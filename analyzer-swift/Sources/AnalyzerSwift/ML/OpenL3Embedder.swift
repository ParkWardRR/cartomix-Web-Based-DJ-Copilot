import Accelerate
import CoreML
import Foundation

/// OpenL3 embedding result for a single time window
public struct OpenL3WindowEmbedding: Sendable {
    /// 512-dimensional embedding vector
    public let vector: [Float]
    /// Window start time in seconds
    public let timestamp: Double
    /// Window duration in seconds
    public let duration: Double
}

/// Track-level OpenL3 embedding with temporal info
public struct OpenL3Embedding: Sendable {
    /// Pooled 512-dimensional embedding (mean of all windows)
    public let vector: [Float]
    /// Per-window embeddings for fine-grained similarity
    public let windows: [OpenL3WindowEmbedding]
    /// Number of windows processed
    public var windowCount: Int { windows.count }

    /// Calculate cosine similarity to another embedding
    public func similarity(to other: OpenL3Embedding) -> Float {
        guard vector.count == other.vector.count, !vector.isEmpty else { return 0 }

        var dotProduct: Float = 0
        var normA: Float = 0
        var normB: Float = 0

        vDSP_dotpr(vector, 1, other.vector, 1, &dotProduct, vDSP_Length(vector.count))
        vDSP_svesq(vector, 1, &normA, vDSP_Length(vector.count))
        vDSP_svesq(other.vector, 1, &normB, vDSP_Length(other.vector.count))

        let denominator = sqrt(normA) * sqrt(normB)
        return denominator > 0 ? dotProduct / denominator : 0
    }
}

/// Errors that can occur during OpenL3 embedding generation
public enum OpenL3Error: Error, CustomStringConvertible {
    case modelNotFound
    case modelLoadFailed(String)
    case inferenceError(String)
    case invalidInput(String)

    public var description: String {
        switch self {
        case .modelNotFound:
            return "OpenL3 Core ML model not found in bundle"
        case .modelLoadFailed(let msg):
            return "Failed to load OpenL3 model: \(msg)"
        case .inferenceError(let msg):
            return "OpenL3 inference failed: \(msg)"
        case .invalidInput(let msg):
            return "Invalid input: \(msg)"
        }
    }
}

/// OpenL3 embedding generator using Core ML
/// Generates 512-dimensional audio embeddings from mel spectrograms
public final class OpenL3Embedder: @unchecked Sendable {
    // Model parameters (must match training config)
    private let sampleRate: Double = 48000
    private let fftSize: Int = 2048
    private let hopSize: Int = 242    // ~199 frames per second
    private let melBands: Int = 128
    private let timeFrames: Int = 199  // 1 second window
    private let embeddingDim: Int = 512

    // Window parameters
    private let windowDuration: Double = 1.0    // 1 second
    private let windowHop: Double = 0.5         // 50% overlap

    // Mel filterbank (computed once)
    private let melFilterbank: [[Float]]

    // Core ML model
    private let model: MLModel?
    private let modelAvailable: Bool

    // FFT resources
    private let fftSetup: FFTSetup
    private let log2n: vDSP_Length
    private let window: [Float]

    public init() {
        // Initialize FFT
        self.log2n = vDSP_Length(log2(Double(fftSize)))
        self.fftSetup = vDSP_create_fftsetup(log2n, FFTRadix(kFFTRadix2))!

        // Create Hann window
        var window = [Float](repeating: 0, count: fftSize)
        vDSP_hann_window(&window, vDSP_Length(fftSize), Int32(vDSP_HANN_NORM))
        self.window = window

        // Compute mel filterbank
        self.melFilterbank = OpenL3Embedder.computeMelFilterbank(
            sampleRate: sampleRate,
            fftSize: fftSize,
            melBands: melBands
        )

        // Load Core ML model
        do {
            let config = MLModelConfiguration()
            config.computeUnits = .all  // Use ANE when available

            // Try to find the model in the bundle
            if let modelURL = Bundle.module.url(forResource: "OpenL3Music", withExtension: "mlmodelc") {
                self.model = try MLModel(contentsOf: modelURL, configuration: config)
                self.modelAvailable = true
            } else if let modelURL = Bundle.module.url(forResource: "OpenL3Music", withExtension: "mlpackage") {
                // Compile mlpackage if needed
                let compiledURL = try MLModel.compileModel(at: modelURL)
                self.model = try MLModel(contentsOf: compiledURL, configuration: config)
                self.modelAvailable = true
            } else {
                print("Warning: OpenL3 model not found - embedding generation will use fallback")
                self.model = nil
                self.modelAvailable = false
            }
        } catch {
            print("Warning: Failed to load OpenL3 model: \(error) - using fallback")
            self.model = nil
            self.modelAvailable = false
        }
    }

    deinit {
        vDSP_destroy_fftsetup(fftSetup)
    }

    /// Check if OpenL3 model is available
    public var isAvailable: Bool { modelAvailable }

    /// Generate OpenL3 embedding from audio samples
    /// - Parameter samples: Audio samples at 48kHz
    /// - Returns: Track-level embedding with per-window details
    public func generate(_ samples: [Float]) -> OpenL3Embedding {
        guard !samples.isEmpty else {
            return OpenL3Embedding(
                vector: [Float](repeating: 0, count: embeddingDim),
                windows: []
            )
        }

        // Generate windowed embeddings
        let windowSamples = Int(windowDuration * sampleRate)
        let hopSamples = Int(windowHop * sampleRate)

        var windowEmbeddings: [OpenL3WindowEmbedding] = []
        var offset = 0

        while offset + windowSamples <= samples.count {
            let windowStart = offset
            let windowEnd = min(offset + windowSamples, samples.count)
            let windowSamples = Array(samples[windowStart..<windowEnd])

            // Compute mel spectrogram for this window
            let melSpec = computeMelSpectrogram(windowSamples)

            // Run inference or use fallback
            let embedding: [Float]
            if modelAvailable, let model = self.model {
                embedding = runInference(melSpec, model: model)
            } else {
                embedding = fallbackEmbedding(melSpec)
            }

            let timestamp = Double(offset) / sampleRate
            windowEmbeddings.append(OpenL3WindowEmbedding(
                vector: embedding,
                timestamp: timestamp,
                duration: windowDuration
            ))

            offset += hopSamples
        }

        // Pool embeddings (mean across all windows)
        let pooled = poolEmbeddings(windowEmbeddings.map { $0.vector })

        return OpenL3Embedding(vector: pooled, windows: windowEmbeddings)
    }

    /// Compute mel spectrogram from audio samples
    /// - Returns: Mel spectrogram [128 bands × 199 frames] in the format expected by OpenL3
    private func computeMelSpectrogram(_ samples: [Float]) -> [[Float]] {
        var padded = samples
        // Pad to ensure we get exactly timeFrames frames
        let targetLength = fftSize + hopSize * (timeFrames - 1)
        if padded.count < targetLength {
            padded.append(contentsOf: [Float](repeating: 0, count: targetLength - padded.count))
        }

        var melFrames: [[Float]] = []

        for frame in 0..<timeFrames {
            let start = frame * hopSize
            let end = start + fftSize
            guard end <= padded.count else { break }

            let frameSamples = Array(padded[start..<end])
            let powerSpectrum = computePowerSpectrum(frameSamples)
            let melBandValues = applyMelFilterbank(powerSpectrum)

            // Convert to log scale (add small epsilon to avoid log(0))
            let logMel = melBandValues.map { log(max($0, 1e-10)) }
            melFrames.append(logMel)
        }

        // Ensure exactly timeFrames
        while melFrames.count < timeFrames {
            melFrames.append([Float](repeating: -23.0, count: melBands)) // log(1e-10)
        }

        return melFrames
    }

    /// Compute power spectrum using FFT
    private func computePowerSpectrum(_ samples: [Float]) -> [Float] {
        var windowed = [Float](repeating: 0, count: fftSize)
        vDSP_vmul(samples, 1, window, 1, &windowed, 1, vDSP_Length(fftSize))

        let halfN = fftSize / 2
        var realp = [Float](repeating: 0, count: halfN)
        var imagp = [Float](repeating: 0, count: halfN)
        var power = [Float](repeating: 0, count: halfN)

        realp.withUnsafeMutableBufferPointer { realpPtr in
            imagp.withUnsafeMutableBufferPointer { imagpPtr in
                power.withUnsafeMutableBufferPointer { powerPtr in
                    windowed.withUnsafeBufferPointer { ptr in
                        ptr.baseAddress!.withMemoryRebound(to: DSPComplex.self, capacity: halfN) { complexPtr in
                            var splitComplex = DSPSplitComplex(realp: realpPtr.baseAddress!, imagp: imagpPtr.baseAddress!)
                            vDSP_ctoz(complexPtr, 2, &splitComplex, 1, vDSP_Length(halfN))
                        }
                    }

                    var splitComplex = DSPSplitComplex(realp: realpPtr.baseAddress!, imagp: imagpPtr.baseAddress!)
                    vDSP_fft_zrip(fftSetup, &splitComplex, 1, log2n, FFTDirection(FFT_FORWARD))

                    // Compute power spectrum (magnitude squared)
                    vDSP_zvmags(&splitComplex, 1, powerPtr.baseAddress!, 1, vDSP_Length(halfN))
                }
            }
        }

        return power
    }

    /// Apply mel filterbank to power spectrum
    private func applyMelFilterbank(_ powerSpectrum: [Float]) -> [Float] {
        var melBandValues = [Float](repeating: 0, count: melBands)

        for (bandIdx, filterCoeffs) in melFilterbank.enumerated() {
            var sum: Float = 0
            let specLen = min(powerSpectrum.count, filterCoeffs.count)
            vDSP_dotpr(powerSpectrum, 1, filterCoeffs, 1, &sum, vDSP_Length(specLen))
            melBandValues[bandIdx] = sum
        }

        return melBandValues
    }

    /// Run Core ML inference
    private func runInference(_ melSpec: [[Float]], model: MLModel) -> [Float] {
        do {
            // Create MLMultiArray with shape [1, 128, 199, 1]
            let inputArray = try MLMultiArray(shape: [1, 128, 199, 1] as [NSNumber], dataType: .float32)

            // Fill the array - melSpec is [199 frames][128 bands]
            // Model expects [batch, mel_bands, time_frames, channels]
            for frameIdx in 0..<min(timeFrames, melSpec.count) {
                let frame = melSpec[frameIdx]
                for bandIdx in 0..<min(melBands, frame.count) {
                    let index = [0, bandIdx, frameIdx, 0] as [NSNumber]
                    inputArray[index] = NSNumber(value: frame[bandIdx])
                }
            }

            // Run prediction
            let input = try MLDictionaryFeatureProvider(dictionary: ["melspectrogram": inputArray])
            let output = try model.prediction(from: input)

            // Extract embedding from output (named 'var_227' from conversion)
            if let embeddingArray = output.featureValue(for: "var_227")?.multiArrayValue {
                var embedding = [Float](repeating: 0, count: embeddingDim)
                for i in 0..<min(embeddingDim, embeddingArray.count) {
                    embedding[i] = embeddingArray[[0, i] as [NSNumber]].floatValue
                }
                return embedding
            }
        } catch {
            print("OpenL3 inference error: \(error)")
        }

        return fallbackEmbedding(melSpec)
    }

    /// Fallback embedding when model is unavailable
    private func fallbackEmbedding(_ melSpec: [[Float]]) -> [Float] {
        // Generate a simpler embedding from mel spectrogram statistics
        var embedding = [Float](repeating: 0, count: embeddingDim)

        guard !melSpec.isEmpty else { return embedding }

        // Compute statistics per mel band
        for bandIdx in 0..<min(melBands, embeddingDim / 4) {
            var bandValues = [Float]()
            for frame in melSpec {
                if bandIdx < frame.count {
                    bandValues.append(frame[bandIdx])
                }
            }

            if !bandValues.isEmpty {
                // Mean
                var mean: Float = 0
                vDSP_meanv(bandValues, 1, &mean, vDSP_Length(bandValues.count))
                embedding[bandIdx] = mean

                // Std
                var std: Float = 0
                var tempMean: Float = 0
                vDSP_normalize(bandValues, 1, nil, 1, &tempMean, &std, vDSP_Length(bandValues.count))
                embedding[bandIdx + 128] = std

                // Delta (first diff mean)
                if bandValues.count > 1 {
                    var delta: Float = 0
                    for i in 1..<bandValues.count {
                        delta += abs(bandValues[i] - bandValues[i-1])
                    }
                    embedding[bandIdx + 256] = delta / Float(bandValues.count - 1)
                }

                // Max
                var maxVal: Float = 0
                vDSP_maxv(bandValues, 1, &maxVal, vDSP_Length(bandValues.count))
                embedding[bandIdx + 384] = maxVal
            }
        }

        // Normalize
        var maxMag: Float = 0
        vDSP_maxmgv(embedding, 1, &maxMag, vDSP_Length(embedding.count))
        if maxMag > 0 {
            var scale = 1.0 / maxMag
            vDSP_vsmul(embedding, 1, &scale, &embedding, 1, vDSP_Length(embedding.count))
        }

        return embedding
    }

    /// Pool multiple embeddings into one (mean pooling)
    private func poolEmbeddings(_ embeddings: [[Float]]) -> [Float] {
        guard !embeddings.isEmpty else {
            return [Float](repeating: 0, count: embeddingDim)
        }

        var pooled = [Float](repeating: 0, count: embeddingDim)

        for emb in embeddings {
            vDSP_vadd(pooled, 1, emb, 1, &pooled, 1, vDSP_Length(min(pooled.count, emb.count)))
        }

        var scale = Float(1.0 / Float(embeddings.count))
        vDSP_vsmul(pooled, 1, &scale, &pooled, 1, vDSP_Length(pooled.count))

        return pooled
    }

    /// Compute mel filterbank coefficients
    private static func computeMelFilterbank(sampleRate: Double, fftSize: Int, melBands: Int) -> [[Float]] {
        let halfN = fftSize / 2
        let fMin: Double = 0
        let fMax: Double = sampleRate / 2

        // Convert to mel scale
        func hzToMel(_ hz: Double) -> Double {
            return 2595 * log10(1 + hz / 700)
        }

        func melToHz(_ mel: Double) -> Double {
            return 700 * (pow(10, mel / 2595) - 1)
        }

        let melMin = hzToMel(fMin)
        let melMax = hzToMel(fMax)

        // Create mel-spaced center frequencies
        var melPoints = [Double]()
        for i in 0...(melBands + 1) {
            let mel = melMin + (melMax - melMin) * Double(i) / Double(melBands + 1)
            melPoints.append(melToHz(mel))
        }

        // Convert to FFT bin indices
        let binPoints = melPoints.map { hz -> Int in
            return Int(round(hz * Double(fftSize) / sampleRate))
        }

        // Create triangular filters
        var filterbank = [[Float]]()

        for m in 0..<melBands {
            var filter = [Float](repeating: 0, count: halfN)
            let left = binPoints[m]
            let center = binPoints[m + 1]
            let right = binPoints[m + 2]

            // Rising slope
            for k in left..<center where k < halfN {
                if center > left {
                    filter[k] = Float(k - left) / Float(center - left)
                }
            }

            // Falling slope
            for k in center..<right where k < halfN {
                if right > center {
                    filter[k] = Float(right - k) / Float(right - center)
                }
            }

            filterbank.append(filter)
        }

        return filterbank
    }
}

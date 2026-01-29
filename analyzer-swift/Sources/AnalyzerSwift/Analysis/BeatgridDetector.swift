import Accelerate

/// Beat marker with timing information
public struct BeatMarker: Sendable {
    public let index: Int
    public let time: Double
    public let isDownbeat: Bool

    public init(index: Int, time: Double, isDownbeat: Bool) {
        self.index = index
        self.time = time
        self.isDownbeat = isDownbeat
    }
}

/// Tempo map node
public struct TempoNode: Sendable {
    public let beatIndex: Int
    public let bpm: Double

    public init(beatIndex: Int, bpm: Double) {
        self.beatIndex = beatIndex
        self.bpm = bpm
    }
}

/// Beatgrid analysis result
public struct BeatgridResult: Sendable {
    public let beats: [BeatMarker]
    public let tempoMap: [TempoNode]
    public let confidence: Double

    public init(beats: [BeatMarker], tempoMap: [TempoNode], confidence: Double) {
        self.beats = beats
        self.tempoMap = tempoMap
        self.confidence = confidence
    }
}

/// Detects beat positions and tempo using onset detection and autocorrelation
public final class BeatgridDetector: @unchecked Sendable {
    private let fft: FFTProcessor
    private let hopSize: Int
    private let sampleRate: Double
    private let tempoFloor: Double
    private let tempoCeil: Double

    public init(
        sampleRate: Double = 48000,
        fftSize: Int = 2048,
        hopSize: Int = 512,
        tempoFloor: Double = 60,
        tempoCeil: Double = 180
    ) {
        self.fft = FFTProcessor(fftSize: fftSize)
        self.hopSize = hopSize
        self.sampleRate = sampleRate
        self.tempoFloor = tempoFloor
        self.tempoCeil = tempoCeil
    }

    /// Detect beatgrid from audio samples
    public func detect(_ samples: [Float]) -> BeatgridResult {
        // 1. Compute STFT
        let spectrogram = fft.stft(samples, hopSize: hopSize)

        // 2. Compute onset strength (spectral flux)
        let onsetStrength = fft.spectralFlux(spectrogram)

        // 3. Estimate tempo via autocorrelation
        let tempo = estimateTempo(onsetStrength)

        // 4. Find beat positions
        let beats = findBeats(onsetStrength, tempo: tempo)

        // 5. Calculate confidence based on beat regularity
        let confidence = calculateConfidence(beats, tempo: tempo)

        return BeatgridResult(
            beats: beats,
            tempoMap: [TempoNode(beatIndex: 0, bpm: tempo)],
            confidence: confidence
        )
    }

    private func estimateTempo(_ onsetStrength: [Float]) -> Double {
        let frameRate = sampleRate / Double(hopSize)

        // Autocorrelation-based tempo estimation
        let minLag = Int(60.0 / tempoCeil * frameRate)
        let maxLag = Int(60.0 / tempoFloor * frameRate)

        guard maxLag < onsetStrength.count else {
            return 120.0 // Default tempo
        }

        var bestLag = minLag
        var bestCorr: Float = 0

        for lag in minLag..<maxLag {
            var corr: Float = 0
            var count = 0

            for i in 0..<(onsetStrength.count - lag) {
                corr += onsetStrength[i] * onsetStrength[i + lag]
                count += 1
            }

            if count > 0 {
                corr /= Float(count)
            }

            if corr > bestCorr {
                bestCorr = corr
                bestLag = lag
            }
        }

        let tempo = 60.0 / (Double(bestLag) / frameRate)
        return clamp(tempo, min: tempoFloor, max: tempoCeil)
    }

    private func findBeats(_ onsetStrength: [Float], tempo: Double) -> [BeatMarker] {
        let frameRate = sampleRate / Double(hopSize)
        let expectedBeatInterval = 60.0 / tempo * frameRate

        // Peak picking with adaptive threshold
        let threshold = calculateAdaptiveThreshold(onsetStrength)
        var peaks = findPeaks(onsetStrength, threshold: threshold, minDistance: Int(expectedBeatInterval * 0.5))

        // Align peaks to regular grid
        var beats = [BeatMarker]()
        var beatIndex = 0

        if let firstPeak = peaks.first {
            // Find the best starting position
            let startFrame = firstPeak

            var frame = startFrame
            while frame < onsetStrength.count {
                let time = Double(frame) * Double(hopSize) / sampleRate
                let isDownbeat = beatIndex % 4 == 0

                beats.append(BeatMarker(index: beatIndex, time: time, isDownbeat: isDownbeat))
                beatIndex += 1
                frame += Int(expectedBeatInterval)
            }
        }

        return beats
    }

    private func calculateAdaptiveThreshold(_ signal: [Float]) -> Float {
        guard !signal.isEmpty else { return 0 }

        // Calculate mean and standard deviation
        var mean: Float = 0
        var stdDev: Float = 0
        var length = vDSP_Length(signal.count)

        vDSP_meanv(signal, 1, &mean, length)
        vDSP_normalize(signal, 1, nil, 1, &mean, &stdDev, length)

        return mean + 0.5 * stdDev
    }

    private func findPeaks(_ signal: [Float], threshold: Float, minDistance: Int) -> [Int] {
        var peaks = [Int]()
        let windowSize = max(3, minDistance / 2)

        for i in windowSize..<(signal.count - windowSize) {
            let value = signal[i]

            if value < threshold {
                continue
            }

            // Check if local maximum
            var isMax = true
            for j in (i - windowSize)..<(i + windowSize) where j != i {
                if signal[j] >= value {
                    isMax = false
                    break
                }
            }

            if isMax {
                // Check minimum distance from last peak
                if let lastPeak = peaks.last, (i - lastPeak) < minDistance {
                    // Keep the higher peak
                    if value > signal[lastPeak] {
                        peaks[peaks.count - 1] = i
                    }
                } else {
                    peaks.append(i)
                }
            }
        }

        return peaks
    }

    private func calculateConfidence(_ beats: [BeatMarker], tempo: Double) -> Double {
        guard beats.count > 2 else { return 0.0 }

        let expectedInterval = 60.0 / tempo
        var deviations = [Double]()

        for i in 1..<beats.count {
            let interval = beats[i].time - beats[i - 1].time
            let deviation = abs(interval - expectedInterval) / expectedInterval
            deviations.append(deviation)
        }

        let meanDeviation = deviations.reduce(0, +) / Double(deviations.count)

        // Confidence is inverse of mean deviation, clamped to [0, 1]
        return clamp(1.0 - meanDeviation * 2, min: 0, max: 1)
    }
}

private func clamp<T: Comparable>(_ value: T, min: T, max: T) -> T {
    Swift.min(Swift.max(value, min), max)
}

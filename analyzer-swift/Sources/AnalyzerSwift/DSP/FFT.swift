import Accelerate

/// FFT processor using Accelerate framework for high-performance spectral analysis
public final class FFTProcessor: @unchecked Sendable {
    private let fftSetup: FFTSetup
    private let log2n: vDSP_Length
    public let fftSize: Int
    private let window: [Float]

    public init(fftSize: Int = 2048) {
        precondition(fftSize.isPowerOfTwo, "FFT size must be power of 2")

        self.fftSize = fftSize
        self.log2n = vDSP_Length(log2(Double(fftSize)))
        self.fftSetup = vDSP_create_fftsetup(log2n, FFTRadix(kFFTRadix2))!

        // Create Hann window
        var window = [Float](repeating: 0, count: fftSize)
        vDSP_hann_window(&window, vDSP_Length(fftSize), Int32(vDSP_HANN_NORM))
        self.window = window
    }

    deinit {
        vDSP_destroy_fftsetup(fftSetup)
    }

    /// Compute magnitude spectrum from time-domain samples
    public func magnitudeSpectrum(_ samples: [Float]) -> [Float] {
        precondition(samples.count >= fftSize, "Input must be at least fftSize")

        // Apply window
        var windowed = [Float](repeating: 0, count: fftSize)
        vDSP_vmul(samples, 1, window, 1, &windowed, 1, vDSP_Length(fftSize))

        // Prepare split complex
        let halfN = fftSize / 2
        var realp = [Float](repeating: 0, count: halfN)
        var imagp = [Float](repeating: 0, count: halfN)
        var magnitudes = [Float](repeating: 0, count: halfN)

        // Use withUnsafeMutableBufferPointer to ensure pointer lifetime
        realp.withUnsafeMutableBufferPointer { realpPtr in
            imagp.withUnsafeMutableBufferPointer { imagpPtr in
                magnitudes.withUnsafeMutableBufferPointer { magPtr in
                    // Convert to split complex
                    windowed.withUnsafeBufferPointer { ptr in
                        ptr.baseAddress!.withMemoryRebound(to: DSPComplex.self, capacity: halfN) { complexPtr in
                            var splitComplex = DSPSplitComplex(realp: realpPtr.baseAddress!, imagp: imagpPtr.baseAddress!)
                            vDSP_ctoz(complexPtr, 2, &splitComplex, 1, vDSP_Length(halfN))
                        }
                    }

                    // Perform FFT
                    var splitComplex = DSPSplitComplex(realp: realpPtr.baseAddress!, imagp: imagpPtr.baseAddress!)
                    vDSP_fft_zrip(fftSetup, &splitComplex, 1, log2n, FFTDirection(FFT_FORWARD))

                    // Compute magnitude
                    vDSP_zvmags(&splitComplex, 1, magPtr.baseAddress!, 1, vDSP_Length(halfN))

                    // Convert to dB scale
                    var one: Float = 1.0
                    vDSP_vdbcon(magPtr.baseAddress!, 1, &one, magPtr.baseAddress!, 1, vDSP_Length(halfN), 1)
                }
            }
        }

        return magnitudes
    }

    /// Compute Short-Time Fourier Transform (spectrogram)
    public func stft(_ samples: [Float], hopSize: Int) -> [[Float]] {
        let frameCount = (samples.count - fftSize) / hopSize + 1
        var spectrogram = [[Float]]()
        spectrogram.reserveCapacity(frameCount)

        for i in 0..<frameCount {
            let start = i * hopSize
            let frame = Array(samples[start..<(start + fftSize)])
            let spectrum = magnitudeSpectrum(frame)
            spectrogram.append(spectrum)
        }

        return spectrogram
    }

    /// Compute spectral flux (onset detection function)
    public func spectralFlux(_ spectrogram: [[Float]]) -> [Float] {
        guard spectrogram.count > 1 else { return [] }

        var flux = [Float](repeating: 0, count: spectrogram.count)
        flux[0] = 0

        for i in 1..<spectrogram.count {
            var sum: Float = 0
            let current = spectrogram[i]
            let previous = spectrogram[i - 1]

            for j in 0..<min(current.count, previous.count) {
                let diff = current[j] - previous[j]
                if diff > 0 {
                    sum += diff * diff
                }
            }
            flux[i] = sqrt(sum)
        }

        return flux
    }

    /// Compute chroma features for key detection
    public func chromaFeatures(_ samples: [Float], sampleRate: Double, hopSize: Int) -> [[Float]] {
        let spectrogram = stft(samples, hopSize: hopSize)
        let binFreqs = frequencyBins(sampleRate: sampleRate)

        var chroma = [[Float]]()
        chroma.reserveCapacity(spectrogram.count)

        for spectrum in spectrogram {
            var chromaVector = [Float](repeating: 0, count: 12)

            for (bin, magnitude) in spectrum.enumerated() where bin < binFreqs.count {
                let freq = binFreqs[bin]
                if freq > 20 && freq < 5000 {
                    let pitchClass = frequencyToPitchClass(freq)
                    chromaVector[pitchClass] += pow(10, magnitude / 20) // Convert from dB
                }
            }

            // Normalize
            let maxVal = chromaVector.max() ?? 1.0
            if maxVal > 0 {
                for i in 0..<12 {
                    chromaVector[i] /= maxVal
                }
            }

            chroma.append(chromaVector)
        }

        return chroma
    }

    private func frequencyBins(sampleRate: Double) -> [Double] {
        let halfN = fftSize / 2
        return (0..<halfN).map { Double($0) * sampleRate / Double(fftSize) }
    }

    private func frequencyToPitchClass(_ freq: Double) -> Int {
        // A4 = 440 Hz = MIDI 69
        let midi = 12 * log2(freq / 440.0) + 69
        let pitchClass = Int(round(midi)).modulo(12)
        return pitchClass
    }
}

extension Int {
    var isPowerOfTwo: Bool {
        self > 0 && (self & (self - 1)) == 0
    }

    func modulo(_ n: Int) -> Int {
        let result = self % n
        return result >= 0 ? result : result + n
    }
}

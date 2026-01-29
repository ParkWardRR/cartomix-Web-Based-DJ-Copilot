import Accelerate

/// Energy analysis result for a track
public struct EnergyResult: Sendable {
    public let globalEnergy: Int // 1-10 scale
    public let curve: [Float]    // Normalized energy over time
    public let rms: Float        // Root mean square
    public let peak: Float       // Peak amplitude
    public let dynamicRange: Float
    public let lowEnergy: Float  // Sub-bass/bass energy
    public let midEnergy: Float  // Mid-range energy
    public let highEnergy: Float // High frequency energy

    public init(
        globalEnergy: Int,
        curve: [Float],
        rms: Float,
        peak: Float,
        dynamicRange: Float,
        lowEnergy: Float,
        midEnergy: Float,
        highEnergy: Float
    ) {
        self.globalEnergy = globalEnergy
        self.curve = curve
        self.rms = rms
        self.peak = peak
        self.dynamicRange = dynamicRange
        self.lowEnergy = lowEnergy
        self.midEnergy = midEnergy
        self.highEnergy = highEnergy
    }
}

/// Analyzes track energy using spectral and amplitude features
public final class EnergyAnalyzer: @unchecked Sendable {
    private let fft: FFTProcessor
    private let hopSize: Int
    private let sampleRate: Double

    // Frequency band boundaries (Hz)
    private let lowCutoff: Double = 250
    private let midCutoff: Double = 4000

    public init(sampleRate: Double = 48000, fftSize: Int = 2048, hopSize: Int = 1024) {
        self.fft = FFTProcessor(fftSize: fftSize)
        self.hopSize = hopSize
        self.sampleRate = sampleRate
    }

    /// Analyze energy characteristics
    public func analyze(_ samples: [Float]) -> EnergyResult {
        // 1. Calculate RMS and peak
        let (rms, peak) = calculateLoudness(samples)

        // 2. Compute spectrogram for band analysis
        let spectrogram = fft.stft(samples, hopSize: hopSize)

        // 3. Calculate band energies
        let (low, mid, high) = calculateBandEnergies(spectrogram)

        // 4. Build energy curve
        let curve = calculateEnergyCurve(spectrogram)

        // 5. Calculate dynamic range
        let dynamicRange = peak > 0 ? 20 * log10(peak / max(rms, 1e-10)) : 0

        // 6. Compute global energy (1-10 scale)
        let globalEnergy = computeGlobalEnergy(rms: rms, low: low, mid: mid, high: high)

        return EnergyResult(
            globalEnergy: globalEnergy,
            curve: curve,
            rms: rms,
            peak: peak,
            dynamicRange: dynamicRange,
            lowEnergy: low,
            midEnergy: mid,
            highEnergy: high
        )
    }

    private func calculateLoudness(_ samples: [Float]) -> (rms: Float, peak: Float) {
        guard !samples.isEmpty else { return (0, 0) }

        var rms: Float = 0
        vDSP_rmsqv(samples, 1, &rms, vDSP_Length(samples.count))

        var peak: Float = 0
        vDSP_maxmgv(samples, 1, &peak, vDSP_Length(samples.count))

        return (rms, peak)
    }

    private func calculateBandEnergies(_ spectrogram: [[Float]]) -> (low: Float, mid: Float, high: Float) {
        guard !spectrogram.isEmpty, !spectrogram[0].isEmpty else {
            return (0, 0, 0)
        }

        let binCount = spectrogram[0].count
        let binWidth = sampleRate / Double(fft.fftSize)

        let lowBin = Int(lowCutoff / binWidth)
        let midBin = Int(midCutoff / binWidth)

        var lowSum: Float = 0
        var midSum: Float = 0
        var highSum: Float = 0
        var frameCount: Float = 0

        for spectrum in spectrogram {
            for bin in 0..<min(binCount, spectrum.count) {
                let magnitude = pow(10, spectrum[bin] / 20) // Convert from dB

                if bin < lowBin {
                    lowSum += magnitude
                } else if bin < midBin {
                    midSum += magnitude
                } else {
                    highSum += magnitude
                }
            }
            frameCount += 1
        }

        if frameCount > 0 {
            lowSum /= frameCount
            midSum /= frameCount
            highSum /= frameCount
        }

        // Normalize to 0-1 range
        let total = lowSum + midSum + highSum
        if total > 0 {
            return (lowSum / total, midSum / total, highSum / total)
        }

        return (0.33, 0.33, 0.33)
    }

    private func calculateEnergyCurve(_ spectrogram: [[Float]]) -> [Float] {
        var curve = [Float](repeating: 0, count: spectrogram.count)

        for (i, spectrum) in spectrogram.enumerated() {
            var energy: Float = 0
            for magnitude in spectrum {
                energy += pow(10, magnitude / 20) // Convert from dB
            }
            curve[i] = energy / Float(spectrum.count)
        }

        // Normalize curve
        let maxEnergy = curve.max() ?? 1.0
        if maxEnergy > 0 {
            for i in 0..<curve.count {
                curve[i] /= maxEnergy
            }
        }

        // Smooth with moving average
        let windowSize = 10
        var smoothed = [Float](repeating: 0, count: curve.count)
        for i in 0..<curve.count {
            var sum: Float = 0
            var count: Float = 0
            for j in max(0, i - windowSize / 2)..<min(curve.count, i + windowSize / 2 + 1) {
                sum += curve[j]
                count += 1
            }
            smoothed[i] = sum / count
        }

        return smoothed
    }

    private func computeGlobalEnergy(rms: Float, low: Float, mid: Float, high: Float) -> Int {
        // Combine factors for global energy
        // RMS contributes to overall loudness
        // Low frequency energy indicates bass-heavy (club) tracks
        // High energy indicates brightness

        let rmsNorm = min(1.0, rms * 5) // RMS typically 0.1-0.3 for music

        // Weight bass energy more for club music
        let weightedEnergy = rmsNorm * 0.4 + low * 0.35 + mid * 0.15 + high * 0.1

        // Map to 1-10 scale
        let energy = Int(round(weightedEnergy * 9)) + 1
        return min(10, max(1, energy))
    }
}

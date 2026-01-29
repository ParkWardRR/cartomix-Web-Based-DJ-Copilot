import Accelerate

/// Musical key with various representations
public struct MusicalKey: Sendable {
    public let pitchClass: Int  // 0-11, where 0 = C
    public let isMinor: Bool
    public let confidence: Double

    public init(pitchClass: Int, isMinor: Bool, confidence: Double) {
        self.pitchClass = pitchClass
        self.isMinor = isMinor
        self.confidence = confidence
    }

    /// Key name in standard notation (e.g., "Am", "C")
    public var name: String {
        let names = ["C", "C#", "D", "Eb", "E", "F", "F#", "G", "Ab", "A", "Bb", "B"]
        let name = names[pitchClass]
        return isMinor ? "\(name)m" : name
    }

    /// Camelot notation (e.g., "8A", "8B")
    public var camelot: String {
        // Camelot wheel mapping
        // Minor keys (A): 1A=Abm, 2A=Ebm, 3A=Bbm, 4A=Fm, 5A=Cm, 6A=Gm, 7A=Dm, 8A=Am, 9A=Em, 10A=Bm, 11A=F#m, 12A=C#m
        // Major keys (B): 1B=B, 2B=F#, 3B=Db, 4B=Ab, 5B=Eb, 6B=Bb, 7B=F, 8B=C, 9B=G, 10B=D, 11B=A, 12B=E

        // Pitch class to Camelot number
        let minorMap = [5, 12, 7, 2, 9, 4, 11, 6, 1, 8, 3, 10] // C=5A, C#=12A, D=7A, etc.
        let majorMap = [8, 3, 10, 5, 12, 7, 2, 9, 4, 11, 6, 1] // C=8B, C#=3B, D=10B, etc.

        let number = isMinor ? minorMap[pitchClass] : majorMap[pitchClass]
        let letter = isMinor ? "A" : "B"

        return "\(number)\(letter)"
    }

    /// Open Key notation (e.g., "1m", "1d")
    public var openKey: String {
        // Similar to Camelot but with different numbering
        let minorMap = [1, 8, 3, 10, 5, 12, 7, 2, 9, 4, 11, 6]
        let majorMap = [1, 8, 3, 10, 5, 12, 7, 2, 9, 4, 11, 6]

        let number = isMinor ? minorMap[pitchClass] : majorMap[pitchClass]
        let suffix = isMinor ? "m" : "d"

        return "\(number)\(suffix)"
    }
}

/// Key profile templates (Krumhansl-Schmuckler)
private let majorProfile: [Float] = [
    6.35, 2.23, 3.48, 2.33, 4.38, 4.09,
    2.52, 5.19, 2.39, 3.66, 2.29, 2.88
]

private let minorProfile: [Float] = [
    6.33, 2.68, 3.52, 5.38, 2.60, 3.53,
    2.54, 4.75, 3.98, 2.69, 3.34, 3.17
]

/// Detects musical key using chroma-based analysis
public final class KeyDetector: @unchecked Sendable {
    private let fft: FFTProcessor
    private let hopSize: Int
    private let sampleRate: Double

    public init(sampleRate: Double = 48000, fftSize: Int = 4096, hopSize: Int = 2048) {
        self.fft = FFTProcessor(fftSize: fftSize)
        self.hopSize = hopSize
        self.sampleRate = sampleRate
    }

    /// Detect key from audio samples
    public func detect(_ samples: [Float]) -> MusicalKey {
        // 1. Extract chroma features
        let chroma = fft.chromaFeatures(samples, sampleRate: sampleRate, hopSize: hopSize)

        // 2. Average chroma across all frames
        let avgChroma = averageChroma(chroma)

        // 3. Correlate with key profiles
        var bestKey = 0
        var bestIsMinor = false
        var bestCorr: Float = -Float.infinity

        for pitchClass in 0..<12 {
            // Rotate profile to match pitch class
            let majorCorr = correlate(avgChroma, with: rotate(majorProfile, by: pitchClass))
            let minorCorr = correlate(avgChroma, with: rotate(minorProfile, by: pitchClass))

            if majorCorr > bestCorr {
                bestCorr = majorCorr
                bestKey = pitchClass
                bestIsMinor = false
            }

            if minorCorr > bestCorr {
                bestCorr = minorCorr
                bestKey = pitchClass
                bestIsMinor = true
            }
        }

        // Normalize correlation to confidence
        let confidence = normalizeCorrelation(bestCorr)

        return MusicalKey(pitchClass: bestKey, isMinor: bestIsMinor, confidence: confidence)
    }

    private func averageChroma(_ chroma: [[Float]]) -> [Float] {
        guard !chroma.isEmpty else {
            return [Float](repeating: 0, count: 12)
        }

        var avg = [Float](repeating: 0, count: 12)
        let scale = 1.0 / Float(chroma.count)

        for frame in chroma {
            for i in 0..<12 {
                avg[i] += frame[i] * scale
            }
        }

        // Normalize
        let maxVal = avg.max() ?? 1.0
        if maxVal > 0 {
            for i in 0..<12 {
                avg[i] /= maxVal
            }
        }

        return avg
    }

    private func rotate(_ array: [Float], by amount: Int) -> [Float] {
        let n = array.count
        return (0..<n).map { array[($0 + amount) % n] }
    }

    private func correlate(_ a: [Float], with b: [Float]) -> Float {
        guard a.count == b.count else { return 0 }

        var meanA: Float = 0
        var meanB: Float = 0
        vDSP_meanv(a, 1, &meanA, vDSP_Length(a.count))
        vDSP_meanv(b, 1, &meanB, vDSP_Length(b.count))

        var numerator: Float = 0
        var denomA: Float = 0
        var denomB: Float = 0

        for i in 0..<a.count {
            let diffA = a[i] - meanA
            let diffB = b[i] - meanB
            numerator += diffA * diffB
            denomA += diffA * diffA
            denomB += diffB * diffB
        }

        let denom = sqrt(denomA * denomB)
        return denom > 0 ? numerator / denom : 0
    }

    private func normalizeCorrelation(_ corr: Float) -> Double {
        // Map correlation [-1, 1] to confidence [0, 1]
        // Typical good correlations are 0.6-0.9
        return Double(min(1, max(0, (corr + 1) / 2)))
    }
}

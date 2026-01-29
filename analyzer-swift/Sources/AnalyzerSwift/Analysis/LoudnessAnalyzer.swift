import Accelerate

/// EBU R128 loudness measurement result
public struct LoudnessResult: Sendable {
    /// Integrated loudness in LUFS (Loudness Units Full Scale)
    public let integratedLoudness: Float
    /// Loudness range in LU
    public let loudnessRange: Float
    /// Short-term loudness maximum in LUFS
    public let shortTermMax: Float
    /// Momentary loudness maximum in LUFS
    public let momentaryMax: Float
    /// True peak in dBTP (decibels True Peak)
    public let truePeak: Float
    /// Sample peak in dBFS
    public let samplePeak: Float

    public init(
        integratedLoudness: Float,
        loudnessRange: Float,
        shortTermMax: Float,
        momentaryMax: Float,
        truePeak: Float,
        samplePeak: Float
    ) {
        self.integratedLoudness = integratedLoudness
        self.loudnessRange = loudnessRange
        self.shortTermMax = shortTermMax
        self.momentaryMax = momentaryMax
        self.truePeak = truePeak
        self.samplePeak = samplePeak
    }
}

/// EBU R128 loudness analyzer
/// Implements ITU-R BS.1770-4 loudness measurement
public final class LoudnessAnalyzer: @unchecked Sendable {
    private let sampleRate: Double
    private let blockSize: Int  // 400ms for momentary, 3s for short-term
    private let hopSize: Int    // 100ms hop for gating

    // K-weighting filter coefficients (pre-computed for common sample rates)
    private var highShelfB: [Double] = []
    private var highShelfA: [Double] = []
    private var highPassB: [Double] = []
    private var highPassA: [Double] = []

    public init(sampleRate: Double = 48000) {
        self.sampleRate = sampleRate
        self.blockSize = Int(0.4 * sampleRate)  // 400ms block
        self.hopSize = Int(0.1 * sampleRate)    // 100ms hop

        // Initialize K-weighting filters
        initializeFilters()
    }

    /// Analyze loudness according to EBU R128
    public func analyze(_ samples: [Float]) -> LoudnessResult {
        guard samples.count > blockSize else {
            return LoudnessResult(
                integratedLoudness: -70,
                loudnessRange: 0,
                shortTermMax: -70,
                momentaryMax: -70,
                truePeak: -70,
                samplePeak: -70
            )
        }

        // 1. Apply K-weighting filter
        let weighted = applyKWeighting(samples)

        // 2. Calculate momentary loudness (400ms blocks)
        let momentaryLoudness = calculateMomentaryLoudness(weighted)

        // 3. Calculate short-term loudness (3s blocks)
        let shortTermLoudness = calculateShortTermLoudness(weighted)

        // 4. Calculate integrated loudness with gating
        let integratedLoudness = calculateIntegratedLoudness(momentaryLoudness)

        // 5. Calculate loudness range (LRA)
        let loudnessRange = calculateLoudnessRange(shortTermLoudness)

        // 6. Calculate peaks
        let samplePeak = calculateSamplePeak(samples)
        let truePeak = calculateTruePeak(samples)

        return LoudnessResult(
            integratedLoudness: integratedLoudness,
            loudnessRange: loudnessRange,
            shortTermMax: shortTermLoudness.max() ?? -70,
            momentaryMax: momentaryLoudness.max() ?? -70,
            truePeak: truePeak,
            samplePeak: samplePeak
        )
    }

    private func initializeFilters() {
        // High shelf filter (pre-filter) - boosts high frequencies
        // Coefficients for 48kHz, adapted from ITU-R BS.1770-4
        let fc = 1681.974450955533  // Shelf frequency
        let G = 3.999843853973347   // Gain in dB
        let Q = 0.7071752369554196  // Q factor

        let K = tan(.pi * fc / sampleRate)
        let Vh = pow(10, G / 20)
        let Vb = pow(Vh, 0.4996667741545416)

        let a0 = 1 + K / Q + K * K
        highShelfB = [
            (Vh + Vb * K / Q + K * K) / a0,
            2 * (K * K - Vh) / a0,
            (Vh - Vb * K / Q + K * K) / a0
        ]
        highShelfA = [
            1,
            2 * (K * K - 1) / a0,
            (1 - K / Q + K * K) / a0
        ]

        // High-pass filter (RLB weighting)
        let fc2 = 38.13547087602444  // High-pass frequency
        let Q2 = 0.5003270373238773

        let K2 = tan(.pi * fc2 / sampleRate)
        let a02 = 1 + K2 / Q2 + K2 * K2
        highPassB = [
            1 / a02,
            -2 / a02,
            1 / a02
        ]
        highPassA = [
            1,
            2 * (K2 * K2 - 1) / a02,
            (1 - K2 / Q2 + K2 * K2) / a02
        ]
    }

    private func applyKWeighting(_ samples: [Float]) -> [Float] {
        // Apply high shelf filter (stage 1)
        var stage1 = applyBiquad(samples, b: highShelfB, a: highShelfA)

        // Apply high-pass filter (stage 2)
        let stage2 = applyBiquad(stage1, b: highPassB, a: highPassA)

        return stage2
    }

    private func applyBiquad(_ input: [Float], b: [Double], a: [Double]) -> [Float] {
        var output = [Float](repeating: 0, count: input.count)
        var z1: Double = 0
        var z2: Double = 0

        for i in 0..<input.count {
            let x = Double(input[i])
            let y = b[0] * x + z1
            z1 = b[1] * x - a[1] * y + z2
            z2 = b[2] * x - a[2] * y
            output[i] = Float(y)
        }

        return output
    }

    private func calculateMomentaryLoudness(_ weighted: [Float]) -> [Float] {
        var loudness = [Float]()
        let numBlocks = (weighted.count - blockSize) / hopSize + 1

        for i in 0..<numBlocks {
            let start = i * hopSize
            let end = min(start + blockSize, weighted.count)
            let block = Array(weighted[start..<end])

            // Calculate mean square
            var meanSquare: Float = 0
            vDSP_measqv(block, 1, &meanSquare, vDSP_Length(block.count))

            // Convert to LUFS
            let lufs = meanSquare > 0 ? -0.691 + 10 * log10(meanSquare) : -70
            loudness.append(Float(lufs))
        }

        return loudness
    }

    private func calculateShortTermLoudness(_ weighted: [Float]) -> [Float] {
        let shortTermBlockSize = Int(3.0 * sampleRate)  // 3 seconds
        var loudness = [Float]()
        let numBlocks = max(1, (weighted.count - shortTermBlockSize) / hopSize + 1)

        for i in 0..<numBlocks {
            let start = i * hopSize
            let end = min(start + shortTermBlockSize, weighted.count)
            let block = Array(weighted[start..<end])

            var meanSquare: Float = 0
            vDSP_measqv(block, 1, &meanSquare, vDSP_Length(block.count))

            let lufs = meanSquare > 0 ? -0.691 + 10 * log10(meanSquare) : -70
            loudness.append(Float(lufs))
        }

        return loudness
    }

    private func calculateIntegratedLoudness(_ momentary: [Float]) -> Float {
        guard !momentary.isEmpty else { return -70 }

        // Absolute gate at -70 LUFS
        let absoluteGate: Float = -70
        var gatedBlocks = momentary.filter { $0 > absoluteGate }

        guard !gatedBlocks.isEmpty else { return -70 }

        // Calculate ungated mean
        let ungatedMean = gatedBlocks.reduce(0, +) / Float(gatedBlocks.count)

        // Relative gate at -10 LU below ungated mean
        let relativeGate = ungatedMean - 10

        // Apply relative gate
        gatedBlocks = gatedBlocks.filter { $0 > relativeGate }

        guard !gatedBlocks.isEmpty else { return ungatedMean }

        // Calculate final integrated loudness
        // Convert back to linear, average, then convert to LUFS
        var linearSum: Float = 0
        for block in gatedBlocks {
            linearSum += pow(10, block / 10)
        }

        let meanLinear = linearSum / Float(gatedBlocks.count)
        return 10 * log10(meanLinear)
    }

    private func calculateLoudnessRange(_ shortTerm: [Float]) -> Float {
        guard shortTerm.count > 1 else { return 0 }

        // Sort short-term loudness values
        let sorted = shortTerm.filter { $0 > -70 }.sorted()
        guard sorted.count > 10 else { return 0 }

        // Calculate percentiles (10th and 95th)
        let lowIndex = Int(Float(sorted.count) * 0.10)
        let highIndex = Int(Float(sorted.count) * 0.95)

        let lowPercentile = sorted[lowIndex]
        let highPercentile = sorted[highIndex]

        return highPercentile - lowPercentile
    }

    private func calculateSamplePeak(_ samples: [Float]) -> Float {
        var peak: Float = 0
        vDSP_maxmgv(samples, 1, &peak, vDSP_Length(samples.count))

        // Convert to dBFS
        return peak > 0 ? 20 * log10(peak) : -70
    }

    private func calculateTruePeak(_ samples: [Float]) -> Float {
        // True peak requires 4x oversampling
        // For simplicity, we'll use a fast approximation with linear interpolation
        // A full implementation would use polyphase filters

        let oversampleFactor = 4
        var maxPeak: Float = 0

        for i in 0..<(samples.count - 1) {
            let s1 = samples[i]
            let s2 = samples[i + 1]

            // Check original samples
            maxPeak = max(maxPeak, abs(s1))

            // Interpolate between samples
            for j in 1..<oversampleFactor {
                let t = Float(j) / Float(oversampleFactor)
                let interpolated = s1 * (1 - t) + s2 * t
                maxPeak = max(maxPeak, abs(interpolated))
            }
        }

        // Check last sample
        if let last = samples.last {
            maxPeak = max(maxPeak, abs(last))
        }

        // Convert to dBTP
        return maxPeak > 0 ? 20 * log10(maxPeak) : -70
    }
}

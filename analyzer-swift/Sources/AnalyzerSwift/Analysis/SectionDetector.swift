import Accelerate

/// Section types in a DJ track
public enum SectionType: String, Sendable, CaseIterable {
    case intro
    case verse
    case build
    case drop
    case breakdown
    case outro
}

/// A detected section with timing and confidence
public struct Section: Sendable {
    public let type: SectionType
    public let startTime: Double
    public let endTime: Double
    public let startBeat: Int
    public let endBeat: Int
    public let confidence: Double

    public init(
        type: SectionType,
        startTime: Double,
        endTime: Double,
        startBeat: Int,
        endBeat: Int,
        confidence: Double
    ) {
        self.type = type
        self.startTime = startTime
        self.endTime = endTime
        self.startBeat = startBeat
        self.endBeat = endBeat
        self.confidence = confidence
    }

    public var duration: Double { endTime - startTime }
}

/// Section detection result
public struct SectionResult: Sendable {
    public let sections: [Section]
    public let transitionWindows: [(start: Double, end: Double)]
    public let confidence: Double

    public init(sections: [Section], transitionWindows: [(start: Double, end: Double)], confidence: Double) {
        self.sections = sections
        self.transitionWindows = transitionWindows
        self.confidence = confidence
    }
}

/// Detects song sections using energy segmentation and structural analysis
public final class SectionDetector: @unchecked Sendable {
    private let fft: FFTProcessor
    private let hopSize: Int
    private let sampleRate: Double

    // Section detection parameters
    private let minSectionBeats: Int = 16 // Minimum 16 beats per section (4 bars at 4/4)
    private let phraseBeats: Int = 32     // Standard phrase length (8 bars)

    public init(sampleRate: Double = 48000, fftSize: Int = 4096, hopSize: Int = 2048) {
        self.fft = FFTProcessor(fftSize: fftSize)
        self.hopSize = hopSize
        self.sampleRate = sampleRate
    }

    /// Detect sections from audio samples and beatgrid
    public func detect(_ samples: [Float], beats: [BeatMarker], tempo: Double) -> SectionResult {
        guard !beats.isEmpty else {
            return SectionResult(sections: [], transitionWindows: [], confidence: 0)
        }

        // 1. Compute energy curve
        let spectrogram = fft.stft(samples, hopSize: hopSize)
        let energyCurve = computeEnergyCurve(spectrogram)

        // 2. Map energy to beat positions
        let beatEnergies = mapEnergyToBeats(energyCurve, beats: beats)

        // 3. Find section boundaries using energy changes
        let boundaries = findSectionBoundaries(beatEnergies)

        // 4. Classify sections
        let sections = classifySections(boundaries, beatEnergies: beatEnergies, beats: beats)

        // 5. Find transition windows
        let transitionWindows = findTransitionWindows(sections)

        // 6. Calculate overall confidence
        let confidence = calculateConfidence(sections)

        return SectionResult(
            sections: sections,
            transitionWindows: transitionWindows,
            confidence: confidence
        )
    }

    private func computeEnergyCurve(_ spectrogram: [[Float]]) -> [Float] {
        var curve = [Float](repeating: 0, count: spectrogram.count)

        for (i, spectrum) in spectrogram.enumerated() {
            var sum: Float = 0
            for magnitude in spectrum {
                sum += pow(10, magnitude / 20)
            }
            curve[i] = sum / Float(max(1, spectrum.count))
        }

        // Normalize
        let maxVal = curve.max() ?? 1.0
        if maxVal > 0 {
            for i in 0..<curve.count {
                curve[i] /= maxVal
            }
        }

        return curve
    }

    private func mapEnergyToBeats(_ energyCurve: [Float], beats: [BeatMarker]) -> [Float] {
        guard !energyCurve.isEmpty, !beats.isEmpty else { return [] }

        let frameRate = sampleRate / Double(hopSize)
        var beatEnergies = [Float](repeating: 0, count: beats.count)

        for (i, beat) in beats.enumerated() {
            let frame = Int(beat.time * frameRate)
            if frame < energyCurve.count {
                beatEnergies[i] = energyCurve[frame]
            }
        }

        return beatEnergies
    }

    private func findSectionBoundaries(_ beatEnergies: [Float]) -> [Int] {
        guard !beatEnergies.isEmpty else { return [] }

        var boundaries = [0] // Always start at beat 0

        // Calculate moving average for baseline
        let windowSize = phraseBeats
        var baseline = [Float](repeating: 0, count: beatEnergies.count)

        for i in 0..<beatEnergies.count {
            var sum: Float = 0
            var count: Float = 0
            let start = max(0, i - windowSize / 2)
            let end = min(beatEnergies.count, i + windowSize / 2)
            for j in start..<end {
                sum += beatEnergies[j]
                count += 1
            }
            baseline[i] = sum / count
        }

        // Find significant energy changes
        for i in stride(from: phraseBeats, to: beatEnergies.count - minSectionBeats, by: minSectionBeats) {
            // Look for energy jumps at phrase boundaries
            if i % phraseBeats == 0 {
                let before = average(beatEnergies, start: max(0, i - 8), end: i)
                let after = average(beatEnergies, start: i, end: min(beatEnergies.count, i + 8))

                let change = abs(after - before)
                let threshold: Float = 0.15

                if change > threshold {
                    // Significant energy change at phrase boundary
                    if let lastBoundary = boundaries.last, (i - lastBoundary) >= minSectionBeats {
                        boundaries.append(i)
                    }
                }
            }
        }

        // Add end boundary
        boundaries.append(beatEnergies.count)

        return boundaries
    }

    private func average(_ array: [Float], start: Int, end: Int) -> Float {
        guard end > start else { return 0 }
        var sum: Float = 0
        for i in start..<end {
            sum += array[i]
        }
        return sum / Float(end - start)
    }

    private func classifySections(_ boundaries: [Int], beatEnergies: [Float], beats: [BeatMarker]) -> [Section] {
        guard boundaries.count >= 2, !beats.isEmpty else { return [] }

        var sections = [Section]()
        let totalBeats = boundaries.last ?? beatEnergies.count

        for i in 0..<(boundaries.count - 1) {
            let startBeat = boundaries[i]
            let endBeat = boundaries[i + 1]

            guard startBeat < beats.count, endBeat <= beats.count else { continue }

            let startTime = beats[startBeat].time
            let endTime = endBeat < beats.count ? beats[endBeat].time : beats[beats.count - 1].time

            // Calculate average energy for this section
            let sectionEnergy = average(beatEnergies, start: startBeat, end: min(endBeat, beatEnergies.count))

            // Calculate energy variance for section
            var variance: Float = 0
            for b in startBeat..<min(endBeat, beatEnergies.count) {
                let diff = beatEnergies[b] - sectionEnergy
                variance += diff * diff
            }
            variance /= Float(max(1, endBeat - startBeat))

            // Classify based on position and energy
            let relativePosition = Double(startBeat) / Double(max(1, totalBeats))
            let type = classifySectionType(
                energy: sectionEnergy,
                variance: variance,
                position: relativePosition,
                isFirst: i == 0,
                isLast: i == boundaries.count - 2
            )

            let confidence = 0.7 + Double(variance) * 0.3 // Higher variance = more confident classification

            sections.append(Section(
                type: type,
                startTime: startTime,
                endTime: endTime,
                startBeat: startBeat,
                endBeat: endBeat,
                confidence: min(1.0, confidence)
            ))
        }

        return sections
    }

    private func classifySectionType(
        energy: Float,
        variance: Float,
        position: Double,
        isFirst: Bool,
        isLast: Bool
    ) -> SectionType {
        // Position-based heuristics
        if isFirst && position < 0.1 {
            return .intro
        }

        if isLast && position > 0.85 {
            return .outro
        }

        // Energy-based classification
        if energy > 0.75 {
            return .drop
        }

        if energy < 0.35 {
            if variance < 0.05 {
                return .breakdown
            }
            return .verse
        }

        // Medium energy with increasing trend could be build
        if variance > 0.1 && energy > 0.5 {
            return .build
        }

        return .verse
    }

    private func findTransitionWindows(_ sections: [Section]) -> [(start: Double, end: Double)] {
        var windows = [(start: Double, end: Double)]()

        for section in sections {
            switch section.type {
            case .intro:
                // Good to mix out of during intro (incoming track)
                let start = section.endTime - min(16, section.duration / 2)
                windows.append((start, section.endTime))

            case .outro:
                // Good to mix into during outro (outgoing track)
                windows.append((section.startTime, section.startTime + min(16, section.duration / 2)))

            case .breakdown:
                // Breakdowns are good transition points
                windows.append((section.startTime, section.endTime))

            default:
                break
            }
        }

        return windows
    }

    private func calculateConfidence(_ sections: [Section]) -> Double {
        guard !sections.isEmpty else { return 0 }

        // Confidence based on:
        // 1. Having expected structure (intro, body, outro)
        // 2. Section classification confidence

        var structureScore = 0.0

        let hasIntro = sections.first?.type == .intro
        let hasOutro = sections.last?.type == .outro
        let hasDrop = sections.contains { $0.type == .drop }

        if hasIntro { structureScore += 0.25 }
        if hasOutro { structureScore += 0.25 }
        if hasDrop { structureScore += 0.25 }
        if sections.count >= 3 { structureScore += 0.25 }

        let avgConfidence = sections.map(\.confidence).reduce(0, +) / Double(sections.count)

        return structureScore * 0.5 + avgConfidence * 0.5
    }
}

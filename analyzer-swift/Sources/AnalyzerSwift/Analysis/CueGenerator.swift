import Foundation

/// Cue point types for DJ software
public enum CueType: String, Sendable {
    case load = "CUE_LOAD"
    case introStart = "CUE_INTRO_START"
    case introEnd = "CUE_INTRO_END"
    case build = "CUE_BUILD"
    case drop = "CUE_DROP"
    case breakdown = "CUE_BREAKDOWN"
    case outroStart = "CUE_OUTRO_START"
    case outroEnd = "CUE_OUTRO_END"
    case marker = "CUE_MARKER"
}

/// A generated cue point
public struct CuePoint: Sendable {
    public let type: CueType
    public let beatIndex: Int
    public let time: Double
    public let label: String
    public let color: CueColor

    public init(type: CueType, beatIndex: Int, time: Double, label: String, color: CueColor) {
        self.type = type
        self.beatIndex = beatIndex
        self.time = time
        self.label = label
        self.color = color
    }
}

/// Standard cue colors (compatible with Rekordbox/Serato)
public enum CueColor: Int, Sendable {
    case red = 0xFF0000
    case orange = 0xFF8000
    case yellow = 0xFFFF00
    case green = 0x00FF00
    case cyan = 0x00FFFF
    case blue = 0x0000FF
    case purple = 0x8000FF
    case pink = 0xFF00FF

    /// Default color for each cue type
    public static func forType(_ type: CueType) -> CueColor {
        switch type {
        case .load: return .green
        case .introStart: return .blue
        case .introEnd: return .blue
        case .build: return .yellow
        case .drop: return .red
        case .breakdown: return .purple
        case .outroStart: return .orange
        case .outroEnd: return .orange
        case .marker: return .cyan
        }
    }
}

/// Cue generation result
public struct CueResult: Sendable {
    public let cues: [CuePoint]
    public let safeStartBeat: Int  // Safe point to start playback
    public let safeEndBeat: Int    // Safe point to end/transition out

    public init(cues: [CuePoint], safeStartBeat: Int, safeEndBeat: Int) {
        self.cues = cues
        self.safeStartBeat = safeStartBeat
        self.safeEndBeat = safeEndBeat
    }
}

/// Generates DJ cue points from sections and beats
public final class CueGenerator: @unchecked Sendable {
    private let maxCues: Int

    public init(maxCues: Int = 8) {
        self.maxCues = maxCues
    }

    /// Generate cue points from sections and beats
    public func generate(sections: [Section], beats: [BeatMarker]) -> CueResult {
        guard !beats.isEmpty else {
            return CueResult(cues: [], safeStartBeat: 0, safeEndBeat: 0)
        }

        var cues = [CuePoint]()

        // 1. Always add load point at beat 0
        cues.append(CuePoint(
            type: .load,
            beatIndex: 0,
            time: beats[0].time,
            label: "Load",
            color: .forType(.load)
        ))

        // 2. Add section-based cues
        for section in sections {
            let cueType = cueTypeForSection(section.type, isStart: true)

            // Add start cue (beat-aligned)
            let startBeat = alignToDownbeat(section.startBeat, beats: beats)
            if startBeat > 0 && !hasCueNear(cues, beat: startBeat, threshold: 8) {
                let time = startBeat < beats.count ? beats[startBeat].time : section.startTime
                cues.append(CuePoint(
                    type: cueType,
                    beatIndex: startBeat,
                    time: time,
                    label: labelForSection(section.type),
                    color: .forType(cueType)
                ))
            }

            // For intro/outro, also add end markers
            if section.type == .intro || section.type == .outro {
                let endCueType = cueTypeForSection(section.type, isStart: false)
                let endBeat = alignToDownbeat(section.endBeat, beats: beats)
                if endBeat > 0 && !hasCueNear(cues, beat: endBeat, threshold: 8) {
                    let time = endBeat < beats.count ? beats[endBeat].time : section.endTime
                    cues.append(CuePoint(
                        type: endCueType,
                        beatIndex: endBeat,
                        time: time,
                        label: section.type == .intro ? "Intro End" : "Outro End",
                        color: .forType(endCueType)
                    ))
                }
            }
        }

        // 3. Sort by beat position
        cues.sort { $0.beatIndex < $1.beatIndex }

        // 4. Limit to maxCues, keeping most important
        cues = prioritizeAndLimit(cues)

        // 5. Calculate safe bounds
        let (safeStart, safeEnd) = calculateSafeBounds(sections: sections, beats: beats)

        return CueResult(
            cues: cues,
            safeStartBeat: safeStart,
            safeEndBeat: safeEnd
        )
    }

    private func cueTypeForSection(_ sectionType: SectionType, isStart: Bool) -> CueType {
        switch sectionType {
        case .intro:
            return isStart ? .introStart : .introEnd
        case .outro:
            return isStart ? .outroStart : .outroEnd
        case .build:
            return .build
        case .drop:
            return .drop
        case .breakdown:
            return .breakdown
        case .verse:
            return .marker
        }
    }

    private func labelForSection(_ sectionType: SectionType) -> String {
        switch sectionType {
        case .intro: return "Intro"
        case .verse: return "Verse"
        case .build: return "Build"
        case .drop: return "Drop"
        case .breakdown: return "Breakdown"
        case .outro: return "Outro"
        }
    }

    private func alignToDownbeat(_ beat: Int, beats: [BeatMarker]) -> Int {
        // Find nearest downbeat (every 4 beats typically)
        let aligned = (beat / 4) * 4
        return min(aligned, max(0, beats.count - 1))
    }

    private func hasCueNear(_ cues: [CuePoint], beat: Int, threshold: Int) -> Bool {
        return cues.contains { abs($0.beatIndex - beat) < threshold }
    }

    private func prioritizeAndLimit(_ cues: [CuePoint]) -> [CuePoint] {
        guard cues.count > maxCues else { return cues }

        // Priority order: load > drop > intro/outro > build > breakdown > marker
        let priority: [CueType: Int] = [
            .load: 0,
            .drop: 1,
            .introStart: 2,
            .outroStart: 2,
            .build: 3,
            .breakdown: 4,
            .introEnd: 5,
            .outroEnd: 5,
            .marker: 6
        ]

        let sorted = cues.sorted { (priority[$0.type] ?? 99) < (priority[$1.type] ?? 99) }
        let limited = Array(sorted.prefix(maxCues))

        // Re-sort by beat position
        return limited.sorted { $0.beatIndex < $1.beatIndex }
    }

    private func calculateSafeBounds(sections: [Section], beats: [BeatMarker]) -> (start: Int, end: Int) {
        guard !beats.isEmpty else { return (0, 0) }

        var safeStart = 0
        var safeEnd = beats.count - 1

        // Safe start: after intro ends
        if let intro = sections.first(where: { $0.type == .intro }) {
            safeStart = intro.endBeat
        }

        // Safe end: before outro starts (with some buffer)
        if let outro = sections.first(where: { $0.type == .outro }) {
            safeEnd = max(0, outro.startBeat - 32) // 32 beats buffer for mixing
        }

        return (safeStart, safeEnd)
    }
}

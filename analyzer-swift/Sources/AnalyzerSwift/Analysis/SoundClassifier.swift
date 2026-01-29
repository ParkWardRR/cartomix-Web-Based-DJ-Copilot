import Foundation

#if canImport(SoundAnalysis)
import SoundAnalysis
import AVFoundation

/// Sound event detected by Apple's SoundAnalysis framework
public struct SoundEvent: Sendable {
    /// Apple's label (e.g., "music", "speech", "crowd")
    public let label: String
    /// Mapped DJ-relevant category
    public let category: String
    /// Confidence score (0.0 - 1.0)
    public let confidence: Float
    /// Start time in seconds
    public let startTime: Double
    /// End time in seconds
    public let endTime: Double

    public init(label: String, category: String, confidence: Float, startTime: Double, endTime: Double) {
        self.label = label
        self.category = category
        self.confidence = confidence
        self.startTime = startTime
        self.endTime = endTime
    }
}

/// QA flag types for track review
public enum QAFlagType: String, Sendable {
    case needsReview = "needs_review"
    case mixedContent = "mixed_content"
    case lowConfidence = "low_confidence"
    case speechDetected = "speech_detected"
    case silenceDetected = "silence_detected"
    case noiseDetected = "noise_detected"
}

/// QA flag generated from sound classification
public struct QAFlag: Sendable {
    public let type: QAFlagType
    public let reason: String
    public let severity: Int  // 1-3, higher = more important

    public init(type: QAFlagType, reason: String, severity: Int = 1) {
        self.type = type
        self.reason = reason
        self.severity = max(1, min(3, severity))
    }
}

/// Result of sound classification analysis
public struct SoundClassificationResult: Sendable {
    /// All detected sound events
    public let events: [SoundEvent]
    /// Generated QA flags for review
    public let qaFlags: [QAFlag]
    /// Primary audio context (music/speech/noise/mixed)
    public let primaryContext: String
    /// Confidence in primary context
    public let confidence: Double
    /// Total duration analyzed
    public let duration: Double

    public init(
        events: [SoundEvent],
        qaFlags: [QAFlag],
        primaryContext: String,
        confidence: Double,
        duration: Double
    ) {
        self.events = events
        self.qaFlags = qaFlags
        self.primaryContext = primaryContext
        self.confidence = confidence
        self.duration = duration
    }

    /// Get events by category
    public func events(inCategory category: String) -> [SoundEvent] {
        return events.filter { $0.category.lowercased() == category.lowercased() }
    }

    /// Get percentage of track that is music
    public var musicPercentage: Double {
        let musicEvents = events(inCategory: "music")
        let musicDuration = musicEvents.reduce(0.0) { $0 + ($1.endTime - $1.startTime) }
        return duration > 0 ? musicDuration / duration : 0
    }
}

/// Sound classifier using Apple's built-in SoundAnalysis framework
/// Uses SNClassifySoundRequest with .version1 classifier (300+ labels)
@available(macOS 12.0, *)
public final class SoundClassifier: @unchecked Sendable {
    // Configuration
    private let sampleRate: Double
    private let windowDuration: Double
    private let overlapFactor: Double
    private let minConfidence: Float

    // Category mappings from Apple's 300+ labels to DJ-relevant categories
    private let categoryMappings: [String: String] = [
        // Music categories
        "music": "music",
        "singing": "music",
        "musical_instrument": "music",
        "guitar": "music",
        "piano": "music",
        "drum": "music",
        "bass": "music",
        "synthesizer": "music",
        "electronic_music": "music",
        "techno": "music",
        "hip_hop_music": "music",
        "rock_music": "music",
        "pop_music": "music",
        "dance_music": "music",
        "house_music": "music",

        // Speech categories
        "speech": "speech",
        "male_speech": "speech",
        "female_speech": "speech",
        "child_speech": "speech",
        "conversation": "speech",
        "narration": "speech",
        "rap": "speech",  // Could be music, but often has prominent speech
        "crowd": "speech",
        "cheer": "speech",
        "applause": "speech",
        "laughter": "speech",

        // Noise categories
        "noise": "noise",
        "white_noise": "noise",
        "pink_noise": "noise",
        "static": "noise",
        "hum": "noise",
        "buzz": "noise",
        "wind": "noise",
        "rain": "noise",
        "thunder": "noise",
        "traffic": "noise",

        // Silence/ambient
        "silence": "silence",
        "ambient": "ambient",
    ]

    public init(
        sampleRate: Double = 48000,
        windowDuration: Double = 1.0,
        overlapFactor: Double = 0.5,
        minConfidence: Float = 0.3
    ) {
        self.sampleRate = sampleRate
        self.windowDuration = windowDuration
        self.overlapFactor = overlapFactor
        self.minConfidence = minConfidence
    }

    /// Check if SoundAnalysis is available
    public var isAvailable: Bool {
        if #available(macOS 12.0, *) {
            return true
        }
        return false
    }

    /// Classify sound events from audio samples
    /// - Parameters:
    ///   - samples: Audio samples at the configured sample rate
    ///   - duration: Total duration in seconds
    /// - Returns: Classification result with events and QA flags
    public func classify(_ samples: [Float], duration: Double) async throws -> SoundClassificationResult {
        guard !samples.isEmpty else {
            return SoundClassificationResult(
                events: [],
                qaFlags: [QAFlag(type: .lowConfidence, reason: "Empty audio data")],
                primaryContext: "unknown",
                confidence: 0,
                duration: 0
            )
        }

        // Process audio in windows and collect classifications
        var allEvents: [SoundEvent] = []
        var categoryDurations: [String: Double] = [:]

        let windowSamples = Int(windowDuration * sampleRate)
        let hopSamples = Int(windowDuration * (1 - overlapFactor) * sampleRate)
        var offset = 0

        while offset + windowSamples <= samples.count {
            let windowStart = Double(offset) / sampleRate
            let windowEnd = windowStart + windowDuration
            let windowSamplesArray = Array(samples[offset..<min(offset + windowSamples, samples.count)])

            // Classify this window
            let windowClassifications = try await classifyWindow(windowSamplesArray)

            for (label, confidence) in windowClassifications where confidence >= minConfidence {
                let category = mapCategory(label)
                let event = SoundEvent(
                    label: label,
                    category: category,
                    confidence: confidence,
                    startTime: windowStart,
                    endTime: windowEnd
                )
                allEvents.append(event)

                // Track category durations
                let effectiveDuration = windowDuration * (1 - overlapFactor) // Adjust for overlap
                categoryDurations[category, default: 0] += effectiveDuration
            }

            offset += hopSamples
        }

        // Merge overlapping events of the same type
        let mergedEvents = mergeEvents(allEvents)

        // Determine primary context
        let (primaryContext, contextConfidence) = determinePrimaryContext(categoryDurations, totalDuration: duration)

        // Generate QA flags
        let qaFlags = generateQAFlags(
            events: mergedEvents,
            categoryDurations: categoryDurations,
            primaryContext: primaryContext,
            contextConfidence: contextConfidence,
            totalDuration: duration
        )

        return SoundClassificationResult(
            events: mergedEvents,
            qaFlags: qaFlags,
            primaryContext: primaryContext,
            confidence: contextConfidence,
            duration: duration
        )
    }

    /// Classify a single window of audio using SoundAnalysis
    private func classifyWindow(_ samples: [Float]) async throws -> [(String, Float)] {
        // Convert Float samples to AVAudioPCMBuffer
        let format = AVAudioFormat(standardFormatWithSampleRate: sampleRate, channels: 1)!
        guard let buffer = AVAudioPCMBuffer(pcmFormat: format, frameCapacity: AVAudioFrameCount(samples.count)) else {
            throw SoundClassifierError.bufferCreationFailed
        }
        buffer.frameLength = AVAudioFrameCount(samples.count)

        if let channelData = buffer.floatChannelData?[0] {
            for (index, sample) in samples.enumerated() {
                channelData[index] = sample
            }
        }

        // Create analyzer
        let analyzer = SNAudioStreamAnalyzer(format: format)

        // Create classification request with built-in classifier
        let request = try SNClassifySoundRequest(classifierIdentifier: .version1)
        request.windowDuration = CMTime(seconds: windowDuration, preferredTimescale: 48000)
        request.overlapFactor = overlapFactor

        // Create result observer
        let observer = ClassificationObserver()
        try analyzer.add(request, withObserver: observer)

        // Analyze the buffer
        try analyzer.analyze(buffer, atAudioFramePosition: 0)

        // Wait for results (with timeout)
        let timeout: UInt64 = 5_000_000_000 // 5 seconds in nanoseconds
        let startTime = DispatchTime.now()

        while !observer.hasResults {
            if DispatchTime.now().uptimeNanoseconds - startTime.uptimeNanoseconds > timeout {
                throw SoundClassifierError.analysisTimeout
            }
            try await Task.sleep(nanoseconds: 10_000_000) // 10ms
        }

        analyzer.removeAllRequests()

        return observer.classifications
    }

    /// Map Apple's label to DJ-relevant category
    private func mapCategory(_ label: String) -> String {
        // First check direct mapping
        if let category = categoryMappings[label.lowercased()] {
            return category
        }

        // Check for partial matches
        let lowercased = label.lowercased()
        if lowercased.contains("music") || lowercased.contains("instrument") ||
           lowercased.contains("drum") || lowercased.contains("guitar") ||
           lowercased.contains("bass") || lowercased.contains("synth") {
            return "music"
        }
        if lowercased.contains("speech") || lowercased.contains("voice") ||
           lowercased.contains("talk") || lowercased.contains("speak") {
            return "speech"
        }
        if lowercased.contains("noise") || lowercased.contains("static") ||
           lowercased.contains("hum") || lowercased.contains("buzz") {
            return "noise"
        }

        return "other"
    }

    /// Merge overlapping events of the same category
    private func mergeEvents(_ events: [SoundEvent]) -> [SoundEvent] {
        guard !events.isEmpty else { return [] }

        // Group by category
        var categoryEvents: [String: [SoundEvent]] = [:]
        for event in events {
            categoryEvents[event.category, default: []].append(event)
        }

        var mergedEvents: [SoundEvent] = []

        for (_, categoryEvts) in categoryEvents {
            let sorted = categoryEvts.sorted { $0.startTime < $1.startTime }
            var merged: [SoundEvent] = []
            var current: SoundEvent? = nil

            for event in sorted {
                if var curr = current {
                    // Check for overlap or adjacency (within 0.1s)
                    if event.startTime <= curr.endTime + 0.1 {
                        // Merge: extend end time and average confidence
                        let avgConfidence = (curr.confidence + event.confidence) / 2
                        curr = SoundEvent(
                            label: curr.label,
                            category: curr.category,
                            confidence: avgConfidence,
                            startTime: curr.startTime,
                            endTime: max(curr.endTime, event.endTime)
                        )
                        current = curr
                    } else {
                        merged.append(curr)
                        current = event
                    }
                } else {
                    current = event
                }
            }

            if let curr = current {
                merged.append(curr)
            }

            mergedEvents.append(contentsOf: merged)
        }

        return mergedEvents.sorted { $0.startTime < $1.startTime }
    }

    /// Determine primary audio context
    private func determinePrimaryContext(
        _ categoryDurations: [String: Double],
        totalDuration: Double
    ) -> (String, Double) {
        guard !categoryDurations.isEmpty, totalDuration > 0 else {
            return ("unknown", 0)
        }

        // Find dominant category
        let sorted = categoryDurations.sorted { $0.value > $1.value }

        if let (topCategory, topDuration) = sorted.first {
            let percentage = topDuration / totalDuration

            // Check if it's mixed content
            if sorted.count >= 2 {
                let secondDuration = sorted[1].value
                let secondPercentage = secondDuration / totalDuration

                // If second category is significant (>20%), it's mixed
                if secondPercentage > 0.2 && percentage < 0.7 {
                    return ("mixed", percentage)
                }
            }

            return (topCategory, percentage)
        }

        return ("unknown", 0)
    }

    /// Generate QA flags based on classification results
    private func generateQAFlags(
        events: [SoundEvent],
        categoryDurations: [String: Double],
        primaryContext: String,
        contextConfidence: Double,
        totalDuration: Double
    ) -> [QAFlag] {
        var flags: [QAFlag] = []

        // Check for mixed content
        if primaryContext == "mixed" {
            flags.append(QAFlag(
                type: .mixedContent,
                reason: "Track contains multiple audio types (music, speech, etc.)",
                severity: 2
            ))
        }

        // Check for speech in music tracks
        let speechDuration = categoryDurations["speech"] ?? 0
        let speechPercentage = totalDuration > 0 ? speechDuration / totalDuration : 0
        if speechPercentage > 0.1 && primaryContext == "music" {
            flags.append(QAFlag(
                type: .speechDetected,
                reason: String(format: "Speech detected in %.0f%% of track", speechPercentage * 100),
                severity: speechPercentage > 0.3 ? 2 : 1
            ))
        }

        // Check for noise
        let noiseDuration = categoryDurations["noise"] ?? 0
        let noisePercentage = totalDuration > 0 ? noiseDuration / totalDuration : 0
        if noisePercentage > 0.05 {
            flags.append(QAFlag(
                type: .noiseDetected,
                reason: String(format: "Background noise detected in %.0f%% of track", noisePercentage * 100),
                severity: noisePercentage > 0.15 ? 2 : 1
            ))
        }

        // Check for silence
        let silenceDuration = categoryDurations["silence"] ?? 0
        let silencePercentage = totalDuration > 0 ? silenceDuration / totalDuration : 0
        if silencePercentage > 0.1 {
            flags.append(QAFlag(
                type: .silenceDetected,
                reason: String(format: "Silence detected in %.0f%% of track", silencePercentage * 100),
                severity: silencePercentage > 0.3 ? 2 : 1
            ))
        }

        // Check for low confidence
        if contextConfidence < 0.5 {
            flags.append(QAFlag(
                type: .lowConfidence,
                reason: String(format: "Low classification confidence (%.0f%%)", contextConfidence * 100),
                severity: contextConfidence < 0.3 ? 2 : 1
            ))
        }

        // Set needs_review if any high-severity flags
        if flags.contains(where: { $0.severity >= 2 }) {
            flags.insert(QAFlag(
                type: .needsReview,
                reason: "Track flagged for manual review",
                severity: 2
            ), at: 0)
        }

        return flags
    }
}

/// Error types for sound classification
public enum SoundClassifierError: Error, CustomStringConvertible {
    case notAvailable
    case bufferCreationFailed
    case analysisTimeout
    case analysisFailed(String)

    public var description: String {
        switch self {
        case .notAvailable:
            return "SoundAnalysis not available on this platform"
        case .bufferCreationFailed:
            return "Failed to create audio buffer"
        case .analysisTimeout:
            return "Sound analysis timed out"
        case .analysisFailed(let msg):
            return "Sound analysis failed: \(msg)"
        }
    }
}

/// Observer for SoundAnalysis results
@available(macOS 12.0, *)
private final class ClassificationObserver: NSObject, SNResultsObserving {
    var classifications: [(String, Float)] = []
    var hasResults = false
    var error: Error?

    func request(_ request: SNRequest, didProduce result: SNResult) {
        guard let classificationResult = result as? SNClassificationResult else { return }

        // Get top classifications
        let topClassifications = classificationResult.classifications
            .filter { $0.confidence > 0.1 }
            .prefix(10)
            .map { ($0.identifier, Float($0.confidence)) }

        classifications.append(contentsOf: topClassifications)
    }

    func request(_ request: SNRequest, didFailWithError error: Error) {
        self.error = error
        hasResults = true
    }

    func requestDidComplete(_ request: SNRequest) {
        hasResults = true
    }
}

#else

// Fallback for platforms without SoundAnalysis
public struct SoundEvent: Sendable {
    public let label: String
    public let category: String
    public let confidence: Float
    public let startTime: Double
    public let endTime: Double
}

public enum QAFlagType: String, Sendable {
    case needsReview = "needs_review"
    case mixedContent = "mixed_content"
    case lowConfidence = "low_confidence"
    case speechDetected = "speech_detected"
    case silenceDetected = "silence_detected"
    case noiseDetected = "noise_detected"
}

public struct QAFlag: Sendable {
    public let type: QAFlagType
    public let reason: String
    public let severity: Int
}

public struct SoundClassificationResult: Sendable {
    public let events: [SoundEvent]
    public let qaFlags: [QAFlag]
    public let primaryContext: String
    public let confidence: Double
    public let duration: Double
}

public final class SoundClassifier: @unchecked Sendable {
    public var isAvailable: Bool { false }

    public init(
        sampleRate: Double = 48000,
        windowDuration: Double = 1.0,
        overlapFactor: Double = 0.5,
        minConfidence: Float = 0.3
    ) {}

    public func classify(_ samples: [Float], duration: Double) async throws -> SoundClassificationResult {
        return SoundClassificationResult(
            events: [],
            qaFlags: [QAFlag(type: .lowConfidence, reason: "SoundAnalysis not available", severity: 1)],
            primaryContext: "unknown",
            confidence: 0,
            duration: duration
        )
    }
}

public enum SoundClassifierError: Error {
    case notAvailable
}

#endif

import Foundation
import CoreML

#if canImport(SoundAnalysis)
import SoundAnalysis
import AVFoundation

/// Section prediction result
public struct SectionPrediction: Sendable {
    public let label: DJSectionLabel
    public let confidence: Float
    public let startTime: Double
    public let endTime: Double
    public let startBeat: Int
    public let endBeat: Int

    public init(
        label: DJSectionLabel,
        confidence: Float,
        startTime: Double,
        endTime: Double,
        startBeat: Int,
        endBeat: Int
    ) {
        self.label = label
        self.confidence = confidence
        self.startTime = startTime
        self.endTime = endTime
        self.startBeat = startBeat
        self.endBeat = endBeat
    }
}

/// Classification result for a track
public struct DJSectionClassificationResult: Sendable {
    public let sections: [SectionPrediction]
    public let modelVersion: Int
    public let averageConfidence: Double

    public init(sections: [SectionPrediction], modelVersion: Int, averageConfidence: Double) {
        self.sections = sections
        self.modelVersion = modelVersion
        self.averageConfidence = averageConfidence
    }

    /// Get sections by label
    public func sections(ofType label: DJSectionLabel) -> [SectionPrediction] {
        return sections.filter { $0.label == label }
    }

    /// Get primary section at a given time
    public func section(at time: Double) -> SectionPrediction? {
        return sections.first { time >= $0.startTime && time < $0.endTime }
    }
}

/// DJ Section Classifier using custom trained Core ML model
@available(macOS 12.0, *)
public final class DJSectionClassifier: @unchecked Sendable {
    private let sampleRate: Double
    private let windowDuration: Double
    private let windowHop: Double
    private let minConfidence: Float

    private var model: MLModel?
    private var modelVersion: Int = 0
    private let trainer: DJSectionTrainer

    public init(
        sampleRate: Double = 48000,
        windowDuration: Double = 4.0,  // 4-second windows for section detection
        windowHop: Double = 2.0,       // 50% overlap
        minConfidence: Float = 0.5
    ) {
        self.sampleRate = sampleRate
        self.windowDuration = windowDuration
        self.windowHop = windowHop
        self.minConfidence = minConfidence
        self.trainer = DJSectionTrainer()

        // Try to load active model
        loadActiveModel()
    }

    /// Check if a trained model is available
    public var isAvailable: Bool {
        return model != nil
    }

    /// Get current model version
    public var currentModelVersion: Int {
        return modelVersion
    }

    /// Reload the active model (after training a new one)
    public func reloadModel() {
        loadActiveModel()
    }

    /// Classify sections in audio samples
    /// - Parameters:
    ///   - samples: Audio samples at the configured sample rate
    ///   - beats: Beat markers for beat-aligned section boundaries
    ///   - tempo: Track tempo in BPM
    /// - Returns: Classification result with predicted sections
    public func classify(
        _ samples: [Float],
        beats: [BeatMarker],
        tempo: Double
    ) async throws -> DJSectionClassificationResult {
        guard let model = model else {
            throw DJSectionClassifierError.modelNotLoaded
        }

        guard !samples.isEmpty else {
            return DJSectionClassificationResult(sections: [], modelVersion: modelVersion, averageConfidence: 0)
        }

        let duration = Double(samples.count) / sampleRate
        var predictions: [(time: Double, label: DJSectionLabel, confidence: Float)] = []

        // Process audio in windows
        let windowSamples = Int(windowDuration * sampleRate)
        let hopSamples = Int(windowHop * sampleRate)
        var offset = 0

        while offset + windowSamples <= samples.count {
            let windowStart = Double(offset) / sampleRate
            let windowSamplesArray = Array(samples[offset..<min(offset + windowSamples, samples.count)])

            // Run inference on window
            if let (label, confidence) = try await classifyWindow(windowSamplesArray, model: model) {
                if confidence >= minConfidence {
                    predictions.append((windowStart + windowDuration / 2, label, confidence))
                }
            }

            offset += hopSamples
        }

        // Merge predictions into sections
        let sections = mergeIntoSections(predictions, beats: beats, duration: duration)

        let avgConfidence = sections.isEmpty ? 0 : Double(sections.reduce(0) { $0 + $1.confidence }) / Double(sections.count)

        return DJSectionClassificationResult(
            sections: sections,
            modelVersion: modelVersion,
            averageConfidence: avgConfidence
        )
    }

    // MARK: - Private Methods

    private func loadActiveModel() {
        do {
            if let modelURL = try trainer.getActiveModel() {
                let config = MLModelConfiguration()
                config.computeUnits = .all  // Use ANE when available

                model = try MLModel(contentsOf: modelURL, configuration: config)

                // Extract version from path
                let filename = modelURL.lastPathComponent
                let versionStr = filename
                    .replacingOccurrences(of: "dj_section_v", with: "")
                    .replacingOccurrences(of: ".mlmodelc", with: "")
                if let version = Int(versionStr) {
                    modelVersion = version
                }

                print("Loaded DJ Section model v\(modelVersion)")
            }
        } catch {
            print("Failed to load DJ Section model: \(error)")
            model = nil
        }
    }

    private func classifyWindow(_ samples: [Float], model: MLModel) async throws -> (DJSectionLabel, Float)? {
        // Create audio buffer
        let format = AVAudioFormat(standardFormatWithSampleRate: sampleRate, channels: 1)!
        guard let buffer = AVAudioPCMBuffer(pcmFormat: format, frameCapacity: AVAudioFrameCount(samples.count)) else {
            return nil
        }
        buffer.frameLength = AVAudioFrameCount(samples.count)

        if let channelData = buffer.floatChannelData?[0] {
            for (index, sample) in samples.enumerated() {
                channelData[index] = sample
            }
        }

        // Use SoundAnalysis with custom model
        let analyzer = SNAudioStreamAnalyzer(format: format)

        let request = try SNClassifySoundRequest(mlModel: model)
        request.windowDuration = CMTime(seconds: windowDuration, preferredTimescale: 48000)

        let observer = SectionClassificationObserver()
        try analyzer.add(request, withObserver: observer)

        analyzer.analyze(buffer, atAudioFramePosition: 0)

        // Wait for results
        let timeout: UInt64 = 3_000_000_000 // 3 seconds
        let startTime = DispatchTime.now()

        while !observer.hasResult {
            if DispatchTime.now().uptimeNanoseconds - startTime.uptimeNanoseconds > timeout {
                break
            }
            try await Task.sleep(nanoseconds: 10_000_000) // 10ms
        }

        analyzer.removeAllRequests()

        if let (labelStr, confidence) = observer.topClassification,
           let label = DJSectionLabel(rawValue: labelStr) {
            return (label, confidence)
        }

        return nil
    }

    private func mergeIntoSections(
        _ predictions: [(time: Double, label: DJSectionLabel, confidence: Float)],
        beats: [BeatMarker],
        duration: Double
    ) -> [SectionPrediction] {
        guard !predictions.isEmpty else { return [] }

        var sections: [SectionPrediction] = []
        var currentLabel = predictions[0].label
        var currentConfidences: [Float] = [predictions[0].confidence]
        var sectionStartTime = 0.0

        for i in 1..<predictions.count {
            let pred = predictions[i]

            if pred.label != currentLabel {
                // End current section
                let avgConfidence = currentConfidences.reduce(0, +) / Float(currentConfidences.count)
                let sectionEndTime = pred.time - windowHop / 2

                // Find beat boundaries
                let (startBeat, endBeat) = findBeatBoundaries(
                    startTime: sectionStartTime,
                    endTime: sectionEndTime,
                    beats: beats
                )

                sections.append(SectionPrediction(
                    label: currentLabel,
                    confidence: avgConfidence,
                    startTime: sectionStartTime,
                    endTime: sectionEndTime,
                    startBeat: startBeat,
                    endBeat: endBeat
                ))

                // Start new section
                currentLabel = pred.label
                currentConfidences = [pred.confidence]
                sectionStartTime = sectionEndTime
            } else {
                currentConfidences.append(pred.confidence)
            }
        }

        // Add final section
        let avgConfidence = currentConfidences.reduce(0, +) / Float(currentConfidences.count)
        let (startBeat, endBeat) = findBeatBoundaries(
            startTime: sectionStartTime,
            endTime: duration,
            beats: beats
        )

        sections.append(SectionPrediction(
            label: currentLabel,
            confidence: avgConfidence,
            startTime: sectionStartTime,
            endTime: duration,
            startBeat: startBeat,
            endBeat: endBeat
        ))

        // Filter out very short sections (< 4 beats)
        return sections.filter { $0.endBeat - $0.startBeat >= 4 }
    }

    private func findBeatBoundaries(startTime: Double, endTime: Double, beats: [BeatMarker]) -> (Int, Int) {
        var startBeat = 0
        var endBeat = beats.count - 1

        for (_, beat) in beats.enumerated() {
            if beat.time >= startTime && startBeat == 0 {
                startBeat = beat.index
            }
            if beat.time >= endTime {
                endBeat = beat.index
                break
            }
        }

        return (startBeat, endBeat)
    }
}

/// Observer for section classification results
@available(macOS 12.0, *)
private final class SectionClassificationObserver: NSObject, SNResultsObserving {
    var topClassification: (String, Float)?
    var hasResult = false
    var error: Error?

    func request(_ request: SNRequest, didProduce result: SNResult) {
        guard let classificationResult = result as? SNClassificationResult else { return }

        if let top = classificationResult.classifications.first {
            topClassification = (top.identifier, Float(top.confidence))
        }
    }

    func request(_ request: SNRequest, didFailWithError error: Error) {
        self.error = error
        hasResult = true
    }

    func requestDidComplete(_ request: SNRequest) {
        hasResult = true
    }
}

/// Errors for DJ Section Classifier
public enum DJSectionClassifierError: Error, CustomStringConvertible {
    case modelNotLoaded
    case classificationFailed(String)

    public var description: String {
        switch self {
        case .modelNotLoaded:
            return "No trained DJ Section model is loaded"
        case .classificationFailed(let msg):
            return "Classification failed: \(msg)"
        }
    }
}

#else

// Fallback for platforms without SoundAnalysis
public struct SectionPrediction: Sendable {
    public let label: DJSectionLabel
    public let confidence: Float
    public let startTime: Double
    public let endTime: Double
    public let startBeat: Int
    public let endBeat: Int
}

public struct DJSectionClassificationResult: Sendable {
    public let sections: [SectionPrediction]
    public let modelVersion: Int
    public let averageConfidence: Double
}

public final class DJSectionClassifier: @unchecked Sendable {
    public var isAvailable: Bool { false }
    public var currentModelVersion: Int { 0 }

    public init(
        sampleRate: Double = 48000,
        windowDuration: Double = 4.0,
        windowHop: Double = 2.0,
        minConfidence: Float = 0.5
    ) {}

    public func reloadModel() {}

    public func classify(
        _ samples: [Float],
        beats: [BeatMarker],
        tempo: Double
    ) async throws -> DJSectionClassificationResult {
        throw DJSectionClassifierError.modelNotLoaded
    }
}

public enum DJSectionClassifierError: Error {
    case modelNotLoaded
}

#endif

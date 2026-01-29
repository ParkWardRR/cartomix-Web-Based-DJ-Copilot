import Foundation

/// Complete track analysis result
public struct TrackAnalysisResult: Sendable {
    public let path: String
    public let duration: Double
    public let sampleRate: Double

    // Beatgrid
    public let beats: [BeatMarker]
    public let tempoMap: [TempoNode]
    public let beatgridConfidence: Double

    // Key
    public let key: MusicalKey

    // Energy
    public let globalEnergy: Int
    public let energyCurve: [Float]
    public let lowEnergy: Float
    public let midEnergy: Float
    public let highEnergy: Float
    public let rms: Float
    public let peak: Float

    // Sections
    public let sections: [Section]
    public let sectionConfidence: Double
    public let transitionWindows: [(start: Double, end: Double)]

    // Cues
    public let cues: [CuePoint]
    public let safeStartBeat: Int
    public let safeEndBeat: Int

    // Waveform summary (for visualization)
    public let waveformSummary: [Float]

    public init(
        path: String,
        duration: Double,
        sampleRate: Double,
        beats: [BeatMarker],
        tempoMap: [TempoNode],
        beatgridConfidence: Double,
        key: MusicalKey,
        globalEnergy: Int,
        energyCurve: [Float],
        lowEnergy: Float,
        midEnergy: Float,
        highEnergy: Float,
        rms: Float,
        peak: Float,
        sections: [Section],
        sectionConfidence: Double,
        transitionWindows: [(start: Double, end: Double)],
        cues: [CuePoint],
        safeStartBeat: Int,
        safeEndBeat: Int,
        waveformSummary: [Float]
    ) {
        self.path = path
        self.duration = duration
        self.sampleRate = sampleRate
        self.beats = beats
        self.tempoMap = tempoMap
        self.beatgridConfidence = beatgridConfidence
        self.key = key
        self.globalEnergy = globalEnergy
        self.energyCurve = energyCurve
        self.lowEnergy = lowEnergy
        self.midEnergy = midEnergy
        self.highEnergy = highEnergy
        self.rms = rms
        self.peak = peak
        self.sections = sections
        self.sectionConfidence = sectionConfidence
        self.transitionWindows = transitionWindows
        self.cues = cues
        self.safeStartBeat = safeStartBeat
        self.safeEndBeat = safeEndBeat
        self.waveformSummary = waveformSummary
    }

    /// Primary BPM (from first tempo node)
    public var bpm: Double {
        tempoMap.first?.bpm ?? 120.0
    }

    /// Key in Camelot notation
    public var camelotKey: String {
        key.camelot
    }
}

/// Analysis error types
public enum AnalyzerError: Error, Sendable {
    case decodingFailed(String)
    case insufficientData
    case analysisTimeout
    case cancelled
}

/// Analysis progress callback
public typealias AnalysisProgressHandler = @Sendable (AnalysisProgress) -> Void

/// Analysis progress stages
public enum AnalysisProgress: Sendable {
    case decoding
    case beatgrid(progress: Double)
    case key
    case energy
    case sections
    case cues
    case waveform
    case complete
}

/// Main analyzer that orchestrates all analysis stages
public final class Analyzer: @unchecked Sendable {
    private let decoder: AudioDecoder
    private let beatgridDetector: BeatgridDetector
    private let keyDetector: KeyDetector
    private let energyAnalyzer: EnergyAnalyzer
    private let sectionDetector: SectionDetector
    private let cueGenerator: CueGenerator

    // Configuration
    private let targetSampleRate: Double
    private let waveformBins: Int

    public init(
        sampleRate: Double = 48000,
        waveformBins: Int = 200
    ) {
        self.targetSampleRate = sampleRate
        self.waveformBins = waveformBins

        // Initialize components
        self.decoder = AudioDecoder(targetSampleRate: sampleRate, mono: true)
        self.beatgridDetector = BeatgridDetector(sampleRate: sampleRate)
        self.keyDetector = KeyDetector(sampleRate: sampleRate)
        self.energyAnalyzer = EnergyAnalyzer(sampleRate: sampleRate)
        self.sectionDetector = SectionDetector(sampleRate: sampleRate)
        self.cueGenerator = CueGenerator(maxCues: 8)
    }

    /// Analyze a track file
    public func analyze(
        path: String,
        progress: AnalysisProgressHandler? = nil
    ) async throws -> TrackAnalysisResult {
        // 1. Decode audio
        progress?(.decoding)
        let audio: AudioBuffer
        do {
            audio = try decoder.decode(path: path)
        } catch {
            throw AnalyzerError.decodingFailed(error.localizedDescription)
        }

        let samples = audio.monoSamples()

        guard samples.count > Int(targetSampleRate) else {
            throw AnalyzerError.insufficientData
        }

        // 2. Detect beatgrid
        progress?(.beatgrid(progress: 0))
        let beatgridResult = beatgridDetector.detect(samples)
        progress?(.beatgrid(progress: 1))

        // 3. Detect key
        progress?(.key)
        let key = keyDetector.detect(samples)

        // 4. Analyze energy
        progress?(.energy)
        let energyResult = energyAnalyzer.analyze(samples)

        // 5. Detect sections
        progress?(.sections)
        let sectionResult = sectionDetector.detect(
            samples,
            beats: beatgridResult.beats,
            tempo: beatgridResult.tempoMap.first?.bpm ?? 120
        )

        // 6. Generate cues
        progress?(.cues)
        let cueResult = cueGenerator.generate(
            sections: sectionResult.sections,
            beats: beatgridResult.beats
        )

        // 7. Generate waveform summary
        progress?(.waveform)
        let waveformSummary = generateWaveformSummary(samples)

        progress?(.complete)

        return TrackAnalysisResult(
            path: path,
            duration: audio.duration,
            sampleRate: audio.sampleRate,
            beats: beatgridResult.beats,
            tempoMap: beatgridResult.tempoMap,
            beatgridConfidence: beatgridResult.confidence,
            key: key,
            globalEnergy: energyResult.globalEnergy,
            energyCurve: energyResult.curve,
            lowEnergy: energyResult.lowEnergy,
            midEnergy: energyResult.midEnergy,
            highEnergy: energyResult.highEnergy,
            rms: energyResult.rms,
            peak: energyResult.peak,
            sections: sectionResult.sections,
            sectionConfidence: sectionResult.confidence,
            transitionWindows: sectionResult.transitionWindows,
            cues: cueResult.cues,
            safeStartBeat: cueResult.safeStartBeat,
            safeEndBeat: cueResult.safeEndBeat,
            waveformSummary: waveformSummary
        )
    }

    /// Generate a downsampled waveform for visualization
    private func generateWaveformSummary(_ samples: [Float]) -> [Float] {
        guard !samples.isEmpty else { return [] }

        let binSize = samples.count / waveformBins
        guard binSize > 0 else { return samples }

        var summary = [Float](repeating: 0, count: waveformBins)

        for bin in 0..<waveformBins {
            let start = bin * binSize
            let end = min(start + binSize, samples.count)

            var maxVal: Float = 0
            for i in start..<end {
                let absVal = abs(samples[i])
                if absVal > maxVal {
                    maxVal = absVal
                }
            }
            summary[bin] = maxVal
        }

        return summary
    }
}

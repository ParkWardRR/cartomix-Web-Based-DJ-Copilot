import XCTest
@testable import AnalyzerSwift

// MARK: - FFT Tests

final class FFTTests: XCTestCase {
    func testFFTProcessorCreation() {
        let fft = FFTProcessor(fftSize: 2048)
        XCTAssertEqual(fft.fftSize, 2048)
    }

    func testFFTProcessorPowerOfTwo() {
        _ = FFTProcessor(fftSize: 1024)
        _ = FFTProcessor(fftSize: 4096)
    }

    func testMagnitudeSpectrumOutput() {
        let fft = FFTProcessor(fftSize: 1024)

        // Generate a simple sine wave at 440 Hz
        let sampleRate: Double = 48000
        let frequency: Double = 440
        var samples = [Float](repeating: 0, count: 1024)

        for i in 0..<1024 {
            samples[i] = Float(sin(2 * .pi * frequency * Double(i) / sampleRate))
        }

        let spectrum = fft.magnitudeSpectrum(samples)

        XCTAssertEqual(spectrum.count, 512) // Half of FFT size
    }

    func testSTFTOutput() {
        let fft = FFTProcessor(fftSize: 512)
        let samples = [Float](repeating: 0.5, count: 2048)

        let spectrogram = fft.stft(samples, hopSize: 256)

        XCTAssertGreaterThan(spectrogram.count, 0)
        XCTAssertEqual(spectrogram[0].count, 256) // Half of FFT size
    }

    func testSpectralFlux() {
        let fft = FFTProcessor(fftSize: 512)

        let spectrogram: [[Float]] = [
            [Float](repeating: 0.1, count: 256),
            [Float](repeating: 0.2, count: 256),
            [Float](repeating: 0.5, count: 256),
            [Float](repeating: 0.3, count: 256)
        ]

        let flux = fft.spectralFlux(spectrogram)

        XCTAssertEqual(flux.count, 4)
        XCTAssertEqual(flux[0], 0) // First frame has no flux
        XCTAssertGreaterThan(flux[2], flux[1]) // Larger jump
    }

    func testChromaFeatures() {
        let fft = FFTProcessor(fftSize: 2048)

        var samples = [Float](repeating: 0, count: 8192)
        for i in 0..<8192 {
            samples[i] = Float(sin(2 * .pi * 440 * Double(i) / 48000))
        }

        let chroma = fft.chromaFeatures(samples, sampleRate: 48000, hopSize: 1024)

        XCTAssertGreaterThan(chroma.count, 0)
        XCTAssertEqual(chroma[0].count, 12) // 12 pitch classes
    }
}

// MARK: - Beatgrid Detector Tests

final class BeatgridDetectorTests: XCTestCase {
    func testBeatgridDetectorCreation() {
        let detector = BeatgridDetector(sampleRate: 48000)
        XCTAssertNotNil(detector)
    }

    func testBeatgridDetectionWithSilence() {
        let detector = BeatgridDetector(sampleRate: 48000, fftSize: 1024, hopSize: 256)
        let samples = [Float](repeating: 0, count: 48000 * 10) // 10 seconds of silence

        let result = detector.detect(samples)

        XCTAssertNotNil(result.tempoMap.first)
    }

    func testBeatgridTempoRange() {
        let detector = BeatgridDetector(
            sampleRate: 48000,
            tempoFloor: 100,
            tempoCeil: 150
        )

        // Generate click track at ~120 BPM
        var samples = [Float](repeating: 0, count: 48000 * 10)
        let beatInterval = Int(48000 * 60.0 / 120.0)

        for beat in 0..<20 {
            let pos = beat * beatInterval
            if pos < samples.count - 100 {
                for i in 0..<100 {
                    samples[pos + i] = Float.random(in: -1...1)
                }
            }
        }

        let result = detector.detect(samples)

        if let tempo = result.tempoMap.first?.bpm {
            XCTAssertGreaterThanOrEqual(tempo, 100)
            XCTAssertLessThanOrEqual(tempo, 150)
        }
    }

    func testBeatMarkerMonotonicity() {
        let detector = BeatgridDetector(sampleRate: 48000)

        var samples = [Float](repeating: 0, count: 48000 * 5)
        for i in 0..<samples.count {
            samples[i] = Float.random(in: -0.5...0.5)
        }

        let result = detector.detect(samples)

        for i in 1..<result.beats.count {
            XCTAssertGreaterThan(result.beats[i].time, result.beats[i-1].time, "Beats should be monotonically increasing")
            XCTAssertGreaterThan(result.beats[i].index, result.beats[i-1].index, "Beat indices should be monotonically increasing")
        }
    }
}

// MARK: - Key Detector Tests

final class KeyDetectorTests: XCTestCase {
    func testKeyDetectorCreation() {
        let detector = KeyDetector(sampleRate: 48000)
        XCTAssertNotNil(detector)
    }

    func testKeyDetectionReturnsValidKey() {
        let detector = KeyDetector(sampleRate: 48000, fftSize: 2048, hopSize: 1024)

        // Generate a simple A minor chord (A, C, E)
        let frequencies: [Double] = [440, 523.25, 659.25]
        var samples = [Float](repeating: 0, count: 48000 * 3)

        for i in 0..<samples.count {
            var sum: Float = 0
            for freq in frequencies {
                sum += Float(sin(2 * .pi * freq * Double(i) / 48000))
            }
            samples[i] = sum / 3
        }

        let key = detector.detect(samples)

        XCTAssertGreaterThanOrEqual(key.pitchClass, 0)
        XCTAssertLessThan(key.pitchClass, 12)
        XCTAssertGreaterThanOrEqual(key.confidence, 0)
        XCTAssertLessThanOrEqual(key.confidence, 1)
    }

    func testMusicalKeyNameFormat() {
        let majorKey = MusicalKey(pitchClass: 0, isMinor: false, confidence: 0.9)
        XCTAssertEqual(majorKey.name, "C")

        let minorKey = MusicalKey(pitchClass: 9, isMinor: true, confidence: 0.9)
        XCTAssertEqual(minorKey.name, "Am")
    }

    func testCamelotNotation() {
        let cMajor = MusicalKey(pitchClass: 0, isMinor: false, confidence: 1.0)
        XCTAssertEqual(cMajor.camelot, "8B")

        let aMinor = MusicalKey(pitchClass: 9, isMinor: true, confidence: 1.0)
        XCTAssertEqual(aMinor.camelot, "8A")

        let fMajor = MusicalKey(pitchClass: 5, isMinor: false, confidence: 1.0)
        XCTAssertEqual(fMajor.camelot, "7B")
    }

    func testOpenKeyNotation() {
        let key = MusicalKey(pitchClass: 0, isMinor: false, confidence: 1.0)
        XCTAssertEqual(key.openKey, "1d")

        let minorKey = MusicalKey(pitchClass: 0, isMinor: true, confidence: 1.0)
        XCTAssertEqual(minorKey.openKey, "1m")
    }
}

// MARK: - Energy Analyzer Tests

final class EnergyAnalyzerTests: XCTestCase {
    func testEnergyAnalyzerCreation() {
        let analyzer = EnergyAnalyzer(sampleRate: 48000)
        XCTAssertNotNil(analyzer)
    }

    func testEnergyAnalysisReturnsValidRange() {
        let analyzer = EnergyAnalyzer(sampleRate: 48000, fftSize: 1024, hopSize: 512)

        var samples = [Float](repeating: 0, count: 48000 * 3)
        for i in 0..<samples.count {
            samples[i] = Float.random(in: -0.5...0.5)
        }

        let result = analyzer.analyze(samples)

        XCTAssertGreaterThanOrEqual(result.globalEnergy, 1)
        XCTAssertLessThanOrEqual(result.globalEnergy, 10)
        XCTAssertGreaterThanOrEqual(result.rms, 0)
        XCTAssertGreaterThanOrEqual(result.peak, 0)
    }

    func testEnergyBandDistribution() {
        let analyzer = EnergyAnalyzer(sampleRate: 48000)

        var samples = [Float](repeating: 0, count: 48000 * 2)
        for i in 0..<samples.count {
            samples[i] = Float.random(in: -0.3...0.3)
        }

        let result = analyzer.analyze(samples)

        let sum = result.lowEnergy + result.midEnergy + result.highEnergy
        XCTAssertLessThan(abs(sum - 1.0), 0.01)
    }

    func testEnergyCurveNormalization() {
        let analyzer = EnergyAnalyzer(sampleRate: 48000, hopSize: 1024)

        var samples = [Float](repeating: 0, count: 48000 * 2)
        for i in 0..<samples.count {
            samples[i] = Float.random(in: -0.5...0.5)
        }

        let result = analyzer.analyze(samples)

        for value in result.curve {
            XCTAssertGreaterThanOrEqual(value, 0)
            XCTAssertLessThanOrEqual(value, 1)
        }
    }
}

// MARK: - Section Detector Tests

final class SectionDetectorTests: XCTestCase {
    func testSectionDetectorCreation() {
        let detector = SectionDetector(sampleRate: 48000)
        XCTAssertNotNil(detector)
    }

    func testSectionDetectionWithEmptyBeats() {
        let detector = SectionDetector(sampleRate: 48000)
        let samples = [Float](repeating: 0, count: 48000)

        let result = detector.detect(samples, beats: [], tempo: 120)

        XCTAssertTrue(result.sections.isEmpty)
    }

    func testSectionTypes() {
        let allTypes = SectionType.allCases
        XCTAssertEqual(allTypes.count, 6)
        XCTAssertTrue(allTypes.contains(.intro))
        XCTAssertTrue(allTypes.contains(.verse))
        XCTAssertTrue(allTypes.contains(.build))
        XCTAssertTrue(allTypes.contains(.drop))
        XCTAssertTrue(allTypes.contains(.breakdown))
        XCTAssertTrue(allTypes.contains(.outro))
    }
}

// MARK: - Cue Generator Tests

final class CueGeneratorTests: XCTestCase {
    func testCueGeneratorCreation() {
        let generator = CueGenerator(maxCues: 8)
        XCTAssertNotNil(generator)
    }

    func testCueGenerationWithEmptyInput() {
        let generator = CueGenerator()
        let result = generator.generate(sections: [], beats: [])

        XCTAssertTrue(result.cues.isEmpty)
    }

    func testCueGenerationAlwaysIncludesLoadPoint() {
        let generator = CueGenerator()

        let beats = (0..<100).map { i in
            BeatMarker(index: i, time: Double(i) * 0.5, isDownbeat: i % 4 == 0)
        }

        let sections = [
            Section(type: .intro, startTime: 0, endTime: 10, startBeat: 0, endBeat: 20, confidence: 0.9)
        ]

        let result = generator.generate(sections: sections, beats: beats)

        XCTAssertTrue(result.cues.contains { $0.type == .load })
        XCTAssertEqual(result.cues.first?.type, .load)
    }

    func testCueMaxLimit() {
        let generator = CueGenerator(maxCues: 4)

        let beats = (0..<200).map { i in
            BeatMarker(index: i, time: Double(i) * 0.5, isDownbeat: i % 4 == 0)
        }

        let sections = [
            Section(type: .intro, startTime: 0, endTime: 10, startBeat: 0, endBeat: 20, confidence: 0.9),
            Section(type: .build, startTime: 10, endTime: 20, startBeat: 20, endBeat: 40, confidence: 0.9),
            Section(type: .drop, startTime: 20, endTime: 40, startBeat: 40, endBeat: 80, confidence: 0.9),
            Section(type: .breakdown, startTime: 40, endTime: 50, startBeat: 80, endBeat: 100, confidence: 0.9),
            Section(type: .build, startTime: 50, endTime: 60, startBeat: 100, endBeat: 120, confidence: 0.9),
            Section(type: .drop, startTime: 60, endTime: 80, startBeat: 120, endBeat: 160, confidence: 0.9),
            Section(type: .outro, startTime: 80, endTime: 100, startBeat: 160, endBeat: 200, confidence: 0.9),
        ]

        let result = generator.generate(sections: sections, beats: beats)

        XCTAssertLessThanOrEqual(result.cues.count, 4)
    }

    func testCueColors() {
        XCTAssertEqual(CueColor.forType(.load), .green)
        XCTAssertEqual(CueColor.forType(.drop), .red)
        XCTAssertEqual(CueColor.forType(.build), .yellow)
        XCTAssertEqual(CueColor.forType(.breakdown), .purple)
    }
}

// MARK: - Integration Tests

final class AnalyzerIntegrationTests: XCTestCase {
    func testFullAnalyzerCreation() {
        let analyzer = Analyzer(sampleRate: 48000, waveformBins: 100)
        XCTAssertNotNil(analyzer)
    }
}

// MARK: - Audio Buffer Tests

final class AudioBufferTests: XCTestCase {
    func testAudioBufferMonoConversion() {
        let samples: [Float] = [1, 2, 3, 4, 5, 6]
        let buffer = AudioBuffer(samples: samples, sampleRate: 48000, channelCount: 2, duration: 0.0625)

        let mono = buffer.monoSamples()

        XCTAssertEqual(mono.count, 3)
        XCTAssertEqual(mono[0], 1.5) // (1+2)/2
        XCTAssertEqual(mono[1], 3.5) // (3+4)/2
        XCTAssertEqual(mono[2], 5.5) // (5+6)/2
    }

    func testAudioBufferMonoPassthrough() {
        let samples: [Float] = [1, 2, 3, 4]
        let buffer = AudioBuffer(samples: samples, sampleRate: 48000, channelCount: 1, duration: 0.083)

        let mono = buffer.monoSamples()

        XCTAssertEqual(mono, samples)
    }
}

// MARK: - Utility Tests

final class UtilityTests: XCTestCase {
    func testIntModuloExtension() {
        XCTAssertEqual(5.modulo(12), 5)
        XCTAssertEqual(15.modulo(12), 3)
        XCTAssertEqual((-1).modulo(12), 11)
        XCTAssertEqual((-13).modulo(12), 11)
    }

    func testIntIsPowerOfTwo() {
        XCTAssertTrue(1.isPowerOfTwo)
        XCTAssertTrue(2.isPowerOfTwo)
        XCTAssertTrue(1024.isPowerOfTwo)
        XCTAssertTrue(4096.isPowerOfTwo)

        XCTAssertFalse(0.isPowerOfTwo)
        XCTAssertFalse(3.isPowerOfTwo)
        XCTAssertFalse(1000.isPowerOfTwo)
    }
}

// MARK: - Loudness Analyzer Tests

final class LoudnessAnalyzerTests: XCTestCase {
    func testLoudnessAnalyzerCreation() {
        let analyzer = LoudnessAnalyzer(sampleRate: 48000)
        XCTAssertNotNil(analyzer)
    }

    func testLoudnessAnalysisReturnsValidValues() {
        let analyzer = LoudnessAnalyzer(sampleRate: 48000)

        // Generate some noise
        var samples = [Float](repeating: 0, count: 48000 * 5)
        for i in 0..<samples.count {
            samples[i] = Float.random(in: -0.5...0.5)
        }

        let result = analyzer.analyze(samples)

        // Integrated loudness should be negative (in LUFS)
        XCTAssertLessThan(result.integratedLoudness, 0)
        // Loudness range should be non-negative
        XCTAssertGreaterThanOrEqual(result.loudnessRange, 0)
        // Peaks should be reasonable
        XCTAssertGreaterThan(result.truePeak, -100)
    }
}

// MARK: - Embedding Generator Tests

final class EmbeddingGeneratorTests: XCTestCase {
    func testEmbeddingGeneratorCreation() {
        let generator = EmbeddingGenerator(sampleRate: 48000)
        XCTAssertNotNil(generator)
    }

    func testEmbeddingDimension() {
        let generator = EmbeddingGenerator(sampleRate: 48000, embeddingDim: 128)

        var samples = [Float](repeating: 0, count: 48000 * 3)
        for i in 0..<samples.count {
            samples[i] = Float.random(in: -0.5...0.5)
        }

        let embedding = generator.generate(samples)

        XCTAssertEqual(embedding.vector.count, 128)
    }

    func testEmbeddingSimilarityRange() {
        let generator = EmbeddingGenerator(sampleRate: 48000)

        var samples1 = [Float](repeating: 0, count: 48000 * 3)
        var samples2 = [Float](repeating: 0, count: 48000 * 3)

        for i in 0..<samples1.count {
            samples1[i] = Float.random(in: -0.5...0.5)
            samples2[i] = Float.random(in: -0.5...0.5)
        }

        let embedding1 = generator.generate(samples1)
        let embedding2 = generator.generate(samples2)

        let similarity = embedding1.similarity(to: embedding2)

        // Cosine similarity should be between -1 and 1
        XCTAssertGreaterThanOrEqual(similarity, -1)
        XCTAssertLessThanOrEqual(similarity, 1)
    }

    func testEmbeddingSelfSimilarity() {
        let generator = EmbeddingGenerator(sampleRate: 48000)

        var samples = [Float](repeating: 0, count: 48000 * 3)
        for i in 0..<samples.count {
            samples[i] = Float.random(in: -0.5...0.5)
        }

        let embedding = generator.generate(samples)

        // Self-similarity should be 1
        let selfSimilarity = embedding.similarity(to: embedding)
        XCTAssertEqual(selfSimilarity, 1.0, accuracy: 0.001)
    }
}

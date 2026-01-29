import Testing
@testable import AnalyzerSwift

// MARK: - FFT Tests

@Test func fftProcessorCreation() {
    let fft = FFTProcessor(fftSize: 2048)
    #expect(fft.fftSize == 2048)
}

@Test func fftProcessorPowerOfTwoRequired() {
    // This should work
    _ = FFTProcessor(fftSize: 1024)
    _ = FFTProcessor(fftSize: 4096)
}

@Test func magnitudeSpectrumOutput() {
    let fft = FFTProcessor(fftSize: 1024)

    // Generate a simple sine wave at 440 Hz
    let sampleRate: Double = 48000
    let frequency: Double = 440
    var samples = [Float](repeating: 0, count: 1024)

    for i in 0..<1024 {
        samples[i] = Float(sin(2 * .pi * frequency * Double(i) / sampleRate))
    }

    let spectrum = fft.magnitudeSpectrum(samples)

    #expect(spectrum.count == 512) // Half of FFT size
}

@Test func stftOutput() {
    let fft = FFTProcessor(fftSize: 512)
    let samples = [Float](repeating: 0.5, count: 2048)

    let spectrogram = fft.stft(samples, hopSize: 256)

    #expect(spectrogram.count > 0)
    #expect(spectrogram[0].count == 256) // Half of FFT size
}

@Test func spectralFlux() {
    let fft = FFTProcessor(fftSize: 512)

    // Create a spectrogram with increasing energy
    let spectrogram: [[Float]] = [
        [Float](repeating: 0.1, count: 256),
        [Float](repeating: 0.2, count: 256),
        [Float](repeating: 0.5, count: 256),
        [Float](repeating: 0.3, count: 256)
    ]

    let flux = fft.spectralFlux(spectrogram)

    #expect(flux.count == 4)
    #expect(flux[0] == 0) // First frame has no flux
    #expect(flux[2] > flux[1]) // Larger jump
}

@Test func chromaFeatures() {
    let fft = FFTProcessor(fftSize: 2048)

    // Generate test signal
    var samples = [Float](repeating: 0, count: 8192)
    for i in 0..<8192 {
        samples[i] = Float(sin(2 * .pi * 440 * Double(i) / 48000))
    }

    let chroma = fft.chromaFeatures(samples, sampleRate: 48000, hopSize: 1024)

    #expect(chroma.count > 0)
    #expect(chroma[0].count == 12) // 12 pitch classes
}

// MARK: - Beatgrid Detector Tests

@Test func beatgridDetectorCreation() {
    let detector = BeatgridDetector(sampleRate: 48000)
    #expect(detector != nil)
}

@Test func beatgridDetectionWithSilence() {
    let detector = BeatgridDetector(sampleRate: 48000, fftSize: 1024, hopSize: 256)
    let samples = [Float](repeating: 0, count: 48000 * 10) // 10 seconds of silence

    let result = detector.detect(samples)

    // Should still produce a result, but confidence may be low
    #expect(result.tempoMap.first != nil)
}

@Test func beatgridTempoRange() {
    let detector = BeatgridDetector(
        sampleRate: 48000,
        tempoFloor: 100,
        tempoCeil: 150
    )

    // Generate click track at ~120 BPM
    var samples = [Float](repeating: 0, count: 48000 * 10)
    let beatInterval = Int(48000 * 60.0 / 120.0) // Samples per beat at 120 BPM

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
        #expect(tempo >= 100)
        #expect(tempo <= 150)
    }
}

@Test func beatMarkerMonotonicity() {
    let detector = BeatgridDetector(sampleRate: 48000)

    // Simple test signal
    var samples = [Float](repeating: 0, count: 48000 * 5)
    for i in 0..<samples.count {
        samples[i] = Float.random(in: -0.5...0.5)
    }

    let result = detector.detect(samples)

    // Verify beats are monotonically increasing
    for i in 1..<result.beats.count {
        #expect(result.beats[i].time > result.beats[i-1].time, "Beats should be monotonically increasing")
        #expect(result.beats[i].index > result.beats[i-1].index, "Beat indices should be monotonically increasing")
    }
}

// MARK: - Key Detector Tests

@Test func keyDetectorCreation() {
    let detector = KeyDetector(sampleRate: 48000)
    #expect(detector != nil)
}

@Test func keyDetectionReturnsValidKey() {
    let detector = KeyDetector(sampleRate: 48000, fftSize: 2048, hopSize: 1024)

    // Generate a simple A minor chord (A, C, E)
    let frequencies: [Double] = [440, 523.25, 659.25] // A4, C5, E5
    var samples = [Float](repeating: 0, count: 48000 * 3)

    for i in 0..<samples.count {
        var sum: Float = 0
        for freq in frequencies {
            sum += Float(sin(2 * .pi * freq * Double(i) / 48000))
        }
        samples[i] = sum / 3
    }

    let key = detector.detect(samples)

    #expect(key.pitchClass >= 0)
    #expect(key.pitchClass < 12)
    #expect(key.confidence >= 0)
    #expect(key.confidence <= 1)
}

@Test func musicalKeyNameFormat() {
    let majorKey = MusicalKey(pitchClass: 0, isMinor: false, confidence: 0.9)
    #expect(majorKey.name == "C")

    let minorKey = MusicalKey(pitchClass: 9, isMinor: true, confidence: 0.9)
    #expect(minorKey.name == "Am")
}

@Test func camelotNotation() {
    // Test some known mappings
    let cMajor = MusicalKey(pitchClass: 0, isMinor: false, confidence: 1.0)
    #expect(cMajor.camelot == "8B")

    let aMinor = MusicalKey(pitchClass: 9, isMinor: true, confidence: 1.0)
    #expect(aMinor.camelot == "8A")

    let fMajor = MusicalKey(pitchClass: 5, isMinor: false, confidence: 1.0)
    #expect(fMajor.camelot == "7B")
}

@Test func openKeyNotation() {
    let key = MusicalKey(pitchClass: 0, isMinor: false, confidence: 1.0)
    #expect(key.openKey == "1d") // C major = 1d

    let minorKey = MusicalKey(pitchClass: 0, isMinor: true, confidence: 1.0)
    #expect(minorKey.openKey == "1m") // C minor = 1m
}

// MARK: - Energy Analyzer Tests

@Test func energyAnalyzerCreation() {
    let analyzer = EnergyAnalyzer(sampleRate: 48000)
    #expect(analyzer != nil)
}

@Test func energyAnalysisReturnsValidRange() {
    let analyzer = EnergyAnalyzer(sampleRate: 48000, fftSize: 1024, hopSize: 512)

    var samples = [Float](repeating: 0, count: 48000 * 3)
    for i in 0..<samples.count {
        samples[i] = Float.random(in: -0.5...0.5)
    }

    let result = analyzer.analyze(samples)

    #expect(result.globalEnergy >= 1)
    #expect(result.globalEnergy <= 10)
    #expect(result.rms >= 0)
    #expect(result.peak >= 0)
}

@Test func energyBandDistribution() {
    let analyzer = EnergyAnalyzer(sampleRate: 48000)

    var samples = [Float](repeating: 0, count: 48000 * 2)
    for i in 0..<samples.count {
        samples[i] = Float.random(in: -0.3...0.3)
    }

    let result = analyzer.analyze(samples)

    // Band energies should sum to approximately 1 (normalized)
    let sum = result.lowEnergy + result.midEnergy + result.highEnergy
    #expect(abs(sum - 1.0) < 0.01)
}

@Test func energyCurveNormalization() {
    let analyzer = EnergyAnalyzer(sampleRate: 48000, hopSize: 1024)

    var samples = [Float](repeating: 0, count: 48000 * 2)
    for i in 0..<samples.count {
        samples[i] = Float.random(in: -0.5...0.5)
    }

    let result = analyzer.analyze(samples)

    // All values in curve should be 0-1
    for value in result.curve {
        #expect(value >= 0)
        #expect(value <= 1)
    }
}

// MARK: - Section Detector Tests

@Test func sectionDetectorCreation() {
    let detector = SectionDetector(sampleRate: 48000)
    #expect(detector != nil)
}

@Test func sectionDetectionWithEmptyBeats() {
    let detector = SectionDetector(sampleRate: 48000)
    let samples = [Float](repeating: 0, count: 48000)

    let result = detector.detect(samples, beats: [], tempo: 120)

    #expect(result.sections.isEmpty)
}

@Test func sectionTypes() {
    // Verify all section types are accounted for
    let allTypes = SectionType.allCases
    #expect(allTypes.count == 6)
    #expect(allTypes.contains(.intro))
    #expect(allTypes.contains(.verse))
    #expect(allTypes.contains(.build))
    #expect(allTypes.contains(.drop))
    #expect(allTypes.contains(.breakdown))
    #expect(allTypes.contains(.outro))
}

// MARK: - Cue Generator Tests

@Test func cueGeneratorCreation() {
    let generator = CueGenerator(maxCues: 8)
    #expect(generator != nil)
}

@Test func cueGenerationWithEmptyInput() {
    let generator = CueGenerator()
    let result = generator.generate(sections: [], beats: [])

    #expect(result.cues.isEmpty)
}

@Test func cueGenerationAlwaysIncludesLoadPoint() {
    let generator = CueGenerator()

    let beats = (0..<100).map { i in
        BeatMarker(index: i, time: Double(i) * 0.5, isDownbeat: i % 4 == 0)
    }

    let sections = [
        Section(type: .intro, startTime: 0, endTime: 10, startBeat: 0, endBeat: 20, confidence: 0.9)
    ]

    let result = generator.generate(sections: sections, beats: beats)

    #expect(result.cues.contains { $0.type == .load })
    #expect(result.cues.first?.type == .load)
}

@Test func cueMaxLimit() {
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

    #expect(result.cues.count <= 4)
}

@Test func cueColors() {
    // Verify color mappings
    #expect(CueColor.forType(.load) == .green)
    #expect(CueColor.forType(.drop) == .red)
    #expect(CueColor.forType(.build) == .yellow)
    #expect(CueColor.forType(.breakdown) == .purple)
}

// MARK: - Integration Tests

@Test func fullAnalyzerCreation() {
    let analyzer = Analyzer(sampleRate: 48000, waveformBins: 100)
    #expect(analyzer != nil)
}

// MARK: - Audio Buffer Tests

@Test func audioBufferMonoConversion() {
    // Stereo buffer
    let samples: [Float] = [1, 2, 3, 4, 5, 6] // L, R, L, R, L, R
    let buffer = AudioBuffer(samples: samples, sampleRate: 48000, channelCount: 2, duration: 0.0625)

    let mono = buffer.monoSamples()

    #expect(mono.count == 3)
    #expect(mono[0] == 1.5) // (1+2)/2
    #expect(mono[1] == 3.5) // (3+4)/2
    #expect(mono[2] == 5.5) // (5+6)/2
}

@Test func audioBufferMonoPassthrough() {
    let samples: [Float] = [1, 2, 3, 4]
    let buffer = AudioBuffer(samples: samples, sampleRate: 48000, channelCount: 1, duration: 0.083)

    let mono = buffer.monoSamples()

    #expect(mono == samples)
}

// MARK: - Utility Tests

@Test func intModuloExtension() {
    #expect(5.modulo(12) == 5)
    #expect(15.modulo(12) == 3)
    #expect((-1).modulo(12) == 11)
    #expect((-13).modulo(12) == 11)
}

@Test func intIsPowerOfTwo() {
    #expect(1.isPowerOfTwo)
    #expect(2.isPowerOfTwo)
    #expect(1024.isPowerOfTwo)
    #expect(4096.isPowerOfTwo)

    #expect(!0.isPowerOfTwo)
    #expect(!3.isPowerOfTwo)
    #expect(!1000.isPowerOfTwo)
}

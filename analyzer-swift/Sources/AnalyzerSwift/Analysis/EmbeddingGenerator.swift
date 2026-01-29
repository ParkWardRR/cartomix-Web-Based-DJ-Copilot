import Accelerate

/// Audio embedding vector for similarity comparison
public struct AudioEmbedding: Sendable {
    /// 128-dimensional embedding vector
    public let vector: [Float]
    /// Spectral centroid mean (brightness)
    public let spectralCentroid: Float
    /// Spectral rolloff mean
    public let spectralRolloff: Float
    /// Zero crossing rate (percussiveness)
    public let zeroCrossingRate: Float
    /// Spectral flatness (noise vs tonal)
    public let spectralFlatness: Float
    /// Tempo stability score
    public let tempoStability: Float
    /// Harmonic-to-noise ratio
    public let harmonicRatio: Float

    public init(
        vector: [Float],
        spectralCentroid: Float,
        spectralRolloff: Float,
        zeroCrossingRate: Float,
        spectralFlatness: Float,
        tempoStability: Float,
        harmonicRatio: Float
    ) {
        self.vector = vector
        self.spectralCentroid = spectralCentroid
        self.spectralRolloff = spectralRolloff
        self.zeroCrossingRate = zeroCrossingRate
        self.spectralFlatness = spectralFlatness
        self.tempoStability = tempoStability
        self.harmonicRatio = harmonicRatio
    }

    /// Calculate cosine similarity between two embeddings
    public func similarity(to other: AudioEmbedding) -> Float {
        guard vector.count == other.vector.count, !vector.isEmpty else { return 0 }

        var dotProduct: Float = 0
        var normA: Float = 0
        var normB: Float = 0

        vDSP_dotpr(vector, 1, other.vector, 1, &dotProduct, vDSP_Length(vector.count))
        vDSP_svesq(vector, 1, &normA, vDSP_Length(vector.count))
        vDSP_svesq(other.vector, 1, &normB, vDSP_Length(other.vector.count))

        let denominator = sqrt(normA) * sqrt(normB)
        return denominator > 0 ? dotProduct / denominator : 0
    }

    /// Calculate weighted vibe similarity (for DJ set continuity)
    public func vibeSimilarity(to other: AudioEmbedding) -> Float {
        // Weights for different aspects of "vibe"
        let vectorWeight: Float = 0.5       // Overall spectral similarity
        let centroidWeight: Float = 0.15    // Brightness similarity
        let flatnessWeight: Float = 0.15    // Tonal vs noise similarity
        let harmonicWeight: Float = 0.1     // Harmonic content
        let zcrWeight: Float = 0.1          // Percussiveness

        let vectorSim = similarity(to: other)

        // Normalize differences to 0-1 similarity
        let centroidSim = 1 - min(1, abs(spectralCentroid - other.spectralCentroid) / 4000)
        let flatnessSim = 1 - abs(spectralFlatness - other.spectralFlatness)
        let harmonicSim = 1 - abs(harmonicRatio - other.harmonicRatio)
        let zcrSim = 1 - min(1, abs(zeroCrossingRate - other.zeroCrossingRate) / 0.2)

        return vectorSim * vectorWeight +
               centroidSim * centroidWeight +
               flatnessSim * flatnessWeight +
               harmonicSim * harmonicWeight +
               zcrSim * zcrWeight
    }
}

/// Generates audio embeddings for similarity comparison
public final class EmbeddingGenerator: @unchecked Sendable {
    private let fft: FFTProcessor
    private let sampleRate: Double
    private let hopSize: Int
    private let embeddingDim: Int

    public init(sampleRate: Double = 48000, fftSize: Int = 2048, hopSize: Int = 1024, embeddingDim: Int = 128) {
        self.fft = FFTProcessor(fftSize: fftSize)
        self.sampleRate = sampleRate
        self.hopSize = hopSize
        self.embeddingDim = embeddingDim
    }

    /// Generate embedding from audio samples
    public func generate(_ samples: [Float]) -> AudioEmbedding {
        guard !samples.isEmpty else {
            return AudioEmbedding(
                vector: [Float](repeating: 0, count: embeddingDim),
                spectralCentroid: 0,
                spectralRolloff: 0,
                zeroCrossingRate: 0,
                spectralFlatness: 0,
                tempoStability: 0,
                harmonicRatio: 0
            )
        }

        // Compute spectrogram
        let spectrogram = fft.stft(samples, hopSize: hopSize)

        // Extract features
        let spectralCentroid = calculateMeanSpectralCentroid(spectrogram)
        let spectralRolloff = calculateMeanSpectralRolloff(spectrogram)
        let zeroCrossingRate = calculateZeroCrossingRate(samples)
        let spectralFlatness = calculateMeanSpectralFlatness(spectrogram)
        let tempoStability = calculateTempoStability(spectrogram)
        let harmonicRatio = calculateHarmonicRatio(spectrogram)

        // Generate embedding vector from MFCC-like features
        let vector = generateEmbeddingVector(spectrogram)

        return AudioEmbedding(
            vector: vector,
            spectralCentroid: spectralCentroid,
            spectralRolloff: spectralRolloff,
            zeroCrossingRate: zeroCrossingRate,
            spectralFlatness: spectralFlatness,
            tempoStability: tempoStability,
            harmonicRatio: harmonicRatio
        )
    }

    private func calculateMeanSpectralCentroid(_ spectrogram: [[Float]]) -> Float {
        guard !spectrogram.isEmpty, !spectrogram[0].isEmpty else { return 0 }

        var totalCentroid: Float = 0
        let binWidth = Float(sampleRate) / Float(fft.fftSize)

        for spectrum in spectrogram {
            // Convert from dB to linear magnitude
            var linearMagnitudes = spectrum.map { pow(10, $0 / 20) }

            var weightedSum: Float = 0
            var magnitudeSum: Float = 0

            for (bin, magnitude) in linearMagnitudes.enumerated() {
                let frequency = Float(bin) * binWidth
                weightedSum += frequency * magnitude
                magnitudeSum += magnitude
            }

            if magnitudeSum > 0 {
                totalCentroid += weightedSum / magnitudeSum
            }
        }

        return totalCentroid / Float(max(1, spectrogram.count))
    }

    private func calculateMeanSpectralRolloff(_ spectrogram: [[Float]], threshold: Float = 0.85) -> Float {
        guard !spectrogram.isEmpty else { return 0 }

        var totalRolloff: Float = 0
        let binWidth = Float(sampleRate) / Float(fft.fftSize)

        for spectrum in spectrogram {
            let linearMagnitudes = spectrum.map { pow(10, $0 / 20) }
            var totalEnergy: Float = 0
            vDSP_sve(linearMagnitudes, 1, &totalEnergy, vDSP_Length(linearMagnitudes.count))

            let targetEnergy = totalEnergy * threshold
            var cumulativeEnergy: Float = 0
            var rolloffBin = 0

            for (bin, magnitude) in linearMagnitudes.enumerated() {
                cumulativeEnergy += magnitude
                if cumulativeEnergy >= targetEnergy {
                    rolloffBin = bin
                    break
                }
            }

            totalRolloff += Float(rolloffBin) * binWidth
        }

        return totalRolloff / Float(max(1, spectrogram.count))
    }

    private func calculateZeroCrossingRate(_ samples: [Float]) -> Float {
        guard samples.count > 1 else { return 0 }

        var crossings = 0
        for i in 1..<samples.count {
            if (samples[i-1] >= 0 && samples[i] < 0) || (samples[i-1] < 0 && samples[i] >= 0) {
                crossings += 1
            }
        }

        return Float(crossings) / Float(samples.count)
    }

    private func calculateMeanSpectralFlatness(_ spectrogram: [[Float]]) -> Float {
        guard !spectrogram.isEmpty else { return 0 }

        var totalFlatness: Float = 0

        for spectrum in spectrogram {
            let linearMagnitudes = spectrum.map { max(1e-10, pow(10, $0 / 20)) }

            // Geometric mean
            var logSum: Float = 0
            for mag in linearMagnitudes {
                logSum += log(mag)
            }
            let geometricMean = exp(logSum / Float(linearMagnitudes.count))

            // Arithmetic mean
            var arithmeticMean: Float = 0
            vDSP_meanv(linearMagnitudes, 1, &arithmeticMean, vDSP_Length(linearMagnitudes.count))

            // Flatness is ratio of geometric to arithmetic mean
            if arithmeticMean > 0 {
                totalFlatness += geometricMean / arithmeticMean
            }
        }

        return totalFlatness / Float(max(1, spectrogram.count))
    }

    private func calculateTempoStability(_ spectrogram: [[Float]]) -> Float {
        // Analyze onset strength variance to estimate tempo stability
        guard spectrogram.count > 10 else { return 0.5 }

        var onsetStrength = [Float]()

        for i in 1..<spectrogram.count {
            var flux: Float = 0
            for j in 0..<spectrogram[i].count {
                let diff = spectrogram[i][j] - spectrogram[i-1][j]
                flux += max(0, diff)
            }
            onsetStrength.append(flux)
        }

        // Calculate coefficient of variation (lower = more stable)
        var mean: Float = 0
        var stdDev: Float = 0
        vDSP_meanv(onsetStrength, 1, &mean, vDSP_Length(onsetStrength.count))
        vDSP_normalize(onsetStrength, 1, nil, 1, &mean, &stdDev, vDSP_Length(onsetStrength.count))

        let cv = mean > 0 ? stdDev / mean : 1
        return max(0, min(1, 1 - cv))
    }

    private func calculateHarmonicRatio(_ spectrogram: [[Float]]) -> Float {
        // Estimate harmonic content vs noise
        guard !spectrogram.isEmpty, spectrogram[0].count > 10 else { return 0.5 }

        var totalHarmonicRatio: Float = 0

        for spectrum in spectrogram {
            let linearMagnitudes = spectrum.map { pow(10, $0 / 20) }

            // Find peaks (potential harmonics)
            var peakEnergy: Float = 0
            var totalEnergy: Float = 0

            for i in 1..<(linearMagnitudes.count - 1) {
                let current = linearMagnitudes[i]
                totalEnergy += current

                if current > linearMagnitudes[i-1] && current > linearMagnitudes[i+1] {
                    peakEnergy += current
                }
            }

            if totalEnergy > 0 {
                totalHarmonicRatio += peakEnergy / totalEnergy
            }
        }

        return totalHarmonicRatio / Float(max(1, spectrogram.count))
    }

    private func generateEmbeddingVector(_ spectrogram: [[Float]]) -> [Float] {
        guard !spectrogram.isEmpty else {
            return [Float](repeating: 0, count: embeddingDim)
        }

        // Generate MFCC-like features by averaging spectrogram into mel-like bands
        // then taking statistics across time

        let numBands = embeddingDim / 4  // 32 bands
        let binsPerBand = spectrogram[0].count / numBands

        var bandMeans = [Float](repeating: 0, count: numBands)
        var bandStds = [Float](repeating: 0, count: numBands)
        var bandDeltas = [Float](repeating: 0, count: numBands)
        var bandDeltaDeltas = [Float](repeating: 0, count: numBands)

        // Calculate band statistics
        for band in 0..<numBands {
            var bandValues = [Float]()
            let startBin = band * binsPerBand
            let endBin = min(startBin + binsPerBand, spectrogram[0].count)

            for frame in spectrogram {
                var bandSum: Float = 0
                for bin in startBin..<endBin {
                    bandSum += pow(10, frame[bin] / 20)
                }
                bandValues.append(bandSum / Float(endBin - startBin))
            }

            // Mean
            var mean: Float = 0
            vDSP_meanv(bandValues, 1, &mean, vDSP_Length(bandValues.count))
            bandMeans[band] = mean

            // Standard deviation
            var variance: Float = 0
            for value in bandValues {
                variance += (value - mean) * (value - mean)
            }
            bandStds[band] = sqrt(variance / Float(max(1, bandValues.count)))

            // Delta (first derivative)
            if bandValues.count > 2 {
                var deltaSum: Float = 0
                for i in 1..<(bandValues.count - 1) {
                    deltaSum += abs(bandValues[i+1] - bandValues[i-1])
                }
                bandDeltas[band] = deltaSum / Float(bandValues.count - 2)
            }

            // Delta-delta (second derivative)
            if bandValues.count > 4 {
                var ddSum: Float = 0
                for i in 2..<(bandValues.count - 2) {
                    let d1 = bandValues[i] - bandValues[i-2]
                    let d2 = bandValues[i+2] - bandValues[i]
                    ddSum += abs(d2 - d1)
                }
                bandDeltaDeltas[band] = ddSum / Float(bandValues.count - 4)
            }
        }

        // Normalize and concatenate
        var embedding = [Float]()
        embedding.append(contentsOf: normalize(bandMeans))
        embedding.append(contentsOf: normalize(bandStds))
        embedding.append(contentsOf: normalize(bandDeltas))
        embedding.append(contentsOf: normalize(bandDeltaDeltas))

        // Pad or truncate to exact embedding dimension
        while embedding.count < embeddingDim {
            embedding.append(0)
        }
        if embedding.count > embeddingDim {
            embedding = Array(embedding.prefix(embeddingDim))
        }

        return embedding
    }

    private func normalize(_ values: [Float]) -> [Float] {
        guard !values.isEmpty else { return values }

        var maxVal: Float = 0
        vDSP_maxmgv(values, 1, &maxVal, vDSP_Length(values.count))

        if maxVal > 0 {
            return values.map { $0 / maxVal }
        }
        return values
    }
}

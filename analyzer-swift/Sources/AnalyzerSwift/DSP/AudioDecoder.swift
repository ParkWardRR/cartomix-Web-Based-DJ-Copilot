@preconcurrency import AVFoundation
import Accelerate

/// Audio format after decoding
public struct AudioBuffer: Sendable {
    public let samples: [Float]
    public let sampleRate: Double
    public let channelCount: Int
    public let duration: Double

    public var frameCount: Int { samples.count / channelCount }

    /// Get mono samples (average of channels if stereo)
    public func monoSamples() -> [Float] {
        if channelCount == 1 {
            return samples
        }

        var mono = [Float](repeating: 0, count: frameCount)
        let scale = 1.0 / Float(channelCount)

        for i in 0..<frameCount {
            var sum: Float = 0
            for ch in 0..<channelCount {
                sum += samples[i * channelCount + ch]
            }
            mono[i] = sum * scale
        }

        return mono
    }
}

/// Decodes audio files to PCM samples using AVFoundation
public final class AudioDecoder: @unchecked Sendable {
    private let targetSampleRate: Double
    private let mono: Bool

    public init(targetSampleRate: Double = 48000, mono: Bool = false) {
        self.targetSampleRate = targetSampleRate
        self.mono = mono
    }

    /// Decode audio file to float samples
    public func decode(path: String) throws -> AudioBuffer {
        let url = URL(fileURLWithPath: path)

        // Open the audio file
        let file = try AVAudioFile(forReading: url)
        let sourceFormat = file.processingFormat

        // Calculate output format
        let outputChannels: AVAudioChannelCount = mono ? 1 : AVAudioChannelCount(sourceFormat.channelCount)
        guard let outputFormat = AVAudioFormat(
            commonFormat: .pcmFormatFloat32,
            sampleRate: targetSampleRate,
            channels: outputChannels,
            interleaved: true
        ) else {
            throw AudioDecoderError.invalidFormat
        }

        // Create converter if needed
        let needsConversion = sourceFormat.sampleRate != targetSampleRate ||
                             (mono && sourceFormat.channelCount > 1)

        let samples: [Float]
        let duration: Double

        if needsConversion {
            samples = try decodeWithConversion(file: file, to: outputFormat)
        } else {
            samples = try decodeDirectly(file: file)
        }

        duration = Double(samples.count / Int(outputChannels)) / targetSampleRate

        return AudioBuffer(
            samples: samples,
            sampleRate: targetSampleRate,
            channelCount: Int(outputChannels),
            duration: duration
        )
    }

    private func decodeDirectly(file: AVAudioFile) throws -> [Float] {
        let frameCount = AVAudioFrameCount(file.length)
        guard let buffer = AVAudioPCMBuffer(pcmFormat: file.processingFormat, frameCapacity: frameCount) else {
            throw AudioDecoderError.bufferAllocationFailed
        }

        try file.read(into: buffer)

        guard let floatData = buffer.floatChannelData else {
            throw AudioDecoderError.noFloatData
        }

        let channelCount = Int(file.processingFormat.channelCount)
        let sampleCount = Int(buffer.frameLength) * channelCount
        var samples = [Float](repeating: 0, count: sampleCount)

        // Interleave channels
        for frame in 0..<Int(buffer.frameLength) {
            for ch in 0..<channelCount {
                samples[frame * channelCount + ch] = floatData[ch][frame]
            }
        }

        return samples
    }

    private func decodeWithConversion(file: AVAudioFile, to outputFormat: AVAudioFormat) throws -> [Float] {
        guard let converter = AVAudioConverter(from: file.processingFormat, to: outputFormat) else {
            throw AudioDecoderError.converterCreationFailed
        }

        // Calculate output frame count based on sample rate ratio
        let ratio = targetSampleRate / file.processingFormat.sampleRate
        let outputFrameCount = AVAudioFrameCount(Double(file.length) * ratio)

        guard let outputBuffer = AVAudioPCMBuffer(pcmFormat: outputFormat, frameCapacity: outputFrameCount) else {
            throw AudioDecoderError.bufferAllocationFailed
        }

        // Read input in chunks
        let inputFrameCapacity: AVAudioFrameCount = 8192
        guard let inputBuffer = AVAudioPCMBuffer(pcmFormat: file.processingFormat, frameCapacity: inputFrameCapacity) else {
            throw AudioDecoderError.bufferAllocationFailed
        }

        var error: NSError?
        let inputBlock: AVAudioConverterInputBlock = { inNumPackets, outStatus in
            do {
                try file.read(into: inputBuffer, frameCount: min(inNumPackets, inputFrameCapacity))
                if inputBuffer.frameLength == 0 {
                    outStatus.pointee = .endOfStream
                    return nil
                }
                outStatus.pointee = .haveData
                return inputBuffer
            } catch {
                outStatus.pointee = .endOfStream
                return nil
            }
        }

        let status = converter.convert(to: outputBuffer, error: &error, withInputFrom: inputBlock)

        if let error = error {
            throw AudioDecoderError.conversionFailed(error.localizedDescription)
        }

        if status == .error {
            throw AudioDecoderError.conversionFailed("Unknown conversion error")
        }

        guard let floatData = outputBuffer.floatChannelData else {
            throw AudioDecoderError.noFloatData
        }

        let channelCount = Int(outputFormat.channelCount)
        let sampleCount = Int(outputBuffer.frameLength) * channelCount
        var samples = [Float](repeating: 0, count: sampleCount)

        // Interleave channels
        for frame in 0..<Int(outputBuffer.frameLength) {
            for ch in 0..<channelCount {
                samples[frame * channelCount + ch] = floatData[ch][frame]
            }
        }

        return samples
    }
}

public enum AudioDecoderError: Error, Sendable {
    case invalidFormat
    case bufferAllocationFailed
    case noFloatData
    case converterCreationFailed
    case conversionFailed(String)
}

import Foundation
import GRPC
import NIO
import NIOCore
import NIOPosix
import Logging
import SwiftProtobuf
import AnalyzerSwift

/// gRPC server for the analyzer worker service
/// Implements the AnalyzerWorker proto service
public final class AnalyzerGRPCServer {
    private let port: Int
    private let logger: Logger
    private let group: MultiThreadedEventLoopGroup
    private var server: Server?

    public init(port: Int, logger: Logger) {
        self.port = port
        self.logger = logger
        self.group = MultiThreadedEventLoopGroup(numberOfThreads: System.coreCount)
    }

    deinit {
        try? group.syncShutdownGracefully()
    }

    /// Start the gRPC server
    public func start() async throws {
        // Check hardware requirements
        try HardwareCheck.requireMetal()

        let provider = AnalyzerWorkerProvider(logger: logger)

        let server = try await Server.insecure(group: group)
            .withServiceProviders([provider])
            .withLogger(logger)
            .bind(host: "0.0.0.0", port: port)
            .get()

        self.server = server
        logger.info("gRPC server started on port \(port)")

        // Wait for server to close
        try await server.channel.closeFuture.get()
    }

    /// Stop the gRPC server gracefully
    public func stop() async throws {
        logger.info("Stopping gRPC server...")
        try await server?.initiateGracefulShutdown()
    }
}

/// Provider implementation for the AnalyzerWorker service
/// Maps gRPC calls to the Swift analyzer
final class AnalyzerWorkerProvider: CallHandlerProvider {
    var serviceName: Substring { "cartomix.analyzer.AnalyzerWorker" }

    private let logger: Logger
    private let analyzer: Analyzer

    init(logger: Logger) {
        self.logger = logger
        self.analyzer = Analyzer(
            sampleRate: 48000,
            waveformBins: 200,
            enableOpenL3: true,
            enableSoundClassification: true
        )
    }

    func handle(method name: Substring, context: CallHandlerContext) -> GRPCServerHandlerProtocol? {
        switch name {
        case "AnalyzeTrack":
            return UnaryServerHandler(
                context: context,
                requestDeserializer: ProtobufDeserializer<AnalyzeJobRequest>(),
                responseSerializer: ProtobufSerializer<AnalyzeResultResponse>(),
                interceptors: [],
                userFunction: analyzeTrack
            )
        default:
            return nil
        }
    }

    // MARK: - RPC Implementations

    private func analyzeTrack(
        request: AnalyzeJobRequest,
        context: StatusOnlyCallContext
    ) -> EventLoopFuture<AnalyzeResultResponse> {
        let promise = context.eventLoop.makePromise(of: AnalyzeResultResponse.self)

        Task {
            do {
                logger.info("Analyzing track", metadata: ["path": "\(request.path)"])

                // Verify file exists
                guard FileManager.default.fileExists(atPath: request.path) else {
                    promise.fail(GRPCStatus(code: .notFound, message: "File not found: \(request.path)"))
                    return
                }

                // Run analysis with progress logging
                let result = try await analyzer.analyze(path: request.path) { [weak self] progress in
                    self?.logger.debug("Analysis progress: \(progress)")
                }

                // Convert to proto response
                let response = AnalyzeResultResponse(from: result)
                promise.succeed(response)

                logger.info("Analysis complete", metadata: [
                    "path": "\(request.path)",
                    "bpm": "\(result.bpm)",
                    "key": "\(result.camelotKey)"
                ])

            } catch {
                logger.error("Analysis failed", metadata: ["error": "\(error)"])
                promise.fail(GRPCStatus(code: .internal, message: error.localizedDescription))
            }
        }

        return promise.futureResult
    }
}

// MARK: - Proto Message Types

/// Request message for AnalyzeTrack RPC
/// Mirrors proto: cartomix.analyzer.AnalyzeJob
public struct AnalyzeJobRequest: Message, Sendable {
    public static let protoMessageName = "cartomix.analyzer.AnalyzeJob"

    public var contentHash: String = ""
    public var path: String = ""
    public var targetSampleRate: Int32 = 48000
    public var mono: Bool = true
    public var dynamicTempo: Bool = false
    public var tempoFloor: Double = 60.0
    public var tempoCeil: Double = 180.0
    public var maxCues: Int32 = 8
    public var snapToDownbeat: Bool = true
    public var analysisVersion: Int32 = 1

    public init() {}

    public var unknownFields = SwiftProtobuf.UnknownStorage()

    public mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {
        while let fieldNumber = try decoder.nextFieldNumber() {
            switch fieldNumber {
            case 1: // TrackId (nested, extract content_hash)
                var nested = try decoder.decodeNestedMessage()
                while let nestedField = try nested.nextFieldNumber() {
                    if nestedField == 1 {
                        try nested.decodeSingularStringField(value: &contentHash)
                    } else if nestedField == 2 {
                        try nested.decodeSingularStringField(value: &path)
                    }
                }
            case 2:
                try decoder.decodeSingularStringField(value: &path)
            case 3: // DecodeParams
                var nested = try decoder.decodeNestedMessage()
                while let nestedField = try nested.nextFieldNumber() {
                    if nestedField == 1 {
                        try nested.decodeSingularInt32Field(value: &targetSampleRate)
                    } else if nestedField == 2 {
                        try nested.decodeSingularBoolField(value: &mono)
                    }
                }
            case 4: // BeatgridParams
                var nested = try decoder.decodeNestedMessage()
                while let nestedField = try nested.nextFieldNumber() {
                    if nestedField == 1 {
                        try nested.decodeSingularBoolField(value: &dynamicTempo)
                    } else if nestedField == 2 {
                        try nested.decodeSingularDoubleField(value: &tempoFloor)
                    } else if nestedField == 3 {
                        try nested.decodeSingularDoubleField(value: &tempoCeil)
                    }
                }
            case 5: // CueParams
                var nested = try decoder.decodeNestedMessage()
                while let nestedField = try nested.nextFieldNumber() {
                    if nestedField == 1 {
                        try nested.decodeSingularInt32Field(value: &maxCues)
                    } else if nestedField == 2 {
                        try nested.decodeSingularBoolField(value: &snapToDownbeat)
                    }
                }
            case 6:
                try decoder.decodeSingularInt32Field(value: &analysisVersion)
            default:
                break
            }
        }
    }

    public func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        if !contentHash.isEmpty || !path.isEmpty {
            try visitor.visitSingularMessageField(value: TrackIdMessage(contentHash: contentHash, path: path), fieldNumber: 1)
        }
        if !path.isEmpty {
            try visitor.visitSingularStringField(value: path, fieldNumber: 2)
        }
        try visitor.visitSingularMessageField(value: DecodeParamsMessage(sampleRate: targetSampleRate, mono: mono), fieldNumber: 3)
        try visitor.visitSingularMessageField(value: BeatgridParamsMessage(dynamic: dynamicTempo, floor: tempoFloor, ceil: tempoCeil), fieldNumber: 4)
        try visitor.visitSingularMessageField(value: CueParamsMessage(maxCues: maxCues, snap: snapToDownbeat), fieldNumber: 5)
        if analysisVersion != 0 {
            try visitor.visitSingularInt32Field(value: analysisVersion, fieldNumber: 6)
        }
    }

    public static func ==(lhs: AnalyzeJobRequest, rhs: AnalyzeJobRequest) -> Bool {
        return lhs.contentHash == rhs.contentHash &&
               lhs.path == rhs.path &&
               lhs.targetSampleRate == rhs.targetSampleRate &&
               lhs.mono == rhs.mono
    }
}

// Helper nested message types
struct TrackIdMessage: Message, Sendable {
    static let protoMessageName = "cartomix.common.TrackId"
    var contentHash: String
    var path: String
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularStringField(value: contentHash, fieldNumber: 1)
        try visitor.visitSingularStringField(value: path, fieldNumber: 2)
    }
    static func ==(lhs: TrackIdMessage, rhs: TrackIdMessage) -> Bool {
        lhs.contentHash == rhs.contentHash && lhs.path == rhs.path
    }
}

struct DecodeParamsMessage: Message, Sendable {
    static let protoMessageName = "cartomix.analyzer.DecodeParams"
    var sampleRate: Int32
    var mono: Bool
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularInt32Field(value: sampleRate, fieldNumber: 1)
        try visitor.visitSingularBoolField(value: mono, fieldNumber: 2)
    }
    static func ==(lhs: DecodeParamsMessage, rhs: DecodeParamsMessage) -> Bool {
        lhs.sampleRate == rhs.sampleRate && lhs.mono == rhs.mono
    }
}

struct BeatgridParamsMessage: Message, Sendable {
    static let protoMessageName = "cartomix.analyzer.BeatgridParams"
    var dynamic: Bool
    var floor: Double
    var ceil: Double
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularBoolField(value: dynamic, fieldNumber: 1)
        try visitor.visitSingularDoubleField(value: floor, fieldNumber: 2)
        try visitor.visitSingularDoubleField(value: ceil, fieldNumber: 3)
    }
    static func ==(lhs: BeatgridParamsMessage, rhs: BeatgridParamsMessage) -> Bool {
        lhs.dynamic == rhs.dynamic && lhs.floor == rhs.floor && lhs.ceil == rhs.ceil
    }
}

struct CueParamsMessage: Message, Sendable {
    static let protoMessageName = "cartomix.analyzer.CueParams"
    var maxCues: Int32
    var snap: Bool
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularInt32Field(value: maxCues, fieldNumber: 1)
        try visitor.visitSingularBoolField(value: snap, fieldNumber: 2)
    }
    static func ==(lhs: CueParamsMessage, rhs: CueParamsMessage) -> Bool {
        lhs.maxCues == rhs.maxCues && lhs.snap == rhs.snap
    }
}

/// Response message for AnalyzeTrack RPC
/// Mirrors proto: cartomix.analyzer.AnalyzeResult
public struct AnalyzeResultResponse: Message, Sendable {
    public static let protoMessageName = "cartomix.analyzer.AnalyzeResult"

    public var analysis: TrackAnalysisProto
    public var waveformTiles: Data = Data()

    public var unknownFields = SwiftProtobuf.UnknownStorage()

    public init() {
        self.analysis = TrackAnalysisProto()
    }

    public init(from result: TrackAnalysisResult) {
        self.analysis = TrackAnalysisProto(from: result)
        // Pack waveform summary into tiles
        self.waveformTiles = Data(result.waveformSummary.flatMap { value -> [UInt8] in
            var v = value
            return withUnsafeBytes(of: &v) { Array($0) }
        })
    }

    public mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {
        while let fieldNumber = try decoder.nextFieldNumber() {
            switch fieldNumber {
            case 1:
                try decoder.decodeSingularMessageField(value: &analysis)
            case 2:
                try decoder.decodeSingularBytesField(value: &waveformTiles)
            default:
                break
            }
        }
    }

    public func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularMessageField(value: analysis, fieldNumber: 1)
        if !waveformTiles.isEmpty {
            try visitor.visitSingularBytesField(value: waveformTiles, fieldNumber: 2)
        }
    }

    public static func ==(lhs: AnalyzeResultResponse, rhs: AnalyzeResultResponse) -> Bool {
        return lhs.analysis == rhs.analysis && lhs.waveformTiles == rhs.waveformTiles
    }
}

/// Proto representation of TrackAnalysis
/// Mirrors proto: cartomix.common.TrackAnalysis
public struct TrackAnalysisProto: Message, Sendable {
    public static let protoMessageName = "cartomix.common.TrackAnalysis"

    public var contentHash: String = ""
    public var path: String = ""
    public var durationSeconds: Double = 0
    public var bpm: Double = 0
    public var beatgridConfidence: Float = 0
    public var keyValue: String = ""
    public var keyConfidence: Float = 0
    public var energyGlobal: Int32 = 0
    public var cueCount: Int32 = 0
    public var sectionCount: Int32 = 0
    public var integratedLoudness: Float = 0
    public var truePeak: Float = 0
    public var loudnessRange: Float = 0
    public var openL3Vector: [Float] = []
    public var openL3WindowCount: Int32 = 0
    public var soundContext: String = ""
    public var soundContextConfidence: Float = 0
    public var hasQAFlags: Bool = false
    public var analysisVersion: Int32 = 1

    public var unknownFields = SwiftProtobuf.UnknownStorage()

    public init() {}

    public init(from result: TrackAnalysisResult) {
        // Generate content hash from path (simplified)
        self.contentHash = result.path.data(using: .utf8)?.base64EncodedString() ?? ""
        self.path = result.path
        self.durationSeconds = result.duration
        self.bpm = result.bpm
        self.beatgridConfidence = Float(result.beatgridConfidence)
        self.keyValue = result.camelotKey
        self.keyConfidence = Float(result.key.confidence)
        self.energyGlobal = Int32(result.globalEnergy)
        self.cueCount = Int32(result.cues.count)
        self.sectionCount = Int32(result.sections.count)
        self.integratedLoudness = result.loudness.integratedLoudness
        self.truePeak = result.loudness.truePeak
        self.loudnessRange = result.loudness.loudnessRange

        if let openL3 = result.openL3Embedding {
            self.openL3Vector = openL3.vector
            self.openL3WindowCount = Int32(openL3.windowCount)
        }

        if let sound = result.soundClassification {
            self.soundContext = sound.primaryContext
            self.soundContextConfidence = Float(sound.confidence)
            self.hasQAFlags = !sound.qaFlags.isEmpty
        }

        self.analysisVersion = 1
    }

    public mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {
        // Minimal decode - server is primarily write-side
    }

    public func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        // TrackId (field 1)
        try visitor.visitSingularMessageField(value: TrackIdMessage(contentHash: contentHash, path: path), fieldNumber: 1)

        if durationSeconds != 0 {
            try visitor.visitSingularDoubleField(value: durationSeconds, fieldNumber: 2)
        }

        // Beatgrid (field 3) - simplified
        try visitor.visitSingularMessageField(value: BeatgridProto(bpm: bpm, confidence: beatgridConfidence), fieldNumber: 3)

        // Key (field 4)
        try visitor.visitSingularMessageField(value: MusicalKeyProto(value: keyValue, confidence: keyConfidence), fieldNumber: 4)

        if energyGlobal != 0 {
            try visitor.visitSingularInt32Field(value: energyGlobal, fieldNumber: 5)
        }

        // Loudness (field 10)
        try visitor.visitSingularMessageField(value: LoudnessProto(integrated: integratedLoudness, truePeak: truePeak, range: loudnessRange), fieldNumber: 10)

        if analysisVersion != 0 {
            try visitor.visitSingularInt32Field(value: analysisVersion, fieldNumber: 12)
        }

        // OpenL3 (field 13)
        if !openL3Vector.isEmpty {
            try visitor.visitSingularMessageField(value: OpenL3EmbeddingProto(vector: openL3Vector, windowCount: openL3WindowCount), fieldNumber: 13)
        }

        if !soundContext.isEmpty {
            try visitor.visitSingularStringField(value: soundContext, fieldNumber: 15)
        }

        if soundContextConfidence != 0 {
            try visitor.visitSingularFloatField(value: soundContextConfidence, fieldNumber: 16)
        }

        if hasQAFlags {
            try visitor.visitSingularBoolField(value: hasQAFlags, fieldNumber: 17)
        }
    }

    public static func ==(lhs: TrackAnalysisProto, rhs: TrackAnalysisProto) -> Bool {
        lhs.contentHash == rhs.contentHash && lhs.path == rhs.path
    }
}

// Additional helper proto types
struct BeatgridProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.Beatgrid"
    var bpm: Double
    var confidence: Float
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        // tempo_map (field 2) - simplified with single node
        try visitor.visitRepeatedMessageField(value: [TempoNodeProto(beatIndex: 0, bpm: bpm)], fieldNumber: 2)
        try visitor.visitSingularFloatField(value: confidence, fieldNumber: 3)
    }
    static func ==(lhs: BeatgridProto, rhs: BeatgridProto) -> Bool {
        lhs.bpm == rhs.bpm && lhs.confidence == rhs.confidence
    }
}

struct TempoNodeProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.TempoMapNode"
    var beatIndex: Int32
    var bpm: Double
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularInt32Field(value: beatIndex, fieldNumber: 1)
        try visitor.visitSingularDoubleField(value: bpm, fieldNumber: 2)
    }
    static func ==(lhs: TempoNodeProto, rhs: TempoNodeProto) -> Bool {
        lhs.beatIndex == rhs.beatIndex && lhs.bpm == rhs.bpm
    }
}

struct MusicalKeyProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.MusicalKey"
    var value: String
    var confidence: Float
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularStringField(value: value, fieldNumber: 1)
        try visitor.visitSingularInt32Field(value: 2, fieldNumber: 2) // CAMELOT format
        try visitor.visitSingularFloatField(value: confidence, fieldNumber: 3)
    }
    static func ==(lhs: MusicalKeyProto, rhs: MusicalKeyProto) -> Bool {
        lhs.value == rhs.value
    }
}

struct LoudnessProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.Loudness"
    var integrated: Float
    var truePeak: Float
    var range: Float
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularFloatField(value: integrated, fieldNumber: 1)
        try visitor.visitSingularFloatField(value: truePeak, fieldNumber: 2)
        try visitor.visitSingularFloatField(value: range, fieldNumber: 5)
    }
    static func ==(lhs: LoudnessProto, rhs: LoudnessProto) -> Bool {
        lhs.integrated == rhs.integrated
    }
}

struct OpenL3EmbeddingProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.OpenL3Embedding"
    var vector: [Float]
    var windowCount: Int32
    var unknownFields = SwiftProtobuf.UnknownStorage()

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitPackedFloatField(value: vector, fieldNumber: 1)
        try visitor.visitSingularInt32Field(value: windowCount, fieldNumber: 2)
    }
    static func ==(lhs: OpenL3EmbeddingProto, rhs: OpenL3EmbeddingProto) -> Bool {
        lhs.windowCount == rhs.windowCount
    }
}

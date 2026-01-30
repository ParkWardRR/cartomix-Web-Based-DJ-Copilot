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
public final class AnalyzerGRPCServer: @unchecked Sendable {
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
final class AnalyzerWorkerProvider: CallHandlerProvider, @unchecked Sendable {
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

                // Run analysis (without progress callback to avoid Sendable issues)
                let result = try await analyzer.analyze(path: request.path, progress: nil)

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
                promise.fail(GRPCStatus(code: .internalError, message: error.localizedDescription))
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
            case 1: // TrackId (nested message)
                var trackId: TrackIdMessage?
                try decoder.decodeSingularMessageField(value: &trackId)
                if let tid = trackId {
                    self.contentHash = tid.contentHash
                    if !tid.path.isEmpty {
                        self.path = tid.path
                    }
                }
            case 2:
                try decoder.decodeSingularStringField(value: &path)
            case 3: // DecodeParams (nested message)
                var decodeParams: DecodeParamsMessage?
                try decoder.decodeSingularMessageField(value: &decodeParams)
                if let dp = decodeParams {
                    self.targetSampleRate = dp.sampleRate
                    self.mono = dp.mono
                }
            case 4: // BeatgridParams (nested message)
                var beatgridParams: BeatgridParamsMessage?
                try decoder.decodeSingularMessageField(value: &beatgridParams)
                if let bp = beatgridParams {
                    self.dynamicTempo = bp.dynamic
                    self.tempoFloor = bp.floor
                    self.tempoCeil = bp.ceil
                }
            case 5: // CueParams (nested message)
                var cueParams: CueParamsMessage?
                try decoder.decodeSingularMessageField(value: &cueParams)
                if let cp = cueParams {
                    self.maxCues = cp.maxCues
                    self.snapToDownbeat = cp.snap
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

    public func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? AnalyzeJobRequest else { return false }
        return self == other
    }
}

// Helper nested message types
struct TrackIdMessage: Message, Sendable {
    static let protoMessageName = "cartomix.common.TrackId"
    var contentHash: String = ""
    var path: String = ""
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(contentHash: String, path: String) {
        self.contentHash = contentHash
        self.path = path
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {
        while let fieldNumber = try decoder.nextFieldNumber() {
            switch fieldNumber {
            case 1: try decoder.decodeSingularStringField(value: &contentHash)
            case 2: try decoder.decodeSingularStringField(value: &path)
            default: break
            }
        }
    }
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        if !contentHash.isEmpty { try visitor.visitSingularStringField(value: contentHash, fieldNumber: 1) }
        if !path.isEmpty { try visitor.visitSingularStringField(value: path, fieldNumber: 2) }
    }
    static func ==(lhs: TrackIdMessage, rhs: TrackIdMessage) -> Bool {
        lhs.contentHash == rhs.contentHash && lhs.path == rhs.path
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? TrackIdMessage else { return false }
        return self == other
    }
}

struct DecodeParamsMessage: Message, Sendable {
    static let protoMessageName = "cartomix.analyzer.DecodeParams"
    var sampleRate: Int32 = 48000
    var mono: Bool = true
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(sampleRate: Int32, mono: Bool) {
        self.sampleRate = sampleRate
        self.mono = mono
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {
        while let fieldNumber = try decoder.nextFieldNumber() {
            switch fieldNumber {
            case 1: try decoder.decodeSingularInt32Field(value: &sampleRate)
            case 2: try decoder.decodeSingularBoolField(value: &mono)
            default: break
            }
        }
    }
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        if sampleRate != 0 { try visitor.visitSingularInt32Field(value: sampleRate, fieldNumber: 1) }
        if mono { try visitor.visitSingularBoolField(value: mono, fieldNumber: 2) }
    }
    static func ==(lhs: DecodeParamsMessage, rhs: DecodeParamsMessage) -> Bool {
        lhs.sampleRate == rhs.sampleRate && lhs.mono == rhs.mono
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? DecodeParamsMessage else { return false }
        return self == other
    }
}

struct BeatgridParamsMessage: Message, Sendable {
    static let protoMessageName = "cartomix.analyzer.BeatgridParams"
    var dynamic: Bool = false
    var floor: Double = 60.0
    var ceil: Double = 180.0
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(dynamic: Bool, floor: Double, ceil: Double) {
        self.dynamic = dynamic
        self.floor = floor
        self.ceil = ceil
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {
        while let fieldNumber = try decoder.nextFieldNumber() {
            switch fieldNumber {
            case 1: try decoder.decodeSingularBoolField(value: &dynamic)
            case 2: try decoder.decodeSingularDoubleField(value: &floor)
            case 3: try decoder.decodeSingularDoubleField(value: &ceil)
            default: break
            }
        }
    }
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        if dynamic { try visitor.visitSingularBoolField(value: dynamic, fieldNumber: 1) }
        if floor != 0 { try visitor.visitSingularDoubleField(value: floor, fieldNumber: 2) }
        if ceil != 0 { try visitor.visitSingularDoubleField(value: ceil, fieldNumber: 3) }
    }
    static func ==(lhs: BeatgridParamsMessage, rhs: BeatgridParamsMessage) -> Bool {
        lhs.dynamic == rhs.dynamic && lhs.floor == rhs.floor && lhs.ceil == rhs.ceil
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? BeatgridParamsMessage else { return false }
        return self == other
    }
}

struct CueParamsMessage: Message, Sendable {
    static let protoMessageName = "cartomix.analyzer.CueParams"
    var maxCues: Int32 = 8
    var snap: Bool = true
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(maxCues: Int32, snap: Bool) {
        self.maxCues = maxCues
        self.snap = snap
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {
        while let fieldNumber = try decoder.nextFieldNumber() {
            switch fieldNumber {
            case 1: try decoder.decodeSingularInt32Field(value: &maxCues)
            case 2: try decoder.decodeSingularBoolField(value: &snap)
            default: break
            }
        }
    }
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        if maxCues != 0 { try visitor.visitSingularInt32Field(value: maxCues, fieldNumber: 1) }
        if snap { try visitor.visitSingularBoolField(value: snap, fieldNumber: 2) }
    }
    static func ==(lhs: CueParamsMessage, rhs: CueParamsMessage) -> Bool {
        lhs.maxCues == rhs.maxCues && lhs.snap == rhs.snap
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? CueParamsMessage else { return false }
        return self == other
    }
}

/// Response message for AnalyzeTrack RPC
/// Mirrors proto: cartomix.analyzer.AnalyzeResult
public struct AnalyzeResultResponse: Message, Sendable {
    public static let protoMessageName = "cartomix.analyzer.AnalyzeResult"

    public var analysis: TrackAnalysisProto = TrackAnalysisProto()
    public var waveformTiles: Data = Data()

    public var unknownFields = SwiftProtobuf.UnknownStorage()

    public init() {}

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
                var optionalAnalysis: TrackAnalysisProto?
                try decoder.decodeSingularMessageField(value: &optionalAnalysis)
                if let a = optionalAnalysis {
                    self.analysis = a
                }
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

    public func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? AnalyzeResultResponse else { return false }
        return self == other
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

    public func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? TrackAnalysisProto else { return false }
        return self == other
    }
}

// Additional helper proto types
struct BeatgridProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.Beatgrid"
    var bpm: Double = 0
    var confidence: Float = 0
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(bpm: Double, confidence: Float) {
        self.bpm = bpm
        self.confidence = confidence
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        // tempo_map (field 2) - simplified with single node
        try visitor.visitRepeatedMessageField(value: [TempoNodeProto(beatIndex: 0, bpm: bpm)], fieldNumber: 2)
        try visitor.visitSingularFloatField(value: confidence, fieldNumber: 3)
    }
    static func ==(lhs: BeatgridProto, rhs: BeatgridProto) -> Bool {
        lhs.bpm == rhs.bpm && lhs.confidence == rhs.confidence
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? BeatgridProto else { return false }
        return self == other
    }
}

struct TempoNodeProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.TempoMapNode"
    var beatIndex: Int32 = 0
    var bpm: Double = 0
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(beatIndex: Int32, bpm: Double) {
        self.beatIndex = beatIndex
        self.bpm = bpm
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularInt32Field(value: beatIndex, fieldNumber: 1)
        try visitor.visitSingularDoubleField(value: bpm, fieldNumber: 2)
    }
    static func ==(lhs: TempoNodeProto, rhs: TempoNodeProto) -> Bool {
        lhs.beatIndex == rhs.beatIndex && lhs.bpm == rhs.bpm
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? TempoNodeProto else { return false }
        return self == other
    }
}

struct MusicalKeyProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.MusicalKey"
    var value: String = ""
    var confidence: Float = 0
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(value: String, confidence: Float) {
        self.value = value
        self.confidence = confidence
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularStringField(value: value, fieldNumber: 1)
        try visitor.visitSingularInt32Field(value: 2, fieldNumber: 2) // CAMELOT format
        try visitor.visitSingularFloatField(value: confidence, fieldNumber: 3)
    }
    static func ==(lhs: MusicalKeyProto, rhs: MusicalKeyProto) -> Bool {
        lhs.value == rhs.value
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? MusicalKeyProto else { return false }
        return self == other
    }
}

struct LoudnessProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.Loudness"
    var integrated: Float = 0
    var truePeak: Float = 0
    var range: Float = 0
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(integrated: Float, truePeak: Float, range: Float) {
        self.integrated = integrated
        self.truePeak = truePeak
        self.range = range
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitSingularFloatField(value: integrated, fieldNumber: 1)
        try visitor.visitSingularFloatField(value: truePeak, fieldNumber: 2)
        try visitor.visitSingularFloatField(value: range, fieldNumber: 5)
    }
    static func ==(lhs: LoudnessProto, rhs: LoudnessProto) -> Bool {
        lhs.integrated == rhs.integrated
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? LoudnessProto else { return false }
        return self == other
    }
}

struct OpenL3EmbeddingProto: Message, Sendable {
    static let protoMessageName = "cartomix.common.OpenL3Embedding"
    var vector: [Float] = []
    var windowCount: Int32 = 0
    var unknownFields = SwiftProtobuf.UnknownStorage()

    init() {}
    init(vector: [Float], windowCount: Int32) {
        self.vector = vector
        self.windowCount = windowCount
    }

    mutating func decodeMessage<D: SwiftProtobuf.Decoder>(decoder: inout D) throws {}
    func traverse<V: SwiftProtobuf.Visitor>(visitor: inout V) throws {
        try visitor.visitPackedFloatField(value: vector, fieldNumber: 1)
        try visitor.visitSingularInt32Field(value: windowCount, fieldNumber: 2)
    }
    static func ==(lhs: OpenL3EmbeddingProto, rhs: OpenL3EmbeddingProto) -> Bool {
        lhs.windowCount == rhs.windowCount
    }
    func isEqualTo(message: any Message) -> Bool {
        guard let other = message as? OpenL3EmbeddingProto else { return false }
        return self == other
    }
}

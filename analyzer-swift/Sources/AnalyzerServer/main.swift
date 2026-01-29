import ArgumentParser
import Foundation
import AnalyzerSwift
import Logging

/// CLI tool for the Swift audio analyzer
@main
struct AnalyzerServer: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        commandName: "analyzer-swift",
        abstract: "Algiers Audio Analyzer - Apple Silicon optimized",
        version: "0.1.0",
        subcommands: [Analyze.self, Serve.self, Healthcheck.self],
        defaultSubcommand: Analyze.self
    )
}

// MARK: - Analyze Command

struct Analyze: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        abstract: "Analyze a single audio file"
    )

    @Argument(help: "Path to the audio file")
    var path: String

    @Option(name: .shortAndLong, help: "Output format: json, summary")
    var format: OutputFormat = .json

    @Option(name: .shortAndLong, help: "Output file (default: stdout)")
    var output: String?

    @Flag(name: .shortAndLong, help: "Show progress")
    var progress = false

    func run() async throws {
        let logger = Logger(label: "analyzer")

        // Check for Metal/ANE hardware requirement
        do {
            try HardwareCheck.requireMetal()
        } catch let error as HardwareError {
            throw AnalyzerCLIError.hardwareNotSupported(error.description)
        }

        // Validate file exists
        guard FileManager.default.fileExists(atPath: path) else {
            throw AnalyzerCLIError.fileNotFound(path)
        }

        let analyzer = Analyzer(sampleRate: 48000, waveformBins: 200)

        let progressHandler: AnalysisProgressHandler?
        if progress {
            progressHandler = { stage in
                switch stage {
                case .decoding:
                    FileHandle.standardError.write("Decoding...\n".data(using: .utf8)!)
                case .beatgrid(let p):
                    if p == 0 {
                        FileHandle.standardError.write("Detecting beatgrid...\n".data(using: .utf8)!)
                    }
                case .key:
                    FileHandle.standardError.write("Detecting key...\n".data(using: .utf8)!)
                case .energy:
                    FileHandle.standardError.write("Analyzing energy...\n".data(using: .utf8)!)
                case .loudness:
                    FileHandle.standardError.write("Analyzing loudness (EBU R128)...\n".data(using: .utf8)!)
                case .sections:
                    FileHandle.standardError.write("Detecting sections...\n".data(using: .utf8)!)
                case .cues:
                    FileHandle.standardError.write("Generating cues...\n".data(using: .utf8)!)
                case .waveform:
                    FileHandle.standardError.write("Generating waveform...\n".data(using: .utf8)!)
                case .embedding:
                    FileHandle.standardError.write("Generating audio embedding...\n".data(using: .utf8)!)
                case .openL3Embedding:
                    FileHandle.standardError.write("Generating OpenL3 embedding (512-dim)...\n".data(using: .utf8)!)
                case .soundClassification:
                    FileHandle.standardError.write("Running sound classification (Apple ML)...\n".data(using: .utf8)!)
                case .complete:
                    FileHandle.standardError.write("Complete!\n".data(using: .utf8)!)
                }
            }
        } else {
            progressHandler = nil
        }

        let result = try await analyzer.analyze(path: path, progress: progressHandler)

        let outputData: Data
        switch format {
        case .json:
            outputData = try encodeJSON(result)
        case .summary:
            outputData = encodeSummary(result).data(using: .utf8)!
        }

        if let outputPath = output {
            try outputData.write(to: URL(fileURLWithPath: outputPath))
        } else {
            FileHandle.standardOutput.write(outputData)
        }
    }

    private func encodeJSON(_ result: TrackAnalysisResult) throws -> Data {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]

        let json = AnalysisJSON(from: result)
        return try encoder.encode(json)
    }

    private func encodeSummary(_ result: TrackAnalysisResult) -> String {
        var summary = """
        Track: \(result.path)
        Duration: \(String(format: "%.1f", result.duration))s
        BPM: \(String(format: "%.1f", result.bpm)) (confidence: \(String(format: "%.0f%%", result.beatgridConfidence * 100)))
        Key: \(result.key.name) / \(result.camelotKey) (confidence: \(String(format: "%.0f%%", result.key.confidence * 100)))
        Energy: \(result.globalEnergy)/10
        Loudness: \(String(format: "%.1f", result.loudness.integratedLoudness)) LUFS (range: \(String(format: "%.1f", result.loudness.loudnessRange)) LU, peak: \(String(format: "%.1f", result.loudness.truePeak)) dBTP)
        Sections: \(result.sections.count)
        Cues: \(result.cues.count)
        Embedding: \(result.embedding.vector.count)-dim vector (centroid: \(String(format: "%.0f", result.embedding.spectralCentroid)) Hz)
        """

        if let openL3 = result.openL3Embedding {
            summary += "\nOpenL3: \(openL3.vector.count)-dim vector (\(openL3.windowCount) windows)"
        }

        if let soundClass = result.soundClassification {
            summary += "\nSound Context: \(soundClass.primaryContext) (\(String(format: "%.0f%%", soundClass.confidence * 100)) confidence)"
            if !soundClass.qaFlags.isEmpty {
                summary += "\nQA Flags: \(soundClass.qaFlags.map { $0.type.rawValue }.joined(separator: ", "))"
            }
        }

        return summary
    }
}

enum OutputFormat: String, ExpressibleByArgument {
    case json
    case summary
}

// MARK: - Serve Command (HTTP/gRPC server)

struct Serve: AsyncParsableCommand {
    static let configuration = CommandConfiguration(
        abstract: "Start the analyzer server"
    )

    @Option(name: .shortAndLong, help: "Port to listen on")
    var port: Int = 9090

    @Option(name: .long, help: "Protocol: http, grpc")
    var proto: ServerProtocol = .http

    func run() async throws {
        let logger = Logger(label: "analyzer-server")

        // Check for Metal/ANE hardware requirement
        do {
            try HardwareCheck.requireMetal()
        } catch let error as HardwareError {
            throw AnalyzerCLIError.hardwareNotSupported(error.description)
        }

        logger.info("Starting analyzer server on port \(port) (\(proto))")

        // Start the appropriate server
        switch proto {
        case .http:
            try await startHTTPServer(port: port, logger: logger)
        case .grpc:
            try await startGRPCServer(port: port, logger: logger)
        }
    }

    private func startGRPCServer(port: Int, logger: Logger) async throws {
        logger.info("Starting gRPC server on port \(port)")
        logger.info("Service: cartomix.analyzer.AnalyzerWorker")
        logger.info("RPC: AnalyzeTrack - Analyze a track file")

        let server = AnalyzerGRPCServer(port: port, logger: logger)

        // Handle graceful shutdown
        let sigintSource = DispatchSource.makeSignalSource(signal: SIGINT, queue: .main)
        let sigtermSource = DispatchSource.makeSignalSource(signal: SIGTERM, queue: .main)

        signal(SIGINT, SIG_IGN)
        signal(SIGTERM, SIG_IGN)

        let shutdownHandler = {
            logger.info("Received shutdown signal")
            Task {
                try await server.stop()
            }
        }

        sigintSource.setEventHandler(handler: shutdownHandler)
        sigtermSource.setEventHandler(handler: shutdownHandler)
        sigintSource.resume()
        sigtermSource.resume()

        try await server.start()
    }

    private func startHTTPServer(port: Int, logger: Logger) async throws {
        // Use a simple socket-based HTTP server
        // This is a minimal implementation - in production use Vapor or similar

        let serverSocket = try createServerSocket(port: port)
        defer { close(serverSocket) }

        logger.info("HTTP server listening on port \(port)")
        logger.info("POST /analyze - Analyze a track")
        logger.info("GET /health - Health check")

        while true {
            let clientSocket = accept(serverSocket, nil, nil)
            guard clientSocket >= 0 else { continue }

            // Handle in background
            Task {
                await handleHTTPRequest(socket: clientSocket, logger: logger)
            }
        }
    }

    private func createServerSocket(port: Int) throws -> Int32 {
        let sock = socket(AF_INET, SOCK_STREAM, 0)
        guard sock >= 0 else {
            throw AnalyzerCLIError.socketError("Failed to create socket")
        }

        var yes: Int32 = 1
        setsockopt(sock, SOL_SOCKET, SO_REUSEADDR, &yes, socklen_t(MemoryLayout<Int32>.size))

        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = UInt16(port).bigEndian
        addr.sin_addr.s_addr = INADDR_ANY

        let bindResult = withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                bind(sock, $0, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }

        guard bindResult == 0 else {
            close(sock)
            throw AnalyzerCLIError.socketError("Failed to bind to port \(port)")
        }

        guard listen(sock, 10) == 0 else {
            close(sock)
            throw AnalyzerCLIError.socketError("Failed to listen")
        }

        return sock
    }

    private func handleHTTPRequest(socket: Int32, logger: Logger) async {
        defer { close(socket) }

        // Read request
        var buffer = [UInt8](repeating: 0, count: 65536)
        let bytesRead = read(socket, &buffer, buffer.count)
        guard bytesRead > 0 else { return }

        let request = String(bytes: buffer.prefix(bytesRead), encoding: .utf8) ?? ""
        let lines = request.components(separatedBy: "\r\n")
        guard let firstLine = lines.first else { return }

        let parts = firstLine.components(separatedBy: " ")
        guard parts.count >= 2 else { return }

        let method = parts[0]
        let path = parts[1]

        logger.debug("\(method) \(path)")

        // Route request
        let response: String
        if path == "/health" || path == "/healthz" {
            response = httpResponse(status: 200, body: #"{"status":"ok","version":"0.1.0"}"#)
        } else if path == "/analyze" && method == "POST" {
            // Extract body
            if let bodyStart = request.range(of: "\r\n\r\n") {
                let body = String(request[bodyStart.upperBound...])
                response = await handleAnalyzeRequest(body: body, logger: logger)
            } else {
                response = httpResponse(status: 400, body: #"{"error":"missing body"}"#)
            }
        } else {
            response = httpResponse(status: 404, body: #"{"error":"not found"}"#)
        }

        // Send response
        _ = response.withCString { ptr in
            write(socket, ptr, strlen(ptr))
        }
    }

    private func handleAnalyzeRequest(body: String, logger: Logger) async -> String {
        // Parse JSON body
        guard let data = body.data(using: .utf8),
              let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any],
              let path = json["path"] as? String else {
            return httpResponse(status: 400, body: #"{"error":"missing path"}"#)
        }

        guard FileManager.default.fileExists(atPath: path) else {
            return httpResponse(status: 404, body: #"{"error":"file not found"}"#)
        }

        do {
            let analyzer = Analyzer()
            let result = try await analyzer.analyze(path: path)
            let analysisJSON = AnalysisJSON(from: result)
            let encoder = JSONEncoder()
            encoder.outputFormatting = .sortedKeys
            let jsonData = try encoder.encode(analysisJSON)
            let jsonString = String(data: jsonData, encoding: .utf8) ?? "{}"
            return httpResponse(status: 200, body: jsonString)
        } catch {
            logger.error("Analysis failed: \(error)")
            return httpResponse(status: 500, body: #"{"error":"analysis failed: \#(error.localizedDescription)"}"#)
        }
    }

    private func httpResponse(status: Int, body: String) -> String {
        let statusText: String
        switch status {
        case 200: statusText = "OK"
        case 400: statusText = "Bad Request"
        case 404: statusText = "Not Found"
        case 500: statusText = "Internal Server Error"
        default: statusText = "Unknown"
        }

        return """
        HTTP/1.1 \(status) \(statusText)\r
        Content-Type: application/json\r
        Content-Length: \(body.utf8.count)\r
        Connection: close\r
        \r
        \(body)
        """
    }
}

enum ServerProtocol: String, ExpressibleByArgument {
    case http
    case grpc
}

// MARK: - Healthcheck Command

struct Healthcheck: ParsableCommand {
    static let configuration = CommandConfiguration(
        abstract: "Check analyzer health"
    )

    func run() throws {
        let capabilities = HardwareCheck.getCapabilities()

        let status: String
        if capabilities.metalAvailable {
            status = "ok"
        } else {
            status = "error"
        }

        let json: [String: Any] = [
            "status": status,
            "version": "0.1.0",
            "hardware": [
                "metal": capabilities.metalAvailable,
                "device": capabilities.deviceName ?? "none",
                "unified_memory": capabilities.hasUnifiedMemory,
                "ane": capabilities.aneAvailable,
                "compute_units": capabilities.recommendedComputeUnits.rawValue
            ]
        ]

        if let jsonData = try? JSONSerialization.data(withJSONObject: json, options: [.sortedKeys]),
           let jsonString = String(data: jsonData, encoding: .utf8) {
            print(jsonString)
        } else {
            print(#"{"status":"error","version":"0.1.0"}"#)
        }

        // Exit with error code if Metal not available
        if !capabilities.metalAvailable {
            throw AnalyzerCLIError.hardwareNotSupported(HardwareError.metalNotSupported.description)
        }
    }
}

// MARK: - JSON Encoding

struct AnalysisJSON: Codable {
    let path: String
    let duration: Double
    let bpm: Double
    let beatgridConfidence: Double
    let key: KeyJSON
    let energy: Int
    let loudness: LoudnessJSON
    let sections: [SectionJSON]
    let cues: [CueJSON]
    let waveformSummary: [Float]
    let embedding: EmbeddingJSON
    let openL3Embedding: OpenL3EmbeddingJSON?
    let soundClassification: SoundClassificationJSON?

    init(from result: TrackAnalysisResult) {
        self.path = result.path
        self.duration = result.duration
        self.bpm = result.bpm
        self.beatgridConfidence = result.beatgridConfidence
        self.key = KeyJSON(from: result.key)
        self.energy = result.globalEnergy
        self.loudness = LoudnessJSON(from: result.loudness)
        self.sections = result.sections.map { SectionJSON(from: $0) }
        self.cues = result.cues.map { CueJSON(from: $0) }
        self.waveformSummary = result.waveformSummary
        self.embedding = EmbeddingJSON(from: result.embedding)
        self.openL3Embedding = result.openL3Embedding.map { OpenL3EmbeddingJSON(from: $0) }
        self.soundClassification = result.soundClassification.map { SoundClassificationJSON(from: $0) }
    }
}

struct KeyJSON: Codable {
    let name: String
    let camelot: String
    let openKey: String
    let confidence: Double

    init(from key: MusicalKey) {
        self.name = key.name
        self.camelot = key.camelot
        self.openKey = key.openKey
        self.confidence = key.confidence
    }
}

struct SectionJSON: Codable {
    let type: String
    let startTime: Double
    let endTime: Double
    let startBeat: Int
    let endBeat: Int
    let confidence: Double

    init(from section: Section) {
        self.type = section.type.rawValue
        self.startTime = section.startTime
        self.endTime = section.endTime
        self.startBeat = section.startBeat
        self.endBeat = section.endBeat
        self.confidence = section.confidence
    }
}

struct CueJSON: Codable {
    let type: String
    let beatIndex: Int
    let time: Double
    let label: String
    let color: Int

    init(from cue: CuePoint) {
        self.type = cue.type.rawValue
        self.beatIndex = cue.beatIndex
        self.time = cue.time
        self.label = cue.label
        self.color = cue.color.rawValue
    }
}

struct LoudnessJSON: Codable {
    let integratedLUFS: Float
    let loudnessRange: Float
    let shortTermMax: Float
    let momentaryMax: Float
    let truePeak: Float
    let samplePeak: Float

    init(from loudness: LoudnessResult) {
        self.integratedLUFS = loudness.integratedLoudness
        self.loudnessRange = loudness.loudnessRange
        self.shortTermMax = loudness.shortTermMax
        self.momentaryMax = loudness.momentaryMax
        self.truePeak = loudness.truePeak
        self.samplePeak = loudness.samplePeak
    }
}

struct EmbeddingJSON: Codable {
    let vector: [Float]
    let spectralCentroid: Float
    let spectralRolloff: Float
    let zeroCrossingRate: Float
    let spectralFlatness: Float
    let tempoStability: Float
    let harmonicRatio: Float

    init(from embedding: AudioEmbedding) {
        self.vector = embedding.vector
        self.spectralCentroid = embedding.spectralCentroid
        self.spectralRolloff = embedding.spectralRolloff
        self.zeroCrossingRate = embedding.zeroCrossingRate
        self.spectralFlatness = embedding.spectralFlatness
        self.tempoStability = embedding.tempoStability
        self.harmonicRatio = embedding.harmonicRatio
    }
}

struct OpenL3EmbeddingJSON: Codable {
    let vector: [Float]
    let windowCount: Int
    let windows: [OpenL3WindowJSON]?

    init(from embedding: OpenL3Embedding) {
        self.vector = embedding.vector
        self.windowCount = embedding.windowCount
        // Only include first 10 windows to keep JSON size reasonable
        self.windows = embedding.windows.prefix(10).map { OpenL3WindowJSON(from: $0) }
    }
}

struct OpenL3WindowJSON: Codable {
    let timestamp: Double
    let duration: Double
    // Omit vector to save space - use track-level vector for similarity

    init(from window: OpenL3WindowEmbedding) {
        self.timestamp = window.timestamp
        self.duration = window.duration
    }
}

struct SoundClassificationJSON: Codable {
    let primaryContext: String
    let confidence: Double
    let events: [SoundEventJSON]
    let qaFlags: [QAFlagJSON]

    init(from result: SoundClassificationResult) {
        self.primaryContext = result.primaryContext
        self.confidence = result.confidence
        self.events = result.events.map { SoundEventJSON(from: $0) }
        self.qaFlags = result.qaFlags.map { QAFlagJSON(from: $0) }
    }
}

struct SoundEventJSON: Codable {
    let label: String
    let category: String
    let confidence: Float
    let startTime: Double
    let endTime: Double

    init(from event: SoundEvent) {
        self.label = event.label
        self.category = event.category
        self.confidence = event.confidence
        self.startTime = event.startTime
        self.endTime = event.endTime
    }
}

struct QAFlagJSON: Codable {
    let type: String
    let reason: String
    let severity: Int

    init(from flag: QAFlag) {
        self.type = flag.type.rawValue
        self.reason = flag.reason
        self.severity = flag.severity
    }
}

// MARK: - Errors

enum AnalyzerCLIError: Error, CustomStringConvertible {
    case fileNotFound(String)
    case socketError(String)
    case notImplemented(String)
    case hardwareNotSupported(String)

    var description: String {
        switch self {
        case .fileNotFound(let path):
            return "File not found: \(path)"
        case .socketError(let msg):
            return "Socket error: \(msg)"
        case .notImplemented(let feature):
            return "Not implemented: \(feature)"
        case .hardwareNotSupported(let msg):
            return "Hardware not supported:\n\(msg)"
        }
    }
}

import Foundation
import os

actor ProcessManager {
    private var engineProcess: Process?
    private var analyzerProcess: Process?
    private let logger = Logger(subsystem: "com.algiers.app", category: "ProcessManager")

    func startServices() async {
        logger.info("Starting Algiers services...")

        // Find helper binaries in app bundle
        guard let helpersURL = Bundle.main.url(forResource: "Helpers", withExtension: nil, subdirectory: nil) ??
              Bundle.main.privateFrameworksURL?.deletingLastPathComponent().appendingPathComponent("Helpers") else {
            logger.error("Helpers directory not found in bundle")
            return
        }

        // Find web root for static files
        let webRoot = Bundle.main.resourceURL?.appendingPathComponent("web").path ?? ""

        // Setup data directory in Application Support
        let appSupport = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first!
        let dataDir = appSupport.appendingPathComponent("Algiers")
        try? FileManager.default.createDirectory(at: dataDir, withIntermediateDirectories: true)

        // Start analyzer first
        await startAnalyzer(helpersURL: helpersURL)

        // Wait a moment for analyzer to initialize
        try? await Task.sleep(nanoseconds: 500_000_000)

        // Start engine
        await startEngine(helpersURL: helpersURL, dataDir: dataDir.path, webRoot: webRoot)
    }

    private func startEngine(helpersURL: URL, dataDir: String, webRoot: String) async {
        let enginePath = helpersURL.appendingPathComponent("algiers-engine").path

        guard FileManager.default.fileExists(atPath: enginePath) else {
            logger.error("Engine binary not found at \(enginePath)")
            return
        }

        let process = Process()
        process.executableURL = URL(fileURLWithPath: enginePath)
        process.arguments = [
            "--http-port", "8080",
            "--port", "50051",
            "--data-dir", dataDir,
            "--web-root", webRoot,
            "--analyzer-addr", "localhost:50052",
            "--log-level", "info"
        ]

        // Capture output
        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = pipe

        pipe.fileHandleForReading.readabilityHandler = { [weak self] handle in
            let data = handle.availableData
            if let output = String(data: data, encoding: .utf8), !output.isEmpty {
                self?.logger.info("Engine: \(output.trimmingCharacters(in: .whitespacesAndNewlines))")
            }
        }

        do {
            try process.run()
            engineProcess = process
            logger.info("Engine started with PID \(process.processIdentifier)")
        } catch {
            logger.error("Failed to start engine: \(error.localizedDescription)")
        }
    }

    private func startAnalyzer(helpersURL: URL) async {
        let analyzerPath = helpersURL.appendingPathComponent("analyzer-swift").path

        guard FileManager.default.fileExists(atPath: analyzerPath) else {
            logger.warning("Analyzer binary not found at \(analyzerPath) - using CPU fallback")
            return
        }

        let process = Process()
        process.executableURL = URL(fileURLWithPath: analyzerPath)
        process.arguments = ["serve", "--port", "50052", "--proto", "grpc"]

        let pipe = Pipe()
        process.standardOutput = pipe
        process.standardError = pipe

        pipe.fileHandleForReading.readabilityHandler = { [weak self] handle in
            let data = handle.availableData
            if let output = String(data: data, encoding: .utf8), !output.isEmpty {
                self?.logger.info("Analyzer: \(output.trimmingCharacters(in: .whitespacesAndNewlines))")
            }
        }

        do {
            try process.run()
            analyzerProcess = process
            logger.info("Analyzer started with PID \(process.processIdentifier)")
        } catch {
            logger.error("Failed to start analyzer: \(error.localizedDescription)")
        }
    }

    nonisolated func stopServices() {
        Task {
            await stopServicesAsync()
        }
    }

    private func stopServicesAsync() {
        logger.info("Stopping Algiers services...")

        if let engine = engineProcess, engine.isRunning {
            engine.interrupt()
            engine.waitUntilExit()
            logger.info("Engine stopped")
        }

        if let analyzer = analyzerProcess, analyzer.isRunning {
            analyzer.interrupt()
            analyzer.waitUntilExit()
            logger.info("Analyzer stopped")
        }
    }
}

import Foundation

public struct AnalyzerConfig {
    public var targetSampleRate: Int
    public var allowDynamicGrid: Bool

    public init(targetSampleRate: Int = 48_000, allowDynamicGrid: Bool = true) {
        self.targetSampleRate = targetSampleRate
        self.allowDynamicGrid = allowDynamicGrid
    }
}

// Stub analyzer entry point; will call Accelerate/Core ML pipelines.
public final class Analyzer {
    private let config: AnalyzerConfig

    public init(config: AnalyzerConfig = AnalyzerConfig()) {
        self.config = config
    }

    public func analyze(path: String) throws {
        // TODO: wire to actual DSP + ML once protobuf contracts are generated.
        print("Analyzer stub running on", path, "config", config)
    }
}

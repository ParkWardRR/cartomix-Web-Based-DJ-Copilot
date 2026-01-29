import Metal
import Foundation

/// Hardware capability errors
public enum HardwareError: Error, CustomStringConvertible {
    case metalNotSupported
    case aneNotAvailable(reason: String)

    public var description: String {
        switch self {
        case .metalNotSupported:
            return """
            Metal GPU acceleration is required but not available on this system.

            Algiers requires Apple Silicon or a Mac with Metal-capable GPU for:
            - Core ML inference (OpenL3 embeddings)
            - Real-time audio analysis
            - ANE (Apple Neural Engine) acceleration

            Supported hardware:
            - Apple Silicon Macs (M1, M2, M3, M4 family)
            - Intel Macs with Metal-capable AMD/Intel GPU (limited support)

            Please run Algiers on a supported Mac.
            """
        case .aneNotAvailable(let reason):
            return "Apple Neural Engine unavailable: \(reason)"
        }
    }
}

/// Hardware capability checker for ML inference requirements
public struct HardwareCheck {

    /// Check if Metal is available (required for Core ML inference)
    /// - Throws: HardwareError.metalNotSupported if no Metal device found
    public static func requireMetal() throws {
        guard let device = MTLCreateSystemDefaultDevice() else {
            throw HardwareError.metalNotSupported
        }

        // Log the device for debugging
        let deviceName = device.name
        let hasUnifiedMemory = device.hasUnifiedMemory
        let registryID = device.registryID

        print("Metal device: \(deviceName)")
        print("Unified memory: \(hasUnifiedMemory)")
        print("Registry ID: \(registryID)")

        // Apple Silicon has unified memory
        if hasUnifiedMemory {
            print("ANE acceleration: Available (Apple Silicon)")
        } else {
            print("ANE acceleration: Not available (discrete GPU)")
        }
    }

    /// Check hardware and return capabilities without throwing
    public static func getCapabilities() -> HardwareCapabilities {
        guard let device = MTLCreateSystemDefaultDevice() else {
            return HardwareCapabilities(
                metalAvailable: false,
                deviceName: nil,
                hasUnifiedMemory: false,
                aneAvailable: false,
                recommendedComputeUnits: .cpuOnly
            )
        }

        let hasUnifiedMemory = device.hasUnifiedMemory

        return HardwareCapabilities(
            metalAvailable: true,
            deviceName: device.name,
            hasUnifiedMemory: hasUnifiedMemory,
            aneAvailable: hasUnifiedMemory, // ANE only on Apple Silicon
            recommendedComputeUnits: hasUnifiedMemory ? .all : .cpuAndGPU
        )
    }
}

/// Hardware capabilities summary
public struct HardwareCapabilities: Sendable {
    /// Metal GPU is available
    public let metalAvailable: Bool

    /// GPU device name
    public let deviceName: String?

    /// Unified memory (Apple Silicon)
    public let hasUnifiedMemory: Bool

    /// Apple Neural Engine available
    public let aneAvailable: Bool

    /// Recommended Core ML compute units
    public let recommendedComputeUnits: ComputeUnits

    public enum ComputeUnits: String, Sendable {
        case all = "all"           // CPU + GPU + ANE
        case cpuAndGPU = "cpuAndGPU"
        case cpuOnly = "cpuOnly"
    }
}

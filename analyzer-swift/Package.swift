// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "AnalyzerSwift",
    platforms: [
        .macOS(.v15)
    ],
    products: [
        .executable(
            name: "analyzer-swift",
            targets: ["AnalyzerServer"]),
        .library(
            name: "AnalyzerSwift",
            targets: ["AnalyzerSwift"])
    ],
    dependencies: [
        .package(url: "https://github.com/grpc/grpc-swift.git", from: "1.23.0"),
        .package(url: "https://github.com/apple/swift-protobuf.git", from: "1.28.0"),
        .package(url: "https://github.com/apple/swift-argument-parser.git", from: "1.5.0"),
        .package(url: "https://github.com/apple/swift-log.git", from: "1.6.0"),
    ],
    targets: [
        .executableTarget(
            name: "AnalyzerServer",
            dependencies: [
                "AnalyzerSwift",
                .product(name: "GRPC", package: "grpc-swift"),
                .product(name: "ArgumentParser", package: "swift-argument-parser"),
                .product(name: "Logging", package: "swift-log"),
            ]),
        .target(
            name: "AnalyzerSwift",
            dependencies: [
                .product(name: "SwiftProtobuf", package: "swift-protobuf"),
            ]),
        .testTarget(
            name: "AnalyzerSwiftTests",
            dependencies: ["AnalyzerSwift"])
    ]
)

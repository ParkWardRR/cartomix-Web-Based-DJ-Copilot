// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "AnalyzerSwift",
    platforms: [
        .macOS(.v15)
    ],
    products: [
        .library(
            name: "AnalyzerSwift",
            targets: ["AnalyzerSwift"])
    ],
    targets: [
        .target(
            name: "AnalyzerSwift",
            dependencies: []),
        .testTarget(
            name: "AnalyzerSwiftTests",
            dependencies: ["AnalyzerSwift"])
    ]
)

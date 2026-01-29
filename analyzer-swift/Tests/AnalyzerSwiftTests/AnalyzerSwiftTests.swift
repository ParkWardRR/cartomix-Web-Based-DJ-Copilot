import XCTest
@testable import AnalyzerSwift

final class AnalyzerSwiftTests: XCTestCase {
    func testStubAnalyzerRuns() throws {
        let analyzer = Analyzer()
        XCTAssertNoThrow(try analyzer.analyze(path: "/tmp/example.wav"))
    }
}

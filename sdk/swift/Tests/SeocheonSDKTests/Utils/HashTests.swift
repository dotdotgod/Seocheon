import XCTest
@testable import SeocheonSDK

final class HashTests: XCTestCase {

    func testVerifyValidHash() {
        let hash = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
        XCTAssertTrue(HashUtils.verifyActivityHash(hash))
    }

    func testVerifyAllZeroHash() {
        let hash = String(repeating: "0", count: 64)
        XCTAssertTrue(HashUtils.verifyActivityHash(hash))
    }

    func testVerifyAllFHash() {
        let hash = String(repeating: "f", count: 64)
        XCTAssertTrue(HashUtils.verifyActivityHash(hash))
    }

    func testVerifyTooShort() {
        let hash = String(repeating: "a", count: 63)
        XCTAssertFalse(HashUtils.verifyActivityHash(hash))
    }

    func testVerifyTooLong() {
        let hash = String(repeating: "a", count: 65)
        XCTAssertFalse(HashUtils.verifyActivityHash(hash))
    }

    func testVerifyNonHexCharacters() {
        var hash = String(repeating: "a", count: 63) + "g"
        XCTAssertFalse(HashUtils.verifyActivityHash(hash))
        hash = String(repeating: "a", count: 63) + "G"
        XCTAssertFalse(HashUtils.verifyActivityHash(hash))
    }

    func testVerifyEmpty() {
        XCTAssertFalse(HashUtils.verifyActivityHash(""))
    }

    func testComputeActivityHash() {
        let data = Data("hello world".utf8)
        let hash = HashUtils.computeActivityHash(data)
        XCTAssertEqual(hash.count, 64)
        XCTAssertTrue(HashUtils.verifyActivityHash(hash))
        // SHA-256 of "hello world"
        XCTAssertEqual(hash, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9")
    }
}

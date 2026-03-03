import XCTest
@testable import SeocheonSDK

final class KeysTests: XCTestCase {

    func testPrivateKeyFromBytes() throws {
        let bytes = Data(repeating: 0x01, count: 32)
        let key = try PrivateKey(rawBytes: bytes)
        XCTAssertEqual(key.bytes.count, 32)
    }

    func testPublicKeyIs33Bytes() throws {
        let bytes = Data(repeating: 0x01, count: 32)
        let key = try PrivateKey(rawBytes: bytes)
        let pubKey = key.pubKey
        XCTAssertEqual(pubKey.count, 33)
        XCTAssertTrue(pubKey[0] == 0x02 || pubKey[0] == 0x03)
    }

    func testSignProduces64Bytes() throws {
        let bytes = Data(repeating: 0x01, count: 32)
        let key = try PrivateKey(rawBytes: bytes)
        let message = Data("test message".utf8)
        let signature = try key.sign(message)
        XCTAssertEqual(signature.count, 64)
    }

    func testDifferentMessagesProduceDifferentSignatures() throws {
        let bytes = Data(repeating: 0x01, count: 32)
        let key = try PrivateKey(rawBytes: bytes)
        let sig1 = try key.sign(Data("message1".utf8))
        let sig2 = try key.sign(Data("message2".utf8))
        XCTAssertNotEqual(sig1, sig2)
    }

    func testInvalidKeySize() {
        XCTAssertThrowsError(try PrivateKey(rawBytes: Data(repeating: 0x01, count: 16)))
    }
}

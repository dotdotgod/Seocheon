import XCTest
@testable import SeocheonSDK

final class Bip44Tests: XCTestCase {

    func testDeriveKeyFromMnemonic() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        XCTAssertEqual(key.bytes.count, 32)
    }

    func testDeriveKeyConsistency() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key1 = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let key2 = try Bip44.deriveKeyFromMnemonic(mnemonic)
        XCTAssertEqual(key1.bytes, key2.bytes)
    }

    func testDeriveKeyDifferentMnemonics() throws {
        let key1 = try Bip44.deriveKeyFromMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
        let key2 = try Bip44.deriveKeyFromMnemonic("zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong")
        XCTAssertNotEqual(key1.bytes, key2.bytes)
    }

    func testDeriveKeyProducesValidPublicKey() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let pubKey = key.pubKey
        XCTAssertEqual(pubKey.count, 33)
    }
}

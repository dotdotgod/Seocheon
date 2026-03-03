import XCTest
@testable import SeocheonSDK

final class AddressTests: XCTestCase {

    func testAddressFromPubKey() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let address = try Address.fromPubKey(key.pubKey)
        XCTAssertTrue(address.hasPrefix("seocheon1"))
        XCTAssertTrue(address.count > 10)
    }

    func testAddressConsistency() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let addr1 = try Address.fromPubKey(key.pubKey)
        let addr2 = try Address.fromPubKey(key.pubKey)
        XCTAssertEqual(addr1, addr2)
    }

    func testDifferentKeysProduceDifferentAddresses() throws {
        let key1 = try Bip44.deriveKeyFromMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
        let key2 = try Bip44.deriveKeyFromMnemonic("zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong")
        let addr1 = try Address.fromPubKey(key1.pubKey)
        let addr2 = try Address.fromPubKey(key2.pubKey)
        XCTAssertNotEqual(addr1, addr2)
    }
}

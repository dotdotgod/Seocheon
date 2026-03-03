import XCTest
@testable import SeocheonSDK

final class SigningTests: XCTestCase {

    func testDirectServiceCreation() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let service = try DirectSigningService(mnemonic: mnemonic)
        XCTAssertFalse(service.getAddress().isEmpty)
        XCTAssertTrue(service.getAddress().hasPrefix("seocheon1"))
    }

    func testDirectServicePubKey() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let service = try DirectSigningService(mnemonic: mnemonic)
        let pubKey = service.getPubKey()
        XCTAssertEqual(pubKey.count, 33)
    }

    func testDirectServiceSign() async throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let service = try DirectSigningService(mnemonic: mnemonic)
        let sig = try await service.sign(Data("test".utf8))
        XCTAssertEqual(sig.count, 64)
    }

    func testDirectServiceConsistentAddress() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let s1 = try DirectSigningService(mnemonic: mnemonic)
        let s2 = try DirectSigningService(mnemonic: mnemonic)
        XCTAssertEqual(s1.getAddress(), s2.getAddress())
    }

    func testDirectServiceConsistentPubKey() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let s1 = try DirectSigningService(mnemonic: mnemonic)
        let s2 = try DirectSigningService(mnemonic: mnemonic)
        XCTAssertEqual(s1.getPubKey(), s2.getPubKey())
    }

    func testMockSignerAddress() {
        let signer = MockSigner(address: "seocheon1test")
        XCTAssertEqual(signer.getAddress(), "seocheon1test")
    }

    func testMockSignerPubKey() {
        let signer = MockSigner()
        XCTAssertEqual(signer.getPubKey().count, 33)
    }

    func testMockSignerSign() async throws {
        let signer = MockSigner()
        let sig = try await signer.sign(Data("test".utf8))
        XCTAssertEqual(sig.count, 64)
    }
}

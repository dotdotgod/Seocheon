import XCTest
@testable import SeocheonSDK

final class VaultTests: XCTestCase {

    func testVaultServiceCreation() {
        let service = VaultSigningService(endpoint: "http://vault:8200", keyName: "test-key")
        XCTAssertNotNil(service)
    }

    func testVaultServiceAddress() {
        let service = VaultSigningService(endpoint: "http://vault:8200", keyName: "test-key")
        // Without connecting, address should be empty
        let addr = service.getAddress()
        XCTAssertTrue(addr.isEmpty)
    }

    func testVaultServicePubKey() {
        let service = VaultSigningService(endpoint: "http://vault:8200", keyName: "test-key")
        let pubKey = service.getPubKey()
        XCTAssertTrue(pubKey.isEmpty)
    }
}

import Foundation
@testable import SeocheonSDK

final class MockSigner: SigningService, @unchecked Sendable {
    private let address: String
    private let pubKey: Data

    init(address: String = "seocheon1mockaddr", pubKey: Data = Data(repeating: 0x02, count: 33)) {
        self.address = address
        self.pubKey = pubKey
    }

    func sign(_ txBytes: Data) async throws -> Data {
        return Data(repeating: 0xAA, count: 64)
    }

    func getAddress() -> String {
        return address
    }

    func getPubKey() -> Data {
        return pubKey
    }
}

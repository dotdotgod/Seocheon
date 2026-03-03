import Foundation

/// Signs transactions using a mnemonic directly (test/development only).
public final class DirectSigningService: SigningService, @unchecked Sendable {
    private let privKey: PrivateKey
    private let address: String
    private let pubKeyData: Data

    public init(mnemonic: String) throws {
        guard !mnemonic.isEmpty else {
            throw SDKError.signingFailed("mnemonic is required")
        }
        self.privKey = try Bip44.deriveKeyFromMnemonic(mnemonic)
        self.pubKeyData = privKey.pubKey
        self.address = try Address.fromPubKey(pubKeyData)
    }

    public func sign(_ txBytes: Data) async throws -> Data {
        return try privKey.sign(txBytes)
    }

    public func getAddress() -> String {
        return address
    }

    public func getPubKey() -> Data {
        return pubKeyData
    }
}

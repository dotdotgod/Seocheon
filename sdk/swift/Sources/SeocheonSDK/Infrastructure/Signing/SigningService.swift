import Foundation

/// Protocol for transaction signing services.
public protocol SigningService: Sendable {
    /// Signs the given transaction bytes and returns the signature.
    func sign(_ txBytes: Data) async throws -> Data
    /// Returns the signer's bech32 address.
    func getAddress() -> String
    /// Returns the signer's compressed public key bytes.
    func getPubKey() -> Data
}

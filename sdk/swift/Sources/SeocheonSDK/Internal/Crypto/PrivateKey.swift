import Foundation
import CryptoKit
import secp256k1

/// Wraps a secp256k1 private key for signing operations.
internal final class PrivateKey {
    private let key: secp256k1.Signing.PrivateKey

    init(rawBytes: Data) throws {
        guard rawBytes.count == 32 else {
            throw SDKError.signingFailed("private key must be 32 bytes, got \(rawBytes.count)")
        }
        self.key = try secp256k1.Signing.PrivateKey(dataRepresentation: rawBytes)
    }

    init(key: secp256k1.Signing.PrivateKey) {
        self.key = key
    }

    /// Returns the compressed 33-byte public key.
    var pubKey: Data {
        return Data(key.publicKey.dataRepresentation)
    }

    /// Returns the raw 32-byte private key.
    var bytes: Data {
        return Data(key.dataRepresentation)
    }

    /// Signs the given data with SHA-256 hashing + secp256k1.
    /// Returns a 64-byte compact signature (R || S).
    func sign(_ data: Data) throws -> Data {
        let hash = SHA256.hash(data: data)
        let hashData = Data(hash)
        let sig = try key.signature(for: hashData)
        let compactSig = sig.dataRepresentation
        // secp256k1.swift returns 64-byte compact R||S
        guard compactSig.count == 64 else {
            throw SDKError.signingFailed("unexpected signature length: \(compactSig.count)")
        }
        return compactSig
    }

    /// Verifies a 64-byte compact signature against data using the public key.
    static func verify(pubKeyBytes: Data, data: Data, signature: Data) -> Bool {
        guard signature.count == 64 else { return false }
        do {
            let pubKey = try secp256k1.Signing.PublicKey(dataRepresentation: pubKeyBytes, format: .compressed)
            let hash = SHA256.hash(data: data)
            let hashData = Data(hash)
            let ecdsaSig = try secp256k1.Signing.ECDSASignature(dataRepresentation: signature)
            return pubKey.isValidSignature(ecdsaSig, for: hashData)
        } catch {
            return false
        }
    }
}

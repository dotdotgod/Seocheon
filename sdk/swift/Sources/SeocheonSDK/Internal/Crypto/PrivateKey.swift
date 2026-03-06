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
    /// Returns a 64-byte compact signature (R || S) in big-endian format.
    ///
    /// secp256k1.swift uses a 4×64-bit scalar representation (scalar_4x64),
    /// so ECDSASignature.dataRepresentation stores the raw little-endian limb format,
    /// NOT the compact big-endian R||S that Cosmos SDK expects.
    /// We must call compactRepresentation (serialize_compact) to get correct bytes.
    ///
    /// Signing uses the Digest overload to avoid double-hashing:
    /// secp256k1.swift has two overloads — Digest (signs hash directly) and
    /// DataProtocol (applies SHA256 then signs). Passing SHA256.Digest uses the
    /// Digest overload, giving secp256k1_sign(SHA256(data)) — one hash, correct.
    func sign(_ data: Data) throws -> Data {
        let digest = SHA256.hash(data: data)
        let sig = try key.signature(for: digest)
        // Use compactRepresentation (serialize_compact) to get big-endian R||S bytes
        let compactSig = try sig.compactRepresentation
        guard compactSig.count == 64 else {
            throw SDKError.signingFailed("unexpected signature length: \(compactSig.count)")
        }
        return compactSig
    }

    /// Verifies a 64-byte compact big-endian signature against data using the public key.
    static func verify(pubKeyBytes: Data, data: Data, signature: Data) -> Bool {
        guard signature.count == 64 else { return false }
        do {
            let pubKey = try secp256k1.Signing.PublicKey(dataRepresentation: pubKeyBytes, format: .compressed)
            let digest = SHA256.hash(data: data)
            // Parse compact (big-endian R||S) signature — not dataRepresentation (internal limb format)
            let ecdsaSig = try secp256k1.Signing.ECDSASignature(compactRepresentation: signature)
            return pubKey.isValidSignature(ecdsaSig, for: digest)
        } catch {
            return false
        }
    }
}

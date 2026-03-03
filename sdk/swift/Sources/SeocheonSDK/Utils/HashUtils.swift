import Foundation
import CryptoKit

/// Hash-related utility functions.
public enum HashUtils {
    /// Validates that the given string is a valid activity hash (64 hex characters, 32 bytes SHA-256).
    public static func verifyActivityHash(_ hash: String) -> Bool {
        guard hash.count == 64 else { return false }
        return hash.allSatisfy { $0.isHexDigit }
    }

    /// Computes a SHA-256 hash and returns a 64-character hex string.
    public static func computeActivityHash(_ data: Data) -> String {
        let digest = SHA256.hash(data: data)
        return digest.map { String(format: "%02x", $0) }.joined()
    }
}

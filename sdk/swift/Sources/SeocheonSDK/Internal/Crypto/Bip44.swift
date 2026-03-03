import Foundation
import CryptoKit

/// BIP39/BIP44 key derivation for Cosmos (m/44'/118'/0'/0/0).
internal enum Bip44 {
    /// Derives a secp256k1 private key from a BIP39 mnemonic.
    /// Uses the Cosmos BIP44 path: m/44'/118'/0'/0/0
    static func deriveKeyFromMnemonic(_ mnemonic: String) throws -> PrivateKey {
        let words = mnemonic.split(separator: " ").map(String.init)
        guard words.count == 12 || words.count == 15 || words.count == 18 ||
              words.count == 21 || words.count == 24 else {
            throw SDKError.signingFailed("invalid mnemonic: expected 12/15/18/21/24 words, got \(words.count)")
        }

        // BIP39: mnemonic -> seed (PBKDF2-HMAC-SHA512)
        let seed = try mnemonicToSeed(mnemonic, passphrase: "")

        // BIP32: seed -> master key
        let (masterKey, masterChainCode) = try hmacSHA512(key: Data("Bitcoin seed".utf8), data: seed)

        // Derive m/44'/118'/0'/0/0
        let purpose = try deriveChild(key: masterKey, chainCode: masterChainCode, index: 44 + 0x80000000)
        let coinType = try deriveChild(key: purpose.key, chainCode: purpose.chainCode, index: 118 + 0x80000000)
        let account = try deriveChild(key: coinType.key, chainCode: coinType.chainCode, index: 0 + 0x80000000)
        let change = try deriveChild(key: account.key, chainCode: account.chainCode, index: 0)
        let address = try deriveChild(key: change.key, chainCode: change.chainCode, index: 0)

        return try PrivateKey(rawBytes: address.key)
    }

    // MARK: - Private

    private static func mnemonicToSeed(_ mnemonic: String, passphrase: String) throws -> Data {
        let password = Data(mnemonic.decomposedStringWithCompatibilityMapping.utf8)
        let salt = Data(("mnemonic" + passphrase).decomposedStringWithCompatibilityMapping.utf8)
        return try pbkdf2HMACSHA512(password: password, salt: salt, iterations: 2048, keyLength: 64)
    }

    private static func pbkdf2HMACSHA512(password: Data, salt: Data, iterations: Int, keyLength: Int) throws -> Data {
        var result = Data()
        let blocks = (keyLength + 63) / 64

        for blockIndex in 1...blocks {
            var blockIndexBytes = Data(count: 4)
            blockIndexBytes[0] = UInt8((blockIndex >> 24) & 0xFF)
            blockIndexBytes[1] = UInt8((blockIndex >> 16) & 0xFF)
            blockIndexBytes[2] = UInt8((blockIndex >> 8) & 0xFF)
            blockIndexBytes[3] = UInt8(blockIndex & 0xFF)

            let key = SymmetricKey(data: password)
            let firstInput = salt + blockIndexBytes
            var u = Data(HMAC<SHA512>.authenticationCode(for: firstInput, using: key))
            var block = u

            for _ in 1..<iterations {
                u = Data(HMAC<SHA512>.authenticationCode(for: u, using: key))
                for i in 0..<block.count {
                    block[i] ^= u[i]
                }
            }
            result.append(block)
        }

        return Data(result.prefix(keyLength))
    }

    private struct DerivedKey {
        let key: Data
        let chainCode: Data
    }

    private static func hmacSHA512(key: Data, data: Data) throws -> (Data, Data) {
        let hmacKey = SymmetricKey(data: key)
        let hmac = Data(HMAC<SHA512>.authenticationCode(for: data, using: hmacKey))
        return (Data(hmac.prefix(32)), Data(hmac.suffix(32)))
    }

    private static func deriveChild(key: Data, chainCode: Data, index: UInt32) throws -> DerivedKey {
        var inputData = Data()
        if index >= 0x80000000 {
            // Hardened: 0x00 || key || index
            inputData.append(0x00)
            inputData.append(key)
        } else {
            // Normal: pubkey || index
            let privKey = try PrivateKey(rawBytes: key)
            inputData.append(privKey.pubKey)
        }

        var indexBytes = Data(count: 4)
        indexBytes[0] = UInt8((index >> 24) & 0xFF)
        indexBytes[1] = UInt8((index >> 16) & 0xFF)
        indexBytes[2] = UInt8((index >> 8) & 0xFF)
        indexBytes[3] = UInt8(index & 0xFF)
        inputData.append(indexBytes)

        let hmacKey = SymmetricKey(data: chainCode)
        let hmac = Data(HMAC<SHA512>.authenticationCode(for: inputData, using: hmacKey))
        let childKey = Data(hmac.prefix(32))
        let childChainCode = Data(hmac.suffix(32))

        // Add parent key to child key (mod curve order) for proper BIP32
        let resultKey = try addPrivateKeys(childKey, key)

        return DerivedKey(key: resultKey, chainCode: childChainCode)
    }

    private static func addPrivateKeys(_ a: Data, _ b: Data) throws -> Data {
        // secp256k1 curve order
        let curveOrder: [UInt8] = [
            0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
            0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
            0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B,
            0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x41,
        ]

        var carry: UInt16 = 0
        var result = Data(count: 32)
        for i in stride(from: 31, through: 0, by: -1) {
            let sum = UInt16(a[i]) + UInt16(b[i]) + carry
            result[i] = UInt8(sum & 0xFF)
            carry = sum >> 8
        }

        // Reduce mod curve order if needed
        var isGreaterOrEqual = false
        for i in 0..<32 {
            if result[i] > curveOrder[i] { isGreaterOrEqual = true; break }
            if result[i] < curveOrder[i] { break }
        }

        if isGreaterOrEqual {
            var borrow: Int16 = 0
            for i in stride(from: 31, through: 0, by: -1) {
                let diff = Int16(result[i]) - Int16(curveOrder[i]) - borrow
                if diff < 0 {
                    result[i] = UInt8((diff + 256) & 0xFF)
                    borrow = 1
                } else {
                    result[i] = UInt8(diff & 0xFF)
                    borrow = 0
                }
            }
        }

        return result
    }
}

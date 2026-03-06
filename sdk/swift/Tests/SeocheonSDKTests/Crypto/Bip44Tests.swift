import XCTest
@testable import SeocheonSDK

final class Bip44Tests: XCTestCase {

    func testDeriveKeyFromMnemonic() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        XCTAssertEqual(key.bytes.count, 32)
    }

    func testDeriveKeyConsistency() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key1 = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let key2 = try Bip44.deriveKeyFromMnemonic(mnemonic)
        XCTAssertEqual(key1.bytes, key2.bytes)
    }

    func testDeriveKeyDifferentMnemonics() throws {
        let key1 = try Bip44.deriveKeyFromMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
        let key2 = try Bip44.deriveKeyFromMnemonic("zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong")
        XCTAssertNotEqual(key1.bytes, key2.bytes)
    }

    func testDeriveKeyProducesValidPublicKey() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let pubKey = key.pubKey
        XCTAssertEqual(pubKey.count, 33)
    }

    /// Cross-validates derived address against Go SDK (seocheon19rl4cm2hmr8afy4kldpxz3fka4jguq0astzff0)
    func testDeriveAddressMatchesGoSDK() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let addr = try Address.fromPubKey(key.pubKey)
        XCTAssertEqual(addr, "seocheon19rl4cm2hmr8afy4kldpxz3fka4jguq0astzff0",
                       "Swift SDK address must match Go SDK for same mnemonic")
    }

    /// Debug: trace BIP32 derivation step by step
    func testBip32DerivationSteps() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let seed = try Bip44._mnemonicToSeedForTesting(mnemonic)

        // Master key
        let (masterKey, masterChainCode) = try Bip44._deriveMasterKeyForTesting(seed: seed)
        print("master key:        \(masterKey.map { String(format:"%02x",$0) }.joined())")
        print("master chain code: \(masterChainCode.map { String(format:"%02x",$0) }.joined())")

        // m/44'
        let (k44, cc44) = try Bip44._deriveChildForTesting(key: masterKey, chainCode: masterChainCode, index: 44 + 0x80000000)
        print("m/44' key:         \(k44.map { String(format:"%02x",$0) }.joined())")

        // m/44'/118'
        let (k118, cc118) = try Bip44._deriveChildForTesting(key: k44, chainCode: cc44, index: 118 + 0x80000000)
        print("m/44'/118' key:    \(k118.map { String(format:"%02x",$0) }.joined())")

        // m/44'/118'/0'
        let (k0h, cc0h) = try Bip44._deriveChildForTesting(key: k118, chainCode: cc118, index: 0x80000000)
        print("m/44'/118'/0' key: \(k0h.map { String(format:"%02x",$0) }.joined())")

        // m/44'/118'/0'/0
        let (k0, cc0) = try Bip44._deriveChildForTesting(key: k0h, chainCode: cc0h, index: 0)
        print("m/44'/118'/0'/0 key: \(k0.map { String(format:"%02x",$0) }.joined())")

        // m/44'/118'/0'/0/0
        let (kFinal, _) = try Bip44._deriveChildForTesting(key: k0, chainCode: cc0, index: 0)
        print("m/44'/118'/0'/0/0 key: \(kFinal.map { String(format:"%02x",$0) }.joined())")
        XCTAssertEqual(kFinal.map { String(format:"%02x",$0) }.joined(),
                       "c4a48e2fce1481cd3294b4490f6678090ea98d3d0e5cd984558ab0968741b104")
    }

    func testBip32ChainCodes() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        let seed = try Bip44._mnemonicToSeedForTesting(mnemonic)
        let (masterKey, masterChainCode) = try Bip44._deriveMasterKeyForTesting(seed: seed)
        print("master cc: \(masterChainCode.map { String(format:"%02x",$0) }.joined())")
        let (k44, cc44) = try Bip44._deriveChildForTesting(key: masterKey, chainCode: masterChainCode, index: 44 + 0x80000000)
        print("m/44' cc:  \(cc44.map { String(format:"%02x",$0) }.joined())")
        let (_, cc118) = try Bip44._deriveChildForTesting(key: k44, chainCode: cc44, index: 118 + 0x80000000)
        print("m/44'/118' cc: \(cc118.map { String(format:"%02x",$0) }.joined())")
    }

    /// Debug: check BIP39 seed against known test vector
    func testBip39SeedVector() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        // Known BIP39 seed for this mnemonic (no passphrase)
        let expectedSeedHex = "5eb00bbddcf069084889a8ab9155568165f5c453ccb85e70811aaed6f6da5fc19a5ac40b389cd370d086206dec8aa6c43daea6690f20ad3d8d48b2d2ce9e38e4"
        let seed = try Bip44._mnemonicToSeedForTesting(mnemonic)
        let seedHex = seed.map { String(format: "%02x", $0) }.joined()
        print("BIP39 seed: \(seedHex)")
        XCTAssertEqual(seedHex, expectedSeedHex, "BIP39 seed must match known vector")
    }

    /// Debug: compare intermediate values against Go SDK reference
    func testDeriveKeyMatchesGoSDKVector() throws {
        let mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        // Go SDK reference values (m/44'/118'/0'/0/0):
        let expectedPrivKeyHex = "c4a48e2fce1481cd3294b4490f6678090ea98d3d0e5cd984558ab0968741b104"
        let expectedPubKeyHex  = "024f4e2ad99c34d60b9ba6283c9431a8418af8673212961f97a77b6377fcd05b62"

        let key = try Bip44.deriveKeyFromMnemonic(mnemonic)
        let privHex = key.bytes.map { String(format: "%02x", $0) }.joined()
        let pubHex  = key.pubKey.map { String(format: "%02x", $0) }.joined()

        print("privKey: \(privHex)")
        print("pubKey:  \(pubHex)")

        XCTAssertEqual(privHex, expectedPrivKeyHex, "Private key must match Go SDK")
        XCTAssertEqual(pubHex,  expectedPubKeyHex,  "Public key must match Go SDK")
    }
}

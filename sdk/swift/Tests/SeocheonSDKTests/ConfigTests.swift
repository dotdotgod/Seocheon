import XCTest
@testable import SeocheonSDK

final class ConfigTests: XCTestCase {

    func testValidDirectConfig() throws {
        let config = SDKConfig(
            chain: ChainConfig(chainId: "seocheon-1", rpcEndpoint: "http://localhost:26657", grpcEndpoint: "http://localhost:1317"),
            signing: SigningConfig(mode: .direct, mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
        )
        XCTAssertNoThrow(try config.validate())
    }

    func testValidVaultConfig() throws {
        let config = SDKConfig(
            chain: ChainConfig(chainId: "seocheon-1", rpcEndpoint: "http://localhost:26657", grpcEndpoint: "http://localhost:1317"),
            signing: SigningConfig(mode: .vault, vaultEndpoint: "http://vault:8200", keyName: "mykey")
        )
        XCTAssertNoThrow(try config.validate())
    }

    func testValidKeystoreConfig() throws {
        let config = SDKConfig(
            chain: ChainConfig(chainId: "seocheon-1", rpcEndpoint: "http://localhost:26657", grpcEndpoint: "http://localhost:1317"),
            signing: SigningConfig(mode: .keystore, keystorePath: "/path/to/keystore.json", passphraseEnv: "MY_PASS")
        )
        XCTAssertNoThrow(try config.validate())
    }

    func testMissingChainID() {
        let config = SDKConfig(
            chain: ChainConfig(chainId: "", rpcEndpoint: "http://localhost:26657", grpcEndpoint: "http://localhost:1317"),
            signing: SigningConfig(mode: .direct, mnemonic: "test mnemonic")
        )
        XCTAssertThrowsError(try config.validate()) { error in
            guard let sdkErr = error as? SDKError else { XCTFail("Expected SDKError"); return }
            XCTAssertEqual(sdkErr.code, 9005)
        }
    }

    func testMissingRPCEndpoint() {
        let config = SDKConfig(
            chain: ChainConfig(chainId: "seocheon-1", rpcEndpoint: "", grpcEndpoint: "http://localhost:1317"),
            signing: SigningConfig(mode: .direct, mnemonic: "test mnemonic")
        )
        XCTAssertThrowsError(try config.validate())
    }

    func testMissingMnemonic() {
        let config = SDKConfig(
            chain: ChainConfig(chainId: "seocheon-1", rpcEndpoint: "http://localhost:26657", grpcEndpoint: "http://localhost:1317"),
            signing: SigningConfig(mode: .direct)
        )
        XCTAssertThrowsError(try config.validate())
    }

    func testDefaultConfig() {
        let config = SDKConfig.defaultConfig(
            chainId: "seocheon-1",
            rpcEndpoint: "http://localhost:26657",
            grpcEndpoint: "http://localhost:1317",
            signing: SigningConfig(mode: .direct, mnemonic: "test mnemonic")
        )
        XCTAssertEqual(config.chain.chainId, "seocheon-1")
        XCTAssertEqual(config.tx.broadcastMode, "sync")
        XCTAssertEqual(config.tx.confirmTimeoutMs, 30000)
    }

    func testTxConfigDefaults() {
        let txConfig = TxConfig()
        XCTAssertEqual(txConfig.broadcastMode, "sync")
        XCTAssertEqual(txConfig.confirmTimeoutMs, 30000)
        XCTAssertEqual(txConfig.confirmPollIntervalMs, 1000)
    }
}

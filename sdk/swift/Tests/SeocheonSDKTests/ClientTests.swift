import XCTest
@testable import SeocheonSDK

final class ClientTests: XCTestCase {

    private func makeConfig() -> SDKConfig {
        return SDKConfig(
            chain: ChainConfig(chainId: "seocheon-1", rpcEndpoint: "http://localhost:26657", grpcEndpoint: "http://localhost:1317"),
            signing: SigningConfig(mode: .direct, mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
        )
    }

    func testSDKCreation() throws {
        let sdk = try SeocheonSDK(config: makeConfig())
        XCTAssertNotNil(sdk.activity)
        XCTAssertNotNil(sdk.epoch)
        XCTAssertNotNil(sdk.node)
        XCTAssertNotNil(sdk.rewards)
        XCTAssertNotNil(sdk.cosmos)
    }

    func testSDKIsNotConnectedByDefault() throws {
        let sdk = try SeocheonSDK(config: makeConfig())
        XCTAssertFalse(sdk.isConnected())
    }

    func testSDKGetConfig() throws {
        let config = makeConfig()
        let sdk = try SeocheonSDK(config: config)
        let returned = sdk.getConfig()
        XCTAssertEqual(returned.chain.chainId, "seocheon-1")
    }

    func testSDKInvalidConfigThrows() {
        let config = SDKConfig(
            chain: ChainConfig(chainId: "", rpcEndpoint: "", grpcEndpoint: ""),
            signing: SigningConfig(mode: .direct)
        )
        XCTAssertThrowsError(try SeocheonSDK(config: config))
    }

    func testSDKWithMockClient() async throws {
        let config = makeConfig()
        let mockClient = MockChainClient()
        let mockSigner = MockSigner()
        let sdk = SeocheonSDK(config: config, client: mockClient, signer: mockSigner)

        XCTAssertFalse(sdk.isConnected())
        try await sdk.connect()
        XCTAssertTrue(sdk.isConnected())
        await sdk.disconnect()
        XCTAssertFalse(sdk.isConnected())
    }
}

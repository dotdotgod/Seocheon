import XCTest
@testable import SeocheonSDK

final class PipelineTests: XCTestCase {

    func testPipelineConfigDefault() {
        let config = PipelineConfig.defaultConfig(chainID: "seocheon-1")
        XCTAssertEqual(config.chainID, "seocheon-1")
        XCTAssertEqual(config.gasPrice, Gas.defaultGasPrice)
    }

    func testPipelineConfigCustom() {
        let config = PipelineConfig(
            chainID: "test-chain",
            gasPrice: 500,
            confirmTimeoutSeconds: 5,
            pollIntervalSeconds: 0.5
        )
        XCTAssertEqual(config.chainID, "test-chain")
        XCTAssertEqual(config.gasPrice, 500)
    }

    func testTxRequestWithDefaultGas() {
        let msg = MsgSubmitActivity(submitter: "seocheon1abc", activityHash: String(repeating: "a", count: 64), contentURI: "ipfs://test")
        let request = TxRequest(message: msg)
        XCTAssertEqual(request.gasLimit, 0)
    }

    func testTxRequestWithCustomGas() {
        let msg = MsgSubmitActivity(submitter: "seocheon1abc", activityHash: String(repeating: "a", count: 64), contentURI: "ipfs://test")
        var request = TxRequest(message: msg)
        request.gasLimit = 500_000
        XCTAssertEqual(request.gasLimit, 500_000)
    }

    func testTxResultStruct() {
        let result = TxResult(txHash: "ABC123", height: 100, code: 0, gasUsed: 50000, gasWanted: 100000, rawLog: "ok", events: [])
        XCTAssertEqual(result.txHash, "ABC123")
        XCTAssertEqual(result.code, 0)
        XCTAssertEqual(result.height, 100)
    }

    func testChainClientAdapterConformance() {
        let mock = MockChainClient()
        let adapter = ChainClientAdapter(client: mock)
        XCTAssertNotNil(adapter)
    }

    func testSigningServiceAdapterConformance() {
        let mock = MockSigner()
        let adapter = SigningServiceAdapter(service: mock)
        XCTAssertNotNil(adapter)
    }
}

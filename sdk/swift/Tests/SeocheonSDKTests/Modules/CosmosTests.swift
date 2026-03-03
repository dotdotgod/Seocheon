import XCTest
@testable import SeocheonSDK

final class CosmosTests: XCTestCase {

    private func makeModule() -> (CosmosModule, MockChainClient) {
        let client = MockChainClient()
        let signer = MockSigner()
        let module = CosmosModule(client: client, signer: signer, chainID: "seocheon-1")
        return (module, client)
    }

    func testGetBalance() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "by_denom", json: """
        {"balance":{"denom":"uppyeo","amount":"50000000000"}}
        """)

        let result = try await module.getBalance()
        XCTAssertEqual(result.balance, "50000000000")
        XCTAssertEqual(result.balanceKkot, "5.0000000000")
        XCTAssertEqual(result.address, "seocheon1mockaddr")
    }

    func testGetBalanceCustomAddress() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "by_denom", json: """
        {"balance":{"denom":"uppyeo","amount":"100"}}
        """)

        let result = try await module.getBalance(address: "seocheon1custom")
        XCTAssertEqual(result.address, "seocheon1custom")
    }

    func testSendTokensEmptyAddress() async {
        let (module, _) = makeModule()
        do {
            _ = try await module.sendTokens(toAddress: "", amount: "1000")
            XCTFail("Expected invalidAddress error")
        } catch let err as SDKError {
            XCTAssertEqual(err, SDKError.invalidAddress)
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testGetBlockInfo() async throws {
        let (module, _) = makeModule()
        let info = try await module.getBlockInfo()
        XCTAssertEqual(info.blockHeight, 34560)
        XCTAssertEqual(info.chainId, "seocheon-1")
    }

    func testGetTxResult() async throws {
        let (module, _) = makeModule()
        let result = try await module.getTxResult(txHash: "ABC123")
        XCTAssertEqual(result.txHash, "ABC123")
        XCTAssertEqual(result.code, 0)
    }
}

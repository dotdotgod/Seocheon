import XCTest
@testable import SeocheonSDK

final class ChainClientTests: XCTestCase {

    func testMockClientConnect() async throws {
        let client = MockChainClient()
        XCTAssertFalse(client.isConnected())
        try await client.connect()
        XCTAssertTrue(client.isConnected())
    }

    func testMockClientDisconnect() async throws {
        let client = MockChainClient()
        try await client.connect()
        await client.disconnect()
        XCTAssertFalse(client.isConnected())
    }

    func testMockClientQueryREST() async throws {
        let client = MockChainClient()
        client.setResponse(path: "/test", json: "{\"result\":\"ok\"}")
        let data = try await client.queryREST(path: "/test")
        let str = String(data: data, encoding: .utf8)
        XCTAssertTrue(str?.contains("ok") ?? false)
    }

    func testMockClientGetLatestBlock() async throws {
        let client = MockChainClient()
        let block = try await client.getLatestBlock()
        XCTAssertEqual(block.height, 34560)
        XCTAssertEqual(block.chainId, "seocheon-1")
    }

    func testMockClientGetTx() async throws {
        let client = MockChainClient()
        let tx = try await client.getTx(txHash: "ABC123")
        XCTAssertEqual(tx.txHash, "ABC123")
        XCTAssertEqual(tx.code, 0)
    }
}

import XCTest
@testable import SeocheonSDK

final class EpochModuleTests: XCTestCase {

    private func makeModule() -> (EpochModule, MockChainClient) {
        let client = MockChainClient()
        let module = EpochModule(client: client)
        return (module, client)
    }

    func testGetInfo() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/activity/v1/epoch-info", json: """
        {"current_epoch":"2","current_window":"3","epoch_start_block":"34561","blocks_until_next_epoch":"5000"}
        """)
        client.latestBlock = BlockResponse(height: 40000, time: "2025-01-01T00:00:00Z", chainId: "seocheon-1", numTxs: 0)

        let info = try await module.getInfo()
        XCTAssertEqual(info.epochNumber, 2)
        XCTAssertEqual(info.windowNumber, 3)
        XCTAssertEqual(info.epochStartBlock, 34561)
    }

    func testGetQualificationRequiresNodeID() async {
        let (module, _) = makeModule()
        do {
            _ = try await module.getQualification(nodeID: "")
            XCTFail("Expected error for empty nodeID")
        } catch let err as SDKError {
            XCTAssertEqual(err.code, 9005)
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testGetQualification() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/activity/v1/epoch-info", json: """
        {"current_epoch":"1","current_window":"5","epoch_start_block":"17281","blocks_until_next_epoch":"8000"}
        """)
        client.setResponse(path: "epochs/1", json: """
        {"summary":{"active_windows":"5","eligible":false}}
        """)
        client.latestBlock = BlockResponse(height: 25000, time: "2025-01-01T00:00:00Z", chainId: "seocheon-1", numTxs: 0)

        let qual = try await module.getQualification(nodeID: "node1", epochNumber: 1)
        XCTAssertEqual(qual.activeWindows, 5)
        XCTAssertFalse(qual.isQualified)
        XCTAssertEqual(qual.requiredWindows, 8)
    }
}

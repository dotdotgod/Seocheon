import XCTest
@testable import SeocheonSDK

final class NodeTests: XCTestCase {

    private func makeModule() -> (NodeModule, MockChainClient) {
        let client = MockChainClient()
        let signer = MockSigner()
        let module = NodeModule(client: client, signer: signer, chainID: "seocheon-test")
        return (module, client)
    }

    func testGetInfoWithNodeID() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/nodes/node1", json: """
        {"node":{"id":"node1","operator":"seocheon1op","agent_address":"seocheon1ag","status":"NODE_STATUS_ACTIVE","description":"Test Node","website":"https://test.com","tags":["ai"],"agent_share":"0.2","validator_address":"","registered_at":"1000"}}
        """)
        let info = try await module.getInfo(nodeID: "node1")
        XCTAssertEqual(info.nodeId, "node1")
        XCTAssertEqual(info.operator, "seocheon1op")
        XCTAssertEqual(info.status, "ACTIVE")
        XCTAssertEqual(info.tags, ["ai"])
    }

    func testGetInfoNodeNotFound() async {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/nodes/nonexist", json: "{}")
        do {
            _ = try await module.getInfo(nodeID: "nonexist")
            XCTFail("Expected nodeNotFound")
        } catch let err as SDKError {
            XCTAssertEqual(err, SDKError.nodeNotFound)
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testSearch() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/nodes", json: """
        {"nodes":[{"id":"n1","status":"NODE_STATUS_ACTIVE","tags":["ai"],"description":"Node 1"},{"id":"n2","status":"NODE_STATUS_REGISTERED","tags":[],"description":"Node 2"}]}
        """)
        let result = try await module.search()
        XCTAssertEqual(result.totalCount, 2)
        XCTAssertEqual(result.nodes.count, 2)
    }

    func testSearchWithStatusFilter() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/nodes", json: """
        {"nodes":[{"id":"n1","status":"NODE_STATUS_ACTIVE","tags":[],"description":""},{"id":"n2","status":"NODE_STATUS_REGISTERED","tags":[],"description":""}]}
        """)
        let result = try await module.search(status: "ACTIVE")
        XCTAssertEqual(result.totalCount, 1)
    }

    func testGetDelegationStatus() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/delegation-confirmation/seocheon1del/seocheonvaloper1val", json: """
        {"expiry_epoch":"90","current_epoch":"5","in_renewal_window":false,"renewal_window_start":"83"}
        """)
        let result = try await module.getDelegationStatus(
            delegatorAddress: "seocheon1del",
            validatorAddress: "seocheonvaloper1val"
        )
        XCTAssertEqual(result.expiryEpoch, 90)
        XCTAssertEqual(result.currentEpoch, 5)
        XCTAssertFalse(result.inRenewalWindow)
        XCTAssertEqual(result.renewalWindowStart, 83)
    }

    func testConfirmDelegation() async throws {
        let (module, _) = makeModule()
        let result = try await module.confirmDelegation(validatorAddress: "seocheonvaloper1val")
        XCTAssertFalse(result.txHash.isEmpty)
        XCTAssertEqual(result.code, 0)
    }
}

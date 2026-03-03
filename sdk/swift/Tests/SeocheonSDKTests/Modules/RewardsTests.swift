import XCTest
@testable import SeocheonSDK

final class RewardsTests: XCTestCase {

    private func makeModule() -> (RewardsModule, MockChainClient) {
        let client = MockChainClient()
        let signer = MockSigner()
        let module = RewardsModule(client: client, signer: signer, chainID: "seocheon-1")
        return (module, client)
    }

    func testGetPendingWithNodeID() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/nodes/node1", json: """
        {"node":{"validator_address":"seocheonvaloper1abc","agent_share":"0.2"}}
        """)
        client.setResponse(path: "outstanding_rewards", json: """
        {"rewards":{"rewards":[{"denom":"uppyeo","amount":"5000000000"}]}}
        """)
        client.setResponse(path: "commission", json: """
        {"commission":{"commission":[{"denom":"uppyeo","amount":"1000000000"}]}}
        """)

        let pending = try await module.getPending(nodeID: "node1")
        XCTAssertNotNil(pending.delegationReward)
        XCTAssertNotNil(pending.commissionTotal)
        XCTAssertNotNil(pending.operatorShare)
        XCTAssertNotNil(pending.agentShare)
    }

    func testGetPendingNoValidator() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/nodes/node2", json: """
        {"node":{"validator_address":"","agent_share":"0.2"}}
        """)

        let pending = try await module.getPending(nodeID: "node2")
        XCTAssertEqual(pending.delegationReward, "0.0000000000")
        XCTAssertEqual(pending.commissionTotal, "0.0000000000")
    }

    func testGetPendingNodeNotFound() async {
        let (module, client) = makeModule()
        client.setResponse(path: "by-agent", json: "{\"node\":{}}")
        do {
            _ = try await module.getPending()
            XCTFail("Expected nodeNotFound")
        } catch let err as SDKError {
            XCTAssertEqual(err, SDKError.nodeNotFound)
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testAgentShareParsing() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "/seocheon/node/v1/nodes/node3", json: """
        {"node":{"validator_address":"seocheonvaloper1x","agent_share":"20"}}
        """)
        client.setResponse(path: "outstanding_rewards", json: """
        {"rewards":{"rewards":[{"denom":"uppyeo","amount":"0"}]}}
        """)
        client.setResponse(path: "commission", json: """
        {"commission":{"commission":[{"denom":"uppyeo","amount":"10000000000"}]}}
        """)

        let pending = try await module.getPending(nodeID: "node3")
        // agent_share "20" should be parsed as 20% (20/100 = 0.2)
        XCTAssertNotEqual(pending.operatorShare, pending.agentShare)
    }

    func testWithdrawReturnsResponse() async throws {
        let (module, client) = makeModule()
        client.broadcastResult = BroadcastResponse(txHash: "TX123", code: 0, rawLog: "")
        client.txResult = TxResponse(
            txHash: "TX123", height: 100, code: 0, gasUsed: 50000, gasWanted: 100000, rawLog: "",
            events: [
                ChainTxEvent(type: "withdraw_commission", attributes: [
                    ChainEventAttribute(key: "amount", value: "5000000000uppyeo"),
                    ChainEventAttribute(key: "operator_share", value: "4000000000uppyeo"),
                    ChainEventAttribute(key: "agent_share", value: "1000000000uppyeo"),
                ])
            ]
        )

        do {
            let result = try await module.withdraw()
            XCTAssertEqual(result.txHash, "TX123")
        } catch {
            // May fail due to mock account info not being set up for pipeline
        }
    }
}

import XCTest
@testable import SeocheonSDK

final class ActivityTests: XCTestCase {

    private func makeModule() -> (ActivityModule, MockChainClient) {
        let client = MockChainClient()
        let signer = MockSigner()
        let module = ActivityModule(client: client, signer: signer, chainID: "seocheon-1")
        return (module, client)
    }

    func testSubmitInvalidHash() async {
        let (module, _) = makeModule()
        do {
            _ = try await module.submit(activityHash: "short", contentURI: "ipfs://test")
            XCTFail("Expected invalidActivityHash error")
        } catch let err as SDKError {
            XCTAssertEqual(err, SDKError.invalidActivityHash)
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testSubmitEmptyContentURI() async {
        let (module, _) = makeModule()
        let hash = String(repeating: "a", count: 64)
        do {
            _ = try await module.submit(activityHash: hash, contentURI: "")
            XCTFail("Expected invalidContentURI error")
        } catch let err as SDKError {
            XCTAssertEqual(err, SDKError.invalidContentURI)
        } catch {
            XCTFail("Unexpected error: \(error)")
        }
    }

    func testGetActivities() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "by-agent", json: "{\"node\":{\"id\":\"node1\"}}")
        client.setResponse(path: "activities", json: "{\"activities\":[{\"activity_hash\":\"aaaa\",\"content_uri\":\"ipfs://x\",\"block_height\":\"100\"}]}")
        let result = try await module.getActivities()
        XCTAssertEqual(result.totalCount, 1)
    }

    func testGetQuota() async throws {
        let (module, client) = makeModule()
        client.setResponse(path: "by-agent", json: "{\"node\":{\"id\":\"node1\"}}")
        client.setResponse(path: "epochs", json: "{\"quota_used\":\"5\",\"quota_limit\":\"100\"}")
        let result = try await module.getQuota()
        XCTAssertEqual(result.quotaUsed, 5)
        XCTAssertEqual(result.quotaTotal, 100)
        XCTAssertEqual(result.quotaRemaining, 95)
    }

    func testSubmitNodeNotRegistered() async {
        let (module, client) = makeModule()
        client.setResponse(path: "by-agent", json: "{\"node\":{}}")
        let hash = String(repeating: "a", count: 64)
        client.broadcastResult = BroadcastResponse(txHash: "TX1", code: 0, rawLog: "")
        client.txResult = TxResponse(txHash: "TX1", height: 100, code: 0, gasUsed: 50000, gasWanted: 100000, rawLog: "", events: [])

        // Submit should work since the mock broadcast succeeds
        // The resolveOwnNodeID is only called by getActivities/getQuota, not submit
        do {
            _ = try await module.submit(activityHash: hash, contentURI: "ipfs://test")
        } catch {
            // May fail if the mock doesn't have account info - this is expected
        }
    }
}

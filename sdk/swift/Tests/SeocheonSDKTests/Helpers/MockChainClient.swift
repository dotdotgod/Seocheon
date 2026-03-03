import Foundation
@testable import SeocheonSDK

final class MockChainClient: ChainClient, @unchecked Sendable {
    var connected = false
    var responses: [String: Data] = [:]
    var broadcastResult = BroadcastResponse(txHash: "AABB00", code: 0, rawLog: "")
    var latestBlock = BlockResponse(height: 34560, time: "2025-01-01T00:00:00Z", chainId: "seocheon-1", numTxs: 5)
    var txResult: TxResponse?
    var accountResult = AccountInfo(accountNumber: 1, sequence: 0)
    var queryError: Error?
    var broadcastError: Error?

    func connect() async throws {
        connected = true
    }

    func disconnect() async {
        connected = false
    }

    func isConnected() -> Bool {
        return connected
    }

    func queryREST(path: String) async throws -> Data {
        if let err = queryError { throw err }
        if let data = responses[path] { return data }
        for (key, data) in responses {
            if path.contains(key) { return data }
        }
        return Data("{}".utf8)
    }

    func broadcastTx(txBytes: Data, mode: String) async throws -> BroadcastResponse {
        if let err = broadcastError { throw err }
        return broadcastResult
    }

    func getLatestBlock() async throws -> BlockResponse {
        return latestBlock
    }

    func getTx(txHash: String) async throws -> TxResponse {
        if let tx = txResult { return tx }
        return TxResponse(txHash: txHash, height: 100, code: 0, gasUsed: 50000, gasWanted: 100000, rawLog: "", events: [])
    }

    func getAccountInfo(address: String) async throws -> AccountInfo {
        return accountResult
    }

    func setResponse(path: String, json: String) {
        responses[path] = Data(json.utf8)
    }
}

import Foundation

/// Chain communication response types.
public struct BroadcastResponse: Sendable {
    public let txHash: String
    public let code: UInt32
    public let rawLog: String
}

public struct BlockResponse: Sendable {
    public let height: Int64
    public let time: String
    public let chainId: String
    public let numTxs: Int
}

public struct TxResponse: Sendable {
    public let txHash: String
    public let height: Int64
    public let code: UInt32
    public let gasUsed: UInt64
    public let gasWanted: UInt64
    public let rawLog: String
    public let events: [ChainTxEvent]
}

public struct ChainTxEvent: Sendable {
    public let type: String
    public let attributes: [ChainEventAttribute]
}

public struct ChainEventAttribute: Sendable {
    public let key: String
    public let value: String
}

public struct AccountInfo: Sendable {
    public let accountNumber: UInt64
    public let sequence: UInt64
}

/// Protocol defining chain communication.
public protocol ChainClient: Sendable {
    func connect() async throws
    func disconnect() async
    func isConnected() -> Bool

    func queryREST(path: String) async throws -> Data
    func broadcastTx(txBytes: Data, mode: String) async throws -> BroadcastResponse
    func getLatestBlock() async throws -> BlockResponse
    func getTx(txHash: String) async throws -> TxResponse
    func getAccountInfo(address: String) async throws -> AccountInfo
}

/// HTTP-based implementation of ChainClient.
public final class URLSessionChainClient: ChainClient, @unchecked Sendable {
    private let rpcEndpoint: String
    private let grpcEndpoint: String
    private let session: URLSession
    private var _connected: Bool = false

    public init(rpcEndpoint: String, grpcEndpoint: String) {
        self.rpcEndpoint = rpcEndpoint.hasSuffix("/") ? String(rpcEndpoint.dropLast()) : rpcEndpoint
        self.grpcEndpoint = grpcEndpoint.hasSuffix("/") ? String(grpcEndpoint.dropLast()) : grpcEndpoint
        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        self.session = URLSession(configuration: config)
    }

    public func connect() async throws {
        _ = try await getLatestBlock()
        _connected = true
    }

    public func disconnect() async {
        _connected = false
    }

    public func isConnected() -> Bool {
        return _connected
    }

    public func queryREST(path: String) async throws -> Data {
        let url = URL(string: grpcEndpoint + path)!
        let (data, response) = try await session.data(from: url)
        guard let httpResponse = response as? HTTPURLResponse, httpResponse.statusCode == 200 else {
            throw SDKError.queryFailed("query failed with status: \(String(data: data, encoding: .utf8) ?? "unknown")")
        }
        return data
    }

    public func broadcastTx(txBytes: Data, mode: String) async throws -> BroadcastResponse {
        let base64Tx = txBytes.base64EncodedString()
        let protoMode = mode == "async" ? "BROADCAST_MODE_ASYNC" : "BROADCAST_MODE_SYNC"
        let payload = "{\"tx_bytes\":\"\(base64Tx)\",\"mode\":\"\(protoMode)\"}"

        let url = URL(string: grpcEndpoint + "/cosmos/tx/v1beta1/txs")!
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = Data(payload.utf8)

        let (data, _) = try await session.data(for: request)

        struct BroadcastResult: Codable {
            struct TxResp: Codable {
                let txhash: String?
                let code: UInt32?
                let raw_log: String?
            }
            let tx_response: TxResp?
        }

        let result = try JSONDecoder().decode(BroadcastResult.self, from: data)
        return BroadcastResponse(
            txHash: result.tx_response?.txhash ?? "",
            code: result.tx_response?.code ?? 0,
            rawLog: result.tx_response?.raw_log ?? ""
        )
    }

    public func getLatestBlock() async throws -> BlockResponse {
        let data = try await queryREST(path: "/cosmos/base/tendermint/v1beta1/blocks/latest")

        struct BlockResult: Codable {
            struct Block: Codable {
                struct Header: Codable {
                    let height: String?
                    let time: String?
                    let chain_id: String?
                }
                struct Data_: Codable {
                    let txs: [String]?
                }
                let header: Header?
                let data: Data_?
            }
            let block: Block?
        }

        let result = try JSONDecoder().decode(BlockResult.self, from: data)
        return BlockResponse(
            height: Int64(result.block?.header?.height ?? "0") ?? 0,
            time: result.block?.header?.time ?? "",
            chainId: result.block?.header?.chain_id ?? "",
            numTxs: result.block?.data?.txs?.count ?? 0
        )
    }

    public func getTx(txHash: String) async throws -> TxResponse {
        let data = try await queryREST(path: "/cosmos/tx/v1beta1/txs/\(txHash)")

        struct TxQueryResult: Codable {
            struct TxResp: Codable {
                let txhash: String?
                let height: String?
                let code: UInt32?
                let gas_used: String?
                let gas_wanted: String?
                let raw_log: String?
                let events: [EventProto]?
            }
            struct EventProto: Codable {
                let type: String?
                let attributes: [AttrProto]?
            }
            struct AttrProto: Codable {
                let key: String?
                let value: String?
            }
            let tx_response: TxResp?
        }

        let result = try JSONDecoder().decode(TxQueryResult.self, from: data)
        let resp = result.tx_response
        let events = (resp?.events ?? []).map { e in
            ChainTxEvent(
                type: e.type ?? "",
                attributes: (e.attributes ?? []).map { ChainEventAttribute(key: $0.key ?? "", value: $0.value ?? "") }
            )
        }
        return TxResponse(
            txHash: resp?.txhash ?? "",
            height: Int64(resp?.height ?? "0") ?? 0,
            code: resp?.code ?? 0,
            gasUsed: UInt64(resp?.gas_used ?? "0") ?? 0,
            gasWanted: UInt64(resp?.gas_wanted ?? "0") ?? 0,
            rawLog: resp?.raw_log ?? "",
            events: events
        )
    }

    public func getAccountInfo(address: String) async throws -> AccountInfo {
        let data = try await queryREST(path: "/cosmos/auth/v1beta1/accounts/\(address)")

        struct AccountResult: Codable {
            struct Account: Codable {
                let account_number: String?
                let sequence: String?
            }
            let account: Account?
        }

        let result = try JSONDecoder().decode(AccountResult.self, from: data)
        return AccountInfo(
            accountNumber: UInt64(result.account?.account_number ?? "0") ?? 0,
            sequence: UInt64(result.account?.sequence ?? "0") ?? 0
        )
    }
}

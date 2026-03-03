import Foundation

/// Provides standard Cosmos operations.
public final class CosmosModule: @unchecked Sendable {
    private let client: ChainClient
    private let signer: SigningService
    private let txQuerier: ChainQuerier
    private let txConfig: PipelineConfig
    private let txSigner: TxSigner

    internal init(client: ChainClient, signer: SigningService, chainID: String) {
        self.client = client
        self.signer = signer
        self.txQuerier = ChainClientAdapter(client: client)
        self.txConfig = PipelineConfig.defaultConfig(chainID: chainID)
        self.txSigner = SigningServiceAdapter(service: signer)
    }

    /// Returns the token balance for an address.
    public func getBalance(address: String = "", denom: String = "") async throws -> BalanceResponse {
        let effectiveAddr = address.isEmpty ? signer.getAddress() : address
        let effectiveDenom = denom.isEmpty ? "uppyeo" : denom

        let path = "/cosmos/bank/v1beta1/balances/\(effectiveAddr)/by_denom?denom=\(effectiveDenom)"
        let data = try await client.queryREST(path: path)

        struct BalanceResult: Codable {
            struct Balance: Codable { let denom: String?; let amount: String? }
            let balance: Balance?
        }

        let result = try JSONDecoder().decode(BalanceResult.self, from: data)
        let amountStr = result.balance?.amount ?? "0"
        let balanceUppyeo = parseIntSafe(amountStr)

        return BalanceResponse(
            address: effectiveAddr,
            balance: amountStr,
            balanceKkot: ConvertUtils.formatKkot(balanceUppyeo)
        )
    }

    /// Sends tokens to the specified address.
    public func sendTokens(toAddress: String, amount: String, denom: String = "") async throws -> SendTokensResponse {
        guard !toAddress.isEmpty else { throw SDKError.invalidAddress }
        guard !amount.isEmpty else { throw SDKError.invalidConfig("amount is required") }

        let effectiveDenom = denom.isEmpty ? "uppyeo" : denom
        let msg = MsgSend(
            fromAddress: signer.getAddress(),
            toAddress: toAddress,
            amount: [Coin(denom: effectiveDenom, amount: amount)]
        )

        let result = try await TxPipeline.executeTx(querier: txQuerier, signer: txSigner, config: txConfig, request: TxRequest(message: msg))

        if result.code != 0 {
            throw SDKError.fromABCICode(result.code)
        }

        return SendTokensResponse(txHash: result.txHash, blockHeight: result.height)
    }

    /// Returns the latest block information.
    public func getBlockInfo() async throws -> BlockInfoResponse {
        let block = try await client.getLatestBlock()
        return BlockInfoResponse(
            blockHeight: block.height,
            blockTime: block.time,
            chainId: block.chainId,
            numTxs: UInt64(block.numTxs)
        )
    }

    /// Returns the result of a transaction by its hash.
    public func getTxResult(txHash: String) async throws -> TxResultResponse {
        guard !txHash.isEmpty else { throw SDKError.txNotFound }

        let tx = try await client.getTx(txHash: txHash)
        let events = tx.events.map { e in
            TxEvent(
                type: e.type,
                attributes: e.attributes.map { EventAttribute(key: $0.key, value: $0.value) }
            )
        }

        return TxResultResponse(
            txHash: tx.txHash,
            height: tx.height,
            code: tx.code,
            gasUsed: tx.gasUsed,
            gasWanted: tx.gasWanted,
            rawLog: tx.rawLog,
            events: events
        )
    }

    private func parseIntSafe(_ s: String) -> Int64 {
        var result: Int64 = 0
        for c in s {
            guard c.isNumber else { break }
            result = result * 10 + Int64(c.asciiValue! - Character("0").asciiValue!)
        }
        return result
    }
}

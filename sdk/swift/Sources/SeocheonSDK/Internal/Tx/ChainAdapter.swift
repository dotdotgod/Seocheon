import Foundation

/// Adapts ChainClient to the ChainQuerier interface.
internal final class ChainClientAdapter: ChainQuerier {
    private let client: ChainClient

    init(client: ChainClient) {
        self.client = client
    }

    func getAccountInfo(address: String) async throws -> (accountNumber: UInt64, sequence: UInt64) {
        let info = try await client.getAccountInfo(address: address)
        return (info.accountNumber, info.sequence)
    }

    func broadcastTxSync(txBytes: Data) async throws -> (txHash: String, code: UInt32, rawLog: String) {
        let resp = try await client.broadcastTx(txBytes: txBytes, mode: "sync")
        return (resp.txHash, resp.code, resp.rawLog)
    }

    func getTxResult(txHash: String) async throws -> TxResult {
        let resp = try await client.getTx(txHash: txHash)
        let events = resp.events.map { e in
            TxEventInternal(
                type: e.type,
                attributes: e.attributes.map { TxEventAttribute(key: $0.key, value: $0.value) }
            )
        }
        return TxResult(
            txHash: resp.txHash,
            height: resp.height,
            code: resp.code,
            gasUsed: resp.gasUsed,
            gasWanted: resp.gasWanted,
            rawLog: resp.rawLog,
            events: events
        )
    }
}

/// Wraps a SigningService to conform to TxSigner.
internal final class SigningServiceAdapter: TxSigner {
    private let service: SigningService

    init(service: SigningService) {
        self.service = service
    }

    func sign(_ data: Data) async throws -> Data {
        return try await service.sign(data)
    }

    func getAddress() -> String {
        return service.getAddress()
    }

    func getPubKey() -> Data {
        return service.getPubKey()
    }
}

import Foundation

/// Holds parameters for building and broadcasting a transaction.
internal struct TxRequest {
    let message: MessageEncoder
    var memo: String = ""
    var timeoutHeight: UInt64 = 0
    var gasLimit: UInt64 = 0
    var feeAmount: UInt64 = 0
    var feeDenom: String = ""
}

/// Holds the result of a broadcast and confirmed transaction.
internal struct TxResult {
    var txHash: String = ""
    var height: Int64 = 0
    var code: UInt32 = 0
    var gasUsed: UInt64 = 0
    var gasWanted: UInt64 = 0
    var rawLog: String = ""
    var events: [TxEventInternal] = []
}

/// Internal transaction event representation.
internal struct TxEventInternal {
    let type: String
    let attributes: [TxEventAttribute]
}

/// Internal event attribute.
internal struct TxEventAttribute {
    let key: String
    let value: String
}

/// Abstracts the signing capability needed by the TX pipeline.
internal protocol TxSigner {
    func sign(_ data: Data) async throws -> Data
    func getAddress() -> String
    func getPubKey() -> Data
}

/// Abstracts the chain queries needed by the TX pipeline.
internal protocol ChainQuerier {
    func getAccountInfo(address: String) async throws -> (accountNumber: UInt64, sequence: UInt64)
    func broadcastTxSync(txBytes: Data) async throws -> (txHash: String, code: UInt32, rawLog: String)
    func getTxResult(txHash: String) async throws -> TxResult
}

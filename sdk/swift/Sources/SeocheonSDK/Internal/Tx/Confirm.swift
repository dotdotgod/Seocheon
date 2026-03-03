import Foundation

/// TX confirmation polling.
internal enum TxConfirm {
    /// Polls for a transaction result until confirmed or timeout.
    static func pollTxConfirmation(
        querier: ChainQuerier,
        txHash: String,
        timeout: TimeInterval = 30,
        pollInterval: TimeInterval = 1
    ) async throws -> TxResult {
        let deadline = Date().addingTimeInterval(timeout)

        while Date() < deadline {
            do {
                let result = try await querier.getTxResult(txHash: txHash)
                return result
            } catch {
                // TX not yet indexed, continue polling
            }
            try await Task.sleep(nanoseconds: UInt64(pollInterval * 1_000_000_000))
        }

        throw SDKError.txTimeout
    }
}

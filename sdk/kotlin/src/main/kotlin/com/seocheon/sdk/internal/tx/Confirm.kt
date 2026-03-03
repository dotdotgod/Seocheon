package com.seocheon.sdk.internal.tx

import com.seocheon.sdk.errors.SdkError
import kotlinx.coroutines.delay

/**
 * Polls for a transaction result until confirmed or timeout.
 */
object Confirm {

    suspend fun pollTxConfirmation(
        querier: ChainQuerier,
        txHash: String,
        timeoutMs: Long = 30000,
        pollIntervalMs: Long = 1000,
    ): TxResult {
        val deadline = System.currentTimeMillis() + timeoutMs

        while (System.currentTimeMillis() < deadline) {
            try {
                return querier.getTxResult(txHash)
            } catch (_: Exception) {
                // TX not yet indexed, continue polling
            }
            delay(pollIntervalMs)
        }
        throw SdkError.TxTimeout()
    }
}

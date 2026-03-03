package com.seocheon.sdk.internal.tx

import com.seocheon.sdk.constants.ChainConstants
import com.seocheon.sdk.errors.SdkError

/**
 * 4-phase TX pipeline configuration.
 */
data class PipelineConfig(
    val chainId: String,
    val gasPrice: Long = ChainConstants.GAS_PRICE,
    val confirmTimeoutMs: Long = ChainConstants.DEFAULT_CONFIRM_TIMEOUT_MS,
    val pollIntervalMs: Long = ChainConstants.DEFAULT_CONFIRM_POLL_MS,
)

/**
 * Executes the full 4-phase TX pipeline:
 *   Phase 1 - Assembly: query account, build TxBody + AuthInfo + SignDoc
 *   Phase 2 - Signing: sign the SignDoc
 *   Phase 3 - Broadcast: encode TxRaw and broadcast
 *   Phase 4 - Confirmation: poll for TX inclusion
 */
object Pipeline {

    suspend fun executeTx(
        querier: ChainQuerier,
        signer: Signer,
        config: PipelineConfig,
        request: TxRequest,
    ): TxResult {
        // Phase 1: Assembly
        val address = signer.getAddress()
        val pubKey = signer.getPubKey()

        val (accountNumber, sequence) = querier.getAccountInfo(address)

        // Determine gas limit
        val gasLimit = if (request.gasLimit > 0) request.gasLimit
            else Gas.defaultGasForMessage(request.message.typeUrl())

        // Determine fee
        val gasPrice = if (config.gasPrice > 0) config.gasPrice else ChainConstants.GAS_PRICE
        val feeAmount = if (request.feeAmount > 0) request.feeAmount
            else Gas.calculateFee(gasLimit, gasPrice)

        val feeDenom = request.feeDenom.ifEmpty { ChainConstants.FEE_DENOM }

        // Encode TxBody
        val bodyBytes = Envelope.encodeTxBody(listOf(request.message), request.memo, request.timeoutHeight)

        // Encode AuthInfo
        val feeCoins = listOf(Coin(feeDenom, feeAmount.toString()))
        val authInfoBytes = Envelope.encodeAuthInfo(pubKey, sequence, feeCoins, gasLimit)

        // Encode SignDoc
        val signDocBytes = Envelope.encodeSignDoc(bodyBytes, authInfoBytes, config.chainId, accountNumber)

        // Phase 2: Signing
        val signature = try {
            signer.sign(signDocBytes)
        } catch (e: Exception) {
            throw SdkError.SigningFailed(e)
        }

        // Phase 3: Broadcast
        val txRawBytes = Envelope.encodeTxRaw(bodyBytes, authInfoBytes, signature)

        val (txHash, code, rawLog) = try {
            querier.broadcastTxSync(txRawBytes)
        } catch (e: Exception) {
            throw SdkError.BroadcastFailed(e)
        }

        // Check broadcast result
        if (code != 0u) {
            throw SdkError.fromAbciCode(code)
        }

        // Phase 4: Confirmation
        return Confirm.pollTxConfirmation(querier, txHash, config.confirmTimeoutMs, config.pollIntervalMs)
    }
}

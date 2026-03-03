package com.seocheon.sdk.modules

import com.seocheon.sdk.constants.ChainConstants
import com.seocheon.sdk.infrastructure.ChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.internal.tx.*
import com.seocheon.sdk.types.*
import com.seocheon.sdk.utils.ConvertUtils
import kotlinx.serialization.json.*

/**
 * Cosmos module for standard chain operations.
 */
class CosmosModule(
    private val client: ChainClient,
    private val signer: SigningService,
    private val pipelineConfig: PipelineConfig,
) {

    /**
     * Queries the token balance for an address.
     */
    suspend fun getBalance(address: String? = null, denom: String = ChainConstants.TOKEN_BASE_DENOM): BalanceResponse {
        val addr = address ?: signer.getAddress()
        val resp = client.queryRest("/cosmos/bank/v1beta1/balances/$addr/by_denom?denom=$denom")

        val balance = resp.jsonObject["balance"]?.jsonObject
        val amount = balance?.get("amount")?.jsonPrimitive?.content ?: "0"
        val balanceKkot = ConvertUtils.formatKkot(amount.toLongOrNull() ?: 0L)

        return BalanceResponse(
            address = addr,
            balance = amount,
            balanceKkot = balanceKkot,
        )
    }

    /**
     * Sends tokens to another address.
     */
    suspend fun sendTokens(toAddress: String, amount: String, denom: String = ChainConstants.TOKEN_BASE_DENOM): SendTokensResponse {
        val msg = MsgSend(
            fromAddress = signer.getAddress(),
            toAddress = toAddress,
            amount = listOf(Coin(denom, amount)),
        )

        val result = Pipeline.executeTx(
            client, signerAdapter(), pipelineConfig,
            TxRequest(message = msg),
        )

        return SendTokensResponse(
            txHash = result.txHash,
            blockHeight = result.height,
        )
    }

    /**
     * Queries the latest block information.
     */
    suspend fun getBlockInfo(): BlockInfoResponse {
        val resp = client.getLatestBlock()
        val block = resp.jsonObject["block"]?.jsonObject
        val header = block?.get("header")?.jsonObject

        return BlockInfoResponse(
            blockHeight = header?.get("height")?.jsonPrimitive?.longOrNull ?: 0L,
            blockTime = header?.get("time")?.jsonPrimitive?.content ?: "",
            chainId = header?.get("chain_id")?.jsonPrimitive?.content ?: "",
            numTxs = block?.get("data")?.jsonObject?.get("txs")?.jsonArray?.size?.toLong() ?: 0L,
        )
    }

    /**
     * Queries a transaction result by hash.
     */
    suspend fun getTxResult(txHash: String): TxResultResponse {
        val result = client.getTxResult(txHash)

        return TxResultResponse(
            txHash = result.txHash,
            height = result.height,
            code = result.code,
            gasUsed = result.gasUsed,
            gasWanted = result.gasWanted,
            rawLog = result.rawLog,
            events = result.events.map { event ->
                TxEvent(
                    type = event.type,
                    attributes = event.attributes.map { attr ->
                        EventAttribute(key = attr.key, value = attr.value)
                    },
                )
            },
        )
    }

    private fun signerAdapter(): Signer = object : Signer {
        override suspend fun sign(data: ByteArray): ByteArray = signer.sign(data)
        override fun getAddress(): String = signer.getAddress()
        override fun getPubKey(): ByteArray = signer.getPubKey()
    }
}

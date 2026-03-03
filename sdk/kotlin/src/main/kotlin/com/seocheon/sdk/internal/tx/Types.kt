package com.seocheon.sdk.internal.tx

/**
 * TX pipeline types.
 */
data class TxRequest(
    val message: MessageEncoder,
    val memo: String = "",
    val timeoutHeight: Long = 0,
    val gasLimit: Long = 0,
    val feeAmount: Long = 0,
    val feeDenom: String = "",
)

data class TxResult(
    val txHash: String,
    val height: Long = 0,
    val code: UInt = 0u,
    val gasUsed: Long = 0,
    val gasWanted: Long = 0,
    val rawLog: String = "",
    val events: List<TxEventData> = emptyList(),
)

data class TxEventData(
    val type: String,
    val attributes: List<TxEventAttribute>,
)

data class TxEventAttribute(
    val key: String,
    val value: String,
)

/**
 * Abstracts the signing capability needed by the TX pipeline.
 */
interface Signer {
    suspend fun sign(data: ByteArray): ByteArray
    fun getAddress(): String
    fun getPubKey(): ByteArray
}

/**
 * Abstracts the chain queries needed by the TX pipeline.
 */
interface ChainQuerier {
    suspend fun getAccountInfo(address: String): Pair<Long, Long> // accountNumber, sequence
    suspend fun broadcastTxSync(txBytes: ByteArray): Triple<String, UInt, String> // txHash, code, rawLog
    suspend fun getTxResult(txHash: String): TxResult
}

package com.seocheon.sdk.infrastructure

import com.seocheon.sdk.ChainConfig
import com.seocheon.sdk.errors.SdkError
import com.seocheon.sdk.internal.tx.ChainQuerier
import com.seocheon.sdk.internal.tx.TxEventAttribute
import com.seocheon.sdk.internal.tx.TxEventData
import com.seocheon.sdk.internal.tx.TxResult
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.*
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.Base64
import java.util.concurrent.TimeUnit

/**
 * Chain client interface for blockchain interaction.
 */
interface ChainClient : ChainQuerier {
    suspend fun connect()
    suspend fun disconnect()
    fun isConnected(): Boolean
    suspend fun queryRest(path: String): JsonElement
    suspend fun getLatestBlock(): JsonElement
}

/**
 * HTTP-based chain client implementation using OkHttp.
 */
class OkHttpChainClient(private val config: ChainConfig) : ChainClient {

    private var connected = false
    private val client = OkHttpClient.Builder()
        .connectTimeout(10, TimeUnit.SECONDS)
        .readTimeout(30, TimeUnit.SECONDS)
        .build()
    private val json = Json { ignoreUnknownKeys = true }

    override suspend fun connect() {
        // Validate endpoints by making a status request
        try {
            getLatestBlock()
            connected = true
        } catch (e: Exception) {
            throw SdkError.QueryFailed("failed to connect: ${e.message}", e)
        }
    }

    override suspend fun disconnect() {
        connected = false
    }

    override fun isConnected(): Boolean = connected

    override suspend fun queryRest(path: String): JsonElement = withContext(Dispatchers.IO) {
        val url = "${config.grpcEndpoint.trimEnd('/')}$path"
        val request = Request.Builder().url(url).get().build()
        val response = client.newCall(request).execute()
        val body = response.body?.string() ?: throw SdkError.QueryFailed("empty response from $path")
        if (!response.isSuccessful) {
            throw SdkError.QueryFailed("query failed ($path): ${response.code} $body")
        }
        json.parseToJsonElement(body)
    }

    override suspend fun getLatestBlock(): JsonElement =
        queryRest("/cosmos/base/tendermint/v1beta1/blocks/latest")

    override suspend fun getAccountInfo(address: String): Pair<Long, Long> {
        val resp = queryRest("/cosmos/auth/v1beta1/accounts/$address")
        val account = resp.jsonObject["account"]?.jsonObject ?: resp.jsonObject
        val accountNumber = account["account_number"]?.jsonPrimitive?.longOrNull ?: 0L
        val sequence = account["sequence"]?.jsonPrimitive?.longOrNull ?: 0L
        return Pair(accountNumber, sequence)
    }

    override suspend fun broadcastTxSync(txBytes: ByteArray): Triple<String, UInt, String> =
        withContext(Dispatchers.IO) {
            val url = "${config.rpcEndpoint.trimEnd('/')}/broadcast_tx_sync"
            val txBase64 = Base64.getEncoder().encodeToString(txBytes)
            val reqBody = buildJsonObject {
                put("jsonrpc", "2.0")
                put("id", 1)
                put("method", "broadcast_tx_sync")
                put("params", buildJsonObject {
                    put("tx", txBase64)
                })
            }.toString()

            val request = Request.Builder()
                .url(url)
                .post(reqBody.toRequestBody("application/json".toMediaType()))
                .build()

            val response = client.newCall(request).execute()
            val body = response.body?.string() ?: throw SdkError.BroadcastFailed()

            val parsed = json.parseToJsonElement(body).jsonObject
            val result = parsed["result"]?.jsonObject ?: throw SdkError.BroadcastFailed()

            val txHash = result["hash"]?.jsonPrimitive?.content ?: ""
            val code = result["code"]?.jsonPrimitive?.longOrNull?.toUInt() ?: 0u
            val rawLog = result["log"]?.jsonPrimitive?.content ?: ""

            Triple(txHash, code, rawLog)
        }

    override suspend fun getTxResult(txHash: String): TxResult {
        val resp = queryRest("/cosmos/tx/v1beta1/txs/$txHash")
        val txResponse = resp.jsonObject["tx_response"]?.jsonObject
            ?: throw SdkError.TxNotFound()

        val events = txResponse["events"]?.jsonArray?.map { eventEl ->
            val eventObj = eventEl.jsonObject
            TxEventData(
                type = eventObj["type"]?.jsonPrimitive?.content ?: "",
                attributes = eventObj["attributes"]?.jsonArray?.map { attrEl ->
                    val attrObj = attrEl.jsonObject
                    TxEventAttribute(
                        key = attrObj["key"]?.jsonPrimitive?.content ?: "",
                        value = attrObj["value"]?.jsonPrimitive?.content ?: "",
                    )
                } ?: emptyList()
            )
        } ?: emptyList()

        return TxResult(
            txHash = txResponse["txhash"]?.jsonPrimitive?.content ?: txHash,
            height = txResponse["height"]?.jsonPrimitive?.longOrNull ?: 0L,
            code = txResponse["code"]?.jsonPrimitive?.longOrNull?.toUInt() ?: 0u,
            gasUsed = txResponse["gas_used"]?.jsonPrimitive?.longOrNull ?: 0L,
            gasWanted = txResponse["gas_wanted"]?.jsonPrimitive?.longOrNull ?: 0L,
            rawLog = txResponse["raw_log"]?.jsonPrimitive?.content ?: "",
            events = events,
        )
    }
}

package com.seocheon.sdk.infrastructure.signing

import com.seocheon.sdk.utils.HashUtils
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit

/**
 * Signs transactions via an external vault server (production).
 */
class VaultService private constructor(
    private val endpoint: String,
    private val keyName: String,
    private val address: String,
    private val pubKey: ByteArray,
    private val client: OkHttpClient,
) : SigningService {

    override suspend fun sign(txBytes: ByteArray): ByteArray = withContext(Dispatchers.IO) {
        val url = "$endpoint/v1/keys/$keyName/sign"
        val reqBody = json.encodeToString(
            VaultSignRequest.serializer(),
            VaultSignRequest(data = HashUtils.bytesToHex(txBytes))
        )

        val request = Request.Builder()
            .url(url)
            .post(reqBody.toRequestBody(JSON_MEDIA_TYPE))
            .build()

        val response = client.newCall(request).execute()
        require(response.isSuccessful) {
            "vault sign failed with status ${response.code}: ${response.body?.string()}"
        }

        val body = response.body?.string() ?: throw IllegalStateException("empty response")
        val result = json.decodeFromString(VaultSignResponse.serializer(), body)
        HashUtils.hexToBytes(result.signature)
    }

    override fun getAddress(): String = address

    override fun getPubKey(): ByteArray = pubKey

    companion object {
        private val json = Json { ignoreUnknownKeys = true }
        private val JSON_MEDIA_TYPE = "application/json".toMediaType()

        suspend fun create(endpoint: String, keyName: String): VaultService = withContext(Dispatchers.IO) {
            require(endpoint.isNotBlank()) { "vault endpoint is required" }
            require(keyName.isNotBlank()) { "key name is required" }

            val baseUrl = endpoint.trimEnd('/')
            val client = OkHttpClient.Builder()
                .connectTimeout(10, TimeUnit.SECONDS)
                .readTimeout(10, TimeUnit.SECONDS)
                .build()

            // Fetch address
            val addrResp = doGet(client, "$baseUrl/v1/keys/$keyName/address")
            val addrResult = json.decodeFromString(VaultAddressResponse.serializer(), addrResp)

            // Fetch public key
            val pubKeyResp = doGet(client, "$baseUrl/v1/keys/$keyName/pubkey")
            val pubKeyResult = json.decodeFromString(VaultPubKeyResponse.serializer(), pubKeyResp)
            val pubKeyBytes = HashUtils.hexToBytes(pubKeyResult.pubKey)

            VaultService(baseUrl, keyName, addrResult.address, pubKeyBytes, client)
        }

        private fun doGet(client: OkHttpClient, url: String): String {
            val request = Request.Builder().url(url).get().build()
            val response = client.newCall(request).execute()
            require(response.isSuccessful) {
                "vault request failed with status ${response.code}: ${response.body?.string()}"
            }
            return response.body?.string() ?: throw IllegalStateException("empty response")
        }
    }
}

@Serializable
private data class VaultSignRequest(val data: String)

@Serializable
private data class VaultSignResponse(val signature: String)

@Serializable
private data class VaultAddressResponse(val address: String)

@Serializable
private data class VaultPubKeyResponse(@SerialName("pubkey") val pubKey: String)

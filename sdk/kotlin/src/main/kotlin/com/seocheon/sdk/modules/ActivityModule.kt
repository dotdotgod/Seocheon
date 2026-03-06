package com.seocheon.sdk.modules

import com.seocheon.sdk.errors.SdkError
import com.seocheon.sdk.infrastructure.ChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.internal.tx.*
import com.seocheon.sdk.types.ActivityItem
import com.seocheon.sdk.types.GetActivitiesResponse
import com.seocheon.sdk.types.GetQuotaResponse
import com.seocheon.sdk.types.SubmitActivityResponse
import com.seocheon.sdk.utils.EpochUtils
import com.seocheon.sdk.utils.HashUtils
import kotlinx.serialization.json.*

/**
 * Activity module for submitting activities and querying quotas.
 */
class ActivityModule(
    private val client: ChainClient,
    private val signer: SigningService,
    private val pipelineConfig: PipelineConfig,
) {

    /**
     * Submits an activity to the chain.
     */
    suspend fun submit(activityHash: String, contentUri: String): SubmitActivityResponse {
        if (!HashUtils.verifyActivityHash(activityHash)) {
            throw SdkError.InvalidActivityHash()
        }
        if (contentUri.isBlank()) {
            throw SdkError.InvalidContentURI()
        }

        val msg = MsgSubmitActivity(
            submitter = signer.getAddress(),
            activityHash = activityHash,
            contentUri = contentUri,
        )

        val result = Pipeline.executeTx(
            client, signerAdapter(), pipelineConfig,
            TxRequest(message = msg),
        )

        val blockHeight = result.height
        val epochNumber = EpochUtils.computeEpoch(blockHeight)
        val windowNumber = EpochUtils.computeWindow(blockHeight)

        return SubmitActivityResponse(
            txHash = result.txHash,
            blockHeight = blockHeight,
            epochNumber = epochNumber,
            windowNumber = windowNumber,
            quotaRemaining = 0, // Could be parsed from events
        )
    }

    /**
     * Queries activities for a node in a given epoch.
     */
    suspend fun getActivities(nodeId: String, epochNumber: Long? = null): GetActivitiesResponse {
        var path = "/seocheon/activity/v1/nodes/$nodeId/activities"
        if (epochNumber != null) {
            path += "?epoch_number=$epochNumber"
        }

        val resp = client.queryRest(path)
        val activities = resp.jsonObject["activities"]?.jsonArray?.map { item ->
            val obj = item.jsonObject
            ActivityItem(
                activityHash = obj["activity_hash"]?.jsonPrimitive?.content ?: "",
                contentUri = obj["content_uri"]?.jsonPrimitive?.content ?: "",
                blockHeight = obj["block_height"]?.jsonPrimitive?.longOrNull ?: 0L,
                windowNumber = obj["window_number"]?.jsonPrimitive?.longOrNull ?: 0L,
                txHash = obj["tx_hash"]?.jsonPrimitive?.content ?: "",
            )
        } ?: emptyList()

        val totalCount = resp.jsonObject["total_count"]?.jsonPrimitive?.longOrNull ?: activities.size.toLong()

        return GetActivitiesResponse(activities = activities, totalCount = totalCount)
    }

    /**
     * Queries the activity quota for a node.
     */
    suspend fun getQuota(): GetQuotaResponse {
        val nodeId = resolveOwnNodeId()
        val epochResp = client.queryRest("/seocheon/activity/v1/epoch-info")
        val currentEpoch = epochResp.jsonObject["epoch_number"]?.jsonPrimitive?.longOrNull ?: 0L

        val resp = client.queryRest("/seocheon/activity/v1/nodes/$nodeId/epochs/$currentEpoch")
        val obj = resp.jsonObject

        val quotaLimit = obj["quota_limit"]?.jsonPrimitive?.longOrNull ?: 0L
        val quotaUsed = obj["quota_used"]?.jsonPrimitive?.longOrNull ?: 0L

        return GetQuotaResponse(
            epochNumber = currentEpoch,
            quotaTotal = quotaLimit,
            quotaUsed = quotaUsed,
            quotaRemaining = maxOf(0L, quotaLimit - quotaUsed),
            isFeegrant = false,
            feegrantExpiry = null,
        )
    }

    private fun signerAdapter(): Signer = object : Signer {
        override suspend fun sign(data: ByteArray): ByteArray = signer.sign(data)
        override fun getAddress(): String = signer.getAddress()
        override fun getPubKey(): ByteArray = signer.getPubKey()
    }

    private suspend fun resolveOwnNodeId(): String {
        val address = signer.getAddress()
        val resp = client.queryRest("/seocheon/node/v1/nodes/by-agent/$address")
        return resp.jsonObject["node"]?.jsonObject?.get("id")?.jsonPrimitive?.content
            ?: throw SdkError.NodeNotFound()
    }
}

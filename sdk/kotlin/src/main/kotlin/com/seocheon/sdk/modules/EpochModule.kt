package com.seocheon.sdk.modules

import com.seocheon.sdk.constants.ChainConstants
import com.seocheon.sdk.infrastructure.ChainClient
import com.seocheon.sdk.types.EpochInfoResponse
import com.seocheon.sdk.types.QualificationResponse
import com.seocheon.sdk.types.WindowActivity
import com.seocheon.sdk.utils.EpochUtils
import kotlinx.serialization.json.*

/**
 * Epoch module for querying epoch/window state and activity qualification.
 */
class EpochModule(
    private val client: ChainClient,
) {

    /**
     * Returns current epoch and window information.
     */
    suspend fun getInfo(): EpochInfoResponse {
        val blockResp = client.getLatestBlock()
        val block = blockResp.jsonObject["block"]?.jsonObject
        val header = block?.get("header")?.jsonObject
        val blockHeight = header?.get("height")?.jsonPrimitive?.longOrNull ?: 0L

        val epochNumber = EpochUtils.computeEpoch(blockHeight)
        val windowNumber = EpochUtils.computeWindow(blockHeight)

        val epochStart = EpochUtils.epochStartBlock(epochNumber)
        val epochEnd = EpochUtils.epochEndBlock(epochNumber)
        val windowStart = EpochUtils.windowStartBlock(epochStart, windowNumber)
        val windowEnd = EpochUtils.windowEndBlock(windowStart)

        val epochProgress = if (epochEnd > epochStart)
            "%.2f".format((blockHeight - epochStart).toDouble() / (epochEnd - epochStart) * 100)
        else "0.00"

        val windowProgress = if (windowEnd >= windowStart)
            "%.2f".format((blockHeight - windowStart).toDouble() / ChainConstants.WINDOW_LENGTH * 100)
        else "0.00"

        return EpochInfoResponse(
            blockHeight = blockHeight,
            epochNumber = epochNumber,
            epochStartBlock = epochStart,
            epochEndBlock = epochEnd,
            epochProgress = epochProgress,
            windowNumber = windowNumber,
            windowStartBlock = windowStart,
            windowEndBlock = windowEnd,
            windowProgress = windowProgress,
            blocksUntilNextWindow = windowEnd - blockHeight,
            blocksUntilNextEpoch = epochEnd - blockHeight,
        )
    }

    /**
     * Queries reward qualification status for a node.
     */
    suspend fun getQualification(nodeId: String, epochNumber: Long? = null): QualificationResponse {
        val currentEpoch = epochNumber ?: EpochUtils.computeEpoch(getCurrentBlockHeight())

        val resp = client.queryRest("/seocheon/activity/v1/qualification/$nodeId?epoch_number=$currentEpoch")
        val obj = resp.jsonObject

        val windowDetail = obj["window_detail"]?.jsonArray?.map { item ->
            val w = item.jsonObject
            WindowActivity(
                windowNumber = w["window_number"]?.jsonPrimitive?.longOrNull ?: 0L,
                submissionCount = w["submission_count"]?.jsonPrimitive?.longOrNull ?: 0L,
                hasActivity = w["has_activity"]?.jsonPrimitive?.booleanOrNull ?: false,
            )
        } ?: emptyList()

        val activeWindows = obj["active_windows"]?.jsonPrimitive?.longOrNull
            ?: windowDetail.count { it.hasActivity }.toLong()
        val requiredWindows = ChainConstants.MIN_ACTIVE_WINDOWS
        val isQualified = activeWindows >= requiredWindows

        val elapsedWindows = obj["elapsed_windows"]?.jsonPrimitive?.longOrNull
            ?: ChainConstants.WINDOWS_PER_EPOCH
        val remainingNeeded = maxOf(0, requiredWindows - activeWindows)
        val remainingWindows = ChainConstants.WINDOWS_PER_EPOCH - elapsedWindows
        val canStillQualify = activeWindows + remainingWindows >= requiredWindows

        return QualificationResponse(
            epochNumber = currentEpoch,
            totalWindows = ChainConstants.WINDOWS_PER_EPOCH,
            elapsedWindows = elapsedWindows,
            activeWindows = activeWindows,
            requiredWindows = requiredWindows,
            isQualified = isQualified,
            remainingNeeded = remainingNeeded,
            canStillQualify = canStillQualify,
            windowDetail = windowDetail,
        )
    }

    private suspend fun getCurrentBlockHeight(): Long {
        val blockResp = client.getLatestBlock()
        val block = blockResp.jsonObject["block"]?.jsonObject
        val header = block?.get("header")?.jsonObject
        return header?.get("height")?.jsonPrimitive?.longOrNull ?: 0L
    }
}

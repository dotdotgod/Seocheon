package com.seocheon.sdk.types

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class SubmitActivityResponse(
    @SerialName("tx_hash") val txHash: String,
    @SerialName("block_height") val blockHeight: Long,
    @SerialName("window_number") val windowNumber: Long,
    @SerialName("epoch_number") val epochNumber: Long,
    @SerialName("quota_remaining") val quotaRemaining: Long,
)

@Serializable
data class ActivityItem(
    @SerialName("activity_hash") val activityHash: String,
    @SerialName("content_uri") val contentUri: String,
    @SerialName("block_height") val blockHeight: Long,
    @SerialName("window_number") val windowNumber: Long,
    @SerialName("tx_hash") val txHash: String,
)

@Serializable
data class GetActivitiesResponse(
    val activities: List<ActivityItem>,
    @SerialName("total_count") val totalCount: Long,
)

@Serializable
data class GetQuotaResponse(
    @SerialName("epoch_number") val epochNumber: Long,
    @SerialName("quota_total") val quotaTotal: Long,
    @SerialName("quota_used") val quotaUsed: Long,
    @SerialName("quota_remaining") val quotaRemaining: Long,
    @SerialName("is_feegrant") val isFeegrant: Boolean,
    @SerialName("feegrant_expiry") val feegrantExpiry: Long? = null,
)

@Serializable
data class EpochInfoResponse(
    @SerialName("block_height") val blockHeight: Long,
    @SerialName("epoch_number") val epochNumber: Long,
    @SerialName("epoch_start_block") val epochStartBlock: Long,
    @SerialName("epoch_end_block") val epochEndBlock: Long,
    @SerialName("epoch_progress") val epochProgress: String,
    @SerialName("window_number") val windowNumber: Long,
    @SerialName("window_start_block") val windowStartBlock: Long,
    @SerialName("window_end_block") val windowEndBlock: Long,
    @SerialName("window_progress") val windowProgress: String,
    @SerialName("blocks_until_next_window") val blocksUntilNextWindow: Long,
    @SerialName("blocks_until_next_epoch") val blocksUntilNextEpoch: Long,
)

@Serializable
data class WindowActivity(
    @SerialName("window_number") val windowNumber: Long,
    @SerialName("submission_count") val submissionCount: Long,
    @SerialName("has_activity") val hasActivity: Boolean,
)

@Serializable
data class QualificationResponse(
    @SerialName("epoch_number") val epochNumber: Long,
    @SerialName("total_windows") val totalWindows: Long,
    @SerialName("elapsed_windows") val elapsedWindows: Long,
    @SerialName("active_windows") val activeWindows: Long,
    @SerialName("required_windows") val requiredWindows: Long,
    @SerialName("is_qualified") val isQualified: Boolean,
    @SerialName("remaining_needed") val remainingNeeded: Long,
    @SerialName("can_still_qualify") val canStillQualify: Boolean,
    @SerialName("window_detail") val windowDetail: List<WindowActivity>,
)

@Serializable
data class NodeInfoResponse(
    @SerialName("node_id") val nodeId: String,
    val operator: String,
    @SerialName("agent_address") val agentAddress: String,
    val status: String,
    val description: String,
    val website: String,
    val tags: List<String>,
    @SerialName("commission_rate") val commissionRate: String,
    @SerialName("agent_share") val agentShare: String,
    @SerialName("total_delegation") val totalDelegation: String,
    @SerialName("self_delegation") val selfDelegation: String,
    @SerialName("validator_address") val validatorAddress: String,
    @SerialName("registered_at") val registeredAt: Long,
)

@Serializable
data class NodeSummary(
    @SerialName("node_id") val nodeId: String,
    val status: String,
    val tags: List<String>,
    @SerialName("total_delegation") val totalDelegation: String,
    val description: String,
)

@Serializable
data class NodeSearchResponse(
    val nodes: List<NodeSummary>,
    @SerialName("total_count") val totalCount: Long,
)

@Serializable
data class PendingRewardsResponse(
    @SerialName("delegation_reward") val delegationReward: String,
    @SerialName("activity_reward") val activityReward: String,
    @SerialName("total_reward") val totalReward: String,
    @SerialName("commission_total") val commissionTotal: String,
    @SerialName("operator_share") val operatorShare: String,
    @SerialName("agent_share") val agentShare: String,
)

@Serializable
data class WithdrawRewardsResponse(
    @SerialName("tx_hash") val txHash: String,
    @SerialName("withdrawn_total") val withdrawnTotal: String,
    @SerialName("to_operator") val toOperator: String,
    @SerialName("to_agent") val toAgent: String,
)

@Serializable
data class BalanceResponse(
    val address: String,
    val balance: String,
    @SerialName("balance_kkot") val balanceKkot: String,
)

@Serializable
data class SendTokensResponse(
    @SerialName("tx_hash") val txHash: String,
    @SerialName("block_height") val blockHeight: Long,
)

@Serializable
data class BlockInfoResponse(
    @SerialName("block_height") val blockHeight: Long,
    @SerialName("block_time") val blockTime: String,
    @SerialName("chain_id") val chainId: String,
    @SerialName("num_txs") val numTxs: Long,
)

@Serializable
data class EventAttribute(
    val key: String,
    val value: String,
)

@Serializable
data class TxEvent(
    val type: String,
    val attributes: List<EventAttribute>,
)

@Serializable
data class TxResultResponse(
    @SerialName("tx_hash") val txHash: String,
    val height: Long,
    val code: UInt,
    @SerialName("gas_used") val gasUsed: Long,
    @SerialName("gas_wanted") val gasWanted: Long,
    @SerialName("raw_log") val rawLog: String,
    val events: List<TxEvent>,
)

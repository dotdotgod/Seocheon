package com.seocheon.sdk.modules

import com.seocheon.sdk.errors.SdkError
import com.seocheon.sdk.infrastructure.ChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.internal.tx.*
import com.seocheon.sdk.types.PendingRewardsResponse
import com.seocheon.sdk.types.WithdrawRewardsResponse
import kotlinx.serialization.json.*

/**
 * Rewards module for pending rewards and withdrawals.
 */
class RewardsModule(
    private val client: ChainClient,
    private val signer: SigningService,
    private val pipelineConfig: PipelineConfig,
) {

    /**
     * Queries pending rewards for a node.
     */
    suspend fun getPending(nodeId: String? = null): PendingRewardsResponse {
        val id = nodeId ?: resolveOwnNodeId()
        val nodeResp = client.queryRest("/seocheon/node/v1/nodes/$id")
        val node = nodeResp.jsonObject["node"]?.jsonObject ?: nodeResp.jsonObject
        val validatorAddr = node["validator_address"]?.jsonPrimitive?.content ?: ""
        val operatorAddr = node["operator"]?.jsonPrimitive?.content ?: ""
        val agentShareStr = node["agent_share"]?.jsonPrimitive?.content ?: "20"

        // Query outstanding rewards
        var delegationReward = "0"
        var activityReward = "0"
        try {
            val rewardsResp = client.queryRest(
                "/cosmos/distribution/v1beta1/delegators/$operatorAddr/rewards/$validatorAddr"
            )
            val rewards = rewardsResp.jsonObject["rewards"]?.jsonArray
            rewards?.forEach { r ->
                val denom = r.jsonObject["denom"]?.jsonPrimitive?.content
                val amount = r.jsonObject["amount"]?.jsonPrimitive?.content ?: "0"
                if (denom == "uppyeo") delegationReward = amount.split(".").first()
            }
        } catch (_: Exception) { /* no rewards */ }

        // Query commission
        var commissionTotal = "0"
        try {
            val commResp = client.queryRest(
                "/cosmos/distribution/v1beta1/validators/$validatorAddr/commission"
            )
            val commission = commResp.jsonObject["commission"]?.jsonObject?.get("commission")?.jsonArray
            commission?.forEach { c ->
                val denom = c.jsonObject["denom"]?.jsonPrimitive?.content
                val amount = c.jsonObject["amount"]?.jsonPrimitive?.content ?: "0"
                if (denom == "uppyeo") commissionTotal = amount.split(".").first()
            }
        } catch (_: Exception) { /* no commission */ }

        val agentSharePct = parseAgentShare(agentShareStr)
        val totalRewardLong = delegationReward.toLongOrNull() ?: 0L
        val commissionLong = commissionTotal.toLongOrNull() ?: 0L
        val totalAll = totalRewardLong + commissionLong
        val agentAmount = (totalAll * agentSharePct / 100)
        val operatorAmount = totalAll - agentAmount

        return PendingRewardsResponse(
            delegationReward = delegationReward,
            activityReward = activityReward,
            totalReward = totalAll.toString(),
            commissionTotal = commissionTotal,
            operatorShare = operatorAmount.toString(),
            agentShare = agentAmount.toString(),
        )
    }

    /**
     * Withdraws pending rewards.
     */
    suspend fun withdraw(nodeId: String? = null): WithdrawRewardsResponse {
        val address = signer.getAddress()
        val msg = MsgWithdrawNodeCommission(operator = address)

        val result = Pipeline.executeTx(
            client, signerAdapter(), pipelineConfig,
            TxRequest(message = msg),
        )

        // Parse withdrawn amounts from events
        var withdrawnTotal = "0"
        var toOperator = "0"
        var toAgent = "0"

        for (event in result.events) {
            if (event.type == "withdraw_rewards" || event.type == "withdraw_commission") {
                for (attr in event.attributes) {
                    when (attr.key) {
                        "amount" -> withdrawnTotal = attr.value.replace("uppyeo", "")
                    }
                }
            }
        }

        return WithdrawRewardsResponse(
            txHash = result.txHash,
            withdrawnTotal = withdrawnTotal,
            toOperator = toOperator,
            toAgent = toAgent,
        )
    }

    private fun parseAgentShare(share: String): Long {
        val cleaned = share.replace("%", "").trim()
        return cleaned.toLongOrNull() ?: 20
    }

    private suspend fun resolveOwnNodeId(): String {
        val address = signer.getAddress()
        val resp = client.queryRest("/seocheon/node/v1/nodes/by-agent/$address")
        return resp.jsonObject["node_id"]?.jsonPrimitive?.content
            ?: throw SdkError.NodeNotFound()
    }

    private fun signerAdapter(): Signer = object : Signer {
        override suspend fun sign(data: ByteArray): ByteArray = signer.sign(data)
        override fun getAddress(): String = signer.getAddress()
        override fun getPubKey(): ByteArray = signer.getPubKey()
    }
}

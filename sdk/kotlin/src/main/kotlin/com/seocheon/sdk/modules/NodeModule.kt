package com.seocheon.sdk.modules

import com.seocheon.sdk.infrastructure.ChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.internal.tx.MsgConfirmDelegation
import com.seocheon.sdk.internal.tx.Pipeline
import com.seocheon.sdk.internal.tx.PipelineConfig
import com.seocheon.sdk.internal.tx.TxRequest
import com.seocheon.sdk.types.DelegationStatusResponse
import com.seocheon.sdk.types.NodeInfoResponse
import com.seocheon.sdk.types.NodeSearchResponse
import com.seocheon.sdk.types.NodeSummary
import com.seocheon.sdk.types.TxResultResponse
import kotlinx.serialization.json.*

/**
 * Node module for querying node information and searching.
 */
class NodeModule(
    private val client: ChainClient,
    private val signer: SigningService,
    private val pipelineConfig: PipelineConfig? = null,
) {

    /**
     * Retrieves detailed information for a node.
     */
    suspend fun getInfo(nodeId: String? = null): NodeInfoResponse {
        val id = nodeId ?: resolveOwnNodeId()
        val resp = client.queryRest("/seocheon/node/v1/nodes/$id")
        val node = resp.jsonObject["node"]?.jsonObject ?: resp.jsonObject

        return NodeInfoResponse(
            nodeId = node["id"]?.jsonPrimitive?.content ?: id,
            operator = node["operator"]?.jsonPrimitive?.content ?: "",
            agentAddress = node["agent_address"]?.jsonPrimitive?.content ?: "",
            status = node["status"]?.jsonPrimitive?.content ?: "UNSPECIFIED",
            description = node["description"]?.jsonPrimitive?.content ?: "",
            website = node["website"]?.jsonPrimitive?.content ?: "",
            tags = node["tags"]?.jsonArray?.map { it.jsonPrimitive.content } ?: emptyList(),
            commissionRate = node["commission_rate"]?.jsonPrimitive?.content ?: "0",
            agentShare = node["agent_share"]?.jsonPrimitive?.content ?: "0",
            totalDelegation = node["total_delegation"]?.jsonPrimitive?.content ?: "0",
            selfDelegation = node["self_delegation"]?.jsonPrimitive?.content ?: "0",
            validatorAddress = node["validator_address"]?.jsonPrimitive?.content ?: "",
            registeredAt = node["registered_at"]?.jsonPrimitive?.longOrNull ?: 0L,
        )
    }

    /**
     * Searches for nodes with optional filters.
     */
    suspend fun search(
        tag: String? = null,
        status: String? = null,
        limit: Int = 20,
        orderBy: String = "delegation",
    ): NodeSearchResponse {
        val params = mutableListOf<String>()
        tag?.let { params.add("tag=$it") }
        status?.let { params.add("status=$it") }
        params.add("limit=$limit")
        params.add("order_by=$orderBy")

        val query = if (params.isNotEmpty()) "?${params.joinToString("&")}" else ""
        val resp = client.queryRest("/seocheon/node/v1/nodes$query")

        val nodes = resp.jsonObject["nodes"]?.jsonArray?.map { item ->
            val obj = item.jsonObject
            NodeSummary(
                nodeId = obj["id"]?.jsonPrimitive?.content ?: "",
                status = obj["status"]?.jsonPrimitive?.content ?: "UNSPECIFIED",
                tags = obj["tags"]?.jsonArray?.map { it.jsonPrimitive.content } ?: emptyList(),
                totalDelegation = obj["total_delegation"]?.jsonPrimitive?.content ?: "0",
                description = obj["description"]?.jsonPrimitive?.content ?: "",
            )
        } ?: emptyList()

        val totalCount = resp.jsonObject["total_count"]?.jsonPrimitive?.longOrNull ?: nodes.size.toLong()
        return NodeSearchResponse(nodes = nodes, totalCount = totalCount)
    }

    /**
     * Queries the delegation confirmation status.
     */
    suspend fun getDelegationStatus(delegatorAddress: String, validatorAddress: String): DelegationStatusResponse {
        val resp = client.queryRest("/seocheon/node/v1/delegation-confirmation/$delegatorAddress/$validatorAddress")
        val obj = resp.jsonObject
        return DelegationStatusResponse(
            expiryEpoch = obj["expiry_epoch"]?.jsonPrimitive?.longOrNull ?: 0L,
            currentEpoch = obj["current_epoch"]?.jsonPrimitive?.longOrNull ?: 0L,
            inRenewalWindow = obj["in_renewal_window"]?.jsonPrimitive?.booleanOrNull ?: false,
            renewalWindowStart = obj["renewal_window_start"]?.jsonPrimitive?.longOrNull ?: 0L,
        )
    }

    /**
     * Confirms delegation to a validator.
     */
    suspend fun confirmDelegation(validatorAddress: String): TxResultResponse {
        val cfg = pipelineConfig ?: throw com.seocheon.sdk.errors.SdkError.InvalidConfig("PipelineConfig required for TX operations")
        val msg = MsgConfirmDelegation(
            delegatorAddress = signer.getAddress(),
            validatorAddress = validatorAddress,
        )
        val result = Pipeline.executeTx(client, signerAdapter(), cfg, TxRequest(message = msg))
        return TxResultResponse(
            txHash = result.txHash,
            height = result.height,
            code = result.code,
            gasUsed = result.gasUsed,
            gasWanted = result.gasWanted,
            rawLog = result.rawLog,
            events = emptyList(),
        )
    }

    private fun signerAdapter(): com.seocheon.sdk.internal.tx.Signer = object : com.seocheon.sdk.internal.tx.Signer {
        override suspend fun sign(data: ByteArray): ByteArray = signer.sign(data)
        override fun getAddress(): String = signer.getAddress()
        override fun getPubKey(): ByteArray = signer.getPubKey()
    }

    private suspend fun resolveOwnNodeId(): String {
        val address = signer.getAddress()
        val resp = client.queryRest("/seocheon/node/v1/nodes/by-agent/$address")
        return resp.jsonObject["node"]?.jsonObject?.get("id")?.jsonPrimitive?.content
            ?: throw com.seocheon.sdk.errors.SdkError.NodeNotFound()
    }
}

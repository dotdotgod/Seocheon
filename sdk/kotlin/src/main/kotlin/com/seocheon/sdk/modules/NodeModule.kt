package com.seocheon.sdk.modules

import com.seocheon.sdk.infrastructure.ChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.types.NodeInfoResponse
import com.seocheon.sdk.types.NodeSearchResponse
import com.seocheon.sdk.types.NodeSummary
import kotlinx.serialization.json.*

/**
 * Node module for querying node information and searching.
 */
class NodeModule(
    private val client: ChainClient,
    private val signer: SigningService,
) {

    /**
     * Retrieves detailed information for a node.
     */
    suspend fun getInfo(nodeId: String? = null): NodeInfoResponse {
        val id = nodeId ?: resolveOwnNodeId()
        val resp = client.queryRest("/seocheon/node/v1/nodes/$id")
        val node = resp.jsonObject["node"]?.jsonObject ?: resp.jsonObject

        return NodeInfoResponse(
            nodeId = node["node_id"]?.jsonPrimitive?.content ?: id,
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
        sortBy: String = "delegation",
        limit: Int = 20,
        offset: Int = 0,
    ): NodeSearchResponse {
        val params = mutableListOf<String>()
        tag?.let { params.add("tag=$it") }
        status?.let { params.add("status=$it") }
        params.add("sort_by=$sortBy")
        params.add("limit=$limit")
        params.add("offset=$offset")

        val query = if (params.isNotEmpty()) "?${params.joinToString("&")}" else ""
        val resp = client.queryRest("/seocheon/node/v1/nodes$query")

        val nodes = resp.jsonObject["nodes"]?.jsonArray?.map { item ->
            val obj = item.jsonObject
            NodeSummary(
                nodeId = obj["node_id"]?.jsonPrimitive?.content ?: "",
                status = obj["status"]?.jsonPrimitive?.content ?: "UNSPECIFIED",
                tags = obj["tags"]?.jsonArray?.map { it.jsonPrimitive.content } ?: emptyList(),
                totalDelegation = obj["total_delegation"]?.jsonPrimitive?.content ?: "0",
                description = obj["description"]?.jsonPrimitive?.content ?: "",
            )
        } ?: emptyList()

        val totalCount = resp.jsonObject["total_count"]?.jsonPrimitive?.longOrNull ?: nodes.size.toLong()
        return NodeSearchResponse(nodes = nodes, totalCount = totalCount)
    }

    private suspend fun resolveOwnNodeId(): String {
        val address = signer.getAddress()
        val resp = client.queryRest("/seocheon/node/v1/nodes/by-agent/$address")
        return resp.jsonObject["node_id"]?.jsonPrimitive?.content
            ?: throw com.seocheon.sdk.errors.SdkError.NodeNotFound()
    }
}

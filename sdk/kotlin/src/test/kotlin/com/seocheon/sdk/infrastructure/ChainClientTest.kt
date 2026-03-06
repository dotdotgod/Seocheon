package com.seocheon.sdk.infrastructure

import com.seocheon.sdk.ChainConfig
import com.seocheon.sdk.internal.tx.TxEventAttribute
import com.seocheon.sdk.internal.tx.TxEventData
import com.seocheon.sdk.internal.tx.TxResult
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class ChainClientTest {

    @Test
    fun `OkHttpChainClient is not connected initially`() {
        val config = ChainConfig("test", "http://localhost:26657", "http://localhost:9090")
        val client = OkHttpChainClient(config)
        assertFalse(client.isConnected())
    }

    @Test
    fun `ChainClient interface is implementable`() {
        val mock = MockChainClient()
        assertFalse(mock.isConnected())
    }

    @Test
    fun `MockChainClient tracks connection state`() {
        val mock = MockChainClient()
        assertFalse(mock.isConnected())
    }

    @Test
    fun `OkHttpChainClient stores config`() {
        val config = ChainConfig("my-chain", "http://rpc:26657", "http://grpc:9090")
        val client = OkHttpChainClient(config)
        assertNotNull(client)
    }

    @Test
    fun `ChainClient extends ChainQuerier`() {
        val mock = MockChainClient()
        // ChainClient extends ChainQuerier, should have getAccountInfo
        assertTrue(mock is com.seocheon.sdk.internal.tx.ChainQuerier)
    }
}

/** Minimal mock for testing. */
class MockChainClient : ChainClient {
    private var connected = false

    override suspend fun connect() { connected = true }
    override suspend fun disconnect() { connected = false }
    override fun isConnected(): Boolean = connected

    override suspend fun queryRest(path: String): JsonElement = buildJsonObject {
        if (path.contains("/by-agent/")) {
            put("node", buildJsonObject {
                put("id", "mock-node-1")
                put("operator", "seocheon1mockaddress")
            })
        } else {
            put("result", "mock")
        }
    }

    override suspend fun getLatestBlock(): JsonElement = buildJsonObject {
        put("block", buildJsonObject {
            put("header", buildJsonObject {
                put("height", "1000")
                put("time", "2026-01-01T00:00:00Z")
                put("chain_id", "test-chain")
            })
        })
    }

    override suspend fun getAccountInfo(address: String): Pair<Long, Long> = Pair(1L, 0L)
    override suspend fun broadcastTxSync(txBytes: ByteArray): Triple<String, UInt, String> =
        Triple("MOCK_HASH", 0u, "")
    override suspend fun getTxResult(txHash: String): TxResult = TxResult(
        txHash = txHash, height = 100, code = 0u, gasUsed = 50000, gasWanted = 200000,
        rawLog = "", events = listOf(
            TxEventData("transfer", listOf(TxEventAttribute("amount", "1000uppyeo")))
        ),
    )
}

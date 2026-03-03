package com.seocheon.sdk.internal.tx

import com.seocheon.sdk.errors.SdkError
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class PipelineTest {

    private class MockSigner : Signer {
        override suspend fun sign(data: ByteArray): ByteArray = ByteArray(64) { 0x01 }
        override fun getAddress(): String = "seocheon1mockaddress"
        override fun getPubKey(): ByteArray = ByteArray(33) { if (it == 0) 0x02 else 0x01 }
    }

    private class MockQuerier(
        private val broadcastCode: UInt = 0u,
        private val txResult: TxResult = TxResult(txHash = "ABCD1234", height = 100, code = 0u),
    ) : ChainQuerier {
        override suspend fun getAccountInfo(address: String): Pair<Long, Long> = Pair(1L, 5L)
        override suspend fun broadcastTxSync(txBytes: ByteArray): Triple<String, UInt, String> =
            Triple("ABCD1234", broadcastCode, if (broadcastCode != 0u) "error" else "")
        override suspend fun getTxResult(txHash: String): TxResult = txResult
    }

    @Test
    fun `executeTx succeeds with valid inputs`() = runTest {
        val result = Pipeline.executeTx(
            MockQuerier(),
            MockSigner(),
            PipelineConfig("test-chain"),
            TxRequest(MsgSubmitActivity("addr", "a".repeat(64), "uri")),
        )
        assertEquals("ABCD1234", result.txHash)
        assertEquals(100L, result.height)
    }

    @Test
    fun `executeTx uses default gas for message type`() = runTest {
        val result = Pipeline.executeTx(
            MockQuerier(),
            MockSigner(),
            PipelineConfig("test-chain"),
            TxRequest(MsgSubmitActivity("addr", "a".repeat(64), "uri")),
        )
        assertNotNull(result)
    }

    @Test
    fun `executeTx with custom gas limit`() = runTest {
        val result = Pipeline.executeTx(
            MockQuerier(),
            MockSigner(),
            PipelineConfig("test-chain"),
            TxRequest(MsgSubmitActivity("addr", "a".repeat(64), "uri"), gasLimit = 500000),
        )
        assertNotNull(result)
    }

    @Test
    fun `executeTx throws on broadcast failure code`() = runTest {
        assertThrows(SdkError::class.java) {
            kotlinx.coroutines.test.runTest {
                Pipeline.executeTx(
                    MockQuerier(broadcastCode = 1203u),
                    MockSigner(),
                    PipelineConfig("test-chain"),
                    TxRequest(MsgSubmitActivity("addr", "a".repeat(64), "uri")),
                )
            }
        }
    }

    @Test
    fun `executeTx throws SigningFailed on signer error`() = runTest {
        val failingSigner = object : Signer {
            override suspend fun sign(data: ByteArray): ByteArray = throw RuntimeException("key error")
            override fun getAddress(): String = "seocheon1mock"
            override fun getPubKey(): ByteArray = ByteArray(33) { 0x02 }
        }
        assertThrows(SdkError.SigningFailed::class.java) {
            kotlinx.coroutines.test.runTest {
                Pipeline.executeTx(
                    MockQuerier(),
                    failingSigner,
                    PipelineConfig("test-chain"),
                    TxRequest(MsgSubmitActivity("addr", "a".repeat(64), "uri")),
                )
            }
        }
    }

    @Test
    fun `executeTx throws BroadcastFailed on network error`() = runTest {
        val failingQuerier = object : ChainQuerier {
            override suspend fun getAccountInfo(address: String): Pair<Long, Long> = Pair(1L, 5L)
            override suspend fun broadcastTxSync(txBytes: ByteArray): Triple<String, UInt, String> =
                throw RuntimeException("network error")
            override suspend fun getTxResult(txHash: String): TxResult =
                throw RuntimeException("not found")
        }
        assertThrows(SdkError.BroadcastFailed::class.java) {
            kotlinx.coroutines.test.runTest {
                Pipeline.executeTx(
                    failingQuerier,
                    MockSigner(),
                    PipelineConfig("test-chain"),
                    TxRequest(MsgSubmitActivity("addr", "a".repeat(64), "uri")),
                )
            }
        }
    }

    @Test
    fun `PipelineConfig has sensible defaults`() {
        val config = PipelineConfig("test-chain")
        assertEquals("test-chain", config.chainId)
        assertEquals(250L, config.gasPrice)
        assertEquals(30000L, config.confirmTimeoutMs)
        assertEquals(1000L, config.pollIntervalMs)
    }
}

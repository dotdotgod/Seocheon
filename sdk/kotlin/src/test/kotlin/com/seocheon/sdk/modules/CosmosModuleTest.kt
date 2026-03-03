package com.seocheon.sdk.modules

import com.seocheon.sdk.infrastructure.MockChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.internal.tx.PipelineConfig
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class CosmosModuleTest {

    private val mockSigner = object : SigningService {
        override suspend fun sign(txBytes: ByteArray): ByteArray = ByteArray(64) { 0x01 }
        override fun getAddress(): String = "seocheon1mockaddress"
        override fun getPubKey(): ByteArray = ByteArray(33) { if (it == 0) 0x02 else 0x01 }
    }

    private fun createModule() = CosmosModule(
        MockChainClient().also { kotlinx.coroutines.test.runTest { it.connect() } },
        mockSigner,
        PipelineConfig("test-chain"),
    )

    @Test
    fun `getBalance returns response`() = runTest {
        val module = createModule()
        val result = module.getBalance("seocheon1test")
        assertNotNull(result)
        assertEquals("seocheon1test", result.address)
    }

    @Test
    fun `getBalance uses signer address by default`() = runTest {
        val module = createModule()
        val result = module.getBalance()
        assertEquals("seocheon1mockaddress", result.address)
    }

    @Test
    fun `getBlockInfo returns response`() = runTest {
        val module = createModule()
        val result = module.getBlockInfo()
        assertNotNull(result)
        assertTrue(result.blockHeight >= 0)
    }

    @Test
    fun `sendTokens returns response`() = runTest {
        val module = createModule()
        val result = module.sendTokens("seocheon1target", "1000000")
        assertNotNull(result.txHash)
    }

    @Test
    fun `getTxResult returns response`() = runTest {
        val module = createModule()
        val result = module.getTxResult("ABCD1234")
        assertNotNull(result)
        assertEquals("ABCD1234", result.txHash)
    }
}

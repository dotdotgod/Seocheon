package com.seocheon.sdk.modules

import com.seocheon.sdk.infrastructure.MockChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.internal.tx.PipelineConfig
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class RewardsModuleTest {

    private val mockSigner = object : SigningService {
        override suspend fun sign(txBytes: ByteArray): ByteArray = ByteArray(64) { 0x01 }
        override fun getAddress(): String = "seocheon1mockaddress"
        override fun getPubKey(): ByteArray = ByteArray(33) { if (it == 0) 0x02 else 0x01 }
    }

    private fun createModule() = RewardsModule(
        MockChainClient().also { kotlinx.coroutines.test.runTest { it.connect() } },
        mockSigner,
        PipelineConfig("test-chain"),
    )

    @Test
    fun `getPending returns response for node ID`() = runTest {
        val module = createModule()
        val result = module.getPending("node1")
        assertNotNull(result)
    }

    @Test
    fun `getPending has default agent share`() = runTest {
        val module = createModule()
        val result = module.getPending("node1")
        assertNotNull(result.agentShare)
        assertNotNull(result.operatorShare)
    }

    @Test
    fun `withdraw returns response`() = runTest {
        val module = createModule()
        val result = module.withdraw()
        assertNotNull(result.txHash)
    }

    @Test
    fun `module is constructable`() {
        val module = createModule()
        assertNotNull(module)
    }

    @Test
    fun `getPending returns string amounts`() = runTest {
        val module = createModule()
        val result = module.getPending("node1")
        assertNotNull(result.totalReward)
        assertNotNull(result.commissionTotal)
    }
}

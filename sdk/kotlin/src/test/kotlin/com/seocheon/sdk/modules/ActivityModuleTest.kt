package com.seocheon.sdk.modules

import com.seocheon.sdk.errors.SdkError
import com.seocheon.sdk.infrastructure.MockChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import com.seocheon.sdk.internal.tx.PipelineConfig
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class ActivityModuleTest {

    private val mockSigner = object : SigningService {
        override suspend fun sign(txBytes: ByteArray): ByteArray = ByteArray(64) { 0x01 }
        override fun getAddress(): String = "seocheon1mockaddress"
        override fun getPubKey(): ByteArray = ByteArray(33) { if (it == 0) 0x02 else 0x01 }
    }

    private fun createModule() = ActivityModule(
        MockChainClient().also { kotlinx.coroutines.test.runTest { it.connect() } },
        mockSigner,
        PipelineConfig("test-chain"),
    )

    @Test
    fun `submit rejects invalid activity hash`() = runTest {
        val module = createModule()
        assertThrows(SdkError.InvalidActivityHash::class.java) {
            kotlinx.coroutines.test.runTest {
                module.submit("invalid", "https://example.com")
            }
        }
    }

    @Test
    fun `submit rejects blank content URI`() = runTest {
        val module = createModule()
        assertThrows(SdkError.InvalidContentURI::class.java) {
            kotlinx.coroutines.test.runTest {
                module.submit("a".repeat(64), "")
            }
        }
    }

    @Test
    fun `submit accepts valid hash and URI`() = runTest {
        val module = createModule()
        val result = module.submit("a".repeat(64), "https://example.com")
        assertNotNull(result.txHash)
        assertTrue(result.blockHeight >= 0)
    }

    @Test
    fun `getActivities returns response`() = runTest {
        val module = createModule()
        val result = module.getActivities("node1")
        assertNotNull(result)
    }

    @Test
    fun `getQuota returns response`() = runTest {
        val module = createModule()
        val result = module.getQuota("node1")
        assertNotNull(result)
    }
}

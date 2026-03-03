package com.seocheon.sdk.modules

import com.seocheon.sdk.infrastructure.MockChainClient
import com.seocheon.sdk.infrastructure.signing.SigningService
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class NodeModuleTest {

    private val mockSigner = object : SigningService {
        override suspend fun sign(txBytes: ByteArray): ByteArray = ByteArray(64)
        override fun getAddress(): String = "seocheon1mockaddress"
        override fun getPubKey(): ByteArray = ByteArray(33) { if (it == 0) 0x02 else 0x01 }
    }

    private fun createModule() = NodeModule(
        MockChainClient().also { kotlinx.coroutines.test.runTest { it.connect() } },
        mockSigner,
    )

    @Test
    fun `getInfo returns response for node ID`() = runTest {
        val module = createModule()
        val result = module.getInfo("node1")
        assertNotNull(result)
    }

    @Test
    fun `search returns response`() = runTest {
        val module = createModule()
        val result = module.search()
        assertNotNull(result)
    }

    @Test
    fun `search with tag filter`() = runTest {
        val module = createModule()
        val result = module.search(tag = "ai-agent")
        assertNotNull(result)
    }

    @Test
    fun `search with status filter`() = runTest {
        val module = createModule()
        val result = module.search(status = "ACTIVE")
        assertNotNull(result)
    }
}

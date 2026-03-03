package com.seocheon.sdk.modules

import com.seocheon.sdk.infrastructure.MockChainClient
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class EpochModuleTest {

    private fun createModule() = EpochModule(
        MockChainClient().also { kotlinx.coroutines.test.runTest { it.connect() } },
    )

    @Test
    fun `getInfo returns epoch info`() = runTest {
        val module = createModule()
        val result = module.getInfo()
        assertNotNull(result)
        assertTrue(result.blockHeight >= 0)
        assertTrue(result.epochNumber >= 0)
        assertTrue(result.windowNumber >= 0)
    }

    @Test
    fun `getQualification returns response`() = runTest {
        val module = createModule()
        val result = module.getQualification("node1", epochNumber = 0)
        assertNotNull(result)
        assertEquals(0L, result.epochNumber)
    }

    @Test
    fun `getQualification has correct required windows`() = runTest {
        val module = createModule()
        val result = module.getQualification("node1", epochNumber = 0)
        assertEquals(8L, result.requiredWindows)
    }
}

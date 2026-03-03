package com.seocheon.sdk.infrastructure.signing

import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class SigningServiceTest {

    private val testMnemonic =
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

    @Test
    fun `DirectService creates from valid mnemonic`() {
        val service = DirectService(testMnemonic)
        assertNotNull(service)
    }

    @Test
    fun `DirectService getAddress returns seocheon1 prefix`() {
        val service = DirectService(testMnemonic)
        assertTrue(service.getAddress().startsWith("seocheon1"))
    }

    @Test
    fun `DirectService getPubKey returns 33 bytes`() {
        val service = DirectService(testMnemonic)
        assertEquals(33, service.getPubKey().size)
    }

    @Test
    fun `DirectService sign returns 64-byte signature`() = runTest {
        val service = DirectService(testMnemonic)
        val sig = service.sign("test data".toByteArray())
        assertEquals(64, sig.size)
    }

    @Test
    fun `DirectService rejects blank mnemonic`() {
        assertThrows(IllegalArgumentException::class.java) {
            DirectService("")
        }
    }

    @Test
    fun `DirectService is deterministic`() {
        val s1 = DirectService(testMnemonic)
        val s2 = DirectService(testMnemonic)
        assertEquals(s1.getAddress(), s2.getAddress())
        assertArrayEquals(s1.getPubKey(), s2.getPubKey())
    }

    @Test
    fun `DirectService implements SigningService`() {
        val service: SigningService = DirectService(testMnemonic)
        assertNotNull(service.getAddress())
    }

    @Test
    fun `VaultService rejects blank endpoint`() = runTest {
        assertThrows(IllegalArgumentException::class.java) {
            kotlinx.coroutines.test.runTest {
                VaultService.create("", "key1")
            }
        }
    }
}

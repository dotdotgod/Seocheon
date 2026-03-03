package com.seocheon.sdk.utils

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class HashUtilsTest {

    @Test
    fun `verifyActivityHash accepts valid 64-char hex`() {
        val hash = "a".repeat(64)
        assertTrue(HashUtils.verifyActivityHash(hash))
    }

    @Test
    fun `verifyActivityHash accepts mixed case hex`() {
        val hash = "aAbBcCdDeEfF0123456789" + "0".repeat(42)
        assertTrue(HashUtils.verifyActivityHash(hash))
    }

    @Test
    fun `verifyActivityHash rejects short string`() {
        assertFalse(HashUtils.verifyActivityHash("abc123"))
    }

    @Test
    fun `verifyActivityHash rejects long string`() {
        assertFalse(HashUtils.verifyActivityHash("a".repeat(65)))
    }

    @Test
    fun `verifyActivityHash rejects non-hex characters`() {
        val hash = "g" + "0".repeat(63)
        assertFalse(HashUtils.verifyActivityHash(hash))
    }

    @Test
    fun `verifyActivityHash rejects empty string`() {
        assertFalse(HashUtils.verifyActivityHash(""))
    }

    @Test
    fun `computeActivityHash produces valid 64-char hex`() {
        val data = "test data".toByteArray()
        val hash = HashUtils.computeActivityHash(data)
        assertEquals(64, hash.length)
        assertTrue(HashUtils.verifyActivityHash(hash))
    }

    @Test
    fun `computeActivityHash is deterministic`() {
        val data = "hello seocheon".toByteArray()
        val hash1 = HashUtils.computeActivityHash(data)
        val hash2 = HashUtils.computeActivityHash(data)
        assertEquals(hash1, hash2)
    }
}

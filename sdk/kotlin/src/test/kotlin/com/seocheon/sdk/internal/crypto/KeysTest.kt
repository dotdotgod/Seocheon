package com.seocheon.sdk.internal.crypto

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class KeysTest {

    @Test
    fun `fromBytes creates key from 32 bytes`() {
        val bytes = ByteArray(32) { 1 }
        val key = PrivateKey.fromBytes(bytes)
        assertNotNull(key)
        assertArrayEquals(bytes, key.bytes())
    }

    @Test
    fun `fromBytes rejects non-32-byte input`() {
        assertThrows(IllegalArgumentException::class.java) {
            PrivateKey.fromBytes(ByteArray(16))
        }
    }

    @Test
    fun `pubKey returns 33-byte compressed key`() {
        val key = PrivateKey.fromBytes(ByteArray(32) { 1 })
        val pubKey = key.pubKey()
        assertEquals(33, pubKey.size)
        assertTrue(pubKey[0] == 0x02.toByte() || pubKey[0] == 0x03.toByte())
    }

    @Test
    fun `sign returns 64-byte signature`() {
        val key = PrivateKey.fromBytes(ByteArray(32) { 1 })
        val data = "test data".toByteArray()
        val sig = key.sign(data)
        assertEquals(64, sig.size)
    }

    @Test
    fun `verify validates correct signature`() {
        val key = PrivateKey.fromBytes(ByteArray(32) { 1 })
        val data = "test data".toByteArray()
        val sig = key.sign(data)
        assertTrue(PrivateKey.verify(key.pubKey(), data, sig))
    }
}

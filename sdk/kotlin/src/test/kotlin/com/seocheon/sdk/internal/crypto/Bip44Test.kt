package com.seocheon.sdk.internal.crypto

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class Bip44Test {

    private val testMnemonic =
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

    @Test
    fun `deriveKeyFromMnemonic produces 32-byte private key`() {
        val key = Bip44.deriveKeyFromMnemonic(testMnemonic)
        assertEquals(32, key.bytes().size)
    }

    @Test
    fun `deriveKeyFromMnemonic produces 33-byte public key`() {
        val key = Bip44.deriveKeyFromMnemonic(testMnemonic)
        assertEquals(33, key.pubKey().size)
    }

    @Test
    fun `derivation is deterministic`() {
        val key1 = Bip44.deriveKeyFromMnemonic(testMnemonic)
        val key2 = Bip44.deriveKeyFromMnemonic(testMnemonic)
        assertArrayEquals(key1.bytes(), key2.bytes())
        assertArrayEquals(key1.pubKey(), key2.pubKey())
    }

    @Test
    fun `isValidMnemonic returns true for valid mnemonic`() {
        assertTrue(Bip44.isValidMnemonic(testMnemonic))
    }
}

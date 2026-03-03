package com.seocheon.sdk.internal.crypto

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class AddressTest {

    private val testMnemonic =
        "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

    @Test
    fun `fromPubKey produces seocheon1 prefix address`() {
        val key = Bip44.deriveKeyFromMnemonic(testMnemonic)
        val address = Address.fromPubKey(key.pubKey())
        assertTrue(address.startsWith("seocheon1"))
    }

    @Test
    fun `fromPubKey is deterministic`() {
        val key = Bip44.deriveKeyFromMnemonic(testMnemonic)
        val addr1 = Address.fromPubKey(key.pubKey())
        val addr2 = Address.fromPubKey(key.pubKey())
        assertEquals(addr1, addr2)
    }

    @Test
    fun `fromPubKey rejects non-33-byte input`() {
        assertThrows(IllegalArgumentException::class.java) {
            Address.fromPubKey(ByteArray(20))
        }
    }
}

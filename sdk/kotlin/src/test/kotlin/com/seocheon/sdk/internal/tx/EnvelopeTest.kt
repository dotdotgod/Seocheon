package com.seocheon.sdk.internal.tx

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class EnvelopeTest {

    @Test
    fun `encodeTxBody produces non-empty bytes`() {
        val msg = MsgSubmitActivity("addr", "a".repeat(64), "uri")
        val body = Envelope.encodeTxBody(listOf(msg))
        assertTrue(body.isNotEmpty())
    }

    @Test
    fun `encodeTxBody with memo includes memo`() {
        val msg = MsgSubmitActivity("addr", "a".repeat(64), "uri")
        val withMemo = Envelope.encodeTxBody(listOf(msg), "test memo")
        val withoutMemo = Envelope.encodeTxBody(listOf(msg))
        assertTrue(withMemo.size > withoutMemo.size)
    }

    @Test
    fun `encodeAuthInfo produces non-empty bytes`() {
        val pubKey = ByteArray(33) { 0x02.toByte() }
        val coins = listOf(Coin("uppyeo", "50000000"))
        val result = Envelope.encodeAuthInfo(pubKey, 0, coins, 200000)
        assertTrue(result.isNotEmpty())
    }

    @Test
    fun `encodeSignDoc produces non-empty bytes`() {
        val body = byteArrayOf(1, 2, 3)
        val authInfo = byteArrayOf(4, 5, 6)
        val result = Envelope.encodeSignDoc(body, authInfo, "test-chain", 0)
        assertTrue(result.isNotEmpty())
    }

    @Test
    fun `encodeTxRaw produces non-empty bytes`() {
        val body = byteArrayOf(1, 2, 3)
        val authInfo = byteArrayOf(4, 5, 6)
        val sig = byteArrayOf(7, 8, 9)
        val result = Envelope.encodeTxRaw(body, authInfo, sig)
        assertTrue(result.isNotEmpty())
    }

    @Test
    fun `encodeTxRaw includes signature`() {
        val body = byteArrayOf(1, 2)
        val authInfo = byteArrayOf(3, 4)
        val sig = ByteArray(64) { it.toByte() }
        val result = Envelope.encodeTxRaw(body, authInfo, sig)
        // Result should contain the signature bytes somewhere
        assertTrue(result.size > body.size + authInfo.size)
    }
}

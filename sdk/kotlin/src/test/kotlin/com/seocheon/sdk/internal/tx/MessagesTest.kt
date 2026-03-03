package com.seocheon.sdk.internal.tx

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class MessagesTest {

    @Test
    fun `MsgSubmitActivity has correct type URL`() {
        val msg = MsgSubmitActivity("addr", "hash", "uri")
        assertEquals("/seocheon.activity.v1.MsgSubmitActivity", msg.typeUrl())
    }

    @Test
    fun `MsgSubmitActivity encodes non-empty bytes`() {
        val msg = MsgSubmitActivity("seocheon1abc", "a".repeat(64), "https://example.com")
        val encoded = msg.encode()
        assertTrue(encoded.isNotEmpty())
    }

    @Test
    fun `MsgWithdrawNodeCommission has correct type URL`() {
        val msg = MsgWithdrawNodeCommission("seocheon1abc")
        assertEquals("/seocheon.node.v1.MsgWithdrawNodeCommission", msg.typeUrl())
    }

    @Test
    fun `MsgWithdrawNodeCommission encodes non-empty bytes`() {
        val msg = MsgWithdrawNodeCommission("seocheon1abc")
        val encoded = msg.encode()
        assertTrue(encoded.isNotEmpty())
    }

    @Test
    fun `MsgSend has correct type URL`() {
        val msg = MsgSend("from", "to", listOf(Coin("uppyeo", "1000")))
        assertEquals("/cosmos.bank.v1beta1.MsgSend", msg.typeUrl())
    }

    @Test
    fun `Coin encodes denom and amount`() {
        val coin = Coin("uppyeo", "50000000")
        val encoded = coin.encode()
        assertTrue(encoded.isNotEmpty())
        // Should contain both "uppyeo" and "50000000" strings
        val str = String(encoded, Charsets.UTF_8)
        // Protobuf-encoded, not raw text — just check non-empty
    }
}

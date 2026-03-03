package com.seocheon.sdk.internal.tx

import org.junit.jupiter.api.Assertions.*
import org.junit.jupiter.api.Test

class ProtobufTest {

    @Test
    fun `encodeVarint encodes small values`() {
        val result = Protobuf.encodeVarint(1L)
        assertArrayEquals(byteArrayOf(0x01), result)
    }

    @Test
    fun `encodeVarint encodes multi-byte values`() {
        val result = Protobuf.encodeVarint(300L)
        assertEquals(2, result.size)
        // 300 = 0b100101100 → varint: 0xAC 0x02
        assertArrayEquals(byteArrayOf(0xAC.toByte(), 0x02), result)
    }

    @Test
    fun `encodeVarint encodes zero`() {
        val result = Protobuf.encodeVarint(0L)
        assertArrayEquals(byteArrayOf(0x00), result)
    }

    @Test
    fun `encodeFieldVarint omits zero value`() {
        val result = Protobuf.encodeFieldVarint(1, 0L)
        assertEquals(0, result.size)
    }

    @Test
    fun `encodeFieldBytes omits empty data`() {
        val result = Protobuf.encodeFieldBytes(1, ByteArray(0))
        assertEquals(0, result.size)
    }

    @Test
    fun `encodeFieldString encodes non-empty string`() {
        val result = Protobuf.encodeFieldString(1, "hello")
        assertTrue(result.isNotEmpty())
        // field 1, wire type 2 → tag = 0x0A, then length 5, then "hello"
        assertEquals(0x0A.toByte(), result[0])
        assertEquals(5.toByte(), result[1])
    }
}

package com.seocheon.sdk.internal.tx

/**
 * Minimal protobuf wire format encoder.
 * Only supports types needed for Cosmos TX encoding:
 * - varint (wire type 0)
 * - length-delimited (wire type 2)
 */
object Protobuf {

    /** Encodes a ULong as a protobuf varint. */
    fun encodeVarint(v: ULong): ByteArray {
        val buf = ByteArray(10)
        var value = v
        var n = 0
        while (value >= 0x80u) {
            buf[n] = (value.toByte().toInt() or 0x80).toByte()
            value = value shr 7
            n++
        }
        buf[n] = value.toByte()
        return buf.sliceArray(0..n)
    }

    /** Encodes a Long as a varint. */
    fun encodeVarint(v: Long): ByteArray = encodeVarint(v.toULong())

    /** Encodes a varint field (wire type 0). Omits zero values. */
    fun encodeFieldVarint(fieldNumber: Int, value: ULong): ByteArray {
        if (value == 0uL) return ByteArray(0)
        val tag = encodeVarint((fieldNumber.toULong() shl 3) or 0uL)
        return tag + encodeVarint(value)
    }

    /** Encodes a varint field from Long. */
    fun encodeFieldVarint(fieldNumber: Int, value: Long): ByteArray =
        encodeFieldVarint(fieldNumber, value.toULong())

    /** Encodes a length-delimited field (wire type 2). Omits empty data. */
    fun encodeFieldBytes(fieldNumber: Int, data: ByteArray): ByteArray {
        if (data.isEmpty()) return ByteArray(0)
        val tag = encodeVarint((fieldNumber.toULong() shl 3) or 2uL)
        val length = encodeVarint(data.size.toULong())
        return tag + length + data
    }

    /** Encodes a string field (wire type 2). Omits empty strings. */
    fun encodeFieldString(fieldNumber: Int, value: String): ByteArray {
        if (value.isEmpty()) return ByteArray(0)
        return encodeFieldBytes(fieldNumber, value.toByteArray(Charsets.UTF_8))
    }

    /** Concatenates multiple byte arrays, skipping empty ones. */
    fun concatBytes(vararg parts: ByteArray): ByteArray {
        val totalSize = parts.sumOf { it.size }
        val result = ByteArray(totalSize)
        var offset = 0
        for (part in parts) {
            part.copyInto(result, offset)
            offset += part.size
        }
        return result
    }
}

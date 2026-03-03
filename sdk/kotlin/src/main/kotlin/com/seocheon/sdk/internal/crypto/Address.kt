package com.seocheon.sdk.internal.crypto

import com.seocheon.sdk.constants.ChainConstants
import java.security.MessageDigest

/**
 * Cosmos bech32 address encoding utilities.
 */
object Address {

    /**
     * Derives a Cosmos bech32 address from a compressed 33-byte public key.
     * Returns a "seocheon1..." address.
     */
    fun fromPubKey(pubKey: ByteArray): String {
        require(pubKey.size == 33) {
            "public key must be 33 bytes (compressed), got ${pubKey.size}"
        }

        // SHA256 hash
        val sha256 = MessageDigest.getInstance("SHA-256").digest(pubKey)

        // RIPEMD160 hash
        val ripemd160 = ripemd160(sha256)

        // Bech32 encode with "seocheon" HRP
        return bech32Encode(ChainConstants.ADDRESS_PREFIX, ripemd160)
    }

    private fun ripemd160(data: ByteArray): ByteArray {
        val digest = org.bouncycastle.crypto.digests.RIPEMD160Digest()
        digest.update(data, 0, data.size)
        val result = ByteArray(digest.digestSize)
        digest.doFinal(result, 0)
        return result
    }

    private fun bech32Encode(hrp: String, data: ByteArray): String {
        val converted = convertBits(data, 8, 5, true)

        val values = converted.toMutableList()
        values.addAll(listOf(0, 0, 0, 0, 0, 0))
        val polymod = bech32Polymod(expandHRP(hrp) + values) xor 1
        for (i in 0 until 6) {
            values[converted.size + i] = ((polymod shr (5 * (5 - i))) and 31).toByte()
        }

        val charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
        val sb = StringBuilder(hrp)
        sb.append('1')
        for (v in values) {
            sb.append(charset[v.toInt() and 0xFF])
        }
        return sb.toString()
    }

    private fun convertBits(data: ByteArray, fromBits: Int, toBits: Int, pad: Boolean): List<Byte> {
        var acc = 0
        var bits = 0
        val result = mutableListOf<Byte>()
        val maxV = (1 shl toBits) - 1

        for (b in data) {
            acc = (acc shl fromBits) or (b.toInt() and 0xFF)
            bits += fromBits
            while (bits >= toBits) {
                bits -= toBits
                result.add(((acc shr bits) and maxV).toByte())
            }
        }

        if (pad) {
            if (bits > 0) {
                result.add(((acc shl (toBits - bits)) and maxV).toByte())
            }
        }

        return result
    }

    private fun expandHRP(hrp: String): List<Byte> {
        val result = mutableListOf<Byte>()
        for (c in hrp) {
            result.add((c.code shr 5).toByte())
        }
        result.add(0)
        for (c in hrp) {
            result.add((c.code and 31).toByte())
        }
        return result
    }

    private fun bech32Polymod(values: List<Byte>): Int {
        val gen = intArrayOf(0x3b6a57b2, 0x26508e6d, 0x1ea119fa.toInt(), 0x3d4233dd, 0x2a1462b3)
        var chk = 1
        for (v in values) {
            val top = chk shr 25
            chk = ((chk and 0x1ffffff) shl 5) xor (v.toInt() and 0xFF)
            for (i in 0 until 5) {
                if ((top shr i) and 1 == 1) {
                    chk = chk xor gen[i]
                }
            }
        }
        return chk
    }
}

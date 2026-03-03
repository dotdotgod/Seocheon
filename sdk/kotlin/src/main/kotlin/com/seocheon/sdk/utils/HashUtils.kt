package com.seocheon.sdk.utils

import java.security.MessageDigest

/**
 * SHA-256 hash utilities for activity hash verification and computation.
 */
object HashUtils {

    private val HEX_CHARS = "0123456789abcdef".toCharArray()

    /**
     * Validates that the given string is a valid activity hash.
     * A valid activity hash is exactly 64 hex characters (representing 32 bytes / SHA-256).
     */
    fun verifyActivityHash(hash: String): Boolean {
        if (hash.length != 64) return false
        return hash.all { it in '0'..'9' || it in 'a'..'f' || it in 'A'..'F' }
    }

    /**
     * Computes a SHA-256 hash of the given data and returns
     * it as a 64-character lowercase hex string suitable for use as an activity_hash.
     */
    fun computeActivityHash(data: ByteArray): String {
        val digest = MessageDigest.getInstance("SHA-256")
        val hashBytes = digest.digest(data)
        return bytesToHex(hashBytes)
    }

    /**
     * Converts a byte array to a lowercase hex string.
     */
    fun bytesToHex(bytes: ByteArray): String {
        val sb = StringBuilder(bytes.size * 2)
        for (b in bytes) {
            sb.append(HEX_CHARS[(b.toInt() shr 4) and 0x0F])
            sb.append(HEX_CHARS[b.toInt() and 0x0F])
        }
        return sb.toString()
    }

    /**
     * Converts a hex string to a byte array.
     */
    fun hexToBytes(hex: String): ByteArray {
        require(hex.length % 2 == 0) { "Hex string must have even length" }
        return ByteArray(hex.length / 2) { i ->
            val hi = Character.digit(hex[i * 2], 16)
            val lo = Character.digit(hex[i * 2 + 1], 16)
            require(hi >= 0 && lo >= 0) { "Invalid hex character" }
            ((hi shl 4) or lo).toByte()
        }
    }
}

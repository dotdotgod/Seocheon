package com.seocheon.sdk.infrastructure.signing

/**
 * Interface for transaction signing.
 */
interface SigningService {
    /** Signs the given transaction bytes and returns the signature. */
    suspend fun sign(txBytes: ByteArray): ByteArray

    /** Returns the signer's bech32 address. */
    fun getAddress(): String

    /** Returns the signer's compressed public key bytes. */
    fun getPubKey(): ByteArray
}

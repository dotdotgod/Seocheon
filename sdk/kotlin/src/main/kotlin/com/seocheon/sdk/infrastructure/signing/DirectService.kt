package com.seocheon.sdk.infrastructure.signing

import com.seocheon.sdk.internal.crypto.Address
import com.seocheon.sdk.internal.crypto.Bip44
import com.seocheon.sdk.internal.crypto.PrivateKey

/**
 * Signs transactions using a mnemonic directly (test/development only).
 */
class DirectService(mnemonic: String) : SigningService {

    private val privKey: PrivateKey
    private val address: String
    private val pubKey: ByteArray

    init {
        require(mnemonic.isNotBlank()) { "mnemonic is required" }
        privKey = Bip44.deriveKeyFromMnemonic(mnemonic)
        pubKey = privKey.pubKey()
        address = Address.fromPubKey(pubKey)
    }

    override suspend fun sign(txBytes: ByteArray): ByteArray = privKey.sign(txBytes)

    override fun getAddress(): String = address

    override fun getPubKey(): ByteArray = pubKey
}

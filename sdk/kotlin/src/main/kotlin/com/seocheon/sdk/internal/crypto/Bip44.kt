package com.seocheon.sdk.internal.crypto

import org.bitcoinj.crypto.HDKeyDerivation
import org.bitcoinj.crypto.MnemonicCode
import org.bitcoinj.crypto.ChildNumber

/**
 * BIP39/BIP44 key derivation for Cosmos-compatible chains.
 * Path: m/44'/118'/0'/0/0
 */
object Bip44 {

    /**
     * Derives a secp256k1 private key from a BIP39 mnemonic
     * using the Cosmos BIP44 path: m/44'/118'/0'/0/0
     */
    fun deriveKeyFromMnemonic(mnemonic: String): PrivateKey {
        val words = mnemonic.trim().split("\\s+".toRegex())
        MnemonicCode.INSTANCE.check(words)

        val seed = MnemonicCode.toSeed(words, "")

        // Create master key
        val masterKey = HDKeyDerivation.createMasterPrivateKey(seed)

        // Derive m/44'/118'/0'/0/0
        val purpose = HDKeyDerivation.deriveChildKey(masterKey, ChildNumber(44, true))
        val coinType = HDKeyDerivation.deriveChildKey(purpose, ChildNumber(118, true))
        val account = HDKeyDerivation.deriveChildKey(coinType, ChildNumber(0, true))
        val change = HDKeyDerivation.deriveChildKey(account, ChildNumber(0, false))
        val addressKey = HDKeyDerivation.deriveChildKey(change, ChildNumber(0, false))

        return PrivateKey.fromBytes(addressKey.privKeyBytes)
    }

    /**
     * Validates a BIP39 mnemonic phrase.
     */
    fun isValidMnemonic(mnemonic: String): Boolean {
        return try {
            val words = mnemonic.trim().split("\\s+".toRegex())
            MnemonicCode.INSTANCE.check(words)
            true
        } catch (_: Exception) {
            false
        }
    }
}

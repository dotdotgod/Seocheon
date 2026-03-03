package com.seocheon.sdk.infrastructure.signing

import com.seocheon.sdk.internal.crypto.Address
import com.seocheon.sdk.internal.crypto.PrivateKey
import com.seocheon.sdk.utils.HashUtils
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import java.io.File
import java.security.MessageDigest
import javax.crypto.Cipher
import javax.crypto.spec.IvParameterSpec
import javax.crypto.spec.SecretKeySpec

/**
 * Signs transactions using a local encrypted keystore file.
 * Compatible with the Web3 Secret Storage format (scrypt + AES-128-CTR).
 */
class KeystoreService private constructor(
    private val privKey: PrivateKey,
    private val address: String,
    private val pubKey: ByteArray,
) : SigningService {

    override suspend fun sign(txBytes: ByteArray): ByteArray = privKey.sign(txBytes)

    override fun getAddress(): String = address

    override fun getPubKey(): ByteArray = pubKey

    companion object {
        private val json = Json { ignoreUnknownKeys = true }

        fun create(keystorePath: String, passphrase: String): KeystoreService {
            require(keystorePath.isNotBlank()) { "keystore path is required" }
            require(passphrase.isNotBlank()) { "passphrase is required" }

            val key = loadAndDecrypt(keystorePath, passphrase)
            val pubKey = key.pubKey()
            val address = Address.fromPubKey(pubKey)
            return KeystoreService(key, address, pubKey)
        }

        private fun loadAndDecrypt(path: String, passphrase: String): PrivateKey {
            val data = File(path).readText()
            val ks = json.decodeFromString<KeystoreFile>(data)

            require(ks.crypto.kdf == "scrypt") {
                "unsupported KDF: ${ks.crypto.kdf} (only scrypt is supported)"
            }
            require(ks.crypto.cipher == "aes-128-ctr") {
                "unsupported cipher: ${ks.crypto.cipher} (only aes-128-ctr is supported)"
            }

            val salt = HashUtils.hexToBytes(ks.crypto.kdfParams.salt)
            val iv = HashUtils.hexToBytes(ks.crypto.cipherParams.iv)
            val cipherText = HashUtils.hexToBytes(ks.crypto.cipherText)
            val mac = HashUtils.hexToBytes(ks.crypto.mac)

            val dkLen = if (ks.crypto.kdfParams.dkLen > 0) ks.crypto.kdfParams.dkLen else 32

            // Derive key using scrypt
            val derivedKey = org.bouncycastle.crypto.generators.SCrypt.generate(
                passphrase.toByteArray(),
                salt,
                ks.crypto.kdfParams.n,
                ks.crypto.kdfParams.r,
                ks.crypto.kdfParams.p,
                dkLen
            )

            // Verify MAC: SHA256(derivedKey[16:32] + cipherText)
            val macInput = derivedKey.sliceArray(16 until 32) + cipherText
            val calculatedMac = MessageDigest.getInstance("SHA-256").digest(macInput)
            require(calculatedMac.contentEquals(mac)) {
                "MAC verification failed: incorrect passphrase or corrupted keystore"
            }

            // Decrypt using AES-128-CTR
            val aesKey = derivedKey.sliceArray(0 until 16)
            val cipher = Cipher.getInstance("AES/CTR/NoPadding")
            cipher.init(Cipher.DECRYPT_MODE, SecretKeySpec(aesKey, "AES"), IvParameterSpec(iv))
            val privKeyBytes = cipher.doFinal(cipherText)

            return PrivateKey.fromBytes(privKeyBytes)
        }
    }
}

@Serializable
private data class KeystoreFile(
    val crypto: KeystoreCrypto,
)

@Serializable
private data class KeystoreCrypto(
    val cipher: String,
    @SerialName("ciphertext") val cipherText: String,
    @SerialName("cipherparams") val cipherParams: KeystoreCipherParams,
    val kdf: String,
    @SerialName("kdfparams") val kdfParams: KeystoreKdfParams,
    val mac: String,
)

@Serializable
private data class KeystoreCipherParams(
    val iv: String,
)

@Serializable
private data class KeystoreKdfParams(
    @SerialName("dklen") val dkLen: Int = 32,
    val n: Int,
    val r: Int,
    val p: Int,
    val salt: String,
)

package com.seocheon.sdk.internal.crypto

import org.bouncycastle.crypto.params.ECPrivateKeyParameters
import org.bouncycastle.crypto.params.ECPublicKeyParameters
import org.bouncycastle.crypto.signers.ECDSASigner
import org.bouncycastle.crypto.signers.HMacDSAKCalculator
import org.bouncycastle.crypto.digests.SHA256Digest
import org.bouncycastle.jce.ECNamedCurveTable
import org.bouncycastle.jce.spec.ECNamedCurveParameterSpec
import org.bouncycastle.math.ec.ECPoint
import java.math.BigInteger
import java.security.MessageDigest

/**
 * secp256k1 private key wrapper using BouncyCastle.
 */
class PrivateKey private constructor(
    private val keyBytes: ByteArray,
    private val spec: ECNamedCurveParameterSpec,
) {
    private val d: BigInteger = BigInteger(1, keyBytes)

    /**
     * Returns the compressed 33-byte public key.
     */
    fun pubKey(): ByteArray {
        val q: ECPoint = spec.g.multiply(d).normalize()
        return q.getEncoded(true)
    }

    /**
     * Returns the raw 32-byte private key.
     */
    fun bytes(): ByteArray = keyBytes.copyOf()

    /**
     * Signs the given data using secp256k1. Hashes with SHA-256 first,
     * returns a 64-byte compact signature (R || S).
     */
    fun sign(data: ByteArray): ByteArray {
        val hash = MessageDigest.getInstance("SHA-256").digest(data)

        val signer = ECDSASigner(HMacDSAKCalculator(SHA256Digest()))
        val privParams = ECPrivateKeyParameters(d, org.bouncycastle.crypto.params.ECDomainParameters(
            spec.curve, spec.g, spec.n, spec.h
        ))
        signer.init(true, privParams)

        val components = signer.generateSignature(hash)
        var r = components[0]
        var s = components[1]

        // Enforce low-S (BIP-62)
        val halfN = spec.n.shiftRight(1)
        if (s > halfN) {
            s = spec.n.subtract(s)
        }

        val rBytes = bigIntTo32Bytes(r)
        val sBytes = bigIntTo32Bytes(s)
        return rBytes + sBytes
    }

    companion object {
        private val SPEC: ECNamedCurveParameterSpec = ECNamedCurveTable.getParameterSpec("secp256k1")

        /**
         * Creates a PrivateKey from raw 32-byte private key.
         */
        fun fromBytes(privKeyBytes: ByteArray): PrivateKey {
            require(privKeyBytes.size == 32) {
                "private key must be 32 bytes, got ${privKeyBytes.size}"
            }
            return PrivateKey(privKeyBytes.copyOf(), SPEC)
        }

        /**
         * Verifies a 64-byte compact signature against data using the public key.
         */
        fun verify(pubKeyBytes: ByteArray, data: ByteArray, sigBytes: ByteArray): Boolean {
            if (sigBytes.size != 64) return false

            val hash = MessageDigest.getInstance("SHA-256").digest(data)

            val point = SPEC.curve.decodePoint(pubKeyBytes)
            val pubParams = ECPublicKeyParameters(point, org.bouncycastle.crypto.params.ECDomainParameters(
                SPEC.curve, SPEC.g, SPEC.n, SPEC.h
            ))

            val r = BigInteger(1, sigBytes.sliceArray(0 until 32))
            val s = BigInteger(1, sigBytes.sliceArray(32 until 64))

            val verifier = ECDSASigner()
            verifier.init(false, pubParams)
            return verifier.verifySignature(hash, r, s)
        }

        private fun bigIntTo32Bytes(value: BigInteger): ByteArray {
            val bytes = value.toByteArray()
            return when {
                bytes.size == 32 -> bytes
                bytes.size > 32 -> bytes.sliceArray(bytes.size - 32 until bytes.size)
                else -> ByteArray(32 - bytes.size) + bytes
            }
        }
    }
}

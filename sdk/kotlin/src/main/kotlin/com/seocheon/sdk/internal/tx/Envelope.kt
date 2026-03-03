package com.seocheon.sdk.internal.tx

/**
 * TX envelope encoding: TxBody, AuthInfo, SignDoc, TxRaw.
 */
object Envelope {

    /** google.protobuf.Any: type_url(1, string), value(2, bytes) */
    private fun encodeAny(typeUrl: String, value: ByteArray): ByteArray = Protobuf.concatBytes(
        Protobuf.encodeFieldString(1, typeUrl),
        Protobuf.encodeFieldBytes(2, value),
    )

    /** cosmos.crypto.secp256k1.PubKey as Any */
    private fun encodePubKeyAny(pubKey: ByteArray): ByteArray {
        val innerMsg = Protobuf.encodeFieldBytes(1, pubKey)
        return encodeAny("/cosmos.crypto.secp256k1.PubKey", innerMsg)
    }

    /** ModeInfo for SIGN_MODE_DIRECT: ModeInfo { single { mode: 1 } } */
    private fun encodeModeInfoDirect(): ByteArray {
        val modeField = Protobuf.encodeFieldVarint(1, 1L) // SIGN_MODE_DIRECT = 1
        return Protobuf.encodeFieldBytes(1, modeField)
    }

    /** SignerInfo: public_key(1), mode_info(2), sequence(3) */
    private fun encodeSignerInfo(pubKey: ByteArray, sequence: Long): ByteArray {
        val pubKeyAny = encodePubKeyAny(pubKey)
        val modeInfo = encodeModeInfoDirect()
        return Protobuf.concatBytes(
            Protobuf.encodeFieldBytes(1, pubKeyAny),
            Protobuf.encodeFieldBytes(2, modeInfo),
            Protobuf.encodeFieldVarint(3, sequence),
        )
    }

    /** Fee: amount(1, repeated Coin), gas_limit(2, uint64) */
    private fun encodeFee(coins: List<Coin>, gasLimit: Long): ByteArray {
        val parts = mutableListOf<ByteArray>()
        for (coin in coins) {
            parts.add(Protobuf.encodeFieldBytes(1, coin.encode()))
        }
        parts.add(Protobuf.encodeFieldVarint(2, gasLimit))
        return Protobuf.concatBytes(*parts.toTypedArray())
    }

    /**
     * Encodes a TxBody.
     * Fields: messages(1, repeated Any), memo(2, string), timeout_height(3, uint64)
     */
    fun encodeTxBody(messages: List<MessageEncoder>, memo: String = "", timeoutHeight: Long = 0): ByteArray {
        val parts = mutableListOf<ByteArray>()
        for (msg in messages) {
            val anyBytes = encodeAny(msg.typeUrl(), msg.encode())
            parts.add(Protobuf.encodeFieldBytes(1, anyBytes))
        }
        if (memo.isNotEmpty()) {
            parts.add(Protobuf.encodeFieldString(2, memo))
        }
        if (timeoutHeight > 0) {
            parts.add(Protobuf.encodeFieldVarint(3, timeoutHeight))
        }
        return Protobuf.concatBytes(*parts.toTypedArray())
    }

    /**
     * Encodes an AuthInfo.
     * Fields: signer_infos(1, repeated SignerInfo), fee(2, Fee)
     */
    fun encodeAuthInfo(pubKey: ByteArray, sequence: Long, feeCoins: List<Coin>, gasLimit: Long): ByteArray {
        val signerInfo = encodeSignerInfo(pubKey, sequence)
        val fee = encodeFee(feeCoins, gasLimit)
        return Protobuf.concatBytes(
            Protobuf.encodeFieldBytes(1, signerInfo),
            Protobuf.encodeFieldBytes(2, fee),
        )
    }

    /**
     * Encodes a SignDoc for SIGN_MODE_DIRECT.
     * Fields: body_bytes(1), auth_info_bytes(2), chain_id(3), account_number(4)
     */
    fun encodeSignDoc(bodyBytes: ByteArray, authInfoBytes: ByteArray, chainId: String, accountNumber: Long): ByteArray =
        Protobuf.concatBytes(
            Protobuf.encodeFieldBytes(1, bodyBytes),
            Protobuf.encodeFieldBytes(2, authInfoBytes),
            Protobuf.encodeFieldString(3, chainId),
            Protobuf.encodeFieldVarint(4, accountNumber),
        )

    /**
     * Encodes a TxRaw for broadcast.
     * Fields: body_bytes(1), auth_info_bytes(2), signatures(3, repeated)
     */
    fun encodeTxRaw(bodyBytes: ByteArray, authInfoBytes: ByteArray, vararg signatures: ByteArray): ByteArray {
        val parts = mutableListOf(
            Protobuf.encodeFieldBytes(1, bodyBytes),
            Protobuf.encodeFieldBytes(2, authInfoBytes),
        )
        for (sig in signatures) {
            parts.add(Protobuf.encodeFieldBytes(3, sig))
        }
        return Protobuf.concatBytes(*parts.toTypedArray())
    }
}

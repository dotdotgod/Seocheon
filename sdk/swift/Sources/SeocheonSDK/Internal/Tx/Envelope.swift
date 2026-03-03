import Foundation

/// Protobuf envelope encoding for Cosmos transactions.
internal enum Envelope {
    /// Encodes google.protobuf.Any.
    static func encodeAny(typeURL: String, value: Data) -> Data {
        Protobuf.concatBytes(
            Protobuf.encodeFieldString(1, value: typeURL),
            Protobuf.encodeFieldBytes(2, data: value)
        )
    }

    /// Encodes cosmos.crypto.secp256k1.PubKey as Any.
    static func encodePubKeyAny(_ pubKey: Data) -> Data {
        let innerMsg = Protobuf.encodeFieldBytes(1, data: pubKey)
        return encodeAny(typeURL: "/cosmos.crypto.secp256k1.PubKey", value: innerMsg)
    }

    /// Encodes ModeInfo for SIGN_MODE_DIRECT.
    static func encodeModeInfoDirect() -> Data {
        let modeField = Protobuf.encodeFieldVarint(1, value: 1) // SIGN_MODE_DIRECT = 1
        return Protobuf.encodeFieldBytes(1, data: modeField)
    }

    /// Encodes SignerInfo.
    static func encodeSignerInfo(pubKey: Data, sequence: UInt64) -> Data {
        let pubKeyAny = encodePubKeyAny(pubKey)
        let modeInfo = encodeModeInfoDirect()
        return Protobuf.concatBytes(
            Protobuf.encodeFieldBytes(1, data: pubKeyAny),
            Protobuf.encodeFieldBytes(2, data: modeInfo),
            Protobuf.encodeFieldVarint(3, value: sequence)
        )
    }

    /// Encodes Fee.
    static func encodeFee(coins: [Coin], gasLimit: UInt64) -> Data {
        var parts: [Data] = []
        for coin in coins {
            parts.append(Protobuf.encodeFieldBytes(1, data: coin.encode()))
        }
        parts.append(Protobuf.encodeFieldVarint(2, value: gasLimit))
        return Protobuf.concatByteArray(parts)
    }

    /// Encodes TxBody.
    static func encodeTxBody(messages: [MessageEncoder], memo: String, timeoutHeight: UInt64) -> Data {
        var parts: [Data] = []
        for msg in messages {
            let anyBytes = encodeAny(typeURL: msg.typeURL, value: msg.encode())
            parts.append(Protobuf.encodeFieldBytes(1, data: anyBytes))
        }
        if !memo.isEmpty {
            parts.append(Protobuf.encodeFieldString(2, value: memo))
        }
        if timeoutHeight > 0 {
            parts.append(Protobuf.encodeFieldVarint(3, value: timeoutHeight))
        }
        return Protobuf.concatByteArray(parts)
    }

    /// Encodes AuthInfo.
    static func encodeAuthInfo(pubKey: Data, sequence: UInt64, feeCoins: [Coin], gasLimit: UInt64) -> Data {
        let signerInfo = encodeSignerInfo(pubKey: pubKey, sequence: sequence)
        let fee = encodeFee(coins: feeCoins, gasLimit: gasLimit)
        return Protobuf.concatBytes(
            Protobuf.encodeFieldBytes(1, data: signerInfo),
            Protobuf.encodeFieldBytes(2, data: fee)
        )
    }

    /// Encodes SignDoc for SIGN_MODE_DIRECT.
    static func encodeSignDoc(bodyBytes: Data, authInfoBytes: Data, chainID: String, accountNumber: UInt64) -> Data {
        Protobuf.concatBytes(
            Protobuf.encodeFieldBytes(1, data: bodyBytes),
            Protobuf.encodeFieldBytes(2, data: authInfoBytes),
            Protobuf.encodeFieldString(3, value: chainID),
            Protobuf.encodeFieldVarint(4, value: accountNumber)
        )
    }

    /// Encodes TxRaw for broadcast.
    static func encodeTxRaw(bodyBytes: Data, authInfoBytes: Data, signatures: [Data]) -> Data {
        var parts: [Data] = [
            Protobuf.encodeFieldBytes(1, data: bodyBytes),
            Protobuf.encodeFieldBytes(2, data: authInfoBytes),
        ]
        for sig in signatures {
            parts.append(Protobuf.encodeFieldBytes(3, data: sig))
        }
        return Protobuf.concatByteArray(parts)
    }
}

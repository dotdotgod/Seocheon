namespace Seocheon.Sdk.Internal.Tx;

/// <summary>
/// Transaction envelope construction: TxBody, AuthInfo, SignDoc, TxRaw.
/// </summary>
public static class Envelope
{
    // === PubKey type URL ===
    private const string PubKeyTypeUrl = "/cosmos.crypto.secp256k1.PubKey";

    /// <summary>
    /// Builds a protobuf Any message wrapping a typed message.
    /// Any: 1=type_url, 2=value
    /// </summary>
    public static byte[] EncodeAny(string typeUrl, byte[] value)
    {
        return Protobuf.Concat(
            Protobuf.EncodeString(1, typeUrl),
            Protobuf.EncodeBytes(2, value)
        );
    }

    /// <summary>
    /// Builds TxBody.
    /// TxBody: 1=messages (repeated Any), 2=memo, 3=timeout_height
    /// </summary>
    public static byte[] EncodeTxBody(byte[] messageAny, string memo = "", ulong timeoutHeight = 0)
    {
        var result = Protobuf.EncodeMessage(1, messageAny);
        if (!string.IsNullOrEmpty(memo))
            result = Protobuf.Concat(result, Protobuf.EncodeString(2, memo));
        if (timeoutHeight > 0)
            result = Protobuf.Concat(result, Protobuf.EncodeUint64(3, timeoutHeight));
        return result;
    }

    /// <summary>
    /// Builds AuthInfo.
    /// AuthInfo: 1=signer_infos (repeated), 2=fee
    /// SignerInfo: 1=public_key (Any), 2=mode_info, 3=sequence
    /// ModeInfo: 1=single (Single)
    /// Single: 1=mode (SIGN_MODE_DIRECT = 1)
    /// Fee: 1=amount (repeated Coin), 2=gas_limit
    /// Coin: 1=denom, 2=amount
    /// </summary>
    public static byte[] EncodeAuthInfo(byte[] pubKey, ulong sequence, ulong gasLimit, ulong feeAmount, string feeDenom)
    {
        // PubKey Any
        var pubKeyValue = Protobuf.EncodeBytes(1, pubKey);
        var pubKeyAny = EncodeAny(PubKeyTypeUrl, pubKeyValue);

        // ModeInfo: Single { mode = SIGN_MODE_DIRECT (1) }
        var single = Protobuf.EncodeUint64(1, 1); // SIGN_MODE_DIRECT
        var modeInfo = Protobuf.EncodeMessage(1, single);

        // SignerInfo
        var signerInfo = Protobuf.Concat(
            Protobuf.EncodeMessage(1, pubKeyAny),
            Protobuf.EncodeMessage(2, modeInfo),
            Protobuf.EncodeUint64(3, sequence)
        );

        // Fee
        var coin = Protobuf.Concat(
            Protobuf.EncodeString(1, feeDenom),
            Protobuf.EncodeString(2, feeAmount.ToString())
        );
        var fee = Protobuf.Concat(
            Protobuf.EncodeMessage(1, coin),
            Protobuf.EncodeUint64(2, gasLimit)
        );

        return Protobuf.Concat(
            Protobuf.EncodeMessage(1, signerInfo),
            Protobuf.EncodeMessage(2, fee)
        );
    }

    /// <summary>
    /// Builds SignDoc.
    /// SignDoc: 1=body_bytes, 2=auth_info_bytes, 3=chain_id, 4=account_number
    /// </summary>
    public static byte[] EncodeSignDoc(byte[] bodyBytes, byte[] authInfoBytes, string chainId, ulong accountNumber)
    {
        return Protobuf.Concat(
            Protobuf.EncodeBytes(1, bodyBytes),
            Protobuf.EncodeBytes(2, authInfoBytes),
            Protobuf.EncodeString(3, chainId),
            Protobuf.EncodeUint64(4, accountNumber)
        );
    }

    /// <summary>
    /// Builds TxRaw.
    /// TxRaw: 1=body_bytes, 2=auth_info_bytes, 3=signatures (repeated)
    /// </summary>
    public static byte[] EncodeTxRaw(byte[] bodyBytes, byte[] authInfoBytes, byte[] signature)
    {
        return Protobuf.Concat(
            Protobuf.EncodeBytes(1, bodyBytes),
            Protobuf.EncodeBytes(2, authInfoBytes),
            Protobuf.EncodeBytes(3, signature)
        );
    }
}

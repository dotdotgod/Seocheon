namespace Seocheon.Sdk.Internal.Tx;

/// <summary>
/// Protobuf message encoders for Seocheon transaction types.
/// </summary>
public static class Messages
{
    // === Message Type URLs ===

    public const string TypeMsgSubmitActivity = "/seocheon.activity.v1.MsgSubmitActivity";
    public const string TypeMsgWithdrawRewards = "/seocheon.node.v1.MsgWithdrawNodeCommission";
    public const string TypeMsgConfirmDelegation = "/seocheon.node.v1.MsgConfirmDelegation";
    public const string TypeMsgSend = "/cosmos.bank.v1beta1.MsgSend";

    /// <summary>
    /// Encodes MsgSubmitActivity.
    /// Fields: 1=submitter, 2=activity_hash, 3=content_uri
    /// </summary>
    public static byte[] EncodeMsgSubmitActivity(string submitter, string activityHash, string contentUri)
    {
        return Protobuf.Concat(
            Protobuf.EncodeString(1, submitter),
            Protobuf.EncodeString(2, activityHash),
            Protobuf.EncodeString(3, contentUri)
        );
    }

    /// <summary>
    /// Encodes MsgWithdrawNodeCommission.
    /// Fields: 1=operator_address
    /// </summary>
    public static byte[] EncodeMsgWithdrawRewards(string operatorAddress)
    {
        return Protobuf.EncodeString(1, operatorAddress);
    }

    /// <summary>
    /// Encodes MsgConfirmDelegation.
    /// Fields: 1=delegator_address, 2=validator_address
    /// </summary>
    public static byte[] EncodeMsgConfirmDelegation(string delegatorAddress, string validatorAddress)
    {
        return Protobuf.Concat(
            Protobuf.EncodeString(1, delegatorAddress),
            Protobuf.EncodeString(2, validatorAddress)
        );
    }

    /// <summary>
    /// Encodes MsgSend.
    /// Fields: 1=from_address, 2=to_address, 3=amount (repeated Coin)
    /// Coin: 1=denom, 2=amount
    /// </summary>
    public static byte[] EncodeMsgSend(string fromAddress, string toAddress, string amount, string denom)
    {
        var coin = Protobuf.Concat(
            Protobuf.EncodeString(1, denom),
            Protobuf.EncodeString(2, amount)
        );

        return Protobuf.Concat(
            Protobuf.EncodeString(1, fromAddress),
            Protobuf.EncodeString(2, toAddress),
            Protobuf.EncodeMessage(3, coin)
        );
    }
}

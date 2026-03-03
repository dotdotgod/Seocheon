namespace Seocheon.Sdk.Errors;

/// <summary>
/// SDK exception with numeric error code and structured message.
/// </summary>
public class SdkException : Exception
{
    /// <summary>Error code (SDK 9000-9012, x/node 1100-1112, x/activity 1200-1210).</summary>
    public uint Code { get; }

    public SdkException(uint code, string message) : base(message)
    {
        Code = code;
    }

    public SdkException(uint code, string message, Exception innerException)
        : base(message, innerException)
    {
        Code = code;
    }

    public override string ToString() =>
        Code > 0 ? $"[{Code}] {Message}" : Message;
}

/// <summary>
/// Predefined SDK error codes and factory methods.
/// </summary>
public static class SdkErrors
{
    // === SDK-level errors (9000-9012) ===

    public const uint CodeNotConnected = 9000;
    public const uint CodeBroadcastFailed = 9001;
    public const uint CodeTxTimeout = 9002;
    public const uint CodeTxNotFound = 9003;
    public const uint CodeSigningFailed = 9004;
    public const uint CodeInvalidConfig = 9005;
    public const uint CodeQueryFailed = 9006;
    public const uint CodeInvalidAddress = 9007;
    public const uint CodeInvalidHash = 9008;
    public const uint CodeInvalidContentUri = 9009;
    public const uint CodeSerializationFailed = 9010;
    public const uint CodeAccountNotFound = 9011;
    public const uint CodeInsufficientFunds = 9012;

    // === x/node ABCI errors (1100-1112) ===

    public const uint CodeNodeNotFound = 1101;
    public const uint CodeNodeAlreadyExists = 1102;
    public const uint CodeInvalidNodeStatus = 1103;
    public const uint CodeInvalidOperator = 1104;
    public const uint CodeInvalidAgentAddress = 1105;
    public const uint CodeInvalidDescription = 1106;
    public const uint CodeInvalidCommission = 1107;
    public const uint CodeUnauthorizedOperator = 1108;
    public const uint CodeUnauthorizedAgentMsg = 1109;
    public const uint CodeAgentAddressCooldown = 1110;
    public const uint CodeRegistrationPoolInsufficient = 1111;
    public const uint CodeInvalidDelegationConfirmation = 1112;

    // === x/activity ABCI errors (1200-1210) ===

    public const uint CodeSubmitterNotRegistered = 1200;
    public const uint CodeNodeNotEligible = 1201;
    public const uint CodeDuplicateActivityHash = 1202;
    public const uint CodeQuotaExceeded = 1203;
    public const uint CodeInvalidActivityHash = 1204;
    public const uint CodeInvalidContentUri_Chain = 1205;
    public const uint CodeInvalidEpoch = 1206;
    public const uint CodeInvalidWindow = 1207;
    public const uint CodeRewardDistributionFailed = 1208;
    public const uint CodeInvalidFeeModel = 1209;
    public const uint CodeFeegrantExpired = 1210;

    // === Factory methods ===

    public static SdkException NotConnected() =>
        new(CodeNotConnected, "SDK is not connected to chain");

    public static SdkException BroadcastFailed(string detail) =>
        new(CodeBroadcastFailed, $"Transaction broadcast failed: {detail}");

    public static SdkException TxTimeout(string txHash) =>
        new(CodeTxTimeout, $"Transaction confirmation timed out: {txHash}");

    public static SdkException TxNotFound(string txHash) =>
        new(CodeTxNotFound, $"Transaction not found: {txHash}");

    public static SdkException SigningFailed(string detail, Exception? inner = null) =>
        inner != null
            ? new(CodeSigningFailed, $"Signing failed: {detail}", inner)
            : new(CodeSigningFailed, $"Signing failed: {detail}");

    public static SdkException InvalidConfig(string detail) =>
        new(CodeInvalidConfig, $"Invalid configuration: {detail}");

    public static SdkException QueryFailed(string detail) =>
        new(CodeQueryFailed, $"Query failed: {detail}");

    public static SdkException InvalidAddress(string address) =>
        new(CodeInvalidAddress, $"Invalid address: {address}");

    public static SdkException InvalidHash(string detail) =>
        new(CodeInvalidHash, $"Invalid activity hash: {detail}");

    public static SdkException InvalidContentUri(string detail) =>
        new(CodeInvalidContentUri, $"Invalid content URI: {detail}");

    public static SdkException SerializationFailed(string detail, Exception? inner = null) =>
        inner != null
            ? new(CodeSerializationFailed, $"Serialization failed: {detail}", inner)
            : new(CodeSerializationFailed, $"Serialization failed: {detail}");

    public static SdkException AccountNotFound(string address) =>
        new(CodeAccountNotFound, $"Account not found: {address}");

    public static SdkException InsufficientFunds(string detail) =>
        new(CodeInsufficientFunds, $"Insufficient funds: {detail}");

    /// <summary>
    /// Maps an ABCI error code to an SdkException.
    /// </summary>
    public static SdkException AbciCodeToError(uint code, string rawLog = "")
    {
        var message = code switch
        {
            // x/node
            CodeNodeNotFound => "Node not found",
            CodeNodeAlreadyExists => "Node already exists",
            CodeInvalidNodeStatus => "Invalid node status",
            CodeInvalidOperator => "Invalid operator",
            CodeInvalidAgentAddress => "Invalid agent address",
            CodeInvalidDescription => "Invalid description",
            CodeInvalidCommission => "Invalid commission rate",
            CodeUnauthorizedOperator => "Unauthorized operator",
            CodeUnauthorizedAgentMsg => "Unauthorized agent message type",
            CodeAgentAddressCooldown => "Agent address change cooldown active",
            CodeRegistrationPoolInsufficient => "Registration pool insufficient",
            CodeInvalidDelegationConfirmation => "Invalid delegation confirmation",

            // x/activity
            CodeSubmitterNotRegistered => "Submitter not registered",
            CodeNodeNotEligible => "Node not eligible for activity",
            CodeDuplicateActivityHash => "Duplicate activity hash",
            CodeQuotaExceeded => "Activity quota exceeded",
            CodeInvalidActivityHash => "Invalid activity hash format",
            CodeInvalidContentUri_Chain => "Invalid content URI",
            CodeInvalidEpoch => "Invalid epoch",
            CodeInvalidWindow => "Invalid window",
            CodeRewardDistributionFailed => "Reward distribution failed",
            CodeInvalidFeeModel => "Invalid fee model",
            CodeFeegrantExpired => "Feegrant expired",

            _ => string.IsNullOrEmpty(rawLog) ? $"Unknown error (code {code})" : rawLog,
        };

        return new SdkException(code, message);
    }
}

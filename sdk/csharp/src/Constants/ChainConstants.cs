namespace Seocheon.Sdk.Constants;

/// <summary>
/// Seocheon blockchain constants. All values match on-chain parameters.
/// </summary>
public static class ChainConstants
{
    // === Epoch / Window ===

    /// <summary>Blocks per epoch (~24h at 5s/block).</summary>
    public const long EpochBlocks = 17_280;

    /// <summary>Windows per epoch.</summary>
    public const int WindowsPerEpoch = 12;

    /// <summary>Minimum active windows for reward qualification.</summary>
    public const int MinActiveWindows = 8;

    /// <summary>Blocks per window (EpochBlocks / WindowsPerEpoch).</summary>
    public const long WindowBlocks = EpochBlocks / WindowsPerEpoch; // 1440

    // === Quota ===

    /// <summary>Self-funded submissions per epoch.</summary>
    public const int SelfFundedQuota = 100;

    /// <summary>Feegrant submissions per epoch.</summary>
    public const int FeegrantQuota = 10;

    // === Fee Model ===

    /// <summary>Base activity fee in uppyeo (1 KKOT).</summary>
    public const long BaseActivityFee = 10_000_000_000;

    /// <summary>Fee exponent (0.5 basis points).</summary>
    public const int FeeExponent = 5000;

    /// <summary>Maximum activity fee in uppyeo (100 KKOT).</summary>
    public const long MaxActivityFee = 1_000_000_000_000;

    /// <summary>Whether feegrant submissions are fee-exempt.</summary>
    public const bool FeegrantFeeExempt = true;

    // === Dual Reward Pool ===

    /// <summary>Delegation minimum ratio (0.3 = 3000/10000).</summary>
    public const int DMin = 3000;

    /// <summary>Fee to activity pool ratio (0.8 = 80/20 split).</summary>
    public const int FeeToActivityPoolRatio = 8000;

    // === Token Denominations (6-stage system) ===

    /// <summary>Base denomination.</summary>
    public const string BaseDenom = "uppyeo";

    /// <summary>Display denomination.</summary>
    public const string DisplayDenom = "kkot";

    /// <summary>Uppyeo per sal (10^2).</summary>
    public const long UppyeoPerSal = 100;

    /// <summary>Uppyeo per pi (10^4).</summary>
    public const long UppyeoPerPi = 10_000;

    /// <summary>Uppyeo per sum (10^6).</summary>
    public const long UppyeoPerSum = 1_000_000;

    /// <summary>Uppyeo per hon (10^8).</summary>
    public const long UppyeoPerHon = 100_000_000;

    /// <summary>Uppyeo per kkot (10^10).</summary>
    public const long UppyeoPerKkot = 10_000_000_000;

    /// <summary>Denomination names in order.</summary>
    public static readonly string[] DenomNames = ["uppyeo", "sal", "pi", "sum", "hon", "kkot"];

    /// <summary>Denomination conversion factors (uppyeo units per denom).</summary>
    public static readonly long[] DenomFactors = [1, UppyeoPerSal, UppyeoPerPi, UppyeoPerSum, UppyeoPerHon, UppyeoPerKkot];

    // === Time Conversions ===

    /// <summary>Blocks per hour at 5s/block.</summary>
    public const long BlocksPerHour = 720;

    /// <summary>Blocks per day.</summary>
    public const long BlocksPerDay = 17_280;

    /// <summary>Blocks per year (~365.25 days).</summary>
    public const long BlocksPerYear = 6_307_200;

    /// <summary>Unbonding period in blocks (~21 days).</summary>
    public const long UnbondingPeriodBlocks = 362_880;

    /// <summary>Feegrant expiry in blocks (~180 days).</summary>
    public const long FeegrantExpiryBlocks = 3_110_400;

    // === Gas Defaults ===

    /// <summary>Gas limit for MsgSubmitActivity.</summary>
    public const ulong GasSubmitActivity = 200_000;

    /// <summary>Gas limit for MsgWithdrawRewards.</summary>
    public const ulong GasWithdraw = 300_000;

    /// <summary>Gas limit for MsgSend.</summary>
    public const ulong GasSend = 100_000;

    /// <summary>Fallback gas limit.</summary>
    public const ulong GasFallback = 200_000;

    /// <summary>Default gas price in uppyeo.</summary>
    public const ulong DefaultGasPrice = 250;

    /// <summary>Default gas price string.</summary>
    public const string DefaultGasPriceStr = "250uppyeo";

    /// <summary>Default gas adjustment multiplier.</summary>
    public const double DefaultGasAdjustment = 1.3;

    // === Broadcast & Confirm ===

    /// <summary>Default broadcast mode.</summary>
    public const string DefaultBroadcastMode = "sync";

    /// <summary>Default confirmation timeout in milliseconds.</summary>
    public const int DefaultConfirmTimeoutMs = 30_000;

    /// <summary>Default confirmation poll interval in milliseconds.</summary>
    public const int DefaultConfirmPollMs = 1_000;

    // === Agent Permissions ===

    /// <summary>Message types allowed for agent addresses.</summary>
    public static readonly string[] AgentAllowedMsgTypes =
    [
        "/seocheon.activity.v1.MsgSubmitActivity",
        "/cosmos.bank.v1beta1.MsgSend"
    ];

    /// <summary>Message types allowed for agent feegrant.</summary>
    public static readonly string[] AgentFeegrantAllowedMsgTypes =
    [
        "/seocheon.activity.v1.MsgSubmitActivity"
    ];

    /// <summary>Agent address change cooldown in blocks (1 epoch).</summary>
    public const long AgentAddressChangeCooldown = EpochBlocks;

    // === BIP44 / Address ===

    /// <summary>BIP44 derivation path for Cosmos (coin type 118).</summary>
    public const string Bip44Path = "m/44'/118'/0'/0/0";

    /// <summary>Bech32 address prefix.</summary>
    public const string AddressPrefix = "seocheon";

    /// <summary>Expected SHA-256 hash length in hex characters.</summary>
    public const int ActivityHashLength = 64;
}

import Foundation

/// Chain parameter constants for the Seocheon blockchain.
public enum ChainConstants {
    // MARK: - Epoch/Window parameters

    /// Blocks per epoch (~24h at 5s/block)
    public static let epochLength: Int64 = 17280
    /// Windows per epoch
    public static let windowsPerEpoch: Int64 = 12
    /// Minimum active windows for reward qualification
    public static let minActiveWindows: Int64 = 8
    /// Blocks per window (epochLength / windowsPerEpoch)
    public static let windowLength: Int64 = 1440

    // MARK: - Quota parameters

    /// Epoch quota for self-funded nodes
    public static let selfFundedQuota: UInt64 = 100
    /// Epoch quota for feegrant nodes
    public static let feegrantQuota: UInt64 = 10

    // MARK: - Pruning parameters

    /// Activity pruning keep blocks (~1 year)
    public static let activityPruningKeepBlocks: Int64 = 6307200

    // MARK: - Fee model parameters

    public static let feeThresholdMultiplier: Int64 = 3
    /// 1 KKOT in uppyeo
    public static let baseActivityFee: Int64 = 10_000_000_000
    /// Basis points (0.5)
    public static let feeExponent: Int64 = 5000
    /// 100 KKOT in uppyeo
    public static let maxActivityFee: Int64 = 1_000_000_000_000
    public static let minFeegrantQuota: UInt64 = 8
    /// Basis points (0.5)
    public static let quotaReductionRate: Int64 = 5000
    public static let feegrantFeeExempt: Bool = true

    // MARK: - Dual reward pool parameters

    /// D_min basis points (0.3)
    public static let dMin: Int64 = 3000
    /// Fee to activity pool ratio basis points (0.8)
    public static let feeToActivityPoolRatio: Int64 = 8000

    // MARK: - Node registration parameters

    public static let maxRegistrationsPerBlock: Int64 = 5
    public static let registrationCooldownBlks: Int64 = 100
    public static let registrationDeposit: String = "0"
    public static let maxTags: Int64 = 10
    public static let maxTagLength: Int64 = 32

    // MARK: - Agent permission parameters

    public static let agentAllowedMsgTypes: [String] = [
        "/seocheon.activity.v1.MsgSubmitActivity",
        "/cosmos.bank.v1beta1.MsgSend",
    ]
    public static let agentFeegrantAllowedMsgTypes: [String] = [
        "/seocheon.activity.v1.MsgSubmitActivity",
    ]
    public static let agentAddressChangeCooldown: Int64 = 17280

    // MARK: - Time-block conversion (5s/block)

    public static let blocksPerHour: Int64 = 720
    public static let blocksPerDay: Int64 = 17280
    public static let blocksPerYear: Int64 = 6307200
    /// ~21 days
    public static let unbondingPeriodBlocks: Int64 = 362880
    /// ~180 days
    public static let feegrantExpiryBlocks: Int64 = 3110400

    // MARK: - Token denomination constants (6-stage system)

    /// Base denomination (10^0)
    public static let tokenBaseDenom: String = "uppyeo"
    /// sal denomination (10^2)
    public static let tokenSalDenom: String = "sal"
    /// pi denomination (10^4)
    public static let tokenPiDenom: String = "pi"
    /// sum denomination (10^6)
    public static let tokenSumDenom: String = "sum"
    /// hon denomination (10^8)
    public static let tokenHonDenom: String = "hon"
    /// Display denomination (10^10)
    public static let tokenDisplayDenom: String = "kkot"

    // MARK: - Denomination conversion factors

    public static let uppyeoPerSal: Int64 = 100
    public static let uppyeoPerPi: Int64 = 10_000
    public static let uppyeoPerSum: Int64 = 1_000_000
    public static let uppyeoPerHon: Int64 = 100_000_000
    public static let uppyeoPerKkot: Int64 = 10_000_000_000

    // MARK: - Default SDK configuration

    public static let defaultGasPrice: String = "250uppyeo"
    public static let defaultGasAdjustment: Double = 1.3
    public static let defaultBroadcastMode: String = "sync"
    public static let defaultConfirmTimeoutMs: UInt64 = 30000
    public static let defaultConfirmPollMs: UInt64 = 1000
}

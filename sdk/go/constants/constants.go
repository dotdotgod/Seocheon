// Package constants defines chain parameters, token denominations, and error codes
// for the Seocheon blockchain SDK.
package constants

// Epoch/Window parameters (x/activity params)
const (
	EpochLength       int64 = 17280 // blocks per epoch (~24h at 5s/block)
	WindowsPerEpoch   int64 = 12    // windows per epoch
	MinActiveWindows  int64 = 8     // minimum active windows for reward qualification
	WindowLength      int64 = 1440  // blocks per window (EpochLength / WindowsPerEpoch)
)

// Quota parameters (x/activity params)
const (
	SelfFundedQuota uint64 = 100 // epoch quota for self-funded nodes
	FeegrantQuota   uint64 = 10  // epoch quota for feegrant nodes
)

// Pruning parameters
const (
	ActivityPruningKeepBlocks int64 = 6307200 // ~1 year
)

// Fee model parameters (x/activity params)
const (
	FeeThresholdMultiplier int64  = 3
	BaseActivityFee        int64  = 10000000000   // 1 KKOT in uppyeo
	FeeExponent            int64  = 5000           // basis points (0.5)
	MaxActivityFee         int64  = 1000000000000  // 100 KKOT in uppyeo
	MinFeegrantQuota       uint64 = 8
	QuotaReductionRate     int64  = 5000 // basis points (0.5)
	FeegrantFeeExempt      bool   = true
)

// Dual reward pool parameters (x/activity params)
const (
	DMin                   int64 = 3000 // basis points (0.3)
	FeeToActivityPoolRatio int64 = 8000 // basis points (0.8)
)

// Node registration parameters (x/node params)
const (
	MaxRegistrationsPerBlock int64  = 5
	RegistrationCooldownBlks int64  = 100
	RegistrationDeposit      string = "0" // uppyeo
	MaxTags                  int64  = 10
	MaxTagLength             int64  = 32
)

// Agent permission parameters
var (
	AgentAllowedMsgTypes = []string{
		"/seocheon.activity.v1.MsgSubmitActivity",
		"/cosmos.bank.v1beta1.MsgSend",
	}
	AgentFeegrantAllowedMsgTypes = []string{
		"/seocheon.activity.v1.MsgSubmitActivity",
	}
)

const (
	AgentAddressChangeCooldown int64 = 17280 // 1 epoch
)

// Time-block conversion (5s/block)
const (
	BlocksPerHour         int64 = 720
	BlocksPerDay          int64 = 17280
	BlocksPerYear         int64 = 6307200
	UnbondingPeriodBlocks int64 = 362880  // ~21 days
	FeegrantExpiryBlocks  int64 = 3110400 // ~180 days
)

// Token denomination constants (6-stage system)
// uppyeo(0) → sal(2) → pi(4) → sum(6) → hon(8) → kkot(10)
const (
	TokenBaseDenom    = "uppyeo" // base denomination (10^0)
	TokenSalDenom     = "sal"    // sal denomination (10^2)
	TokenPiDenom      = "pi"     // pi denomination (10^4)
	TokenSumDenom     = "sum"    // sum denomination (10^6)
	TokenHonDenom     = "hon"    // hon denomination (10^8)
	TokenDisplayDenom = "kkot"   // display denomination (10^10)
)

// Denomination conversion factors (base unit: uppyeo)
const (
	UppyeoPerSal  int64 = 100
	UppyeoPerPi   int64 = 10000
	UppyeoPerSum  int64 = 1000000
	UppyeoPerHon  int64 = 100000000
	UppyeoPerKkot int64 = 10000000000
)

// Default SDK configuration values
const (
	DefaultGasPrice         = "250uppyeo"
	DefaultGasAdjustment    = 1.3
	DefaultBroadcastMode    = "sync"
	DefaultConfirmTimeoutMs = 30000
	DefaultConfirmPollMs    = 1000
)

package types

import "fmt"

// Default parameter values.
var (
	DefaultEpochLength              = int64(17280)   // ~1 day in blocks
	DefaultWindowsPerEpoch          = int64(12)      // 12 windows per epoch
	DefaultMinActiveWindows         = int64(8)       // 8 out of 12 windows required
	DefaultSelfFundedQuota          = uint64(100)    // max 100 activities per epoch (self-funded)
	DefaultFeegrantQuota            = uint64(10)     // max 10 activities per epoch (feegrant)
	DefaultActivityPruningKeepBlocks = int64(6307200) // ~365 days (1 year) in blocks

	// Activity Cost Model defaults.
	DefaultFeeThresholdMultiplier = uint64(3)           // fee activates when N_a > N_d * 3
	DefaultBaseActivityFee       = uint64(1_000_000)    // 1 KKOT in usum
	DefaultFeeExponent           = uint64(5000)         // 0.5 in basis points (sqrt curve)
	DefaultMaxActivityFee        = uint64(100_000_000)  // 100 KKOT in usum
	DefaultMinFeegrantQuota      = uint64(8)            // minimum feegrant quota (matches min_active_windows)
	DefaultQuotaReductionRate    = uint64(5000)         // 0.5 in basis points
	DefaultFeegrantFeeExempt     = true                 // feegrant nodes exempt from activity fees

	// Dual Reward Pool defaults.
	DefaultDMin                    = uint64(3000) // 0.3 in basis points — delegation pool minimum 30%
	DefaultFeeToActivityPoolRatio  = uint64(8000) // 80% of collected activity fees → activity reward pool
)

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		EpochLength:               DefaultEpochLength,
		WindowsPerEpoch:           DefaultWindowsPerEpoch,
		MinActiveWindows:          DefaultMinActiveWindows,
		SelfFundedQuota:           DefaultSelfFundedQuota,
		FeegrantQuota:             DefaultFeegrantQuota,
		ActivityPruningKeepBlocks: DefaultActivityPruningKeepBlocks,
		FeeThresholdMultiplier:    DefaultFeeThresholdMultiplier,
		BaseActivityFee:           DefaultBaseActivityFee,
		FeeExponent:               DefaultFeeExponent,
		MaxActivityFee:            DefaultMaxActivityFee,
		MinFeegrantQuota:          DefaultMinFeegrantQuota,
		QuotaReductionRate:        DefaultQuotaReductionRate,
		FeegrantFeeExempt:         DefaultFeegrantFeeExempt,
		DMin:                      DefaultDMin,
		FeeToActivityPoolRatio:    DefaultFeeToActivityPoolRatio,
	}
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.EpochLength <= 0 {
		return fmt.Errorf("epoch_length must be positive, got %d", p.EpochLength)
	}
	if p.WindowsPerEpoch <= 0 {
		return fmt.Errorf("windows_per_epoch must be positive, got %d", p.WindowsPerEpoch)
	}
	if p.MinActiveWindows <= 0 {
		return fmt.Errorf("min_active_windows must be positive, got %d", p.MinActiveWindows)
	}
	if p.MinActiveWindows > p.WindowsPerEpoch {
		return fmt.Errorf("min_active_windows (%d) cannot exceed windows_per_epoch (%d)", p.MinActiveWindows, p.WindowsPerEpoch)
	}
	if p.EpochLength%p.WindowsPerEpoch != 0 {
		return fmt.Errorf("epoch_length (%d) must be evenly divisible by windows_per_epoch (%d)", p.EpochLength, p.WindowsPerEpoch)
	}
	if p.SelfFundedQuota == 0 {
		return fmt.Errorf("self_funded_quota must be positive")
	}
	if p.FeegrantQuota == 0 {
		return fmt.Errorf("feegrant_quota must be positive")
	}
	if p.ActivityPruningKeepBlocks <= 0 {
		return fmt.Errorf("activity_pruning_keep_blocks must be positive, got %d", p.ActivityPruningKeepBlocks)
	}
	if p.FeeThresholdMultiplier == 0 {
		return fmt.Errorf("fee_threshold_multiplier must be positive")
	}
	if p.FeeExponent > 10000 {
		return fmt.Errorf("fee_exponent must be <= 10000 basis points, got %d", p.FeeExponent)
	}
	if p.MaxActivityFee > 0 && p.MaxActivityFee < p.BaseActivityFee {
		return fmt.Errorf("max_activity_fee (%d) must be >= base_activity_fee (%d)", p.MaxActivityFee, p.BaseActivityFee)
	}
	if p.MinFeegrantQuota > p.FeegrantQuota {
		return fmt.Errorf("min_feegrant_quota (%d) cannot exceed feegrant_quota (%d)", p.MinFeegrantQuota, p.FeegrantQuota)
	}
	if p.QuotaReductionRate > 10000 {
		return fmt.Errorf("quota_reduction_rate must be <= 10000 basis points, got %d", p.QuotaReductionRate)
	}
	if p.DMin > 10000 {
		return fmt.Errorf("d_min must be <= 10000 basis points, got %d", p.DMin)
	}
	if p.FeeToActivityPoolRatio > 10000 {
		return fmt.Errorf("fee_to_activity_pool_ratio must be <= 10000 basis points, got %d", p.FeeToActivityPoolRatio)
	}
	return nil
}

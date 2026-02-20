package types

import "fmt"

// Default parameter values.
var (
	DefaultEpochLength              = int64(17280)   // ~1 day in blocks
	DefaultWindowsPerEpoch          = int64(12)      // 12 windows per epoch
	DefaultMinActiveWindows         = int64(8)       // 8 out of 12 windows required
	DefaultSelfFundedQuota          = uint64(100)    // max 100 activities per epoch (self-funded)
	DefaultFeegrantQuota            = uint64(10)     // max 10 activities per epoch (feegrant)
	DefaultActivityPruningKeepBlocks = int64(1555200) // ~90 days in blocks
)

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		EpochLength:              DefaultEpochLength,
		WindowsPerEpoch:          DefaultWindowsPerEpoch,
		MinActiveWindows:         DefaultMinActiveWindows,
		SelfFundedQuota:          DefaultSelfFundedQuota,
		FeegrantQuota:            DefaultFeegrantQuota,
		ActivityPruningKeepBlocks: DefaultActivityPruningKeepBlocks,
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
	return nil
}

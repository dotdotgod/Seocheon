package testutil

import (
	activitytypes "seocheon/x/activity/types"
)

// FastTestnetActivityParams returns x/activity params optimized for fast testnet
// iteration. EpochLength is reduced to 120 blocks (~2 minutes at 1s/block)
// while preserving the same window/threshold ratios as production.
func FastTestnetActivityParams() activitytypes.Params {
	p := activitytypes.DefaultParams()

	// Production: 17,280 blocks/epoch (~24h at 5s/block)
	// Fast testnet: 120 blocks/epoch (~2min at 1s/block)
	p.EpochLength = 120
	p.WindowsPerEpoch = 12       // same ratio as production
	p.MinActiveWindows = 8       // same threshold as production
	p.SelfFundedQuota = 200      // relaxed for burst testing

	return p
}

// MediumTestnetActivityParams returns moderate-speed params for integration testing.
// EpochLength is 1,440 blocks (~24 minutes at 1s/block).
func MediumTestnetActivityParams() activitytypes.Params {
	p := activitytypes.DefaultParams()

	p.EpochLength = 1440
	p.WindowsPerEpoch = 12
	p.MinActiveWindows = 8

	return p
}

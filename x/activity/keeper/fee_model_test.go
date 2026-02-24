package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"seocheon/x/activity/types"
)

func TestCalculateActivityFee_Bootstrap(t *testing.T) {
	params := types.DefaultParams()

	// N_a = 450, N_d = 150, threshold = 150*3 = 450 → S = 1.0 → no fee
	fee := calculateActivityFee(450, 150, params)
	require.Equal(t, uint64(0), fee)

	// N_a = 100, N_d = 150, threshold = 450 → S < 1.0 → no fee
	fee = calculateActivityFee(100, 150, params)
	require.Equal(t, uint64(0), fee)

	// N_a = 0 → no fee
	fee = calculateActivityFee(0, 150, params)
	require.Equal(t, uint64(0), fee)
}

func TestCalculateActivityFee_Growth(t *testing.T) {
	params := types.DefaultParams()

	// N_a = 900, N_d = 150, threshold = 450
	// S = 900/450 = 2.0, S-1 = 1.0
	// fee = 1_000_000 * (1.0)^0.5 = 1_000_000
	fee := calculateActivityFee(900, 150, params)
	require.Equal(t, uint64(1_000_000), fee)

	// N_a = 2250, N_d = 150, threshold = 450
	// S = 2250/450 = 5.0, S-1 = 4.0
	// fee = 1_000_000 * (4.0)^0.5 = 1_000_000 * 2 = 2_000_000
	fee = calculateActivityFee(2250, 150, params)
	require.Equal(t, uint64(2_000_000), fee)
}

func TestCalculateActivityFee_MaxCap(t *testing.T) {
	params := types.DefaultParams()

	// Very high N_a should be capped at max_activity_fee
	fee := calculateActivityFee(100000, 150, params)
	require.LessOrEqual(t, fee, params.MaxActivityFee)
}

func TestCalculateActivityFee_ZeroBaseFee(t *testing.T) {
	params := types.DefaultParams()
	params.BaseActivityFee = 0

	fee := calculateActivityFee(1000, 150, params)
	require.Equal(t, uint64(0), fee)
}

func TestCalculateActivityFee_ZeroMultiplier(t *testing.T) {
	params := types.DefaultParams()
	params.FeeThresholdMultiplier = 0

	fee := calculateActivityFee(1000, 150, params)
	require.Equal(t, uint64(0), fee)
}

func TestCalculateEffectiveQuota_Bootstrap(t *testing.T) {
	params := types.DefaultParams()

	// N_a = 100, N_d = 150, threshold = 450 → S < 1.0 → no reduction
	quota := calculateEffectiveQuota(100, 150, params)
	require.Equal(t, params.FeegrantQuota, quota)

	// N_a = 450 → S = 1.0 → no reduction
	quota = calculateEffectiveQuota(450, 150, params)
	require.Equal(t, params.FeegrantQuota, quota)
}

func TestCalculateEffectiveQuota_Growth(t *testing.T) {
	params := types.DefaultParams()

	// N_a = 900, N_d = 150, threshold = 450
	// S-1 = 1.0, reduction = 10 * 0.5 * 1.0 = 5
	// effective = 10 - 5 = 5 (but min is 8)
	quota := calculateEffectiveQuota(900, 150, params)
	require.Equal(t, params.MinFeegrantQuota, quota)
}

func TestCalculateEffectiveQuota_HighSaturation(t *testing.T) {
	params := types.DefaultParams()

	// Very high saturation → quota should be min_feegrant_quota
	quota := calculateEffectiveQuota(10000, 150, params)
	require.Equal(t, params.MinFeegrantQuota, quota)
}

func TestCalculateEffectiveQuota_ZeroMultiplier(t *testing.T) {
	params := types.DefaultParams()
	params.FeeThresholdMultiplier = 0

	quota := calculateEffectiveQuota(1000, 150, params)
	require.Equal(t, params.FeegrantQuota, quota)
}

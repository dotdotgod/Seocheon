package keeper

import (
	"testing"

	sdkmath "cosmossdk.io/math"
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

func TestDecPow(t *testing.T) {
	tests := []struct {
		name     string
		base     sdkmath.LegacyDec
		expBP    uint64
		expected sdkmath.LegacyDec
	}{
		{
			name:     "exponent 0 → 1",
			base:     sdkmath.LegacyNewDec(4),
			expBP:    0,
			expected: sdkmath.LegacyOneDec(),
		},
		{
			name:     "exponent 10000 → base",
			base:     sdkmath.LegacyNewDec(4),
			expBP:    10000,
			expected: sdkmath.LegacyNewDec(4),
		},
		{
			name:     "4^0.5 = 2",
			base:     sdkmath.LegacyNewDec(4),
			expBP:    5000,
			expected: sdkmath.LegacyNewDec(2),
		},
		{
			name:     "1^0.5 = 1",
			base:     sdkmath.LegacyOneDec(),
			expBP:    5000,
			expected: sdkmath.LegacyOneDec(),
		},
		{
			name:     "9^0.5 = 3",
			base:     sdkmath.LegacyNewDec(9),
			expBP:    5000,
			expected: sdkmath.LegacyNewDec(3),
		},
		{
			name:     "8^(10000/30000) = 8^(1/3) = 2",
			base:     sdkmath.LegacyNewDec(8),
			expBP:    10000 / 3, // 3333 bp → reduced with GCD
			expected: sdkmath.LegacyNewDec(2),
		},
		{
			name:     "4^1.5 = 8",
			base:     sdkmath.LegacyNewDec(4),
			expBP:    15000,
			expected: sdkmath.LegacyNewDec(8),
		},
		{
			name:     "4^2.0 = 16",
			base:     sdkmath.LegacyNewDec(4),
			expBP:    20000,
			expected: sdkmath.LegacyNewDec(16),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := decPow(tc.base, tc.expBP)
			require.NoError(t, err)
			// Allow precision error from ApproxRoot (Newton's method, 100 iterations).
			// High-order roots (e.g., 10000th root) have lower precision than sqrt.
			diff := result.Sub(tc.expected).Abs()
			require.True(t, diff.LT(sdkmath.LegacyNewDecWithPrec(1, 3)),
				"expected %s, got %s (diff %s)", tc.expected, result, diff)
		})
	}
}

func TestGcd(t *testing.T) {
	require.Equal(t, uint64(5000), gcd(5000, 10000))
	require.Equal(t, uint64(2500), gcd(7500, 10000))
	require.Equal(t, uint64(10000), gcd(10000, 10000))
	require.Equal(t, uint64(5000), gcd(15000, 10000))
	require.Equal(t, uint64(10000), gcd(20000, 10000))
	require.Equal(t, uint64(1), gcd(1, 10000))
}

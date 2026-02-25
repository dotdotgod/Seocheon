package keeper

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestCalculateActivityRatio_NoActivityNodes(t *testing.T) {
	// N_a = 0 → all to DPoS, activity_ratio = 0
	ratio := CalculateActivityRatio(0, 150, 3000)
	require.True(t, ratio.IsZero())
}

func TestCalculateActivityRatio_EqualNodes(t *testing.T) {
	// N_a = 150, N_d = 150, D_min = 0.3
	// natural = 150/300 = 0.5 > D_min(0.3)
	// activity_ratio = 1 - 0.5 = 0.5
	ratio := CalculateActivityRatio(150, 150, 3000)
	expected := math.LegacyNewDecWithPrec(5, 1) // 0.5
	require.True(t, ratio.Equal(expected), "expected 0.5, got %s", ratio)
}

func TestCalculateActivityRatio_ManyActivityNodes(t *testing.T) {
	// N_a = 850, N_d = 150, D_min = 0.3
	// natural = 150/1000 = 0.15 < D_min(0.3)
	// delegation_ratio = D_min = 0.3
	// activity_ratio = 1 - 0.3 = 0.7
	ratio := CalculateActivityRatio(850, 150, 3000)
	expected := math.LegacyNewDecWithPrec(7, 1) // 0.7
	require.True(t, ratio.Equal(expected), "expected 0.7, got %s", ratio)
}

func TestCalculateActivityRatio_FewActivityNodes(t *testing.T) {
	// N_a = 10, N_d = 150, D_min = 0.3
	// natural = 150/160 = 0.9375 > D_min
	// activity_ratio = 1 - 0.9375 = 0.0625
	ratio := CalculateActivityRatio(10, 150, 3000)
	expected := math.LegacyOneDec().Sub(math.LegacyNewDec(150).Quo(math.LegacyNewDec(160)))
	require.True(t, ratio.Equal(expected), "expected %s, got %s", expected, ratio)
}

func TestCalculateActivityRatio_DMinFloor(t *testing.T) {
	// N_a = 10000, N_d = 1 → natural = 1/10001 ≈ 0.0001 < D_min(0.3)
	// delegation_ratio = D_min = 0.3
	// activity_ratio = 0.7
	ratio := CalculateActivityRatio(10000, 1, 3000)
	expected := math.LegacyNewDecWithPrec(7, 1) // 0.7
	require.True(t, ratio.Equal(expected), "expected 0.7, got %s", ratio)
}

func TestCalculateActivityRatio_HighDMin(t *testing.T) {
	// D_min = 9000 (0.9) → activity_ratio = max 0.1
	// N_a = 1000, N_d = 100 → natural = 100/1100 = 0.0909 < 0.9
	// delegation_ratio = 0.9 → activity_ratio = 0.1
	ratio := CalculateActivityRatio(1000, 100, 9000)
	expected := math.LegacyNewDecWithPrec(1, 1) // 0.1
	require.True(t, ratio.Equal(expected), "expected 0.1, got %s", ratio)
}

func TestCalculateActivityRatio_DMinZero(t *testing.T) {
	// D_min = 0 → no floor, purely natural
	// N_a = 900, N_d = 100 → natural = 100/1000 = 0.1
	// activity_ratio = 0.9
	ratio := CalculateActivityRatio(900, 100, 0)
	expected := math.LegacyNewDecWithPrec(9, 1) // 0.9
	require.True(t, ratio.Equal(expected), "expected 0.9, got %s", ratio)
}

func TestCalculateActivityRatio_DMinFull(t *testing.T) {
	// D_min = 10000 (1.0) → all to DPoS, activity_ratio = 0
	ratio := CalculateActivityRatio(100, 100, 10000)
	require.True(t, ratio.IsZero(), "expected 0, got %s", ratio)
}

func TestCalculateActivityRatio_SingleValidator(t *testing.T) {
	// N_d = 1 is typical for testnet / early genesis
	// N_a = 5, N_d = 1, D_min = 3000 (0.3)
	// natural = 1/6 = 0.1667 < 0.3
	// delegation_ratio = 0.3 → activity_ratio = 0.7
	ratio := CalculateActivityRatio(5, 1, 3000)
	expected := math.LegacyNewDecWithPrec(7, 1)
	require.True(t, ratio.Equal(expected), "expected 0.7, got %s", ratio)
}

func TestCalculateActivityRatio_NaturalExactlyDMin(t *testing.T) {
	// Edge case: natural ratio equals D_min exactly.
	// N_d / (N_a + N_d) = 0.3 → N_d = 0.3*(N_a+N_d) → 0.7*N_d = 0.3*N_a → N_a = 7/3*N_d
	// Use N_d = 3, N_a = 7 → natural = 3/10 = 0.3 = D_min
	// delegation_ratio = 0.3, activity_ratio = 0.7
	ratio := CalculateActivityRatio(7, 3, 3000)
	expected := math.LegacyNewDecWithPrec(7, 1) // 0.7
	require.True(t, ratio.Equal(expected), "expected 0.7, got %s", ratio)
}

package keeper_test

import (
	"testing"

	"seocheon/x/activity/keeper"
	"seocheon/x/activity/types"
)

func TestGetCurrentEpoch(t *testing.T) {
	params := types.DefaultParams() // EpochLength=17280

	tests := []struct {
		name        string
		blockHeight int64
		expected    int64
	}{
		{"block 0", 0, 0},
		{"block 1 (first block)", 1, 0},
		{"block 17280 (last block of epoch 0)", 17280, 0},
		{"block 17281 (first block of epoch 1)", 17281, 1},
		{"block 34560 (last block of epoch 1)", 34560, 1},
		{"block 34561 (first block of epoch 2)", 34561, 2},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := keeper.GetCurrentEpoch(tc.blockHeight, params)
			if got != tc.expected {
				t.Errorf("GetCurrentEpoch(%d) = %d, want %d", tc.blockHeight, got, tc.expected)
			}
		})
	}
}

func TestGetCurrentWindow(t *testing.T) {
	params := types.DefaultParams() // EpochLength=17280, WindowsPerEpoch=12, WindowLength=1440

	tests := []struct {
		name        string
		blockHeight int64
		expected    int64
	}{
		{"block 0", 0, 0},
		{"block 1 (first block, window 0)", 1, 0},
		{"block 1440 (last block of window 0)", 1440, 0},
		{"block 1441 (first block of window 1)", 1441, 1},
		{"block 2880 (last block of window 1)", 2880, 1},
		{"block 2881 (first block of window 2)", 2881, 2},
		{"block 17280 (last block of epoch 0, window 11)", 17280, 11},
		{"block 17281 (first block of epoch 1, window 0)", 17281, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := keeper.GetCurrentWindow(tc.blockHeight, params)
			if got != tc.expected {
				t.Errorf("GetCurrentWindow(%d) = %d, want %d", tc.blockHeight, got, tc.expected)
			}
		})
	}
}

func TestGetWindowLength(t *testing.T) {
	params := types.DefaultParams()
	got := keeper.GetWindowLength(params)
	if got != 1440 {
		t.Errorf("GetWindowLength() = %d, want 1440", got)
	}
}

func TestIsEpochBoundary(t *testing.T) {
	params := types.DefaultParams()

	tests := []struct {
		blockHeight int64
		expected    bool
	}{
		{0, false},
		{1, false},
		{17279, false},
		{17280, true},
		{17281, false},
		{34560, true},
	}

	for _, tc := range tests {
		got := keeper.IsEpochBoundary(tc.blockHeight, params)
		if got != tc.expected {
			t.Errorf("IsEpochBoundary(%d) = %v, want %v", tc.blockHeight, got, tc.expected)
		}
	}
}

func TestIsWindowBoundary(t *testing.T) {
	params := types.DefaultParams()

	tests := []struct {
		blockHeight int64
		expected    bool
	}{
		{0, false},
		{1, false},
		{1440, true},
		{1441, false},
		{2880, true},
		{17280, true}, // also epoch boundary
	}

	for _, tc := range tests {
		got := keeper.IsWindowBoundary(tc.blockHeight, params)
		if got != tc.expected {
			t.Errorf("IsWindowBoundary(%d) = %v, want %v", tc.blockHeight, got, tc.expected)
		}
	}
}

func TestGetEpochStartBlock(t *testing.T) {
	params := types.DefaultParams()

	tests := []struct {
		epoch    int64
		expected int64
	}{
		{0, 1},
		{1, 17281},
		{2, 34561},
	}

	for _, tc := range tests {
		got := keeper.GetEpochStartBlock(tc.epoch, params)
		if got != tc.expected {
			t.Errorf("GetEpochStartBlock(%d) = %d, want %d", tc.epoch, got, tc.expected)
		}
	}
}

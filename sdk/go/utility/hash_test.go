package utility

import (
	"testing"

	"github.com/seocheon/sdk-go/constants"
)

func TestVerifyActivityHash_Valid(t *testing.T) {
	// Valid SHA-256 hex string (64 chars)
	hash := "a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890"
	if !VerifyActivityHash(hash) {
		t.Errorf("expected valid hash, got invalid")
	}
}

func TestVerifyActivityHash_ComputedHash(t *testing.T) {
	hash := ComputeActivityHash([]byte("hello world"))
	if !VerifyActivityHash(hash) {
		t.Errorf("computed hash should be valid: %s", hash)
	}
	if len(hash) != 64 {
		t.Errorf("expected 64 chars, got %d", len(hash))
	}
}

func TestVerifyActivityHash_Invalid(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"too short", "a1b2c3"},
		{"too long", "a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f6789000"},
		{"non-hex chars", "g1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890a1b2c3d4e5f67890"},
		{"spaces", "a1b2c3d4 e5f67890 a1b2c3d4 e5f67890 a1b2c3d4 e5f67890 a1b2c3d4 e5f6789"},
		{"uppercase valid", "A1B2C3D4E5F67890A1B2C3D4E5F67890A1B2C3D4E5F67890A1B2C3D4E5F67890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyActivityHash(tt.hash)
			// Only "uppercase valid" should pass (hex is case-insensitive)
			if tt.name == "uppercase valid" {
				if !result {
					t.Errorf("uppercase hex should be valid")
				}
				return
			}
			if result {
				t.Errorf("expected invalid hash for %q", tt.name)
			}
		})
	}
}

func TestComputeActivityHash_Deterministic(t *testing.T) {
	data := []byte("test activity data")
	hash1 := ComputeActivityHash(data)
	hash2 := ComputeActivityHash(data)
	if hash1 != hash2 {
		t.Errorf("hash should be deterministic: %s != %s", hash1, hash2)
	}
}

func TestComputeActivityHash_DifferentInputs(t *testing.T) {
	hash1 := ComputeActivityHash([]byte("input 1"))
	hash2 := ComputeActivityHash([]byte("input 2"))
	if hash1 == hash2 {
		t.Errorf("different inputs should produce different hashes")
	}
}

func TestComputeEpoch(t *testing.T) {
	tests := []struct {
		blockHeight int64
		epochLength int64
		expected    int64
	}{
		{1, constants.EpochLength, 0},                      // first block = epoch 0
		{constants.EpochLength, constants.EpochLength, 0},  // last block of epoch 0
		{constants.EpochLength + 1, constants.EpochLength, 1}, // first block of epoch 1
		{17280, 17280, 0},
		{17281, 17280, 1},
		{34560, 17280, 1},
		{34561, 17280, 2},
		{0, 17280, 0},  // edge: zero height
		{1, 0, 0},      // edge: zero epoch length
	}

	for _, tt := range tests {
		result := ComputeEpoch(tt.blockHeight, tt.epochLength)
		if result != tt.expected {
			t.Errorf("ComputeEpoch(%d, %d) = %d, want %d",
				tt.blockHeight, tt.epochLength, result, tt.expected)
		}
	}
}

func TestComputeWindow(t *testing.T) {
	tests := []struct {
		blockHeight     int64
		epochLength     int64
		windowsPerEpoch int64
		expected        int64
	}{
		{1, 17280, 12, 0},      // first block = window 0
		{1440, 17280, 12, 0},   // last block of window 0
		{1441, 17280, 12, 1},   // first block of window 1
		{2880, 17280, 12, 1},   // last block of window 1
		{2881, 17280, 12, 2},   // first block of window 2
		{17280, 17280, 12, 11}, // last block of epoch 0 = window 11
		{17281, 17280, 12, 0},  // first block of epoch 1 = window 0
		{0, 17280, 12, 0},      // edge: zero height
		{1, 0, 12, 0},          // edge: zero epoch length
		{1, 17280, 0, 0},       // edge: zero windows
	}

	for _, tt := range tests {
		result := ComputeWindow(tt.blockHeight, tt.epochLength, tt.windowsPerEpoch)
		if result != tt.expected {
			t.Errorf("ComputeWindow(%d, %d, %d) = %d, want %d",
				tt.blockHeight, tt.epochLength, tt.windowsPerEpoch, result, tt.expected)
		}
	}
}

func TestEpochStartBlock(t *testing.T) {
	if EpochStartBlock(0, 17280) != 1 {
		t.Error("epoch 0 start should be 1")
	}
	if EpochStartBlock(1, 17280) != 17281 {
		t.Error("epoch 1 start should be 17281")
	}
}

func TestEpochEndBlock(t *testing.T) {
	if EpochEndBlock(0, 17280) != 17280 {
		t.Error("epoch 0 end should be 17280")
	}
	if EpochEndBlock(1, 17280) != 34560 {
		t.Error("epoch 1 end should be 34560")
	}
}

func TestWindowStartBlock(t *testing.T) {
	epochStart := int64(1)
	if WindowStartBlock(epochStart, 0, 1440) != 1 {
		t.Error("window 0 start should be 1")
	}
	if WindowStartBlock(epochStart, 1, 1440) != 1441 {
		t.Error("window 1 start should be 1441")
	}
}

func TestWindowEndBlock(t *testing.T) {
	if WindowEndBlock(1, 1440) != 1440 {
		t.Error("window 0 end should be 1440")
	}
	if WindowEndBlock(1441, 1440) != 2880 {
		t.Error("window 1 end should be 2880")
	}
}

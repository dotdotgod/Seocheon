package utility

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// Additional hash/compute tests for spec alignment

func TestVerifyActivityHash_ExactLength(t *testing.T) {
	// Exactly 64 hex chars (32 bytes SHA-256)
	hash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if !VerifyActivityHash(hash) {
		t.Error("valid SHA-256 hash should pass")
	}
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}
}

func TestComputeActivityHash_SHA256(t *testing.T) {
	data := []byte("test data")
	hash := ComputeActivityHash(data)

	// Verify against standard library
	expected := sha256.Sum256(data)
	expectedHex := hex.EncodeToString(expected[:])
	if hash != expectedHex {
		t.Errorf("ComputeActivityHash = %s, want %s", hash, expectedHex)
	}
}

func TestComputeActivityHash_EmptyInput(t *testing.T) {
	hash := ComputeActivityHash([]byte{})
	// SHA-256 of empty input
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expected {
		t.Errorf("ComputeActivityHash(empty) = %s, want %s", hash, expected)
	}
}

func TestComputeEpoch_BoundaryBlocks(t *testing.T) {
	// Verify exact boundary between epoch 0 and epoch 1
	epochLen := int64(17280)
	if ComputeEpoch(epochLen, epochLen) != 0 {
		t.Error("block 17280 should be in epoch 0")
	}
	if ComputeEpoch(epochLen+1, epochLen) != 1 {
		t.Error("block 17281 should be in epoch 1")
	}
}

func TestComputeWindow_AllWindows(t *testing.T) {
	// Verify all 12 windows in an epoch map correctly
	epochLen := int64(17280)
	windowsPerEpoch := int64(12)
	windowLen := epochLen / windowsPerEpoch // 1440

	for w := int64(0); w < windowsPerEpoch; w++ {
		blockHeight := w*windowLen + 1 // first block of each window
		result := ComputeWindow(blockHeight, epochLen, windowsPerEpoch)
		if result != w {
			t.Errorf("block %d should be window %d, got %d", blockHeight, w, result)
		}
	}
}

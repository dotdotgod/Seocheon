package utility

import (
	"crypto/sha256"
	"encoding/hex"
)

// VerifyActivityHash validates that the given string is a valid activity hash.
// A valid activity hash is exactly 64 hex characters (representing 32 bytes / SHA-256).
func VerifyActivityHash(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	_, err := hex.DecodeString(hash)
	return err == nil
}

// ComputeActivityHash computes a SHA-256 hash of the given data and returns
// it as a 64-character hex string suitable for use as an activity_hash.
func ComputeActivityHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// ComputeEpoch calculates the epoch number for a given block height.
// epoch_number = (block_height - 1) / epoch_length
func ComputeEpoch(blockHeight, epochLength int64) int64 {
	if blockHeight <= 0 || epochLength <= 0 {
		return 0
	}
	return (blockHeight - 1) / epochLength
}

// ComputeWindow calculates the window index within an epoch for a given block height.
// window_index = ((block_height - 1) % epoch_length) / window_length
func ComputeWindow(blockHeight, epochLength, windowsPerEpoch int64) int64 {
	if epochLength <= 0 || windowsPerEpoch <= 0 {
		return 0
	}
	windowLength := epochLength / windowsPerEpoch
	return ((blockHeight - 1) % epochLength) / windowLength
}

// EpochStartBlock returns the starting block height for the given epoch number.
func EpochStartBlock(epochNumber, epochLength int64) int64 {
	return epochNumber*epochLength + 1
}

// EpochEndBlock returns the ending block height for the given epoch number.
func EpochEndBlock(epochNumber, epochLength int64) int64 {
	return (epochNumber + 1) * epochLength
}

// WindowStartBlock returns the starting block height for a given window within an epoch.
func WindowStartBlock(epochStartBlock, windowIndex, windowLength int64) int64 {
	return epochStartBlock + windowIndex*windowLength
}

// WindowEndBlock returns the ending block height for a given window.
func WindowEndBlock(windowStart, windowLength int64) int64 {
	return windowStart + windowLength - 1
}

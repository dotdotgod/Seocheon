package keeper

import "seocheon/x/activity/types"

// GetCurrentEpoch returns the current epoch number for the given block height.
// Epoch 0 starts at block 1, epoch 1 starts at block EpochLength+1, etc.
func GetCurrentEpoch(blockHeight int64, params types.Params) int64 {
	if blockHeight <= 0 {
		return 0
	}
	return (blockHeight - 1) / params.EpochLength
}

// GetCurrentWindow returns the current window number within the epoch (0-indexed).
func GetCurrentWindow(blockHeight int64, params types.Params) int64 {
	if blockHeight <= 0 {
		return 0
	}
	windowLength := GetWindowLength(params)
	blockInEpoch := (blockHeight - 1) % params.EpochLength
	return blockInEpoch / windowLength
}

// GetWindowLength returns the number of blocks per window.
func GetWindowLength(params types.Params) int64 {
	return params.EpochLength / params.WindowsPerEpoch
}

// IsEpochBoundary returns true if the given block height is an epoch boundary
// (the last block of an epoch, i.e., the block where epoch transitions happen).
func IsEpochBoundary(blockHeight int64, params types.Params) bool {
	if blockHeight <= 0 {
		return false
	}
	return blockHeight%params.EpochLength == 0
}

// IsWindowBoundary returns true if the given block height is a window boundary.
func IsWindowBoundary(blockHeight int64, params types.Params) bool {
	if blockHeight <= 0 {
		return false
	}
	windowLength := GetWindowLength(params)
	return blockHeight%windowLength == 0
}

// GetEpochStartBlock returns the first block height of the given epoch.
func GetEpochStartBlock(epoch int64, params types.Params) int64 {
	return epoch*params.EpochLength + 1
}

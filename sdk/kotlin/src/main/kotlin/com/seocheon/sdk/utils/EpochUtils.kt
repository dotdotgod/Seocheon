package com.seocheon.sdk.utils

import com.seocheon.sdk.constants.ChainConstants

/**
 * Epoch and window calculation utilities.
 */
object EpochUtils {

    /**
     * Calculates the epoch number for a given block height.
     * epoch_number = (block_height - 1) / epoch_length
     */
    fun computeEpoch(
        blockHeight: Long,
        epochLength: Long = ChainConstants.EPOCH_LENGTH,
    ): Long {
        if (blockHeight <= 0 || epochLength <= 0) return 0
        return (blockHeight - 1) / epochLength
    }

    /**
     * Calculates the window index within an epoch for a given block height.
     * window_index = ((block_height - 1) % epoch_length) / window_length
     */
    fun computeWindow(
        blockHeight: Long,
        epochLength: Long = ChainConstants.EPOCH_LENGTH,
        windowsPerEpoch: Long = ChainConstants.WINDOWS_PER_EPOCH,
    ): Long {
        if (epochLength <= 0 || windowsPerEpoch <= 0) return 0
        val windowLength = epochLength / windowsPerEpoch
        return ((blockHeight - 1) % epochLength) / windowLength
    }

    /**
     * Returns the starting block height for the given epoch number.
     */
    fun epochStartBlock(
        epochNumber: Long,
        epochLength: Long = ChainConstants.EPOCH_LENGTH,
    ): Long {
        return epochNumber * epochLength + 1
    }

    /**
     * Returns the ending block height for the given epoch number.
     */
    fun epochEndBlock(
        epochNumber: Long,
        epochLength: Long = ChainConstants.EPOCH_LENGTH,
    ): Long {
        return (epochNumber + 1) * epochLength
    }

    /**
     * Returns the starting block height for a given window within an epoch.
     */
    fun windowStartBlock(
        epochStartBlock: Long,
        windowIndex: Long,
        windowLength: Long = ChainConstants.WINDOW_LENGTH,
    ): Long {
        return epochStartBlock + windowIndex * windowLength
    }

    /**
     * Returns the ending block height for a given window.
     */
    fun windowEndBlock(
        windowStart: Long,
        windowLength: Long = ChainConstants.WINDOW_LENGTH,
    ): Long {
        return windowStart + windowLength - 1
    }
}

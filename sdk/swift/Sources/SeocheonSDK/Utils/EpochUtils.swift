import Foundation

/// Epoch and window calculation utilities.
public enum EpochUtils {
    /// Calculates the epoch number for a given block height.
    /// epoch_number = (block_height - 1) / epoch_length
    public static func computeEpoch(blockHeight: Int64, epochLength: Int64 = ChainConstants.epochLength) -> Int64 {
        guard blockHeight > 0, epochLength > 0 else { return 0 }
        return (blockHeight - 1) / epochLength
    }

    /// Calculates the window index within an epoch.
    /// window_index = ((block_height - 1) % epoch_length) / window_length
    public static func computeWindow(
        blockHeight: Int64,
        epochLength: Int64 = ChainConstants.epochLength,
        windowsPerEpoch: Int64 = ChainConstants.windowsPerEpoch
    ) -> Int64 {
        guard epochLength > 0, windowsPerEpoch > 0 else { return 0 }
        let windowLength = epochLength / windowsPerEpoch
        return ((blockHeight - 1) % epochLength) / windowLength
    }

    /// Returns the starting block height for the given epoch number.
    public static func epochStartBlock(epochNumber: Int64, epochLength: Int64 = ChainConstants.epochLength) -> Int64 {
        return epochNumber * epochLength + 1
    }

    /// Returns the ending block height for the given epoch number.
    public static func epochEndBlock(epochNumber: Int64, epochLength: Int64 = ChainConstants.epochLength) -> Int64 {
        return (epochNumber + 1) * epochLength
    }

    /// Returns the starting block height for a given window within an epoch.
    public static func windowStartBlock(epochStartBlock: Int64, windowIndex: Int64, windowLength: Int64 = ChainConstants.windowLength) -> Int64 {
        return epochStartBlock + windowIndex * windowLength
    }

    /// Returns the ending block height for a given window.
    public static func windowEndBlock(windowStart: Int64, windowLength: Int64 = ChainConstants.windowLength) -> Int64 {
        return windowStart + windowLength - 1
    }
}

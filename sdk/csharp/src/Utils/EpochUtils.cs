using Seocheon.Sdk.Constants;

namespace Seocheon.Sdk.Utils;

/// <summary>
/// Epoch and window computation utilities.
/// Block 1 is the first block. Epoch 0 starts at block 1.
/// </summary>
public static class EpochUtils
{
    /// <summary>
    /// Computes the epoch number for a given block height.
    /// </summary>
    public static long ComputeEpoch(long blockHeight, long epochLength = ChainConstants.EpochBlocks)
    {
        if (blockHeight <= 0) return 0;
        return (blockHeight - 1) / epochLength;
    }

    /// <summary>
    /// Computes the window index (0-based) within the current epoch.
    /// </summary>
    public static long ComputeWindow(
        long blockHeight,
        long epochLength = ChainConstants.EpochBlocks,
        int windowsPerEpoch = ChainConstants.WindowsPerEpoch)
    {
        if (blockHeight <= 0) return 0;

        var windowLength = epochLength / windowsPerEpoch;
        var positionInEpoch = (blockHeight - 1) % epochLength;
        return positionInEpoch / windowLength;
    }

    /// <summary>
    /// Returns the first block of the given epoch.
    /// </summary>
    public static long EpochStartBlock(long epochNumber, long epochLength = ChainConstants.EpochBlocks)
    {
        return epochNumber * epochLength + 1;
    }

    /// <summary>
    /// Returns the last block of the given epoch.
    /// </summary>
    public static long EpochEndBlock(long epochNumber, long epochLength = ChainConstants.EpochBlocks)
    {
        return (epochNumber + 1) * epochLength;
    }

    /// <summary>
    /// Returns the first block of a window within an epoch.
    /// </summary>
    public static long WindowStartBlock(long epochStartBlock, long windowIndex, long windowLength = ChainConstants.WindowBlocks)
    {
        return epochStartBlock + windowIndex * windowLength;
    }

    /// <summary>
    /// Returns the last block of a window.
    /// </summary>
    public static long WindowEndBlock(long windowStartBlock, long windowLength = ChainConstants.WindowBlocks)
    {
        return windowStartBlock + windowLength - 1;
    }

    /// <summary>
    /// Computes the number of elapsed windows in the current epoch at the given block height.
    /// </summary>
    public static long ElapsedWindows(long blockHeight, long epochLength = ChainConstants.EpochBlocks, int windowsPerEpoch = ChainConstants.WindowsPerEpoch)
    {
        var currentWindow = ComputeWindow(blockHeight, epochLength, windowsPerEpoch);
        return currentWindow + 1;
    }

    /// <summary>
    /// Returns blocks remaining until the next epoch boundary.
    /// </summary>
    public static long BlocksUntilNextEpoch(long blockHeight, long epochLength = ChainConstants.EpochBlocks)
    {
        if (blockHeight <= 0) return epochLength;
        var epoch = ComputeEpoch(blockHeight, epochLength);
        var endBlock = EpochEndBlock(epoch, epochLength);
        return endBlock - blockHeight;
    }

    /// <summary>
    /// Returns blocks remaining until the next window boundary.
    /// </summary>
    public static long BlocksUntilNextWindow(
        long blockHeight,
        long epochLength = ChainConstants.EpochBlocks,
        int windowsPerEpoch = ChainConstants.WindowsPerEpoch)
    {
        if (blockHeight <= 0) return epochLength / windowsPerEpoch;

        var windowLength = epochLength / windowsPerEpoch;
        var positionInEpoch = (blockHeight - 1) % epochLength;
        var positionInWindow = positionInEpoch % windowLength;
        return windowLength - positionInWindow - 1;
    }
}

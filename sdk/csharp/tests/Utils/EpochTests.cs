using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Utils;
using Xunit;

namespace Seocheon.Sdk.Tests.Utils;

public class EpochTests
{
    [Theory]
    [InlineData(1, 0)]        // First block → epoch 0
    [InlineData(17280, 0)]    // Last block of epoch 0
    [InlineData(17281, 1)]    // First block of epoch 1
    [InlineData(34560, 1)]    // Last block of epoch 1
    [InlineData(34561, 2)]    // First block of epoch 2
    public void ComputeEpoch(long blockHeight, long expectedEpoch)
    {
        Assert.Equal(expectedEpoch, EpochUtils.ComputeEpoch(blockHeight));
    }

    [Theory]
    [InlineData(1, 0)]        // First block → window 0
    [InlineData(1440, 0)]     // Last block of window 0
    [InlineData(1441, 1)]     // First block of window 1
    [InlineData(17280, 11)]   // Last block of epoch → window 11
    [InlineData(17281, 0)]    // First block of epoch 1 → window 0
    public void ComputeWindow(long blockHeight, long expectedWindow)
    {
        Assert.Equal(expectedWindow, EpochUtils.ComputeWindow(blockHeight));
    }

    [Theory]
    [InlineData(0, 1)]       // Epoch 0 starts at block 1
    [InlineData(1, 17281)]   // Epoch 1 starts at block 17281
    [InlineData(2, 34561)]   // Epoch 2 starts at block 34561
    public void EpochStartBlock(long epochNumber, long expectedStart)
    {
        Assert.Equal(expectedStart, EpochUtils.EpochStartBlock(epochNumber));
    }

    [Theory]
    [InlineData(0, 17280)]   // Epoch 0 ends at block 17280
    [InlineData(1, 34560)]   // Epoch 1 ends at block 34560
    public void EpochEndBlock(long epochNumber, long expectedEnd)
    {
        Assert.Equal(expectedEnd, EpochUtils.EpochEndBlock(epochNumber));
    }

    [Fact]
    public void WindowStartBlock_EpochZero()
    {
        var epochStart = EpochUtils.EpochStartBlock(0); // 1
        Assert.Equal(1, EpochUtils.WindowStartBlock(epochStart, 0));
        Assert.Equal(1441, EpochUtils.WindowStartBlock(epochStart, 1));
        Assert.Equal(2881, EpochUtils.WindowStartBlock(epochStart, 2));
    }

    [Fact]
    public void WindowEndBlock_Calculation()
    {
        var windowStart = EpochUtils.WindowStartBlock(1, 0); // = 1
        Assert.Equal(1440, EpochUtils.WindowEndBlock(windowStart));
    }

    [Theory]
    [InlineData(0, 17280)]
    [InlineData(1, 17279)]
    [InlineData(17280, 0)]
    [InlineData(17281, 17279)]
    public void BlocksUntilNextEpoch(long blockHeight, long expected)
    {
        Assert.Equal(expected, EpochUtils.BlocksUntilNextEpoch(blockHeight == 0 ? 0 : blockHeight));
    }

    [Fact]
    public void BlocksUntilNextWindow_FirstBlock()
    {
        Assert.Equal(1439, EpochUtils.BlocksUntilNextWindow(1));
    }

    [Fact]
    public void BlocksUntilNextWindow_LastBlockOfWindow()
    {
        Assert.Equal(0, EpochUtils.BlocksUntilNextWindow(1440));
    }

    [Fact]
    public void ElapsedWindows_MiddleOfEpoch()
    {
        // Block 2881 = window 2 (0-indexed) → elapsed = 3
        Assert.Equal(3, EpochUtils.ElapsedWindows(2881));
    }

    [Fact]
    public void ZeroBlockHeight_ReturnsZero()
    {
        Assert.Equal(0, EpochUtils.ComputeEpoch(0));
        Assert.Equal(0, EpochUtils.ComputeWindow(0));
    }
}

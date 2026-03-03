using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Modules;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Modules;

public class EpochModuleTests
{
    private readonly MockChainClient _client = new();
    private readonly MockSigner _signer = new();

    private EpochModule CreateModule()
    {
        _client.ConnectAsync().Wait();
        return new EpochModule(_client, _signer);
    }

    [Fact]
    public async Task GetInfo_ReturnsCurrentEpoch()
    {
        _client.CurrentBlockHeight = 17281; // Epoch 1, Window 0
        var module = CreateModule();

        var info = await module.GetInfo();

        Assert.Equal(17281, info.BlockHeight);
        Assert.Equal(1, info.EpochNumber);
        Assert.Equal(0, info.WindowNumber);
        Assert.Equal(17281, info.EpochStartBlock);
        Assert.Equal(34560, info.EpochEndBlock);
    }

    [Fact]
    public async Task GetInfo_WindowProgress()
    {
        _client.CurrentBlockHeight = 1;
        var module = CreateModule();

        var info = await module.GetInfo();
        Assert.Contains("/", info.WindowProgress);
        Assert.Contains("/", info.EpochProgress);
    }

    [Fact]
    public async Task GetQualification_ReturnsQualification()
    {
        var module = CreateModule();
        _client.SetQueryResponse("qualification", new
        {
            elapsed_windows = "6",
            active_windows = "5",
            window_detail = new[]
            {
                new { window_number = "0", submission_count = "2" },
                new { window_number = "1", submission_count = "1" },
                new { window_number = "2", submission_count = "0" },
                new { window_number = "3", submission_count = "1" },
                new { window_number = "4", submission_count = "3" },
                new { window_number = "5", submission_count = "1" }
            }
        });

        var result = await module.GetQualification("node-1", 1);

        Assert.Equal(1, result.EpochNumber);
        Assert.Equal(ChainConstants.WindowsPerEpoch, (int)result.TotalWindows);
        Assert.Equal(5UL, result.ActiveWindows);
        Assert.False(result.IsQualified); // 5 < 8
        Assert.True(result.CanStillQualify); // 3 needed, 6 windows left
    }
}

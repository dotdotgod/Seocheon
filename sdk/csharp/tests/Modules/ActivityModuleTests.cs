using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Modules;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Modules;

public class ActivityModuleTests
{
    private readonly MockChainClient _client = new();
    private readonly MockSigner _signer = new();

    private readonly PipelineConfig _config = new()
    {
        ChainId = "seocheon-test-1",
        GasPrice = 250,
        ConfirmTimeout = TimeSpan.FromSeconds(5),
        PollInterval = TimeSpan.FromMilliseconds(100)
    };

    private ActivityModule CreateModule()
    {
        _client.ConnectAsync().Wait();
        return new ActivityModule(_client, _signer, _config);
    }

    [Fact]
    public async Task Submit_InvalidHash_Throws()
    {
        var module = CreateModule();
        await Assert.ThrowsAsync<SdkException>(
            () => module.Submit("short", "https://example.com")
        );
    }

    [Fact]
    public async Task Submit_EmptyContentUri_Throws()
    {
        var module = CreateModule();
        var hash = "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3";
        await Assert.ThrowsAsync<SdkException>(
            () => module.Submit(hash, "")
        );
    }

    [Fact]
    public async Task Submit_ValidInput_Succeeds()
    {
        var module = CreateModule();
        var hash = "a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3";

        var result = await module.Submit(hash, "https://example.com/report");

        Assert.NotEmpty(result.TxHash);
        Assert.True(result.BlockHeight > 0);
    }

    [Fact]
    public async Task GetQuota_ReturnsResponse()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/nodes/by-agent/", new
        {
            node = new { id = "node-1" }
        });
        _client.SetQueryResponse("/seocheon/activity/v1/nodes/", new
        {
            quota_limit = "10",
            quota_used = "3"
        });

        var result = await module.GetQuota();
        Assert.Equal(1, result.EpochNumber);
        Assert.Equal(10UL, result.QuotaTotal);
        Assert.Equal(3UL, result.QuotaUsed);
        Assert.Equal(7UL, result.QuotaRemaining);
        Assert.False(result.IsFeegrant);
    }

    [Fact]
    public async Task GetActivities_WithNodeId()
    {
        var module = CreateModule();
        _client.SetQueryResponse("activities", new
        {
            activities = new[]
            {
                new
                {
                    activity_hash = "abc123def456abc123def456abc123def456abc123def456abc123def456abc1",
                    content_uri = "https://example.com",
                    block_height = "17300",
                    tx_hash = "TX123"
                }
            },
            total_count = "1"
        });

        var result = await module.GetActivities("node-1", 1);
        Assert.Single(result.Activities);
        Assert.Equal(1UL, result.TotalCount);
    }
}

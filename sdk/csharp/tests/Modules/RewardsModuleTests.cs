using Seocheon.Sdk.Modules;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Modules;

public class RewardsModuleTests
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

    private RewardsModule CreateModule()
    {
        _client.ConnectAsync().Wait();
        return new RewardsModule(_client, _signer, _config);
    }

    [Fact]
    public async Task GetPending_ReturnsRewards()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/rewards/", new
        {
            delegation_reward = "500000",
            activity_reward = "300000",
            total_reward = "800000",
            commission_total = "100000"
        });

        var result = await module.GetPending("node-1");
        Assert.Equal("500000", result.DelegationReward);
        Assert.Equal("300000", result.ActivityReward);
        Assert.Equal("800000", result.TotalReward);
        Assert.Equal("100000", result.CommissionTotal);
        // 80/20 split
        Assert.Equal("80000", result.OperatorShare);
        Assert.Equal("20000", result.AgentShare);
    }

    [Fact]
    public async Task GetPending_ZeroCommission()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/rewards/", new
        {
            delegation_reward = "0",
            activity_reward = "0",
            total_reward = "0",
            commission_total = "0"
        });

        var result = await module.GetPending("node-1");
        Assert.Equal("0", result.OperatorShare);
        Assert.Equal("0", result.AgentShare);
    }

    [Fact]
    public async Task Withdraw_Succeeds()
    {
        var module = CreateModule();

        var result = await module.Withdraw();
        Assert.NotEmpty(result.TxHash);
    }

    [Fact]
    public async Task Withdraw_ParsesEvents()
    {
        var module = CreateModule();

        _client.SetTxResponse(
            "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234",
            new TxResponse
            {
                TxHash = "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234",
                Height = 17300,
                Code = 0,
                Events =
                [
                    new TxEventData("withdraw_commission", new Dictionary<string, string>
                    {
                        ["amount"] = "500000"
                    })
                ]
            }
        );

        var result = await module.Withdraw();
        Assert.Equal("500000", result.WithdrawnTotal);
        Assert.Equal("400000", result.ToOperator); // 80%
        Assert.Equal("100000", result.ToAgent);     // 20%
    }

    [Fact]
    public async Task GetPending_WithNodeIdResolution()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/node_by_agent/", new
        {
            node = new { node_id = "resolved-node" }
        });
        _client.SetQueryResponse("/seocheon/node/v1/rewards/", new
        {
            delegation_reward = "100",
            activity_reward = "200",
            total_reward = "300",
            commission_total = "50"
        });

        var result = await module.GetPending();
        Assert.Equal("300", result.TotalReward);
    }
}

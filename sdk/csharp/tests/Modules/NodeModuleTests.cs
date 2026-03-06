using Seocheon.Sdk.Modules;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Modules;

public class NodeModuleTests
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

    private NodeModule CreateModule()
    {
        _client.ConnectAsync().Wait();
        return new NodeModule(_client, _signer, _config);
    }

    [Fact]
    public async Task GetInfo_ByNodeId()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/nodes/", new
        {
            node = new
            {
                id = "node-1",
                @operator = "seocheon1operator",
                agent_address = "seocheon1agent",
                status = "ACTIVE",
                description = "Test node",
                website = "https://example.com",
                tags = new[] { "ai", "research" },
                commission_rate = "0.1",
                agent_share = "0.2",
                total_delegation = "1000000",
                self_delegation = "500000",
                validator_address = "seocheonvaloper1xxx",
                registered_at = "100"
            }
        });

        var info = await module.GetInfo("node-1");
        Assert.Equal("node-1", info.NodeId);
        Assert.Equal("ACTIVE", info.Status);
        Assert.Equal(2, info.Tags.Count);
        Assert.Contains("ai", info.Tags);
    }

    [Fact]
    public async Task Search_ReturnsNodes()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/nodes", new
        {
            nodes = new[]
            {
                new
                {
                    id = "node-1",
                    status = "ACTIVE",
                    tags = new[] { "ai" },
                    total_delegation = "1000000",
                    description = "AI node"
                },
                new
                {
                    id = "node-2",
                    status = "REGISTERED",
                    tags = new[] { "research" },
                    total_delegation = "500000",
                    description = "Research node"
                }
            },
            total_count = "2"
        });

        var result = await module.Search();
        Assert.Equal(2, result.Nodes.Count);
        Assert.Equal(2UL, result.TotalCount);
    }

    [Fact]
    public async Task Search_WithFilters()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/nodes", new
        {
            nodes = new[]
            {
                new
                {
                    id = "node-1",
                    status = "ACTIVE",
                    tags = new[] { "ai" },
                    total_delegation = "1000000",
                    description = "AI node"
                }
            },
            total_count = "1"
        });

        var result = await module.Search(tag: "ai", status: "ACTIVE", limit: 10);
        Assert.Single(result.Nodes);
    }

    [Fact]
    public async Task GetDelegationStatus_ReturnsResponse()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/delegation-confirmation/", new
        {
            expiry_epoch = "90",
            current_epoch = "5",
            in_renewal_window = false,
            renewal_window_start = "83"
        });

        var result = await module.GetDelegationStatus("seocheon1delegator", "seocheonvaloper1val");
        Assert.Equal(90L, result.ExpiryEpoch);
        Assert.Equal(5L, result.CurrentEpoch);
        Assert.False(result.InRenewalWindow);
        Assert.Equal(83L, result.RenewalWindowStart);
    }

    [Fact]
    public async Task ConfirmDelegation_ReturnsTxHash()
    {
        var module = CreateModule();
        var result = await module.ConfirmDelegation("seocheonvaloper1val");
        Assert.NotNull(result.TxHash);
        Assert.NotEmpty(result.TxHash);
        Assert.Equal(150000UL, result.GasUsed);
        Assert.Equal(200000UL, result.GasWanted);
    }
}

using Seocheon.Sdk.Modules;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Modules;

public class NodeModuleTests
{
    private readonly MockChainClient _client = new();
    private readonly MockSigner _signer = new();

    private NodeModule CreateModule()
    {
        _client.ConnectAsync().Wait();
        return new NodeModule(_client, _signer);
    }

    [Fact]
    public async Task GetInfo_ByNodeId()
    {
        var module = CreateModule();
        _client.SetQueryResponse("/seocheon/node/v1/node/", new
        {
            node = new
            {
                node_id = "node-1",
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
                    node_id = "node-1",
                    status = "ACTIVE",
                    tags = new[] { "ai" },
                    total_delegation = "1000000",
                    description = "AI node"
                },
                new
                {
                    node_id = "node-2",
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
                    node_id = "node-1",
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
}

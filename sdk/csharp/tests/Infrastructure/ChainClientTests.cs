using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Infrastructure;

public class ChainClientTests
{
    [Fact]
    public async Task ConnectAndDisconnect()
    {
        var client = new MockChainClient();
        Assert.False(client.IsConnected);

        await client.ConnectAsync();
        Assert.True(client.IsConnected);

        await client.DisconnectAsync();
        Assert.False(client.IsConnected);
    }

    [Fact]
    public async Task GetLatestBlock_ReturnsBlockHeight()
    {
        var client = new MockChainClient { CurrentBlockHeight = 12345 };
        await client.ConnectAsync();

        var block = await client.GetLatestBlock();
        var height = block.GetProperty("block").GetProperty("header").GetProperty("height").GetString();
        Assert.Equal("12345", height);
    }

    [Fact]
    public async Task GetAccountInfo_ReturnsInfo()
    {
        var client = new MockChainClient();
        await client.ConnectAsync();

        var info = await client.GetAccountInfo("seocheon1test");
        Assert.Equal(42UL, info.AccountNumber);
        Assert.Equal(7UL, info.Sequence);
    }

    [Fact]
    public async Task BroadcastTx_ReturnsHash()
    {
        var client = new MockChainClient();
        await client.ConnectAsync();

        var result = await client.BroadcastTx([0x01], "sync");
        Assert.NotEmpty(result.TxHash);
        Assert.Equal(0u, result.Code);
    }

    [Fact]
    public async Task HttpChainClient_NotConnected_Throws()
    {
        var client = new HttpChainClient("http://localhost:26657", "http://localhost:1317");
        await Assert.ThrowsAsync<SdkException>(() => client.QueryRest("/test"));
    }
}

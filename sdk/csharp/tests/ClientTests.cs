using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests;

public class ClientTests
{
    private static SdkConfig TestConfig() => new()
    {
        Chain = new ChainConfig
        {
            ChainId = "seocheon-test-1",
            RpcEndpoint = "http://localhost:26657",
            GrpcEndpoint = "http://localhost:1317"
        },
        Signing = new SigningConfig
        {
            Mode = SigningMode.Direct,
            Mnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
        }
    };

    [Fact]
    public void CreateSdk_WithMock()
    {
        var client = new MockChainClient();
        var signer = new MockSigner();

        var sdk = new SeocheonSdk(TestConfig(), client, signer);

        Assert.NotNull(sdk.Activity);
        Assert.NotNull(sdk.Epoch);
        Assert.NotNull(sdk.Node);
        Assert.NotNull(sdk.Rewards);
        Assert.NotNull(sdk.Cosmos);
        Assert.False(sdk.IsConnected);
    }

    [Fact]
    public async Task ConnectAndDisconnect()
    {
        var client = new MockChainClient();
        var signer = new MockSigner();

        using var sdk = new SeocheonSdk(TestConfig(), client, signer);

        await sdk.ConnectAsync();
        Assert.True(sdk.IsConnected);

        await sdk.DisconnectAsync();
        Assert.False(sdk.IsConnected);
    }

    [Fact]
    public void InvalidConfig_Throws()
    {
        var config = new SdkConfig
        {
            Chain = new ChainConfig { ChainId = "", RpcEndpoint = "", GrpcEndpoint = "" },
            Signing = new SigningConfig { Mode = SigningMode.Direct }
        };

        Assert.Throws<SdkException>(() => new SeocheonSdk(config));
    }

    [Fact]
    public void Config_Property_ReturnsConfig()
    {
        var config = TestConfig();
        var client = new MockChainClient();
        var signer = new MockSigner();

        var sdk = new SeocheonSdk(config, client, signer);
        Assert.Equal("seocheon-test-1", sdk.Config.Chain.ChainId);
    }

    [Fact]
    public void GetAddress_ReturnsMockAddress()
    {
        var client = new MockChainClient();
        var signer = new MockSigner { Address = "seocheon1myaddress" };

        var sdk = new SeocheonSdk(TestConfig(), client, signer);
        Assert.Equal("seocheon1myaddress", sdk.GetAddress());
    }
}

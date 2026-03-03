using Seocheon.Sdk.Modules;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Modules;

public class CosmosModuleTests
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

    private CosmosModule CreateModule()
    {
        _client.ConnectAsync().Wait();
        return new CosmosModule(_client, _signer, _config);
    }

    [Fact]
    public async Task GetBalance_ReturnsFormattedBalance()
    {
        var module = CreateModule();
        _client.SetQueryResponse("balances", new
        {
            balance = new { denom = "uppyeo", amount = "10000000000" }
        });

        var result = await module.GetBalance("seocheon1test");
        Assert.Equal("seocheon1test", result.Address);
        Assert.Equal("10000000000", result.Balance);
        Assert.Equal("1.0000000000", result.BalanceKkot);
    }

    [Fact]
    public async Task GetBalance_DefaultsToSignerAddress()
    {
        var module = CreateModule();
        _client.SetQueryResponse("balances", new
        {
            balance = new { denom = "uppyeo", amount = "5000000000" }
        });

        var result = await module.GetBalance();
        Assert.Equal(_signer.Address, result.Address);
    }

    [Fact]
    public async Task SendTokens_Succeeds()
    {
        var module = CreateModule();

        var result = await module.SendTokens("seocheon1recipient", "1000000", "uppyeo");
        Assert.NotEmpty(result.TxHash);
        Assert.True(result.BlockHeight > 0);
    }

    [Fact]
    public async Task GetBlockInfo_ReturnsInfo()
    {
        _client.CurrentBlockHeight = 54321;
        var module = CreateModule();

        var info = await module.GetBlockInfo();
        Assert.Equal(54321, info.BlockHeight);
        Assert.Equal("seocheon-testnet-1", info.ChainId);
        Assert.NotEmpty(info.BlockTime);
    }

    [Fact]
    public async Task GetTxResult_ReturnsTxInfo()
    {
        var module = CreateModule();
        _client.SetTxResponse("TX_HASH_123", new TxResponse
        {
            TxHash = "TX_HASH_123",
            Height = 12345,
            Code = 0,
            RawLog = "success",
            Events =
            [
                new TxEventData("transfer", new Dictionary<string, string>
                {
                    ["sender"] = "seocheon1from",
                    ["recipient"] = "seocheon1to",
                    ["amount"] = "1000uppyeo"
                })
            ]
        });

        var result = await module.GetTxResult("TX_HASH_123");
        Assert.Equal("TX_HASH_123", result.TxHash);
        Assert.Equal(12345, result.Height);
        Assert.Equal(0u, result.Code);
        Assert.Single(result.Events);
    }
}

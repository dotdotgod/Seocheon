using Seocheon.Sdk;
using Xunit;
using Xunit.Abstractions;

namespace Seocheon.Sdk.Tests;

/// <summary>
/// E2E integration tests for the Seocheon C# SDK.
///
/// Skip conditions:
///   - SEOCHEON_GRPC not set
///   - SEOCHEON_MNEMONIC not set
///
/// Run with:
///   dotnet test --filter "Category=e2e"
/// </summary>
[Trait("Category", "e2e")]
public sealed class E2EIntegrationTests : IAsyncLifetime
{
    private static readonly string Grpc = Environment.GetEnvironmentVariable("SEOCHEON_GRPC") ?? "";
    private static readonly string Mnemonic = Environment.GetEnvironmentVariable("SEOCHEON_MNEMONIC") ?? "";
    private static readonly string Rpc = Environment.GetEnvironmentVariable("SEOCHEON_RPC") ?? "http://localhost:26657";
    private static readonly string ChainId = Environment.GetEnvironmentVariable("SEOCHEON_CHAIN_ID") ?? "seocheon-e2e";

    private static bool ShouldSkip => string.IsNullOrEmpty(Grpc) || string.IsNullOrEmpty(Mnemonic);

    private SeocheonSdk? _sdk;
    private readonly ITestOutputHelper _output;

    public E2EIntegrationTests(ITestOutputHelper output)
    {
        _output = output;
    }

    private static SdkConfig BuildConfig() => new()
    {
        Chain = new ChainConfig
        {
            ChainId = ChainId,
            RpcEndpoint = Rpc,
            GrpcEndpoint = Grpc,
            GasPrice = "250uppyeo",
        },
        Signing = new SigningConfig
        {
            Mode = SigningMode.Direct,
            Mnemonic = Mnemonic,
        },
    };

    public async Task InitializeAsync()
    {
        if (ShouldSkip) return;
        _sdk = new SeocheonSdk(BuildConfig());
        await _sdk.ConnectAsync();
    }

    public async Task DisposeAsync()
    {
        if (_sdk is not null)
            await _sdk.DisconnectAsync();
    }

    [Fact]
    public void Connect_IsConnected_ReturnTrue()
    {
        if (ShouldSkip)
        {
            _output.WriteLine("E2E 스킵: SEOCHEON_GRPC 또는 SEOCHEON_MNEMONIC 미설정");
            return;
        }
        Assert.True(_sdk!.IsConnected, "Connect 후 IsConnected = false");
    }

    [Fact]
    public async Task GetBlockInfo_ReturnsPositiveBlockHeight()
    {
        if (ShouldSkip)
        {
            _output.WriteLine("E2E 스킵: SEOCHEON_GRPC 또는 SEOCHEON_MNEMONIC 미설정");
            return;
        }
        var block = await _sdk!.Cosmos.GetBlockInfo();
        Assert.True(block.BlockHeight > 0, $"블록 높이가 양수여야 함: {block.BlockHeight}");
        _output.WriteLine($"최신 블록: height={block.BlockHeight} chainId={block.ChainId}");
    }

    [Fact]
    public async Task NodeSearch_EndpointResponds()
    {
        if (ShouldSkip)
        {
            _output.WriteLine("E2E 스킵: SEOCHEON_GRPC 또는 SEOCHEON_MNEMONIC 미설정");
            return;
        }
        var resp = await _sdk!.Node.Search(limit: 10);
        Assert.NotNull(resp);
        _output.WriteLine($"x/node 조회 성공: total={resp.TotalCount}");
    }

    [Fact]
    public async Task EpochGetInfo_ReturnsValidEpoch()
    {
        if (ShouldSkip)
        {
            _output.WriteLine("E2E 스킵: SEOCHEON_GRPC 또는 SEOCHEON_MNEMONIC 미설정");
            return;
        }
        var info = await _sdk!.Epoch.GetInfo();
        Assert.True(info.BlockHeight > 0, $"에포크 블록 높이가 양수여야 함: {info.BlockHeight}");
        _output.WriteLine($"에포크: epoch={info.EpochNumber} window={info.WindowNumber} height={info.BlockHeight}");
    }

    [Fact]
    public async Task GetBalance_ReturnsNonNegative()
    {
        if (ShouldSkip)
        {
            _output.WriteLine("E2E 스킵: SEOCHEON_GRPC 또는 SEOCHEON_MNEMONIC 미설정");
            return;
        }
        var result = await _sdk!.Cosmos.GetBalance(_sdk.GetAddress());
        Assert.NotNull(result.Balance);
        _output.WriteLine($"잔액: {result.Balance} uppyeo ({result.BalanceKkot} KKOT)");
    }

    [Fact]
    public async Task SubmitActivity_ReturnsValidTxHash()
    {
        if (ShouldSkip)
        {
            _output.WriteLine("E2E 스킵: SEOCHEON_GRPC 또는 SEOCHEON_MNEMONIC 미설정");
            return;
        }
        var hash = new string('a', 64);
        var result = await _sdk!.Activity.Submit(hash, "ipfs://QmTestCSharpE2E");
        Assert.NotEmpty(result.TxHash);
        _output.WriteLine($"활동 제출 성공: txHash={result.TxHash} height={result.BlockHeight}");
    }
}

using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Tests.TestHelpers;
using Xunit;

namespace Seocheon.Sdk.Tests.Internal.Tx;

public class PipelineTests
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

    [Fact]
    public async Task ExecuteTx_SuccessfulSubmission()
    {
        await _client.ConnectAsync();
        var request = new TxRequest
        {
            Message = Messages.EncodeMsgSubmitActivity("addr", "hash", "uri"),
            MessageTypeUrl = Messages.TypeMsgSubmitActivity
        };

        var result = await Pipeline.ExecuteTx(_client, _signer, _config, request);

        Assert.NotEmpty(result.TxHash);
        Assert.True(result.Height > 0);
        Assert.Equal(0u, result.Code);
    }

    [Fact]
    public async Task ExecuteTx_WithCustomGas()
    {
        await _client.ConnectAsync();
        var request = new TxRequest
        {
            Message = Messages.EncodeMsgSend("from", "to", "100", "uppyeo"),
            MessageTypeUrl = Messages.TypeMsgSend,
            GasLimit = 500_000
        };

        var result = await Pipeline.ExecuteTx(_client, _signer, _config, request);
        Assert.NotEmpty(result.TxHash);
    }

    [Fact]
    public async Task ExecuteTx_WithMemo()
    {
        await _client.ConnectAsync();
        var request = new TxRequest
        {
            Message = Messages.EncodeMsgSend("from", "to", "100", "uppyeo"),
            MessageTypeUrl = Messages.TypeMsgSend,
            Memo = "test transfer"
        };

        var result = await Pipeline.ExecuteTx(_client, _signer, _config, request);
        Assert.NotEmpty(result.TxHash);
    }

    [Fact]
    public async Task ExecuteTx_BroadcastFailure_Throws()
    {
        var failClient = new FailBroadcastClient();
        await failClient.ConnectAsync();

        var request = new TxRequest
        {
            Message = [0x01],
            MessageTypeUrl = Messages.TypeMsgSend
        };

        await Assert.ThrowsAsync<Errors.SdkException>(
            () => Pipeline.ExecuteTx(failClient, _signer, _config, request)
        );
    }

    [Fact]
    public async Task ExecuteTx_Timeout_Throws()
    {
        var neverConfirmClient = new NeverConfirmClient();
        await neverConfirmClient.ConnectAsync();

        var timeoutConfig = _config with
        {
            ConfirmTimeout = TimeSpan.FromMilliseconds(200),
            PollInterval = TimeSpan.FromMilliseconds(50)
        };

        var request = new TxRequest
        {
            Message = [0x01],
            MessageTypeUrl = Messages.TypeMsgSend
        };

        await Assert.ThrowsAsync<Errors.SdkException>(
            () => Pipeline.ExecuteTx(neverConfirmClient, _signer, timeoutConfig, request)
        );
    }

    [Fact]
    public void TxRequest_DefaultValues()
    {
        var request = new TxRequest
        {
            Message = [],
            MessageTypeUrl = "/test"
        };
        Assert.Equal(0UL, request.GasLimit);
        Assert.Equal("uppyeo", request.FeeDenom);
        Assert.Equal("", request.Memo);
    }

    [Fact]
    public void PipelineConfig_DefaultValues()
    {
        var config = new PipelineConfig { ChainId = "test" };
        Assert.Equal(250UL, config.GasPrice);
        Assert.Equal(TimeSpan.FromSeconds(30), config.ConfirmTimeout);
    }

    // Helper test clients

    private class FailBroadcastClient : MockChainClient
    {
        public new Task<BroadcastResponse> BroadcastTx(byte[] txBytes, string mode, CancellationToken ct = default)
        {
            return Task.FromResult(new BroadcastResponse { TxHash = "", Code = 5, RawLog = "broadcast error" });
        }
    }

    private class NeverConfirmClient : MockChainClient
    {
        public new Task<TxResponse?> GetTx(string txHash, CancellationToken ct = default)
        {
            return Task.FromResult<TxResponse?>(null);
        }
    }
}

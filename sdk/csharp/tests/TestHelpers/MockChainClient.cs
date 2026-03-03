using System.Text.Json;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Internal.Tx;

namespace Seocheon.Sdk.Tests.TestHelpers;

/// <summary>
/// In-memory mock chain client for testing.
/// </summary>
public class MockChainClient : IChainClient
{
    public bool IsConnected { get; private set; }
    public long CurrentBlockHeight { get; set; } = 17281; // Epoch 1, Window 0

    private readonly Dictionary<string, JsonElement> _queryResponses = new();
    private readonly Dictionary<string, TxResponse> _txResponses = new();

    public Task ConnectAsync(CancellationToken ct = default)
    {
        IsConnected = true;
        return Task.CompletedTask;
    }

    public Task DisconnectAsync()
    {
        IsConnected = false;
        return Task.CompletedTask;
    }

    public void SetQueryResponse(string path, object response)
    {
        var json = JsonSerializer.Serialize(response);
        _queryResponses[path] = JsonDocument.Parse(json).RootElement.Clone();
    }

    public void SetTxResponse(string txHash, TxResponse response)
    {
        _txResponses[txHash] = response;
    }

    public Task<JsonElement> QueryRest(string path, CancellationToken ct = default)
    {
        // Try exact match first, then prefix match
        if (_queryResponses.TryGetValue(path, out var exact))
            return Task.FromResult(exact);

        foreach (var (key, value) in _queryResponses)
        {
            if (path.StartsWith(key) || path.Contains(key.TrimStart('/')))
                return Task.FromResult(value);
        }

        // Default empty response
        var empty = JsonDocument.Parse("{}").RootElement.Clone();
        return Task.FromResult(empty);
    }

    public Task<BroadcastResponse> BroadcastTx(byte[] txBytes, string mode, CancellationToken ct = default)
    {
        return Task.FromResult(new BroadcastResponse
        {
            TxHash = "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234",
            Code = 0,
            RawLog = ""
        });
    }

    public Task<JsonElement> GetLatestBlock(CancellationToken ct = default)
    {
        var block = new
        {
            block = new
            {
                header = new
                {
                    height = CurrentBlockHeight.ToString(),
                    time = "2026-03-03T00:00:00Z",
                    chain_id = "seocheon-testnet-1"
                },
                data = new
                {
                    txs = Array.Empty<string>()
                }
            }
        };

        var json = JsonSerializer.Serialize(block);
        return Task.FromResult(JsonDocument.Parse(json).RootElement.Clone());
    }

    public Task<TxResponse?> GetTx(string txHash, CancellationToken ct = default)
    {
        if (_txResponses.TryGetValue(txHash, out var response))
            return Task.FromResult<TxResponse?>(response);

        // Default successful TX
        return Task.FromResult<TxResponse?>(new TxResponse
        {
            TxHash = txHash,
            Height = CurrentBlockHeight,
            Code = 0,
            RawLog = "",
            Events = []
        });
    }

    public Task<AccountInfo> GetAccountInfo(string address, CancellationToken ct = default)
    {
        return Task.FromResult(new AccountInfo
        {
            AccountNumber = 42,
            Sequence = 7
        });
    }
}

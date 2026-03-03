using System.Text;
using System.Text.Json;
using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Internal.Tx;

namespace Seocheon.Sdk.Infrastructure;

/// <summary>
/// Chain client interface for REST queries and TX broadcast.
/// </summary>
public interface IChainClient
{
    /// <summary>Establishes connection to the chain.</summary>
    Task ConnectAsync(CancellationToken ct = default);

    /// <summary>Closes the connection.</summary>
    Task DisconnectAsync();

    /// <summary>Returns whether the client is connected.</summary>
    bool IsConnected { get; }

    /// <summary>Queries a REST endpoint.</summary>
    Task<JsonElement> QueryRest(string path, CancellationToken ct = default);

    /// <summary>Broadcasts a signed transaction.</summary>
    Task<BroadcastResponse> BroadcastTx(byte[] txBytes, string mode, CancellationToken ct = default);

    /// <summary>Gets the latest block info.</summary>
    Task<JsonElement> GetLatestBlock(CancellationToken ct = default);

    /// <summary>Queries a transaction by hash.</summary>
    Task<TxResponse?> GetTx(string txHash, CancellationToken ct = default);

    /// <summary>Gets account info (number + sequence).</summary>
    Task<AccountInfo> GetAccountInfo(string address, CancellationToken ct = default);
}

/// <summary>
/// HTTP-based chain client implementation.
/// </summary>
public sealed class HttpChainClient : IChainClient, IDisposable
{
    private readonly string _rpcEndpoint;
    private readonly string _grpcEndpoint;
    private HttpClient? _http;
    private bool _connected;

    /// <summary>
    /// Creates an HttpChainClient for the given endpoints.
    /// </summary>
    public HttpChainClient(string rpcEndpoint, string grpcEndpoint)
    {
        _rpcEndpoint = rpcEndpoint.TrimEnd('/');
        _grpcEndpoint = grpcEndpoint.TrimEnd('/');
    }

    /// <inheritdoc />
    public bool IsConnected => _connected;

    /// <inheritdoc />
    public Task ConnectAsync(CancellationToken ct = default)
    {
        _http = new HttpClient { Timeout = TimeSpan.FromSeconds(30) };
        _connected = true;
        return Task.CompletedTask;
    }

    /// <inheritdoc />
    public Task DisconnectAsync()
    {
        _http?.Dispose();
        _http = null;
        _connected = false;
        return Task.CompletedTask;
    }

    /// <inheritdoc />
    public async Task<JsonElement> QueryRest(string path, CancellationToken ct = default)
    {
        EnsureConnected();
        try
        {
            var url = $"{_grpcEndpoint}{path}";
            var response = await _http!.GetAsync(url, ct);

            if (!response.IsSuccessStatusCode)
                throw SdkErrors.QueryFailed($"HTTP {response.StatusCode}: {path}");

            var body = await response.Content.ReadAsStringAsync(ct);
            return JsonDocument.Parse(body).RootElement.Clone();
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.QueryFailed($"{path}: {ex.Message}");
        }
    }

    /// <inheritdoc />
    public async Task<BroadcastResponse> BroadcastTx(byte[] txBytes, string mode, CancellationToken ct = default)
    {
        EnsureConnected();
        try
        {
            var payload = JsonSerializer.Serialize(new
            {
                tx_bytes = Convert.ToBase64String(txBytes),
                mode = mode == "async" ? "BROADCAST_MODE_ASYNC" : "BROADCAST_MODE_SYNC"
            });

            var content = new StringContent(payload, Encoding.UTF8, "application/json");
            var response = await _http!.PostAsync(
                $"{_grpcEndpoint}/cosmos/tx/v1beta1/txs",
                content,
                ct
            );

            var body = await response.Content.ReadAsStringAsync(ct);
            using var doc = JsonDocument.Parse(body);
            var txResponse = doc.RootElement.GetProperty("tx_response");

            return new BroadcastResponse
            {
                TxHash = txResponse.GetProperty("txhash").GetString() ?? "",
                Code = txResponse.GetProperty("code").GetUInt32(),
                RawLog = txResponse.TryGetProperty("raw_log", out var rl) ? rl.GetString() ?? "" : ""
            };
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.BroadcastFailed(ex.Message);
        }
    }

    /// <inheritdoc />
    public async Task<JsonElement> GetLatestBlock(CancellationToken ct = default)
    {
        return await QueryRest("/cosmos/base/tendermint/v1beta1/blocks/latest", ct);
    }

    /// <inheritdoc />
    public async Task<TxResponse?> GetTx(string txHash, CancellationToken ct = default)
    {
        EnsureConnected();
        try
        {
            var response = await _http!.GetAsync(
                $"{_grpcEndpoint}/cosmos/tx/v1beta1/txs/{txHash}",
                ct
            );

            if (!response.IsSuccessStatusCode)
                return null;

            var body = await response.Content.ReadAsStringAsync(ct);
            using var doc = JsonDocument.Parse(body);
            var txResp = doc.RootElement.GetProperty("tx_response");

            var events = new List<TxEventData>();
            if (txResp.TryGetProperty("events", out var eventsEl))
            {
                foreach (var evt in eventsEl.EnumerateArray())
                {
                    var attrs = new Dictionary<string, string>();
                    if (evt.TryGetProperty("attributes", out var attrsEl))
                    {
                        foreach (var attr in attrsEl.EnumerateArray())
                        {
                            var key = attr.GetProperty("key").GetString() ?? "";
                            var val = attr.TryGetProperty("value", out var v) ? v.GetString() ?? "" : "";
                            attrs[key] = val;
                        }
                    }
                    events.Add(new TxEventData(
                        evt.GetProperty("type").GetString() ?? "",
                        attrs
                    ));
                }
            }

            return new TxResponse
            {
                TxHash = txResp.GetProperty("txhash").GetString() ?? "",
                Height = long.Parse(txResp.GetProperty("height").GetString() ?? "0"),
                Code = txResp.GetProperty("code").GetUInt32(),
                RawLog = txResp.TryGetProperty("raw_log", out var rl) ? rl.GetString() ?? "" : "",
                Events = events
            };
        }
        catch (SdkException)
        {
            throw;
        }
        catch
        {
            return null;
        }
    }

    /// <inheritdoc />
    public async Task<AccountInfo> GetAccountInfo(string address, CancellationToken ct = default)
    {
        try
        {
            var result = await QueryRest($"/cosmos/auth/v1beta1/accounts/{address}", ct);
            var account = result.GetProperty("account");

            ulong accountNumber = 0;
            ulong sequence = 0;

            if (account.TryGetProperty("account_number", out var an))
                ulong.TryParse(an.GetString(), out accountNumber);
            if (account.TryGetProperty("sequence", out var seq))
                ulong.TryParse(seq.GetString(), out sequence);

            // Handle nested base_account (vesting accounts)
            if (account.TryGetProperty("base_account", out var baseAcct))
            {
                if (baseAcct.TryGetProperty("account_number", out var ban))
                    ulong.TryParse(ban.GetString(), out accountNumber);
                if (baseAcct.TryGetProperty("sequence", out var bseq))
                    ulong.TryParse(bseq.GetString(), out sequence);
            }

            return new AccountInfo { AccountNumber = accountNumber, Sequence = sequence };
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.AccountNotFound($"{address}: {ex.Message}");
        }
    }

    /// <summary>Disposes the HTTP client.</summary>
    public void Dispose()
    {
        _http?.Dispose();
    }

    private void EnsureConnected()
    {
        if (!_connected || _http == null)
            throw SdkErrors.NotConnected();
    }
}

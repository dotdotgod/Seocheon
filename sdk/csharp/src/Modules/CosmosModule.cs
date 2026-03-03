using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Infrastructure.Signing;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Types;
using Seocheon.Sdk.Utils;

namespace Seocheon.Sdk.Modules;

/// <summary>
/// Cosmos module: balance queries, token transfers, block and TX queries.
/// </summary>
public sealed class CosmosModule
{
    private readonly IChainClient _client;
    private readonly ISigningService _signer;
    private readonly PipelineConfig _pipelineConfig;

    internal CosmosModule(IChainClient client, ISigningService signer, PipelineConfig pipelineConfig)
    {
        _client = client;
        _signer = signer;
        _pipelineConfig = pipelineConfig;
    }

    /// <summary>
    /// Queries token balance for an address.
    /// </summary>
    /// <param name="address">Address to query. Empty uses signer's address.</param>
    /// <param name="denom">Denomination to query. Empty uses "uppyeo".</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<BalanceResponse> GetBalance(string address = "", string denom = "", CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(address))
            address = _signer.GetAddress();
        if (string.IsNullOrEmpty(denom))
            denom = "uppyeo";

        var result = await _client.QueryRest(
            $"/cosmos/bank/v1beta1/balances/{address}/by_denom?denom={denom}",
            ct
        );

        var balance = "0";
        if (result.TryGetProperty("balance", out var balEl))
        {
            if (balEl.TryGetProperty("amount", out var amt))
                balance = amt.GetString() ?? "0";
        }

        var balanceKkot = "0.0000000000";
        if (long.TryParse(balance, out var uppyeo))
            balanceKkot = ConvertUtils.FormatKkot(uppyeo);

        return new BalanceResponse(address, balance, balanceKkot);
    }

    /// <summary>
    /// Sends tokens to another address.
    /// </summary>
    /// <param name="toAddress">Recipient address.</param>
    /// <param name="amount">Amount in base units (string for large values).</param>
    /// <param name="denom">Denomination. Empty uses "uppyeo".</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<SendTokensResponse> SendTokens(string toAddress, string amount, string denom = "", CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(denom))
            denom = "uppyeo";

        var msgBytes = Messages.EncodeMsgSend(
            _signer.GetAddress(),
            toAddress,
            amount,
            denom
        );

        var result = await Pipeline.ExecuteTx(_client, _signer, _pipelineConfig, new TxRequest
        {
            Message = msgBytes,
            MessageTypeUrl = Messages.TypeMsgSend
        }, ct);

        return new SendTokensResponse(result.TxHash, result.Height);
    }

    /// <summary>
    /// Returns latest block information.
    /// </summary>
    public async Task<BlockInfoResponse> GetBlockInfo(CancellationToken ct = default)
    {
        var result = await _client.GetLatestBlock(ct);

        var block = result.GetProperty("block");
        var header = block.GetProperty("header");

        var height = long.Parse(header.GetProperty("height").GetString() ?? "0");
        var blockTime = header.TryGetProperty("time", out var t) ? t.GetString() ?? "" : "";
        var chainId = header.TryGetProperty("chain_id", out var ci) ? ci.GetString() ?? "" : "";

        ulong numTxs = 0;
        if (block.TryGetProperty("data", out var data) && data.TryGetProperty("txs", out var txs))
            numTxs = (ulong)txs.GetArrayLength();

        return new BlockInfoResponse(height, blockTime, chainId, numTxs);
    }

    /// <summary>
    /// Queries a transaction result by hash.
    /// </summary>
    public async Task<TxResultResponse> GetTxResult(string txHash, CancellationToken ct = default)
    {
        var txResp = await _client.GetTx(txHash, ct)
                     ?? throw Errors.SdkErrors.TxNotFound(txHash);

        var events = txResp.Events.Select(e => new TxEvent(
            e.Type,
            e.Attributes
        )).ToList();

        return new TxResultResponse(
            TxHash: txResp.TxHash,
            Height: txResp.Height,
            Code: txResp.Code,
            GasUsed: 0, // Not available from basic TX query
            GasWanted: 0,
            RawLog: txResp.RawLog,
            Events: events
        );
    }
}

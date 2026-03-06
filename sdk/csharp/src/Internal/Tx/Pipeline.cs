using System.Text.Json;
using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Infrastructure.Signing;

namespace Seocheon.Sdk.Internal.Tx;

/// <summary>
/// 4-phase transaction pipeline: Build → Sign → Broadcast → Confirm.
/// </summary>
public static class Pipeline
{
    /// <summary>
    /// Executes a full transaction lifecycle.
    /// </summary>
    public static async Task<TxResult> ExecuteTx(
        IChainClient client,
        ISigningService signer,
        PipelineConfig config,
        TxRequest request,
        CancellationToken ct = default)
    {
        // Phase 1: Build
        var accountInfo = await client.GetAccountInfo(signer.GetAddress(), ct);
        var gasLimit = Gas.ResolveGasLimit(request.MessageTypeUrl, request.GasLimit);
        var feeAmount = request.FeeAmount > 0
            ? request.FeeAmount
            : Gas.CalculateFee(gasLimit, config.GasPrice);
        var feeDenom = string.IsNullOrEmpty(request.FeeDenom) ? "uppyeo" : request.FeeDenom;

        var messageAny = Envelope.EncodeAny(request.MessageTypeUrl, request.Message);
        var bodyBytes = Envelope.EncodeTxBody(messageAny, request.Memo, request.TimeoutHeight);
        var authInfoBytes = Envelope.EncodeAuthInfo(
            signer.GetPubKey(),
            accountInfo.Sequence,
            gasLimit,
            feeAmount,
            feeDenom
        );

        // Phase 2: Sign
        byte[] signature;
        try
        {
            var signDoc = Envelope.EncodeSignDoc(
                bodyBytes,
                authInfoBytes,
                config.ChainId,
                accountInfo.AccountNumber
            );
            signature = signer.Sign(signDoc);
        }
        catch (SdkException)
        {
            throw;
        }
        catch (Exception ex)
        {
            throw SdkErrors.SigningFailed(ex.Message, ex);
        }

        // Phase 3: Broadcast
        var txRaw = Envelope.EncodeTxRaw(bodyBytes, authInfoBytes, signature);
        var broadcastResult = await client.BroadcastTx(txRaw, "sync", ct);

        var txHash = broadcastResult.TxHash;
        if (broadcastResult.Code != 0)
            throw SdkErrors.BroadcastFailed($"code={broadcastResult.Code}: {broadcastResult.RawLog}");

        // Phase 4: Confirm
        return await ConfirmTx(client, txHash, config.ConfirmTimeout, config.PollInterval, ct);
    }

    private static async Task<TxResult> ConfirmTx(
        IChainClient client,
        string txHash,
        TimeSpan timeout,
        TimeSpan pollInterval,
        CancellationToken ct)
    {
        using var cts = CancellationTokenSource.CreateLinkedTokenSource(ct);
        cts.CancelAfter(timeout);

        while (!cts.IsCancellationRequested)
        {
            try
            {
                var txResponse = await client.GetTx(txHash, cts.Token);
                if (txResponse != null)
                {
                    if (txResponse.Code != 0)
                        throw SdkErrors.AbciCodeToError(txResponse.Code, txResponse.RawLog);

                    return new TxResult
                    {
                        TxHash = txHash,
                        Height = txResponse.Height,
                        Code = txResponse.Code,
                        GasUsed = txResponse.GasUsed,
                        GasWanted = txResponse.GasWanted,
                        RawLog = txResponse.RawLog,
                        Events = txResponse.Events
                    };
                }
            }
            catch (SdkException)
            {
                throw;
            }
            catch
            {
                // TX not yet indexed, continue polling
            }

            try
            {
                await Task.Delay(pollInterval, cts.Token);
            }
            catch (OperationCanceledException)
            {
                break;
            }
        }

        throw SdkErrors.TxTimeout(txHash);
    }
}

/// <summary>
/// Broadcast response from the chain.
/// </summary>
public sealed record BroadcastResponse
{
    public string TxHash { get; init; } = "";
    public uint Code { get; init; }
    public string RawLog { get; init; } = "";
}

/// <summary>
/// Transaction response from GetTx query.
/// </summary>
public sealed record TxResponse
{
    public string TxHash { get; init; } = "";
    public long Height { get; init; }
    public uint Code { get; init; }
    public ulong GasUsed { get; init; }
    public ulong GasWanted { get; init; }
    public string RawLog { get; init; } = "";
    public IReadOnlyList<TxEventData> Events { get; init; } = [];
}

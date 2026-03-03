using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Infrastructure.Signing;
using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Types;
using Seocheon.Sdk.Utils;

namespace Seocheon.Sdk.Modules;

/// <summary>
/// Epoch module: query epoch/window state and qualification status.
/// </summary>
public sealed class EpochModule
{
    private readonly IChainClient _client;
    private readonly ISigningService _signer;

    internal EpochModule(IChainClient client, ISigningService signer)
    {
        _client = client;
        _signer = signer;
    }

    /// <summary>
    /// Returns current epoch and window information.
    /// </summary>
    public async Task<EpochInfoResponse> GetInfo(CancellationToken ct = default)
    {
        var block = await _client.GetLatestBlock(ct);
        var height = long.Parse(
            block.GetProperty("block").GetProperty("header").GetProperty("height").GetString() ?? "1"
        );

        var epoch = EpochUtils.ComputeEpoch(height);
        var window = EpochUtils.ComputeWindow(height);
        var epochStart = EpochUtils.EpochStartBlock(epoch);
        var epochEnd = EpochUtils.EpochEndBlock(epoch);
        var windowStart = EpochUtils.WindowStartBlock(epochStart, window);
        var windowEnd = EpochUtils.WindowEndBlock(windowStart);

        var epochElapsed = height - epochStart + 1;
        var epochTotal = ChainConstants.EpochBlocks;
        var windowElapsed = height - windowStart + 1;
        var windowTotal = ChainConstants.WindowBlocks;

        return new EpochInfoResponse(
            BlockHeight: height,
            EpochNumber: epoch,
            EpochStartBlock: epochStart,
            EpochEndBlock: epochEnd,
            EpochProgress: $"{epochElapsed}/{epochTotal}",
            WindowNumber: window,
            WindowStartBlock: windowStart,
            WindowEndBlock: windowEnd,
            WindowProgress: $"{windowElapsed}/{windowTotal}",
            BlocksUntilNextWindow: EpochUtils.BlocksUntilNextWindow(height),
            BlocksUntilNextEpoch: EpochUtils.BlocksUntilNextEpoch(height)
        );
    }

    /// <summary>
    /// Returns qualification status for a node in an epoch.
    /// </summary>
    /// <param name="nodeId">Node ID. Empty resolves from signer's agent address.</param>
    /// <param name="epochNumber">Epoch number. 0 means current epoch.</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<QualificationResponse> GetQualification(string nodeId = "", long epochNumber = 0, CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(nodeId))
            nodeId = await ResolveOwnNodeId(ct);

        if (epochNumber == 0)
            epochNumber = await ComputeCurrentEpoch(ct);

        var result = await _client.QueryRest(
            $"/seocheon/activity/v1/qualification/{nodeId}?epoch={epochNumber}",
            ct
        );

        var totalWindows = ChainConstants.WindowsPerEpoch;
        var requiredWindows = ChainConstants.MinActiveWindows;

        long elapsedWindows = 0;
        if (result.TryGetProperty("elapsed_windows", out var ew))
            long.TryParse(ew.GetString(), out elapsedWindows);

        ulong activeWindows = 0;
        if (result.TryGetProperty("active_windows", out var aw))
            ulong.TryParse(aw.GetString(), out activeWindows);

        var isQualified = activeWindows >= (ulong)requiredWindows;
        var remainingNeeded = Math.Max(0, requiredWindows - (long)activeWindows);
        var windowsLeft = totalWindows - elapsedWindows;
        var canStillQualify = isQualified || remainingNeeded <= windowsLeft;

        var windowDetail = new List<WindowActivity>();
        if (result.TryGetProperty("window_detail", out var wd))
        {
            foreach (var w in wd.EnumerateArray())
            {
                var wn = w.TryGetProperty("window_number", out var wnum)
                    ? long.Parse(wnum.GetString() ?? "0") : 0;
                ulong sc = 0;
                if (w.TryGetProperty("submission_count", out var scv))
                    ulong.TryParse(scv.GetString(), out sc);
                windowDetail.Add(new WindowActivity(wn, sc, sc > 0));
            }
        }

        return new QualificationResponse(
            EpochNumber: epochNumber,
            TotalWindows: totalWindows,
            ElapsedWindows: elapsedWindows,
            ActiveWindows: activeWindows,
            RequiredWindows: requiredWindows,
            IsQualified: isQualified,
            RemainingNeeded: remainingNeeded,
            CanStillQualify: canStillQualify,
            WindowDetail: windowDetail
        );
    }

    private async Task<string> ResolveOwnNodeId(CancellationToken ct)
    {
        var address = _signer.GetAddress();
        var result = await _client.QueryRest($"/seocheon/node/v1/node_by_agent/{address}", ct);
        return result.GetProperty("node").GetProperty("node_id").GetString()
               ?? throw SdkErrors.QueryFailed("Could not resolve node ID");
    }

    private async Task<long> ComputeCurrentEpoch(CancellationToken ct)
    {
        var block = await _client.GetLatestBlock(ct);
        var height = long.Parse(
            block.GetProperty("block").GetProperty("header").GetProperty("height").GetString() ?? "1"
        );
        return EpochUtils.ComputeEpoch(height);
    }
}

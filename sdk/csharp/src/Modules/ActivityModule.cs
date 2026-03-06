using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Infrastructure.Signing;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Types;
using Seocheon.Sdk.Utils;

namespace Seocheon.Sdk.Modules;

/// <summary>
/// Activity module: submit activities, query activities and quota.
/// </summary>
public sealed class ActivityModule
{
    private readonly IChainClient _client;
    private readonly ISigningService _signer;
    private readonly PipelineConfig _pipelineConfig;

    internal ActivityModule(IChainClient client, ISigningService signer, PipelineConfig pipelineConfig)
    {
        _client = client;
        _signer = signer;
        _pipelineConfig = pipelineConfig;
    }

    /// <summary>
    /// Submits an activity to the chain.
    /// </summary>
    /// <param name="activityHash">64-character lowercase hex SHA-256 hash.</param>
    /// <param name="contentUri">URI pointing to the off-chain activity report.</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<SubmitActivityResponse> Submit(string activityHash, string contentUri, CancellationToken ct = default)
    {
        if (!HashUtils.ValidateActivityHash(activityHash))
            throw SdkErrors.InvalidHash($"Must be {ChainConstants.ActivityHashLength} hex chars");
        if (string.IsNullOrWhiteSpace(contentUri))
            throw SdkErrors.InvalidContentUri("Content URI is required");

        var msgBytes = Messages.EncodeMsgSubmitActivity(
            _signer.GetAddress(),
            activityHash,
            contentUri
        );

        var result = await Pipeline.ExecuteTx(_client, _signer, _pipelineConfig, new TxRequest
        {
            Message = msgBytes,
            MessageTypeUrl = Messages.TypeMsgSubmitActivity
        }, ct);

        var epoch = EpochUtils.ComputeEpoch(result.Height);
        var window = EpochUtils.ComputeWindow(result.Height);

        return new SubmitActivityResponse(
            TxHash: result.TxHash,
            BlockHeight: result.Height,
            WindowNumber: window,
            EpochNumber: epoch,
            QuotaRemaining: 0 // Estimated; chain doesn't return in TX response
        );
    }

    /// <summary>
    /// Queries activity records for a node in a given epoch.
    /// </summary>
    /// <param name="nodeId">Node ID. Empty resolves from signer's agent address.</param>
    /// <param name="epochNumber">Epoch number. 0 means current epoch.</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<GetActivitiesResponse> GetActivities(string nodeId = "", long epochNumber = 0, CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(nodeId))
            nodeId = await ResolveOwnNodeId(ct);

        if (epochNumber == 0)
            epochNumber = await ComputeCurrentEpoch(ct);

        var result = await _client.QueryRest(
            $"/seocheon/activity/v1/nodes/{nodeId}/activities?epoch={epochNumber}",
            ct
        );

        var activities = new List<ActivityItem>();
        if (result.TryGetProperty("activities", out var arr))
        {
            foreach (var item in arr.EnumerateArray())
            {
                var height = long.Parse(item.GetProperty("block_height").GetString() ?? "0");
                activities.Add(new ActivityItem(
                    ActivityHash: item.GetProperty("activity_hash").GetString() ?? "",
                    ContentUri: item.TryGetProperty("content_uri", out var cu) ? cu.GetString() ?? "" : "",
                    BlockHeight: height,
                    WindowNumber: EpochUtils.ComputeWindow(height),
                    TxHash: item.TryGetProperty("tx_hash", out var tx) ? tx.GetString() ?? "" : ""
                ));
            }
        }

        ulong total = 0;
        if (result.TryGetProperty("total_count", out var tc))
            ulong.TryParse(tc.GetString(), out total);

        return new GetActivitiesResponse(activities, total);
    }

    /// <summary>
    /// Queries quota status for the current epoch.
    /// </summary>
    public async Task<GetQuotaResponse> GetQuota(CancellationToken ct = default)
    {
        var nodeId = await ResolveOwnNodeId(ct);
        var epochBlock = await _client.GetLatestBlock(ct);
        var height = long.Parse(
            epochBlock.GetProperty("block").GetProperty("header").GetProperty("height").GetString() ?? "1"
        );
        var epoch = EpochUtils.ComputeEpoch(height);

        var result = await _client.QueryRest(
            $"/seocheon/activity/v1/nodes/{nodeId}/epochs/{epoch}",
            ct
        );

        ulong limit = 0, used = 0;
        if (result.TryGetProperty("quota_limit", out var ql)) ulong.TryParse(ql.GetString(), out limit);
        if (result.TryGetProperty("quota_used", out var qu)) ulong.TryParse(qu.GetString(), out used);
        var remaining = limit >= used ? limit - used : 0UL;

        return new GetQuotaResponse(epoch, limit, used, remaining, false, null);
    }

    private async Task<string> ResolveOwnNodeId(CancellationToken ct)
    {
        var address = _signer.GetAddress();
        var result = await _client.QueryRest(
            $"/seocheon/node/v1/nodes/by-agent/{address}",
            ct
        );
        return result.GetProperty("node").GetProperty("id").GetString()
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

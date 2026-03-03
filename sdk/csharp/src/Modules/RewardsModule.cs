using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Infrastructure.Signing;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Types;

namespace Seocheon.Sdk.Modules;

/// <summary>
/// Rewards module: query pending rewards and withdraw.
/// </summary>
public sealed class RewardsModule
{
    private readonly IChainClient _client;
    private readonly ISigningService _signer;
    private readonly PipelineConfig _pipelineConfig;

    internal RewardsModule(IChainClient client, ISigningService signer, PipelineConfig pipelineConfig)
    {
        _client = client;
        _signer = signer;
        _pipelineConfig = pipelineConfig;
    }

    /// <summary>
    /// Queries pending rewards for a node.
    /// </summary>
    /// <param name="nodeId">Node ID. Empty resolves from signer's agent address.</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<PendingRewardsResponse> GetPending(string nodeId = "", CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(nodeId))
            nodeId = await ResolveOwnNodeId(ct);

        var result = await _client.QueryRest($"/seocheon/node/v1/rewards/{nodeId}", ct);

        var delegationReward = result.TryGetProperty("delegation_reward", out var dr) ? dr.GetString() ?? "0" : "0";
        var activityReward = result.TryGetProperty("activity_reward", out var ar) ? ar.GetString() ?? "0" : "0";
        var totalReward = result.TryGetProperty("total_reward", out var tr) ? tr.GetString() ?? "0" : "0";
        var commissionTotal = result.TryGetProperty("commission_total", out var ct2) ? ct2.GetString() ?? "0" : "0";

        // 80/20 split estimate
        if (long.TryParse(commissionTotal, out var commTotal) && commTotal > 0)
        {
            var operatorShare = (commTotal * 80 / 100).ToString();
            var agentShare = (commTotal - commTotal * 80 / 100).ToString();
            return new PendingRewardsResponse(delegationReward, activityReward, totalReward, commissionTotal, operatorShare, agentShare);
        }

        return new PendingRewardsResponse(delegationReward, activityReward, totalReward, commissionTotal, "0", "0");
    }

    /// <summary>
    /// Withdraws pending rewards.
    /// </summary>
    public async Task<WithdrawRewardsResponse> Withdraw(CancellationToken ct = default)
    {
        var msgBytes = Messages.EncodeMsgWithdrawRewards(_signer.GetAddress());

        var result = await Pipeline.ExecuteTx(_client, _signer, _pipelineConfig, new TxRequest
        {
            Message = msgBytes,
            MessageTypeUrl = Messages.TypeMsgWithdrawRewards
        }, ct);

        // Parse withdrawn amounts from events
        var withdrawnTotal = "0";
        var toOperator = "0";
        var toAgent = "0";

        foreach (var evt in result.Events)
        {
            if (evt.Type == "withdraw_commission")
            {
                if (evt.Attributes.TryGetValue("amount", out var amount))
                {
                    withdrawnTotal = amount;
                    // Estimate 80/20 split
                    if (long.TryParse(amount.Replace("uppyeo", ""), out var total))
                    {
                        toOperator = (total * 80 / 100).ToString();
                        toAgent = (total - total * 80 / 100).ToString();
                    }
                }
            }
        }

        return new WithdrawRewardsResponse(result.TxHash, withdrawnTotal, toOperator, toAgent);
    }

    private async Task<string> ResolveOwnNodeId(CancellationToken ct)
    {
        var address = _signer.GetAddress();
        var result = await _client.QueryRest($"/seocheon/node/v1/node_by_agent/{address}", ct);
        return result.GetProperty("node").GetProperty("node_id").GetString()
               ?? throw SdkErrors.QueryFailed("Could not resolve node ID");
    }
}

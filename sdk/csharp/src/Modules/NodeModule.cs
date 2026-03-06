using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Infrastructure.Signing;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Types;

namespace Seocheon.Sdk.Modules;

/// <summary>
/// Node module: query node information and search.
/// </summary>
public sealed class NodeModule
{
    private readonly IChainClient _client;
    private readonly ISigningService _signer;
    private readonly PipelineConfig _pipelineConfig;

    internal NodeModule(IChainClient client, ISigningService signer, PipelineConfig? pipelineConfig = null)
    {
        _client = client;
        _signer = signer;
        _pipelineConfig = pipelineConfig ?? new PipelineConfig { ChainId = "seocheon-1" };
    }

    /// <summary>
    /// Returns detailed information for a specific node.
    /// </summary>
    /// <param name="nodeId">Node ID. Empty resolves from signer's agent address.</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<NodeInfoResponse> GetInfo(string nodeId = "", CancellationToken ct = default)
    {
        if (string.IsNullOrEmpty(nodeId))
            nodeId = await ResolveOwnNodeId(ct);

        var result = await _client.QueryRest($"/seocheon/node/v1/nodes/{nodeId}", ct);
        var node = result.GetProperty("node");

        var tags = new List<string>();
        if (node.TryGetProperty("tags", out var tagsEl))
        {
            foreach (var t in tagsEl.EnumerateArray())
                tags.Add(t.GetString() ?? "");
        }

        return new NodeInfoResponse(
            NodeId: node.GetProperty("id").GetString() ?? "",
            Operator: node.TryGetProperty("operator", out var op) ? op.GetString() ?? "" : "",
            AgentAddress: node.TryGetProperty("agent_address", out var aa) ? aa.GetString() ?? "" : "",
            Status: node.TryGetProperty("status", out var st) ? st.GetString() ?? "REGISTERED" : "REGISTERED",
            Description: node.TryGetProperty("description", out var desc) ? desc.GetString() ?? "" : "",
            Website: node.TryGetProperty("website", out var ws) ? ws.GetString() ?? "" : "",
            Tags: tags,
            CommissionRate: node.TryGetProperty("commission_rate", out var cr) ? cr.GetString() ?? "0" : "0",
            AgentShare: node.TryGetProperty("agent_share", out var @as) ? @as.GetString() ?? "0" : "0",
            TotalDelegation: node.TryGetProperty("total_delegation", out var td) ? td.GetString() ?? "0" : "0",
            SelfDelegation: node.TryGetProperty("self_delegation", out var sd) ? sd.GetString() ?? "0" : "0",
            ValidatorAddress: node.TryGetProperty("validator_address", out var va) ? va.GetString() ?? "" : "",
            RegisteredAt: node.TryGetProperty("registered_at", out var ra)
                ? long.Parse(ra.GetString() ?? "0") : 0
        );
    }

    /// <summary>
    /// Searches for nodes with optional filters.
    /// </summary>
    /// <param name="tag">Filter by tag. Empty for all.</param>
    /// <param name="status">Filter by status. Empty for all.</param>
    /// <param name="limit">Maximum results (default 20).</param>
    /// <param name="orderBy">"delegation" (default) or "registered_at".</param>
    /// <param name="ct">Cancellation token.</param>
    public async Task<NodeSearchResponse> Search(
        string tag = "",
        string status = "",
        int limit = 20,
        string orderBy = "delegation",
        CancellationToken ct = default)
    {
        var query = $"/seocheon/node/v1/nodes?limit={limit}&order_by={orderBy}";
        if (!string.IsNullOrEmpty(tag))
            query += $"&tag={Uri.EscapeDataString(tag)}";
        if (!string.IsNullOrEmpty(status))
            query += $"&status={Uri.EscapeDataString(status)}";

        var result = await _client.QueryRest(query, ct);

        var nodes = new List<NodeSummary>();
        if (result.TryGetProperty("nodes", out var nodesEl))
        {
            foreach (var n in nodesEl.EnumerateArray())
            {
                var tags = new List<string>();
                if (n.TryGetProperty("tags", out var tagsEl))
                {
                    foreach (var t in tagsEl.EnumerateArray())
                        tags.Add(t.GetString() ?? "");
                }

                nodes.Add(new NodeSummary(
                    NodeId: n.GetProperty("id").GetString() ?? "",
                    Status: n.TryGetProperty("status", out var st) ? st.GetString() ?? "" : "",
                    Tags: tags,
                    TotalDelegation: n.TryGetProperty("total_delegation", out var td) ? td.GetString() ?? "0" : "0",
                    Description: n.TryGetProperty("description", out var desc) ? desc.GetString() ?? "" : ""
                ));
            }
        }

        ulong total = 0;
        if (result.TryGetProperty("total_count", out var tc))
            ulong.TryParse(tc.GetString(), out total);

        return new NodeSearchResponse(nodes, total);
    }

    /// <summary>
    /// Queries delegation confirmation status.
    /// </summary>
    public async Task<DelegationStatusResponse> GetDelegationStatus(
        string delegatorAddress, string validatorAddress, CancellationToken ct = default)
    {
        var result = await _client.QueryRest(
            $"/seocheon/node/v1/delegation-confirmation/{delegatorAddress}/{validatorAddress}", ct);

        return new DelegationStatusResponse(
            ExpiryEpoch: result.TryGetProperty("expiry_epoch", out var ee) ? long.Parse(ee.GetString() ?? "0") : 0,
            CurrentEpoch: result.TryGetProperty("current_epoch", out var ce) ? long.Parse(ce.GetString() ?? "0") : 0,
            InRenewalWindow: result.TryGetProperty("in_renewal_window", out var rw) && rw.GetBoolean(),
            RenewalWindowStart: result.TryGetProperty("renewal_window_start", out var rws) ? long.Parse(rws.GetString() ?? "0") : 0
        );
    }

    /// <summary>
    /// Confirms delegation for a validator.
    /// </summary>
    public async Task<TxResultResponse> ConfirmDelegation(string validatorAddress, CancellationToken ct = default)
    {
        var msgBytes = Messages.EncodeMsgConfirmDelegation(
            _signer.GetAddress(),
            validatorAddress
        );

        var result = await Pipeline.ExecuteTx(_client, _signer, _pipelineConfig, new TxRequest
        {
            Message = msgBytes,
            MessageTypeUrl = Messages.TypeMsgConfirmDelegation
        }, ct);

        return new TxResultResponse(
            TxHash: result.TxHash,
            Height: result.Height,
            Code: result.Code,
            GasUsed: result.GasUsed,
            GasWanted: result.GasWanted,
            RawLog: result.RawLog,
            Events: Array.Empty<TxEvent>()
        );
    }

    private async Task<string> ResolveOwnNodeId(CancellationToken ct)
    {
        var address = _signer.GetAddress();
        var result = await _client.QueryRest($"/seocheon/node/v1/nodes/by-agent/{address}", ct);
        return result.GetProperty("node").GetProperty("id").GetString()
               ?? throw SdkErrors.QueryFailed("Could not resolve node ID");
    }
}

namespace Seocheon.Sdk.Types;

/// <summary>
/// On-chain node status.
/// </summary>
public enum NodeStatus
{
    /// <summary>Registered but not in Active Validator Set.</summary>
    Registered,

    /// <summary>In Active Validator Set.</summary>
    Active,

    /// <summary>Inactive (jailed or unbonded).</summary>
    Inactive
}

/// <summary>
/// Extension methods for NodeStatus.
/// </summary>
public static class NodeStatusExtensions
{
    /// <summary>
    /// Parses a status string from chain response to NodeStatus enum.
    /// </summary>
    public static NodeStatus ParseNodeStatus(string status)
    {
        return status.ToUpperInvariant() switch
        {
            "REGISTERED" => NodeStatus.Registered,
            "ACTIVE" => NodeStatus.Active,
            "INACTIVE" or "JAILED" => NodeStatus.Inactive,
            _ => NodeStatus.Registered,
        };
    }

    /// <summary>
    /// Converts NodeStatus to its on-chain string representation.
    /// </summary>
    public static string ToChainString(this NodeStatus status)
    {
        return status switch
        {
            NodeStatus.Registered => "REGISTERED",
            NodeStatus.Active => "ACTIVE",
            NodeStatus.Inactive => "INACTIVE",
            _ => "REGISTERED",
        };
    }
}

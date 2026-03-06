namespace Seocheon.Sdk.Types;

// === Activity ===

/// <summary>Result of submitting an activity.</summary>
public sealed record SubmitActivityResponse(
    string TxHash,
    long BlockHeight,
    long WindowNumber,
    long EpochNumber,
    ulong QuotaRemaining
);

/// <summary>Single activity record.</summary>
public sealed record ActivityItem(
    string ActivityHash,
    string ContentUri,
    long BlockHeight,
    long WindowNumber,
    string TxHash
);

/// <summary>Activity query response.</summary>
public sealed record GetActivitiesResponse(
    IReadOnlyList<ActivityItem> Activities,
    ulong TotalCount
);

/// <summary>Quota status for current epoch.</summary>
public sealed record GetQuotaResponse(
    long EpochNumber,
    ulong QuotaTotal,
    ulong QuotaUsed,
    ulong QuotaRemaining,
    bool IsFeegrant,
    long? FeegrantExpiry
);

// === Epoch ===

/// <summary>Current epoch and window information.</summary>
public sealed record EpochInfoResponse(
    long BlockHeight,
    long EpochNumber,
    long EpochStartBlock,
    long EpochEndBlock,
    string EpochProgress,
    long WindowNumber,
    long WindowStartBlock,
    long WindowEndBlock,
    string WindowProgress,
    long BlocksUntilNextWindow,
    long BlocksUntilNextEpoch
);

/// <summary>Activity per window.</summary>
public sealed record WindowActivity(
    long WindowNumber,
    ulong SubmissionCount,
    bool HasActivity
);

/// <summary>Qualification status for a node in an epoch.</summary>
public sealed record QualificationResponse(
    long EpochNumber,
    long TotalWindows,
    long ElapsedWindows,
    ulong ActiveWindows,
    long RequiredWindows,
    bool IsQualified,
    long RemainingNeeded,
    bool CanStillQualify,
    IReadOnlyList<WindowActivity> WindowDetail
);

// === Node ===

/// <summary>Detailed node information.</summary>
public sealed record NodeInfoResponse(
    string NodeId,
    string Operator,
    string AgentAddress,
    string Status,
    string Description,
    string Website,
    IReadOnlyList<string> Tags,
    string CommissionRate,
    string AgentShare,
    string TotalDelegation,
    string SelfDelegation,
    string ValidatorAddress,
    long RegisteredAt
);

/// <summary>Summarized node entry for search results.</summary>
public sealed record NodeSummary(
    string NodeId,
    string Status,
    IReadOnlyList<string> Tags,
    string TotalDelegation,
    string Description
);

/// <summary>Node search results.</summary>
public sealed record NodeSearchResponse(
    IReadOnlyList<NodeSummary> Nodes,
    ulong TotalCount
);

// === Rewards ===

/// <summary>Pending rewards breakdown.</summary>
public sealed record PendingRewardsResponse(
    string DelegationReward,
    string ActivityReward,
    string TotalReward,
    string CommissionTotal,
    string OperatorShare,
    string AgentShare
);

/// <summary>Withdrawal result.</summary>
public sealed record WithdrawRewardsResponse(
    string TxHash,
    string WithdrawnTotal,
    string ToOperator,
    string ToAgent
);

// === Cosmos ===

/// <summary>Token balance.</summary>
public sealed record BalanceResponse(
    string Address,
    string Balance,
    string BalanceKkot
);

/// <summary>Token send result.</summary>
public sealed record SendTokensResponse(
    string TxHash,
    long BlockHeight
);

/// <summary>Block information.</summary>
public sealed record BlockInfoResponse(
    long BlockHeight,
    string BlockTime,
    string ChainId,
    ulong NumTxs
);

/// <summary>Single transaction event.</summary>
public sealed record TxEvent(
    string Type,
    IReadOnlyDictionary<string, string> Attributes
);

/// <summary>Delegation confirmation status.</summary>
public sealed record DelegationStatusResponse(
    long ExpiryEpoch,
    long CurrentEpoch,
    bool InRenewalWindow,
    long RenewalWindowStart
);

/// <summary>Transaction result.</summary>
public sealed record TxResultResponse(
    string TxHash,
    long Height,
    uint Code,
    ulong GasUsed,
    ulong GasWanted,
    string RawLog,
    IReadOnlyList<TxEvent> Events
);

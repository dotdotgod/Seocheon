namespace Seocheon.Sdk.Internal.Tx;

/// <summary>
/// Transaction request for the pipeline.
/// </summary>
public sealed record TxRequest
{
    /// <summary>Protobuf-encoded message.</summary>
    public required byte[] Message { get; init; }

    /// <summary>Message type URL (e.g., "/seocheon.activity.v1.MsgSubmitActivity").</summary>
    public required string MessageTypeUrl { get; init; }

    /// <summary>Optional gas limit override. If 0, uses default for message type.</summary>
    public ulong GasLimit { get; init; }

    /// <summary>Optional fee amount override in uppyeo.</summary>
    public ulong FeeAmount { get; init; }

    /// <summary>Fee denomination (default: "uppyeo").</summary>
    public string FeeDenom { get; init; } = "uppyeo";

    /// <summary>Optional memo.</summary>
    public string Memo { get; init; } = "";

    /// <summary>Optional timeout block height.</summary>
    public ulong TimeoutHeight { get; init; }
}

/// <summary>
/// Transaction pipeline result.
/// </summary>
public sealed record TxResult
{
    /// <summary>Transaction hash.</summary>
    public required string TxHash { get; init; }

    /// <summary>Block height of inclusion.</summary>
    public long Height { get; init; }

    /// <summary>ABCI response code (0 = success).</summary>
    public uint Code { get; init; }

    /// <summary>Gas actually consumed.</summary>
    public ulong GasUsed { get; init; }

    /// <summary>Gas limit requested.</summary>
    public ulong GasWanted { get; init; }

    /// <summary>Raw log from the chain.</summary>
    public string RawLog { get; init; } = "";

    /// <summary>Transaction events.</summary>
    public IReadOnlyList<TxEventData> Events { get; init; } = [];
}

/// <summary>
/// Event data from a transaction.
/// </summary>
public sealed record TxEventData(
    string Type,
    IReadOnlyDictionary<string, string> Attributes
);

/// <summary>
/// Pipeline configuration.
/// </summary>
public sealed record PipelineConfig
{
    /// <summary>Chain identifier.</summary>
    public required string ChainId { get; init; }

    /// <summary>Gas price in uppyeo.</summary>
    public ulong GasPrice { get; init; } = 250;

    /// <summary>Confirmation timeout.</summary>
    public TimeSpan ConfirmTimeout { get; init; } = TimeSpan.FromSeconds(30);

    /// <summary>Poll interval for confirmation.</summary>
    public TimeSpan PollInterval { get; init; } = TimeSpan.FromSeconds(1);
}

/// <summary>
/// Account information from the chain.
/// </summary>
public sealed record AccountInfo
{
    /// <summary>Account number.</summary>
    public ulong AccountNumber { get; init; }

    /// <summary>Sequence number (nonce).</summary>
    public ulong Sequence { get; init; }
}

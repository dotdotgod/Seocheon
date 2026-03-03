using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Errors;

namespace Seocheon.Sdk;

/// <summary>
/// Signing mode for transaction authorization.
/// </summary>
public enum SigningMode
{
    /// <summary>External vault server (production).</summary>
    Vault,
    /// <summary>Encrypted local keystore.</summary>
    Keystore,
    /// <summary>Mnemonic-based, development only.</summary>
    Direct
}

/// <summary>
/// Chain connection configuration.
/// </summary>
public sealed record ChainConfig
{
    /// <summary>Blockchain identifier (e.g., "seocheon-1").</summary>
    public required string ChainId { get; init; }

    /// <summary>RPC HTTP endpoint.</summary>
    public required string RpcEndpoint { get; init; }

    /// <summary>gRPC/REST endpoint.</summary>
    public required string GrpcEndpoint { get; init; }

    /// <summary>Gas price per unit (default: "250uppyeo").</summary>
    public string GasPrice { get; init; } = ChainConstants.DefaultGasPriceStr;

    /// <summary>Gas limit multiplier (default: 1.3).</summary>
    public double GasAdjustment { get; init; } = ChainConstants.DefaultGasAdjustment;
}

/// <summary>
/// Signing service configuration.
/// </summary>
public sealed record SigningConfig
{
    /// <summary>Signing mode.</summary>
    public required SigningMode Mode { get; init; }

    /// <summary>Vault server endpoint (vault mode).</summary>
    public string VaultEndpoint { get; init; } = "";

    /// <summary>Key name in vault (vault mode).</summary>
    public string KeyName { get; init; } = "";

    /// <summary>Keystore file path (keystore mode).</summary>
    public string KeystorePath { get; init; } = "";

    /// <summary>Environment variable containing keystore passphrase.</summary>
    public string PassphraseEnv { get; init; } = "";

    /// <summary>BIP39 mnemonic (direct mode, development only).</summary>
    public string Mnemonic { get; init; } = "";
}

/// <summary>
/// Transaction broadcast and confirmation configuration.
/// </summary>
public sealed record TxConfig
{
    /// <summary>Broadcast mode: "sync" or "async".</summary>
    public string BroadcastMode { get; init; } = ChainConstants.DefaultBroadcastMode;

    /// <summary>Confirmation timeout in milliseconds.</summary>
    public int ConfirmTimeoutMs { get; init; } = ChainConstants.DefaultConfirmTimeoutMs;

    /// <summary>Confirmation poll interval in milliseconds.</summary>
    public int ConfirmPollIntervalMs { get; init; } = ChainConstants.DefaultConfirmPollMs;
}

/// <summary>
/// Top-level SDK configuration.
/// </summary>
public sealed record SdkConfig
{
    /// <summary>Chain connection settings.</summary>
    public required ChainConfig Chain { get; init; }

    /// <summary>Signing service settings.</summary>
    public required SigningConfig Signing { get; init; }

    /// <summary>Transaction settings.</summary>
    public TxConfig Tx { get; init; } = new();

    /// <summary>
    /// Validates the configuration and throws SdkException on invalid settings.
    /// </summary>
    public void Validate()
    {
        // Chain
        if (string.IsNullOrWhiteSpace(Chain.ChainId))
            throw SdkErrors.InvalidConfig("chain_id is required");
        if (string.IsNullOrWhiteSpace(Chain.RpcEndpoint))
            throw SdkErrors.InvalidConfig("rpc_endpoint is required");
        if (string.IsNullOrWhiteSpace(Chain.GrpcEndpoint))
            throw SdkErrors.InvalidConfig("grpc_endpoint is required");
        if (Chain.GasAdjustment <= 0)
            throw SdkErrors.InvalidConfig("gas_adjustment must be positive");

        // Signing
        switch (Signing.Mode)
        {
            case SigningMode.Vault:
                if (string.IsNullOrWhiteSpace(Signing.VaultEndpoint))
                    throw SdkErrors.InvalidConfig("vault_endpoint required for vault mode");
                if (string.IsNullOrWhiteSpace(Signing.KeyName))
                    throw SdkErrors.InvalidConfig("key_name required for vault mode");
                break;

            case SigningMode.Keystore:
                if (string.IsNullOrWhiteSpace(Signing.KeystorePath))
                    throw SdkErrors.InvalidConfig("keystore_path required for keystore mode");
                if (string.IsNullOrWhiteSpace(Signing.PassphraseEnv))
                    throw SdkErrors.InvalidConfig("passphrase_env required for keystore mode");
                break;

            case SigningMode.Direct:
                if (string.IsNullOrWhiteSpace(Signing.Mnemonic))
                    throw SdkErrors.InvalidConfig("mnemonic required for direct mode");
                break;

            default:
                throw SdkErrors.InvalidConfig($"Unknown signing mode: {Signing.Mode}");
        }

        // Tx
        if (Tx.BroadcastMode != "sync" && Tx.BroadcastMode != "async")
            throw SdkErrors.InvalidConfig("broadcast_mode must be 'sync' or 'async'");
        if (Tx.ConfirmTimeoutMs <= 0)
            throw SdkErrors.InvalidConfig("confirm_timeout_ms must be positive");
        if (Tx.ConfirmPollIntervalMs <= 0)
            throw SdkErrors.InvalidConfig("confirm_poll_interval_ms must be positive");
    }
}

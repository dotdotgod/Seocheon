using Seocheon.Sdk.Constants;
using Seocheon.Sdk.Errors;
using Seocheon.Sdk.Infrastructure;
using Seocheon.Sdk.Infrastructure.Signing;
using Seocheon.Sdk.Internal.Tx;
using Seocheon.Sdk.Modules;

namespace Seocheon.Sdk;

/// <summary>
/// Main Seocheon SDK client. Entry point for all blockchain operations.
/// </summary>
public sealed class SeocheonSdk : IDisposable
{
    private readonly SdkConfig _config;
    private readonly IChainClient _client;
    private readonly ISigningService _signer;
    private bool _connected;

    /// <summary>Activity submission and query.</summary>
    public ActivityModule Activity { get; }

    /// <summary>Epoch and window information.</summary>
    public EpochModule Epoch { get; }

    /// <summary>Node information and search.</summary>
    public NodeModule Node { get; }

    /// <summary>Rewards query and withdrawal.</summary>
    public RewardsModule Rewards { get; }

    /// <summary>Balance, transfer, block, and TX queries.</summary>
    public CosmosModule Cosmos { get; }

    /// <summary>
    /// Creates a new SDK instance with the given configuration.
    /// </summary>
    /// <param name="config">SDK configuration.</param>
    /// <param name="chainClient">Optional custom chain client (for testing).</param>
    /// <param name="signingService">Optional custom signing service (for testing).</param>
    public SeocheonSdk(SdkConfig config, IChainClient? chainClient = null, ISigningService? signingService = null)
    {
        config.Validate();
        _config = config;

        _client = chainClient ?? new HttpChainClient(config.Chain.RpcEndpoint, config.Chain.GrpcEndpoint);
        _signer = signingService ?? CreateSigner(config.Signing);

        var pipelineConfig = new PipelineConfig
        {
            ChainId = config.Chain.ChainId,
            GasPrice = ParseGasPrice(config.Chain.GasPrice),
            ConfirmTimeout = TimeSpan.FromMilliseconds(config.Tx.ConfirmTimeoutMs),
            PollInterval = TimeSpan.FromMilliseconds(config.Tx.ConfirmPollIntervalMs)
        };

        Activity = new ActivityModule(_client, _signer, pipelineConfig);
        Epoch = new EpochModule(_client, _signer);
        Node = new NodeModule(_client, _signer);
        Rewards = new RewardsModule(_client, _signer, pipelineConfig);
        Cosmos = new CosmosModule(_client, _signer, pipelineConfig);
    }

    /// <summary>
    /// Connects to the blockchain.
    /// </summary>
    public async Task ConnectAsync(CancellationToken ct = default)
    {
        if (_connected)
            return;

        await _client.ConnectAsync(ct);
        _connected = true;
    }

    /// <summary>
    /// Disconnects from the blockchain.
    /// </summary>
    public async Task DisconnectAsync()
    {
        if (!_connected)
            return;

        await _client.DisconnectAsync();
        _connected = false;
    }

    /// <summary>
    /// Returns whether the SDK is connected to the chain.
    /// </summary>
    public bool IsConnected => _connected;

    /// <summary>
    /// Returns the current SDK configuration.
    /// </summary>
    public SdkConfig Config => _config;

    /// <summary>
    /// Returns the signer's address.
    /// </summary>
    public string GetAddress() => _signer.GetAddress();

    /// <summary>
    /// Disposes the SDK and releases resources.
    /// </summary>
    public void Dispose()
    {
        DisconnectAsync().GetAwaiter().GetResult();
        if (_client is IDisposable disposable)
            disposable.Dispose();
    }

    private static ISigningService CreateSigner(SigningConfig config)
    {
        return config.Mode switch
        {
            SigningMode.Direct => new DirectService(config.Mnemonic),
            SigningMode.Keystore => new KeystoreService(
                config.KeystorePath,
                Environment.GetEnvironmentVariable(config.PassphraseEnv)
                    ?? throw SdkErrors.InvalidConfig($"Environment variable '{config.PassphraseEnv}' not set")
            ),
            SigningMode.Vault => new VaultService(config.VaultEndpoint, config.KeyName),
            _ => throw SdkErrors.InvalidConfig($"Unknown signing mode: {config.Mode}")
        };
    }

    private static ulong ParseGasPrice(string gasPriceStr)
    {
        // Parse "250uppyeo" → 250
        var numStr = new string(gasPriceStr.TakeWhile(char.IsDigit).ToArray());
        return ulong.TryParse(numStr, out var price) ? price : ChainConstants.DefaultGasPrice;
    }
}

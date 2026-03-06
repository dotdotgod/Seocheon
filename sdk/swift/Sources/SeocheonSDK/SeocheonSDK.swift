import Foundation

/// Main entry point for the Seocheon blockchain SDK.
public final class SeocheonSDK: @unchecked Sendable {
    private let config: SDKConfig
    private let client: ChainClient
    private let signer: SigningService
    private var _connected: Bool = false

    /// Activity submission and query operations.
    public let activity: ActivityModule
    /// Epoch and window information.
    public let epoch: EpochModule
    /// Node registration and search operations.
    public let node: NodeModule
    /// Reward query and withdrawal operations.
    public let rewards: RewardsModule
    /// Standard Cosmos operations (balance, send, block info).
    public let cosmos: CosmosModule

    /// Creates a new SDK instance with the given configuration.
    public init(config: SDKConfig) throws {
        try config.validate()

        self.config = config
        self.client = URLSessionChainClient(rpcEndpoint: config.chain.rpcEndpoint, grpcEndpoint: config.chain.grpcEndpoint)
        self.signer = try SeocheonSDK.createSigner(config.signing)

        self.activity = ActivityModule(client: client, signer: signer, chainID: config.chain.chainId)
        self.epoch = EpochModule(client: client)
        self.node = NodeModule(client: client, signer: signer, chainID: config.chain.chainId)
        self.rewards = RewardsModule(client: client, signer: signer, chainID: config.chain.chainId)
        self.cosmos = CosmosModule(client: client, signer: signer, chainID: config.chain.chainId)
    }

    /// Internal initializer for testing with custom client and signer.
    internal init(config: SDKConfig, client: ChainClient, signer: SigningService) {
        self.config = config
        self.client = client
        self.signer = signer

        self.activity = ActivityModule(client: client, signer: signer, chainID: config.chain.chainId)
        self.epoch = EpochModule(client: client)
        self.node = NodeModule(client: client, signer: signer, chainID: config.chain.chainId)
        self.rewards = RewardsModule(client: client, signer: signer, chainID: config.chain.chainId)
        self.cosmos = CosmosModule(client: client, signer: signer, chainID: config.chain.chainId)
    }

    /// Establishes a connection to the blockchain node.
    public func connect() async throws {
        try await client.connect()
        _connected = true
    }

    /// Closes the connection to the blockchain node.
    public func disconnect() async {
        _connected = false
        await client.disconnect()
    }

    /// Returns whether the SDK is connected to the chain.
    public func isConnected() -> Bool {
        return _connected
    }

    /// Returns the current SDK configuration.
    public func getConfig() -> SDKConfig {
        return config
    }

    private static func createSigner(_ cfg: SigningConfig) throws -> SigningService {
        switch cfg.mode {
        case .vault:
            guard let endpoint = cfg.vaultEndpoint, let keyName = cfg.keyName else {
                throw SDKError.invalidConfig("vault mode requires vault_endpoint and key_name")
            }
            return VaultSigningService(endpoint: endpoint, keyName: keyName)
        case .keystore:
            guard let path = cfg.keystorePath, let envVar = cfg.passphraseEnv else {
                throw SDKError.invalidConfig("keystore mode requires keystore_path and passphrase_env")
            }
            let passphrase = ProcessInfo.processInfo.environment[envVar] ?? ""
            return try KeystoreSigningService(keystorePath: path, passphrase: passphrase)
        case .direct:
            guard let mnemonic = cfg.mnemonic else {
                throw SDKError.invalidConfig("direct mode requires mnemonic")
            }
            return try DirectSigningService(mnemonic: mnemonic)
        }
    }
}

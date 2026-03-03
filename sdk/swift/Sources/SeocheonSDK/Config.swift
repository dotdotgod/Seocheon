import Foundation

/// Blockchain connection settings.
public struct ChainConfig: Codable, Sendable {
    public var chainId: String
    public var rpcEndpoint: String
    public var grpcEndpoint: String
    public var gasPrice: String
    public var gasAdjustment: Double

    public init(chainId: String, rpcEndpoint: String, grpcEndpoint: String,
                gasPrice: String = ChainConstants.defaultGasPrice,
                gasAdjustment: Double = ChainConstants.defaultGasAdjustment) {
        self.chainId = chainId
        self.rpcEndpoint = rpcEndpoint
        self.grpcEndpoint = grpcEndpoint
        self.gasPrice = gasPrice
        self.gasAdjustment = gasAdjustment
    }
}

/// How transactions are signed.
public enum SigningMode: String, Codable, Sendable {
    case vault
    case keystore
    case direct
}

/// Transaction signing settings.
public struct SigningConfig: Codable, Sendable {
    public var mode: SigningMode
    public var vaultEndpoint: String?
    public var keyName: String?
    public var keystorePath: String?
    public var passphraseEnv: String?
    public var mnemonic: String?

    public init(mode: SigningMode, vaultEndpoint: String? = nil, keyName: String? = nil,
                keystorePath: String? = nil, passphraseEnv: String? = nil, mnemonic: String? = nil) {
        self.mode = mode
        self.vaultEndpoint = vaultEndpoint
        self.keyName = keyName
        self.keystorePath = keystorePath
        self.passphraseEnv = passphraseEnv
        self.mnemonic = mnemonic
    }
}

/// Transaction broadcast settings.
public struct TxConfig: Codable, Sendable {
    public var broadcastMode: String
    public var confirmTimeoutMs: UInt64
    public var confirmPollIntervalMs: UInt64

    public init(broadcastMode: String = ChainConstants.defaultBroadcastMode,
                confirmTimeoutMs: UInt64 = ChainConstants.defaultConfirmTimeoutMs,
                confirmPollIntervalMs: UInt64 = ChainConstants.defaultConfirmPollMs) {
        self.broadcastMode = broadcastMode
        self.confirmTimeoutMs = confirmTimeoutMs
        self.confirmPollIntervalMs = confirmPollIntervalMs
    }
}

/// Top-level SDK configuration.
public struct SDKConfig: Codable, Sendable {
    public var chain: ChainConfig
    public var signing: SigningConfig
    public var tx: TxConfig

    public init(chain: ChainConfig, signing: SigningConfig, tx: TxConfig = TxConfig()) {
        self.chain = chain
        self.signing = signing
        self.tx = tx
    }

    /// Creates a config with default values applied.
    public static func defaultConfig(chainId: String, rpcEndpoint: String, grpcEndpoint: String, signing: SigningConfig) -> SDKConfig {
        return SDKConfig(
            chain: ChainConfig(chainId: chainId, rpcEndpoint: rpcEndpoint, grpcEndpoint: grpcEndpoint),
            signing: signing,
            tx: TxConfig()
        )
    }

    /// Validates the configuration.
    public func validate() throws {
        guard !chain.chainId.isEmpty else {
            throw SDKError.invalidConfig("chain_id is required")
        }
        guard !chain.rpcEndpoint.isEmpty else {
            throw SDKError.invalidConfig("rpc_endpoint is required")
        }
        guard !chain.grpcEndpoint.isEmpty else {
            throw SDKError.invalidConfig("grpc_endpoint is required")
        }
        switch signing.mode {
        case .vault:
            guard let ve = signing.vaultEndpoint, !ve.isEmpty,
                  let kn = signing.keyName, !kn.isEmpty else {
                throw SDKError.invalidConfig("vault mode requires vault_endpoint and key_name")
            }
        case .keystore:
            guard let kp = signing.keystorePath, !kp.isEmpty,
                  let pe = signing.passphraseEnv, !pe.isEmpty else {
                throw SDKError.invalidConfig("keystore mode requires keystore_path and passphrase_env")
            }
        case .direct:
            guard let m = signing.mnemonic, !m.isEmpty else {
                throw SDKError.invalidConfig("direct mode requires mnemonic")
            }
        }
    }
}

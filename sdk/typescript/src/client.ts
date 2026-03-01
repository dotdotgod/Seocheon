import type { SDKConfig, ResolvedTxConfig } from "./types/config.js";
import { ActivityModule } from "./modules/activity.js";
import { EpochModule } from "./modules/epoch.js";
import { NodeModule } from "./modules/node.js";
import { RewardsModule } from "./modules/rewards.js";
import { CosmosModule } from "./modules/cosmos.js";
import {
  HttpChainClient,
  type ChainClient,
} from "./infrastructure/chain-client.js";
import {
  DirectSigningService,
  VaultSigningService,
  KeystoreSigningService,
  type SigningService,
} from "./infrastructure/signing-service.js";
import { ConnectionError, ValidationError } from "./errors/errors.js";
import { ERR_NOT_CONNECTED, ERR_INVALID_CONFIG } from "./constants/errors.js";

export class SeocheonSDK {
  private readonly config: SDKConfig;
  private readonly chainClient: ChainClient;
  private readonly signingService: SigningService;
  private readonly txConfig: ResolvedTxConfig;
  private connected = false;

  readonly activity: ActivityModule;
  readonly epoch: EpochModule;
  readonly node: NodeModule;
  readonly rewards: RewardsModule;
  readonly cosmos: CosmosModule;

  constructor(config: SDKConfig) {
    this.validateConfig(config);
    this.config = config;

    this.txConfig = {
      broadcast_mode: config.tx?.broadcast_mode ?? "sync",
      confirm_timeout_ms: config.tx?.confirm_timeout_ms ?? 30000,
      confirm_poll_interval_ms: config.tx?.confirm_poll_interval_ms ?? 1000,
    };

    this.chainClient = new HttpChainClient();
    this.signingService = this.createSigningService(config);

    this.activity = new ActivityModule(
      this.chainClient,
      this.signingService,
      this.txConfig,
    );
    this.epoch = new EpochModule(this.chainClient, this.signingService);
    this.node = new NodeModule(this.chainClient, this.signingService);
    this.rewards = new RewardsModule(
      this.chainClient,
      this.signingService,
      this.txConfig,
    );
    this.cosmos = new CosmosModule(
      this.chainClient,
      this.signingService,
      this.txConfig,
    );
  }

  async connect(): Promise<void> {
    await this.chainClient.connect(
      this.config.chain.rpc_endpoint,
      this.config.chain.grpc_endpoint,
    );
    this.connected = true;
  }

  async disconnect(): Promise<void> {
    await this.chainClient.disconnect();
    this.connected = false;
  }

  isConnected(): boolean {
    return this.connected && this.chainClient.isConnected();
  }

  getConfig(): SDKConfig {
    return { ...this.config };
  }

  private validateConfig(config: SDKConfig): void {
    if (!config.chain) {
      throw new ValidationError("chain config is required", ERR_INVALID_CONFIG);
    }
    if (!config.chain.chain_id) {
      throw new ValidationError("chain_id is required", ERR_INVALID_CONFIG);
    }
    if (!config.chain.rpc_endpoint) {
      throw new ValidationError(
        "rpc_endpoint is required",
        ERR_INVALID_CONFIG,
      );
    }
    if (!config.chain.grpc_endpoint) {
      throw new ValidationError(
        "grpc_endpoint is required",
        ERR_INVALID_CONFIG,
      );
    }
    if (!config.signing) {
      throw new ValidationError(
        "signing config is required",
        ERR_INVALID_CONFIG,
      );
    }
    if (!["vault", "keystore", "direct"].includes(config.signing.mode)) {
      throw new ValidationError(
        `invalid signing mode: ${config.signing.mode}`,
        ERR_INVALID_CONFIG,
      );
    }
  }

  private createSigningService(config: SDKConfig): SigningService {
    switch (config.signing.mode) {
      case "direct":
        if (!config.signing.mnemonic) {
          throw new ValidationError(
            "mnemonic is required for direct signing mode",
            ERR_INVALID_CONFIG,
          );
        }
        return new DirectSigningService(config.signing.mnemonic);
      case "vault":
        if (!config.signing.vault_endpoint || !config.signing.key_name) {
          throw new ValidationError(
            "vault_endpoint and key_name are required for vault signing mode",
            ERR_INVALID_CONFIG,
          );
        }
        return new VaultSigningService(
          config.signing.vault_endpoint,
          config.signing.key_name,
        );
      case "keystore": {
        if (
          !config.signing.keystore_path ||
          !config.signing.passphrase_env
        ) {
          throw new ValidationError(
            "keystore_path and passphrase_env are required for keystore signing mode",
            ERR_INVALID_CONFIG,
          );
        }
        const passphrase =
          process.env[config.signing.passphrase_env] ?? "";
        return new KeystoreSigningService(
          config.signing.keystore_path,
          passphrase,
        );
      }
      default:
        throw new ValidationError(
          `unsupported signing mode: ${config.signing.mode}`,
          ERR_INVALID_CONFIG,
        );
    }
  }
}

// Seocheon Client SDK for TypeScript
// Entry point: re-exports all public API

export { SeocheonSDK } from "./client.js";

// Types
export type {
  SDKConfig,
  ChainConfig,
  SigningConfig,
  TxConfig,
  SigningMode,
  BroadcastMode,
} from "./types/config.js";

export type {
  SubmitActivityResponse,
  ActivityItem,
  GetActivitiesResponse,
  GetQuotaResponse,
  EpochInfoResponse,
  QualificationResponse,
  WindowActivity,
  NodeInfoResponse,
  NodeSummary,
  NodeSearchResponse,
  NodeStatus,
  PendingRewardsResponse,
  WithdrawRewardsResponse,
  BalanceResponse,
  SendTokensResponse,
  BlockInfoResponse,
  TxResultResponse,
  TxEvent,
  EventAttribute,
} from "./types/responses.js";

// Errors
export {
  SeocheonError,
  ConnectionError,
  QueryError,
  TransactionError,
  ValidationError,
  TimeoutError,
} from "./errors/errors.js";

// Constants
export * from "./constants/chain.js";
export * from "./constants/errors.js";

// Utilities
export {
  convertDenom,
  uppyeoToKkot,
  kkotToUppyeo,
  formatKkot,
  type Denom,
} from "./utils/denom.js";
export {
  verifyActivityHash,
  isValidActivityHash,
  isValidContentUri,
} from "./utils/hash.js";
export {
  computeEpoch,
  computeWindow,
  epochStartBlock,
  epochEndBlock,
  windowStartBlock,
  windowEndBlock,
  formatProgress,
} from "./utils/epoch.js";

// Modules (for advanced usage)
export { ActivityModule } from "./modules/activity.js";
export { EpochModule } from "./modules/epoch.js";
export { NodeModule } from "./modules/node.js";
export { RewardsModule } from "./modules/rewards.js";
export { CosmosModule } from "./modules/cosmos.js";

// Infrastructure (for custom implementations)
export type { SigningService } from "./infrastructure/signing-service.js";
export {
  DirectSigningService,
  VaultSigningService,
  KeystoreSigningService,
} from "./infrastructure/signing-service.js";
export type { ChainClient } from "./infrastructure/chain-client.js";
export { HttpChainClient } from "./infrastructure/chain-client.js";

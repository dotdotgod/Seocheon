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

// Protobuf encoding (for advanced usage)
export {
  encodeVarint,
  encodeFieldVarint,
  encodeFieldBytes,
  encodeFieldString,
  concatBytes,
} from "./infrastructure/protobuf.js";

// Message encoders
export type { MessageEncoder } from "./infrastructure/messages.js";
export {
  MsgSubmitActivity,
  MsgWithdrawNodeCommission,
  MsgSend,
  Coin,
} from "./infrastructure/messages.js";

// TX envelope
export {
  encodeTxBody,
  encodeAuthInfo,
  encodeSignDoc,
  encodeTxRaw,
} from "./infrastructure/envelope.js";

// Gas & fee
export {
  DEFAULT_GAS_SUBMIT_ACTIVITY,
  DEFAULT_GAS_WITHDRAW_NODE_COMMISSION,
  DEFAULT_GAS_SEND,
  DEFAULT_GAS_FALLBACK,
  DEFAULT_FEE_DENOM,
  DEFAULT_GAS_PRICE,
  defaultGasForMessage,
  calculateFee,
} from "./infrastructure/gas.js";

// TX pipeline
export type {
  PipelineConfig,
  TxRequest,
  TxResult,
  Signer,
  ChainQuerier,
} from "./infrastructure/tx-pipeline.js";
export {
  executeTx,
  pollTxConfirmation,
  ChainClientAdapter,
} from "./infrastructure/tx-pipeline.js";

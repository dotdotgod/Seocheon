export type SigningMode = "vault" | "keystore" | "direct";
export type BroadcastMode = "sync" | "async";

export interface ChainConfig {
  chain_id: string;
  rpc_endpoint: string;
  grpc_endpoint: string;
  gas_price?: string;
  gas_adjustment?: number;
}

export interface SigningConfig {
  mode: SigningMode;
  vault_endpoint?: string;
  key_name?: string;
  keystore_path?: string;
  passphrase_env?: string;
  mnemonic?: string;
}

export interface TxConfig {
  broadcast_mode?: BroadcastMode;
  confirm_timeout_ms?: number;
  confirm_poll_interval_ms?: number;
}

export interface SDKConfig {
  chain: ChainConfig;
  signing: SigningConfig;
  tx?: TxConfig;
}

export interface ResolvedTxConfig {
  broadcast_mode: BroadcastMode;
  confirm_timeout_ms: number;
  confirm_poll_interval_ms: number;
}

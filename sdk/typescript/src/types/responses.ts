export interface SubmitActivityResponse {
  tx_hash: string;
  block_height: number;
  window_number: number;
  epoch_number: number;
  quota_remaining: number;
}

export interface ActivityItem {
  activity_hash: string;
  content_uri: string;
  block_height: number;
  window_number: number;
  tx_hash: string;
}

export interface GetActivitiesResponse {
  activities: ActivityItem[];
  total_count: number;
}

export interface GetQuotaResponse {
  epoch_number: number;
  quota_total: number;
  quota_used: number;
  quota_remaining: number;
  is_feegrant: boolean;
  feegrant_expiry: number | null;
}

export interface EpochInfoResponse {
  block_height: number;
  epoch_number: number;
  epoch_start_block: number;
  epoch_end_block: number;
  epoch_progress: string;
  window_number: number;
  window_start_block: number;
  window_end_block: number;
  window_progress: string;
  blocks_until_next_window: number;
  blocks_until_next_epoch: number;
}

export interface WindowActivity {
  window_number: number;
  submission_count: number;
  has_activity: boolean;
}

export interface QualificationResponse {
  epoch_number: number;
  total_windows: number;
  elapsed_windows: number;
  active_windows: number;
  required_windows: number;
  is_qualified: boolean;
  remaining_needed: number;
  can_still_qualify: boolean;
  window_detail: WindowActivity[];
}

export type NodeStatus =
  | "UNSPECIFIED"
  | "REGISTERED"
  | "ACTIVE"
  | "INACTIVE"
  | "JAILED";

export interface NodeInfoResponse {
  node_id: string;
  operator: string;
  agent_address: string;
  status: NodeStatus;
  description: string;
  website: string;
  tags: string[];
  commission_rate: string;
  agent_share: string;
  total_delegation: string;
  self_delegation: string;
  validator_address: string;
  registered_at: number;
}

export interface NodeSummary {
  node_id: string;
  status: NodeStatus;
  tags: string[];
  total_delegation: string;
  description: string;
}

export interface NodeSearchResponse {
  nodes: NodeSummary[];
  total_count: number;
}

export interface PendingRewardsResponse {
  delegation_reward: string;
  activity_reward: string;
  total_reward: string;
  commission_total: string;
  operator_share: string;
  agent_share: string;
}

export interface WithdrawRewardsResponse {
  tx_hash: string;
  withdrawn_total: string;
  to_operator: string;
  to_agent: string;
}

export interface BalanceResponse {
  address: string;
  balance: string;
  balance_kkot: string;
}

export interface SendTokensResponse {
  tx_hash: string;
  block_height: number;
}

export interface BlockInfoResponse {
  block_height: number;
  block_time: string;
  chain_id: string;
  num_txs: number;
}

export interface DelegationStatusResponse {
  expiry_epoch: number;
  current_epoch: number;
  in_renewal_window: boolean;
  renewal_window_start: number;
}

export interface TxEvent {
  type: string;
  attributes: EventAttribute[];
}

export interface EventAttribute {
  key: string;
  value: string;
}

export interface TxResultResponse {
  tx_hash: string;
  height: number;
  code: number;
  gas_used: number;
  gas_wanted: number;
  raw_log: string;
  events: TxEvent[];
}

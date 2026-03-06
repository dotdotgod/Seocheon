package types

// SubmitActivityResponse is returned after submitting an activity.
type SubmitActivityResponse struct {
	TxHash         string `json:"tx_hash"`
	BlockHeight    int64  `json:"block_height"`
	WindowNumber   int64  `json:"window_number"`
	EpochNumber    int64  `json:"epoch_number"`
	QuotaRemaining uint64 `json:"quota_remaining"`
}

// ActivityItem represents a single activity record.
type ActivityItem struct {
	ActivityHash string `json:"activity_hash"`
	ContentURI   string `json:"content_uri"`
	BlockHeight  int64  `json:"block_height"`
	WindowNumber int64  `json:"window_number"`
	TxHash       string `json:"tx_hash"`
}

// GetActivitiesResponse is returned when querying activities.
type GetActivitiesResponse struct {
	Activities []ActivityItem `json:"activities"`
	TotalCount uint64         `json:"total_count"`
}

// GetQuotaResponse is returned when querying activity quota.
type GetQuotaResponse struct {
	EpochNumber    int64   `json:"epoch_number"`
	QuotaTotal     uint64  `json:"quota_total"`
	QuotaUsed      uint64  `json:"quota_used"`
	QuotaRemaining uint64  `json:"quota_remaining"`
	IsFeegrant     bool    `json:"is_feegrant"`
	FeegrantExpiry *int64  `json:"feegrant_expiry,omitempty"`
}

// EpochInfoResponse is returned when querying epoch information.
type EpochInfoResponse struct {
	BlockHeight          int64  `json:"block_height"`
	EpochNumber          int64  `json:"epoch_number"`
	EpochStartBlock      int64  `json:"epoch_start_block"`
	EpochEndBlock        int64  `json:"epoch_end_block"`
	EpochProgress        string `json:"epoch_progress"`
	WindowNumber         int64  `json:"window_number"`
	WindowStartBlock     int64  `json:"window_start_block"`
	WindowEndBlock       int64  `json:"window_end_block"`
	WindowProgress       string `json:"window_progress"`
	BlocksUntilNextWindow int64 `json:"blocks_until_next_window"`
	BlocksUntilNextEpoch  int64 `json:"blocks_until_next_epoch"`
}

// WindowActivity represents activity within a single window.
type WindowActivity struct {
	WindowNumber    int64  `json:"window_number"`
	SubmissionCount uint64 `json:"submission_count"`
	HasActivity     bool   `json:"has_activity"`
}

// QualificationResponse is returned when querying reward qualification.
type QualificationResponse struct {
	EpochNumber      int64            `json:"epoch_number"`
	TotalWindows     int64            `json:"total_windows"`
	ElapsedWindows   int64            `json:"elapsed_windows"`
	ActiveWindows    uint64           `json:"active_windows"`
	RequiredWindows  int64            `json:"required_windows"`
	IsQualified      bool             `json:"is_qualified"`
	RemainingNeeded  int64            `json:"remaining_needed"`
	CanStillQualify  bool             `json:"can_still_qualify"`
	WindowDetail     []WindowActivity `json:"window_detail"`
}

// NodeInfoResponse is returned when querying node information.
type NodeInfoResponse struct {
	NodeID           string   `json:"node_id"`
	Operator         string   `json:"operator"`
	AgentAddress     string   `json:"agent_address"`
	Status           string   `json:"status"`
	Description      string   `json:"description"`
	Website          string   `json:"website"`
	Tags             []string `json:"tags"`
	CommissionRate   string   `json:"commission_rate"`
	AgentShare       string   `json:"agent_share"`
	TotalDelegation  string   `json:"total_delegation"`
	SelfDelegation   string   `json:"self_delegation"`
	ValidatorAddress string   `json:"validator_address"`
	RegisteredAt     int64    `json:"registered_at"`
}

// NodeSummary is a condensed view of a node.
type NodeSummary struct {
	NodeID          string   `json:"node_id"`
	Status          string   `json:"status"`
	Tags            []string `json:"tags"`
	TotalDelegation string   `json:"total_delegation"`
	Description     string   `json:"description"`
}

// NodeSearchResponse is returned from a node search.
type NodeSearchResponse struct {
	Nodes      []NodeSummary `json:"nodes"`
	TotalCount uint64        `json:"total_count"`
}

// PendingRewardsResponse is returned when querying pending rewards.
type PendingRewardsResponse struct {
	DelegationReward string `json:"delegation_reward"`
	ActivityReward   string `json:"activity_reward"`
	TotalReward      string `json:"total_reward"`
	CommissionTotal  string `json:"commission_total"`
	OperatorShare    string `json:"operator_share"`
	AgentShare       string `json:"agent_share"`
}

// WithdrawRewardsResponse is returned after withdrawing rewards.
type WithdrawRewardsResponse struct {
	TxHash         string `json:"tx_hash"`
	WithdrawnTotal string `json:"withdrawn_total"`
	ToOperator     string `json:"to_operator"`
	ToAgent        string `json:"to_agent"`
}

// BalanceResponse is returned when querying balance.
type BalanceResponse struct {
	Address     string `json:"address"`
	Balance     string `json:"balance"`
	BalanceKkot string `json:"balance_kkot"`
}

// SendTokensResponse is returned after sending tokens.
type SendTokensResponse struct {
	TxHash      string `json:"tx_hash"`
	BlockHeight int64  `json:"block_height"`
}

// BlockInfoResponse is returned when querying block information.
type BlockInfoResponse struct {
	BlockHeight int64  `json:"block_height"`
	BlockTime   string `json:"block_time"`
	ChainID     string `json:"chain_id"`
	NumTxs      uint64 `json:"num_txs"`
}

// TxEvent represents an event emitted by a transaction.
type TxEvent struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

// EventAttribute is a key-value pair within a TxEvent.
type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DelegationStatusResponse is returned when querying delegation confirmation status.
type DelegationStatusResponse struct {
	ExpiryEpoch        int64 `json:"expiry_epoch"`
	CurrentEpoch       int64 `json:"current_epoch"`
	InRenewalWindow    bool  `json:"in_renewal_window"`
	RenewalWindowStart int64 `json:"renewal_window_start"`
}

// TxResultResponse is returned when querying a transaction result.
type TxResultResponse struct {
	TxHash    string    `json:"tx_hash"`
	Height    int64     `json:"height"`
	Code      uint32    `json:"code"`
	GasUsed   uint64    `json:"gas_used"`
	GasWanted uint64    `json:"gas_wanted"`
	RawLog    string    `json:"raw_log"`
	Events    []TxEvent `json:"events"`
}

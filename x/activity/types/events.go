package types

// Event types for the activity module.
const (
	EventTypeActivitySubmitted = "activity_submitted"
	EventTypeActivityPruned    = "activity_pruned"
	EventTypeWindowCompleted   = "window_completed"
	EventTypeEpochCompleted    = "epoch_completed"
	EventTypeEpochFeeState     = "epoch_fee_state"
	EventTypeFeesCollected     = "fees_collected"
	EventTypeActivityFeeCharged = "activity_fee_charged"
)

// Event attribute keys.
const (
	AttributeKeyNodeID        = "node_id"
	AttributeKeySubmitter     = "submitter"
	AttributeKeyActivityHash  = "activity_hash"
	AttributeKeyContentURI    = "content_uri"
	AttributeKeyEpoch         = "epoch"
	AttributeKeyWindow        = "window"
	AttributeKeySequence      = "sequence"
	AttributeKeyBlockHeight   = "block_height"
	AttributeKeyPrunedCount   = "pruned_count"
	AttributeKeyEligibleCount = "eligible_count"
	AttributeKeyActiveNodes   = "active_nodes"
	AttributeKeyQuotaType     = "quota_type"
)

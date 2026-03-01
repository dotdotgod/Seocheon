package types

// Event types for the randomness module.
const (
	EventTypeBeaconSubmitted      = "beacon_submitted"
	EventTypeBeaconVerified       = "beacon_verified"
	EventTypeRandomnessRequested  = "randomness_requested"
	EventTypeRandomnessFulfilled  = "randomness_fulfilled"
	EventTypeRandomnessExpired    = "randomness_expired"
)

// Event attribute keys.
const (
	AttributeKeyRound          = "round"
	AttributeKeyRandomness     = "randomness"
	AttributeKeySubmitter      = "submitter"
	AttributeKeyBlockHeight    = "block_height"
	AttributeKeyVerified       = "verified"
	AttributeKeyRequestID      = "request_id"
	AttributeKeyRequester      = "requester"
	AttributeKeyCommitHash     = "commit_hash"
	AttributeKeyTargetRound    = "target_round"
	AttributeKeyNumWords       = "num_words"
	AttributeKeyRequestFee     = "request_fee"
	AttributeKeyResultHash     = "result_hash"
	AttributeKeyBeaconSubmitter = "beacon_submitter"
	AttributeKeyFeePaid        = "fee_paid"
	AttributeKeyRefundAmount   = "refund_amount"
)

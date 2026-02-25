package types

// Event types for the node module.
const (
	EventTypeNodeRegistered        = "node_registered"
	EventTypeNodeUpdated           = "node_updated"
	EventTypeNodeDeactivated       = "node_deactivated"
	EventTypeNodeActivated         = "node_activated"
	EventTypeAgentAddressChanged   = "agent_address_changed"
	EventTypeAgentShareScheduled   = "agent_share_change_scheduled"
	EventTypeAgentShareChanged     = "agent_share_changed"
	EventTypeCommissionWithdrawn   = "node_commission_withdrawn"
	EventTypeFeegrantRenewed       = "feegrant_renewed"
	EventTypeFeegrantGrantFailed   = "feegrant_grant_failed"
	EventTypeNodeJailed            = "node_jailed"
	EventTypeUndelegateFailed      = "undelegate_failed"
)

// Event attribute keys.
const (
	AttributeKeyNodeID           = "node_id"
	AttributeKeyOperator         = "operator"
	AttributeKeyAgentAddress     = "agent_address"
	AttributeKeyOldAgentAddress  = "old_agent_address"
	AttributeKeyNewAgentAddress  = "new_agent_address"
	AttributeKeyValidatorAddress = "validator_address"
	AttributeKeyBlockHeight      = "block_height"
	AttributeKeyNewAgentShare    = "new_agent_share"
	AttributeKeyApplyAtBlock     = "apply_at_block"
	AttributeKeyOperatorAmount   = "operator_amount"
	AttributeKeyAgentAmount      = "agent_amount"
	AttributeKeyError            = "error"
)

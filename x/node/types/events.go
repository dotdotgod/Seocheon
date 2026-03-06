package types

// Event types for the node module.
const (
	EventTypeNodeRegistered              = "node_registered"
	EventTypeNodeUpdated                 = "node_updated"
	EventTypeNodeDeactivated             = "node_deactivated"
	EventTypeNodeActivated               = "node_activated"
	EventTypeAgentAddressChanged         = "agent_address_changed"
	EventTypeAgentShareScheduled         = "agent_share_change_scheduled"
	EventTypeAgentShareChanged           = "agent_share_changed"
	EventTypeCommissionWithdrawn         = "node_commission_withdrawn"
	EventTypeFeegrantGrantFailed         = "feegrant_grant_failed"
	EventTypeNodeJailed                  = "node_jailed"
	EventTypeUndelegateFailed            = "undelegate_failed"
	EventTypeBoostDistributed            = "boost_distributed"
	EventTypeDelegationConfirmed         = "delegation_confirmed"
	EventTypeDelegationForceUnbonded     = "delegation_force_unbonded"
	EventTypeDelegationRenewalWindowOpen = "delegation_renewal_window_open"
)

// Event attribute keys.
const (
	AttributeKeyNodeID             = "node_id"
	AttributeKeyOperator           = "operator"
	AttributeKeyAgentAddress       = "agent_address"
	AttributeKeyOldAgentAddress    = "old_agent_address"
	AttributeKeyNewAgentAddress    = "new_agent_address"
	AttributeKeyValidatorAddress   = "validator_address"
	AttributeKeyBlockHeight        = "block_height"
	AttributeKeyNewAgentShare      = "new_agent_share"
	AttributeKeyApplyAtBlock       = "apply_at_block"
	AttributeKeyOperatorAmount     = "operator_amount"
	AttributeKeyAgentAmount        = "agent_amount"
	AttributeKeyError              = "error"
	AttributeKeyBoostAmount        = "boost_amount"
	AttributeKeyBoostRecipients    = "boost_recipients"
	AttributeKeyBoostPerValidator  = "boost_per_validator"
	AttributeKeyDelegator          = "delegator"
	AttributeKeyExpiryEpoch        = "expiry_epoch"
	AttributeKeyUnbondCount        = "unbond_count"
	AttributeKeyRenewalWindowStart = "renewal_window_start"
)

package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values.
var (
	DefaultMaxRegistrationsPerBlock     = uint64(5)
	DefaultRegistrationCooldownBlocks   = uint64(100)
	DefaultRegistrationDeposit          = math.ZeroInt()
	DefaultAgentAllowedMsgTypes         = []string{
		"/seocheon.activity.v1.MsgSubmitActivity",
		"/cosmwasm.wasm.v1.MsgExecuteContract",
		"/cosmos.bank.v1beta1.MsgSend",
	}
	DefaultAgentFeegrantAllowedMsgTypes = []string{
		"/seocheon.activity.v1.MsgSubmitActivity",
		"/cosmwasm.wasm.v1.MsgExecuteContract",
	}
	DefaultAgentAddressChangeCooldown = uint64(17280) // 1 epoch
	DefaultMaxTags                    = uint32(10)
	DefaultMaxTagLength               = uint32(32)
)

// NewParams creates a new Params instance with the given values.
func NewParams() Params {
	return Params{
		MaxRegistrationsPerBlock:         DefaultMaxRegistrationsPerBlock,
		RegistrationCooldownBlocks:       DefaultRegistrationCooldownBlocks,
		RegistrationDeposit:              DefaultRegistrationDeposit,
		AgentAllowedMsgTypes:             DefaultAgentAllowedMsgTypes,
		AgentFeegrantAllowedMsgTypes:     DefaultAgentFeegrantAllowedMsgTypes,
		AgentAddressChangeCooldown:       DefaultAgentAddressChangeCooldown,
		MaxTags:                          DefaultMaxTags,
		MaxTagLength:                     DefaultMaxTagLength,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams()
}

// DefaultAgentShareForTest returns a valid agent_share value for testing (35).
// Used to test agent_share changes that are within the default max change rate.
func DefaultAgentShareForTest() math.LegacyDec {
	return math.LegacyNewDec(35)
}

// Validate validates the set of params.
func (p Params) Validate() error {
	if p.MaxRegistrationsPerBlock == 0 {
		return fmt.Errorf("max_registrations_per_block must be positive")
	}
	if p.RegistrationDeposit.IsNegative() {
		return fmt.Errorf("registration_deposit must be non-negative")
	}
	if p.MaxTags == 0 {
		return fmt.Errorf("max_tags must be positive")
	}
	if p.MaxTagLength == 0 {
		return fmt.Errorf("max_tag_length must be positive")
	}
	return nil
}

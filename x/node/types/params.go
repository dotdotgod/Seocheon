package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// Default parameter values.
var (
	DefaultMaxRegistrationsPerBlock   = uint64(5)
	DefaultRegistrationCooldownBlocks = uint64(100)
	DefaultRegistrationDeposit        = math.ZeroInt()
	DefaultAgentAllowedMsgTypes       = []string{
		"/seocheon.activity.v1.MsgSubmitActivity",
		"/cosmos.bank.v1beta1.MsgSend",
	}
	DefaultAgentFeegrantAllowedMsgTypes = []string{
		"/seocheon.activity.v1.MsgSubmitActivity",
	}
	DefaultAgentAddressChangeCooldown   = uint64(17280) // 1 epoch
	DefaultMaxTags                      = uint32(10)
	DefaultMaxTagLength                 = uint32(32)
	DefaultDelegationConfirmationPeriod = uint64(90)         // ~90 days
	DefaultDelegationRenewalWindow      = uint64(7)          // ~7 days
	DefaultParamEpochLength             = DefaultEpochLength // 17280 blocks (~1 day)
)

// NewParams creates a new Params instance with the given values.
func NewParams() Params {
	return Params{
		MaxRegistrationsPerBlock:     DefaultMaxRegistrationsPerBlock,
		RegistrationCooldownBlocks:   DefaultRegistrationCooldownBlocks,
		RegistrationDeposit:          DefaultRegistrationDeposit,
		AgentAllowedMsgTypes:         DefaultAgentAllowedMsgTypes,
		AgentFeegrantAllowedMsgTypes: DefaultAgentFeegrantAllowedMsgTypes,
		AgentAddressChangeCooldown:   DefaultAgentAddressChangeCooldown,
		MaxTags:                      DefaultMaxTags,
		MaxTagLength:                 DefaultMaxTagLength,
		DelegationConfirmationPeriod: DefaultDelegationConfirmationPeriod,
		DelegationRenewalWindow:      DefaultDelegationRenewalWindow,
		EpochLength:                  DefaultParamEpochLength,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams()
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
	if p.EpochLength <= 0 {
		return fmt.Errorf("epoch_length must be positive")
	}
	if p.DelegationConfirmationPeriod > 0 {
		if p.DelegationRenewalWindow == 0 {
			return fmt.Errorf("delegation_renewal_window must be positive when confirmation period is set")
		}
		if p.DelegationRenewalWindow >= p.DelegationConfirmationPeriod {
			return fmt.Errorf("delegation_renewal_window must be less than delegation_confirmation_period")
		}
	}
	return nil
}

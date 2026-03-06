package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default pool balances.
// Registration Pool: 0.1 KKOT = 1,000,000,000 uppyeo (10^9)
// Feegrant Pool: 1 KKOT = 10,000,000,000 uppyeo (10^10)
// Boost Pool: 13,500 KKOT = 135,000,000,000,000 uppyeo (1.35×10^14)
var (
	DefaultRegistrationPoolBalance = sdk.NewCoins(sdk.NewCoin("uppyeo", math.NewInt(1_000_000_000)))
	DefaultFeegrantPoolBalance     = sdk.NewCoins(sdk.NewCoin("uppyeo", math.NewInt(10_000_000_000)))
	DefaultBoostPoolBalance        = sdk.NewCoins(sdk.NewCoin("uppyeo", math.NewInt(135_000_000_000_000)))
	DefaultBoostTargetEpochs       = uint64(730)
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                  DefaultParams(),
		Nodes:                   []Node{},
		RegistrationPoolBalance: DefaultRegistrationPoolBalance,
		FeegrantPoolBalance:     DefaultFeegrantPoolBalance,
		BoostPoolBalance:        DefaultBoostPoolBalance,
		BoostTargetEpochs:       DefaultBoostTargetEpochs,
		DelegationConfirmations: []DelegationConfirmation{},
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}

	nodeIDs := make(map[string]bool)
	operators := make(map[string]bool)
	agents := make(map[string]bool)

	for i, node := range gs.Nodes {
		if node.Id == "" {
			return fmt.Errorf("node %d has empty ID", i)
		}
		if nodeIDs[node.Id] {
			return fmt.Errorf("duplicate node ID: %s", node.Id)
		}
		nodeIDs[node.Id] = true

		if node.Operator == "" {
			return fmt.Errorf("node %s has empty operator", node.Id)
		}
		if operators[node.Operator] {
			return fmt.Errorf("duplicate operator address: %s", node.Operator)
		}
		operators[node.Operator] = true

		if node.AgentAddress != "" {
			if agents[node.AgentAddress] {
				return fmt.Errorf("duplicate agent address: %s", node.AgentAddress)
			}
			agents[node.AgentAddress] = true
		}
	}

	if !gs.RegistrationPoolBalance.IsValid() {
		return fmt.Errorf("invalid registration pool balance")
	}
	if !gs.FeegrantPoolBalance.IsValid() {
		return fmt.Errorf("invalid feegrant pool balance")
	}
	if !gs.BoostPoolBalance.IsValid() {
		return fmt.Errorf("invalid boost pool balance")
	}
	if gs.BoostTargetEpochs == 0 && gs.BoostPoolBalance.IsAllPositive() {
		return fmt.Errorf("boost_target_epochs must be > 0 when boost pool has balance")
	}

	// Validate delegation confirmations.
	delKeys := make(map[string]bool)
	for i, dc := range gs.DelegationConfirmations {
		if dc.DelegatorAddress == "" {
			return fmt.Errorf("delegation confirmation %d has empty delegator address", i)
		}
		if dc.ValidatorAddress == "" {
			return fmt.Errorf("delegation confirmation %d has empty validator address", i)
		}
		key := dc.DelegatorAddress + "/" + dc.ValidatorAddress
		if delKeys[key] {
			return fmt.Errorf("duplicate delegation confirmation: %s", key)
		}
		delKeys[key] = true
	}

	return nil
}

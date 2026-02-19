package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Default pool balances.
// Registration Pool: 1,000 SCN = 1,000,000,000,000 usum (10^12)
// Feegrant Pool: 10,000 SCN = 10,000,000,000,000 usum (10^13)
var (
	DefaultRegistrationPoolBalance = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1_000_000_000_000)))
	DefaultFeegrantPoolBalance     = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(10_000_000_000_000)))
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                  DefaultParams(),
		Nodes:                   []Node{},
		RegistrationPoolBalance: DefaultRegistrationPoolBalance,
		FeegrantPoolBalance:     DefaultFeegrantPoolBalance,
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

	return nil
}

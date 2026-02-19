package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:                  DefaultParams(),
		Nodes:                   []Node{},
		RegistrationPoolBalance: sdk.NewCoins(),
		FeegrantPoolBalance:     sdk.NewCoins(),
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

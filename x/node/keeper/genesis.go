package keeper

import (
	"context"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	if err := k.Params.Set(ctx, genState.Params); err != nil {
		return err
	}

	// Initialize nodes and their indexes.
	for _, node := range genState.Nodes {
		if err := k.Nodes.Set(ctx, node.Id, node); err != nil {
			return err
		}
		if err := k.OperatorIndex.Set(ctx, node.Operator, node.Id); err != nil {
			return err
		}
		if node.AgentAddress != "" {
			if err := k.AgentIndex.Set(ctx, node.AgentAddress, node.Id); err != nil {
				return err
			}
		}
		if node.ValidatorAddress != "" {
			if err := k.ValidatorIndex.Set(ctx, node.ValidatorAddress, node.Id); err != nil {
				return err
			}
		}
		for _, tag := range node.Tags {
			if err := k.TagIndex.Set(ctx, collections.Join(tag, node.Id)); err != nil {
				return err
			}
		}
	}

	// Fund Registration Pool module account.
	if genState.RegistrationPoolBalance.IsAllPositive() {
		regPoolAddr := k.authKeeper.GetModuleAddress(types.RegistrationPoolName)
		if regPoolAddr != nil {
			// The actual funding is handled by x/bank genesis via module account balances.
			_ = regPoolAddr
		}
	}

	// Fund Feegrant Pool module account.
	if genState.FeegrantPoolBalance.IsAllPositive() {
		fgPoolAddr := k.authKeeper.GetModuleAddress(types.FeegrantPoolName)
		if fgPoolAddr != nil {
			_ = fgPoolAddr
		}
	}

	return nil
}

// ExportGenesis returns the module's exported genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	var nodes []types.Node
	err = k.Nodes.Walk(ctx, nil, func(key string, node types.Node) (bool, error) {
		nodes = append(nodes, node)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Export pool balances.
	var regPoolBalance, fgPoolBalance sdk.Coins
	regPoolAddr := k.authKeeper.GetModuleAddress(types.RegistrationPoolName)
	if regPoolAddr != nil {
		regPoolBalance = k.bankKeeper.GetAllBalances(ctx, regPoolAddr)
	}
	fgPoolAddr := k.authKeeper.GetModuleAddress(types.FeegrantPoolName)
	if fgPoolAddr != nil {
		fgPoolBalance = k.bankKeeper.GetAllBalances(ctx, fgPoolAddr)
	}

	return &types.GenesisState{
		Params:                  params,
		Nodes:                   nodes,
		RegistrationPoolBalance: regPoolBalance,
		FeegrantPoolBalance:     fgPoolBalance,
	}, nil
}

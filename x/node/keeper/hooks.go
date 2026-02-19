package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/types"
)

// Hooks wraps the Keeper to implement the staking hooks interface.
type Hooks struct {
	k Keeper
}

var _ stakingtypes.StakingHooks = Hooks{}

// Hooks returns the staking hooks for the node module.
func (k Keeper) Hooks() Hooks {
	return Hooks{k: k}
}

// StakingHooksWrapper returns the hooks wrapped for depinject registration.
func (k Keeper) StakingHooksWrapper() stakingtypes.StakingHooksWrapper {
	return stakingtypes.StakingHooksWrapper{StakingHooks: k.Hooks()}
}

// AfterValidatorBonded is called when a validator transitions to the bonded state.
// This indicates the validator has entered the Active Validator Set.
// We set the corresponding node status to ACTIVE.
func (h Hooks) AfterValidatorBonded(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	valAddrStr, err := sdk.Bech32ifyAddressBytes("seocheonvaloper", valAddr)
	if err != nil {
		return nil // silently skip if address conversion fails
	}

	nodeID, err := h.k.ValidatorIndex.Get(ctx, valAddrStr)
	if err != nil {
		return nil // not a registered node, skip
	}

	node, err := h.k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return nil
	}

	// Only transition to ACTIVE if the node is in REGISTERED status.
	if node.Status == types.NodeStatus_NODE_STATUS_REGISTERED {
		node.Status = types.NodeStatus_NODE_STATUS_ACTIVE
		if err := h.k.Nodes.Set(ctx, nodeID, node); err != nil {
			return err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			"node_activated",
			sdk.NewAttribute("node_id", nodeID),
			sdk.NewAttribute("validator_address", valAddrStr),
		))
	}

	return nil
}

// AfterValidatorBeginUnbonding is called when a validator begins unbonding.
// This indicates the validator has left the Active Validator Set.
// We set the corresponding node status back to REGISTERED (unless it's INACTIVE).
func (h Hooks) AfterValidatorBeginUnbonding(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	valAddrStr, err := sdk.Bech32ifyAddressBytes("seocheonvaloper", valAddr)
	if err != nil {
		return nil
	}

	nodeID, err := h.k.ValidatorIndex.Get(ctx, valAddrStr)
	if err != nil {
		return nil // not a registered node, skip
	}

	node, err := h.k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return nil
	}

	// Only transition from ACTIVE to REGISTERED (not from INACTIVE).
	if node.Status == types.NodeStatus_NODE_STATUS_ACTIVE {
		node.Status = types.NodeStatus_NODE_STATUS_REGISTERED
		if err := h.k.Nodes.Set(ctx, nodeID, node); err != nil {
			return err
		}

		sdkCtx := sdk.UnwrapSDKContext(ctx)
		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			"node_deactivated",
			sdk.NewAttribute("node_id", nodeID),
			sdk.NewAttribute("validator_address", valAddrStr),
		))
	}

	return nil
}

// Unused hooks - required by interface.
func (h Hooks) AfterValidatorCreated(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeValidatorModified(ctx context.Context, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterValidatorRemoved(ctx context.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationCreated(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationSharesModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeDelegationRemoved(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) AfterDelegationModified(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) error {
	return nil
}
func (h Hooks) BeforeValidatorSlashed(ctx context.Context, valAddr sdk.ValAddress, fraction math.LegacyDec) error {
	return nil
}
func (h Hooks) AfterUnbondingInitiated(ctx context.Context, id uint64) error {
	return nil
}

package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/types"
)

// DeactivateNode deactivates a node by setting its status to INACTIVE
// and begins unbonding the 1 usum self-delegation.
func (k msgServer) DeactivateNode(ctx context.Context, msg *types.MsgDeactivateNode) (*types.MsgDeactivateNodeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Look up node by operator.
	nodeID, err := k.OperatorIndex.Get(ctx, msg.Operator)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrNodeNotFound, "no node found for operator %s", msg.Operator)
	}

	node, err := k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrNodeNotFound, "node %s not found", nodeID)
	}

	// Can only deactivate REGISTERED or ACTIVE nodes.
	if node.Status == types.NodeStatus_NODE_STATUS_INACTIVE {
		return nil, errorsmod.Wrap(types.ErrNodeInactive, "node is already inactive")
	}

	// Set status to INACTIVE.
	node.Status = types.NodeStatus_NODE_STATUS_INACTIVE
	if err := k.Nodes.Set(ctx, nodeID, node); err != nil {
		return nil, errorsmod.Wrap(err, "failed to update node status")
	}

	// Remove any pending agent share changes.
	_ = k.PendingAgentShareChanges.Remove(ctx, nodeID)

	// Revoke agent feegrant immediately.
	k.revokeAgentFeegrant(ctx, node.AgentAddress)

	// Reclaim 1 usum from operator to Registration Pool (pre-paid return).
	// The 1 usum will also return to operator after unbonding completes via x/staking.
	if bondDenom, bondErr := k.stakingKeeper.BondDenom(ctx); bondErr == nil {
		operatorAddr, addrErr := sdk.AccAddressFromBech32(msg.Operator)
		if addrErr == nil {
			oneUsum := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewInt(1)))
			_ = k.bankKeeper.SendCoinsFromAccountToModule(ctx, operatorAddr, types.RegistrationPoolName, oneUsum)
		}
	}

	// Begin unbonding 1 usum self-delegation via staking MsgServer.
	if k.stakingMsgServer != nil && node.ValidatorAddress != "" {
		bondDenom, bondErr := k.stakingKeeper.BondDenom(ctx)
		if bondErr == nil {
			undelegateMsg := &stakingtypes.MsgUndelegate{
				DelegatorAddress: msg.Operator,
				ValidatorAddress: node.ValidatorAddress,
				Amount:           sdk.NewCoin(bondDenom, math.NewInt(1)),
			}
			if _, undErr := k.stakingMsgServer.Undelegate(ctx, undelegateMsg); undErr != nil {
				// Best-effort: emit warning but don't block deactivation.
				sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
					types.EventTypeUndelegateFailed,
					sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
					sdk.NewAttribute(types.AttributeKeyError, undErr.Error()),
				))
			}
		}
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeNodeDeactivated,
		sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
		sdk.NewAttribute(types.AttributeKeyOperator, msg.Operator),
	))

	return &types.MsgDeactivateNodeResponse{}, nil
}

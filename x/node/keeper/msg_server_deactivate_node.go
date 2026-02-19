package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// DeactivateNode deactivates a node by setting its status to INACTIVE.
// In a full implementation, this would also begin unbonding the 1 usum self-delegation.
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

	// Revoke feegrant from agent (if exists).
	if node.AgentAddress != "" && k.feegrantKeeper != nil {
		feegrantPoolAddr := k.authKeeper.GetModuleAddress(types.FeegrantPoolName)
		agentAddr, addrErr := sdk.AccAddressFromBech32(node.AgentAddress)
		if addrErr == nil {
			_ = k.feegrantKeeper.RevokeAllowance(ctx, feegrantPoolAddr, agentAddr)
		}
	}

	// TODO: Begin unbonding 1 usum self-delegation via stakingMsgServer.Undelegate()
	// when staking MsgServer wiring is complete.

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"node_deactivated",
		sdk.NewAttribute("node_id", nodeID),
		sdk.NewAttribute("operator", msg.Operator),
	))

	return &types.MsgDeactivateNodeResponse{}, nil
}

package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// UpdateNodeAgentShare requests an agent_share change toward a target value.
// The change is stored as a PendingAgentShareChange and applied gradually
// over multiple epochs (max_agent_share_change_rate per epoch) by EndBlocker.
func (k msgServer) UpdateNodeAgentShare(ctx context.Context, msg *types.MsgUpdateNodeAgentShare) (*types.MsgUpdateNodeAgentShareResponse, error) {
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

	// Verify the node is not inactive.
	if node.Status == types.NodeStatus_NODE_STATUS_INACTIVE {
		return nil, errorsmod.Wrap(types.ErrNodeInactive, "cannot update agent share for an inactive node")
	}

	// Validate range.
	hundred := math.LegacyNewDec(100)
	if msg.NewAgentShare.IsNegative() || msg.NewAgentShare.GT(hundred) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAgentShare, "new_agent_share %s out of range [0, 100]", msg.NewAgentShare)
	}

	// Check no existing pending change (must complete before submitting a new target).
	if has, _ := k.PendingAgentShareChanges.Has(ctx, nodeID); has {
		return nil, errorsmod.Wrap(types.ErrAgentShareChangePending, "a pending agent share change already exists; wait for completion before submitting a new target")
	}

	// Calculate apply_at_block: next epoch boundary.
	currentBlock := sdkCtx.BlockHeight()
	epochLength := types.EpochLength
	nextEpochBoundary := ((currentBlock / epochLength) + 1) * epochLength

	pending := types.PendingAgentShareChange{
		NodeId:        nodeID,
		NewAgentShare: msg.NewAgentShare,
		ApplyAtBlock:  nextEpochBoundary,
	}

	if err := k.PendingAgentShareChanges.Set(ctx, nodeID, pending); err != nil {
		return nil, errorsmod.Wrap(err, "failed to store pending agent share change")
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeAgentShareScheduled,
		sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
		sdk.NewAttribute(types.AttributeKeyNewAgentShare, msg.NewAgentShare.String()),
		sdk.NewAttribute(types.AttributeKeyApplyAtBlock, fmt.Sprintf("%d", nextEpochBoundary)),
	))

	return &types.MsgUpdateNodeAgentShareResponse{}, nil
}

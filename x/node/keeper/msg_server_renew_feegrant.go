package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// RenewFeegrant renews an expired feegrant for an eligible node.
//
// Eligibility requires:
// - Node is REGISTERED or ACTIVE (not INACTIVE)
// - Node has an agent_address set
// - Activity history: at least 20 out of the last 30 epochs with activity qualification
//   (This check requires x/activity module — stubbed as always-eligible in Phase 0)
//
// The renewed feegrant has the same terms as the initial grant:
// - PeriodicAllowance: period=17280 blocks, limit=1 SCN/epoch, expiry=6 months
func (k msgServer) RenewFeegrant(ctx context.Context, msg *types.MsgRenewFeegrant) (*types.MsgRenewFeegrantResponse, error) {
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

	// Node must not be inactive.
	if node.Status == types.NodeStatus_NODE_STATUS_INACTIVE {
		return nil, errorsmod.Wrap(types.ErrNodeInactive, "cannot renew feegrant for an inactive node")
	}

	// Node must have an agent_address.
	if node.AgentAddress == "" {
		return nil, errorsmod.Wrap(types.ErrNodeNotFound, "node has no agent address")
	}

	// TODO: Phase 1 — Check activity history via x/activity module.
	// Requirement: At least 20 out of the last 30 epochs with activity qualification.
	// For Phase 0, we skip this check and allow renewal for any eligible node.
	//
	// activityQualified := k.activityKeeper.CheckActivityHistory(ctx, nodeID, 30, 20)
	// if !activityQualified {
	//     return nil, errorsmod.Wrap(types.ErrInsufficientActivity, "...")
	// }

	// TODO: Grant new feegrant via feegrantKeeper when wired.
	// feegrantPoolAddr := k.authKeeper.GetModuleAddress(types.FeegrantPoolName)
	// agentAddr, _ := sdk.AccAddressFromBech32(node.AgentAddress)
	// allowance := &feegrant.PeriodicAllowance{...}
	// k.feegrantKeeper.GrantAllowance(ctx, feegrantPoolAddr, agentAddr, allowance)

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"feegrant_renewed",
		sdk.NewAttribute("node_id", nodeID),
		sdk.NewAttribute("agent_address", node.AgentAddress),
	))

	return &types.MsgRenewFeegrantResponse{}, nil
}

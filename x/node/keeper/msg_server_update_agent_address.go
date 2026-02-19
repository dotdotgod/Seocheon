package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// UpdateAgentAddress changes the node's agent wallet address.
// Enforces a cooldown period between changes.
// Empty new_agent_address deactivates the agent.
func (k msgServer) UpdateAgentAddress(ctx context.Context, msg *types.MsgUpdateAgentAddress) (*types.MsgUpdateAgentAddressResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get params")
	}

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
		return nil, errorsmod.Wrap(types.ErrNodeInactive, "cannot update agent address for an inactive node")
	}

	// Check cooldown.
	currentBlock := sdkCtx.BlockHeight()
	lastChangeBlock, err := k.LastAgentChangeBlock.Get(ctx, nodeID)
	if err == nil {
		// Cooldown exists — check if elapsed.
		if uint64(currentBlock-lastChangeBlock) < params.AgentAddressChangeCooldown {
			return nil, errorsmod.Wrapf(types.ErrAgentAddressChangeCooldown,
				"must wait %d blocks, only %d elapsed",
				params.AgentAddressChangeCooldown, currentBlock-lastChangeBlock)
		}
	}

	// Check new agent_address uniqueness (if not empty).
	if msg.NewAgentAddress != "" && msg.NewAgentAddress != node.AgentAddress {
		if has, _ := k.AgentIndex.Has(ctx, msg.NewAgentAddress); has {
			return nil, errorsmod.Wrapf(types.ErrAgentAddressAlreadyUsed,
				"agent address %s already registered", msg.NewAgentAddress)
		}
	}

	// Remove old agent index and revoke feegrant.
	oldAgent := node.AgentAddress
	if oldAgent != "" {
		if err := k.AgentIndex.Remove(ctx, oldAgent); err != nil {
			return nil, errorsmod.Wrap(err, "failed to remove old agent index")
		}
		// Revoke feegrant from old agent (best-effort: skip if feegrantKeeper not wired).
		if k.feegrantKeeper != nil {
			feegrantPoolAddr := k.authKeeper.GetModuleAddress(types.FeegrantPoolName)
			oldAgentAddr, addrErr := sdk.AccAddressFromBech32(oldAgent)
			if addrErr == nil {
				// Ignore revoke error — feegrant may not exist.
				_ = k.feegrantKeeper.RevokeAllowance(ctx, feegrantPoolAddr, oldAgentAddr)
			}
		}
	}

	// Set new agent address.
	node.AgentAddress = msg.NewAgentAddress

	if msg.NewAgentAddress != "" {
		if err := k.AgentIndex.Set(ctx, msg.NewAgentAddress, nodeID); err != nil {
			return nil, errorsmod.Wrap(err, "failed to set new agent index")
		}
		// Grant feegrant to new agent (best-effort).
		// Full feegrant granting logic will be added when feegrant wiring is complete.
	}

	if err := k.Nodes.Set(ctx, nodeID, node); err != nil {
		return nil, errorsmod.Wrap(err, "failed to update node")
	}

	// Record cooldown.
	if err := k.LastAgentChangeBlock.Set(ctx, nodeID, currentBlock); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set last agent change block")
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"agent_address_changed",
		sdk.NewAttribute("node_id", nodeID),
		sdk.NewAttribute("old_agent_address", oldAgent),
		sdk.NewAttribute("new_agent_address", msg.NewAgentAddress),
		sdk.NewAttribute("block_height", fmt.Sprintf("%d", currentBlock)),
	))

	return &types.MsgUpdateAgentAddressResponse{}, nil
}

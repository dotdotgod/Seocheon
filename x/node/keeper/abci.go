package keeper

import (
	"context"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// EndBlocker is called at the end of every block.
// It applies pending agent_share changes at epoch boundaries.
func (k Keeper) EndBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()

	// Only process at epoch boundaries.
	if blockHeight%types.EpochLength != 0 {
		return nil
	}

	return k.applyPendingAgentShareChanges(ctx, blockHeight)
}

// applyPendingAgentShareChanges iterates all pending changes and applies those
// scheduled for the current block or earlier.
func (k Keeper) applyPendingAgentShareChanges(ctx context.Context, blockHeight int64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	iter, err := k.PendingAgentShareChanges.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	var toRemove []string

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		nodeID := kv.Key
		pending := kv.Value

		if pending.ApplyAtBlock > blockHeight {
			continue
		}

		// Apply the change.
		node, err := k.Nodes.Get(ctx, nodeID)
		if err != nil {
			// Node was deleted/deactivated — just remove the pending change.
			toRemove = append(toRemove, nodeID)
			continue
		}

		node.AgentShare = pending.NewAgentShare
		if err := k.Nodes.Set(ctx, nodeID, node); err != nil {
			return err
		}

		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			"agent_share_changed",
			sdk.NewAttribute("node_id", nodeID),
			sdk.NewAttribute("new_agent_share", pending.NewAgentShare.String()),
		))

		toRemove = append(toRemove, nodeID)
	}

	// Remove applied changes.
	for _, nodeID := range toRemove {
		if err := k.PendingAgentShareChanges.Remove(ctx, nodeID); err != nil {
			return err
		}
	}

	return nil
}

// NodesByTagIterator returns an iterator over all node IDs with the given tag.
func (k Keeper) NodesByTagIterator(ctx context.Context, tag string) (collections.KeySetIterator[collections.Pair[string, string]], error) {
	rng := collections.NewPrefixedPairRange[string, string](tag)
	return k.TagIndex.Iterate(ctx, rng)
}

package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
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
// Each epoch, agent_share moves by at most max_agent_share_change_rate toward the target.
// If the target is not yet reached, the pending change is kept for the next epoch.
func (k Keeper) applyPendingAgentShareChanges(ctx context.Context, blockHeight int64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	iter, err := k.PendingAgentShareChanges.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	type pendingAction struct {
		nodeID  string
		pending types.PendingAgentShareChange
	}
	var toProcess []pendingAction

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}
		if kv.Value.ApplyAtBlock > blockHeight {
			continue
		}
		toProcess = append(toProcess, pendingAction{nodeID: kv.Key, pending: kv.Value})
	}

	for _, action := range toProcess {
		nodeID := action.nodeID
		pending := action.pending

		node, err := k.Nodes.Get(ctx, nodeID)
		if err != nil {
			// Node was deleted/deactivated — just remove the pending change.
			_ = k.PendingAgentShareChanges.Remove(ctx, nodeID)
			continue
		}

		target := pending.NewAgentShare
		diff := target.Sub(node.AgentShare)
		absDiff := diff.Abs()
		maxRate := node.MaxAgentShareChangeRate

		var newShare math.LegacyDec
		if absDiff.LTE(maxRate) {
			// Can reach target in this epoch.
			newShare = target
		} else {
			// Move by max_change_rate toward target.
			if diff.IsPositive() {
				newShare = node.AgentShare.Add(maxRate)
			} else {
				newShare = node.AgentShare.Sub(maxRate)
			}
		}

		node.AgentShare = newShare
		if err := k.Nodes.Set(ctx, nodeID, node); err != nil {
			return err
		}

		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeAgentShareChanged,
			sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
			sdk.NewAttribute(types.AttributeKeyNewAgentShare, newShare.String()),
		))

		if newShare.Equal(target) {
			// Target reached — remove pending.
			_ = k.PendingAgentShareChanges.Remove(ctx, nodeID)
		} else {
			// Not yet at target — schedule next step at next epoch boundary.
			pending.ApplyAtBlock = blockHeight + types.EpochLength
			_ = k.PendingAgentShareChanges.Set(ctx, nodeID, pending)
		}
	}

	return nil
}

// NodesByTagIterator returns an iterator over all node IDs with the given tag.
func (k Keeper) NodesByTagIterator(ctx context.Context, tag string) (collections.KeySetIterator[collections.Pair[string, string]], error) {
	rng := collections.NewPrefixedPairRange[string, string](tag)
	return k.TagIndex.Iterate(ctx, rng)
}

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/activity/types"
)

// EndBlocker is called at the end of every block.
// It emits window/epoch boundary events and performs pruning at epoch boundaries.
func (k Keeper) EndBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Window boundary: emit window_completed event.
	if IsWindowBoundary(blockHeight, params) {
		epoch := GetCurrentEpoch(blockHeight, params)
		window := GetCurrentWindow(blockHeight, params)

		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeWindowCompleted,
			sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epoch)),
			sdk.NewAttribute(types.AttributeKeyWindow, fmt.Sprintf("%d", window)),
		))
	}

	// Epoch boundary: emit epoch_completed + fee state calculation + pruning.
	if IsEpochBoundary(blockHeight, params) {
		completedEpoch := GetCurrentEpoch(blockHeight, params)

		// Count eligible nodes.
		eligibleCount := int64(len(k.getEligibleNodeIDs(ctx, completedEpoch)))

		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeEpochCompleted,
			sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", completedEpoch)),
			sdk.NewAttribute(types.AttributeKeyEligibleCount, fmt.Sprintf("%d", eligibleCount)),
		))

		// Distribute collected fees from the completed epoch (80% → activity pool, 20% → community pool).
		if err := k.DistributeCollectedFees(ctx, completedEpoch); err != nil {
			return err
		}

		// Distribute activity reward pool equally to eligible nodes.
		if err := k.DistributeActivityRewards(ctx, completedEpoch); err != nil {
			return err
		}

		// Calculate and cache fee state for the next epoch.
		nextEpoch := completedEpoch + 1
		if err := k.CalculateAndCacheEpochFeeState(ctx, nextEpoch); err != nil {
			return err
		}

		// Pruning.
		prunedCount, err := k.pruneOldActivities(ctx, blockHeight, params)
		if err != nil {
			return err
		}

		if prunedCount > 0 {
			sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
				types.EventTypeActivityPruned,
				sdk.NewAttribute(types.AttributeKeyPrunedCount, fmt.Sprintf("%d", prunedCount)),
			))
		}
	}

	return nil
}

// pruneOldActivities removes activity records older than pruning_keep_blocks.
// Returns the number of pruned records. Bounded to max 1000 per epoch for gas safety.
func (k Keeper) pruneOldActivities(ctx context.Context, currentHeight int64, params types.Params) (int64, error) {
	cutoffHeight := currentHeight - params.ActivityPruningKeepBlocks
	if cutoffHeight <= 0 {
		return 0, nil
	}

	var pruned int64
	const maxPrunePerEpoch = 1000

	// Iterate BlockActivities from lowest block height.
	iter, err := k.BlockActivities.Iterate(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	type pruneEntry struct {
		blockHeight int64
		seq         uint64
		nodeID      string
	}
	var toDelete []pruneEntry

	for ; iter.Valid(); iter.Next() {
		if pruned >= maxPrunePerEpoch {
			break
		}

		kv, err := iter.KeyValue()
		if err != nil {
			return pruned, err
		}

		blockHeight := kv.Key.K1()
		if blockHeight >= cutoffHeight {
			break // BlockActivities is ordered by block height
		}

		toDelete = append(toDelete, pruneEntry{
			blockHeight: blockHeight,
			seq:         kv.Key.K2(),
			nodeID:      kv.Value,
		})
		pruned++
	}

	// Delete in a separate pass to avoid iterator mutation.
	for _, entry := range toDelete {
		// Remove from BlockActivities.
		if err := k.BlockActivities.Remove(ctx, collections.Join(entry.blockHeight, entry.seq)); err != nil {
			return pruned, err
		}

		// Find and remove the actual activity record.
		// We need to find it by iterating the node's activities at this block height.
		nodeIter, err := k.Activities.Iterate(ctx, collections.NewPrefixedTripleRange[string, int64, uint64](entry.nodeID))
		if err != nil {
			continue
		}

		for ; nodeIter.Valid(); nodeIter.Next() {
			record, err := nodeIter.Value()
			if err != nil {
				break
			}
			if record.BlockHeight == entry.blockHeight {
				// Remove from Activities.
				if err := k.Activities.Remove(ctx, collections.Join3(record.NodeId, record.Epoch, record.Sequence)); err != nil {
					nodeIter.Close()
					return pruned, err
				}
				// Remove from HashIndex.
				_ = k.HashIndex.Remove(ctx, collections.Join3(record.NodeId, record.Epoch, record.ActivityHash))
				// Remove from GlobalHashIndex.
				_ = k.GlobalHashIndex.Remove(ctx, record.ActivityHash)
				break
			}
		}
		nodeIter.Close()
	}

	return pruned, nil
}

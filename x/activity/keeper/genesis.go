package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"seocheon/x/activity/types"
)

// InitGenesis initializes the module state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	if err := k.Params.Set(ctx, genState.Params); err != nil {
		return err
	}

	for _, activity := range genState.Activities {
		// Store activity record.
		if err := k.Activities.Set(ctx, collections.Join3(activity.NodeId, activity.Epoch, activity.Sequence), activity); err != nil {
			return err
		}

		// Rebuild HashIndex.
		if err := k.HashIndex.Set(ctx, collections.Join3(activity.NodeId, activity.Epoch, activity.ActivityHash)); err != nil {
			return err
		}

		// Rebuild BlockActivities.
		blockSeq, err := k.getNextBlockSeq(ctx, activity.BlockHeight)
		if err != nil {
			return err
		}
		if err := k.BlockActivities.Set(ctx, collections.Join(activity.BlockHeight, blockSeq), activity.NodeId); err != nil {
			return err
		}

		// Rebuild EpochQuotaUsed.
		quotaUsed, err := k.EpochQuotaUsed.Get(ctx, collections.Join(activity.NodeId, activity.Epoch))
		if err != nil {
			quotaUsed = 0
		}
		if err := k.EpochQuotaUsed.Set(ctx, collections.Join(activity.NodeId, activity.Epoch), quotaUsed+1); err != nil {
			return err
		}

		// Rebuild ActivitySequence.
		seq, err := k.ActivitySequence.Get(ctx, collections.Join(activity.NodeId, activity.Epoch))
		if err != nil {
			seq = 0
		}
		if activity.Sequence >= seq {
			if err := k.ActivitySequence.Set(ctx, collections.Join(activity.NodeId, activity.Epoch), activity.Sequence+1); err != nil {
				return err
			}
		}

		// Rebuild WindowActivity and EpochSummary.
		params := genState.Params
		window := GetCurrentWindow(activity.BlockHeight, params)

		windowCount, err := k.WindowActivity.Get(ctx, collections.Join3(activity.NodeId, activity.Epoch, window))
		if err != nil {
			windowCount = 0
		}
		if err := k.WindowActivity.Set(ctx, collections.Join3(activity.NodeId, activity.Epoch, window), windowCount+1); err != nil {
			return err
		}

		// Update EpochSummary.
		summary, err := k.EpochSummary.Get(ctx, collections.Join(activity.NodeId, activity.Epoch))
		if err != nil {
			summary = types.EpochActivitySummary{}
		}
		summary.TotalActivities++
		if windowCount == 0 {
			summary.ActiveWindows++
		}
		summary.Eligible = int64(summary.ActiveWindows) >= params.MinActiveWindows
		if err := k.EpochSummary.Set(ctx, collections.Join(activity.NodeId, activity.Epoch), summary); err != nil {
			return err
		}
	}

	return nil
}

// ExportGenesis returns the module's genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	var activities []types.ActivityRecord
	iter, err := k.Activities.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		record, err := iter.Value()
		if err != nil {
			return nil, err
		}
		activities = append(activities, record)
	}

	return &types.GenesisState{
		Params:     params,
		Activities: activities,
	}, nil
}

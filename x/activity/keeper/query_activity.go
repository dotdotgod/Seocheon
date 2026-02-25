package keeper

import (
	"context"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/activity/types"
)

// Activity queries a single activity by hash using HashIndex scan.
func (qs queryServer) Activity(ctx context.Context, req *types.QueryActivityRequest) (*types.QueryActivityResponse, error) {
	if req.ActivityHash == "" {
		return nil, types.ErrInvalidActivityHash
	}

	// Scan HashIndex for matching hash_hex (3rd key component).
	iter, err := qs.k.HashIndex.Iterate(ctx, nil)
	if err != nil {
		return nil, types.ErrActivityNotFound
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			continue
		}
		if key.K3() != req.ActivityHash {
			continue
		}
		// Found matching hash — look up the activity by (node_id, epoch).
		nodeID := key.K1()
		epoch := key.K2()
		// Find the record with matching hash in this node+epoch.
		rng := collections.NewSuperPrefixedTripleRange[string, int64, uint64](nodeID, epoch)
		actIter, err := qs.k.Activities.Iterate(ctx, rng)
		if err != nil {
			continue
		}
		for ; actIter.Valid(); actIter.Next() {
			record, err := actIter.Value()
			if err != nil {
				break
			}
			if record.ActivityHash == req.ActivityHash {
				actIter.Close()
				return &types.QueryActivityResponse{Activity: record}, nil
			}
		}
		actIter.Close()
	}

	return nil, types.ErrActivityNotFound
}

// ActivitiesByNode queries activities by node_id with optional epoch filter and pagination.
func (qs queryServer) ActivitiesByNode(ctx context.Context, req *types.QueryActivitiesByNodeRequest) (*types.QueryActivitiesByNodeResponse, error) {
	if req.NodeId == "" {
		return nil, types.ErrNodeNotFound
	}

	var activities []types.ActivityRecord

	if req.Epoch > 0 {
		// Filter by specific epoch.
		rng := collections.NewSuperPrefixedTripleRange[string, int64, uint64](req.NodeId, req.Epoch)
		iter, err := qs.k.Activities.Iterate(ctx, rng)
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

		return &types.QueryActivitiesByNodeResponse{
			Activities: activities,
		}, nil
	}

	// All epochs for this node.
	rng := collections.NewPrefixedTripleRange[string, int64, uint64](req.NodeId)
	iter, err := qs.k.Activities.Iterate(ctx, rng)
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

	return &types.QueryActivitiesByNodeResponse{
		Activities: activities,
	}, nil
}

// ActivitiesByBlock queries activities submitted at a specific block height.
func (qs queryServer) ActivitiesByBlock(ctx context.Context, req *types.QueryActivitiesByBlockRequest) (*types.QueryActivitiesByBlockResponse, error) {
	var activities []types.ActivityRecord

	rng := collections.NewPrefixedPairRange[int64, uint64](req.BlockHeight)
	iter, err := qs.k.BlockActivities.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return nil, err
		}
		nodeID := kv.Value

		// We need to find the actual record. Iterate the node's activities to find by block height.
		nodeIter, err := qs.k.Activities.Iterate(ctx, collections.NewPrefixedTripleRange[string, int64, uint64](nodeID))
		if err != nil {
			return nil, err
		}

		for ; nodeIter.Valid(); nodeIter.Next() {
			record, err := nodeIter.Value()
			if err != nil {
				nodeIter.Close()
				return nil, err
			}
			if record.BlockHeight == req.BlockHeight {
				activities = append(activities, record)
			}
		}
		nodeIter.Close()
	}

	return &types.QueryActivitiesByBlockResponse{Activities: activities}, nil
}

// EpochInfo queries current epoch and window information.
func (qs queryServer) EpochInfo(ctx context.Context, _ *types.QueryEpochInfoRequest) (*types.QueryEpochInfoResponse, error) {
	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()

	epoch := GetCurrentEpoch(blockHeight, params)
	window := GetCurrentWindow(blockHeight, params)
	epochStart := GetEpochStartBlock(epoch, params)
	nextEpochStart := GetEpochStartBlock(epoch+1, params)
	blocksUntilNextEpoch := nextEpochStart - blockHeight

	return &types.QueryEpochInfoResponse{
		CurrentEpoch:          epoch,
		CurrentWindow:         window,
		EpochStartBlock:       epochStart,
		BlocksUntilNextEpoch:  blocksUntilNextEpoch,
	}, nil
}

// NodeEpochActivity queries a node's activity summary for a given epoch.
func (qs queryServer) NodeEpochActivity(ctx context.Context, req *types.QueryNodeEpochActivityRequest) (*types.QueryNodeEpochActivityResponse, error) {
	summary, err := qs.k.EpochSummary.Get(ctx, collections.Join(req.NodeId, req.Epoch))
	if err != nil {
		summary = types.EpochActivitySummary{}
	}

	quotaUsed, err := qs.k.EpochQuotaUsed.Get(ctx, collections.Join(req.NodeId, req.Epoch))
	if err != nil {
		quotaUsed = 0
	}

	params, err := qs.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Default to self-funded quota.
	quotaLimit := params.SelfFundedQuota

	return &types.QueryNodeEpochActivityResponse{
		Summary:    summary,
		QuotaUsed:  quotaUsed,
		QuotaLimit: quotaLimit,
	}, nil
}


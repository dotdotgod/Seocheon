package keeper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/activity/types"
)

// Activity queries a single activity by hash using GlobalHashIndex.
func (qs queryServer) Activity(ctx context.Context, req *types.QueryActivityRequest) (*types.QueryActivityResponse, error) {
	if req.ActivityHash == "" {
		return nil, types.ErrInvalidActivityHash
	}

	// Look up in global hash index.
	value, err := qs.k.GlobalHashIndex.Get(ctx, req.ActivityHash)
	if err != nil {
		return nil, types.ErrActivityNotFound
	}

	// Parse "node_id:epoch:sequence".
	nodeID, epoch, seq, err := parseGlobalHashValue(value)
	if err != nil {
		return nil, fmt.Errorf("corrupt global hash index: %w", err)
	}

	record, err := qs.k.Activities.Get(ctx, collections.Join3(nodeID, epoch, seq))
	if err != nil {
		return nil, types.ErrActivityNotFound
	}

	return &types.QueryActivityResponse{Activity: record}, nil
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

// parseGlobalHashValue parses "node_id:epoch:sequence" string.
func parseGlobalHashValue(value string) (string, int64, uint64, error) {
	parts := strings.SplitN(value, ":", 3)
	if len(parts) != 3 {
		return "", 0, 0, fmt.Errorf("invalid format: %s", value)
	}

	epoch, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid epoch: %w", err)
	}

	seq, err := strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("invalid sequence: %w", err)
	}

	return parts[0], epoch, seq, nil
}

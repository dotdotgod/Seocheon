package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"

	"seocheon/x/activity/keeper"
	"seocheon/x/activity/types"
)

func TestQueryActivity_ByHash(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	// Register a node and submit activity.
	f.nodeKeeper.registerNode("node1", "agent1", 2) // ACTIVE
	hash := generateHash(1)
	_, err := f.submitActivity(ctx, "agent1", hash, "ipfs://content1")
	require.NoError(t, err)

	qs := keeper.NewQueryServerImpl(f.keeper)

	// Query by hash.
	resp, err := qs.Activity(ctx, &types.QueryActivityRequest{ActivityHash: hash})
	require.NoError(t, err)
	require.Equal(t, "node1", resp.Activity.NodeId)
	require.Equal(t, hash, resp.Activity.ActivityHash)
	require.Equal(t, "ipfs://content1", resp.Activity.ContentUri)
}

func TestQueryActivity_ByHash_NotFound(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	qs := keeper.NewQueryServerImpl(f.keeper)

	_, err := qs.Activity(ctx, &types.QueryActivityRequest{ActivityHash: generateHash(999)})
	require.Error(t, err)
}

func TestQueryActivity_ByHash_EmptyHash(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	qs := keeper.NewQueryServerImpl(f.keeper)

	_, err := qs.Activity(ctx, &types.QueryActivityRequest{ActivityHash: ""})
	require.Error(t, err)
}

func TestQueryActivitiesByNode(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	f.nodeKeeper.registerNode("node1", "agent1", 2)
	f.nodeKeeper.registerNode("node2", "agent2", 2)

	// Submit 3 activities for node1 and 1 for node2.
	for i := 0; i < 3; i++ {
		_, err := f.submitActivity(ctx, "agent1", generateHash(i), "ipfs://c")
		require.NoError(t, err)
	}
	_, err := f.submitActivity(ctx, "agent2", generateHash(100), "ipfs://c")
	require.NoError(t, err)

	qs := keeper.NewQueryServerImpl(f.keeper)

	// Query node1 — should get 3.
	resp, err := qs.ActivitiesByNode(ctx, &types.QueryActivitiesByNodeRequest{NodeId: "node1"})
	require.NoError(t, err)
	require.Len(t, resp.Activities, 3)

	// Query node2 — should get 1.
	resp, err = qs.ActivitiesByNode(ctx, &types.QueryActivitiesByNodeRequest{NodeId: "node2"})
	require.NoError(t, err)
	require.Len(t, resp.Activities, 1)
}

func TestQueryActivitiesByNode_EpochFilter(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)
	params, _ := f.keeper.Params.Get(f.ctx)

	// Submit in epoch 1 (block EpochLength + 100).
	ctx1 := f.freshCtx(params.EpochLength + 100)
	_, err := f.submitActivity(ctx1, "agent1", generateHash(1), "ipfs://c")
	require.NoError(t, err)

	// Submit in epoch 2 (block 2*EpochLength + 100).
	ctx2 := f.freshCtx(2*params.EpochLength + 100)
	_, err = f.submitActivity(ctx2, "agent1", generateHash(2), "ipfs://c")
	require.NoError(t, err)

	qs := keeper.NewQueryServerImpl(f.keeper)

	// Query epoch 1 (epoch > 0, triggers filter).
	resp, err := qs.ActivitiesByNode(ctx1, &types.QueryActivitiesByNodeRequest{NodeId: "node1", Epoch: 1})
	require.NoError(t, err)
	require.Len(t, resp.Activities, 1)
	require.Equal(t, generateHash(1), resp.Activities[0].ActivityHash)

	// Query epoch 2.
	resp, err = qs.ActivitiesByNode(ctx2, &types.QueryActivitiesByNodeRequest{NodeId: "node1", Epoch: 2})
	require.NoError(t, err)
	require.Len(t, resp.Activities, 1)
	require.Equal(t, generateHash(2), resp.Activities[0].ActivityHash)

	// Query all epochs (epoch=0).
	resp, err = qs.ActivitiesByNode(ctx2, &types.QueryActivitiesByNodeRequest{NodeId: "node1", Epoch: 0})
	require.NoError(t, err)
	require.Len(t, resp.Activities, 2)
}

func TestQueryActivitiesByBlock(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)
	f.nodeKeeper.registerNode("node2", "agent2", 2)

	// Submit at block 500.
	ctx500 := f.freshCtx(500)
	_, err := f.submitActivity(ctx500, "agent1", generateHash(1), "ipfs://c")
	require.NoError(t, err)
	_, err = f.submitActivity(ctx500, "agent2", generateHash(2), "ipfs://c")
	require.NoError(t, err)

	// Submit at block 501.
	ctx501 := f.freshCtx(501)
	_, err = f.submitActivity(ctx501, "agent1", generateHash(3), "ipfs://c")
	require.NoError(t, err)

	qs := keeper.NewQueryServerImpl(f.keeper)

	// Block 500 should have 2.
	resp, err := qs.ActivitiesByBlock(ctx500, &types.QueryActivitiesByBlockRequest{BlockHeight: 500})
	require.NoError(t, err)
	require.Len(t, resp.Activities, 2)

	// Block 501 should have 1.
	resp, err = qs.ActivitiesByBlock(ctx501, &types.QueryActivitiesByBlockRequest{BlockHeight: 501})
	require.NoError(t, err)
	require.Len(t, resp.Activities, 1)
}

func TestQueryEpochInfo(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	// Block 100 should be in epoch 0.
	ctx := f.freshCtx(100)
	qs := keeper.NewQueryServerImpl(f.keeper)

	resp, err := qs.EpochInfo(ctx, &types.QueryEpochInfoRequest{})
	require.NoError(t, err)
	require.Equal(t, int64(0), resp.CurrentEpoch)
	require.True(t, resp.BlocksUntilNextEpoch > 0)

	// Epoch 1 boundary.
	ctx2 := f.freshCtx(params.EpochLength + 1)
	resp2, err := qs.EpochInfo(ctx2, &types.QueryEpochInfoRequest{})
	require.NoError(t, err)
	require.Equal(t, int64(1), resp2.CurrentEpoch)
}

func TestQueryNodeEpochActivity(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit 3 activities.
	for i := 0; i < 3; i++ {
		_, err := f.submitActivity(ctx, "agent1", generateHash(i), "ipfs://c")
		require.NoError(t, err)
	}

	qs := keeper.NewQueryServerImpl(f.keeper)

	resp, err := qs.NodeEpochActivity(ctx, &types.QueryNodeEpochActivityRequest{NodeId: "node1", Epoch: 0})
	require.NoError(t, err)
	require.Equal(t, uint64(3), resp.Summary.TotalActivities)
	require.Equal(t, uint64(3), resp.QuotaUsed)
	require.Equal(t, uint64(100), resp.QuotaLimit) // DefaultParams().SelfFundedQuota

	// Node with no activity.
	resp2, err := qs.NodeEpochActivity(ctx, &types.QueryNodeEpochActivityRequest{NodeId: "nonexistent", Epoch: 0})
	require.NoError(t, err)
	require.Equal(t, uint64(0), resp2.Summary.TotalActivities)
	require.Equal(t, uint64(0), resp2.QuotaUsed)
}

func TestQueryParams(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	qs := keeper.NewQueryServerImpl(f.keeper)

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), resp.Params)
}

// Verify unused variable warning is fixed.
func TestQueryActivitiesByNode_EmptyNodeID(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	qs := keeper.NewQueryServerImpl(f.keeper)
	_, err := qs.ActivitiesByNode(ctx, &types.QueryActivitiesByNodeRequest{NodeId: ""})
	require.Error(t, err)
}

// Verify Activity query finds record via HashIndex scan.
func TestQueryActivity_HashIndexScan(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	f.nodeKeeper.registerNode("node1", "agent1", 2)
	hash := generateHash(42)
	_, err := f.submitActivity(ctx, "agent1", hash, "ipfs://test")
	require.NoError(t, err)

	// Verify HashIndex contains the entry.
	has, err := f.keeper.HashIndex.Has(ctx, collections.Join3("node1", int64(0), hash))
	require.NoError(t, err)
	require.True(t, has)

	// Query service should find the record via HashIndex scan.
	qs := keeper.NewQueryServerImpl(f.keeper)
	resp, err := qs.Activity(ctx, &types.QueryActivityRequest{ActivityHash: hash})
	require.NoError(t, err)
	require.Equal(t, int64(100), resp.Activity.BlockHeight)
}

// Ensure BlockActivities is populated for ActivitiesByBlock.
func TestQueryActivitiesByBlock_Empty(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(100)

	qs := keeper.NewQueryServerImpl(f.keeper)
	resp, err := qs.ActivitiesByBlock(ctx, &types.QueryActivitiesByBlockRequest{BlockHeight: 999})
	require.NoError(t, err)
	require.Empty(t, resp.Activities)
}

// Unused variables in query_activity.go.
var _ = collections.NewPrefixedTripleRange[string, int64, uint64]

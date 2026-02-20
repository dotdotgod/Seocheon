package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"

	"seocheon/x/activity/types"
)

func TestGenesis_DefaultRoundtrip(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// Export default genesis.
	genState, err := f.keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultParams(), genState.Params)
	require.Empty(t, genState.Activities)

	// Re-initialize from exported genesis.
	err = f.keeper.InitGenesis(ctx, *genState)
	require.NoError(t, err)

	// Export again and verify.
	genState2, err := f.keeper.ExportGenesis(ctx)
	require.NoError(t, err)
	require.Equal(t, genState.Params, genState2.Params)
}

func TestGenesis_WithActivities(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)
	f.nodeKeeper.registerNode("node2", "agent2", 2)

	// Submit activities in different blocks/epochs.
	ctx100 := f.freshCtx(100)
	_, err := f.submitActivity(ctx100, "agent1", generateHash(1), "ipfs://c1")
	require.NoError(t, err)

	ctx200 := f.freshCtx(200)
	_, err = f.submitActivity(ctx200, "agent1", generateHash(2), "ipfs://c2")
	require.NoError(t, err)

	ctx300 := f.freshCtx(300)
	_, err = f.submitActivity(ctx300, "agent2", generateHash(3), "ipfs://c3")
	require.NoError(t, err)

	// Export.
	genState, err := f.keeper.ExportGenesis(ctx300)
	require.NoError(t, err)
	require.Len(t, genState.Activities, 3)

	// Init a fresh fixture with exported genesis.
	f2 := initFixture(t)
	f2.nodeKeeper.registerNode("node1", "agent1", 2)
	f2.nodeKeeper.registerNode("node2", "agent2", 2)
	ctx2 := f2.freshCtx(1)

	err = f2.keeper.InitGenesis(ctx2, *genState)
	require.NoError(t, err)

	// Verify activities are restored.
	genState2, err := f2.keeper.ExportGenesis(ctx2)
	require.NoError(t, err)
	require.Len(t, genState2.Activities, 3)

	// Verify HashIndex is rebuilt.
	for _, act := range genState.Activities {
		has, err := f2.keeper.HashIndex.Has(ctx2, collections.Join3(act.NodeId, act.Epoch, act.ActivityHash))
		require.NoError(t, err)
		require.True(t, has, "HashIndex missing for %s", act.ActivityHash)
	}

	// Verify GlobalHashIndex is rebuilt.
	for _, act := range genState.Activities {
		_, err := f2.keeper.GlobalHashIndex.Get(ctx2, act.ActivityHash)
		require.NoError(t, err, "GlobalHashIndex missing for %s", act.ActivityHash)
	}

	// Verify EpochQuotaUsed is rebuilt.
	quotaUsed, err := f2.keeper.EpochQuotaUsed.Get(ctx2, collections.Join("node1", int64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(2), quotaUsed) // node1 has 2 activities in epoch 0

	quotaUsed2, err := f2.keeper.EpochQuotaUsed.Get(ctx2, collections.Join("node2", int64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(1), quotaUsed2)
}

func TestGenesis_IndexRebuild(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit an activity.
	ctx := f.freshCtx(100)
	hash := generateHash(42)
	_, err := f.submitActivity(ctx, "agent1", hash, "ipfs://test")
	require.NoError(t, err)

	// Export.
	genState, err := f.keeper.ExportGenesis(ctx)
	require.NoError(t, err)

	// Init fresh.
	f2 := initFixture(t)
	ctx2 := f2.freshCtx(1)
	err = f2.keeper.InitGenesis(ctx2, *genState)
	require.NoError(t, err)

	// Verify ActivitySequence is rebuilt.
	seq, err := f2.keeper.ActivitySequence.Get(ctx2, collections.Join("node1", int64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(1), seq) // Next sequence after 0.

	// Verify BlockActivities is rebuilt.
	nodeID, err := f2.keeper.BlockActivities.Get(ctx2, collections.Join(int64(100), uint64(0)))
	require.NoError(t, err)
	require.Equal(t, "node1", nodeID)

	// Verify WindowActivity is rebuilt.
	params, _ := f2.keeper.Params.Get(ctx2)
	window := int64(0) // block 100 is in window 0 (100 / (17280/12) = 0)
	windowCount, err := f2.keeper.WindowActivity.Get(ctx2, collections.Join3("node1", int64(0), window))
	require.NoError(t, err)
	require.Equal(t, uint64(1), windowCount)

	// Verify EpochSummary is rebuilt.
	summary, err := f2.keeper.EpochSummary.Get(ctx2, collections.Join("node1", int64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(1), summary.TotalActivities)
	require.Equal(t, uint64(1), summary.ActiveWindows)

	_ = params
}

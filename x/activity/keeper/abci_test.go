package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"github.com/stretchr/testify/require"

	"seocheon/x/activity/types"
)

func TestEndBlocker_NonBoundary(t *testing.T) {
	f := initFixture(t)
	// Block 100 is not a window or epoch boundary.
	ctx := f.freshCtx(100)

	err := f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	// No events should be emitted.
	require.Equal(t, 0, countEvents(ctx, types.EventTypeWindowCompleted))
	require.Equal(t, 0, countEvents(ctx, types.EventTypeEpochCompleted))
}

func TestEndBlocker_WindowBoundary(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)
	windowLength := params.EpochLength / params.WindowsPerEpoch

	// Block exactly at window boundary.
	ctx := f.freshCtx(windowLength)

	err := f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	event := requireEvent(t, ctx, types.EventTypeWindowCompleted)
	require.Equal(t, "0", eventAttribute(event, types.AttributeKeyEpoch))
	require.Equal(t, "0", eventAttribute(event, types.AttributeKeyWindow))
}

func TestEndBlocker_EpochBoundary(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	// Submit activities in 8 windows to make "node1" eligible.
	f.nodeKeeper.registerNode("node1", "agent1", 2)
	windowLength := params.EpochLength / params.WindowsPerEpoch

	for w := int64(0); w < 8; w++ {
		blockInWindow := w*windowLength + 1
		ctx := f.freshCtx(blockInWindow)
		_, err := f.submitActivity(ctx, "agent1", generateHash(int(w)), "ipfs://c")
		require.NoError(t, err)
	}

	// Run EndBlocker at epoch boundary.
	ctx := f.freshCtx(params.EpochLength)
	err := f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Both window_completed (since epoch boundary is also window boundary) and epoch_completed.
	event := requireEvent(t, ctx, types.EventTypeEpochCompleted)
	require.Equal(t, "0", eventAttribute(event, types.AttributeKeyEpoch))
	require.Equal(t, "1", eventAttribute(event, types.AttributeKeyEligibleCount))
}

func TestEndBlocker_EpochBoundary_NoEligible(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	// No activities submitted at all.
	ctx := f.freshCtx(params.EpochLength)
	err := f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	event := requireEvent(t, ctx, types.EventTypeEpochCompleted)
	require.Equal(t, "0", eventAttribute(event, types.AttributeKeyEligibleCount))
}

func TestEndBlocker_Pruning(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	// Set a shorter pruning window for testing.
	params.ActivityPruningKeepBlocks = 100
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit activity at block 10.
	ctx10 := f.freshCtx(10)
	_, err := f.submitActivity(ctx10, "agent1", generateHash(1), "ipfs://c")
	require.NoError(t, err)

	// Verify it exists.
	_, err = f.keeper.BlockActivities.Get(ctx10, collections.Join(int64(10), uint64(0)))
	require.NoError(t, err)

	// Run EndBlocker at a block far enough to trigger pruning.
	// cutoffHeight = 17280 - 100 = 17180; block 10 < 17180 → prunable.
	ctx := f.freshCtx(params.EpochLength) // epoch boundary triggers pruning
	err = f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	// The activity at block 10 should be pruned.
	_, err = f.keeper.BlockActivities.Get(ctx, collections.Join(int64(10), uint64(0)))
	require.Error(t, err) // Should be gone.

	// Pruning event should be emitted.
	requireEvent(t, ctx, types.EventTypeActivityPruned)
}

func TestEndBlocker_Pruning_RecentNotDeleted(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit at a block within the keep window.
	recentBlock := params.EpochLength - 10
	ctxRecent := f.freshCtx(recentBlock)
	_, err := f.submitActivity(ctxRecent, "agent1", generateHash(1), "ipfs://c")
	require.NoError(t, err)

	// Run EndBlocker at epoch boundary.
	// cutoffHeight = EpochLength - pruning_keep_blocks = 17280 - 1555200 < 0 → no pruning.
	ctx := f.freshCtx(params.EpochLength)
	err = f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Activity should still exist.
	_, err = f.keeper.BlockActivities.Get(ctx, collections.Join(recentBlock, uint64(0)))
	require.NoError(t, err)
}

func TestEndBlocker_Pruning_Bounded(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	// Short pruning keep and high quota to trigger bounded pruning.
	params.ActivityPruningKeepBlocks = 100
	params.SelfFundedQuota = 2000
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit 1050 activities at low block heights (all in epoch 0, different blocks).
	for i := 0; i < 1050; i++ {
		ctx := f.freshCtx(int64(i + 1))
		_, err := f.submitActivity(ctx, "agent1", generateHash(i), "ipfs://c")
		require.NoError(t, err)
	}

	// Run EndBlocker — should prune max 1000.
	ctx := f.freshCtx(params.EpochLength)
	err := f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	event := requireEvent(t, ctx, types.EventTypeActivityPruned)
	require.Equal(t, "1000", eventAttribute(event, types.AttributeKeyPrunedCount))
}

func TestEndBlocker_MultipleWindowBoundaries(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)
	windowLength := params.EpochLength / params.WindowsPerEpoch

	// Test multiple window boundaries in sequence.
	for w := int64(1); w <= 3; w++ {
		ctx := f.freshCtx(w * windowLength)
		err := f.keeper.EndBlocker(ctx)
		require.NoError(t, err)

		event := requireEvent(t, ctx, types.EventTypeWindowCompleted)
		require.NotEmpty(t, eventAttribute(event, types.AttributeKeyEpoch))
	}
}

// Verify pruning cleans up all related indices.
func TestEndBlocker_Pruning_CleanupIndices(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	params.ActivityPruningKeepBlocks = 100
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit activity at block 10.
	ctx10 := f.freshCtx(10)
	hash := generateHash(42)
	_, err := f.submitActivity(ctx10, "agent1", hash, "ipfs://c")
	require.NoError(t, err)

	// Verify HashIndex exists.
	has, err := f.keeper.HashIndex.Has(ctx10, collections.Join3("node1", int64(0), hash))
	require.NoError(t, err)
	require.True(t, has)

	// Run EndBlocker at epoch boundary to trigger pruning.
	ctx := f.freshCtx(params.EpochLength)
	err = f.keeper.EndBlocker(ctx)
	require.NoError(t, err)

	// Verify indices are cleaned up.
	has, err = f.keeper.HashIndex.Has(ctx, collections.Join3("node1", int64(0), hash))
	require.NoError(t, err)
	require.False(t, has)

	// Check that activity record itself is pruned.
	_, err = f.keeper.Activities.Get(ctx, collections.Join3("node1", int64(0), uint64(0)))
	require.Error(t, err) // Should be gone.
}

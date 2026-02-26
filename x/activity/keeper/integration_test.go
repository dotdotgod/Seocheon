package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/activity/keeper"
	"seocheon/x/activity/types"
)

// TestIntegration_FullLifecycle tests the complete activity lifecycle:
// submit activities across multiple windows → verify window tracking →
// reach epoch boundary → epoch_completed event → quota boundary → pruning → invariants.
func TestIntegration_FullLifecycle(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)
	windowLength := params.EpochLength / params.WindowsPerEpoch

	f.nodeKeeper.registerNode("node1", "agent1", 2)
	f.nodeKeeper.registerNode("node2", "agent2", 2)

	// --- Phase 1: Submit activities across 8+ windows for node1 ---
	hashIdx := 0
	for w := int64(0); w < 10; w++ {
		blockInWindow := w*windowLength + 1
		ctx := f.freshCtx(blockInWindow)

		// node1 submits in every window.
		_, err := f.submitActivity(ctx, "agent1", generateHash(hashIdx), "ipfs://n1")
		require.NoError(t, err)
		hashIdx++

		// node2 submits in only 5 windows (not enough for eligibility).
		if w < 5 {
			_, err = f.submitActivity(ctx, "agent2", generateHash(hashIdx), "ipfs://n2")
			require.NoError(t, err)
			hashIdx++
		}
	}

	// --- Phase 2: Verify window tracking ---
	ctxCheck := f.freshCtx(1)

	// node1: 10 active windows → eligible.
	summary1, err := f.keeper.EpochSummary.Get(ctxCheck, collections.Join("node1", int64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(10), summary1.ActiveWindows)
	require.True(t, summary1.Eligible)

	// node2: 5 active windows → NOT eligible (min 8).
	summary2, err := f.keeper.EpochSummary.Get(ctxCheck, collections.Join("node2", int64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(5), summary2.ActiveWindows)
	require.False(t, summary2.Eligible)

	// --- Phase 3: Epoch boundary ---
	ctxEpoch := f.freshCtx(params.EpochLength)
	err = f.keeper.EndBlocker(ctxEpoch)
	require.NoError(t, err)

	// Verify epoch_completed event with eligible_count=1 (only node1).
	event := requireEvent(t, ctxEpoch, types.EventTypeEpochCompleted)
	require.Equal(t, "1", eventAttribute(event, types.AttributeKeyEligibleCount))

	// --- Phase 4: Verify quota enforcement ---
	// Submit activities until quota is reached.
	params2, _ := f.keeper.Params.Get(f.ctx)
	quotaUsed, err := f.keeper.EpochQuotaUsed.Get(ctxCheck, collections.Join("node1", int64(0)))
	require.NoError(t, err)
	require.Equal(t, uint64(10), quotaUsed) // 10 activities submitted.

	// Submit more until quota.
	remaining := int(params2.SelfFundedQuota) - int(quotaUsed)
	for i := 0; i < remaining; i++ {
		ctx := f.freshCtx(int64(i + 1000)) // different blocks
		_, err := f.submitActivity(ctx, "agent1", generateHash(hashIdx), "ipfs://c")
		require.NoError(t, err)
		hashIdx++
	}

	// Next submission should fail.
	ctxOverQuota := f.freshCtx(2000)
	_, err = f.submitActivity(ctxOverQuota, "agent1", generateHash(hashIdx), "ipfs://c")
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrQuotaExceeded)

	// --- Phase 5: Run all invariants ---
	sdkCtx := sdk.UnwrapSDKContext(f.freshCtx(1))
	msg, broken := keeper.AllInvariants(f.keeper)(sdkCtx)
	require.False(t, broken, "invariants should pass: %s", msg)

	// --- Phase 6: Genesis round-trip ---
	genState, err := f.keeper.ExportGenesis(f.freshCtx(1))
	require.NoError(t, err)
	require.Greater(t, len(genState.Activities), 0)

	f2 := initFixture(t)
	ctx2 := f2.freshCtx(1)
	err = f2.keeper.InitGenesis(ctx2, *genState)
	require.NoError(t, err)

	genState2, err := f2.keeper.ExportGenesis(ctx2)
	require.NoError(t, err)
	require.Equal(t, len(genState.Activities), len(genState2.Activities))
}

// TestIntegration_CrossEpochBehavior tests global (hash, uri) uniqueness across epochs.
func TestIntegration_CrossEpochBehavior(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	hash := generateHash(1)
	uri := "ipfs://e0"

	// Submit in epoch 0.
	ctx0 := f.freshCtx(100)
	_, err := f.submitActivity(ctx0, "agent1", hash, uri)
	require.NoError(t, err)

	// Same hash + same URI in epoch 0 → should fail (global duplicate).
	ctx0b := f.freshCtx(200)
	_, err = f.submitActivity(ctx0b, "agent1", hash, uri)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDuplicateActivityHash)

	// Same hash + different URI in epoch 0 → allowed (hash collision resolution).
	_, err = f.submitActivity(ctx0b, "agent1", hash, "ipfs://e0alt")
	require.NoError(t, err)

	// Same hash + same URI in epoch 1 → still rejected (global duplicate).
	ctx1 := f.freshCtx(params.EpochLength + 100)
	_, err = f.submitActivity(ctx1, "agent1", hash, uri)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrDuplicateActivityHash)

	// Same hash + new URI in epoch 1 → allowed.
	_, err = f.submitActivity(ctx1, "agent1", hash, "ipfs://e1")
	require.NoError(t, err)
}

// TestIntegration_EligibilityCheck tests the keeper's eligibility methods.
func TestIntegration_EligibilityCheck(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)
	windowLength := params.EpochLength / params.WindowsPerEpoch

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit in 8 windows → eligible.
	for w := int64(0); w < 8; w++ {
		ctx := f.freshCtx(w*windowLength + 1)
		_, err := f.submitActivity(ctx, "agent1", generateHash(int(w)), "ipfs://c")
		require.NoError(t, err)
	}

	eligible, err := f.keeper.IsNodeEligibleForEpoch(f.freshCtx(1), "node1", 0)
	require.NoError(t, err)
	require.True(t, eligible)

	// Non-existent node → not eligible (no error).
	eligible2, err := f.keeper.IsNodeEligibleForEpoch(f.freshCtx(1), "nonexistent", 0)
	require.NoError(t, err)
	require.False(t, eligible2)
}

// TestIntegration_CountEligibleEpochs tests the lookback count method.
func TestIntegration_CountEligibleEpochs(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)
	windowLength := params.EpochLength / params.WindowsPerEpoch

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Make node1 eligible in epoch 0: 8 windows.
	for w := int64(0); w < 8; w++ {
		ctx := f.freshCtx(w*windowLength + 1)
		_, err := f.submitActivity(ctx, "agent1", generateHash(int(w)), "ipfs://c")
		require.NoError(t, err)
	}

	// Make node1 eligible in epoch 1: 8 windows.
	for w := int64(0); w < 8; w++ {
		ctx := f.freshCtx(params.EpochLength + w*windowLength + 1)
		_, err := f.submitActivity(ctx, "agent1", generateHash(100+int(w)), "ipfs://c")
		require.NoError(t, err)
	}

	// Count eligible epochs looking back from epoch 3 with lookback 3.
	count, err := f.keeper.CountEligibleEpochs(f.freshCtx(1), "node1", 3, 3)
	require.NoError(t, err)
	require.Equal(t, int64(2), count) // Epochs 0 and 1 are eligible.
}

// TestIntegration_PruningWithResubmission tests that after pruning, the same hash
// After pruning (TTL expired), both Activities and HashIndex are removed,
// so the same (hash, uri) pair can be resubmitted.
func TestIntegration_PruningWithResubmission(t *testing.T) {
	f := initFixture(t)
	params, _ := f.keeper.Params.Get(f.ctx)

	// Short pruning window for testing.
	params.ActivityPruningKeepBlocks = 100
	require.NoError(t, f.keeper.Params.Set(f.ctx, params))

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	hash := generateHash(1)
	uri := "ipfs://original"

	// Submit in epoch 0 at block 10.
	ctx10 := f.freshCtx(10)
	_, err := f.submitActivity(ctx10, "agent1", hash, uri)
	require.NoError(t, err)

	// Run EndBlocker to prune (both Activities and HashIndex).
	ctxEpoch := f.freshCtx(params.EpochLength)
	err = f.keeper.EndBlocker(ctxEpoch)
	require.NoError(t, err)

	// HashIndex is pruned along with Activities.
	has, err := f.keeper.HashIndex.Has(ctxEpoch, collections.Join(hash, uri))
	require.NoError(t, err)
	require.False(t, has)

	// After pruning, the same (hash, uri) can be resubmitted.
	ctx1 := f.freshCtx(params.EpochLength + 100)
	_, err = f.submitActivity(ctx1, "agent1", hash, uri)
	require.NoError(t, err)
}

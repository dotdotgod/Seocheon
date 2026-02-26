package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/activity/keeper"
)

func TestInvariants_AllPass(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	// Submit a few activities.
	ctx := f.freshCtx(100)
	for i := 0; i < 5; i++ {
		_, err := f.submitActivity(ctx, "agent1", generateHash(i), "ipfs://c")
		require.NoError(t, err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msg, broken := keeper.AllInvariants(f.keeper)(sdkCtx)
	require.False(t, broken, "invariants should pass: %s", msg)
}

func TestInvariant_HashIndexConsistency_Broken(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	ctx := f.freshCtx(100)
	hash := generateHash(1)
	_, err := f.submitActivity(ctx, "agent1", hash, "ipfs://c")
	require.NoError(t, err)

	// Corrupt: remove hash index entry.
	err = f.keeper.HashIndex.Remove(ctx, collections.Join(hash, "ipfs://c"))
	require.NoError(t, err)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msg, broken := keeper.HashIndexConsistencyInvariant(f.keeper)(sdkCtx)
	require.True(t, broken, "invariant should be broken")
	require.Contains(t, msg, "missing from hash index")
}

func TestInvariant_QuotaConsistency_Broken(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, "agent1", generateHash(1), "ipfs://c")
	require.NoError(t, err)

	// Corrupt: set quota to wrong value.
	err = f.keeper.EpochQuotaUsed.Set(ctx, collections.Join("node1", int64(0)), 999)
	require.NoError(t, err)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msg, broken := keeper.QuotaConsistencyInvariant(f.keeper)(sdkCtx)
	require.True(t, broken, "invariant should be broken")
	require.Contains(t, msg, "quota mismatch")
}

func TestInvariant_WindowActivityConsistency_Broken(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, "agent1", generateHash(1), "ipfs://c")
	require.NoError(t, err)

	// Corrupt: set window activity to wrong value.
	err = f.keeper.WindowActivity.Set(ctx, collections.Join3("node1", int64(0), int64(0)), 999)
	require.NoError(t, err)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msg, broken := keeper.WindowActivityConsistencyInvariant(f.keeper)(sdkCtx)
	require.True(t, broken, "invariant should be broken")
	require.Contains(t, msg, "window activity mismatch")
}

func TestInvariants_MultiNode(t *testing.T) {
	f := initFixture(t)

	f.nodeKeeper.registerNode("node1", "agent1", 2)
	f.nodeKeeper.registerNode("node2", "agent2", 2)

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, "agent1", generateHash(1), "ipfs://c")
	require.NoError(t, err)
	_, err = f.submitActivity(ctx, "agent2", generateHash(2), "ipfs://c")
	require.NoError(t, err)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	msg, broken := keeper.AllInvariants(f.keeper)(sdkCtx)
	require.False(t, broken, "invariants should pass: %s", msg)
}

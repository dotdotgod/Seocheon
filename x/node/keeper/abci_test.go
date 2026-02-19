package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestEndBlocker_AppliesPendingAgentShareChange(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_end1_____________")).String()
	agent := sdk.AccAddress([]byte("ag_end1_____________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Schedule agent share change.
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(35),
	})
	require.NoError(t, err)

	// Verify pending exists.
	has, err := f.keeper.PendingAgentShareChanges.Has(ctx, nodeID)
	require.NoError(t, err)
	require.True(t, has)

	// EndBlocker at non-epoch boundary: nothing happens.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	midCtx := sdkCtx.WithBlockHeight(types.EpochLength / 2)
	err = f.keeper.EndBlocker(midCtx)
	require.NoError(t, err)

	// Pending should still exist.
	has, err = f.keeper.PendingAgentShareChanges.Has(ctx, nodeID)
	require.NoError(t, err)
	require.True(t, has)

	// Node's agent_share should still be 30.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, math.LegacyNewDec(30), node.AgentShare)

	// EndBlocker at epoch boundary.
	epochCtx := sdkCtx.WithBlockHeight(types.EpochLength)
	err = f.keeper.EndBlocker(epochCtx)
	require.NoError(t, err)

	// Pending should be removed.
	has, err = f.keeper.PendingAgentShareChanges.Has(ctx, nodeID)
	require.NoError(t, err)
	require.False(t, has)

	// Node's agent_share should be updated to 35.
	node, err = f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, math.LegacyNewDec(35), node.AgentShare)
}

func TestEndBlocker_SkipsNonEpochBoundary(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx

	// No pending changes, just verify no errors at random block.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	midCtx := sdkCtx.WithBlockHeight(100)
	err := f.keeper.EndBlocker(midCtx)
	require.NoError(t, err)
}

func TestEndBlocker_SkipsFutureChanges(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_end2_____________")).String()
	agent := sdk.AccAddress([]byte("ag_end2_____________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Schedule change (will be set to apply at EpochLength).
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(25),
	})
	require.NoError(t, err)

	// Manually set the apply_at_block to a future epoch.
	pending, _ := f.keeper.PendingAgentShareChanges.Get(ctx, nodeID)
	pending.ApplyAtBlock = types.EpochLength * 3 // far future
	err = f.keeper.PendingAgentShareChanges.Set(ctx, nodeID, pending)
	require.NoError(t, err)

	// EndBlocker at first epoch: change should NOT be applied.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	epochCtx := sdkCtx.WithBlockHeight(types.EpochLength)
	err = f.keeper.EndBlocker(epochCtx)
	require.NoError(t, err)

	// Pending should still exist.
	has, err := f.keeper.PendingAgentShareChanges.Has(ctx, nodeID)
	require.NoError(t, err)
	require.True(t, has)

	// Node's agent_share should remain 30.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, math.LegacyNewDec(30), node.AgentShare)
}

func TestEndBlocker_MultipleChanges(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator1 := sdk.AccAddress([]byte("op_end3a____________")).String()
	agent1 := sdk.AccAddress([]byte("ag_end3a____________")).String()
	operator2 := sdk.AccAddress([]byte("op_end3b____________")).String()
	agent2 := sdk.AccAddress([]byte("ag_end3b____________")).String()
	nodeID1 := registerTestNode(t, f, operator1, agent1)
	nodeID2 := registerTestNode(t, f, operator2, agent2)

	// Schedule changes for both.
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator1,
		NewAgentShare: math.LegacyNewDec(35),
	})
	require.NoError(t, err)

	_, err = msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator2,
		NewAgentShare: math.LegacyNewDec(25),
	})
	require.NoError(t, err)

	// EndBlocker at epoch boundary.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	epochCtx := sdkCtx.WithBlockHeight(types.EpochLength)
	err = f.keeper.EndBlocker(epochCtx)
	require.NoError(t, err)

	// Both should be applied.
	node1, err := f.keeper.Nodes.Get(ctx, nodeID1)
	require.NoError(t, err)
	require.Equal(t, math.LegacyNewDec(35), node1.AgentShare)

	node2, err := f.keeper.Nodes.Get(ctx, nodeID2)
	require.NoError(t, err)
	require.Equal(t, math.LegacyNewDec(25), node2.AgentShare)

	// Pending should be cleared.
	has1, _ := f.keeper.PendingAgentShareChanges.Has(ctx, nodeID1)
	has2, _ := f.keeper.PendingAgentShareChanges.Has(ctx, nodeID2)
	require.False(t, has1)
	require.False(t, has2)
}

func TestEndBlocker_DeletedNodePendingChange(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_end4_____________")).String()
	agent := sdk.AccAddress([]byte("ag_end4_____________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Schedule agent share change.
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(35),
	})
	require.NoError(t, err)

	// Now deactivate — this removes the pending change.
	_, err = msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{Operator: operator})
	require.NoError(t, err)

	// But let's manually re-add a stale pending change to test EndBlocker resilience.
	err = f.keeper.PendingAgentShareChanges.Set(ctx, nodeID, types.PendingAgentShareChange{
		NodeId:        nodeID,
		NewAgentShare: math.LegacyNewDec(99),
		ApplyAtBlock:  types.EpochLength,
	})
	require.NoError(t, err)

	// Delete the node from Nodes store to simulate a truly deleted node.
	err = f.keeper.Nodes.Remove(ctx, nodeID)
	require.NoError(t, err)

	// EndBlocker at epoch: should handle the missing node gracefully.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	epochCtx := sdkCtx.WithBlockHeight(types.EpochLength)
	err = f.keeper.EndBlocker(epochCtx)
	require.NoError(t, err)

	// Stale pending change should be cleaned up.
	has, err := f.keeper.PendingAgentShareChanges.Has(ctx, nodeID)
	require.NoError(t, err)
	require.False(t, has)
}

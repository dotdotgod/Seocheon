package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestDeactivateNode_Success(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_deact1___________")).String()
	agent := sdk.AccAddress([]byte("ag_deact1___________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	require.NoError(t, err)

	// Verify status is INACTIVE.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, types.NodeStatus_NODE_STATUS_INACTIVE, node.Status)
}

func TestDeactivateNode_NotFound(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_deact_notfound___")).String()
	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	require.ErrorIs(t, err, types.ErrNodeNotFound)
}

func TestDeactivateNode_AlreadyInactive(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_deact2___________")).String()
	agent := sdk.AccAddress([]byte("ag_deact2___________")).String()
	registerTestNode(t, f, operator, agent)

	// First deactivation.
	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	require.NoError(t, err)

	// Second deactivation should fail.
	_, err = msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	require.ErrorIs(t, err, types.ErrNodeInactive)
}

func TestDeactivateNode_UndelegateFails_BestEffort(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_deact_undfail_____")).String()
	agent := sdk.AccAddress([]byte("ag_deact_undfail_____")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Set undelegate to fail.
	f.stakingMsgServer.undelegateErr = fmt.Errorf("mock undelegate failure")

	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	// Should succeed despite undelegate failure (best-effort).
	require.NoError(t, err)

	// Node should still be INACTIVE.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, types.NodeStatus_NODE_STATUS_INACTIVE, node.Status)

	// undelegate_failed event should have been emitted.
	evt := requireEvent(t, ctx, types.EventTypeUndelegateFailed)
	require.Equal(t, nodeID, eventAttribute(evt, types.AttributeKeyNodeID))
	require.Contains(t, eventAttribute(evt, types.AttributeKeyError), "mock undelegate failure")

	// node_deactivated event also emitted.
	requireEvent(t, ctx, types.EventTypeNodeDeactivated)
}

func TestDeactivateNode_NilStakingMsgServer(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx

	operator := sdk.AccAddress([]byte("op_deact_nosms______")).String()
	agent := sdk.AccAddress([]byte("ag_deact_nosms______")).String()

	// Register node normally (with staking msg server).
	nodeID := registerTestNode(t, f, operator, agent)

	// Now clear staking msg server and create new MsgServer for deactivation.
	f.keeper.SetStakingMsgServer(nil)
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	require.NoError(t, err)

	// Node should be INACTIVE (undelegate skipped silently).
	deactivated, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, types.NodeStatus_NODE_STATUS_INACTIVE, deactivated.Status)

	// No undelegate_failed event should exist (staking msg server was nil).
	requireNoEvent(t, ctx, types.EventTypeUndelegateFailed)
}

func TestDeactivateNode_EmitsDeactivatedEvent(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_deact_evt________")).String()
	agent := sdk.AccAddress([]byte("ag_deact_evt________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	require.NoError(t, err)

	evt := requireEvent(t, ctx, types.EventTypeNodeDeactivated)
	require.Equal(t, nodeID, eventAttribute(evt, types.AttributeKeyNodeID))
	require.Equal(t, operator, eventAttribute(evt, types.AttributeKeyOperator))
}

func TestDeactivateNode_ClearsPendingAgentShareChange(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_deact3___________")).String()
	agent := sdk.AccAddress([]byte("ag_deact3___________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Schedule an agent share change first.
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: types.DefaultAgentShareForTest(),
	})
	require.NoError(t, err)

	// Verify pending change exists.
	has, err := f.keeper.PendingAgentShareChanges.Has(ctx, nodeID)
	require.NoError(t, err)
	require.True(t, has)

	// Deactivate.
	_, err = msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{
		Operator: operator,
	})
	require.NoError(t, err)

	// Pending change should be removed.
	has, err = f.keeper.PendingAgentShareChanges.Has(ctx, nodeID)
	require.NoError(t, err)
	require.False(t, has)
}

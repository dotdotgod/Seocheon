package keeper_test

import (
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

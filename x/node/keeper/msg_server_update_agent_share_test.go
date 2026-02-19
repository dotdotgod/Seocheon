package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestUpdateNodeAgentShare_Success(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_share1___________")).String()
	agent := sdk.AccAddress([]byte("ag_share1___________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Current agent_share = 30, MaxChangeRate = 10.
	// Request change to 35 (within rate).
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(35),
	})
	require.NoError(t, err)

	// Pending change should exist.
	pending, err := f.keeper.PendingAgentShareChanges.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, math.LegacyNewDec(35), pending.NewAgentShare)
	require.Equal(t, types.EpochLength, pending.ApplyAtBlock) // block 0 → next epoch = EpochLength
}

func TestUpdateNodeAgentShare_ExceedsMaxRate(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_share2___________")).String()
	agent := sdk.AccAddress([]byte("ag_share2___________")).String()
	registerTestNode(t, f, operator, agent)

	// Current agent_share = 30, MaxChangeRate = 10.
	// Request change to 50 (exceeds rate by 10).
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(50),
	})
	require.ErrorIs(t, err, types.ErrAgentShareChangeExceedsMax)
}

func TestUpdateNodeAgentShare_OutOfRange(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_share3___________")).String()
	agent := sdk.AccAddress([]byte("ag_share3___________")).String()
	registerTestNode(t, f, operator, agent)

	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(101),
	})
	require.ErrorIs(t, err, types.ErrInvalidAgentShare)

	_, err = msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(-1),
	})
	require.ErrorIs(t, err, types.ErrInvalidAgentShare)
}

func TestUpdateNodeAgentShare_NotFound(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_notfound_________")).String()
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(50),
	})
	require.ErrorIs(t, err, types.ErrNodeNotFound)
}

func TestUpdateNodeAgentShare_DuplicatePending(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_share4___________")).String()
	agent := sdk.AccAddress([]byte("ag_share4___________")).String()
	registerTestNode(t, f, operator, agent)

	// First change request: OK.
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(35),
	})
	require.NoError(t, err)

	// Second change request while pending: should fail.
	_, err = msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(25),
	})
	require.ErrorIs(t, err, types.ErrAgentShareChangeExceedsMax)
}

func TestUpdateNodeAgentShare_InactiveNode(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_share5___________")).String()
	agent := sdk.AccAddress([]byte("ag_share5___________")).String()
	registerTestNode(t, f, operator, agent)

	// Deactivate.
	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{Operator: operator})
	require.NoError(t, err)

	_, err = msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(35),
	})
	require.ErrorIs(t, err, types.ErrNodeInactive)
}

func TestUpdateNodeAgentShare_Decrease(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_share6___________")).String()
	agent := sdk.AccAddress([]byte("ag_share6___________")).String()
	registerTestNode(t, f, operator, agent)

	// Current = 30, decrease to 20 (diff=10, within max rate=10).
	_, err := msgServer.UpdateNodeAgentShare(ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator,
		NewAgentShare: math.LegacyNewDec(20),
	})
	require.NoError(t, err)
}

package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestRenewFeegrant_Success(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_renew1___________")).String()
	agent := sdk.AccAddress([]byte("ag_renew1___________")).String()
	registerTestNode(t, f, operator, agent)

	resp, err := msgServer.RenewFeegrant(ctx, &types.MsgRenewFeegrant{
		Operator: operator,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Event should be emitted.
	evt := requireEvent(t, ctx, types.EventTypeFeegrantRenewed)
	require.NotEmpty(t, eventAttribute(evt, types.AttributeKeyNodeID))
	require.Equal(t, agent, eventAttribute(evt, types.AttributeKeyAgentAddress))
}

func TestRenewFeegrant_NotFound(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_renew_nf_________")).String()
	_, err := msgServer.RenewFeegrant(ctx, &types.MsgRenewFeegrant{
		Operator: operator,
	})
	require.ErrorIs(t, err, types.ErrNodeNotFound)
}

func TestRenewFeegrant_InactiveNode(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_renew2___________")).String()
	agent := sdk.AccAddress([]byte("ag_renew2___________")).String()
	registerTestNode(t, f, operator, agent)

	// Deactivate node.
	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{Operator: operator})
	require.NoError(t, err)

	_, err = msgServer.RenewFeegrant(ctx, &types.MsgRenewFeegrant{
		Operator: operator,
	})
	require.ErrorIs(t, err, types.ErrNodeInactive)
}

func TestRenewFeegrant_NoAgentAddress(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_renew3___________")).String()

	// Register node with no agent.
	seed := byte(pubkeySeed.Add(1))
	_, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            "",
		AgentShare:              types.DefaultAgentShareForTest(),
		MaxAgentShareChangeRate: types.DefaultAgentShareForTest(),
		Description:             "no agent",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)

	_, err = msgServer.RenewFeegrant(ctx, &types.MsgRenewFeegrant{
		Operator: operator,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no agent address")
}

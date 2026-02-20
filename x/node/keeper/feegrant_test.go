package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestGrantAgentFeegrant_NilFeegrantKeeper(t *testing.T) {
	f := initFixture(t)

	// Clear feegrant keeper BEFORE creating MsgServer (value copy).
	f.keeper.SetFeegrantKeeper(nil)
	ms := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_fg_nil___________")).String()

	// Register without feegrant keeper. Should succeed without error.
	seed := byte(pubkeySeed.Add(1))
	resp, err := ms.RegisterNode(f.ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            sdk.AccAddress([]byte("ag_fg_nil___________")).String(),
		AgentShare:              types.DefaultAgentShareForTest(),
		MaxAgentShareChangeRate: types.DefaultAgentShareForTest(),
		Description:             "no feegrant keeper",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.NodeId)

	// Feegrant should NOT have been called (keeper is nil).
	require.Len(t, f.feegrantKeeper.grantCalls, 0)
}

func TestGrantAgentFeegrant_EmptyAgentAddress(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_fg_empty_________")).String()

	// Register with empty agent address — feegrant should not be called.
	seed := byte(pubkeySeed.Add(1))
	resp, err := ms.RegisterNode(f.ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            "",
		AgentShare:              types.DefaultAgentShareForTest(),
		MaxAgentShareChangeRate: types.DefaultAgentShareForTest(),
		Description:             "empty agent",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.NodeId)

	// Feegrant should NOT have been called for empty agent.
	require.Len(t, f.feegrantKeeper.grantCalls, 0)
}

func TestGrantAgentFeegrant_NilModuleAddress(t *testing.T) {
	f := initFixture(t)

	// Remove feegrant pool module address BEFORE creating MsgServer.
	delete(f.authKeeper.moduleAddresses, types.FeegrantPoolName)
	ms := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_fg_nomod_________")).String()
	agent := sdk.AccAddress([]byte("ag_fg_nomod_________")).String()

	seed := byte(pubkeySeed.Add(1))
	resp, err := ms.RegisterNode(f.ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agent,
		AgentShare:              types.DefaultAgentShareForTest(),
		MaxAgentShareChangeRate: types.DefaultAgentShareForTest(),
		Description:             "no module addr",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.NodeId)

	// No feegrant calls because module address is nil.
	require.Len(t, f.feegrantKeeper.grantCalls, 0)
}

func TestGrantAgentFeegrant_GrantFails_EmitsEvent(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	// Set feegrant keeper to always return error.
	f.feegrantKeeper.grantErr = fmt.Errorf("mock feegrant failure")

	operator := sdk.AccAddress([]byte("op_fg_fail__________")).String()
	agent := sdk.AccAddress([]byte("ag_fg_fail__________")).String()

	seed := byte(pubkeySeed.Add(1))
	resp, err := ms.RegisterNode(f.ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agent,
		AgentShare:              types.DefaultAgentShareForTest(),
		MaxAgentShareChangeRate: types.DefaultAgentShareForTest(),
		Description:             "feegrant failure",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	// Registration should succeed despite feegrant failure (best-effort).
	require.NoError(t, err)
	require.NotEmpty(t, resp.NodeId)

	// Verify feegrant_grant_failed event was emitted.
	evt := requireEvent(t, f.ctx, types.EventTypeFeegrantGrantFailed)
	require.Equal(t, agent, eventAttribute(evt, types.AttributeKeyAgentAddress))
	require.Contains(t, eventAttribute(evt, types.AttributeKeyError), "mock feegrant failure")
}

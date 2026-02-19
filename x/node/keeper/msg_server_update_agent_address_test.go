package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestUpdateAgentAddress_Success(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_addr1____________")).String()
	oldAgent := sdk.AccAddress([]byte("old_agent___________")).String()
	newAgent := sdk.AccAddress([]byte("new_agent___________")).String()
	nodeID := registerTestNode(t, f, operator, oldAgent)

	// Change agent address.
	_, err := msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: newAgent,
	})
	require.NoError(t, err)

	// Verify node updated.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, newAgent, node.AgentAddress)

	// Old agent index removed.
	has, err := f.keeper.AgentIndex.Has(ctx, oldAgent)
	require.NoError(t, err)
	require.False(t, has)

	// New agent index set.
	id, err := f.keeper.AgentIndex.Get(ctx, newAgent)
	require.NoError(t, err)
	require.Equal(t, nodeID, id)
}

func TestUpdateAgentAddress_Cooldown(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_addr2____________")).String()
	oldAgent := sdk.AccAddress([]byte("old_agent2__________")).String()
	newAgent1 := sdk.AccAddress([]byte("new_agent2a_________")).String()
	newAgent2 := sdk.AccAddress([]byte("new_agent2b_________")).String()
	registerTestNode(t, f, operator, oldAgent)

	// First change: OK.
	_, err := msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: newAgent1,
	})
	require.NoError(t, err)

	// Second change immediately: should fail (cooldown not elapsed).
	_, err = msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: newAgent2,
	})
	require.ErrorIs(t, err, types.ErrAgentAddressChangeCooldown)
}

func TestUpdateAgentAddress_CooldownElapsed(t *testing.T) {
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_addr3____________")).String()
	oldAgent := sdk.AccAddress([]byte("old_agent3__________")).String()
	newAgent1 := sdk.AccAddress([]byte("new_agent3a_________")).String()
	newAgent2 := sdk.AccAddress([]byte("new_agent3b_________")).String()

	ctx := f.ctx
	registerTestNode(t, f, operator, oldAgent)

	// First change at block 0.
	_, err := msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: newAgent1,
	})
	require.NoError(t, err)

	// Advance block height past cooldown.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx = sdkCtx.WithBlockHeight(int64(types.DefaultAgentAddressChangeCooldown) + 1)
	newCtx := sdkCtx

	// Second change after cooldown: should succeed.
	_, err = msgServer.UpdateAgentAddress(newCtx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: newAgent2,
	})
	require.NoError(t, err)
}

func TestUpdateAgentAddress_Deactivate(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_addr4____________")).String()
	agent := sdk.AccAddress([]byte("agent_deact_________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Deactivate agent by setting empty address.
	_, err := msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: "",
	})
	require.NoError(t, err)

	// Verify agent is cleared.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, "", node.AgentAddress)

	// Agent index removed.
	has, err := f.keeper.AgentIndex.Has(ctx, agent)
	require.NoError(t, err)
	require.False(t, has)
}

func TestUpdateAgentAddress_DuplicateAgent(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator1 := sdk.AccAddress([]byte("op_addr5a___________")).String()
	operator2 := sdk.AccAddress([]byte("op_addr5b___________")).String()
	agent1 := sdk.AccAddress([]byte("agent5a_____________")).String()
	agent2 := sdk.AccAddress([]byte("agent5b_____________")).String()
	registerTestNode(t, f, operator1, agent1)
	registerTestNode(t, f, operator2, agent2)

	// Try to change operator1's agent to agent2 (already used).
	_, err := msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator1,
		NewAgentAddress: agent2,
	})
	require.ErrorIs(t, err, types.ErrAgentAddressAlreadyUsed)
}

func TestUpdateAgentAddress_NotFound(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_notfound2________")).String()
	_, err := msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: "some_addr",
	})
	require.ErrorIs(t, err, types.ErrNodeNotFound)
}

func TestUpdateAgentAddress_InactiveNode(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_addr6____________")).String()
	agent := sdk.AccAddress([]byte("agent6______________")).String()
	registerTestNode(t, f, operator, agent)

	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{Operator: operator})
	require.NoError(t, err)

	_, err = msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: "new_addr",
	})
	require.ErrorIs(t, err, types.ErrNodeInactive)
}

func TestUpdateAgentAddress_SameAddress(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_addr7____________")).String()
	agent := sdk.AccAddress([]byte("agent7______________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Setting the same agent address should succeed (no uniqueness violation).
	_, err := msgServer.UpdateAgentAddress(ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator,
		NewAgentAddress: agent,
	})
	require.NoError(t, err)

	// Node should still have the same agent.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, agent, node.AgentAddress)
}

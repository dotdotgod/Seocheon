package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"seocheon/x/node/types"
)

func TestIsRegisteredAgent(t *testing.T) {
	f := initFixture(t)

	agentAddr := "seocheon1registered_agent"
	nodeID := "test-node-for-agent"

	// Before registering, should return false.
	require.False(t, f.keeper.IsRegisteredAgent(f.ctx, agentAddr))

	// Register agent in the AgentIndex.
	err := f.keeper.AgentIndex.Set(f.ctx, agentAddr, nodeID)
	require.NoError(t, err)

	// After registering, should return true.
	require.True(t, f.keeper.IsRegisteredAgent(f.ctx, agentAddr))

	// Unknown address should still return false.
	require.False(t, f.keeper.IsRegisteredAgent(f.ctx, "seocheon1unknown_addr"))
}

func TestGetAllowedAgentMsgTypes(t *testing.T) {
	f := initFixture(t)

	// With default params, should return the default allowed msg types.
	allowed, err := f.keeper.GetAllowedAgentMsgTypes(f.ctx)
	require.NoError(t, err)
	require.Equal(t, types.DefaultAgentAllowedMsgTypes, allowed)
	require.Contains(t, allowed, "/seocheon.activity.v1.MsgSubmitActivity")
	require.Contains(t, allowed, "/cosmwasm.wasm.v1.MsgExecuteContract")
	require.Contains(t, allowed, "/cosmos.bank.v1beta1.MsgSend")
}

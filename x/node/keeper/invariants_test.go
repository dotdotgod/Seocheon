package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestModuleAccountInvariant(t *testing.T) {
	f := initFixture(t)
	ctx := sdk.UnwrapSDKContext(f.ctx)

	t.Run("passes with valid module accounts", func(t *testing.T) {
		msg, broken := keeper.ModuleAccountInvariant(f.keeper)(ctx)
		require.False(t, broken, msg)
	})

	t.Run("fails when registration pool address is nil", func(t *testing.T) {
		// Remove registration pool address from mock.
		delete(f.authKeeper.moduleAddresses, types.RegistrationPoolName)
		msg, broken := keeper.ModuleAccountInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "registration pool module account does not exist")

		// Restore for subsequent tests.
		f.authKeeper.moduleAddresses[types.RegistrationPoolName] = testAddr("reg_pool__________")
	})

	t.Run("fails when feegrant pool address is nil", func(t *testing.T) {
		delete(f.authKeeper.moduleAddresses, types.FeegrantPoolName)
		msg, broken := keeper.ModuleAccountInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "feegrant pool module account does not exist")
	})
}

func TestIndexConsistencyInvariant(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)
	ctx := sdk.UnwrapSDKContext(f.ctx)

	// Register a node to have data in the store.
	operator1 := testAddr("invop1____________")
	agent1 := testAddr("invag1____________")
	msg := &types.MsgRegisterNode{
		Operator:                operator1.String(),
		AgentAddress:            agent1.String(),
		AgentShare:              defaultAgentShare,
		MaxAgentShareChangeRate: defaultMaxChangeRate,
		Description:             "invariant test node",
		Tags:                    []string{"test"},
		ConsensusPubkey:         testPubKey(50),
	}
	resp, err := ms.RegisterNode(f.ctx, msg)
	require.NoError(t, err)

	t.Run("passes with consistent indexes", func(t *testing.T) {
		msg, broken := keeper.IndexConsistencyInvariant(f.keeper)(ctx)
		require.False(t, broken, msg)
	})

	t.Run("fails when operator index is missing", func(t *testing.T) {
		// Remove operator index manually.
		err := f.keeper.OperatorIndex.Remove(f.ctx, operator1.String())
		require.NoError(t, err)

		msg, broken := keeper.IndexConsistencyInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "operator index missing")

		// Restore.
		err = f.keeper.OperatorIndex.Set(f.ctx, operator1.String(), resp.NodeId)
		require.NoError(t, err)
	})

	t.Run("fails when agent index is missing", func(t *testing.T) {
		err := f.keeper.AgentIndex.Remove(f.ctx, agent1.String())
		require.NoError(t, err)

		msg, broken := keeper.IndexConsistencyInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "agent index missing")

		// Restore.
		err = f.keeper.AgentIndex.Set(f.ctx, agent1.String(), resp.NodeId)
		require.NoError(t, err)
	})

	t.Run("fails when operator index points to non-existent node", func(t *testing.T) {
		// Add a dangling operator index.
		err := f.keeper.OperatorIndex.Set(f.ctx, "dangling_operator", "nonexistent_node_id")
		require.NoError(t, err)

		msg, broken := keeper.IndexConsistencyInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "non-existent node")

		// Cleanup.
		err = f.keeper.OperatorIndex.Remove(f.ctx, "dangling_operator")
		require.NoError(t, err)
	})

	t.Run("fails when node count != operator index count", func(t *testing.T) {
		// Add an extra operator index without a corresponding node.
		err := f.keeper.OperatorIndex.Set(f.ctx, "extra_operator", resp.NodeId)
		require.NoError(t, err)

		msg, broken := keeper.IndexConsistencyInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "node count")

		// Cleanup.
		err = f.keeper.OperatorIndex.Remove(f.ctx, "extra_operator")
		require.NoError(t, err)
	})
}

func TestNodeStatusInvariant(t *testing.T) {
	f := initFixture(t)
	ctx := sdk.UnwrapSDKContext(f.ctx)

	t.Run("passes with no nodes", func(t *testing.T) {
		msg, broken := keeper.NodeStatusInvariant(f.keeper)(ctx)
		require.False(t, broken, msg)
	})

	t.Run("passes with valid node statuses", func(t *testing.T) {
		// Insert a node with REGISTERED status.
		node := types.Node{
			Id:       "status-test-node",
			Operator: "operator_status",
			Status:   types.NodeStatus_NODE_STATUS_REGISTERED,
		}
		err := f.keeper.Nodes.Set(f.ctx, node.Id, node)
		require.NoError(t, err)
		err = f.keeper.OperatorIndex.Set(f.ctx, node.Operator, node.Id)
		require.NoError(t, err)

		msg, broken := keeper.NodeStatusInvariant(f.keeper)(ctx)
		require.False(t, broken, msg)

		// Cleanup.
		_ = f.keeper.Nodes.Remove(f.ctx, node.Id)
		_ = f.keeper.OperatorIndex.Remove(f.ctx, node.Operator)
	})

	t.Run("fails with UNSPECIFIED status", func(t *testing.T) {
		node := types.Node{
			Id:       "bad-status-node",
			Operator: "operator_bad",
			Status:   types.NodeStatus_NODE_STATUS_UNSPECIFIED,
		}
		err := f.keeper.Nodes.Set(f.ctx, node.Id, node)
		require.NoError(t, err)

		msg, broken := keeper.NodeStatusInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "UNSPECIFIED")

		// Cleanup.
		_ = f.keeper.Nodes.Remove(f.ctx, node.Id)
	})

	t.Run("fails with empty operator", func(t *testing.T) {
		node := types.Node{
			Id:       "empty-op-node",
			Operator: "",
			Status:   types.NodeStatus_NODE_STATUS_REGISTERED,
		}
		err := f.keeper.Nodes.Set(f.ctx, node.Id, node)
		require.NoError(t, err)

		msg, broken := keeper.NodeStatusInvariant(f.keeper)(ctx)
		require.True(t, broken)
		require.Contains(t, msg, "empty operator")

		// Cleanup.
		_ = f.keeper.Nodes.Remove(f.ctx, node.Id)
	})
}

func TestAllInvariants(t *testing.T) {
	f := initFixture(t)
	ctx := sdk.UnwrapSDKContext(f.ctx)

	t.Run("all pass on clean state", func(t *testing.T) {
		msg, broken := keeper.AllInvariants(f.keeper)(ctx)
		require.False(t, broken, msg)
	})
}

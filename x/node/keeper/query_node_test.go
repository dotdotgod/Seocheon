package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

// insertTestNode stores a test node and its indexes in the keeper directly.
func insertTestNode(t *testing.T, f *fixture, id, operator, agentAddr string, tags []string) types.Node {
	t.Helper()

	node := types.Node{
		Id:                      id,
		Operator:                operator,
		AgentAddress:            agentAddr,
		AgentShare:              math.LegacyNewDec(30),
		MaxAgentShareChangeRate: math.LegacyNewDec(5),
		Tags:                    tags,
		Status:                  types.NodeStatus_NODE_STATUS_REGISTERED,
	}

	err := f.keeper.Nodes.Set(f.ctx, id, node)
	require.NoError(t, err)

	err = f.keeper.OperatorIndex.Set(f.ctx, operator, id)
	require.NoError(t, err)

	if agentAddr != "" {
		err = f.keeper.AgentIndex.Set(f.ctx, agentAddr, id)
		require.NoError(t, err)
	}

	for _, tag := range tags {
		err = f.keeper.TagIndex.Set(f.ctx, collections.Join(tag, id))
		require.NoError(t, err)
	}

	return node
}

func TestQueryNode(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)

	node := insertTestNode(t, f, "node-001", "seocheon1op1", "seocheon1ag1", []string{"ai"})

	t.Run("found by ID", func(t *testing.T) {
		resp, err := qs.Node(f.ctx, &types.QueryNodeRequest{NodeId: "node-001"})
		require.NoError(t, err)
		require.Equal(t, node.Id, resp.Node.Id)
		require.Equal(t, node.Operator, resp.Node.Operator)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := qs.Node(f.ctx, &types.QueryNodeRequest{NodeId: "nonexistent"})
		require.Error(t, err)
		require.ErrorContains(t, err, "not found")
	})

	t.Run("empty node_id", func(t *testing.T) {
		_, err := qs.Node(f.ctx, &types.QueryNodeRequest{NodeId: ""})
		require.Error(t, err)
		require.ErrorContains(t, err, "node_id is required")
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := qs.Node(f.ctx, nil)
		require.Error(t, err)
	})
}

func TestQueryNodeByOperator(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)

	node := insertTestNode(t, f, "node-002", "seocheon1op2", "seocheon1ag2", nil)

	t.Run("found by operator", func(t *testing.T) {
		resp, err := qs.NodeByOperator(f.ctx, &types.QueryNodeByOperatorRequest{Operator: "seocheon1op2"})
		require.NoError(t, err)
		require.Equal(t, node.Id, resp.Node.Id)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := qs.NodeByOperator(f.ctx, &types.QueryNodeByOperatorRequest{Operator: "seocheon1unknown"})
		require.Error(t, err)
		require.ErrorContains(t, err, "not found")
	})

	t.Run("empty operator", func(t *testing.T) {
		_, err := qs.NodeByOperator(f.ctx, &types.QueryNodeByOperatorRequest{Operator: ""})
		require.Error(t, err)
		require.ErrorContains(t, err, "operator is required")
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := qs.NodeByOperator(f.ctx, nil)
		require.Error(t, err)
	})
}

func TestQueryNodeByAgentAddress(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)

	node := insertTestNode(t, f, "node-003", "seocheon1op3", "seocheon1ag3", nil)

	t.Run("found by agent address", func(t *testing.T) {
		resp, err := qs.NodeByAgentAddress(f.ctx, &types.QueryNodeByAgentAddressRequest{AgentAddress: "seocheon1ag3"})
		require.NoError(t, err)
		require.Equal(t, node.Id, resp.Node.Id)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := qs.NodeByAgentAddress(f.ctx, &types.QueryNodeByAgentAddressRequest{AgentAddress: "seocheon1unknown_agent"})
		require.Error(t, err)
		require.ErrorContains(t, err, "not found")
	})

	t.Run("empty agent_address", func(t *testing.T) {
		_, err := qs.NodeByAgentAddress(f.ctx, &types.QueryNodeByAgentAddressRequest{AgentAddress: ""})
		require.Error(t, err)
		require.ErrorContains(t, err, "agent_address is required")
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := qs.NodeByAgentAddress(f.ctx, nil)
		require.Error(t, err)
	})
}

func TestQueryNodesByTag(t *testing.T) {
	f := initFixture(t)
	qs := keeper.NewQueryServerImpl(f.keeper)

	// Insert nodes with different tags.
	insertTestNode(t, f, "node-t1", "seocheon1opt1", "seocheon1agt1", []string{"ai", "nlp"})
	insertTestNode(t, f, "node-t2", "seocheon1opt2", "seocheon1agt2", []string{"ai", "vision"})
	insertTestNode(t, f, "node-t3", "seocheon1opt3", "seocheon1agt3", []string{"blockchain"})

	t.Run("multiple nodes with same tag", func(t *testing.T) {
		resp, err := qs.NodesByTag(f.ctx, &types.QueryNodesByTagRequest{Tag: "ai"})
		require.NoError(t, err)
		require.Len(t, resp.Nodes, 2)

		ids := make([]string, len(resp.Nodes))
		for i, n := range resp.Nodes {
			ids[i] = n.Id
		}
		require.Contains(t, ids, "node-t1")
		require.Contains(t, ids, "node-t2")
	})

	t.Run("single node with unique tag", func(t *testing.T) {
		resp, err := qs.NodesByTag(f.ctx, &types.QueryNodesByTagRequest{Tag: "blockchain"})
		require.NoError(t, err)
		require.Len(t, resp.Nodes, 1)
		require.Equal(t, "node-t3", resp.Nodes[0].Id)
	})

	t.Run("no nodes with tag", func(t *testing.T) {
		resp, err := qs.NodesByTag(f.ctx, &types.QueryNodesByTagRequest{Tag: "nonexistent_tag"})
		require.NoError(t, err)
		require.Empty(t, resp.Nodes)
	})

	t.Run("empty tag", func(t *testing.T) {
		_, err := qs.NodesByTag(f.ctx, &types.QueryNodesByTagRequest{Tag: ""})
		require.Error(t, err)
		require.ErrorContains(t, err, "tag is required")
	})

	t.Run("nil request", func(t *testing.T) {
		_, err := qs.NodesByTag(f.ctx, nil)
		require.Error(t, err)
	})
}

func TestQueryAllNodes(t *testing.T) {
	t.Run("empty store", func(t *testing.T) {
		f := initFixture(t)
		qs := keeper.NewQueryServerImpl(f.keeper)

		resp, err := qs.AllNodes(f.ctx, &types.QueryAllNodesRequest{})
		require.NoError(t, err)
		require.Empty(t, resp.Nodes)
	})

	t.Run("multiple nodes", func(t *testing.T) {
		f := initFixture(t)
		qs := keeper.NewQueryServerImpl(f.keeper)

		insertTestNode(t, f, "all-n1", "seocheon1all_op1", "seocheon1all_ag1", nil)
		insertTestNode(t, f, "all-n2", "seocheon1all_op2", "seocheon1all_ag2", nil)
		insertTestNode(t, f, "all-n3", "seocheon1all_op3", "", nil)

		resp, err := qs.AllNodes(f.ctx, &types.QueryAllNodesRequest{})
		require.NoError(t, err)
		require.Len(t, resp.Nodes, 3)
	})

	t.Run("nil request", func(t *testing.T) {
		f := initFixture(t)
		qs := keeper.NewQueryServerImpl(f.keeper)

		_, err := qs.AllNodes(f.ctx, nil)
		require.Error(t, err)
	})
}

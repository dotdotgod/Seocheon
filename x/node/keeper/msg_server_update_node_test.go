package keeper_test

import (
	"sync/atomic"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

// pubkeySeed provides unique seeds for test pubkeys across all tests.
var pubkeySeed atomic.Uint64

func init() {
	pubkeySeed.Store(200) // start above the seeds used in register_node_test.go
}

func registerTestNode(t *testing.T, f *fixture, operator, agentAddr string) string {
	t.Helper()
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	seed := byte(pubkeySeed.Add(1))
	resp, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agentAddr,
		AgentShare:              math.LegacyNewDec(30),
		MaxAgentShareChangeRate: math.LegacyNewDec(10),
		Description:             "test node",
		Website:                 "https://example.com",
		Tags:                    []string{"ai", "test"},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)
	return resp.NodeId
}

func TestUpdateNode_Success(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("operator1___________")).String()
	agent := sdk.AccAddress([]byte("agent1______________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Update description, website, and tags.
	_, err := msgServer.UpdateNode(ctx, &types.MsgUpdateNode{
		Operator:    operator,
		Description: "updated description",
		Website:     "https://new-site.com",
		Tags:        []string{"blockchain", "defi"},
	})
	require.NoError(t, err)

	// Verify.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, "updated description", node.Description)
	require.Equal(t, "https://new-site.com", node.Website)
	require.Equal(t, []string{"blockchain", "defi"}, node.Tags)

	// Old tags should be removed from index.
	has, err := f.keeper.TagIndex.Has(ctx, collections.Join("ai", nodeID))
	require.NoError(t, err)
	require.False(t, has)
	has, err = f.keeper.TagIndex.Has(ctx, collections.Join("test", nodeID))
	require.NoError(t, err)
	require.False(t, has)

	// New tags should be in index.
	has, err = f.keeper.TagIndex.Has(ctx, collections.Join("blockchain", nodeID))
	require.NoError(t, err)
	require.True(t, has)
	has, err = f.keeper.TagIndex.Has(ctx, collections.Join("defi", nodeID))
	require.NoError(t, err)
	require.True(t, has)
}

func TestUpdateNode_NotFound(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("nonexistent_________")).String()
	_, err := msgServer.UpdateNode(ctx, &types.MsgUpdateNode{
		Operator:    operator,
		Description: "should fail",
	})
	require.ErrorIs(t, err, types.ErrNodeNotFound)
}

func TestUpdateNode_InactiveNode(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("operator2___________")).String()
	agent := sdk.AccAddress([]byte("agent2______________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Deactivate the node.
	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{Operator: operator})
	require.NoError(t, err)

	// Verify it's inactive.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Equal(t, types.NodeStatus_NODE_STATUS_INACTIVE, node.Status)

	// Try to update — should fail.
	_, err = msgServer.UpdateNode(ctx, &types.MsgUpdateNode{
		Operator:    operator,
		Description: "should fail",
	})
	require.ErrorIs(t, err, types.ErrNodeInactive)
}

func TestUpdateNode_TooManyTags(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("operator3___________")).String()
	agent := sdk.AccAddress([]byte("agent3______________")).String()
	registerTestNode(t, f, operator, agent)

	tags := make([]string, 11) // MaxTags default is 10
	for i := range tags {
		tags[i] = "tag"
	}
	_, err := msgServer.UpdateNode(ctx, &types.MsgUpdateNode{
		Operator: operator,
		Tags:     tags,
	})
	require.ErrorIs(t, err, types.ErrInvalidTags)
}

func TestUpdateNode_EmptyTag(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("operator4___________")).String()
	agent := sdk.AccAddress([]byte("agent4______________")).String()
	registerTestNode(t, f, operator, agent)

	_, err := msgServer.UpdateNode(ctx, &types.MsgUpdateNode{
		Operator: operator,
		Tags:     []string{"valid", ""},
	})
	require.ErrorIs(t, err, types.ErrInvalidTags)
}

func TestUpdateNode_ClearTags(t *testing.T) {
	f := initFixture(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("operator5___________")).String()
	agent := sdk.AccAddress([]byte("agent5______________")).String()
	nodeID := registerTestNode(t, f, operator, agent)

	// Clear all tags.
	_, err := msgServer.UpdateNode(ctx, &types.MsgUpdateNode{
		Operator: operator,
		Tags:     []string{},
	})
	require.NoError(t, err)

	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	require.Empty(t, node.Tags)

	// Old tags should be cleaned up.
	has, err := f.keeper.TagIndex.Has(ctx, collections.Join("ai", nodeID))
	require.NoError(t, err)
	require.False(t, has)
}

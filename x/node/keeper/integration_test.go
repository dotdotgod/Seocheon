package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

// Package-level test defaults shared across test files (including invariants_test.go).
var (
	defaultAgentShare    = math.LegacyNewDec(30)
	defaultMaxChangeRate = math.LegacyNewDec(5)
)

// TestNodeLifecycle is an end-to-end integration test covering the full node lifecycle:
// Register → UpdateMetadata → ScheduleAgentShareChange → EpochBoundary →
// UpdateAgentAddress → WithdrawCommission → Deactivate → InactiveGuard → Invariants
func TestNodeLifecycle(t *testing.T) {
	f := initFixture(t)

	// Wire distribution keeper before creating MsgServer (value copy).
	dk := newMockDistributionKeeper()
	f.keeper.SetDistributionKeeper(dk)

	ms := keeper.NewMsgServerImpl(f.keeper)

	operator := testAddr("lifecycle_op______")
	operatorStr := operator.String()
	agent1 := testAddr("lifecycle_ag1_____")
	agent1Str := agent1.String()
	agent2 := testAddr("lifecycle_ag2_____")
	agent2Str := agent2.String()

	var nodeID string

	// ── Step 1: RegisterNode ──
	t.Run("1_register", func(t *testing.T) {
		resp, err := ms.RegisterNode(f.ctx, &types.MsgRegisterNode{
			Operator:                operatorStr,
			AgentAddress:            agent1Str,
			AgentShare:              defaultAgentShare,
			MaxAgentShareChangeRate: defaultMaxChangeRate,
			Description:             "lifecycle test node",
			Website:                 "https://lifecycle.test",
			Tags:                    []string{"ai", "test"},
			ConsensusPubkey:         testPubKey(100),
		})
		require.NoError(t, err)
		require.NotEmpty(t, resp.NodeId)
		require.NotEmpty(t, resp.ValidatorAddress)
		nodeID = resp.NodeId

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_REGISTERED, node.Status)
		require.Equal(t, operatorStr, node.Operator)
		require.Equal(t, agent1Str, node.AgentAddress)
		require.Equal(t, defaultAgentShare, node.AgentShare)

		// Verify indexes.
		id, err := f.keeper.OperatorIndex.Get(f.ctx, operatorStr)
		require.NoError(t, err)
		require.Equal(t, nodeID, id)

		id, err = f.keeper.AgentIndex.Get(f.ctx, agent1Str)
		require.NoError(t, err)
		require.Equal(t, nodeID, id)

		// Verify feegrant was called for agent.
		require.Len(t, f.feegrantKeeper.grantCalls, 1)
	})

	// ── Step 2: UpdateNode (metadata) ──
	t.Run("2_update_metadata", func(t *testing.T) {
		_, err := ms.UpdateNode(f.ctx, &types.MsgUpdateNode{
			Operator:    operatorStr,
			Description: "updated description",
			Website:     "https://updated.test",
			Tags:        []string{"ai", "updated"},
		})
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, "updated description", node.Description)
		require.Equal(t, "https://updated.test", node.Website)
		require.Equal(t, []string{"ai", "updated"}, node.Tags)
	})

	// ── Step 3: UpdateNodeAgentShare (schedule pending change) ──
	t.Run("3_schedule_agent_share_change", func(t *testing.T) {
		newShare := math.LegacyNewDec(35) // +5, within max_change_rate
		_, err := ms.UpdateNodeAgentShare(f.ctx, &types.MsgUpdateNodeAgentShare{
			Operator:      operatorStr,
			NewAgentShare: newShare,
		})
		require.NoError(t, err)

		// Pending change stored.
		pending, err := f.keeper.PendingAgentShareChanges.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, newShare, pending.NewAgentShare)
		require.Equal(t, types.EpochLength, pending.ApplyAtBlock)

		// Agent share NOT yet changed.
		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, defaultAgentShare, node.AgentShare, "agent_share should not change before epoch boundary")
	})

	// ── Step 4: EndBlocker at epoch boundary → apply change ──
	t.Run("4_epoch_applies_change", func(t *testing.T) {
		sdkCtx := sdk.UnwrapSDKContext(f.ctx)
		epochCtx := sdkCtx.WithBlockHeight(types.EpochLength)

		err := f.keeper.EndBlocker(epochCtx)
		require.NoError(t, err)

		// Agent share should now be 35.
		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, math.LegacyNewDec(35), node.AgentShare)

		// Pending change removed.
		has, err := f.keeper.PendingAgentShareChanges.Has(f.ctx, nodeID)
		require.NoError(t, err)
		require.False(t, has)
	})

	// ── Step 5: UpdateAgentAddress ──
	t.Run("5_swap_agent_address", func(t *testing.T) {
		sdkCtx := sdk.UnwrapSDKContext(f.ctx)
		f.ctx = sdkCtx.WithBlockHeight(types.EpochLength + 1)

		_, err := ms.UpdateAgentAddress(f.ctx, &types.MsgUpdateAgentAddress{
			Operator:        operatorStr,
			NewAgentAddress: agent2Str,
		})
		require.NoError(t, err)

		// Old agent index removed.
		has, err := f.keeper.AgentIndex.Has(f.ctx, agent1Str)
		require.NoError(t, err)
		require.False(t, has)

		// New agent index set.
		id, err := f.keeper.AgentIndex.Get(f.ctx, agent2Str)
		require.NoError(t, err)
		require.Equal(t, nodeID, id)

		// Node updated.
		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, agent2Str, node.AgentAddress)

		// Feegrant granted to new agent (register + update = 2).
		require.Len(t, f.feegrantKeeper.grantCalls, 2)
	})

	// ── Step 6: WithdrawNodeCommission ──
	t.Run("6_withdraw_commission", func(t *testing.T) {
		// Get the validator address to configure mock.
		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		valAddrBytes, err := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
		require.NoError(t, err)
		valAddr := sdk.ValAddress(valAddrBytes)

		// Configure mock to return 1000 usum commission.
		dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1000)))

		resp, err := ms.WithdrawNodeCommission(f.ctx, &types.MsgWithdrawNodeCommission{
			Operator: operatorStr,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Agent share is 35%, so agent gets 350, operator keeps 650.
		require.Equal(t, "650usum", resp.OperatorAmount)
		require.Equal(t, "350usum", resp.AgentAmount)

		// Verify bank SendCoins was called for agent split.
		require.Len(t, f.bankKeeper.sentCoinsRecords, 1)
		sent := f.bankKeeper.sentCoinsRecords[0]
		require.Equal(t, sdk.NewCoin("usum", math.NewInt(350)), sent.Amount[0])
	})

	// ── Step 7: DeactivateNode ──
	t.Run("7_deactivate", func(t *testing.T) {
		_, err := ms.DeactivateNode(f.ctx, &types.MsgDeactivateNode{
			Operator: operatorStr,
		})
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_INACTIVE, node.Status)
	})

	// ── Step 8: Operations fail on inactive node ──
	t.Run("8_inactive_guard", func(t *testing.T) {
		_, err := ms.UpdateNode(f.ctx, &types.MsgUpdateNode{
			Operator: operatorStr, Description: "fail",
		})
		require.ErrorContains(t, err, "inactive")

		_, err = ms.UpdateNodeAgentShare(f.ctx, &types.MsgUpdateNodeAgentShare{
			Operator: operatorStr, NewAgentShare: math.LegacyNewDec(40),
		})
		require.ErrorContains(t, err, "inactive")

		_, err = ms.DeactivateNode(f.ctx, &types.MsgDeactivateNode{
			Operator: operatorStr,
		})
		require.ErrorContains(t, err, "already inactive")

		_, err = ms.WithdrawNodeCommission(f.ctx, &types.MsgWithdrawNodeCommission{
			Operator: operatorStr,
		})
		require.ErrorContains(t, err, "inactive")
	})

	// ── Step 9: All invariants pass ──
	t.Run("9_all_invariants_pass", func(t *testing.T) {
		sdkCtx := sdk.UnwrapSDKContext(f.ctx)
		msg, broken := keeper.AllInvariants(f.keeper)(sdkCtx)
		require.False(t, broken, msg)
	})
}

// TestEndBlockerNonEpochIsNoop verifies EndBlocker does nothing on non-epoch blocks.
func TestEndBlockerNonEpochIsNoop(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	operator := testAddr("endb_op___________")
	agent := testAddr("endb_ag___________")

	resp, err := ms.RegisterNode(f.ctx, &types.MsgRegisterNode{
		Operator:                operator.String(),
		AgentAddress:            agent.String(),
		AgentShare:              defaultAgentShare,
		MaxAgentShareChangeRate: defaultMaxChangeRate,
		ConsensusPubkey:         testPubKey(110),
	})
	require.NoError(t, err)

	_, err = ms.UpdateNodeAgentShare(f.ctx, &types.MsgUpdateNodeAgentShare{
		Operator:      operator.String(),
		NewAgentShare: math.LegacyNewDec(35),
	})
	require.NoError(t, err)

	// EndBlocker at a non-epoch block should be a no-op.
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	nonEpochCtx := sdkCtx.WithBlockHeight(100)
	err = f.keeper.EndBlocker(nonEpochCtx)
	require.NoError(t, err)

	// Pending change should still exist.
	has, err := f.keeper.PendingAgentShareChanges.Has(f.ctx, resp.NodeId)
	require.NoError(t, err)
	require.True(t, has, "pending change should survive non-epoch EndBlocker")

	// Agent share unchanged.
	node, err := f.keeper.Nodes.Get(f.ctx, resp.NodeId)
	require.NoError(t, err)
	require.Equal(t, defaultAgentShare, node.AgentShare)
}

// TestAgentAddressChangeCooldown verifies the cooldown between agent address changes.
func TestAgentAddressChangeCooldown(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	operator := testAddr("cool_op___________")
	agent1 := testAddr("cool_ag1__________")
	agent2 := testAddr("cool_ag2__________")
	agent3 := testAddr("cool_ag3__________")

	_, err := ms.RegisterNode(f.ctx, &types.MsgRegisterNode{
		Operator:                operator.String(),
		AgentAddress:            agent1.String(),
		AgentShare:              defaultAgentShare,
		MaxAgentShareChangeRate: defaultMaxChangeRate,
		ConsensusPubkey:         testPubKey(130),
	})
	require.NoError(t, err)

	params, err := f.keeper.Params.Get(f.ctx)
	require.NoError(t, err)

	// First change should succeed.
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	f.ctx = sdkCtx.WithBlockHeight(100)

	_, err = ms.UpdateAgentAddress(f.ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator.String(),
		NewAgentAddress: agent2.String(),
	})
	require.NoError(t, err)

	// Second change within cooldown should fail.
	sdkCtx = sdk.UnwrapSDKContext(f.ctx)
	f.ctx = sdkCtx.WithBlockHeight(100 + int64(params.AgentAddressChangeCooldown) - 1)

	_, err = ms.UpdateAgentAddress(f.ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator.String(),
		NewAgentAddress: agent3.String(),
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "must wait")

	// After cooldown elapses, should succeed.
	sdkCtx = sdk.UnwrapSDKContext(f.ctx)
	f.ctx = sdkCtx.WithBlockHeight(100 + int64(params.AgentAddressChangeCooldown))

	_, err = ms.UpdateAgentAddress(f.ctx, &types.MsgUpdateAgentAddress{
		Operator:        operator.String(),
		NewAgentAddress: agent3.String(),
	})
	require.NoError(t, err)
}

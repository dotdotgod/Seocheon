package keeper_test

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

// mockDistributionKeeper implements types.DistributionKeeper for testing.
type mockDistributionKeeper struct {
	commissions map[string]sdk.Coins // validator_address -> coins
	withdrawErr error
}

func newMockDistributionKeeper() *mockDistributionKeeper {
	return &mockDistributionKeeper{
		commissions: make(map[string]sdk.Coins),
	}
}

func (m *mockDistributionKeeper) WithdrawValidatorCommission(_ context.Context, valAddr sdk.ValAddress) (sdk.Coins, error) {
	if m.withdrawErr != nil {
		return nil, m.withdrawErr
	}
	coins, ok := m.commissions[valAddr.String()]
	if !ok {
		return sdk.NewCoins(), nil
	}
	delete(m.commissions, valAddr.String())
	return coins, nil
}

func initFixtureWithDistribution(t *testing.T) (*fixture, *mockDistributionKeeper) {
	t.Helper()
	f := initFixture(t)

	dk := newMockDistributionKeeper()
	f.keeper.SetDistributionKeeper(dk)

	return f, dk
}

func TestWithdrawNodeCommission_Split30Percent(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm1____________")).String()
	agent := sdk.AccAddress([]byte("ag_comm1____________")).String()
	nodeID := registerTestNode(t, f, operator, agent) // agent_share = 30

	// Get the validator address for this node.
	node, err := f.keeper.Nodes.Get(ctx, nodeID)
	require.NoError(t, err)
	valAddrBytes, err := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	require.NoError(t, err)
	valAddr := sdk.ValAddress(valAddrBytes)

	// Set commission: 1000 usum.
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1000)))

	resp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)

	// 30% of 1000 = 300 to agent, 700 to operator.
	require.Equal(t, "300usum", resp.AgentAmount)
	require.Equal(t, "700usum", resp.OperatorAmount)

	// Verify bank.SendCoins was called with agent's share.
	require.Len(t, f.bankKeeper.sentCoinsRecords, 1)
	require.Equal(t, agent, f.bankKeeper.sentCoinsRecords[0].To.String())
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(300))), f.bankKeeper.sentCoinsRecords[0].Amount)
}

func TestWithdrawNodeCommission_ZeroAgentShare(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm2____________")).String()
	agent := sdk.AccAddress([]byte("ag_comm2____________")).String()

	// Register with 0% agent share.
	seed := byte(pubkeySeed.Add(1))
	resp, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agent,
		AgentShare:              math.LegacyZeroDec(),
		MaxAgentShareChangeRate: math.LegacyNewDec(10),
		Description:             "zero share",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)

	// Set commission.
	node, _ := f.keeper.Nodes.Get(ctx, resp.NodeId)
	valAddrBytes, _ := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	valAddr := sdk.ValAddress(valAddrBytes)
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(500)))

	withdrawResp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)

	// All goes to operator.
	require.Equal(t, "500usum", withdrawResp.OperatorAmount)
	require.Equal(t, "", withdrawResp.AgentAmount)

	// No SendCoins called (agent gets nothing).
	require.Len(t, f.bankKeeper.sentCoinsRecords, 0)
}

func TestWithdrawNodeCommission_100PercentAgent(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm3____________")).String()
	agent := sdk.AccAddress([]byte("ag_comm3____________")).String()

	// Register with 100% agent share.
	seed := byte(pubkeySeed.Add(1))
	resp, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agent,
		AgentShare:              math.LegacyNewDec(100),
		MaxAgentShareChangeRate: math.LegacyNewDec(10),
		Description:             "full agent share",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)

	node, _ := f.keeper.Nodes.Get(ctx, resp.NodeId)
	valAddrBytes, _ := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	valAddr := sdk.ValAddress(valAddrBytes)
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1000)))

	withdrawResp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)

	require.Equal(t, "1000usum", withdrawResp.AgentAmount)
	require.Equal(t, "", withdrawResp.OperatorAmount)
}

func TestWithdrawNodeCommission_ZeroCommission(t *testing.T) {
	f, _ := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm4____________")).String()
	agent := sdk.AccAddress([]byte("ag_comm4____________")).String()
	registerTestNode(t, f, operator, agent)

	// No commission set — should return zero gracefully.
	resp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)
	require.Equal(t, "", resp.OperatorAmount)
	require.Equal(t, "", resp.AgentAmount)
}

func TestWithdrawNodeCommission_NotFound(t *testing.T) {
	f, _ := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_notfound3________")).String()
	_, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.ErrorIs(t, err, types.ErrNodeNotFound)
}

func TestWithdrawNodeCommission_InactiveNode(t *testing.T) {
	f, _ := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm5____________")).String()
	agent := sdk.AccAddress([]byte("ag_comm5____________")).String()
	registerTestNode(t, f, operator, agent)

	_, err := msgServer.DeactivateNode(ctx, &types.MsgDeactivateNode{Operator: operator})
	require.NoError(t, err)

	_, err = msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.ErrorIs(t, err, types.ErrNodeInactive)
}

func TestWithdrawNodeCommission_DistributionError(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm6____________")).String()
	agent := sdk.AccAddress([]byte("ag_comm6____________")).String()
	registerTestNode(t, f, operator, agent)

	dk.withdrawErr = fmt.Errorf("distribution module error")

	_, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "distribution module error")
}

func TestWithdrawNodeCommission_TruncateSmallAmount(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm_trunc1______")).String()
	agent := sdk.AccAddress([]byte("ag_comm_trunc1______")).String()

	// Register with 33% agent share.
	seed := byte(pubkeySeed.Add(1))
	resp, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agent,
		AgentShare:              math.LegacyNewDec(33),
		MaxAgentShareChangeRate: math.LegacyNewDec(10),
		Description:             "truncation test",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)

	node, _ := f.keeper.Nodes.Get(ctx, resp.NodeId)
	valAddrBytes, _ := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	valAddr := sdk.ValAddress(valAddrBytes)

	// 1 usum at 33% → agent = TruncateInt(0.33) = 0, operator = 1.
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1)))

	withdrawResp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)

	// 33% of 1 = 0.33, truncated to 0. Agent gets nothing, operator gets 1.
	require.Equal(t, "1usum", withdrawResp.OperatorAmount)
	require.Equal(t, "", withdrawResp.AgentAmount)

	// No SendCoins because agent amount is zero.
	require.Len(t, f.bankKeeper.sentCoinsRecords, 0)
}

func TestWithdrawNodeCommission_TruncateAt10(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm_trunc2______")).String()
	agent := sdk.AccAddress([]byte("ag_comm_trunc2______")).String()

	seed := byte(pubkeySeed.Add(1))
	resp, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agent,
		AgentShare:              math.LegacyNewDec(33),
		MaxAgentShareChangeRate: math.LegacyNewDec(10),
		Description:             "truncation test 10",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)

	node, _ := f.keeper.Nodes.Get(ctx, resp.NodeId)
	valAddrBytes, _ := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	valAddr := sdk.ValAddress(valAddrBytes)

	// 10 usum at 33% → agent = TruncateInt(3.3) = 3, operator = 7.
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(10)))

	withdrawResp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)
	require.Equal(t, "3usum", withdrawResp.AgentAmount)
	require.Equal(t, "7usum", withdrawResp.OperatorAmount)
}

func TestWithdrawNodeCommission_1PercentShare(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm_1pct________")).String()
	agent := sdk.AccAddress([]byte("ag_comm_1pct________")).String()

	seed := byte(pubkeySeed.Add(1))
	resp, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            agent,
		AgentShare:              math.LegacyNewDec(1),
		MaxAgentShareChangeRate: math.LegacyNewDec(10),
		Description:             "1% share",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)

	node, _ := f.keeper.Nodes.Get(ctx, resp.NodeId)
	valAddrBytes, _ := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	valAddr := sdk.ValAddress(valAddrBytes)

	// 100 usum at 1% → agent = 1, operator = 99.
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(100)))

	withdrawResp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)
	require.Equal(t, "1usum", withdrawResp.AgentAmount)
	require.Equal(t, "99usum", withdrawResp.OperatorAmount)
}

func TestWithdrawNodeCommission_EmitsEvent(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm_evt_________")).String()
	agent := sdk.AccAddress([]byte("ag_comm_evt_________")).String()
	registerTestNode(t, f, operator, agent)

	node, _ := f.keeper.Nodes.Get(ctx, expectedNodeID(operator))
	valAddrBytes, _ := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	valAddr := sdk.ValAddress(valAddrBytes)
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(100)))

	_, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)

	evt := requireEvent(t, ctx, types.EventTypeCommissionWithdrawn)
	require.NotEmpty(t, eventAttribute(evt, types.AttributeKeyNodeID))
	require.NotEmpty(t, eventAttribute(evt, types.AttributeKeyOperatorAmount))
}

func TestWithdrawNodeCommission_NoAgent(t *testing.T) {
	f, dk := initFixtureWithDistribution(t)
	ctx := f.ctx
	msgServer := keeper.NewMsgServerImpl(f.keeper)

	operator := sdk.AccAddress([]byte("op_comm7____________")).String()

	// Register with no agent.
	seed := byte(pubkeySeed.Add(1))
	resp, err := msgServer.RegisterNode(ctx, &types.MsgRegisterNode{
		Operator:                operator,
		AgentAddress:            "", // no agent
		AgentShare:              math.LegacyNewDec(30),
		MaxAgentShareChangeRate: math.LegacyNewDec(10),
		Description:             "no agent",
		Tags:                    []string{},
		ConsensusPubkey:         testPubKey(seed),
	})
	require.NoError(t, err)

	node, _ := f.keeper.Nodes.Get(ctx, resp.NodeId)
	valAddrBytes, _ := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	valAddr := sdk.ValAddress(valAddrBytes)
	dk.commissions[valAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1000)))

	withdrawResp, err := msgServer.WithdrawNodeCommission(ctx, &types.MsgWithdrawNodeCommission{
		Operator: operator,
	})
	require.NoError(t, err)

	// All goes to operator since no agent address.
	require.Equal(t, "1000usum", withdrawResp.OperatorAmount)
	require.Equal(t, "", withdrawResp.AgentAmount)
}

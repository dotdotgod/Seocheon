package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/activity/types"
)

// makeAddr creates a deterministic bech32 address for testing.
func makeAddr(name string) string {
	return sdk.AccAddress([]byte(name + "________________")[:20]).String()
}

func TestDistributeActivityRewards_NoEligibleNodes(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// Fund pool but have no eligible nodes.
	f.bankKeeper.fundModule(types.ActivityRewardPoolName, sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1000000))))

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	// No sends should have occurred.
	require.Empty(t, f.bankKeeper.moduleToAccSent)
}

func TestDistributeActivityRewards_EmptyPool(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// Set up eligible node but empty pool.
	f.keeper.EpochSummary.Set(ctx, collections.Join("node1", int64(0)), types.EpochActivitySummary{
		TotalActivities: 10,
		ActiveWindows:   8,
		Eligible:        true,
	})

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	// No sends.
	require.Empty(t, f.bankKeeper.moduleToAccSent)
}

func TestDistributeActivityRewards_SingleNode_NoAgentShare(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	operatorAddr := makeAddr("operator1")
	agentWalletAddr := makeAddr("agent1")

	f.nodeKeeper.registerFullNode("node1", "agent1_submitter", operatorAddr, agentWalletAddr, 2, math.LegacyZeroDec())

	// Make node eligible.
	f.keeper.EpochSummary.Set(ctx, collections.Join("node1", int64(0)), types.EpochActivitySummary{
		TotalActivities: 10,
		ActiveWindows:   8,
		Eligible:        true,
	})

	// Fund pool with 1,000,000 usum.
	poolAmount := math.NewInt(1000000)
	f.bankKeeper.fundModule(types.ActivityRewardPoolName, sdk.NewCoins(sdk.NewCoin("usum", poolAmount)))

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	// All goes to operator (agent_share = 0).
	require.Len(t, f.bankKeeper.moduleToAccSent, 1)
	require.Equal(t, poolAmount, f.bankKeeper.moduleToAccSent[0].Amount.AmountOf("usum"))

	opAddr, _ := sdk.AccAddressFromBech32(operatorAddr)
	require.True(t, f.bankKeeper.moduleToAccSent[0].To.Equals(opAddr))
}

func TestDistributeActivityRewards_SingleNode_WithAgentShare(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	operatorAddr := makeAddr("operator1")
	agentWalletAddr := makeAddr("agent1")

	// agent_share = 30%
	f.nodeKeeper.registerFullNode("node1", "agent1_submitter", operatorAddr, agentWalletAddr, 2, math.LegacyNewDec(30))

	f.keeper.EpochSummary.Set(ctx, collections.Join("node1", int64(0)), types.EpochActivitySummary{
		TotalActivities: 10,
		ActiveWindows:   8,
		Eligible:        true,
	})

	poolAmount := math.NewInt(1000000)
	f.bankKeeper.fundModule(types.ActivityRewardPoolName, sdk.NewCoins(sdk.NewCoin("usum", poolAmount)))

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	// Expect 2 transfers: operator 700000 + agent 300000
	require.Len(t, f.bankKeeper.moduleToAccSent, 2)

	opAddr, _ := sdk.AccAddressFromBech32(operatorAddr)
	agAddr, _ := sdk.AccAddressFromBech32(agentWalletAddr)

	// First: operator gets 70%.
	require.True(t, f.bankKeeper.moduleToAccSent[0].To.Equals(opAddr))
	require.Equal(t, math.NewInt(700000), f.bankKeeper.moduleToAccSent[0].Amount.AmountOf("usum"))

	// Second: agent gets 30%.
	require.True(t, f.bankKeeper.moduleToAccSent[1].To.Equals(agAddr))
	require.Equal(t, math.NewInt(300000), f.bankKeeper.moduleToAccSent[1].Amount.AmountOf("usum"))
}

func TestDistributeActivityRewards_MultipleNodes_EqualSplit(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// 3 nodes, agent_share = 0 for all
	for i := 1; i <= 3; i++ {
		nodeID := "node" + string(rune('0'+i))
		opAddr := makeAddr("op" + string(rune('0'+i)))
		agAddr := makeAddr("ag" + string(rune('0'+i)))
		f.nodeKeeper.registerFullNode(nodeID, "submitter"+string(rune('0'+i)), opAddr, agAddr, 2, math.LegacyZeroDec())
		f.keeper.EpochSummary.Set(ctx, collections.Join(nodeID, int64(0)), types.EpochActivitySummary{
			TotalActivities: 10,
			ActiveWindows:   8,
			Eligible:        true,
		})
	}

	// Fund pool with 1000 usum.
	poolAmount := math.NewInt(1000)
	f.bankKeeper.fundModule(types.ActivityRewardPoolName, sdk.NewCoins(sdk.NewCoin("usum", poolAmount)))

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	// Each gets 333 usum (1000/3 = 333, dust = 1 stays in pool).
	require.Len(t, f.bankKeeper.moduleToAccSent, 3)
	for _, transfer := range f.bankKeeper.moduleToAccSent {
		require.Equal(t, math.NewInt(333), transfer.Amount.AmountOf("usum"))
	}
}

func TestDistributeActivityRewards_Dust(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// 3 nodes, 100 usum pool → 33 each, 1 usum dust.
	for i := 1; i <= 3; i++ {
		nodeID := "node" + string(rune('0'+i))
		opAddr := makeAddr("op" + string(rune('0'+i)))
		agAddr := makeAddr("ag" + string(rune('0'+i)))
		f.nodeKeeper.registerFullNode(nodeID, "sub"+string(rune('0'+i)), opAddr, agAddr, 2, math.LegacyZeroDec())
		f.keeper.EpochSummary.Set(ctx, collections.Join(nodeID, int64(0)), types.EpochActivitySummary{
			TotalActivities: 10,
			ActiveWindows:   8,
			Eligible:        true,
		})
	}

	f.bankKeeper.fundModule(types.ActivityRewardPoolName, sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(100))))

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	// Total distributed = 33 * 3 = 99.
	totalDistributed := math.ZeroInt()
	for _, transfer := range f.bankKeeper.moduleToAccSent {
		totalDistributed = totalDistributed.Add(transfer.Amount.AmountOf("usum"))
	}
	require.Equal(t, math.NewInt(99), totalDistributed)

	// Event should report pool_remaining = 1.
	event := requireEvent(t, ctx, "activity_rewards_distributed")
	require.Equal(t, "1", eventAttribute(event, "pool_remaining"))
}

func TestDistributeActivityRewards_IneligibleNodesExcluded(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// node1: eligible, node2: not eligible
	f.nodeKeeper.registerFullNode("node1", "sub1", makeAddr("op1"), makeAddr("ag1"), 2, math.LegacyZeroDec())
	f.nodeKeeper.registerFullNode("node2", "sub2", makeAddr("op2"), makeAddr("ag2"), 2, math.LegacyZeroDec())

	f.keeper.EpochSummary.Set(ctx, collections.Join("node1", int64(0)), types.EpochActivitySummary{
		TotalActivities: 10,
		ActiveWindows:   8,
		Eligible:        true,
	})
	f.keeper.EpochSummary.Set(ctx, collections.Join("node2", int64(0)), types.EpochActivitySummary{
		TotalActivities: 5,
		ActiveWindows:   5,
		Eligible:        false,
	})

	f.bankKeeper.fundModule(types.ActivityRewardPoolName, sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1000))))

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	// Only 1 transfer (to node1's operator).
	require.Len(t, f.bankKeeper.moduleToAccSent, 1)
	require.Equal(t, math.NewInt(1000), f.bankKeeper.moduleToAccSent[0].Amount.AmountOf("usum"))
}

func TestDistributeActivityRewards_Event(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	f.nodeKeeper.registerFullNode("node1", "sub1", makeAddr("op1"), makeAddr("ag1"), 2, math.LegacyNewDec(20))
	f.keeper.EpochSummary.Set(ctx, collections.Join("node1", int64(0)), types.EpochActivitySummary{
		TotalActivities: 10,
		ActiveWindows:   8,
		Eligible:        true,
	})

	f.bankKeeper.fundModule(types.ActivityRewardPoolName, sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(500000))))

	err := f.keeper.DistributeActivityRewards(ctx, 0)
	require.NoError(t, err)

	event := requireEvent(t, ctx, "activity_rewards_distributed")
	require.Equal(t, "0", eventAttribute(event, types.AttributeKeyEpoch))
	require.Equal(t, "1", eventAttribute(event, types.AttributeKeyEligibleCount))
	require.Equal(t, "500000", eventAttribute(event, "per_node_reward"))
	require.Equal(t, "500000", eventAttribute(event, "total_distributed"))
	require.Equal(t, "0", eventAttribute(event, "pool_remaining"))
}

func TestDistributeCollectedFees_80_20_Split(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// Collect 10000 usum in fees for epoch 0.
	f.keeper.EpochCollectedFees.Set(ctx, int64(0), uint64(10000))
	// Fund activity module account (where fees accumulate).
	f.bankKeeper.fundModule(types.ModuleName, sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(10000))))

	err := f.keeper.DistributeCollectedFees(ctx, 0)
	require.NoError(t, err)

	// 80% → activity_reward_pool.
	require.Len(t, f.bankKeeper.moduleToModSent, 1)
	require.Equal(t, types.ModuleName, f.bankKeeper.moduleToModSent[0].FromModule)
	require.Equal(t, types.ActivityRewardPoolName, f.bankKeeper.moduleToModSent[0].ToModule)
	require.Equal(t, math.NewInt(8000), f.bankKeeper.moduleToModSent[0].Amount.AmountOf("usum"))

	// 20% → community pool.
	require.Equal(t, math.NewInt(2000), f.distributionKeeper.communityPoolFunded.AmountOf("usum"))

	// EpochCollectedFees should be cleared.
	_, err = f.keeper.EpochCollectedFees.Get(ctx, int64(0))
	require.Error(t, err)
}

func TestDistributeCollectedFees_NoFees(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	err := f.keeper.DistributeCollectedFees(ctx, 0)
	require.NoError(t, err)

	require.Empty(t, f.bankKeeper.moduleToModSent)
}

func TestDistributeCollectedFees_CustomRatio(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	// Set ratio to 5000 (50%).
	params, _ := f.keeper.Params.Get(ctx)
	params.FeeToActivityPoolRatio = 5000
	f.keeper.Params.Set(ctx, params)

	f.keeper.EpochCollectedFees.Set(ctx, int64(0), uint64(10000))
	f.bankKeeper.fundModule(types.ModuleName, sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(10000))))

	err := f.keeper.DistributeCollectedFees(ctx, 0)
	require.NoError(t, err)

	// 50% → activity_reward_pool.
	require.Equal(t, math.NewInt(5000), f.bankKeeper.moduleToModSent[0].Amount.AmountOf("usum"))
	// 50% → community pool.
	require.Equal(t, math.NewInt(5000), f.distributionKeeper.communityPoolFunded.AmountOf("usum"))
}

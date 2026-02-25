package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/activity/types"
)

// DistributeActivityRewards distributes the activity_reward_pool balance equally
// to all nodes that met the activity threshold (Eligible==true) in the completed epoch.
//
// For each eligible node, the reward is split between operator and agent according
// to the node's agent_share setting:
//   - agent receives: perNodeReward × agent_share / 100
//   - operator receives: perNodeReward - agentAmount
//
// Dust (remainder from integer division) stays in the activity_reward_pool.
// Called at epoch boundary in EndBlocker, after DistributeCollectedFees.
func (k Keeper) DistributeActivityRewards(ctx context.Context, epoch int64) error {
	if k.bankKeeper == nil || k.authKeeper == nil || k.nodeKeeper == nil {
		return nil // keepers not wired yet
	}

	// Get the activity_reward_pool balance.
	poolAddr := k.authKeeper.GetModuleAddress(types.ActivityRewardPoolName)
	if poolAddr == nil {
		return nil
	}
	poolBalance := k.bankKeeper.GetBalance(ctx, poolAddr, "usum")
	if poolBalance.IsZero() {
		return nil
	}

	// Collect eligible nodes for this epoch.
	eligibleNodes := k.getEligibleNodeIDs(ctx, epoch)
	if len(eligibleNodes) == 0 {
		return nil // no eligible nodes, pool carries over
	}

	// Calculate per-node reward (integer division, dust stays in pool).
	nA := int64(len(eligibleNodes))
	perNodeReward := poolBalance.Amount.Quo(math.NewInt(nA))
	if perNodeReward.IsZero() {
		return nil // pool too small to distribute
	}

	totalDistributed := math.ZeroInt()
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	for _, nodeID := range eligibleNodes {
		// Get node's operator and agent addresses.
		operatorAddr, err := k.nodeKeeper.GetNodeOperatorAddress(ctx, nodeID)
		if err != nil {
			continue // skip nodes with missing data
		}
		agentAddr, err := k.nodeKeeper.GetNodeAgentAddress(ctx, nodeID)
		if err != nil {
			continue
		}
		agentShare, err := k.nodeKeeper.GetNodeAgentShare(ctx, nodeID)
		if err != nil {
			agentShare = math.LegacyZeroDec()
		}

		// Calculate agent and operator portions.
		// agent_share is in percentage (0-100).
		hundred := math.LegacyNewDec(100)
		agentAmount := agentShare.MulInt(perNodeReward).Quo(hundred).TruncateInt()
		operatorAmount := perNodeReward.Sub(agentAmount)

		// Send operator portion.
		if operatorAmount.IsPositive() {
			opAddr, err := sdk.AccAddressFromBech32(operatorAddr)
			if err != nil {
				continue
			}
			opCoins := sdk.NewCoins(sdk.NewCoin("usum", operatorAmount))
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ActivityRewardPoolName, opAddr, opCoins); err != nil {
				return fmt.Errorf("failed to send activity reward to operator %s: %w", operatorAddr, err)
			}
			totalDistributed = totalDistributed.Add(operatorAmount)
		}

		// Send agent portion.
		if agentAmount.IsPositive() && agentAddr != "" {
			agAddr, err := sdk.AccAddressFromBech32(agentAddr)
			if err != nil {
				// If agent address is invalid, send agent portion to operator.
				opAddr, _ := sdk.AccAddressFromBech32(operatorAddr)
				agentCoins := sdk.NewCoins(sdk.NewCoin("usum", agentAmount))
				if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ActivityRewardPoolName, opAddr, agentCoins); err != nil {
					return fmt.Errorf("failed to send agent portion to operator %s: %w", operatorAddr, err)
				}
				totalDistributed = totalDistributed.Add(agentAmount)
				continue
			}
			agentCoins := sdk.NewCoins(sdk.NewCoin("usum", agentAmount))
			if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ActivityRewardPoolName, agAddr, agentCoins); err != nil {
				return fmt.Errorf("failed to send activity reward to agent %s: %w", agentAddr, err)
			}
			totalDistributed = totalDistributed.Add(agentAmount)
		}
	}

	// Emit distribution event.
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"activity_rewards_distributed",
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute(types.AttributeKeyEligibleCount, fmt.Sprintf("%d", nA)),
		sdk.NewAttribute("per_node_reward", perNodeReward.String()),
		sdk.NewAttribute("total_distributed", totalDistributed.String()),
		sdk.NewAttribute("pool_remaining", poolBalance.Amount.Sub(totalDistributed).String()),
	))

	return nil
}

// getEligibleNodeIDs returns all node IDs that met the activity threshold for the given epoch.
func (k Keeper) getEligibleNodeIDs(ctx context.Context, epoch int64) []string {
	var eligible []string

	iter, err := k.EpochSummary.Iterate(ctx, nil)
	if err != nil {
		return nil
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			continue
		}
		if kv.Key.K2() != epoch {
			continue
		}
		if kv.Value.Eligible {
			eligible = append(eligible, kv.Key.K1())
		}
	}

	return eligible
}

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"seocheon/x/activity/types"
)

// BeginBlocker splits minted block rewards between the activity reward pool and
// the standard DPoS distribution. It runs AFTER x/mint (which deposits into
// fee_collector) and BEFORE x/distribution (which distributes fee_collector to
// validators/delegators).
//
// Flow: fee_collector balance × activity_ratio → activity_reward_pool
// Remainder stays in fee_collector for x/distribution to process.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	if k.bankKeeper == nil || k.authKeeper == nil {
		return nil // keepers not wired yet (during genesis init)
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Get N_d (active validator set size).
	var nD uint64
	if k.stakingKeeper != nil {
		validators, err := k.stakingKeeper.GetBondedValidatorsByPower(ctx)
		if err == nil {
			nD = uint64(len(validators))
		}
	}
	if nD == 0 {
		nD = 1 // prevent division by zero
	}

	// Get N_a (eligible activity nodes from the current or previous epoch).
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	epoch := GetCurrentEpoch(blockHeight, params)
	prevEpoch := epoch - 1
	var nA uint64
	if prevEpoch >= 0 {
		nA = uint64(len(k.getEligibleNodeIDs(ctx, prevEpoch)))
	}

	// Calculate activity ratio.
	activityRatio := CalculateActivityRatio(nA, nD, params.DMin)
	if activityRatio.IsZero() {
		return nil // nothing to split
	}

	// Get fee_collector balance.
	feeCollectorAddr := k.authKeeper.GetModuleAddress(authtypes.FeeCollectorName)
	if feeCollectorAddr == nil {
		return nil
	}
	feeCollectorBalance := k.bankKeeper.GetBalance(ctx, feeCollectorAddr, "usum")
	if feeCollectorBalance.IsZero() {
		return nil
	}

	// Calculate activity pool amount: fee_collector_balance × activity_ratio.
	activityAmount := activityRatio.MulInt(feeCollectorBalance.Amount).TruncateInt()
	if activityAmount.IsZero() || !activityAmount.IsPositive() {
		return nil
	}

	// Transfer from fee_collector to activity_reward_pool.
	activityCoins := sdk.NewCoins(sdk.NewCoin("usum", activityAmount))
	if err := k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		authtypes.FeeCollectorName,
		types.ActivityRewardPoolName,
		activityCoins,
	); err != nil {
		return fmt.Errorf("failed to split block rewards to activity pool: %w", err)
	}

	// Track accumulated reward pool for this epoch.
	current, err := k.EpochActivityRewardPool.Get(ctx, epoch)
	if err != nil {
		current = 0
	}
	if err := k.EpochActivityRewardPool.Set(ctx, epoch, current+activityAmount.Uint64()); err != nil {
		return err
	}

	return nil
}

// CalculateActivityRatio computes the fraction of block rewards allocated to the
// activity reward pool.
//
//	delegation_ratio = max(D_min, N_d / (N_a + N_d))
//	activity_ratio   = 1 - delegation_ratio
//
// D_min is in basis points (e.g., 3000 = 0.3).
// Returns a LegacyDec in range [0, 1-D_min].
func CalculateActivityRatio(nA, nD uint64, dMinBasisPoints uint64) math.LegacyDec {
	if nA == 0 {
		return math.LegacyZeroDec() // no activity nodes → all rewards to DPoS
	}

	dMin := math.LegacyNewDecWithPrec(int64(dMinBasisPoints), 4) // basis points → decimal

	// N_d / (N_a + N_d)
	total := math.LegacyNewDec(int64(nA + nD))
	naturalDelegationRatio := math.LegacyNewDec(int64(nD)).Quo(total)

	// delegation_ratio = max(D_min, natural_ratio)
	delegationRatio := naturalDelegationRatio
	if dMin.GT(delegationRatio) {
		delegationRatio = dMin
	}

	// activity_ratio = 1 - delegation_ratio
	activityRatio := math.LegacyOneDec().Sub(delegationRatio)
	if activityRatio.IsNegative() {
		return math.LegacyZeroDec()
	}

	return activityRatio
}

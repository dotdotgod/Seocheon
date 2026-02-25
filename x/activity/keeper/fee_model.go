package keeper

import (
	"context"
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/activity/types"
)

// CalculateAndCacheEpochFeeState computes the fee parameters for the given epoch
// and stores them in EpochActivityFee and EpochEffectiveQuota.
// Called once at epoch boundary in EndBlocker.
func (k Keeper) CalculateAndCacheEpochFeeState(ctx context.Context, epoch int64) error {
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

	// Get N_a (eligible activity nodes from previous epoch).
	prevEpoch := epoch - 1
	var nA uint64
	if prevEpoch >= 0 {
		nA = uint64(k.countEligibleNodes(ctx, prevEpoch))
	}

	// Calculate fee and effective quota.
	activityFee := calculateActivityFee(nA, nD, params)
	effectiveQuota := calculateEffectiveQuota(nA, nD, params)

	// Cache the values.
	if err := k.EpochActivityFee.Set(ctx, epoch, activityFee); err != nil {
		return err
	}
	if err := k.EpochEffectiveQuota.Set(ctx, epoch, effectiveQuota); err != nil {
		return err
	}

	// Emit event.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeEpochFeeState,
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("activity_fee", fmt.Sprintf("%d", activityFee)),
		sdk.NewAttribute("effective_feegrant_quota", fmt.Sprintf("%d", effectiveQuota)),
		sdk.NewAttribute("n_a", fmt.Sprintf("%d", nA)),
		sdk.NewAttribute("n_d", fmt.Sprintf("%d", nD)),
	))

	return nil
}

// calculateActivityFee computes the activity fee based on saturation ratio.
// S = N_a / (N_d × fee_threshold_multiplier)
// If S <= 1.0: fee = 0
// If S > 1.0: fee = base_fee × (S - 1)^exponent, capped at max_fee
func calculateActivityFee(nA, nD uint64, params types.Params) uint64 {
	if params.FeeThresholdMultiplier == 0 || params.BaseActivityFee == 0 {
		return 0
	}

	threshold := nD * params.FeeThresholdMultiplier
	if nA <= threshold {
		return 0 // Bootstrap phase: no fees
	}

	// S - 1 = (N_a - threshold) / threshold
	excess := nA - threshold
	sMinus1 := float64(excess) / float64(threshold)

	// Apply exponent (basis points: 5000 = 0.5)
	exponent := float64(params.FeeExponent) / 10000.0
	if exponent <= 0 {
		exponent = 0.5 // default sqrt curve
	}

	factor := math.Pow(sMinus1, exponent)
	fee := uint64(float64(params.BaseActivityFee) * factor)

	// Cap at max_activity_fee.
	if params.MaxActivityFee > 0 && fee > params.MaxActivityFee {
		fee = params.MaxActivityFee
	}

	return fee
}

// calculateEffectiveQuota computes the adjusted feegrant quota under saturation.
// effective_quota = max(min_quota, quota - floor(quota × reduction_rate × (S - 1)))
func calculateEffectiveQuota(nA, nD uint64, params types.Params) uint64 {
	if params.FeeThresholdMultiplier == 0 {
		return params.FeegrantQuota
	}

	threshold := nD * params.FeeThresholdMultiplier
	if nA <= threshold {
		return params.FeegrantQuota // Bootstrap: no reduction
	}

	// S - 1 = (N_a - threshold) / threshold
	excess := nA - threshold
	sMinus1 := float64(excess) / float64(threshold)

	// reduction_rate in basis points (5000 = 0.5)
	rate := float64(params.QuotaReductionRate) / 10000.0
	reduction := uint64(float64(params.FeegrantQuota) * rate * sMinus1)

	if reduction >= params.FeegrantQuota {
		return params.MinFeegrantQuota
	}

	effective := params.FeegrantQuota - reduction
	if effective < params.MinFeegrantQuota {
		effective = params.MinFeegrantQuota
	}

	return effective
}

// GetEpochActivityFee returns the cached activity fee for the given epoch.
// Returns 0 if not cached (bootstrap / epoch 0).
func (k Keeper) GetEpochActivityFee(ctx context.Context, epoch int64) uint64 {
	fee, err := k.EpochActivityFee.Get(ctx, epoch)
	if err != nil {
		return 0
	}
	return fee
}

// GetEpochEffectiveQuota returns the cached effective feegrant quota for the given epoch.
// Returns the default feegrant quota if not cached.
func (k Keeper) GetEpochEffectiveQuota(ctx context.Context, epoch int64) uint64 {
	quota, err := k.EpochEffectiveQuota.Get(ctx, epoch)
	if err != nil {
		// Not cached yet; return default.
		params, err := k.Params.Get(ctx)
		if err != nil {
			return 10 // safe fallback
		}
		return params.FeegrantQuota
	}
	return quota
}

// CollectActivityFee records a fee collected from a self-funded node.
func (k Keeper) CollectActivityFee(ctx context.Context, epoch int64, amount uint64) error {
	current, err := k.EpochCollectedFees.Get(ctx, epoch)
	if err != nil {
		current = 0
	}
	return k.EpochCollectedFees.Set(ctx, epoch, current+amount)
}

// DistributeCollectedFees distributes collected activity fees at epoch boundary.
// fee_to_activity_pool_ratio% (default 80%) → activity_reward_pool (for equal distribution to eligible nodes).
// Remainder (default 20%) → community pool via x/distribution FundCommunityPool.
func (k Keeper) DistributeCollectedFees(ctx context.Context, epoch int64) error {
	total, err := k.EpochCollectedFees.Get(ctx, epoch)
	if err != nil || total == 0 {
		return nil // no fees to distribute
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	totalInt := sdkmath.NewIntFromUint64(total)

	// Calculate split amounts.
	activityPoolRatio := sdkmath.LegacyNewDecWithPrec(int64(params.FeeToActivityPoolRatio), 4)
	activityAmt := activityPoolRatio.MulInt(totalInt).TruncateInt()
	communityAmt := totalInt.Sub(activityAmt)

	// 80% → activity_reward_pool
	if activityAmt.IsPositive() && k.bankKeeper != nil {
		activityCoins := sdk.NewCoins(sdk.NewCoin("usum", activityAmt))
		if err := k.bankKeeper.SendCoinsFromModuleToModule(
			ctx,
			types.ModuleName,
			types.ActivityRewardPoolName,
			activityCoins,
		); err != nil {
			return fmt.Errorf("failed to send fees to activity reward pool: %w", err)
		}
	}

	// 20% → community pool
	if communityAmt.IsPositive() && k.bankKeeper != nil && k.distributionKeeper != nil && k.authKeeper != nil {
		communityCoins := sdk.NewCoins(sdk.NewCoin("usum", communityAmt))
		moduleAddr := k.authKeeper.GetModuleAddress(types.ModuleName)
		if moduleAddr != nil {
			if err := k.distributionKeeper.FundCommunityPool(ctx, communityCoins, moduleAddr); err != nil {
				return fmt.Errorf("failed to fund community pool from activity fees: %w", err)
			}
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeFeesCollected,
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute("total_fees", fmt.Sprintf("%d", total)),
		sdk.NewAttribute("activity_pool_amount", activityAmt.String()),
		sdk.NewAttribute("community_pool_amount", communityAmt.String()),
	))

	// Clear collected fees for this epoch.
	return k.EpochCollectedFees.Remove(ctx, epoch)
}

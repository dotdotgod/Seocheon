package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	nodetypes "seocheon/x/node/types"

	"seocheon/x/activity/types"
)

// SubmitActivity handles MsgSubmitActivity.
func (ms msgServer) SubmitActivity(ctx context.Context, msg *types.MsgSubmitActivity) (*types.MsgSubmitActivityResponse, error) {
	// 1. Validate activity_hash (64 hex chars = 32 bytes).
	if !ValidateActivityHash(msg.ActivityHash) {
		return nil, types.ErrInvalidActivityHash
	}

	// 2. Validate content_uri is not empty.
	if msg.ContentUri == "" {
		return nil, types.ErrInvalidContentURI
	}

	// 3. Look up node by agent address.
	nodeID, err := ms.nodeKeeper.GetNodeIDByAgent(ctx, msg.Submitter)
	if err != nil {
		return nil, types.ErrSubmitterNotRegistered.Wrap(err.Error())
	}

	// 4. Check node status (must be REGISTERED=1 or ACTIVE=2).
	status, err := ms.nodeKeeper.GetNodeStatus(ctx, nodeID)
	if err != nil {
		return nil, types.ErrNodeNotFound.Wrap(err.Error())
	}
	if status != 1 && status != 2 {
		return nil, types.ErrNodeNotEligible
	}

	// 5. Get params and compute epoch/window.
	params, err := ms.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()
	epoch := GetCurrentEpoch(blockHeight, params)
	window := GetCurrentWindow(blockHeight, params)

	// 6. Check for duplicate hash within the same epoch.
	hasDuplicate, err := ms.HashIndex.Has(ctx, collections.Join3(nodeID, epoch, msg.ActivityHash))
	if err != nil {
		return nil, err
	}
	if hasDuplicate {
		return nil, types.ErrDuplicateActivityHash
	}

	// 7. Determine quota (feegrant vs self-funded).
	quota := params.SelfFundedQuota
	quotaType := "self_funded"
	isFeegrantNode := false

	if ms.feegrantKeeper != nil && ms.authKeeper != nil {
		feegrantPoolAddr := ms.authKeeper.GetModuleAddress(nodetypes.FeegrantPoolName)
		if feegrantPoolAddr != nil {
			submitterAddr, err := sdk.AccAddressFromBech32(msg.Submitter)
			if err == nil {
				allowance, err := ms.feegrantKeeper.GetAllowance(ctx, feegrantPoolAddr, submitterAddr)
				if err == nil && allowance != nil {
					isFeegrantNode = true
					quotaType = "feegrant"
					// Use effective quota (adjusted for saturation).
					quota = uint64(ms.GetEpochEffectiveQuota(ctx, epoch))
				}
			}
		}
	}

	// 8. Check quota.
	quotaUsed, err := ms.EpochQuotaUsed.Get(ctx, collections.Join(nodeID, epoch))
	if err != nil {
		quotaUsed = 0 // first activity in this epoch
	}
	if quotaUsed >= quota {
		return nil, types.ErrQuotaExceeded.Wrapf("used %d of %d (%s)", quotaUsed, quota, quotaType)
	}

	// 8.5. Charge activity fee (if applicable).
	activityFee := ms.GetEpochActivityFee(ctx, epoch)
	if activityFee > 0 && !(isFeegrantNode && params.FeegrantFeeExempt) {
		// Self-funded nodes (or feegrant nodes when feegrant_fee_exempt=false) pay the fee.
		// Actually transfer the fee from submitter to the activity module account.
		if ms.bankKeeper != nil {
			submitterAddr, addrErr := sdk.AccAddressFromBech32(msg.Submitter)
			if addrErr != nil {
				return nil, addrErr
			}
			feeCoins := sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewIntFromUint64(activityFee)))
			if err := ms.bankKeeper.SendCoinsFromAccountToModule(ctx, submitterAddr, types.ModuleName, feeCoins); err != nil {
				return nil, fmt.Errorf("failed to collect activity fee: %w", err)
			}
		}

		// Track fee amount for epoch-boundary distribution.
		if err := ms.CollectActivityFee(ctx, epoch, activityFee); err != nil {
			return nil, err
		}

		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeActivityFeeCharged,
			sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
			sdk.NewAttribute("fee_amount", fmt.Sprintf("%d", activityFee)),
			sdk.NewAttribute(types.AttributeKeyQuotaType, quotaType),
		))
	}

	// 9. Get next sequence.
	seq, err := ms.ActivitySequence.Get(ctx, collections.Join(nodeID, epoch))
	if err != nil {
		seq = 0
	}

	// 10. Create and store ActivityRecord.
	record := types.ActivityRecord{
		NodeId:       nodeID,
		Epoch:        epoch,
		Sequence:     seq,
		Submitter:    msg.Submitter,
		ActivityHash: msg.ActivityHash,
		ContentUri:   msg.ContentUri,
		BlockHeight:  blockHeight,
	}

	if err := ms.Activities.Set(ctx, collections.Join3(nodeID, epoch, seq), record); err != nil {
		return nil, err
	}

	// Update HashIndex.
	if err := ms.HashIndex.Set(ctx, collections.Join3(nodeID, epoch, msg.ActivityHash)); err != nil {
		return nil, err
	}

	// Update quota used.
	if err := ms.EpochQuotaUsed.Set(ctx, collections.Join(nodeID, epoch), quotaUsed+1); err != nil {
		return nil, err
	}

	// Update window activity count.
	windowCount, err := ms.WindowActivity.Get(ctx, collections.Join3(nodeID, epoch, window))
	if err != nil {
		windowCount = 0
	}
	if err := ms.WindowActivity.Set(ctx, collections.Join3(nodeID, epoch, window), windowCount+1); err != nil {
		return nil, err
	}

	// Update BlockActivities index.
	blockSeq, err := ms.getNextBlockSeq(ctx, blockHeight)
	if err != nil {
		return nil, err
	}
	if err := ms.BlockActivities.Set(ctx, collections.Join(blockHeight, blockSeq), nodeID); err != nil {
		return nil, err
	}

	// Increment sequence.
	if err := ms.ActivitySequence.Set(ctx, collections.Join(nodeID, epoch), seq+1); err != nil {
		return nil, err
	}

	// 11. Update EpochSummary.
	summary, err := ms.EpochSummary.Get(ctx, collections.Join(nodeID, epoch))
	if err != nil {
		summary = types.EpochActivitySummary{}
	}
	summary.TotalActivities++

	// If this is the first activity in this window, increment active_windows.
	if windowCount == 0 {
		summary.ActiveWindows++
	}

	// Check eligibility.
	summary.Eligible = int64(summary.ActiveWindows) >= params.MinActiveWindows

	if err := ms.EpochSummary.Set(ctx, collections.Join(nodeID, epoch), summary); err != nil {
		return nil, err
	}

	// 12. Emit event.
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeActivitySubmitted,
		sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
		sdk.NewAttribute(types.AttributeKeySubmitter, msg.Submitter),
		sdk.NewAttribute(types.AttributeKeyActivityHash, msg.ActivityHash),
		sdk.NewAttribute(types.AttributeKeyContentURI, msg.ContentUri),
		sdk.NewAttribute(types.AttributeKeyEpoch, fmt.Sprintf("%d", epoch)),
		sdk.NewAttribute(types.AttributeKeyWindow, fmt.Sprintf("%d", window)),
		sdk.NewAttribute(types.AttributeKeySequence, fmt.Sprintf("%d", seq)),
		sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", blockHeight)),
		sdk.NewAttribute(types.AttributeKeyQuotaType, quotaType),
	))

	return &types.MsgSubmitActivityResponse{
		Epoch:    epoch,
		Sequence: seq,
	}, nil
}

// getNextBlockSeq returns the next sequence number for BlockActivities at the given block height.
func (k Keeper) getNextBlockSeq(ctx context.Context, blockHeight int64) (uint64, error) {
	var maxSeq uint64
	rng := collections.NewPrefixedPairRange[int64, uint64](blockHeight)
	iter, err := k.BlockActivities.Iterate(ctx, rng)
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return 0, err
		}
		seq := key.K2()
		if seq >= maxSeq {
			maxSeq = seq + 1
		}
	}
	return maxSeq, nil
}

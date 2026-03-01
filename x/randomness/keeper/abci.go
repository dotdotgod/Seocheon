package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/randomness/types"
)

// EndBlocker processes pending randomness requests in 3 phases:
// Phase 1: Fulfill requests whose beacon is available.
// Phase 2: Expire timed-out requests and refund fees.
// Phase 3: Prune old fulfilled/expired requests.
func (k Keeper) EndBlocker(ctx context.Context) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get params: %w", err)
	}

	// Skip if commit-reveal is not enabled.
	if !params.CommitRevealEnabled {
		return nil
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()

	// Phase 1: Fulfill pending requests with available beacons.
	if err := k.endBlockerFulfill(ctx, sdkCtx, params, blockHeight); err != nil {
		return fmt.Errorf("fulfill phase: %w", err)
	}

	// Phase 2: Expire timed-out requests.
	if err := k.endBlockerExpire(ctx, sdkCtx, params, blockHeight); err != nil {
		return fmt.Errorf("expire phase: %w", err)
	}

	// Phase 3: Prune old requests.
	if err := k.endBlockerPrune(ctx, params, blockHeight); err != nil {
		return fmt.Errorf("prune phase: %w", err)
	}

	return nil
}

// endBlockerFulfill fulfills pending requests whose target_round beacons are available.
func (k Keeper) endBlockerFulfill(ctx context.Context, sdkCtx sdk.Context, params types.Params, blockHeight int64) error {
	fulfilled := uint64(0)

	// Walk PendingByRound to find requests whose beacon is available.
	var toFulfill []collections.Pair[uint64, uint64]
	err := k.PendingByRound.Walk(ctx, nil, func(key collections.Pair[uint64, uint64]) (bool, error) {
		if fulfilled >= params.MaxFulfillsPerBlock {
			return true, nil // stop
		}

		targetRound := key.K1()

		// Check if beacon exists for this round.
		_, err := k.Beacons.Get(ctx, targetRound)
		if err != nil {
			return false, nil // beacon not yet available, skip
		}

		toFulfill = append(toFulfill, key)
		fulfilled++
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk pending requests: %w", err)
	}

	// Process fulfillments.
	for _, key := range toFulfill {
		targetRound := key.K1()
		requestID := key.K2()

		request, err := k.Requests.Get(ctx, requestID)
		if err != nil {
			continue // request not found, skip
		}
		if request.Status != types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING {
			// Clean up stale index entry.
			_ = k.PendingByRound.Remove(ctx, key)
			continue
		}

		beacon, err := k.Beacons.Get(ctx, targetRound)
		if err != nil {
			continue
		}

		// Multi-source mixing: result = SHA256(beacon.randomness || AppHash || commit_hash || word_index)
		appHash := hex.EncodeToString(sdkCtx.HeaderHash())
		var resultWords []string
		for i := uint32(0); i < request.NumWords; i++ {
			preimage := beacon.Randomness + appHash + request.CommitHash + fmt.Sprintf("%d", i)
			hash := sha256.Sum256([]byte(preimage))
			resultWords = append(resultWords, hex.EncodeToString(hash[:]))
		}

		// Concatenate all result words.
		var resultConcat string
		for _, w := range resultWords {
			resultConcat += w
		}

		// Update request to FULFILLED.
		request.Status = types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_FULFILLED
		request.FulfilledAt = blockHeight
		request.Result = resultConcat
		request.BeaconAppHash = appHash

		if err := k.Requests.Set(ctx, requestID, request); err != nil {
			return fmt.Errorf("failed to update fulfilled request %d: %w", requestID, err)
		}

		// Remove from pending index.
		if err := k.PendingByRound.Remove(ctx, key); err != nil {
			return fmt.Errorf("failed to remove pending index for request %d: %w", requestID, err)
		}

		// Add to fulfilled block index for pruning.
		if err := k.FulfilledBlockIndex.Set(ctx, collections.Join(blockHeight, requestID)); err != nil {
			return fmt.Errorf("failed to set fulfilled index for request %d: %w", requestID, err)
		}

		// Decrement pending count.
		pendingCount, _ := k.PendingRequestCount.Get(ctx)
		if pendingCount > 0 {
			_ = k.PendingRequestCount.Set(ctx, pendingCount-1)
		}

		// Pay relayer fee: transfer escrowed fee to the beacon submitter.
		if request.RequestFee.IsPositive() {
			relayerAddr, err := sdk.AccAddressFromBech32(beacon.Submitter)
			if err == nil {
				feeCoins := sdk.NewCoins(request.RequestFee)
				_ = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, relayerAddr, feeCoins)
			}
		}

		// Emit fulfillment event.
		sdkCtx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeRandomnessFulfilled,
				sdk.NewAttribute(types.AttributeKeyRequestID, fmt.Sprintf("%d", requestID)),
				sdk.NewAttribute(types.AttributeKeyRequester, request.Requester),
				sdk.NewAttribute(types.AttributeKeyTargetRound, fmt.Sprintf("%d", targetRound)),
				sdk.NewAttribute(types.AttributeKeyResultHash, resultConcat[:64]), // first word
				sdk.NewAttribute(types.AttributeKeyBeaconSubmitter, beacon.Submitter),
				sdk.NewAttribute(types.AttributeKeyFeePaid, request.RequestFee.String()),
			),
		})
	}

	return nil
}

// endBlockerExpire expires pending requests that have exceeded request_timeout_blocks.
func (k Keeper) endBlockerExpire(ctx context.Context, sdkCtx sdk.Context, params types.Params, blockHeight int64) error {
	var toExpire []collections.Pair[uint64, uint64]

	err := k.PendingByRound.Walk(ctx, nil, func(key collections.Pair[uint64, uint64]) (bool, error) {
		requestID := key.K2()

		request, err := k.Requests.Get(ctx, requestID)
		if err != nil {
			return false, nil
		}

		if request.Status != types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING {
			toExpire = append(toExpire, key) // stale index, clean up
			return false, nil
		}

		// Check if request has timed out.
		if blockHeight-request.CreatedAt > int64(params.RequestTimeoutBlocks) {
			toExpire = append(toExpire, key)
		}

		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk pending requests for expiry: %w", err)
	}

	for _, key := range toExpire {
		requestID := key.K2()

		request, err := k.Requests.Get(ctx, requestID)
		if err != nil {
			_ = k.PendingByRound.Remove(ctx, key)
			continue
		}

		if request.Status != types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING {
			_ = k.PendingByRound.Remove(ctx, key)
			continue
		}

		// Mark as EXPIRED.
		request.Status = types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_EXPIRED
		request.FulfilledAt = blockHeight // Record expiry block for pruning.

		if err := k.Requests.Set(ctx, requestID, request); err != nil {
			return fmt.Errorf("failed to update expired request %d: %w", requestID, err)
		}

		// Remove from pending index.
		_ = k.PendingByRound.Remove(ctx, key)

		// Add to fulfilled block index for pruning.
		_ = k.FulfilledBlockIndex.Set(ctx, collections.Join(blockHeight, requestID))

		// Decrement pending count.
		pendingCount, _ := k.PendingRequestCount.Get(ctx)
		if pendingCount > 0 {
			_ = k.PendingRequestCount.Set(ctx, pendingCount-1)
		}

		// Refund escrowed fee to requester.
		if request.RequestFee.IsPositive() {
			requesterAddr, err := sdk.AccAddressFromBech32(request.Requester)
			if err == nil {
				feeCoins := sdk.NewCoins(request.RequestFee)
				_ = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, requesterAddr, feeCoins)
			}
		}

		// Emit expiry event.
		sdkCtx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeRandomnessExpired,
				sdk.NewAttribute(types.AttributeKeyRequestID, fmt.Sprintf("%d", requestID)),
				sdk.NewAttribute(types.AttributeKeyRequester, request.Requester),
				sdk.NewAttribute(types.AttributeKeyTargetRound, fmt.Sprintf("%d", request.TargetRound)),
				sdk.NewAttribute(types.AttributeKeyRefundAmount, request.RequestFee.String()),
			),
		})
	}

	return nil
}

// endBlockerPrune prunes fulfilled/expired requests older than request_pruning_blocks.
func (k Keeper) endBlockerPrune(ctx context.Context, params types.Params, blockHeight int64) error {
	pruneBeforeHeight := blockHeight - int64(params.RequestPruningBlocks)
	if pruneBeforeHeight <= 0 {
		return nil
	}

	// Walk FulfilledBlockIndex up to pruneBeforeHeight.
	rng := new(collections.Range[collections.Pair[int64, uint64]]).
		EndInclusive(collections.Join(pruneBeforeHeight, uint64(^uint64(0))))

	var toPrune []collections.Pair[int64, uint64]
	err := k.FulfilledBlockIndex.Walk(ctx, rng, func(key collections.Pair[int64, uint64]) (bool, error) {
		toPrune = append(toPrune, key)
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk fulfilled index for pruning: %w", err)
	}

	for _, key := range toPrune {
		requestID := key.K2()

		// Get request for cleanup.
		request, err := k.Requests.Get(ctx, requestID)
		if err == nil {
			// Remove from requester index.
			_ = k.RequestsByRequester.Remove(ctx, collections.Join(request.Requester, requestID))
		}

		// Remove request.
		_ = k.Requests.Remove(ctx, requestID)

		// Remove from fulfilled index.
		_ = k.FulfilledBlockIndex.Remove(ctx, key)
	}

	return nil
}

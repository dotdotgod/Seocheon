package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/randomness/types"
)

// RequestRandomness handles MsgRequestRandomness: creates a commit-reveal randomness request.
func (ms msgServer) RequestRandomness(ctx context.Context, msg *types.MsgRequestRandomness) (*types.MsgRequestRandomnessResponse, error) {
	// 1. Get params.
	params, err := ms.k.Params.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	// 2. Check commit_reveal_enabled.
	if !params.CommitRevealEnabled {
		return nil, types.ErrCommitRevealDisabled
	}

	// 3. Validate commit_hash: must be 64-char hex (SHA-256).
	if len(msg.CommitHash) != 64 {
		return nil, types.ErrInvalidCommitHash
	}
	if _, err := hex.DecodeString(msg.CommitHash); err != nil {
		return nil, types.ErrInvalidCommitHash
	}

	// 4. Validate num_words: must be 1-10.
	if msg.NumWords == 0 || msg.NumWords > 10 {
		return nil, types.ErrInvalidNumWords
	}

	// 5. Validate callback_data: must be valid hex, max 256 bytes (512 hex chars).
	if msg.CallbackData != "" {
		if len(msg.CallbackData) > 512 {
			return nil, types.ErrCallbackDataTooLong
		}
		decoded, err := hex.DecodeString(msg.CallbackData)
		if err != nil {
			return nil, types.ErrInvalidCallbackData
		}
		if len(decoded) > 256 {
			return nil, types.ErrCallbackDataTooLong
		}
	}

	// 6. Validate request_fee >= min_request_fee.
	if msg.RequestFee.IsLT(params.MinRequestFee) {
		return nil, types.ErrInsufficientRequestFee
	}

	// 7. Check global pending request limit.
	pendingCount, err := ms.k.PendingRequestCount.Get(ctx)
	if err != nil {
		pendingCount = 0
	}
	if pendingCount >= params.MaxPendingRequests {
		return nil, types.ErrTooManyPendingRequests
	}

	// 8. Check per-requester pending request limit.
	requesterCount := uint64(0)
	rng := collections.NewPrefixedPairRange[string, uint64](msg.Requester)
	err = ms.k.RequestsByRequester.Walk(ctx, rng, func(key collections.Pair[string, uint64]) (bool, error) {
		// Check if this request is still PENDING.
		req, err := ms.k.Requests.Get(ctx, key.K2())
		if err != nil {
			return false, nil // skip if not found
		}
		if req.Status == types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING {
			requesterCount++
		}
		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to count requester requests: %w", err)
	}
	if requesterCount >= params.MaxRequestsPerRequester {
		return nil, types.ErrRequesterLimitExceeded
	}

	// 9. Calculate target_round.
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockTime := sdkCtx.BlockTime().Unix()

	// current_round = (block_time - drand_genesis_time) / drand_period_seconds
	var currentRound uint64
	if blockTime > int64(params.DrandGenesisTime) {
		currentRound = uint64(blockTime-int64(params.DrandGenesisTime)) / params.DrandPeriodSeconds
	}
	targetRound := currentRound + params.MinFutureRounds

	// 10. Escrow the request fee.
	requesterAddr, err := sdk.AccAddressFromBech32(msg.Requester)
	if err != nil {
		return nil, fmt.Errorf("invalid requester address: %w", err)
	}
	feeCoins := sdk.NewCoins(msg.RequestFee)
	if err := ms.k.bankKeeper.SendCoinsFromAccountToModule(ctx, requesterAddr, types.ModuleName, feeCoins); err != nil {
		return nil, fmt.Errorf("failed to escrow request fee: %w", err)
	}

	// 11. Generate request ID.
	requestID, err := ms.k.RequestSequence.Next(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate request ID: %w", err)
	}
	requestID++ // Sequence starts at 0, IDs start at 1.

	// 12. Create and store the request.
	request := types.RandomnessRequest{
		RequestId:    requestID,
		Requester:    msg.Requester,
		CommitHash:   msg.CommitHash,
		NumWords:     msg.NumWords,
		CallbackData: msg.CallbackData,
		TargetRound:  targetRound,
		Status:       types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING,
		CreatedAt:    sdkCtx.BlockHeight(),
		RequestFee:   msg.RequestFee,
	}

	if err := ms.k.Requests.Set(ctx, requestID, request); err != nil {
		return nil, fmt.Errorf("failed to store request: %w", err)
	}

	// 13. Update indexes.
	if err := ms.k.PendingByRound.Set(ctx, collections.Join(targetRound, requestID)); err != nil {
		return nil, fmt.Errorf("failed to set pending index: %w", err)
	}
	if err := ms.k.RequestsByRequester.Set(ctx, collections.Join(msg.Requester, requestID)); err != nil {
		return nil, fmt.Errorf("failed to set requester index: %w", err)
	}
	if err := ms.k.PendingRequestCount.Set(ctx, pendingCount+1); err != nil {
		return nil, fmt.Errorf("failed to update pending count: %w", err)
	}

	// 14. Emit event.
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRandomnessRequested,
			sdk.NewAttribute(types.AttributeKeyRequestID, fmt.Sprintf("%d", requestID)),
			sdk.NewAttribute(types.AttributeKeyRequester, msg.Requester),
			sdk.NewAttribute(types.AttributeKeyCommitHash, msg.CommitHash),
			sdk.NewAttribute(types.AttributeKeyTargetRound, fmt.Sprintf("%d", targetRound)),
			sdk.NewAttribute(types.AttributeKeyNumWords, fmt.Sprintf("%d", msg.NumWords)),
			sdk.NewAttribute(types.AttributeKeyRequestFee, msg.RequestFee.String()),
		),
	})

	return &types.MsgRequestRandomnessResponse{
		RequestId:   requestID,
		TargetRound: targetRound,
	}, nil
}

package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/randomness/types"
)

// SubmitBeacon handles MsgSubmitBeacon: stores a drand randomness beacon on-chain.
func (ms msgServer) SubmitBeacon(ctx context.Context, msg *types.MsgSubmitBeacon) (*types.MsgSubmitBeaconResponse, error) {
	// Validate round.
	if msg.Round == 0 {
		return nil, types.ErrInvalidRound
	}

	// Validate randomness format (64-char hex = 32 bytes).
	if !ValidateRandomnessFormat(msg.Randomness) {
		return nil, types.ErrInvalidRandomness
	}

	// Validate signature format.
	if !ValidateSignatureFormat(msg.Signature) {
		return nil, types.ErrInvalidSignature
	}

	// Check for duplicate.
	has, err := ms.k.Beacons.Has(ctx, msg.Round)
	if err != nil {
		return nil, fmt.Errorf("failed to check beacon existence: %w", err)
	}
	if has {
		return nil, types.ErrDuplicateBeacon
	}

	// Get params for optional verification.
	params, err := ms.k.Params.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get params: %w", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	blockHeight := sdkCtx.BlockHeight()

	// Validate round timing: expected_time = drand_genesis_time + (round * drand_period_seconds).
	// The beacon round should correspond to a time that is not in the future relative to block time.
	if params.DrandGenesisTime > 0 && params.DrandPeriodSeconds > 0 {
		expectedTime := int64(params.DrandGenesisTime) + int64(msg.Round)*int64(params.DrandPeriodSeconds)
		blockTime := sdkCtx.BlockTime().Unix()
		if expectedTime > blockTime {
			return nil, fmt.Errorf("%w: beacon round %d is in the future (expected_time=%d, block_time=%d)",
				types.ErrInvalidRound, msg.Round, expectedTime, blockTime)
		}
	}

	// BLS12-381 signature verification (requires drand/kyber dependency).
	// When beacon_verification_enabled is true, verify the beacon's BLS signature
	// against the drand public key configured in params.
	if params.BeaconVerificationEnabled {
		// Verification steps (to be implemented with drand/kyber-bls12381):
		// 1. Decode params.DrandPublicKey from hex to G2 point
		// 2. Decode msg.Signature from hex to G1 point
		// 3. Hash round number to G1 using the scheme specified by params.DrandSchemeId
		// 4. Verify BLS signature: e(signature, G2_generator) == e(H(round), pubkey)
		// 5. Verify randomness == SHA-256(signature)
		return nil, fmt.Errorf("%w: drand BLS verification not yet implemented (requires drand/kyber dependency)",
			types.ErrVerificationFailed)
	}

	// Store the beacon.
	beacon := types.Beacon{
		Round:       msg.Round,
		Randomness:  msg.Randomness,
		Signature:   msg.Signature,
		SubmittedAt: blockHeight,
		Submitter:   msg.Submitter,
	}

	if err := ms.k.Beacons.Set(ctx, msg.Round, beacon); err != nil {
		return nil, fmt.Errorf("failed to store beacon: %w", err)
	}

	// Update latest round if this is newer.
	latestRound, err := ms.k.LatestRound.Get(ctx)
	if err != nil || msg.Round > latestRound {
		if err := ms.k.LatestRound.Set(ctx, msg.Round); err != nil {
			return nil, fmt.Errorf("failed to update latest round: %w", err)
		}
	}

	// Emit event.
	sdkCtx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeBeaconSubmitted,
			sdk.NewAttribute(types.AttributeKeyRound, fmt.Sprintf("%d", msg.Round)),
			sdk.NewAttribute(types.AttributeKeyRandomness, msg.Randomness),
			sdk.NewAttribute(types.AttributeKeySubmitter, msg.Submitter),
			sdk.NewAttribute(types.AttributeKeyBlockHeight, fmt.Sprintf("%d", blockHeight)),
		),
	})

	return &types.MsgSubmitBeaconResponse{
		Round: msg.Round,
	}, nil
}

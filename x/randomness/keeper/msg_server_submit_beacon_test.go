package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/randomness/keeper"
	"seocheon/x/randomness/types"
)

const (
	validRandomness = "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
	validSignature  = "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"
)

// validRound returns a round number that maps to exactly the block time (not in the future).
// With default params (genesis=1692803367, period=3) and block time=1692803667:
// round 100 → expected_time = 1692803367 + 100*3 = 1692803667 == block_time (valid).
const validRound = uint64(100)

func TestSubmitBeacon_Success(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)
	resp, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
		Round:      validRound,
		Randomness: validRandomness,
		Signature:  validSignature,
		Submitter:  sdk.AccAddress("submitter1____________").String(),
	})
	require.NoError(t, err)
	require.Equal(t, validRound, resp.Round)

	// Beacon should be stored.
	beacon, err := k.Beacons.Get(ctx, validRound)
	require.NoError(t, err)
	require.Equal(t, validRound, beacon.Round)
	require.Equal(t, validRandomness, beacon.Randomness)
	require.Equal(t, validSignature, beacon.Signature)
	require.Equal(t, int64(100), beacon.SubmittedAt)

	// LatestRound should be updated.
	latestRound, err := k.LatestRound.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, validRound, latestRound)
}

func TestSubmitBeacon_InvalidRound(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)
	_, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
		Round:      0,
		Randomness: validRandomness,
		Signature:  validSignature,
		Submitter:  sdk.AccAddress("submitter1____________").String(),
	})
	require.ErrorIs(t, err, types.ErrInvalidRound)
}

func TestSubmitBeacon_InvalidRandomnessFormat(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name       string
		randomness string
	}{
		{"empty", ""},
		{"too short", "a1b2c3d4"},
		{"too long", validRandomness + "ff"},
		{"invalid hex", "g1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
				Round:      validRound,
				Randomness: tc.randomness,
				Signature:  validSignature,
				Submitter:  sdk.AccAddress("submitter1____________").String(),
			})
			require.ErrorIs(t, err, types.ErrInvalidRandomness)
		})
	}
}

func TestSubmitBeacon_InvalidSignatureFormat(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name      string
		signature string
	}{
		{"empty", ""},
		{"invalid hex", "g1b2c3d4e5f6"},
		{"odd length hex", "a1b"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
				Round:      validRound,
				Randomness: validRandomness,
				Signature:  tc.signature,
				Submitter:  sdk.AccAddress("submitter1____________").String(),
			})
			require.ErrorIs(t, err, types.ErrInvalidSignature)
		})
	}
}

func TestSubmitBeacon_DuplicateBeacon(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)

	// First submission succeeds.
	_, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
		Round:      validRound,
		Randomness: validRandomness,
		Signature:  validSignature,
		Submitter:  sdk.AccAddress("submitter1____________").String(),
	})
	require.NoError(t, err)

	// Second submission for the same round fails.
	_, err = ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
		Round:      validRound,
		Randomness: validRandomness,
		Signature:  validSignature,
		Submitter:  sdk.AccAddress("submitter2____________").String(),
	})
	require.ErrorIs(t, err, types.ErrDuplicateBeacon)
}

func TestSubmitBeacon_FutureRound(t *testing.T) {
	k, ctx := testSetup(t)
	// Default params: DrandGenesisTime=1692803367, DrandPeriodSeconds=3.
	// Block time: 1692803667.
	// Round 101: expected_time = 1692803367 + 101*3 = 1692803670 > 1692803667 → future.
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)
	_, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
		Round:      101,
		Randomness: validRandomness,
		Signature:  validSignature,
		Submitter:  sdk.AccAddress("submitter1____________").String(),
	})
	require.ErrorIs(t, err, types.ErrInvalidRound)
}

func TestSubmitBeacon_VerificationEnabled(t *testing.T) {
	k, ctx := testSetup(t)
	params := types.DefaultParams()
	params.BeaconVerificationEnabled = true
	require.NoError(t, k.Params.Set(ctx, params))

	ms := keeper.NewMsgServerImpl(k)
	_, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
		Round:      validRound,
		Randomness: validRandomness,
		Signature:  validSignature,
		Submitter:  sdk.AccAddress("submitter1____________").String(),
	})
	require.ErrorIs(t, err, types.ErrVerificationFailed)
}

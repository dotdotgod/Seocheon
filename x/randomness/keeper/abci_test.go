package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/randomness/keeper"
	"seocheon/x/randomness/types"
)

// advanceForRound returns a context with block time advanced past the given drand round.
// expected_time = drand_genesis_time + (round * drand_period_seconds)
func advanceForRound(ctx sdk.Context, round uint64, height int64) sdk.Context {
	// drand_genesis_time = 1692803367, drand_period_seconds = 3 (quicknet defaults).
	expectedTime := int64(1692803367) + int64(round)*3
	return ctx.WithBlockHeight(height).WithBlockTime(time.Unix(expectedTime+1, 0))
}

func TestEndBlocker_Fulfill(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)

	// Submit a request at block 100.
	resp, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester1____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   2,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.NoError(t, err)

	// Advance time past target_round so SubmitBeacon is valid.
	futureCtx := advanceForRound(ctx, resp.TargetRound, 200)

	// Submit a beacon for the target round.
	_, err = ms.SubmitBeacon(futureCtx, &types.MsgSubmitBeacon{
		Submitter:  sdk.AccAddress("relayer_______________").String(),
		Round:      resp.TargetRound,
		Randomness: "c1c2c3c4c5c6c7c8c9c0c1c2c3c4c5c6c7c8c9c0c1c2c3c4c5c6c7c8c9c0c1c2",
		Signature:  "d1d2d3d4d5d6d7d8d9d0d1d2d3d4d5d6",
	})
	require.NoError(t, err)

	// Run EndBlocker on the future context.
	err = k.EndBlocker(futureCtx)
	require.NoError(t, err)

	// Verify request is fulfilled.
	req, err := k.Requests.Get(ctx, resp.RequestId)
	require.NoError(t, err)
	require.Equal(t, types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_FULFILLED, req.Status)
	require.NotEmpty(t, req.Result)
	require.Equal(t, int64(200), req.FulfilledAt)

	// Result should have 2 words (2 * 64 hex chars = 128 chars).
	require.Equal(t, 128, len(req.Result))

	// Verify pending count is 0.
	count, err := k.PendingRequestCount.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)

	// Verify PendingByRound is cleared.
	has, err := k.PendingByRound.Has(ctx, collections.Join(resp.TargetRound, resp.RequestId))
	require.NoError(t, err)
	require.False(t, has)

	// Verify FulfilledBlockIndex is set.
	has, err = k.FulfilledBlockIndex.Has(ctx, collections.Join(int64(200), resp.RequestId))
	require.NoError(t, err)
	require.True(t, has)
}

func TestEndBlocker_Expire(t *testing.T) {
	k, ctx := testSetup(t)

	params := types.DefaultParams()
	params.CommitRevealEnabled = true
	params.RequestTimeoutBlocks = 10 // Short timeout for testing.
	require.NoError(t, k.Params.Set(ctx, params))

	ms := keeper.NewMsgServerImpl(k)

	// Submit a request at block 100.
	resp, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester1____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.NoError(t, err)

	// Advance to block 111 (timeout = 10 blocks past creation at block 100).
	expireCtx := ctx.WithBlockHeight(111).WithBlockTime(time.Unix(1692803667+55, 0))

	// Run EndBlocker — should expire the request.
	err = k.EndBlocker(expireCtx)
	require.NoError(t, err)

	// Verify request is expired.
	req, err := k.Requests.Get(ctx, resp.RequestId)
	require.NoError(t, err)
	require.Equal(t, types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_EXPIRED, req.Status)

	// Verify pending count is 0.
	count, err := k.PendingRequestCount.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(0), count)
}

func TestEndBlocker_Prune(t *testing.T) {
	k, ctx := testSetup(t)

	params := types.DefaultParams()
	params.CommitRevealEnabled = true
	params.RequestPruningBlocks = 10 // Short pruning window for testing.
	require.NoError(t, k.Params.Set(ctx, params))

	ms := keeper.NewMsgServerImpl(k)

	// Submit a request at block 100.
	resp, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester1____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.NoError(t, err)

	// Advance time past target_round, submit beacon and fulfill.
	fulfillCtx := advanceForRound(ctx, resp.TargetRound, 105)

	_, err = ms.SubmitBeacon(fulfillCtx, &types.MsgSubmitBeacon{
		Submitter:  sdk.AccAddress("relayer_______________").String(),
		Round:      resp.TargetRound,
		Randomness: "c1c2c3c4c5c6c7c8c9c0c1c2c3c4c5c6c7c8c9c0c1c2c3c4c5c6c7c8c9c0c1c2",
		Signature:  "d1d2d3d4d5d6d7d8d9d0d1d2d3d4d5d6",
	})
	require.NoError(t, err)

	err = k.EndBlocker(fulfillCtx)
	require.NoError(t, err)

	// Verify request exists and is fulfilled.
	req, err := k.Requests.Get(ctx, resp.RequestId)
	require.NoError(t, err)
	require.Equal(t, types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_FULFILLED, req.Status)

	// Advance past pruning window (fulfilled at 105, pruning = 10 blocks, so 116+).
	pruneCtx := advanceForRound(ctx, resp.TargetRound+100, 116)

	err = k.EndBlocker(pruneCtx)
	require.NoError(t, err)

	// Verify request is pruned.
	has, err := k.Requests.Has(ctx, resp.RequestId)
	require.NoError(t, err)
	require.False(t, has)
}

func TestEndBlocker_MaxFulfillsPerBlock(t *testing.T) {
	k, ctx := testSetup(t)

	params := types.DefaultParams()
	params.CommitRevealEnabled = true
	params.MaxFulfillsPerBlock = 1 // Only 1 per block.
	require.NoError(t, k.Params.Set(ctx, params))

	ms := keeper.NewMsgServerImpl(k)

	// Submit 2 requests.
	var resps []*types.MsgRequestRandomnessResponse
	for i := 0; i < 2; i++ {
		resp, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
			Requester:  sdk.AccAddress("requester1____________").String(),
			CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			NumWords:   1,
			RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
		})
		require.NoError(t, err)
		resps = append(resps, resp)
	}

	// Both requests should have the same target round since same block time.
	targetRound := resps[0].TargetRound

	// Advance time and submit beacon.
	futureCtx := advanceForRound(ctx, targetRound, 200)

	_, err := ms.SubmitBeacon(futureCtx, &types.MsgSubmitBeacon{
		Submitter:  sdk.AccAddress("relayer_______________").String(),
		Round:      targetRound,
		Randomness: "c1c2c3c4c5c6c7c8c9c0c1c2c3c4c5c6c7c8c9c0c1c2c3c4c5c6c7c8c9c0c1c2",
		Signature:  "d1d2d3d4d5d6d7d8d9d0d1d2d3d4d5d6",
	})
	require.NoError(t, err)

	// Run EndBlocker — should fulfill only 1.
	err = k.EndBlocker(futureCtx)
	require.NoError(t, err)

	// Count fulfilled.
	fulfilled := 0
	for _, resp := range resps {
		req, err := k.Requests.Get(ctx, resp.RequestId)
		require.NoError(t, err)
		if req.Status == types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_FULFILLED {
			fulfilled++
		}
	}
	require.Equal(t, 1, fulfilled)

	// Run again to fulfill the second.
	err = k.EndBlocker(futureCtx)
	require.NoError(t, err)

	fulfilled = 0
	for _, resp := range resps {
		req, err := k.Requests.Get(ctx, resp.RequestId)
		require.NoError(t, err)
		if req.Status == types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_FULFILLED {
			fulfilled++
		}
	}
	require.Equal(t, 2, fulfilled)
}

func TestEndBlocker_DisabledSkips(t *testing.T) {
	k, ctx := testSetup(t)

	// Default params have commit_reveal_enabled = false.
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	// Should not error even though there's nothing to process.
	err := k.EndBlocker(ctx)
	require.NoError(t, err)
}

package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/randomness/keeper"
	"seocheon/x/randomness/types"
)

func TestQueryRandomnessRequest(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)
	qs := keeper.NewQueryServerImpl(k)

	// Submit a request.
	resp, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester1____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.NoError(t, err)

	// Query the request.
	qResp, err := qs.RandomnessRequest(ctx, &types.QueryRandomnessRequestRequest{
		RequestId: resp.RequestId,
	})
	require.NoError(t, err)
	require.Equal(t, resp.RequestId, qResp.Request.RequestId)
	require.Equal(t, types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING, qResp.Request.Status)
}

func TestQueryRandomnessRequest_NotFound(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	qs := keeper.NewQueryServerImpl(k)

	_, err := qs.RandomnessRequest(ctx, &types.QueryRandomnessRequestRequest{
		RequestId: 999,
	})
	require.ErrorIs(t, err, types.ErrRequestNotFound)
}

func TestQueryPendingRequests(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)
	qs := keeper.NewQueryServerImpl(k)

	// Submit 3 requests.
	for i := 0; i < 3; i++ {
		_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
			Requester:  sdk.AccAddress("requester1____________").String(),
			CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			NumWords:   1,
			RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
		})
		require.NoError(t, err)
	}

	// Query pending requests.
	qResp, err := qs.PendingRequests(ctx, &types.QueryPendingRequestsRequest{})
	require.NoError(t, err)
	require.Equal(t, 3, len(qResp.Requests))
	require.Equal(t, uint64(3), qResp.Pagination.Total)
}

func TestQueryRequestsByRequester(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)
	qs := keeper.NewQueryServerImpl(k)

	requester1 := sdk.AccAddress("requester1____________").String()
	requester2 := sdk.AccAddress("requester2____________").String()

	// Submit 2 requests from requester1.
	for i := 0; i < 2; i++ {
		_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
			Requester:  requester1,
			CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			NumWords:   1,
			RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
		})
		require.NoError(t, err)
	}

	// Submit 1 request from requester2.
	_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  requester2,
		CommitHash: "b1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.NoError(t, err)

	// Query requester1's requests.
	qResp1, err := qs.RequestsByRequester(ctx, &types.QueryRequestsByRequesterRequest{
		Requester: requester1,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(qResp1.Requests))

	// Query requester2's requests.
	qResp2, err := qs.RequestsByRequester(ctx, &types.QueryRequestsByRequesterRequest{
		Requester: requester2,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(qResp2.Requests))
}

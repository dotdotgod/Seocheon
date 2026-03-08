package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"seocheon/x/randomness/keeper"
	"seocheon/x/randomness/types"
)

// storeBeacon is a test helper that submits a beacon via MsgServer.
func storeBeacon(t *testing.T, ms types.MsgServer, ctx sdk.Context, round uint64) {
	t.Helper()
	_, err := ms.SubmitBeacon(ctx, &types.MsgSubmitBeacon{
		Round:      round,
		Randomness: validRandomness,
		Signature:  validSignature,
		Submitter:  sdk.AccAddress("submitter1____________").String(),
	})
	require.NoError(t, err)
}

func TestQueryBeacon_Success(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)
	qs := keeper.NewQueryServerImpl(k)

	storeBeacon(t, ms, ctx, validRound)

	resp, err := qs.Beacon(ctx, &types.QueryBeaconRequest{Round: validRound})
	require.NoError(t, err)
	require.Equal(t, validRound, resp.Beacon.Round)
	require.Equal(t, validRandomness, resp.Beacon.Randomness)
}

func TestQueryBeacon_NotFound(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	qs := keeper.NewQueryServerImpl(k)

	// Query nonexistent round.
	_, err := qs.Beacon(ctx, &types.QueryBeaconRequest{Round: 999})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestQueryBeacon_ZeroRound(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	qs := keeper.NewQueryServerImpl(k)

	// Round 0 is rejected as invalid argument.
	_, err := qs.Beacon(ctx, &types.QueryBeaconRequest{Round: 0})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestQueryBeacons_List(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)
	qs := keeper.NewQueryServerImpl(k)

	// Store beacons for rounds 98, 99, 100 (all <= block time).
	for _, round := range []uint64{98, 99, validRound} {
		storeBeacon(t, ms, ctx, round)
	}

	resp, err := qs.Beacons(ctx, &types.QueryBeaconsRequest{})
	require.NoError(t, err)
	require.Equal(t, 3, len(resp.Beacons))
	require.Equal(t, uint64(3), resp.Pagination.Total)
}

func TestQueryLatestBeacon_Success(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)
	qs := keeper.NewQueryServerImpl(k)

	// Store two beacons; round 100 is the latest.
	storeBeacon(t, ms, ctx, 99)
	storeBeacon(t, ms, ctx, validRound)

	resp, err := qs.LatestBeacon(ctx, &types.QueryLatestBeaconRequest{})
	require.NoError(t, err)
	require.Equal(t, validRound, resp.Beacon.Round)
}

func TestQueryLatestBeacon_NoBeacon(t *testing.T) {
	k, ctx := testSetup(t)
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	qs := keeper.NewQueryServerImpl(k)

	// No beacons stored yet.
	_, err := qs.LatestBeacon(ctx, &types.QueryLatestBeaconRequest{})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

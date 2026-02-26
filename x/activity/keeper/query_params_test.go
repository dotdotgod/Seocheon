package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"seocheon/x/activity/keeper"
	"seocheon/x/activity/types"
)

func TestQueryParams_Default(t *testing.T) {
	f := initFixture(t)
	ctx := f.freshCtx(1)

	qs := keeper.NewQueryServerImpl(f.keeper)

	resp, err := qs.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, int64(17280), resp.Params.EpochLength)
	require.Equal(t, int64(12), resp.Params.WindowsPerEpoch)
	require.Equal(t, int64(8), resp.Params.MinActiveWindows)
	require.Equal(t, uint64(100), resp.Params.SelfFundedQuota)
	require.Equal(t, uint64(10), resp.Params.FeegrantQuota)
	require.Equal(t, int64(6307200), resp.Params.ActivityPruningKeepBlocks)
}

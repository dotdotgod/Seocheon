package keeper_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/randomness/keeper"
	"seocheon/x/randomness/types"

	"github.com/cosmos/cosmos-sdk/testutil"
)

// mockBankKeeper implements types.BankKeeper for testing.
type mockBankKeeper struct {
	balances map[string]sdk.Coins
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		balances: make(map[string]sdk.Coins),
	}
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	return nil
}

// testSetup creates a keeper and context for testing.
func testSetup(t *testing.T) (keeper.Keeper, sdk.Context) {
	t.Helper()

	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	ir := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)

	// Use a simple address codec for testing.
	ac := newTestAddressCodec()

	authority := sdk.AccAddress("gov_authority_________")

	k := keeper.NewKeeper(storeService.(corestore.KVStoreService), cdc, ac, authority)
	bk := newMockBankKeeper()
	k.SetBankKeeper(bk)

	sdkCtx := ctx.Ctx.WithBlockHeight(100).WithBlockTime(time.Unix(1692803667, 0)) // 100 rounds after quicknet genesis

	return k, sdkCtx
}

// testAddressCodec is a simple address codec for tests.
type testAddressCodec struct{}

func newTestAddressCodec() address.Codec {
	return testAddressCodec{}
}

func (testAddressCodec) StringToBytes(text string) ([]byte, error) {
	return []byte(text), nil
}

func (testAddressCodec) BytesToString(bz []byte) (string, error) {
	return string(bz), nil
}

func setupCommitRevealParams(t *testing.T, k keeper.Keeper, ctx sdk.Context) {
	t.Helper()
	params := types.DefaultParams()
	params.CommitRevealEnabled = true
	require.NoError(t, k.Params.Set(ctx, params))
}

func TestRequestRandomness_Disabled(t *testing.T) {
	k, ctx := testSetup(t)

	// Set params with commit_reveal_enabled = false (default).
	require.NoError(t, k.Params.Set(ctx, types.DefaultParams()))

	ms := keeper.NewMsgServerImpl(k)
	_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester1____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.ErrorIs(t, err, types.ErrCommitRevealDisabled)
}

func TestRequestRandomness_Success(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)
	resp, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester1____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   3,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.RequestId)
	require.Greater(t, resp.TargetRound, uint64(0))

	// Verify stored request.
	req, err := k.Requests.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING, req.Status)
	require.Equal(t, uint32(3), req.NumWords)
	require.Equal(t, int64(100), req.CreatedAt)

	// Verify pending count.
	count, err := k.PendingRequestCount.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(1), count)

	// Verify PendingByRound index.
	has, err := k.PendingByRound.Has(ctx, collections.Join(resp.TargetRound, uint64(1)))
	require.NoError(t, err)
	require.True(t, has)
}

func TestRequestRandomness_InvalidCommitHash(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name       string
		commitHash string
	}{
		{"too short", "a1b2c3d4"},
		{"too long", "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2ff"},
		{"invalid hex", "g1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2"},
		{"empty", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
				Requester:  sdk.AccAddress("requester1____________").String(),
				CommitHash: tc.commitHash,
				NumWords:   1,
				RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
			})
			require.ErrorIs(t, err, types.ErrInvalidCommitHash)
		})
	}
}

func TestRequestRandomness_InvalidNumWords(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)

	tests := []struct {
		name     string
		numWords uint32
	}{
		{"zero", 0},
		{"too many", 11},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
				Requester:  sdk.AccAddress("requester1____________").String(),
				CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
				NumWords:   tc.numWords,
				RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
			})
			require.ErrorIs(t, err, types.ErrInvalidNumWords)
		})
	}
}

func TestRequestRandomness_CallbackDataTooLong(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)

	// 257 bytes = 514 hex chars.
	longData := ""
	for i := 0; i < 514; i++ {
		longData += "a"
	}

	_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:    sdk.AccAddress("requester1____________").String(),
		CommitHash:   "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:     1,
		CallbackData: longData,
		RequestFee:   sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.ErrorIs(t, err, types.ErrCallbackDataTooLong)
}

func TestRequestRandomness_InvalidCallbackData(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)

	_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:    sdk.AccAddress("requester1____________").String(),
		CommitHash:   "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:     1,
		CallbackData: "not_valid_hex!!!",
		RequestFee:   sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.ErrorIs(t, err, types.ErrInvalidCallbackData)
}

func TestRequestRandomness_MaxPendingRequests(t *testing.T) {
	k, ctx := testSetup(t)

	// Set a low max_pending_requests.
	params := types.DefaultParams()
	params.CommitRevealEnabled = true
	params.MaxPendingRequests = 2
	require.NoError(t, k.Params.Set(ctx, params))

	ms := keeper.NewMsgServerImpl(k)

	// Submit 2 requests.
	for i := 0; i < 2; i++ {
		_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
			Requester:  sdk.AccAddress("requester1____________").String(),
			CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			NumWords:   1,
			RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
		})
		require.NoError(t, err)
	}

	// Third request should fail.
	_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester2____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.ErrorIs(t, err, types.ErrTooManyPendingRequests)
}

func TestRequestRandomness_MaxPerRequester(t *testing.T) {
	k, ctx := testSetup(t)

	params := types.DefaultParams()
	params.CommitRevealEnabled = true
	params.MaxRequestsPerRequester = 1
	require.NoError(t, k.Params.Set(ctx, params))

	ms := keeper.NewMsgServerImpl(k)

	requester := sdk.AccAddress("requester1____________").String()

	// First request succeeds.
	_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  requester,
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.NoError(t, err)

	// Second request from same requester should fail.
	_, err = ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  requester,
		CommitHash: "b1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(1000)),
	})
	require.ErrorIs(t, err, types.ErrRequesterLimitExceeded)
}

func TestRequestRandomness_InsufficientFee(t *testing.T) {
	k, ctx := testSetup(t)
	setupCommitRevealParams(t, k, ctx)

	ms := keeper.NewMsgServerImpl(k)

	_, err := ms.RequestRandomness(ctx, &types.MsgRequestRandomness{
		Requester:  sdk.AccAddress("requester1____________").String(),
		CommitHash: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		NumWords:   1,
		RequestFee: sdk.NewCoin("uppyeo", math.NewInt(500)), // Less than 1000.
	})
	require.ErrorIs(t, err, types.ErrInsufficientRequestFee)
}

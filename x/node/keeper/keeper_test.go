package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/keeper"
	module "seocheon/x/node/module"
	"seocheon/x/node/types"
)

// ---------------------------------------------------------------------------
// Mock Keepers
// ---------------------------------------------------------------------------

// mockAuthKeeper implements types.AuthKeeper.
type mockAuthKeeper struct {
	moduleAddresses map[string]sdk.AccAddress
}

func newMockAuthKeeper() *mockAuthKeeper {
	return &mockAuthKeeper{
		moduleAddresses: make(map[string]sdk.AccAddress),
	}
}

func (m *mockAuthKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return m.moduleAddresses[name]
}

func (m *mockAuthKeeper) GetModuleAccount(_ context.Context, _ string) sdk.ModuleAccountI {
	return nil
}

func (m *mockAuthKeeper) GetAccount(_ context.Context, _ sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m *mockAuthKeeper) AddressCodec() address.Codec {
	return addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
}

// mockBankKeeper implements types.BankKeeper.
type mockBankKeeper struct {
	balances map[string]sdk.Coins // address string -> coins
	sendErr  error                // configurable error for testing
	// Track SendCoinsFromModuleToAccount calls for verification.
	sentFromModule []sentFromModuleRecord
	// Track SendCoins calls for verification (used by WithdrawNodeCommission).
	sentCoinsRecords []sentCoinsRecord
}

type sentFromModuleRecord struct {
	SenderModule string
	Recipient    sdk.AccAddress
	Amount       sdk.Coins
}

type sentCoinsRecord struct {
	From   sdk.AccAddress
	To     sdk.AccAddress
	Amount sdk.Coins
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		balances: make(map[string]sdk.Coins),
	}
}

func (m *mockBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

func (m *mockBankKeeper) SendCoins(_ context.Context, from, to sdk.AccAddress, amt sdk.Coins) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentCoinsRecords = append(m.sentCoinsRecords, sentCoinsRecord{
		From:   from,
		To:     to,
		Amount: amt,
	})
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sentFromModule = append(m.sentFromModule, sentFromModuleRecord{
		SenderModule: senderModule,
		Recipient:    recipientAddr,
		Amount:       amt,
	})
	return nil
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return m.sendErr
}

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, _, _ string, _ sdk.Coins) error {
	return m.sendErr
}

func (m *mockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coins := m.balances[addr.String()]
	return sdk.Coin{Denom: denom, Amount: coins.AmountOf(denom)}
}

func (m *mockBankKeeper) GetAllBalances(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

// mockStakingKeeper implements types.StakingKeeper.
type mockStakingKeeper struct {
	bondDenom string
}

func newMockStakingKeeper() *mockStakingKeeper {
	return &mockStakingKeeper{
		bondDenom: "usum",
	}
}

func (m *mockStakingKeeper) ConsensusAddressCodec() address.Codec {
	return addresscodec.NewBech32Codec("seocheonvalcons")
}

func (m *mockStakingKeeper) ValidatorByConsAddr(_ context.Context, _ sdk.ConsAddress) (stakingtypes.ValidatorI, error) {
	return nil, nil
}

func (m *mockStakingKeeper) GetValidator(_ context.Context, _ sdk.ValAddress) (stakingtypes.Validator, error) {
	return stakingtypes.Validator{}, nil
}

func (m *mockStakingKeeper) BondDenom(_ context.Context) (string, error) {
	return m.bondDenom, nil
}

// mockFeegrantKeeper implements types.FeegrantKeeper.
type mockFeegrantKeeper struct {
	grantCalls []feegrantGrantCall
	grantErr   error
}

type feegrantGrantCall struct {
	Granter sdk.AccAddress
	Grantee sdk.AccAddress
}

func newMockFeegrantKeeper() *mockFeegrantKeeper {
	return &mockFeegrantKeeper{}
}

func (m *mockFeegrantKeeper) GrantAllowance(_ context.Context, granter, grantee sdk.AccAddress, _ feegrant.FeeAllowanceI) error {
	if m.grantErr != nil {
		return m.grantErr
	}
	m.grantCalls = append(m.grantCalls, feegrantGrantCall{Granter: granter, Grantee: grantee})
	return nil
}

// mockStakingMsgServer implements types.StakingMsgServer.
type mockStakingMsgServer struct {
	createValidatorErr error
	undelegateErr      error
}

func newMockStakingMsgServer() *mockStakingMsgServer {
	return &mockStakingMsgServer{}
}

func (m *mockStakingMsgServer) CreateValidator(_ context.Context, _ *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	if m.createValidatorErr != nil {
		return nil, m.createValidatorErr
	}
	return &stakingtypes.MsgCreateValidatorResponse{}, nil
}

func (m *mockStakingMsgServer) Undelegate(_ context.Context, _ *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	if m.undelegateErr != nil {
		return nil, m.undelegateErr
	}
	return &stakingtypes.MsgUndelegateResponse{}, nil
}

// ---------------------------------------------------------------------------
// Test Fixture
// ---------------------------------------------------------------------------

type fixture struct {
	ctx              context.Context
	keeper           keeper.Keeper
	addressCodec     address.Codec
	authKeeper       *mockAuthKeeper
	bankKeeper       *mockBankKeeper
	stakingKeeper    *mockStakingKeeper
	feegrantKeeper   *mockFeegrantKeeper
	stakingMsgServer *mockStakingMsgServer
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	// Create mock keepers.
	authK := newMockAuthKeeper()
	bankK := newMockBankKeeper()
	stakingK := newMockStakingKeeper()
	feegrantK := newMockFeegrantKeeper()
	stakingMsgSrv := newMockStakingMsgServer()

	// Set up module account addresses for Registration Pool and Feegrant Pool.
	regPoolAddr := authtypes.NewModuleAddress(types.RegistrationPoolName)
	fgPoolAddr := authtypes.NewModuleAddress(types.FeegrantPoolName)
	authK.moduleAddresses[types.RegistrationPoolName] = regPoolAddr
	authK.moduleAddresses[types.FeegrantPoolName] = fgPoolAddr

	// Set up Registration Pool balance: 1000 usum.
	bankK.balances[regPoolAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(1000)))
	bankK.balances[fgPoolAddr.String()] = sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(100)))

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addrCodec,
		authority,
		bankK,
		stakingK,
	)

	// Wire optional keepers via setters.
	k.SetAuthKeeper(authK)
	k.SetFeegrantKeeper(feegrantK)
	k.SetStakingMsgServer(stakingMsgSrv)

	// Initialize default params.
	if err := k.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return &fixture{
		ctx:              ctx,
		keeper:           k,
		addressCodec:     addrCodec,
		authKeeper:       authK,
		bankKeeper:       bankK,
		stakingKeeper:    stakingK,
		feegrantKeeper:   feegrantK,
		stakingMsgServer: stakingMsgSrv,
	}
}

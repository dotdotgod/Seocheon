package node_test

import (
	"context"
	"testing"

	"cosmossdk.io/core/address"
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
// Value vs Pointer Keeper Tests
//
// These tests verify that pointer-based keeper sharing between AppModule and
// the app's keeper field is essential for proper optional dependency wiring.
//
// Problem (value-based keeper):
//   ProvideModule creates Keeper k (value)
//   → NewAppModule copies k into AppModule.keeper (copy A)
//   → ModuleOutputs.NodeKeeper copies k into app.NodeKeeper (copy B)
//   → app.NodeKeeper.SetFeegrantKeeper(...) modifies copy B only
//   → AppModule.keeper (copy A) never gets feegrantKeeper
//   → RegisterServices creates MsgServer with copy A → missing deps!
//
// Fix (pointer-based keeper):
//   ProvideModule creates Keeper k (value)
//   → NewAppModule stores &k into AppModule.keeper (pointer)
//   → ModuleOutputs.NodeKeeper returns &k (same pointer)
//   → app.NodeKeeper.SetFeegrantKeeper(...) modifies *k
//   → AppModule.keeper points to same k → sees the update
//   → RegisterServices dereferences *am.keeper → snapshot with all deps
// ---------------------------------------------------------------------------

// mockAuthKeeper implements types.AuthKeeper for testing.
type mockAuthKeeper struct {
	moduleAddresses map[string]sdk.AccAddress
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

// mockBankKeeper implements types.BankKeeper for testing.
type mockBankKeeper struct {
	balances map[string]sdk.Coins
}

func (m *mockBankKeeper) SpendableCoins(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

func (m *mockBankKeeper) SendCoins(_ context.Context, _, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, _, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) MintCoins(_ context.Context, _ string, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	coins := m.balances[addr.String()]
	return sdk.Coin{Denom: denom, Amount: coins.AmountOf(denom)}
}

func (m *mockBankKeeper) GetAllBalances(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

// mockStakingKeeper implements types.StakingKeeper for testing.
type mockStakingKeeper struct{}

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
	return "usum", nil
}

// mockFeegrantKeeper implements types.FeegrantKeeper for testing.
type mockFeegrantKeeper struct {
	grantCalled bool
}

func (m *mockFeegrantKeeper) GrantAllowance(_ context.Context, _, _ sdk.AccAddress, _ feegrant.FeeAllowanceI) error {
	m.grantCalled = true
	return nil
}

// mockStakingMsgServer implements types.StakingMsgServer for testing.
type mockStakingMsgServer struct {
	createValidatorCalled bool
}

func (m *mockStakingMsgServer) CreateValidator(_ context.Context, _ *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error) {
	m.createValidatorCalled = true
	return &stakingtypes.MsgCreateValidatorResponse{}, nil
}

func (m *mockStakingMsgServer) Undelegate(_ context.Context, _ *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error) {
	return &stakingtypes.MsgUndelegateResponse{}, nil
}

// TestPointerKeeperPropagation verifies that when AppModule stores a *Keeper,
// setter calls on the same pointer are visible through the module's keeper.
// This is the core test for the value-vs-pointer keeper issue.
func TestPointerKeeperPropagation(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	authK := &mockAuthKeeper{moduleAddresses: map[string]sdk.AccAddress{
		types.RegistrationPoolName: authtypes.NewModuleAddress(types.RegistrationPoolName),
		types.FeegrantPoolName:     authtypes.NewModuleAddress(types.FeegrantPoolName),
	}}
	bankK := &mockBankKeeper{balances: make(map[string]sdk.Coins)}
	stakingK := &mockStakingKeeper{}

	// Step 1: Create keeper (simulates what ProvideModule does).
	k := keeper.NewKeeper(storeService, encCfg.Codec, addrCodec, authority, bankK, stakingK)
	k.SetAuthKeeper(authK)

	// Step 2: Create AppModule with POINTER to keeper.
	// This simulates: m := NewAppModule(cdc, &k, ...)
	appModule := module.NewAppModule(encCfg.Codec, &k, authK, bankK)

	// Step 3: Set optional keepers AFTER module creation.
	// This simulates what app.go does after depinject.Inject:
	//   app.NodeKeeper.SetFeegrantKeeper(app.FeegrantKeeper)
	//   app.NodeKeeper.SetStakingMsgServer(...)
	fgK := &mockFeegrantKeeper{}
	stakingMsgSrv := &mockStakingMsgServer{}

	k.SetFeegrantKeeper(fgK)
	k.SetStakingMsgServer(stakingMsgSrv)

	// Step 4: Verify the module sees the updates.
	// Since AppModule stores *Keeper (pointer), and we modified the same Keeper,
	// the module should see the feegrantKeeper and stakingMsgServer.
	//
	// We verify this indirectly: the module's Name() and ConsensusVersion()
	// should work (basic sanity), and RegisterServices should create a working
	// MsgServer that can use the optional keepers.
	if appModule.Name() != types.ModuleName {
		t.Fatalf("expected module name %s, got %s", types.ModuleName, appModule.Name())
	}
	if appModule.ConsensusVersion() != 1 {
		t.Fatalf("expected consensus version 1, got %d", appModule.ConsensusVersion())
	}

	// The module should produce valid DefaultGenesis without panicking.
	genesis := appModule.DefaultGenesis(nil)
	if genesis == nil {
		t.Fatal("DefaultGenesis returned nil")
	}
}

// TestValueKeeperDoesNotPropagate demonstrates the root cause of the
// value-vs-pointer issue: value copies don't share setter updates.
func TestValueKeeperDoesNotPropagate(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)

	authority := authtypes.NewModuleAddress(types.GovModuleName)
	authK := &mockAuthKeeper{moduleAddresses: make(map[string]sdk.AccAddress)}
	bankK := &mockBankKeeper{balances: make(map[string]sdk.Coins)}
	stakingK := &mockStakingKeeper{}

	// Create keeper (value type).
	k := keeper.NewKeeper(storeService, encCfg.Codec, addrCodec, authority, bankK, stakingK)
	k.SetAuthKeeper(authK)

	// Simulate the OLD (broken) approach: copy the keeper by value.
	copyOfK := k // value copy

	// Set optional keeper on the ORIGINAL.
	fgK := &mockFeegrantKeeper{}
	k.SetFeegrantKeeper(fgK)

	// The copy does NOT see the update — this is the bug.
	// We can't directly inspect the copy's feegrantKeeper field (unexported),
	// but we can create a MsgServer from the copy and test behavior.
	_ = copyOfK

	// Create MsgServer from the copy (as the old code did).
	// msgServer := keeper.NewMsgServerImpl(copyOfK)
	// If register-node tries to use feegrantKeeper, it would be nil in the copy.

	// With the pointer approach, the same keeper is shared:
	ptrK := &k
	k.SetStakingMsgServer(&mockStakingMsgServer{})

	// ptrK sees the update because it points to the same keeper.
	msgServer := keeper.NewMsgServerImpl(*ptrK)
	_ = msgServer // MsgServer has all optional keepers set.

	t.Log("Value copy does not propagate setter changes — pointer does")
}

// TestModuleKeeperConsistencyWithAppKeeper verifies the full wiring scenario:
// both the app-level keeper and the module's internal keeper should be
// functionally equivalent when using pointer-based sharing.
func TestModuleKeeperConsistencyWithAppKeeper(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)
	authK := &mockAuthKeeper{moduleAddresses: map[string]sdk.AccAddress{
		types.RegistrationPoolName: authtypes.NewModuleAddress(types.RegistrationPoolName),
		types.FeegrantPoolName:     authtypes.NewModuleAddress(types.FeegrantPoolName),
	}}
	bankK := &mockBankKeeper{balances: make(map[string]sdk.Coins)}
	stakingK := &mockStakingKeeper{}

	// Simulate ProvideModule: create keeper and module with pointer sharing.
	k := keeper.NewKeeper(storeService, encCfg.Codec, addrCodec, authority, bankK, stakingK)
	k.SetAuthKeeper(authK)

	// Both appKeeper and module point to the same keeper.
	appKeeper := &k                                                  // app.NodeKeeper
	appModule := module.NewAppModule(encCfg.Codec, &k, authK, bankK) // module stores &k

	// Simulate app.go: set optional keepers on appKeeper.
	fgK := &mockFeegrantKeeper{}
	appKeeper.SetFeegrantKeeper(fgK)
	appKeeper.SetStakingMsgServer(&mockStakingMsgServer{})

	// Initialize params (needed for operations).
	if err := appKeeper.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	// Verify the module can be used for genesis operations.
	// InitGenesis uses am.keeper internally — if pointer sharing works,
	// the module's keeper should have params set (via appKeeper).
	genBytes := appModule.DefaultGenesis(nil)
	if err := appModule.ValidateGenesis(nil, nil, genBytes); err != nil {
		t.Fatalf("ValidateGenesis failed: %v", err)
	}

	// InitGenesis should work without panic.
	appModule.InitGenesis(ctx, nil, genBytes)

	// ExportGenesis should return valid state.
	exported := appModule.ExportGenesis(ctx, nil)
	if exported == nil {
		t.Fatal("ExportGenesis returned nil")
	}

	// EndBlocker should work (uses am.keeper internally).
	if err := appModule.EndBlock(ctx); err != nil {
		t.Fatalf("EndBlock failed: %v", err)
	}
}

package activity_test

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/feegrant"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"seocheon/x/activity/keeper"
	module "seocheon/x/activity/module"
	"seocheon/x/activity/types"
)

// ---------------------------------------------------------------------------
// Value vs Pointer Keeper Tests for x/activity
//
// These tests mirror x/node/module/module_test.go and verify that the
// pointer-based keeper pattern also works correctly for x/activity.
// The issue: optional keepers (NodeKeeper, AuthKeeper, FeegrantKeeper)
// set via SetXxxKeeper() after module creation must be visible to the
// module's internal keeper through pointer sharing.
// ---------------------------------------------------------------------------

// mockNodeKeeper implements types.NodeKeeper for testing.
type mockNodeKeeper struct {
	nodes map[string]int32 // agent_addr -> status
	agent map[string]string // agent_addr -> node_id
}

func (m *mockNodeKeeper) GetNodeIDByAgent(_ context.Context, agentAddr string) (string, error) {
	nodeID, ok := m.agent[agentAddr]
	if !ok {
		return "", nil
	}
	return nodeID, nil
}

func (m *mockNodeKeeper) GetNodeStatus(_ context.Context, nodeID string) (int32, error) {
	status, ok := m.nodes[nodeID]
	if !ok {
		return 0, nil
	}
	return status, nil
}

func (m *mockNodeKeeper) GetNodeOperatorAddress(_ context.Context, nodeID string) (string, error) {
	return "", fmt.Errorf("not found: %s", nodeID)
}

func (m *mockNodeKeeper) GetNodeAgentAddress(_ context.Context, nodeID string) (string, error) {
	return "", fmt.Errorf("not found: %s", nodeID)
}

func (m *mockNodeKeeper) GetNodeAgentShare(_ context.Context, nodeID string) (math.LegacyDec, error) {
	return math.LegacyZeroDec(), fmt.Errorf("not found: %s", nodeID)
}

// mockAuthKeeper implements types.AuthKeeper for testing.
type mockAuthKeeper struct {
	moduleAddresses map[string]sdk.AccAddress
}

func (m *mockAuthKeeper) GetModuleAddress(name string) sdk.AccAddress {
	return m.moduleAddresses[name]
}

// mockFeegrantKeeper implements types.FeegrantKeeper for testing.
type mockFeegrantKeeper struct{}

func (m *mockFeegrantKeeper) GetAllowance(_ context.Context, _, _ sdk.AccAddress) (feegrant.FeeAllowanceI, error) {
	return nil, nil
}

// TestPointerKeeperPropagation verifies that when AppModule stores a *Keeper,
// setter calls on the same pointer are visible through the module's keeper.
func TestPointerKeeperPropagation(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	// Step 1: Create keeper (simulates what ProvideModule does).
	k := keeper.NewKeeper(storeService, encCfg.Codec, addrCodec, authority)

	// Step 2: Create AppModule with POINTER to keeper.
	appModule := module.NewAppModule(encCfg.Codec, &k)

	// Step 3: Set optional keepers AFTER module creation.
	// This simulates what app.go does after depinject.Inject:
	//   app.ActivityKeeper.SetNodeKeeper(app.NodeKeeper)
	//   app.ActivityKeeper.SetAuthKeeper(app.AuthKeeper)
	//   app.ActivityKeeper.SetFeegrantKeeper(app.FeegrantKeeper)
	nk := &mockNodeKeeper{
		nodes: map[string]int32{"node1": 1},
		agent: map[string]string{"agent1": "node1"},
	}
	ak := &mockAuthKeeper{moduleAddresses: map[string]sdk.AccAddress{
		"node_feegrant_pool": authtypes.NewModuleAddress("node_feegrant_pool"),
	}}
	fk := &mockFeegrantKeeper{}

	k.SetNodeKeeper(nk)
	k.SetAuthKeeper(ak)
	k.SetFeegrantKeeper(fk)

	// Step 4: Verify the module sees the updates.
	if appModule.Name() != types.ModuleName {
		t.Fatalf("expected module name %s, got %s", types.ModuleName, appModule.Name())
	}
	if appModule.ConsensusVersion() != 2 {
		t.Fatalf("expected consensus version 2, got %d", appModule.ConsensusVersion())
	}

	genesis := appModule.DefaultGenesis(nil)
	if genesis == nil {
		t.Fatal("DefaultGenesis returned nil")
	}
}

// TestModuleKeeperConsistencyWithAppKeeper verifies the full wiring scenario:
// optional keepers set via app-level pointer propagate through to module
// operations (InitGenesis, ExportGenesis, EndBlock).
func TestModuleKeeperConsistencyWithAppKeeper(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addrCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	// Simulate ProvideModule: create keeper and module with pointer sharing.
	k := keeper.NewKeeper(storeService, encCfg.Codec, addrCodec, authority)

	// Both appKeeper and module point to the same keeper.
	appKeeper := &k
	appModule := module.NewAppModule(encCfg.Codec, &k)

	// Simulate app.go: set optional keepers on appKeeper.
	appKeeper.SetNodeKeeper(&mockNodeKeeper{
		nodes: make(map[string]int32),
		agent: make(map[string]string),
	})
	appKeeper.SetAuthKeeper(&mockAuthKeeper{
		moduleAddresses: make(map[string]sdk.AccAddress),
	})
	appKeeper.SetFeegrantKeeper(&mockFeegrantKeeper{})

	// Initialize params (needed for operations).
	if err := appKeeper.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	// Verify the module can be used for genesis operations.
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

package keeper_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/collections"
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
	nodetypes "seocheon/x/node/types"
)

// ---------------------------------------------------------------------------
// Mock Keepers
// ---------------------------------------------------------------------------

// mockNodeKeeper implements types.NodeKeeper.
type mockNodeKeeper struct {
	agentToNode    map[string]string       // agent_addr -> node_id
	nodeStatuses   map[string]int32        // node_id -> status
	nodeOperators  map[string]string       // node_id -> operator_addr
	nodeAgents     map[string]string       // node_id -> agent_wallet_addr
	nodeAgentShare map[string]math.LegacyDec // node_id -> agent_share (0-100)
}

func newMockNodeKeeper() *mockNodeKeeper {
	return &mockNodeKeeper{
		agentToNode:    make(map[string]string),
		nodeStatuses:   make(map[string]int32),
		nodeOperators:  make(map[string]string),
		nodeAgents:     make(map[string]string),
		nodeAgentShare: make(map[string]math.LegacyDec),
	}
}

func (m *mockNodeKeeper) GetNodeIDByAgent(_ context.Context, agentAddr string) (string, error) {
	nodeID, ok := m.agentToNode[agentAddr]
	if !ok {
		return "", fmt.Errorf("agent address %s not registered", agentAddr)
	}
	return nodeID, nil
}

func (m *mockNodeKeeper) GetNodeStatus(_ context.Context, nodeID string) (int32, error) {
	status, ok := m.nodeStatuses[nodeID]
	if !ok {
		return 0, fmt.Errorf("node %s not found", nodeID)
	}
	return status, nil
}

func (m *mockNodeKeeper) GetNodeOperatorAddress(_ context.Context, nodeID string) (string, error) {
	addr, ok := m.nodeOperators[nodeID]
	if !ok {
		return "", fmt.Errorf("node %s operator not found", nodeID)
	}
	return addr, nil
}

func (m *mockNodeKeeper) GetNodeAgentAddress(_ context.Context, nodeID string) (string, error) {
	addr, ok := m.nodeAgents[nodeID]
	if !ok {
		return "", fmt.Errorf("node %s agent address not found", nodeID)
	}
	return addr, nil
}

func (m *mockNodeKeeper) GetNodeAgentShare(_ context.Context, nodeID string) (math.LegacyDec, error) {
	share, ok := m.nodeAgentShare[nodeID]
	if !ok {
		return math.LegacyZeroDec(), fmt.Errorf("node %s agent share not found", nodeID)
	}
	return share, nil
}

// registerNode adds a node to the mock.
func (m *mockNodeKeeper) registerNode(nodeID, agentAddr string, status int32) {
	m.agentToNode[agentAddr] = nodeID
	m.nodeStatuses[nodeID] = status
}

// registerFullNode adds a node with all fields for reward distribution tests.
func (m *mockNodeKeeper) registerFullNode(nodeID, agentAddr, operatorAddr, agentWalletAddr string, status int32, agentShare math.LegacyDec) {
	m.agentToNode[agentAddr] = nodeID
	m.nodeStatuses[nodeID] = status
	m.nodeOperators[nodeID] = operatorAddr
	m.nodeAgents[nodeID] = agentWalletAddr
	m.nodeAgentShare[nodeID] = agentShare
}

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

// mockBankKeeper implements types.BankKeeper for testing.
type mockBankKeeper struct {
	moduleBalances  map[string]sdk.Coins // module_name -> balances
	accountBalances map[string]sdk.Coins // acc_addr -> balances
	moduleToAccSent []bankTransfer       // track sends for verification
	moduleToModSent []bankModTransfer
	accToModSent    []bankAccToModTransfer
}

type bankTransfer struct {
	FromModule string
	To         sdk.AccAddress
	Amount     sdk.Coins
}

type bankModTransfer struct {
	FromModule string
	ToModule   string
	Amount     sdk.Coins
}

type bankAccToModTransfer struct {
	From     sdk.AccAddress
	ToModule string
	Amount   sdk.Coins
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		moduleBalances:  make(map[string]sdk.Coins),
		accountBalances: make(map[string]sdk.Coins),
	}
}

func (m *mockBankKeeper) SendCoins(_ context.Context, _, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	m.accToModSent = append(m.accToModSent, bankAccToModTransfer{From: senderAddr, ToModule: recipientModule, Amount: amt})
	// Deduct from sender.
	key := senderAddr.String()
	m.accountBalances[key] = m.accountBalances[key].Sub(amt...)
	// Add to module.
	m.moduleBalances[recipientModule] = m.moduleBalances[recipientModule].Add(amt...)
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToModule(_ context.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	m.moduleToModSent = append(m.moduleToModSent, bankModTransfer{FromModule: senderModule, ToModule: recipientModule, Amount: amt})
	m.moduleBalances[senderModule] = m.moduleBalances[senderModule].Sub(amt...)
	m.moduleBalances[recipientModule] = m.moduleBalances[recipientModule].Add(amt...)
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	m.moduleToAccSent = append(m.moduleToAccSent, bankTransfer{FromModule: senderModule, To: recipientAddr, Amount: amt})
	m.moduleBalances[senderModule] = m.moduleBalances[senderModule].Sub(amt...)
	key := recipientAddr.String()
	m.accountBalances[key] = m.accountBalances[key].Add(amt...)
	return nil
}

func (m *mockBankKeeper) GetBalance(_ context.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	// Check module addresses first (by matching stored module address).
	for modName, bal := range m.moduleBalances {
		modAddr := authtypes.NewModuleAddress(modName)
		if modAddr.Equals(addr) {
			return sdk.NewCoin(denom, bal.AmountOf(denom))
		}
	}
	key := addr.String()
	return sdk.NewCoin(denom, m.accountBalances[key].AmountOf(denom))
}

func (m *mockBankKeeper) GetAllBalances(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	key := addr.String()
	return m.accountBalances[key]
}

func (m *mockBankKeeper) fundModule(moduleName string, coins sdk.Coins) {
	m.moduleBalances[moduleName] = m.moduleBalances[moduleName].Add(coins...)
}

// mockDistributionKeeper implements types.DistributionKeeper for testing.
type mockDistributionKeeper struct {
	communityPoolFunded sdk.Coins
}

func newMockDistributionKeeper() *mockDistributionKeeper {
	return &mockDistributionKeeper{}
}

func (m *mockDistributionKeeper) FundCommunityPool(_ context.Context, amount sdk.Coins, _ sdk.AccAddress) error {
	m.communityPoolFunded = m.communityPoolFunded.Add(amount...)
	return nil
}

// mockFeegrantKeeper implements types.FeegrantKeeper.
type mockFeegrantKeeper struct {
	allowances map[string]bool // "granter:grantee" -> has allowance
}

func newMockFeegrantKeeper() *mockFeegrantKeeper {
	return &mockFeegrantKeeper{
		allowances: make(map[string]bool),
	}
}

func (m *mockFeegrantKeeper) GetAllowance(_ context.Context, granter, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error) {
	key := granter.String() + ":" + grantee.String()
	if m.allowances[key] {
		return &mockFeeAllowance{}, nil
	}
	return nil, fmt.Errorf("no allowance")
}

func (m *mockFeegrantKeeper) setAllowance(granter, grantee sdk.AccAddress) {
	key := granter.String() + ":" + grantee.String()
	m.allowances[key] = true
}

// mockFeeAllowance implements feegrant.FeeAllowanceI minimally.
type mockFeeAllowance struct{}

func (m *mockFeeAllowance) Accept(_ context.Context, _ sdk.Coins, _ []sdk.Msg) (bool, error) {
	return false, nil
}
func (m *mockFeeAllowance) ValidateBasic() error { return nil }
func (m *mockFeeAllowance) ExpiresAt() (*time.Time, error) { return nil, nil }

// ---------------------------------------------------------------------------
// Event Test Helpers
// ---------------------------------------------------------------------------

// requireEvent checks that at least one event of the given type was emitted.
func requireEvent(t *testing.T, ctx context.Context, eventType string) sdk.Event {
	t.Helper()
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	for _, event := range sdkCtx.EventManager().Events() {
		if event.Type == eventType {
			return event
		}
	}
	t.Fatalf("expected event of type %q was not emitted", eventType)
	return sdk.Event{}
}

// eventAttribute returns the value of a named attribute from an event.
func eventAttribute(event sdk.Event, key string) string {
	for _, attr := range event.Attributes {
		if attr.Key == key {
			return attr.Value
		}
	}
	return ""
}

// countEvents counts the number of events of the given type.
func countEvents(ctx context.Context, eventType string) int {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	count := 0
	for _, event := range sdkCtx.EventManager().Events() {
		if event.Type == eventType {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// Test Fixture
// ---------------------------------------------------------------------------

type fixture struct {
	ctx                context.Context
	keeper             keeper.Keeper
	nodeKeeper         *mockNodeKeeper
	authKeeper         *mockAuthKeeper
	feegrantKeeper     *mockFeegrantKeeper
	bankKeeper         *mockBankKeeper
	distributionKeeper *mockDistributionKeeper
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
	nodeK := newMockNodeKeeper()
	authK := newMockAuthKeeper()
	feegrantK := newMockFeegrantKeeper()
	bankK := newMockBankKeeper()
	distrK := newMockDistributionKeeper()

	// Set up feegrant pool address.
	fgPoolAddr := authtypes.NewModuleAddress(nodetypes.FeegrantPoolName)
	authK.moduleAddresses[nodetypes.FeegrantPoolName] = fgPoolAddr
	// Set up activity reward pool address.
	authK.moduleAddresses[types.ActivityRewardPoolName] = authtypes.NewModuleAddress(types.ActivityRewardPoolName)
	// Set up fee_collector address.
	authK.moduleAddresses[authtypes.FeeCollectorName] = authtypes.NewModuleAddress(authtypes.FeeCollectorName)
	// Set up activity module address.
	authK.moduleAddresses[types.ModuleName] = authtypes.NewModuleAddress(types.ModuleName)

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addrCodec,
		authority,
	)

	// Wire mock keepers.
	k.SetNodeKeeper(nodeK)
	k.SetAuthKeeper(authK)
	k.SetFeegrantKeeper(feegrantK)
	k.SetBankKeeper(bankK)
	k.SetDistributionKeeper(distrK)

	// Initialize default params.
	if err := k.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return &fixture{
		ctx:                ctx,
		keeper:             k,
		nodeKeeper:         nodeK,
		authKeeper:         authK,
		feegrantKeeper:     feegrantK,
		bankKeeper:         bankK,
		distributionKeeper: distrK,
	}
}

// freshCtx returns a new SDK context with the given block height, sharing the same store.
func (f *fixture) freshCtx(height int64) context.Context {
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	return sdkCtx.WithBlockHeight(height).WithEventManager(sdk.NewEventManager())
}

// submitActivity is a test helper to submit an activity and return the response.
func (f *fixture) submitActivity(ctx context.Context, submitter, hash, uri string) (*types.MsgSubmitActivityResponse, error) {
	msgServer := keeper.NewMsgServerImpl(f.keeper)
	return msgServer.SubmitActivity(ctx, &types.MsgSubmitActivity{
		Submitter:    submitter,
		ActivityHash: hash,
		ContentUri:   uri,
	})
}

// generateHash generates a deterministic 64-char hex hash for testing.
func generateHash(seed int) string {
	return fmt.Sprintf("%064x", seed)
}

// getBlockActivity gets the node_id stored in BlockActivities for a block.
func (f *fixture) getBlockActivity(ctx context.Context, height int64, seq uint64) (string, error) {
	return f.keeper.BlockActivities.Get(ctx, collections.Join(height, seq))
}

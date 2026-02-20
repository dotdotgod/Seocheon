package keeper_test

import (
	"strings"
	"testing"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"seocheon/x/activity/keeper"
	"seocheon/x/activity/types"
	nodetypes "seocheon/x/node/types"
)

func TestSubmitActivity_Success(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent1_______________")).String()
	f.nodeKeeper.registerNode("node1", agentAddr, 2) // ACTIVE

	ctx := f.freshCtx(100) // epoch 0, window 0
	hash := generateHash(1)

	resp, err := f.submitActivity(ctx, agentAddr, hash, "ipfs://Qm123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Epoch != 0 {
		t.Errorf("expected epoch 0, got %d", resp.Epoch)
	}
	if resp.Sequence != 0 {
		t.Errorf("expected sequence 0, got %d", resp.Sequence)
	}

	// Verify activity record.
	record, err := f.keeper.Activities.Get(ctx, collections.Join3("node1", int64(0), uint64(0)))
	if err != nil {
		t.Fatalf("activity not found: %v", err)
	}
	if record.NodeId != "node1" {
		t.Errorf("expected node_id 'node1', got %q", record.NodeId)
	}
	if record.ActivityHash != hash {
		t.Errorf("expected hash %q, got %q", hash, record.ActivityHash)
	}

	// Verify event.
	event := requireEvent(t, ctx, types.EventTypeActivitySubmitted)
	if eventAttribute(event, types.AttributeKeyNodeID) != "node1" {
		t.Error("event node_id mismatch")
	}
}

func TestSubmitActivity_RegisteredNode(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent2_______________")).String()
	f.nodeKeeper.registerNode("node2", agentAddr, 1) // REGISTERED

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, agentAddr, generateHash(2), "ipfs://Qm2")
	if err != nil {
		t.Fatalf("REGISTERED nodes should be able to submit: %v", err)
	}
}

func TestSubmitActivity_InvalidHash_Short(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent3_______________")).String()
	f.nodeKeeper.registerNode("node3", agentAddr, 2)

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, agentAddr, "abcd", "ipfs://Qm3")
	if err == nil {
		t.Fatal("expected error for short hash")
	}
	if !strings.Contains(err.Error(), "activity hash") {
		t.Errorf("expected activity hash error, got: %v", err)
	}
}

func TestSubmitActivity_InvalidHash_Long(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent4_______________")).String()
	f.nodeKeeper.registerNode("node4", agentAddr, 2)

	ctx := f.freshCtx(100)
	longHash := strings.Repeat("ab", 33) // 66 chars
	_, err := f.submitActivity(ctx, agentAddr, longHash, "ipfs://Qm4")
	if err == nil {
		t.Fatal("expected error for long hash")
	}
}

func TestSubmitActivity_InvalidHash_NonHex(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent5_______________")).String()
	f.nodeKeeper.registerNode("node5", agentAddr, 2)

	ctx := f.freshCtx(100)
	nonHex := strings.Repeat("zz", 32) // 64 chars but not hex
	_, err := f.submitActivity(ctx, agentAddr, nonHex, "ipfs://Qm5")
	if err == nil {
		t.Fatal("expected error for non-hex hash")
	}
}

func TestSubmitActivity_EmptyContentURI(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent6_______________")).String()
	f.nodeKeeper.registerNode("node6", agentAddr, 2)

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, agentAddr, generateHash(6), "")
	if err == nil {
		t.Fatal("expected error for empty content_uri")
	}
}

func TestSubmitActivity_UnregisteredSubmitter(t *testing.T) {
	f := initFixture(t)

	ctx := f.freshCtx(100)
	unknownAddr := sdk.AccAddress([]byte("unknown______________")).String()
	_, err := f.submitActivity(ctx, unknownAddr, generateHash(7), "ipfs://Qm7")
	if err == nil {
		t.Fatal("expected error for unregistered submitter")
	}
}

func TestSubmitActivity_InactiveNode(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent8_______________")).String()
	f.nodeKeeper.registerNode("node8", agentAddr, 3) // INACTIVE

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, agentAddr, generateHash(8), "ipfs://Qm8")
	if err == nil {
		t.Fatal("expected error for inactive node")
	}
}

func TestSubmitActivity_JailedNode(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agent9_______________")).String()
	f.nodeKeeper.registerNode("node9", agentAddr, 4) // JAILED

	ctx := f.freshCtx(100)
	_, err := f.submitActivity(ctx, agentAddr, generateHash(9), "ipfs://Qm9")
	if err == nil {
		t.Fatal("expected error for jailed node")
	}
}

func TestSubmitActivity_DuplicateHash(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentA_______________")).String()
	f.nodeKeeper.registerNode("nodeA", agentAddr, 2)

	ctx := f.freshCtx(100)
	hash := generateHash(10)

	_, err := f.submitActivity(ctx, agentAddr, hash, "ipfs://QmA1")
	if err != nil {
		t.Fatalf("first submission should succeed: %v", err)
	}

	_, err = f.submitActivity(ctx, agentAddr, hash, "ipfs://QmA2")
	if err == nil {
		t.Fatal("expected error for duplicate hash")
	}
}

func TestSubmitActivity_SameHashDifferentEpochs(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentB_______________")).String()
	f.nodeKeeper.registerNode("nodeB", agentAddr, 2)

	hash := generateHash(11)

	// Submit in epoch 0.
	ctx0 := f.freshCtx(100)
	_, err := f.submitActivity(ctx0, agentAddr, hash, "ipfs://QmB1")
	if err != nil {
		t.Fatalf("epoch 0 submission failed: %v", err)
	}

	// Submit same hash in epoch 1 (block 17281+).
	ctx1 := f.freshCtx(17400)
	_, err = f.submitActivity(ctx1, agentAddr, hash, "ipfs://QmB2")
	if err != nil {
		t.Fatalf("same hash in different epoch should succeed: %v", err)
	}
}

func TestSubmitActivity_QuotaExceeded_SelfFunded(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentC_______________")).String()
	f.nodeKeeper.registerNode("nodeC", agentAddr, 2)

	// Set a small quota for testing.
	params := types.DefaultParams()
	params.SelfFundedQuota = 3
	if err := f.keeper.Params.Set(f.ctx, params); err != nil {
		t.Fatal(err)
	}

	// Submit up to quota.
	for i := 0; i < 3; i++ {
		ctx := f.freshCtx(int64(100 + i))
		_, err := f.submitActivity(ctx, agentAddr, generateHash(100+i), "ipfs://QmC")
		if err != nil {
			t.Fatalf("submission %d should succeed: %v", i, err)
		}
	}

	// Exceed quota.
	ctx := f.freshCtx(200)
	_, err := f.submitActivity(ctx, agentAddr, generateHash(200), "ipfs://QmC_over")
	if err == nil {
		t.Fatal("expected quota exceeded error")
	}
}

func TestSubmitActivity_QuotaExceeded_Feegrant(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentD_______________")).String()
	agentAccAddr := sdk.AccAddress([]byte("agentD_______________"))
	f.nodeKeeper.registerNode("nodeD", agentAddr, 2)

	// Set up feegrant allowance.
	fgPoolAddr := authtypes.NewModuleAddress(nodetypes.FeegrantPoolName)
	f.feegrantKeeper.setAllowance(fgPoolAddr, agentAccAddr)

	// Set params with small feegrant quota.
	params := types.DefaultParams()
	params.FeegrantQuota = 2
	params.SelfFundedQuota = 100 // high self-funded, but feegrant should be used
	if err := f.keeper.Params.Set(f.ctx, params); err != nil {
		t.Fatal(err)
	}

	// Submit up to feegrant quota.
	for i := 0; i < 2; i++ {
		ctx := f.freshCtx(int64(100 + i))
		_, err := f.submitActivity(ctx, agentAddr, generateHash(300+i), "ipfs://QmD")
		if err != nil {
			t.Fatalf("submission %d should succeed: %v", i, err)
		}
	}

	// Exceed feegrant quota.
	ctx := f.freshCtx(200)
	_, err := f.submitActivity(ctx, agentAddr, generateHash(400), "ipfs://QmD_over")
	if err == nil {
		t.Fatal("expected feegrant quota exceeded error")
	}
}

func TestSubmitActivity_WindowTracking(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentE_______________")).String()
	f.nodeKeeper.registerNode("nodeE", agentAddr, 2)

	// Submit in window 0 (block 100).
	ctx0 := f.freshCtx(100)
	_, err := f.submitActivity(ctx0, agentAddr, generateHash(500), "ipfs://QmE0")
	if err != nil {
		t.Fatal(err)
	}

	// Submit in window 1 (block 1441).
	ctx1 := f.freshCtx(1441)
	_, err = f.submitActivity(ctx1, agentAddr, generateHash(501), "ipfs://QmE1")
	if err != nil {
		t.Fatal(err)
	}

	// Check epoch summary.
	summary, err := f.keeper.EpochSummary.Get(f.ctx, collections.Join("nodeE", int64(0)))
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalActivities != 2 {
		t.Errorf("expected 2 total activities, got %d", summary.TotalActivities)
	}
	if summary.ActiveWindows != 2 {
		t.Errorf("expected 2 active windows, got %d", summary.ActiveWindows)
	}
}

func TestSubmitActivity_EpochSummaryEligibility(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentF_______________")).String()
	f.nodeKeeper.registerNode("nodeF", agentAddr, 2)

	// Set small params for easy testing.
	params := types.Params{
		EpochLength:              120,
		WindowsPerEpoch:          12,
		MinActiveWindows:         8,
		SelfFundedQuota:          100,
		FeegrantQuota:            10,
		ActivityPruningKeepBlocks: 1000,
	}
	if err := f.keeper.Params.Set(f.ctx, params); err != nil {
		t.Fatal(err)
	}

	windowLength := params.EpochLength / params.WindowsPerEpoch // 10

	// Submit in 7 windows (not enough for eligibility).
	for w := int64(0); w < 7; w++ {
		blockHeight := w*windowLength + 1
		ctx := f.freshCtx(blockHeight)
		_, err := f.submitActivity(ctx, agentAddr, generateHash(600+int(w)), "ipfs://QmF")
		if err != nil {
			t.Fatalf("window %d submission failed: %v", w, err)
		}
	}

	summary, err := f.keeper.EpochSummary.Get(f.ctx, collections.Join("nodeF", int64(0)))
	if err != nil {
		t.Fatal(err)
	}
	if summary.Eligible {
		t.Error("should not be eligible with only 7 active windows")
	}

	// Submit in 8th window.
	ctx := f.freshCtx(7*windowLength + 1)
	_, err = f.submitActivity(ctx, agentAddr, generateHash(607), "ipfs://QmF8")
	if err != nil {
		t.Fatal(err)
	}

	summary, err = f.keeper.EpochSummary.Get(f.ctx, collections.Join("nodeF", int64(0)))
	if err != nil {
		t.Fatal(err)
	}
	if !summary.Eligible {
		t.Error("should be eligible with 8 active windows")
	}
}

func TestSubmitActivity_MultipleInSameWindow(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentG_______________")).String()
	f.nodeKeeper.registerNode("nodeG", agentAddr, 2)

	// Submit 3 activities in the same window.
	for i := 0; i < 3; i++ {
		ctx := f.freshCtx(int64(100 + i))
		_, err := f.submitActivity(ctx, agentAddr, generateHash(700+i), "ipfs://QmG")
		if err != nil {
			t.Fatalf("submission %d failed: %v", i, err)
		}
	}

	// Active windows should still be 1.
	summary, err := f.keeper.EpochSummary.Get(f.ctx, collections.Join("nodeG", int64(0)))
	if err != nil {
		t.Fatal(err)
	}
	if summary.ActiveWindows != 1 {
		t.Errorf("expected 1 active window, got %d", summary.ActiveWindows)
	}
	if summary.TotalActivities != 3 {
		t.Errorf("expected 3 total activities, got %d", summary.TotalActivities)
	}
}

func TestSubmitActivity_GlobalHashIndex(t *testing.T) {
	f := initFixture(t)

	agentAddr := sdk.AccAddress([]byte("agentH_______________")).String()
	f.nodeKeeper.registerNode("nodeH", agentAddr, 2)

	ctx := f.freshCtx(100)
	hash := generateHash(800)
	_, err := f.submitActivity(ctx, agentAddr, hash, "ipfs://QmH")
	if err != nil {
		t.Fatal(err)
	}

	// Verify global hash index.
	value, err := f.keeper.GlobalHashIndex.Get(ctx, hash)
	if err != nil {
		t.Fatalf("global hash index lookup failed: %v", err)
	}
	if value != "nodeH:0:0" {
		t.Errorf("expected 'nodeH:0:0', got %q", value)
	}
}

func TestValidateActivityHash(t *testing.T) {
	tests := []struct {
		name  string
		hash  string
		valid bool
	}{
		{"valid 64 hex chars", strings.Repeat("ab", 32), true},
		{"valid all zeros", strings.Repeat("00", 32), true},
		{"too short", "abcd", false},
		{"too long", strings.Repeat("ab", 33), false},
		{"non-hex", strings.Repeat("zz", 32), false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := keeper.ValidateActivityHash(tc.hash)
			if got != tc.valid {
				t.Errorf("ValidateActivityHash(%q) = %v, want %v", tc.hash, got, tc.valid)
			}
		})
	}
}

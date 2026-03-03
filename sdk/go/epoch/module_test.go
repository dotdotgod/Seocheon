package epoch

import (
	"context"
	"testing"

	"github.com/seocheon/sdk-go/testutil"
)

// Cat 20: Epoch module tests (3 tests)

func newTestModule() (*Module, *testutil.MockChainClient) {
	client := testutil.NewMockChainClient()

	// Default params
	client.SetRESTResponse("/seocheon/activity/v1/params", map[string]interface{}{
		"params": map[string]interface{}{
			"epoch_length":       "17280",
			"windows_per_epoch":  "12",
			"min_active_windows": "8",
		},
	})

	// Default epoch-info
	client.SetRESTResponse("/seocheon/activity/v1/epoch-info", map[string]interface{}{
		"current_epoch":           "1",
		"current_window":          "3",
		"epoch_start_block":       "17281",
		"blocks_until_next_epoch": "12960",
	})

	client.LatestBlock.Height = 21601 // epoch 1, window 3

	m := NewModule(client)
	return m, client
}

func TestGetInfo(t *testing.T) {
	m, _ := newTestModule()
	ctx := context.Background()

	resp, err := m.GetInfo(ctx)
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if resp.EpochNumber != 1 {
		t.Errorf("EpochNumber = %d, want 1", resp.EpochNumber)
	}
	if resp.WindowNumber != 3 {
		t.Errorf("WindowNumber = %d, want 3", resp.WindowNumber)
	}
	if resp.EpochStartBlock != 17281 {
		t.Errorf("EpochStartBlock = %d, want 17281", resp.EpochStartBlock)
	}
	if resp.BlockHeight == 0 {
		t.Error("BlockHeight should not be 0")
	}
}

func TestGetQualificationStatus(t *testing.T) {
	m, client := newTestModule()
	ctx := context.Background()

	client.SetRESTResponse("/seocheon/activity/v1/nodes/node-1/epochs/1", map[string]interface{}{
		"summary": map[string]interface{}{
			"total_activities": "15",
			"active_windows":   "6",
			"eligible":         false,
		},
	})

	client.SetRESTResponse("/seocheon/activity/v1/nodes/node-1/activities?epoch=1", map[string]interface{}{
		"activities": []map[string]interface{}{},
	})

	resp, err := m.GetQualification(ctx, "node-1", 1)
	if err != nil {
		t.Fatalf("GetQualification() error = %v", err)
	}
	if resp.ActiveWindows != 6 {
		t.Errorf("ActiveWindows = %d, want 6", resp.ActiveWindows)
	}
	if resp.IsQualified {
		t.Error("should not be qualified with 6/8 windows")
	}
	if resp.RequiredWindows != 8 {
		t.Errorf("RequiredWindows = %d, want 8", resp.RequiredWindows)
	}
}

func TestGetQualificationRequiresId(t *testing.T) {
	m, _ := newTestModule()
	ctx := context.Background()

	_, err := m.GetQualification(ctx, "", 1)
	if err == nil {
		t.Fatal("expected error when node_id is empty")
	}
}

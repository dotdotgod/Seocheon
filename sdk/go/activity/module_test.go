package activity

import (
	"context"
	"testing"

	"github.com/seocheon/sdk-go/testutil"
)

// Cat 18: Activity module tests (5 tests)

func newTestModule() (*Module, *testutil.MockChainClient, *testutil.MockSigner) {
	client := testutil.NewMockChainClient()
	signer := testutil.NewMockSigner()

	// Set up default REST responses
	client.SetRESTResponse("/seocheon/node/v1/nodes/by-agent/"+signer.Address, map[string]interface{}{
		"node": map[string]interface{}{
			"id": "node-1",
		},
	})
	client.SetRESTResponse("/seocheon/activity/v1/params", map[string]interface{}{
		"params": map[string]interface{}{
			"epoch_length":      "17280",
			"windows_per_epoch": "12",
		},
	})

	m := NewModule(client, signer, "seocheon-test-1")
	return m, client, signer
}

func TestGetActivities(t *testing.T) {
	m, client, _ := newTestModule()
	ctx := context.Background()

	client.SetRESTResponse("/seocheon/activity/v1/nodes/node-1/activities?epoch=1", map[string]interface{}{
		"activities": []map[string]interface{}{
			{
				"activity_hash": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				"content_uri":   "ipfs://QmTest1",
				"block_height":  "17281",
			},
			{
				"activity_hash": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
				"content_uri":   "ipfs://QmTest2",
				"block_height":  "17500",
			},
		},
	})

	resp, err := m.GetActivities(ctx, "node-1", 1)
	if err != nil {
		t.Fatalf("GetActivities() error = %v", err)
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
	if len(resp.Activities) != 2 {
		t.Errorf("len(Activities) = %d, want 2", len(resp.Activities))
	}
}

func TestGetQuota(t *testing.T) {
	m, client, signer := newTestModule()
	ctx := context.Background()

	client.SetRESTResponse("/seocheon/activity/v1/nodes/node-1/epochs/1", map[string]interface{}{
		"quota_used":  "3",
		"quota_limit": "10",
	})
	// Mock epoch info for computeCurrentEpoch
	client.LatestBlock.Height = 17281 // epoch 1

	// Mock feegrant check
	client.SetRESTResponse("/cosmos/feegrant/v1beta1/allowance/feegrant_pool/"+signer.Address, map[string]interface{}{
		"allowance": map[string]string{"type": "test"},
	})

	resp, err := m.GetQuota(ctx)
	if err != nil {
		t.Fatalf("GetQuota() error = %v", err)
	}
	if resp.QuotaUsed != 3 {
		t.Errorf("QuotaUsed = %d, want 3", resp.QuotaUsed)
	}
	if resp.QuotaTotal != 10 {
		t.Errorf("QuotaTotal = %d, want 10", resp.QuotaTotal)
	}
	if resp.QuotaRemaining != 7 {
		t.Errorf("QuotaRemaining = %d, want 7", resp.QuotaRemaining)
	}
}

func TestSubmitInvalidHash(t *testing.T) {
	m, _, _ := newTestModule()
	ctx := context.Background()

	// Invalid hash (too short)
	_, err := m.Submit(ctx, "invalidhash", "ipfs://QmTest")
	if err == nil {
		t.Fatal("expected error for invalid activity hash")
	}
}

func TestSubmitEmptyUri(t *testing.T) {
	m, _, _ := newTestModule()
	ctx := context.Background()

	validHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	_, err := m.Submit(ctx, validHash, "")
	if err == nil {
		t.Fatal("expected error for empty content URI")
	}
}

func TestSubmitNotRegistered(t *testing.T) {
	m, client, _ := newTestModule()
	ctx := context.Background()

	// Override the node lookup to return error (node not registered)
	client.SetRESTError("/seocheon/node/v1/nodes/by-agent/"+m.signer.GetAddress(),
		testutil.ErrMockNotFound)

	_, err := m.GetActivities(ctx, "", 0)
	if err == nil {
		t.Fatal("expected error for unregistered submitter")
	}
}

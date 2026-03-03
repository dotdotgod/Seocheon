package node

import (
	"context"
	"testing"

	"github.com/seocheon/sdk-go/testutil"
)

// Cat 19: Node module tests (4 tests)

func newTestModule() (*Module, *testutil.MockChainClient) {
	client := testutil.NewMockChainClient()
	signer := testutil.NewMockSigner()

	// Default node lookup by agent address
	client.SetRESTResponse("/seocheon/node/v1/nodes/by-agent/"+signer.Address, map[string]interface{}{
		"node": map[string]interface{}{
			"id": "node-1",
		},
	})

	m := NewModule(client, signer)
	return m, client
}

func TestGetInfoById(t *testing.T) {
	m, client := newTestModule()
	ctx := context.Background()

	client.SetRESTResponse("/seocheon/node/v1/nodes/node-42", map[string]interface{}{
		"node": map[string]interface{}{
			"id":                "node-42",
			"operator":          "seocheon1operator",
			"agent_address":     "seocheon1agent",
			"status":            1,
			"description":       "Test node",
			"website":           "https://example.com",
			"tags":              []string{"ai", "defi"},
			"agent_share":       "0.2",
			"validator_address": "",
			"registered_at":     "100",
		},
	})

	resp, err := m.GetInfo(ctx, "node-42")
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
	if resp.NodeID != "node-42" {
		t.Errorf("NodeID = %s, want node-42", resp.NodeID)
	}
	if resp.Operator != "seocheon1operator" {
		t.Errorf("Operator = %s, want seocheon1operator", resp.Operator)
	}
	if resp.Status != "REGISTERED" {
		t.Errorf("Status = %s, want REGISTERED", resp.Status)
	}
	if resp.Description != "Test node" {
		t.Errorf("Description = %s, want 'Test node'", resp.Description)
	}
}

func TestGetInfoOwnNode(t *testing.T) {
	m, client := newTestModule()
	ctx := context.Background()

	client.SetRESTResponse("/seocheon/node/v1/nodes/node-1", map[string]interface{}{
		"node": map[string]interface{}{
			"id":                "node-1",
			"operator":          "seocheon1me",
			"agent_address":     "seocheon1agent",
			"status":            2,
			"description":       "My node",
			"validator_address": "",
			"registered_at":     "50",
		},
	})

	// Empty nodeID should resolve to own node
	resp, err := m.GetInfo(ctx, "")
	if err != nil {
		t.Fatalf("GetInfo('') error = %v", err)
	}
	if resp.NodeID != "node-1" {
		t.Errorf("NodeID = %s, want node-1 (own node)", resp.NodeID)
	}
}

func TestSearchNodes(t *testing.T) {
	m, client := newTestModule()
	ctx := context.Background()

	client.SetRESTResponse("/seocheon/node/v1/nodes", map[string]interface{}{
		"nodes": []map[string]interface{}{
			{
				"id":                "node-1",
				"status":            2,
				"description":       "Node A",
				"tags":              []string{"ai"},
				"validator_address": "",
				"registered_at":     "100",
			},
			{
				"id":                "node-2",
				"status":            1,
				"description":       "Node B",
				"tags":              []string{"data"},
				"validator_address": "",
				"registered_at":     "200",
			},
		},
	})

	resp, err := m.Search(ctx, "", "", 10, "")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resp.Nodes) != 2 {
		t.Errorf("len(Nodes) = %d, want 2", len(resp.Nodes))
	}
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestNotFoundFails(t *testing.T) {
	m, client := newTestModule()
	ctx := context.Background()

	client.SetRESTError("/seocheon/node/v1/nodes/nonexistent", testutil.ErrMockNotFound)

	_, err := m.GetInfo(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent node")
	}
}

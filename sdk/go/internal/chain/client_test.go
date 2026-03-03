package chain

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Cat 17: Chain client tests (5 tests)

func TestHTTPClientStartsDisconnected(t *testing.T) {
	client := NewHTTPClient("http://localhost:26657", "http://localhost:1317")
	if client.IsConnected() {
		t.Error("client should start disconnected")
	}
}

func TestHTTPClientConnectStoresEndpoints(t *testing.T) {
	// Create mock server that returns a valid block response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"block": map[string]interface{}{
				"header": map[string]interface{}{
					"height":   "100",
					"time":     "2026-03-01T00:00:00Z",
					"chain_id": "seocheon-test-1",
				},
				"data": map[string]interface{}{
					"txs": []string{},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, server.URL)
	if err := client.Connect(context.Background()); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	if !client.IsConnected() {
		t.Error("client should be connected after Connect()")
	}
}

func TestHTTPClientDisconnectClearsState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"block": map[string]interface{}{
				"header": map[string]interface{}{
					"height":   "100",
					"time":     "2026-03-01T00:00:00Z",
					"chain_id": "seocheon-test-1",
				},
				"data": map[string]interface{}{
					"txs": []string{},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, server.URL)
	client.Connect(context.Background())
	if !client.IsConnected() {
		t.Fatal("should be connected before disconnect test")
	}

	client.Disconnect()
	if client.IsConnected() {
		t.Error("client should not be connected after Disconnect()")
	}
}

func TestHTTPClientStripsTrailingSlash(t *testing.T) {
	client := NewHTTPClient("http://localhost:26657/", "http://localhost:1317/")
	if client.rpcEndpoint != "http://localhost:26657" {
		t.Errorf("rpcEndpoint = %s, want trailing slash stripped", client.rpcEndpoint)
	}
	if client.grpcEndpoint != "http://localhost:1317" {
		t.Errorf("grpcEndpoint = %s, want trailing slash stripped", client.grpcEndpoint)
	}
}

func TestHTTPClientOpsFailWhenDisconnected(t *testing.T) {
	// Using a non-existent server URL to ensure failure
	client := NewHTTPClient("http://127.0.0.1:19999", "http://127.0.0.1:19998")

	ctx := context.Background()

	// QueryREST should fail when endpoint is unreachable
	_, err := client.QueryREST(ctx, "/test")
	if err == nil {
		t.Error("QueryREST should fail when server is unreachable")
	}
}

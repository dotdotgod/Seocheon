package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestConnection_ConnectAndClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Logf("accept error: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")
		// Read loop to handle close frame from client
		for {
			_, _, err := conn.Read(r.Context())
			if err != nil {
				return
			}
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/websocket"

	c := NewConnection(wsURL)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestConnection_Subscribe(t *testing.T) {
	eventSent := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Logf("accept error: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		ctx := r.Context()

		// Read the subscribe request
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(data, &req); err != nil {
			return
		}

		if req.Method != "subscribe" {
			t.Errorf("expected subscribe method, got %s", req.Method)
			return
		}

		// Send subscription confirmation
		confirmResp := JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  json.RawMessage(`{}`),
		}
		confirmBytes, _ := json.Marshal(confirmResp)
		conn.Write(ctx, websocket.MessageText, confirmBytes)

		// Send a test event
		eventData := map[string]interface{}{
			"query": "tm.event='Tx'",
			"data": map[string]interface{}{
				"type":  "tendermint/event/Tx",
				"value": map[string]string{"tx_hash": "ABCD1234"},
			},
		}
		eventBytes, _ := json.Marshal(eventData)
		eventResp := JSONRPCResponse{
			JSONRPC: "2.0",
			Result:  json.RawMessage(eventBytes),
		}
		respBytes, _ := json.Marshal(eventResp)
		conn.Write(ctx, websocket.MessageText, respBytes)
		close(eventSent)

		<-ctx.Done()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/websocket"

	c := NewConnection(wsURL)
	c.reconnect = false

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Connect(ctx); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer c.Close()

	ch, err := c.Subscribe(ctx, "test-1", "tm.event='Tx'")
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	// Wait for the event
	select {
	case event := <-ch:
		if event == nil {
			t.Fatal("received nil event")
		}
		t.Logf("Received event: %s", string(event))
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestNewConnection(t *testing.T) {
	c := NewConnection("ws://localhost:26657/websocket")
	if c.url != "ws://localhost:26657/websocket" {
		t.Errorf("url = %s, want ws://localhost:26657/websocket", c.url)
	}
	if c.maxBackoff != 30*time.Second {
		t.Errorf("maxBackoff = %v, want 30s", c.maxBackoff)
	}
}

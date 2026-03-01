// Package subscription implements WebSocket-based event subscription
// for the Seocheon SDK.
package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// Event represents a chain event received via subscription.
type Event struct {
	Type       string            `json:"type"`
	Attributes map[string]string `json:"attributes"`
	Height     int64             `json:"height"`
}

// BlockEvent represents a new block event.
type BlockEvent struct {
	Height  int64  `json:"height"`
	Time    string `json:"time"`
	ChainID string `json:"chain_id"`
	NumTxs  int    `json:"num_txs"`
}

// Manager manages WebSocket subscriptions to chain events.
type Manager struct {
	rpcEndpoint string
	mu          sync.Mutex
	cancelFuncs []context.CancelFunc
}

// NewManager creates a new subscription Manager.
func NewManager(rpcEndpoint string) *Manager {
	return &Manager{
		rpcEndpoint: rpcEndpoint,
	}
}

// Subscribe creates a subscription with the given Tendermint query.
// The returned channel receives events matching the query.
// Example query: "tm.event='Tx' AND message.action='/seocheon.activity.v1.MsgSubmitActivity'"
func (m *Manager) Subscribe(ctx context.Context, query string) (<-chan Event, error) {
	wsURL, err := m.buildWSURL()
	if err != nil {
		return nil, fmt.Errorf("building WebSocket URL: %w", err)
	}

	ch := make(chan Event, 100)
	subCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.cancelFuncs = append(m.cancelFuncs, cancel)
	m.mu.Unlock()

	// TODO: Implement actual WebSocket connection and message parsing
	// The connection would:
	// 1. Dial the WebSocket endpoint
	// 2. Send a subscribe request with the query
	// 3. Read events and send them to the channel
	// 4. Handle reconnection on disconnects
	_ = wsURL
	_ = query

	go func() {
		<-subCtx.Done()
		close(ch)
	}()

	return ch, nil
}

// SubscribeNewBlock creates a subscription for new block events.
func (m *Manager) SubscribeNewBlock(ctx context.Context) (<-chan BlockEvent, error) {
	ch := make(chan BlockEvent, 100)
	subCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.cancelFuncs = append(m.cancelFuncs, cancel)
	m.mu.Unlock()

	// TODO: Implement actual WebSocket subscription for tm.event='NewBlock'
	go func() {
		<-subCtx.Done()
		close(ch)
	}()

	return ch, nil
}

// SubscribeActivity creates a subscription specifically for MsgSubmitActivity events.
func (m *Manager) SubscribeActivity(ctx context.Context) (<-chan Event, error) {
	return m.Subscribe(ctx, "tm.event='Tx' AND message.action='/seocheon.activity.v1.MsgSubmitActivity'")
}

// UnsubscribeAll cancels all active subscriptions.
func (m *Manager) UnsubscribeAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cancel := range m.cancelFuncs {
		cancel()
	}
	m.cancelFuncs = nil
}

func (m *Manager) buildWSURL() (string, error) {
	endpoint := m.rpcEndpoint
	endpoint = strings.TrimRight(endpoint, "/")

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("parsing RPC endpoint: %w", err)
	}

	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	u.Path = "/websocket"
	return u.String(), nil
}

// ParseEventAttributes parses raw JSON event data into typed attributes.
func ParseEventAttributes(raw json.RawMessage) (map[string]string, error) {
	var attrs []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &attrs); err != nil {
		return nil, err
	}

	result := make(map[string]string, len(attrs))
	for _, a := range attrs {
		result[a.Key] = a.Value
	}
	return result, nil
}

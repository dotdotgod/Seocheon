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
	"sync/atomic"

	"github.com/seocheon/sdk-go/internal/ws"
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
	wsConn      *ws.Connection
	mu          sync.Mutex
	cancelFuncs []context.CancelFunc
	subCounter  atomic.Int64
}

// NewManager creates a new subscription Manager.
func NewManager(rpcEndpoint string) *Manager {
	return &Manager{
		rpcEndpoint: rpcEndpoint,
	}
}

// Connect establishes the WebSocket connection for subscriptions.
func (m *Manager) Connect(ctx context.Context) error {
	wsURL, err := m.buildWSURL()
	if err != nil {
		return fmt.Errorf("building WebSocket URL: %w", err)
	}

	m.wsConn = ws.NewConnection(wsURL)
	return m.wsConn.Connect(ctx)
}

// Subscribe creates a subscription with the given Tendermint query.
// The returned channel receives events matching the query.
// Example query: "tm.event='Tx' AND message.action='/seocheon.activity.v1.MsgSubmitActivity'"
func (m *Manager) Subscribe(ctx context.Context, query string) (<-chan Event, error) {
	if m.wsConn == nil {
		if err := m.Connect(ctx); err != nil {
			return nil, err
		}
	}

	subID := fmt.Sprintf("sub-%d", m.subCounter.Add(1))

	subCtx, cancel := context.WithCancel(ctx)
	m.mu.Lock()
	m.cancelFuncs = append(m.cancelFuncs, cancel)
	m.mu.Unlock()

	rawCh, err := m.wsConn.Subscribe(subCtx, subID, query)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("subscribing to %s: %w", query, err)
	}

	ch := make(chan Event, 100)
	go func() {
		defer close(ch)
		for raw := range rawCh {
			event, err := parseRawEvent(raw)
			if err != nil {
				continue
			}
			select {
			case ch <- event:
			case <-subCtx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// SubscribeNewBlock creates a subscription for new block events.
func (m *Manager) SubscribeNewBlock(ctx context.Context) (<-chan BlockEvent, error) {
	if m.wsConn == nil {
		if err := m.Connect(ctx); err != nil {
			return nil, err
		}
	}

	subID := fmt.Sprintf("block-%d", m.subCounter.Add(1))

	subCtx, cancel := context.WithCancel(ctx)
	m.mu.Lock()
	m.cancelFuncs = append(m.cancelFuncs, cancel)
	m.mu.Unlock()

	rawCh, err := m.wsConn.Subscribe(subCtx, subID, "tm.event='NewBlock'")
	if err != nil {
		cancel()
		return nil, fmt.Errorf("subscribing to new blocks: %w", err)
	}

	ch := make(chan BlockEvent, 100)
	go func() {
		defer close(ch)
		for raw := range rawCh {
			blockEvt, err := parseRawBlockEvent(raw)
			if err != nil {
				continue
			}
			select {
			case ch <- blockEvt:
			case <-subCtx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// SubscribeActivity creates a subscription specifically for MsgSubmitActivity events.
func (m *Manager) SubscribeActivity(ctx context.Context) (<-chan Event, error) {
	return m.Subscribe(ctx, "tm.event='Tx' AND message.action='/seocheon.activity.v1.MsgSubmitActivity'")
}

// UnsubscribeAll cancels all active subscriptions and closes the WebSocket connection.
func (m *Manager) UnsubscribeAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cancel := range m.cancelFuncs {
		cancel()
	}
	m.cancelFuncs = nil

	if m.wsConn != nil {
		m.wsConn.Close()
		m.wsConn = nil
	}
}

func (m *Manager) buildWSURL() (string, error) {
	endpoint := strings.TrimRight(m.rpcEndpoint, "/")

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

func parseRawEvent(raw json.RawMessage) (Event, error) {
	var result struct {
		Data struct {
			Value struct {
				TxResult struct {
					Height json.Number `json:"height"`
					Result struct {
						Events []struct {
							Type       string `json:"type"`
							Attributes []struct {
								Key   string `json:"key"`
								Value string `json:"value"`
							} `json:"attributes"`
						} `json:"events"`
					} `json:"result"`
				} `json:"TxResult"`
			} `json:"value"`
		} `json:"data"`
		Events map[string][]string `json:"events"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return Event{}, err
	}

	height, _ := result.Data.Value.TxResult.Height.Int64()

	// Use top-level events map if available
	attrs := make(map[string]string)
	for k, v := range result.Events {
		if len(v) > 0 {
			attrs[k] = v[0]
		}
	}

	eventType := ""
	if len(result.Data.Value.TxResult.Result.Events) > 0 {
		eventType = result.Data.Value.TxResult.Result.Events[0].Type
		for _, attr := range result.Data.Value.TxResult.Result.Events[0].Attributes {
			attrs[attr.Key] = attr.Value
		}
	}

	return Event{
		Type:       eventType,
		Attributes: attrs,
		Height:     height,
	}, nil
}

func parseRawBlockEvent(raw json.RawMessage) (BlockEvent, error) {
	var result struct {
		Data struct {
			Value struct {
				Block struct {
					Header struct {
						Height  json.Number `json:"height"`
						Time    string      `json:"time"`
						ChainID string      `json:"chain_id"`
					} `json:"header"`
					Data struct {
						Txs []json.RawMessage `json:"txs"`
					} `json:"data"`
				} `json:"block"`
			} `json:"value"`
		} `json:"data"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return BlockEvent{}, err
	}

	height, _ := result.Data.Value.Block.Header.Height.Int64()

	return BlockEvent{
		Height:  height,
		Time:    result.Data.Value.Block.Header.Time,
		ChainID: result.Data.Value.Block.Header.ChainID,
		NumTxs:  len(result.Data.Value.Block.Data.Txs),
	}, nil
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

package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// Connection manages a WebSocket connection to a CometBFT node.
type Connection struct {
	url        string
	conn       *websocket.Conn
	mu         sync.Mutex
	closed     bool
	reconnect  bool
	maxBackoff time.Duration
}

// NewConnection creates a new WebSocket connection manager.
func NewConnection(url string) *Connection {
	return &Connection{
		url:        url,
		reconnect:  true,
		maxBackoff: 30 * time.Second,
	}
}

// Connect establishes the WebSocket connection.
func (c *Connection) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil
	}

	conn, _, err := websocket.Dial(ctx, c.url, &websocket.DialOptions{})
	if err != nil {
		return fmt.Errorf("dialing websocket %s: %w", c.url, err)
	}

	c.conn = conn
	c.closed = false
	return nil
}

// Close closes the WebSocket connection.
func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	c.reconnect = false
	if c.conn != nil {
		err := c.conn.Close(websocket.StatusNormalClosure, "client closing")
		c.conn = nil
		return err
	}
	return nil
}

// JSONRPCRequest is a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	ID      string      `json:"id"`
	Params  interface{} `json:"params"`
}

// JSONRPCResponse is a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// Subscribe sends a subscribe request and returns a channel for receiving events.
// The query follows CometBFT event query syntax.
func (c *Connection) Subscribe(ctx context.Context, id, query string) (<-chan json.RawMessage, error) {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("websocket not connected")
	}
	conn := c.conn
	c.mu.Unlock()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "subscribe",
		ID:      id,
		Params:  map[string]string{"query": query},
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling subscribe request: %w", err)
	}

	if err := conn.Write(ctx, websocket.MessageText, reqBytes); err != nil {
		return nil, fmt.Errorf("sending subscribe request: %w", err)
	}

	ch := make(chan json.RawMessage, 100)

	go c.readLoop(ctx, conn, ch)

	return ch, nil
}

// Unsubscribe sends an unsubscribe request.
func (c *Connection) Unsubscribe(ctx context.Context, id, query string) error {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return fmt.Errorf("websocket not connected")
	}
	conn := c.conn
	c.mu.Unlock()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "unsubscribe",
		ID:      id,
		Params:  map[string]string{"query": query},
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling unsubscribe request: %w", err)
	}

	return conn.Write(ctx, websocket.MessageText, reqBytes)
}

// readLoop reads messages from the WebSocket and dispatches them to the channel.
// It handles reconnection with exponential backoff.
func (c *Connection) readLoop(ctx context.Context, conn *websocket.Conn, ch chan<- json.RawMessage) {
	defer close(ch)

	backoff := time.Second

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

			c.mu.Lock()
			shouldReconnect := c.reconnect && !c.closed
			c.mu.Unlock()

			if !shouldReconnect {
				return
			}

			// Exponential backoff reconnection
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(backoff):
				}

				newConn, _, dialErr := websocket.Dial(ctx, c.url, &websocket.DialOptions{})
				if dialErr != nil {
					backoff *= 2
					if backoff > c.maxBackoff {
						backoff = c.maxBackoff
					}
					continue
				}

				c.mu.Lock()
				c.conn = newConn
				conn = newConn
				c.mu.Unlock()

				backoff = time.Second
				break
			}
			continue
		}

		// Reset backoff on successful read
		backoff = time.Second

		var resp JSONRPCResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			continue
		}

		// Skip initial subscription confirmation response
		if resp.ID != "" && resp.Result != nil {
			// Check if it's just a subscription confirmation (empty object)
			trimmed := string(resp.Result)
			if trimmed == "{}" || trimmed == "null" {
				continue
			}
		}

		if resp.Result != nil {
			select {
			case ch <- resp.Result:
			default:
				// Channel full, drop oldest
			}
		}
	}
}

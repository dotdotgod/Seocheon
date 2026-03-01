// Package chain provides the ChainClient interface and HTTP-based implementation
// for communicating with a Seocheon/Cosmos blockchain node.
package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client defines the interface for chain communication.
type Client interface {
	// Connect initializes the connection to the chain.
	Connect(ctx context.Context) error
	// Disconnect closes the connection.
	Disconnect() error
	// IsConnected returns whether the client is connected.
	IsConnected() bool

	// QueryREST performs a GET request to the REST API.
	QueryREST(ctx context.Context, path string) (json.RawMessage, error)
	// BroadcastTx broadcasts a transaction via REST.
	BroadcastTx(ctx context.Context, txBytes []byte, mode string) (*BroadcastResponse, error)
	// GetLatestBlock returns the latest block info.
	GetLatestBlock(ctx context.Context) (*BlockResponse, error)
	// GetTx queries a transaction by hash.
	GetTx(ctx context.Context, txHash string) (*TxResponse, error)
	// GetAccountInfo returns account number and sequence.
	GetAccountInfo(ctx context.Context, address string) (*AccountInfo, error)
}

// BroadcastResponse holds the result of a TX broadcast.
type BroadcastResponse struct {
	TxHash string `json:"txhash"`
	Code   uint32 `json:"code"`
	RawLog string `json:"raw_log"`
}

// BlockResponse holds basic block information.
type BlockResponse struct {
	Height  int64  `json:"height"`
	Time    string `json:"time"`
	ChainID string `json:"chain_id"`
	NumTxs  int    `json:"num_txs"`
}

// TxResponse holds a transaction query result.
type TxResponse struct {
	TxHash    string      `json:"txhash"`
	Height    int64       `json:"height"`
	Code      uint32      `json:"code"`
	GasUsed   uint64      `json:"gas_used"`
	GasWanted uint64      `json:"gas_wanted"`
	RawLog    string      `json:"raw_log"`
	Events    []TxEvent   `json:"events"`
}

// TxEvent represents an event from a transaction.
type TxEvent struct {
	Type       string           `json:"type"`
	Attributes []EventAttribute `json:"attributes"`
}

// EventAttribute is a key-value pair in a TxEvent.
type EventAttribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AccountInfo holds account number and sequence for TX signing.
type AccountInfo struct {
	AccountNumber uint64 `json:"account_number"`
	Sequence      uint64 `json:"sequence"`
}

// HTTPClient is the default HTTP-based implementation of Client.
type HTTPClient struct {
	rpcEndpoint  string
	grpcEndpoint string
	httpClient   *http.Client
	connected    bool
}

// NewHTTPClient creates a new HTTP-based chain client.
func NewHTTPClient(rpcEndpoint, grpcEndpoint string) *HTTPClient {
	return &HTTPClient{
		rpcEndpoint:  strings.TrimRight(rpcEndpoint, "/"),
		grpcEndpoint: strings.TrimRight(grpcEndpoint, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Connect tests the connection to the chain node.
func (c *HTTPClient) Connect(ctx context.Context) error {
	_, err := c.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to chain: %w", err)
	}
	c.connected = true
	return nil
}

// Disconnect closes the HTTP client.
func (c *HTTPClient) Disconnect() error {
	c.connected = false
	c.httpClient.CloseIdleConnections()
	return nil
}

// IsConnected returns whether the client has successfully connected.
func (c *HTTPClient) IsConnected() bool {
	return c.connected
}

// QueryREST performs a GET request to the REST endpoint.
func (c *HTTPClient) QueryREST(ctx context.Context, path string) (json.RawMessage, error) {
	url := c.grpcEndpoint + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	return json.RawMessage(body), nil
}

// BroadcastTx broadcasts a signed transaction.
func (c *HTTPClient) BroadcastTx(ctx context.Context, txBytes []byte, mode string) (*BroadcastResponse, error) {
	payload := fmt.Sprintf(`{"tx_bytes":"%s","mode":"%s"}`,
		encodeBase64(txBytes), broadcastModeToProto(mode))

	url := c.grpcEndpoint + "/cosmos/tx/v1beta1/txs"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url,
		strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating broadcast request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("broadcasting tx: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading broadcast response: %w", err)
	}

	var result struct {
		TxResponse BroadcastResponse `json:"tx_response"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing broadcast response: %w", err)
	}

	return &result.TxResponse, nil
}

// GetLatestBlock returns the latest block information.
func (c *HTTPClient) GetLatestBlock(ctx context.Context) (*BlockResponse, error) {
	data, err := c.QueryREST(ctx, "/cosmos/base/tendermint/v1beta1/blocks/latest")
	if err != nil {
		return nil, fmt.Errorf("querying latest block: %w", err)
	}

	var result struct {
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
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing block response: %w", err)
	}

	height, _ := result.Block.Header.Height.Int64()
	return &BlockResponse{
		Height:  height,
		Time:    result.Block.Header.Time,
		ChainID: result.Block.Header.ChainID,
		NumTxs:  len(result.Block.Data.Txs),
	}, nil
}

// GetTx queries a transaction by hash.
func (c *HTTPClient) GetTx(ctx context.Context, txHash string) (*TxResponse, error) {
	path := fmt.Sprintf("/cosmos/tx/v1beta1/txs/%s", txHash)
	data, err := c.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("querying tx %s: %w", txHash, err)
	}

	var result struct {
		TxResponse struct {
			TxHash    string      `json:"txhash"`
			Height    json.Number `json:"height"`
			Code      uint32      `json:"code"`
			GasUsed   json.Number `json:"gas_used"`
			GasWanted json.Number `json:"gas_wanted"`
			RawLog    string      `json:"raw_log"`
			Events    []TxEvent   `json:"events"`
		} `json:"tx_response"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing tx response: %w", err)
	}

	height, _ := result.TxResponse.Height.Int64()
	gasUsed, _ := result.TxResponse.GasUsed.Int64()
	gasWanted, _ := result.TxResponse.GasWanted.Int64()

	return &TxResponse{
		TxHash:    result.TxResponse.TxHash,
		Height:    height,
		Code:      result.TxResponse.Code,
		GasUsed:   uint64(gasUsed),
		GasWanted: uint64(gasWanted),
		RawLog:    result.TxResponse.RawLog,
		Events:    result.TxResponse.Events,
	}, nil
}

// GetAccountInfo returns the account number and sequence for an address.
func (c *HTTPClient) GetAccountInfo(ctx context.Context, address string) (*AccountInfo, error) {
	path := fmt.Sprintf("/cosmos/auth/v1beta1/accounts/%s", address)
	data, err := c.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("querying account %s: %w", address, err)
	}

	var result struct {
		Account struct {
			AccountNumber json.Number `json:"account_number"`
			Sequence      json.Number `json:"sequence"`
		} `json:"account"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing account response: %w", err)
	}

	accNum, _ := result.Account.AccountNumber.Int64()
	seq, _ := result.Account.Sequence.Int64()

	return &AccountInfo{
		AccountNumber: uint64(accNum),
		Sequence:      uint64(seq),
	}, nil
}

// encodeBase64 encodes bytes to base64 string.
func encodeBase64(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, 0, (len(data)+2)/3*4)
	for i := 0; i < len(data); i += 3 {
		var b0, b1, b2 byte
		b0 = data[i]
		if i+1 < len(data) {
			b1 = data[i+1]
		}
		if i+2 < len(data) {
			b2 = data[i+2]
		}
		result = append(result, base64Chars[(b0>>2)&0x3F])
		result = append(result, base64Chars[((b0<<4)|(b1>>4))&0x3F])
		if i+1 < len(data) {
			result = append(result, base64Chars[((b1<<2)|(b2>>6))&0x3F])
		} else {
			result = append(result, '=')
		}
		if i+2 < len(data) {
			result = append(result, base64Chars[b2&0x3F])
		} else {
			result = append(result, '=')
		}
	}
	return string(result)
}

// broadcastModeToProto converts SDK broadcast mode to proto enum string.
func broadcastModeToProto(mode string) string {
	switch mode {
	case "async":
		return "BROADCAST_MODE_ASYNC"
	default:
		return "BROADCAST_MODE_SYNC"
	}
}

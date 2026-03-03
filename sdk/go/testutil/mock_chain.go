// Package testutil provides test helpers and mock implementations for the Seocheon SDK.
package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/seocheon/sdk-go/internal/chain"
)

// MockChainClient implements chain.Client for testing without a real chain connection.
type MockChainClient struct {
	mu           sync.Mutex
	connected    bool
	rpcEndpoint  string
	grpcEndpoint string

	// RESTResponses maps REST paths to response data.
	RESTResponses map[string]json.RawMessage
	// RESTErrors maps REST paths to errors.
	RESTErrors map[string]error

	// BroadcastResult is the result returned by BroadcastTx.
	BroadcastResult *chain.BroadcastResponse
	BroadcastErr    error

	// LatestBlock is the result returned by GetLatestBlock.
	LatestBlock *chain.BlockResponse
	LatestErr   error

	// TxResults maps tx hashes to results.
	TxResults map[string]*chain.TxResponse
	TxErr     error

	// AccountResult is the result returned by GetAccountInfo.
	AccountResult *chain.AccountInfo
	AccountErr    error
}

// NewMockChainClient creates a new mock chain client with sensible defaults.
func NewMockChainClient() *MockChainClient {
	return &MockChainClient{
		rpcEndpoint:  "http://localhost:26657",
		grpcEndpoint: "http://localhost:1317",
		RESTResponses: make(map[string]json.RawMessage),
		RESTErrors:    make(map[string]error),
		TxResults:     make(map[string]*chain.TxResponse),
		LatestBlock: &chain.BlockResponse{
			Height:  17280,
			Time:    "2026-03-01T00:00:00Z",
			ChainID: "seocheon-test-1",
			NumTxs:  5,
		},
		AccountResult: &chain.AccountInfo{
			AccountNumber: 42,
			Sequence:      5,
		},
		BroadcastResult: &chain.BroadcastResponse{
			TxHash: "ABCDEF1234567890",
			Code:   0,
		},
	}
}

func (m *MockChainClient) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = true
	return nil
}

func (m *MockChainClient) Disconnect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connected = false
	return nil
}

func (m *MockChainClient) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

func (m *MockChainClient) QueryREST(ctx context.Context, path string) (json.RawMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.RESTErrors[path]; ok {
		return nil, err
	}
	if resp, ok := m.RESTResponses[path]; ok {
		return resp, nil
	}
	return nil, fmt.Errorf("no mock response for path: %s", path)
}

func (m *MockChainClient) BroadcastTx(ctx context.Context, txBytes []byte, mode string) (*chain.BroadcastResponse, error) {
	if m.BroadcastErr != nil {
		return nil, m.BroadcastErr
	}
	return m.BroadcastResult, nil
}

func (m *MockChainClient) GetLatestBlock(ctx context.Context) (*chain.BlockResponse, error) {
	if m.LatestErr != nil {
		return nil, m.LatestErr
	}
	return m.LatestBlock, nil
}

func (m *MockChainClient) GetTx(ctx context.Context, txHash string) (*chain.TxResponse, error) {
	if m.TxErr != nil {
		return nil, m.TxErr
	}
	if result, ok := m.TxResults[txHash]; ok {
		return result, nil
	}
	return nil, fmt.Errorf("tx not found: %s", txHash)
}

func (m *MockChainClient) GetAccountInfo(ctx context.Context, address string) (*chain.AccountInfo, error) {
	if m.AccountErr != nil {
		return nil, m.AccountErr
	}
	return m.AccountResult, nil
}

// SetRESTResponse sets a mock REST response for a given path.
func (m *MockChainClient) SetRESTResponse(path string, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	bytes, _ := json.Marshal(data)
	m.RESTResponses[path] = json.RawMessage(bytes)
}

// SetRESTError sets a mock REST error for a given path.
func (m *MockChainClient) SetRESTError(path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RESTErrors[path] = err
}

// MockSigner implements signing.Service for testing.
type MockSigner struct {
	Address string
	PubKeyB []byte
	SignFn  func(data []byte) ([]byte, error)
}

// NewMockSigner creates a new mock signer with default behavior.
func NewMockSigner() *MockSigner {
	pubKey := make([]byte, 33)
	pubKey[0] = 0x02
	for i := 1; i < 33; i++ {
		pubKey[i] = byte(i)
	}
	return &MockSigner{
		Address: "seocheon1mockaddr123456789",
		PubKeyB: pubKey,
		SignFn: func(data []byte) ([]byte, error) {
			sig := make([]byte, 64)
			for i := range sig {
				sig[i] = byte(i % 256)
			}
			return sig, nil
		},
	}
}

func (m *MockSigner) Sign(data []byte) ([]byte, error) { return m.SignFn(data) }
func (m *MockSigner) GetAddress() string                { return m.Address }
func (m *MockSigner) GetPubKey() ([]byte, error)        { return m.PubKeyB, nil }

// ErrMockNotFound is a sentinel error for mock "not found" responses.
var ErrMockNotFound = fmt.Errorf("mock: not found")

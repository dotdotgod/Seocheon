package tx

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// mockSigner implements the Signer interface for testing.
type mockSigner struct {
	address string
	pubKey  []byte
	signFn  func(data []byte) ([]byte, error)
}

func newMockSigner() *mockSigner {
	pubKey := make([]byte, 33)
	pubKey[0] = 0x02
	for i := 1; i < 33; i++ {
		pubKey[i] = byte(i)
	}
	return &mockSigner{
		address: "seocheon1testaddr",
		pubKey:  pubKey,
		signFn: func(data []byte) ([]byte, error) {
			sig := make([]byte, 64)
			for i := range sig {
				sig[i] = byte(i % 256)
			}
			return sig, nil
		},
	}
}

func (m *mockSigner) Sign(data []byte) ([]byte, error) { return m.signFn(data) }
func (m *mockSigner) GetAddress() string               { return m.address }
func (m *mockSigner) GetPubKey() ([]byte, error)       { return m.pubKey, nil }

// mockQuerier implements the ChainQuerier interface for testing.
type mockQuerier struct {
	accountNumber uint64
	sequence      uint64
	broadcastHash string
	broadcastCode uint32
	broadcastLog  string
	txResult      *TxResult
	txResultErr   error
	broadcastErr  error
	accountErr    error
	getResultCall int
	resultDelay   int // how many GetTxResult calls before returning result
}

func newMockQuerier() *mockQuerier {
	return &mockQuerier{
		accountNumber: 42,
		sequence:      5,
		broadcastHash: "ABCDEF1234567890",
		broadcastCode: 0,
		txResult: &TxResult{
			TxHash:    "ABCDEF1234567890",
			Height:    1000,
			Code:      0,
			GasUsed:   150000,
			GasWanted: 200000,
		},
		resultDelay: 0,
	}
}

func (m *mockQuerier) GetAccountInfo(ctx context.Context, address string) (uint64, uint64, error) {
	if m.accountErr != nil {
		return 0, 0, m.accountErr
	}
	return m.accountNumber, m.sequence, nil
}

func (m *mockQuerier) BroadcastTxSync(ctx context.Context, txBytes []byte) (string, uint32, string, error) {
	if m.broadcastErr != nil {
		return "", 0, "", m.broadcastErr
	}
	return m.broadcastHash, m.broadcastCode, m.broadcastLog, nil
}

func (m *mockQuerier) GetTxResult(ctx context.Context, txHash string) (*TxResult, error) {
	m.getResultCall++
	if m.getResultCall <= m.resultDelay {
		return nil, fmt.Errorf("tx not found yet")
	}
	if m.txResultErr != nil {
		return nil, m.txResultErr
	}
	return m.txResult, nil
}

func TestExecuteTx_Success(t *testing.T) {
	signer := newMockSigner()
	querier := newMockQuerier()

	cfg := PipelineConfig{
		ChainID:        "seocheon-test-1",
		GasPrice:       250,
		ConfirmTimeout: 5 * time.Second,
		PollInterval:   100 * time.Millisecond,
	}

	req := TxRequest{
		Message: &MsgSubmitActivity{
			Submitter:    signer.GetAddress(),
			ActivityHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			ContentURI:   "ipfs://QmTest",
		},
	}

	result, err := ExecuteTx(context.Background(), querier, signer, cfg, req)
	if err != nil {
		t.Fatalf("ExecuteTx() error = %v", err)
	}

	if result.TxHash != "ABCDEF1234567890" {
		t.Errorf("TxHash = %s, want ABCDEF1234567890", result.TxHash)
	}
	if result.Height != 1000 {
		t.Errorf("Height = %d, want 1000", result.Height)
	}
	if result.Code != 0 {
		t.Errorf("Code = %d, want 0", result.Code)
	}
}

func TestExecuteTx_WithPolling(t *testing.T) {
	signer := newMockSigner()
	querier := newMockQuerier()
	querier.resultDelay = 2 // TX not found for first 2 polls

	cfg := PipelineConfig{
		ChainID:        "seocheon-test-1",
		GasPrice:       250,
		ConfirmTimeout: 5 * time.Second,
		PollInterval:   50 * time.Millisecond,
	}

	req := TxRequest{
		Message: &MsgSend{
			FromAddress: signer.GetAddress(),
			ToAddress:   "seocheon1recipient",
			Amount:      []Coin{{Denom: "uppyeo", Amount: "1000000"}},
		},
	}

	result, err := ExecuteTx(context.Background(), querier, signer, cfg, req)
	if err != nil {
		t.Fatalf("ExecuteTx() error = %v", err)
	}

	if result.TxHash != "ABCDEF1234567890" {
		t.Errorf("TxHash = %s, want ABCDEF1234567890", result.TxHash)
	}
	if querier.getResultCall < 3 {
		t.Errorf("expected at least 3 GetTxResult calls, got %d", querier.getResultCall)
	}
}

func TestExecuteTx_BroadcastFailure(t *testing.T) {
	signer := newMockSigner()
	querier := newMockQuerier()
	querier.broadcastCode = 4
	querier.broadcastLog = "signature verification failed"

	cfg := DefaultPipelineConfig("seocheon-test-1")

	req := TxRequest{
		Message: &MsgSubmitActivity{
			Submitter:    signer.GetAddress(),
			ActivityHash: "deadbeef",
			ContentURI:   "https://example.com",
		},
	}

	result, err := ExecuteTx(context.Background(), querier, signer, cfg, req)
	if err == nil {
		t.Fatal("expected error for broadcast failure")
	}
	if result == nil {
		t.Fatal("expected partial result with TX hash")
	}
	if result.Code != 4 {
		t.Errorf("Code = %d, want 4", result.Code)
	}
}

func TestExecuteTx_AccountError(t *testing.T) {
	signer := newMockSigner()
	querier := newMockQuerier()
	querier.accountErr = fmt.Errorf("account not found")

	cfg := DefaultPipelineConfig("seocheon-test-1")

	req := TxRequest{
		Message: &MsgSubmitActivity{
			Submitter:    signer.GetAddress(),
			ActivityHash: "deadbeef",
			ContentURI:   "https://example.com",
		},
	}

	_, err := ExecuteTx(context.Background(), querier, signer, cfg, req)
	if err == nil {
		t.Fatal("expected error for account query failure")
	}
}

func TestExecuteTx_SigningError(t *testing.T) {
	signer := newMockSigner()
	signer.signFn = func(data []byte) ([]byte, error) {
		return nil, fmt.Errorf("signing failed")
	}
	querier := newMockQuerier()

	cfg := DefaultPipelineConfig("seocheon-test-1")

	req := TxRequest{
		Message: &MsgWithdrawNodeCommission{
			Operator: signer.GetAddress(),
		},
	}

	_, err := ExecuteTx(context.Background(), querier, signer, cfg, req)
	if err == nil {
		t.Fatal("expected error for signing failure")
	}
}

func TestExecuteTx_CustomGasAndFee(t *testing.T) {
	signer := newMockSigner()
	querier := newMockQuerier()

	cfg := DefaultPipelineConfig("seocheon-test-1")

	req := TxRequest{
		Message: &MsgSubmitActivity{
			Submitter:    signer.GetAddress(),
			ActivityHash: "deadbeef",
			ContentURI:   "https://example.com",
		},
		GasLimit:  500000,
		FeeAmount: 100000,
		FeeDenom:  "uppyeo",
		Memo:      "test memo",
	}

	result, err := ExecuteTx(context.Background(), querier, signer, cfg, req)
	if err != nil {
		t.Fatalf("ExecuteTx() error = %v", err)
	}
	if result.TxHash != "ABCDEF1234567890" {
		t.Errorf("TxHash = %s, want ABCDEF1234567890", result.TxHash)
	}
}

func TestDefaultGasForMessage(t *testing.T) {
	tests := []struct {
		typeURL  string
		expected uint64
	}{
		{"/seocheon.activity.v1.MsgSubmitActivity", 200000},
		{"/seocheon.node.v1.MsgWithdrawNodeCommission", 300000},
		{"/cosmos.bank.v1beta1.MsgSend", 100000},
		{"/unknown.Message", 200000},
	}

	for _, tt := range tests {
		got := DefaultGasForMessage(tt.typeURL)
		if got != tt.expected {
			t.Errorf("DefaultGasForMessage(%s) = %d, want %d", tt.typeURL, got, tt.expected)
		}
	}
}

func TestCalculateFee(t *testing.T) {
	fee := CalculateFee(200000, 250)
	if fee != 50000000 {
		t.Errorf("CalculateFee(200000, 250) = %d, want 50000000", fee)
	}
}

func TestCalculateFeeZeroInputs(t *testing.T) {
	// Cat 13.6: zero gas or zero price should produce zero fee
	if CalculateFee(0, 250) != 0 {
		t.Errorf("CalculateFee(0, 250) = %d, want 0", CalculateFee(0, 250))
	}
	if CalculateFee(200000, 0) != 0 {
		t.Errorf("CalculateFee(200000, 0) = %d, want 0", CalculateFee(200000, 0))
	}
	if CalculateFee(0, 0) != 0 {
		t.Errorf("CalculateFee(0, 0) = %d, want 0", CalculateFee(0, 0))
	}
}

func TestPollTxConfirmation_Timeout(t *testing.T) {
	querier := newMockQuerier()
	querier.txResultErr = fmt.Errorf("tx not found")

	_, err := PollTxConfirmation(context.Background(), querier, "DEADBEEF", 200*time.Millisecond, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

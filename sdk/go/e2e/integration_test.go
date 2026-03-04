//go:build e2e

package e2e_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"
	"time"

	sdk "github.com/seocheon/sdk-go"
)

// skipIfMissing skips the test if the environment variable is not set and returns its value.
func skipIfMissing(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skipf("E2E 스킵: 환경변수 %s 미설정", key)
	}
	return v
}

// e2eSDK builds a connected SDK instance from environment variables.
// Requires SEOCHEON_GRPC and SEOCHEON_MNEMONIC; uses defaults for RPC and CHAIN_ID.
func e2eSDK(t *testing.T) *sdk.SeocheonSDK {
	t.Helper()

	rpc := os.Getenv("SEOCHEON_RPC")
	if rpc == "" {
		rpc = "http://localhost:26657"
	}
	grpc := skipIfMissing(t, "SEOCHEON_GRPC")
	chainID := os.Getenv("SEOCHEON_CHAIN_ID")
	if chainID == "" {
		chainID = "seocheon-e2e"
	}
	mnemonic := skipIfMissing(t, "SEOCHEON_MNEMONIC")

	cfg := sdk.DefaultConfig(chainID, rpc, grpc, sdk.SigningConfig{
		Mode:     sdk.SigningModeDirect,
		Mnemonic: mnemonic,
	})

	s, err := sdk.New(cfg)
	if err != nil {
		t.Fatalf("SDK 생성 실패: %v", err)
	}
	return s
}

// TestConnect checks that Connect() succeeds and IsConnected() returns true.
func TestConnect(t *testing.T) {
	s := e2eSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}
	defer func() { _ = s.Disconnect() }()

	if !s.IsConnected() {
		t.Fatal("Connect 후 IsConnected() = false")
	}
	t.Log("Connect 성공")
}

// TestGetLatestBlock verifies that Cosmos.GetBlockInfo returns a valid block.
func TestGetLatestBlock(t *testing.T) {
	s := e2eSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}
	defer func() { _ = s.Disconnect() }()

	block, err := s.Cosmos.GetBlockInfo(ctx)
	if err != nil {
		t.Fatalf("Cosmos.GetBlockInfo 실패: %v", err)
	}
	if block.BlockHeight <= 0 {
		t.Fatalf("블록 높이가 양수여야 함, got %d", block.BlockHeight)
	}
	t.Logf("최신 블록: height=%d chainID=%s", block.BlockHeight, block.ChainID)
}

// TestQueryNodeModule verifies that x/node REST endpoint is reachable via Node.Search.
func TestQueryNodeModule(t *testing.T) {
	s := e2eSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}
	defer func() { _ = s.Disconnect() }()

	resp, err := s.Node.Search(ctx, "", "", 10, "asc")
	if err != nil {
		t.Fatalf("Node.Search 실패: %v", err)
	}
	t.Logf("x/node 조회 성공: total=%d", resp.TotalCount)
}

// TestQueryEpochInfo verifies that Epoch.GetInfo returns valid epoch data
// (internally queries /seocheon/activity/v1/params and /seocheon/activity/v1/epoch-info).
func TestQueryEpochInfo(t *testing.T) {
	s := e2eSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}
	defer func() { _ = s.Disconnect() }()

	info, err := s.Epoch.GetInfo(ctx)
	if err != nil {
		t.Fatalf("Epoch.GetInfo 실패: %v", err)
	}
	if info.BlockHeight <= 0 {
		t.Fatalf("에포크 블록 높이가 양수여야 함, got %d", info.BlockHeight)
	}
	t.Logf("에포크 정보: epoch=%d window=%d height=%d",
		info.EpochNumber, info.WindowNumber, info.BlockHeight)
}

// TestQueryBalance verifies that Cosmos.GetBalance returns the agent's balance.
func TestQueryBalance(t *testing.T) {
	s := e2eSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}
	defer func() { _ = s.Disconnect() }()

	resp, err := s.Cosmos.GetBalance(ctx, "", "uppyeo")
	if err != nil {
		t.Fatalf("Cosmos.GetBalance 실패: %v", err)
	}
	t.Logf("잔액: %s uppyeo (%s kkot) 주소=%s", resp.Balance, resp.BalanceKkot, resp.Address)
}

// TestSubmitActivity submits a MsgSubmitActivity and verifies the TxHash is returned.
// This test requires the signer to be a registered node's agent address.
func TestSubmitActivity(t *testing.T) {
	s := e2eSDK(t)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := s.Connect(ctx); err != nil {
		t.Fatalf("Connect 실패: %v", err)
	}
	defer func() { _ = s.Disconnect() }()

	// Deterministic but unique hash per test run
	h := sha256.Sum256([]byte(fmt.Sprintf("e2e-activity-%d", time.Now().UnixNano())))
	activityHash := fmt.Sprintf("%x", h)
	contentURI := "https://example.com/e2e-activity"

	resp, err := s.Activity.Submit(ctx, activityHash, contentURI)
	if err != nil {
		t.Fatalf("Activity.Submit 실패: %v", err)
	}
	if resp.TxHash == "" {
		t.Fatal("TxHash가 비어 있어서는 안 됨")
	}
	t.Logf("활동 제출 성공: txhash=%s height=%d epoch=%d window=%d quota_remaining=%d",
		resp.TxHash, resp.BlockHeight, resp.EpochNumber, resp.WindowNumber, resp.QuotaRemaining)
}

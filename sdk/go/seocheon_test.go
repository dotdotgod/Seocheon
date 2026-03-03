package sdk

import (
	"testing"
)

// Cat 03: Client lifecycle tests (5 tests)

func TestCreateSdkWithValidConfig(t *testing.T) {
	sdk, err := New(SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if sdk == nil {
		t.Fatal("New() returned nil SDK")
	}
}

func TestStartsDisconnected(t *testing.T) {
	sdk, err := New(SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if sdk.IsConnected() {
		t.Error("SDK should start disconnected")
	}
}

func TestAllFiveModulesInitialized(t *testing.T) {
	sdk, err := New(SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if sdk.Activity == nil {
		t.Error("Activity module is nil")
	}
	if sdk.Node == nil {
		t.Error("Node module is nil")
	}
	if sdk.Epoch == nil {
		t.Error("Epoch module is nil")
	}
	if sdk.Rewards == nil {
		t.Error("Rewards module is nil")
	}
	if sdk.Cosmos == nil {
		t.Error("Cosmos module is nil")
	}
}

func TestConnectAndDisconnect(t *testing.T) {
	sdk, err := New(SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Should start disconnected
	if sdk.IsConnected() {
		t.Error("should start disconnected")
	}

	// Disconnect when not connected should not error
	if err := sdk.Disconnect(); err != nil {
		t.Errorf("Disconnect() error = %v", err)
	}
	if sdk.IsConnected() {
		t.Error("should be disconnected after Disconnect()")
	}
}

func TestInvalidConfigReturnsError(t *testing.T) {
	// Missing chain_id
	_, err := New(SDKConfig{
		Chain: ChainConfig{
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "test mnemonic here",
		},
	})
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}

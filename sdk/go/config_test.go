package sdk

import (
	"testing"
)

// Cat 02: Config validation tests (8 tests)

func TestValidDirectConfigPasses(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}

func TestMissingChainIdFails(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "test mnemonic here",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing chain_id")
	}
}

func TestMissingRpcEndpointFails(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "test mnemonic here",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing rpc_endpoint")
	}
}

func TestMissingGrpcEndpointFails(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			ChainID:     "seocheon-test-1",
			RPCEndpoint: "http://localhost:26657",
		},
		Signing: SigningConfig{
			Mode:     SigningModeDirect,
			Mnemonic: "test mnemonic here",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for missing grpc_endpoint")
	}
}

func TestInvalidSigningModeFails(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode: "invalid-mode",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for invalid signing mode")
	}
}

func TestDirectRequiresMnemonic(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode: SigningModeDirect,
			// no mnemonic
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when direct mode has no mnemonic")
	}
}

func TestVaultRequiresEndpointAndKey(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode: SigningModeVault,
			// no vault_endpoint or key_name
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when vault mode has no endpoint/key")
	}
}

func TestKeystoreRequiresPathAndPass(t *testing.T) {
	cfg := SDKConfig{
		Chain: ChainConfig{
			ChainID:      "seocheon-test-1",
			RPCEndpoint:  "http://localhost:26657",
			GRPCEndpoint: "http://localhost:1317",
		},
		Signing: SigningConfig{
			Mode: SigningModeKeystore,
			// no keystore_path or passphrase
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error when keystore mode has no path/passphrase")
	}
}

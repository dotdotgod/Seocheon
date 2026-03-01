// Package sdk provides the Seocheon blockchain client SDK for Go.
package sdk

import (
	"fmt"

	"github.com/seocheon/sdk-go/constants"
)

// ChainConfig holds blockchain connection settings.
type ChainConfig struct {
	ChainID       string  `json:"chain_id"`
	RPCEndpoint   string  `json:"rpc_endpoint"`
	GRPCEndpoint  string  `json:"grpc_endpoint"`
	GasPrice      string  `json:"gas_price"`
	GasAdjustment float64 `json:"gas_adjustment"`
}

// SigningMode specifies how transactions are signed.
type SigningMode string

const (
	SigningModeVault    SigningMode = "vault"
	SigningModeKeystore SigningMode = "keystore"
	SigningModeDirect   SigningMode = "direct"
)

// SigningConfig holds transaction signing settings.
type SigningConfig struct {
	Mode           SigningMode `json:"mode"`
	VaultEndpoint  string      `json:"vault_endpoint,omitempty"`
	KeyName        string      `json:"key_name,omitempty"`
	KeystorePath   string      `json:"keystore_path,omitempty"`
	PassphraseEnv  string      `json:"passphrase_env,omitempty"`
	Mnemonic       string      `json:"mnemonic,omitempty"`
}

// TxConfig holds transaction broadcast settings.
type TxConfig struct {
	BroadcastMode       string `json:"broadcast_mode"`
	ConfirmTimeoutMs    uint64 `json:"confirm_timeout_ms"`
	ConfirmPollInterval uint64 `json:"confirm_poll_interval_ms"`
}

// SDKConfig is the top-level configuration for SeocheonSDK.
type SDKConfig struct {
	Chain   ChainConfig   `json:"chain"`
	Signing SigningConfig `json:"signing"`
	Tx      TxConfig      `json:"tx"`
}

// Option is a functional option for configuring the SDK.
type Option func(*SDKConfig)

// WithGasPrice sets a custom gas price.
func WithGasPrice(price string) Option {
	return func(c *SDKConfig) {
		c.Chain.GasPrice = price
	}
}

// WithGasAdjustment sets a custom gas adjustment factor.
func WithGasAdjustment(adj float64) Option {
	return func(c *SDKConfig) {
		c.Chain.GasAdjustment = adj
	}
}

// WithBroadcastMode sets the TX broadcast mode ("sync" or "async").
func WithBroadcastMode(mode string) Option {
	return func(c *SDKConfig) {
		c.Tx.BroadcastMode = mode
	}
}

// WithConfirmTimeout sets the TX confirmation timeout in milliseconds.
func WithConfirmTimeout(ms uint64) Option {
	return func(c *SDKConfig) {
		c.Tx.ConfirmTimeoutMs = ms
	}
}

// WithConfirmPollInterval sets the TX confirmation polling interval in milliseconds.
func WithConfirmPollInterval(ms uint64) Option {
	return func(c *SDKConfig) {
		c.Tx.ConfirmPollInterval = ms
	}
}

// DefaultConfig returns an SDKConfig with default values applied.
func DefaultConfig(chainID, rpcEndpoint, grpcEndpoint string, signing SigningConfig) SDKConfig {
	return SDKConfig{
		Chain: ChainConfig{
			ChainID:       chainID,
			RPCEndpoint:   rpcEndpoint,
			GRPCEndpoint:  grpcEndpoint,
			GasPrice:      constants.DefaultGasPrice,
			GasAdjustment: constants.DefaultGasAdjustment,
		},
		Signing: signing,
		Tx: TxConfig{
			BroadcastMode:       constants.DefaultBroadcastMode,
			ConfirmTimeoutMs:    constants.DefaultConfirmTimeoutMs,
			ConfirmPollInterval: constants.DefaultConfirmPollMs,
		},
	}
}

// Validate checks that the SDKConfig is valid.
func (c *SDKConfig) Validate() error {
	if c.Chain.ChainID == "" {
		return fmt.Errorf("chain_id is required")
	}
	if c.Chain.RPCEndpoint == "" {
		return fmt.Errorf("rpc_endpoint is required")
	}
	if c.Chain.GRPCEndpoint == "" {
		return fmt.Errorf("grpc_endpoint is required")
	}
	switch c.Signing.Mode {
	case SigningModeVault:
		if c.Signing.VaultEndpoint == "" || c.Signing.KeyName == "" {
			return fmt.Errorf("vault mode requires vault_endpoint and key_name")
		}
	case SigningModeKeystore:
		if c.Signing.KeystorePath == "" || c.Signing.PassphraseEnv == "" {
			return fmt.Errorf("keystore mode requires keystore_path and passphrase_env")
		}
	case SigningModeDirect:
		if c.Signing.Mnemonic == "" {
			return fmt.Errorf("direct mode requires mnemonic")
		}
	default:
		return fmt.Errorf("invalid signing mode: %s", c.Signing.Mode)
	}
	return nil
}

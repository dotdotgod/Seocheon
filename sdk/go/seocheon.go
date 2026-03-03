package sdk

import (
	"context"
	"fmt"

	"github.com/seocheon/sdk-go/activity"
	"github.com/seocheon/sdk-go/cosmos"
	"github.com/seocheon/sdk-go/epoch"
	sdkerrors "github.com/seocheon/sdk-go/errors"
	"github.com/seocheon/sdk-go/internal/chain"
	"github.com/seocheon/sdk-go/internal/signing"
	"github.com/seocheon/sdk-go/node"
	"github.com/seocheon/sdk-go/rewards"
	"github.com/seocheon/sdk-go/subscription"
)

// SeocheonSDK is the main entry point for the Seocheon blockchain SDK.
type SeocheonSDK struct {
	config  SDKConfig
	client  chain.Client
	signer  signing.Service
	connected bool

	// Modules
	Activity     *activity.Module
	Epoch        *epoch.Module
	Node         *node.Module
	Rewards      *rewards.Module
	Cosmos       *cosmos.Module
	Subscription *subscription.Manager
}

// New creates a new SeocheonSDK instance with the given config and options.
func New(config SDKConfig, opts ...Option) (*SeocheonSDK, error) {
	for _, opt := range opts {
		opt(&config)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrInvalidConfig, err)
	}

	client := chain.NewHTTPClient(config.Chain.RPCEndpoint, config.Chain.GRPCEndpoint)

	signer, err := createSigner(config.Signing)
	if err != nil {
		return nil, fmt.Errorf("creating signing service: %w", err)
	}

	sdk := &SeocheonSDK{
		config: config,
		client: client,
		signer: signer,
	}

	sdk.Activity = activity.NewModule(client, signer, config.Chain.ChainID)
	sdk.Epoch = epoch.NewModule(client)
	sdk.Node = node.NewModule(client, signer)
	sdk.Rewards = rewards.NewModule(client, signer, config.Chain.ChainID)
	sdk.Cosmos = cosmos.NewModule(client, signer, config.Chain.ChainID)
	sdk.Subscription = subscription.NewManager(config.Chain.RPCEndpoint)

	return sdk, nil
}

// Connect establishes a connection to the blockchain node.
func (s *SeocheonSDK) Connect(ctx context.Context) error {
	if err := s.client.Connect(ctx); err != nil {
		return fmt.Errorf("connecting to chain: %w", err)
	}
	s.connected = true
	return nil
}

// Disconnect closes the connection to the blockchain node.
func (s *SeocheonSDK) Disconnect() error {
	s.connected = false
	if s.Subscription != nil {
		s.Subscription.UnsubscribeAll()
	}
	return s.client.Disconnect()
}

// IsConnected returns whether the SDK is connected to the chain.
func (s *SeocheonSDK) IsConnected() bool {
	return s.connected
}

// GetConfig returns the current SDK configuration.
func (s *SeocheonSDK) GetConfig() SDKConfig {
	return s.config
}

// createSigner creates the appropriate signing service based on config.
func createSigner(cfg SigningConfig) (signing.Service, error) {
	switch cfg.Mode {
	case SigningModeVault:
		return signing.NewVaultService(cfg.VaultEndpoint, cfg.KeyName)
	case SigningModeKeystore:
		return signing.NewKeystoreService(cfg.KeystorePath, cfg.PassphraseEnv)
	case SigningModeDirect:
		return signing.NewDirectService(cfg.Mnemonic)
	default:
		return nil, fmt.Errorf("unsupported signing mode: %s", cfg.Mode)
	}
}

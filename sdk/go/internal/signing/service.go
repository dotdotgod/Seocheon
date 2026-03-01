// Package signing provides the SigningService interface and implementations
// for transaction signing in the Seocheon SDK.
package signing

import (
	"fmt"
)

// Service defines the interface for transaction signing.
type Service interface {
	// Sign signs the given transaction bytes and returns the signature.
	Sign(txBytes []byte) ([]byte, error)
	// GetAddress returns the signer's address.
	GetAddress() string
	// GetPubKey returns the signer's public key bytes.
	GetPubKey() ([]byte, error)
}

// VaultService signs transactions via an external vault server (production).
type VaultService struct {
	endpoint string
	keyName  string
	address  string
}

// NewVaultService creates a new VaultService.
func NewVaultService(endpoint, keyName string) (*VaultService, error) {
	if endpoint == "" || keyName == "" {
		return nil, fmt.Errorf("vault endpoint and key name are required")
	}
	return &VaultService{
		endpoint: endpoint,
		keyName:  keyName,
	}, nil
}

func (v *VaultService) Sign(txBytes []byte) ([]byte, error) {
	// TODO: Implement vault signing via HTTP call to vault server
	return nil, fmt.Errorf("vault signing not yet implemented")
}

func (v *VaultService) GetAddress() string {
	return v.address
}

func (v *VaultService) GetPubKey() ([]byte, error) {
	// TODO: Implement vault pubkey retrieval
	return nil, fmt.Errorf("vault pubkey retrieval not yet implemented")
}

// KeystoreService signs transactions using a local keystore file.
type KeystoreService struct {
	keystorePath string
	passphrase   string
	address      string
}

// NewKeystoreService creates a new KeystoreService.
func NewKeystoreService(keystorePath, passphrase string) (*KeystoreService, error) {
	if keystorePath == "" || passphrase == "" {
		return nil, fmt.Errorf("keystore path and passphrase are required")
	}
	return &KeystoreService{
		keystorePath: keystorePath,
		passphrase:   passphrase,
	}, nil
}

func (k *KeystoreService) Sign(txBytes []byte) ([]byte, error) {
	// TODO: Implement keystore signing
	return nil, fmt.Errorf("keystore signing not yet implemented")
}

func (k *KeystoreService) GetAddress() string {
	return k.address
}

func (k *KeystoreService) GetPubKey() ([]byte, error) {
	// TODO: Implement keystore pubkey retrieval
	return nil, fmt.Errorf("keystore pubkey retrieval not yet implemented")
}

// DirectService signs transactions using a mnemonic directly (test only).
type DirectService struct {
	mnemonic string
	address  string
	pubKey   []byte
}

// NewDirectService creates a new DirectService from a mnemonic.
func NewDirectService(mnemonic string) (*DirectService, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("mnemonic is required")
	}
	// TODO: Derive key pair from mnemonic using BIP44 path
	return &DirectService{
		mnemonic: mnemonic,
	}, nil
}

func (d *DirectService) Sign(txBytes []byte) ([]byte, error) {
	// TODO: Implement direct signing with derived key
	return nil, fmt.Errorf("direct signing not yet implemented")
}

func (d *DirectService) GetAddress() string {
	return d.address
}

func (d *DirectService) GetPubKey() ([]byte, error) {
	return d.pubKey, nil
}

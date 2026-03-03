// Package signing provides the SigningService interface and implementations
// for transaction signing in the Seocheon SDK.
package signing

import (
	"fmt"
	"net/http"

	"github.com/seocheon/sdk-go/internal/crypto"
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
	pubKey   []byte
	client   *http.Client
}

// NewVaultService creates a new VaultService.
// It connects to the vault server to fetch the address and public key.
func NewVaultService(endpoint, keyName string) (*VaultService, error) {
	if endpoint == "" || keyName == "" {
		return nil, fmt.Errorf("vault endpoint and key name are required")
	}
	return initVaultService(endpoint, keyName)
}

func (v *VaultService) Sign(txBytes []byte) ([]byte, error) {
	return vaultSign(v.client, v.endpoint, v.keyName, txBytes)
}

func (v *VaultService) GetAddress() string {
	return v.address
}

func (v *VaultService) GetPubKey() ([]byte, error) {
	return v.pubKey, nil
}

// KeystoreService signs transactions using a local encrypted keystore file.
type KeystoreService struct {
	privKey *crypto.PrivateKey
	address string
	pubKey  []byte
}

// NewKeystoreService creates a new KeystoreService by decrypting the keystore.
// passphrase is the passphrase or environment variable name containing it.
func NewKeystoreService(keystorePath, passphrase string) (*KeystoreService, error) {
	if keystorePath == "" || passphrase == "" {
		return nil, fmt.Errorf("keystore path and passphrase are required")
	}

	key, err := loadAndDecryptKeystore(keystorePath, passphrase)
	if err != nil {
		return nil, fmt.Errorf("loading keystore: %w", err)
	}

	pubKey := key.PubKey()
	addr, err := crypto.AddressFromPubKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("deriving address from keystore: %w", err)
	}

	return &KeystoreService{
		privKey: key,
		address: addr,
		pubKey:  pubKey,
	}, nil
}

func (k *KeystoreService) Sign(txBytes []byte) ([]byte, error) {
	return k.privKey.Sign(txBytes)
}

func (k *KeystoreService) GetAddress() string {
	return k.address
}

func (k *KeystoreService) GetPubKey() ([]byte, error) {
	return k.pubKey, nil
}

// DirectService signs transactions using a mnemonic directly (test only).
type DirectService struct {
	privKey *crypto.PrivateKey
	address string
	pubKey  []byte
}

// NewDirectService creates a new DirectService from a mnemonic.
func NewDirectService(mnemonic string) (*DirectService, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("mnemonic is required")
	}

	key, err := crypto.DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("deriving key from mnemonic: %w", err)
	}

	pubKey := key.PubKey()
	addr, err := crypto.AddressFromPubKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("deriving address: %w", err)
	}

	return &DirectService{
		privKey: key,
		address: addr,
		pubKey:  pubKey,
	}, nil
}

func (d *DirectService) Sign(txBytes []byte) ([]byte, error) {
	return d.privKey.Sign(txBytes)
}

func (d *DirectService) GetAddress() string {
	return d.address
}

func (d *DirectService) GetPubKey() ([]byte, error) {
	return d.pubKey, nil
}

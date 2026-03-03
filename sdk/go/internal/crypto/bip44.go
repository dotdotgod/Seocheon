package crypto

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/tyler-smith/go-bip39"
)

// DeriveKeyFromMnemonic derives a secp256k1 private key from a BIP39 mnemonic
// using the Cosmos BIP44 path: m/44'/118'/0'/0/0
func DeriveKeyFromMnemonic(mnemonic string) (*PrivateKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}
	seed := bip39.NewSeed(mnemonic, "")

	// Create master key
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("creating master key: %w", err)
	}

	// Derive m/44'/118'/0'/0/0 (Cosmos BIP44 path)
	// 44' - purpose
	purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, fmt.Errorf("deriving purpose: %w", err)
	}
	// 118' - coin type (Cosmos)
	coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + 118)
	if err != nil {
		return nil, fmt.Errorf("deriving coin type: %w", err)
	}
	// 0' - account
	account, err := coinType.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, fmt.Errorf("deriving account: %w", err)
	}
	// 0 - change
	change, err := account.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("deriving change: %w", err)
	}
	// 0 - address index
	addressKey, err := change.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("deriving address index: %w", err)
	}

	// Extract EC private key
	ecPrivKey, err := addressKey.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("extracting private key: %w", err)
	}

	return NewPrivateKeyFromBtcec(ecPrivKey), nil
}

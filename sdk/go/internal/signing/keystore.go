package signing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/scrypt"

	"github.com/seocheon/sdk-go/internal/crypto"
)

// keystoreFile represents the structure of an encrypted keystore file.
// Compatible with the Web3 Secret Storage format (Ethereum-derived).
type keystoreFile struct {
	Crypto keystoreCrypto `json:"crypto"`
}

type keystoreCrypto struct {
	Cipher       string               `json:"cipher"`
	CipherText   string               `json:"ciphertext"`
	CipherParams keystoreCipherParams `json:"cipherparams"`
	KDF          string               `json:"kdf"`
	KDFParams    keystoreKDFParams    `json:"kdfparams"`
	MAC          string               `json:"mac"`
}

type keystoreCipherParams struct {
	IV string `json:"iv"`
}

type keystoreKDFParams struct {
	DKLen int    `json:"dklen"`
	N     int    `json:"n"`
	R     int    `json:"r"`
	P     int    `json:"p"`
	Salt  string `json:"salt"`
}

// loadAndDecryptKeystore reads a keystore file, derives the decryption key
// via scrypt, and decrypts the private key using AES-128-CTR.
func loadAndDecryptKeystore(path, passphrase string) (*crypto.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading keystore file: %w", err)
	}

	var ks keystoreFile
	if err := json.Unmarshal(data, &ks); err != nil {
		return nil, fmt.Errorf("parsing keystore JSON: %w", err)
	}

	if ks.Crypto.KDF != "scrypt" {
		return nil, fmt.Errorf("unsupported KDF: %s (only scrypt is supported)", ks.Crypto.KDF)
	}
	if ks.Crypto.Cipher != "aes-128-ctr" {
		return nil, fmt.Errorf("unsupported cipher: %s (only aes-128-ctr is supported)", ks.Crypto.Cipher)
	}

	// Decode hex values
	salt, err := hex.DecodeString(ks.Crypto.KDFParams.Salt)
	if err != nil {
		return nil, fmt.Errorf("decoding salt: %w", err)
	}
	iv, err := hex.DecodeString(ks.Crypto.CipherParams.IV)
	if err != nil {
		return nil, fmt.Errorf("decoding IV: %w", err)
	}
	cipherText, err := hex.DecodeString(ks.Crypto.CipherText)
	if err != nil {
		return nil, fmt.Errorf("decoding ciphertext: %w", err)
	}
	mac, err := hex.DecodeString(ks.Crypto.MAC)
	if err != nil {
		return nil, fmt.Errorf("decoding MAC: %w", err)
	}

	// Derive key using scrypt
	dkLen := ks.Crypto.KDFParams.DKLen
	if dkLen == 0 {
		dkLen = 32
	}
	derivedKey, err := scrypt.Key([]byte(passphrase), salt, ks.Crypto.KDFParams.N, ks.Crypto.KDFParams.R, ks.Crypto.KDFParams.P, dkLen)
	if err != nil {
		return nil, fmt.Errorf("deriving key via scrypt: %w", err)
	}

	// Verify MAC: SHA256(derivedKey[16:32] + cipherText)
	macInput := append(derivedKey[16:32], cipherText...)
	calculatedMAC := sha256.Sum256(macInput)
	if !equalBytes(calculatedMAC[:], mac) {
		return nil, fmt.Errorf("MAC verification failed: incorrect passphrase or corrupted keystore")
	}

	// Decrypt using AES-128-CTR
	aesKey := derivedKey[:16]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	stream := cipher.NewCTR(block, iv)
	privKeyBytes := make([]byte, len(cipherText))
	stream.XORKeyStream(privKeyBytes, cipherText)

	return crypto.NewPrivateKeyFromBytes(privKeyBytes)
}

// equalBytes compares two byte slices in constant-ish time.
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := range a {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

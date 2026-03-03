package signing

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/scrypt"

	"github.com/seocheon/sdk-go/internal/crypto"
)

// createTestKeystore creates a temporary keystore file for testing.
func createTestKeystore(t *testing.T, privKeyBytes []byte, passphrase string) string {
	t.Helper()

	// Derive key using scrypt
	salt := []byte("testsalt12345678")
	n, r, p, dkLen := 8192, 8, 1, 32
	derivedKey, err := scrypt.Key([]byte(passphrase), salt, n, r, p, dkLen)
	if err != nil {
		t.Fatalf("scrypt: %v", err)
	}

	// Encrypt with AES-128-CTR
	iv := []byte("testiv1234567890") // 16 bytes
	aesKey := derivedKey[:16]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		t.Fatalf("aes: %v", err)
	}

	cipherText := make([]byte, len(privKeyBytes))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(cipherText, privKeyBytes)

	// Compute MAC: SHA256(derivedKey[16:32] + cipherText)
	macInput := append(derivedKey[16:32], cipherText...)
	mac := sha256.Sum256(macInput)

	ks := keystoreFile{
		Crypto: keystoreCrypto{
			Cipher:     "aes-128-ctr",
			CipherText: hex.EncodeToString(cipherText),
			CipherParams: keystoreCipherParams{
				IV: hex.EncodeToString(iv),
			},
			KDF: "scrypt",
			KDFParams: keystoreKDFParams{
				DKLen: dkLen,
				N:     n,
				R:     r,
				P:     p,
				Salt:  hex.EncodeToString(salt),
			},
			MAC: hex.EncodeToString(mac[:]),
		},
	}

	data, err := json.Marshal(ks)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "keystore.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("write: %v", err)
	}

	return path
}

func TestKeystoreService_LoadAndSign(t *testing.T) {
	// Generate a known key from mnemonic
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	origKey, err := crypto.DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("DeriveKeyFromMnemonic: %v", err)
	}

	passphrase := "testpassword123"
	path := createTestKeystore(t, origKey.Bytes(), passphrase)

	// Load keystore
	svc, err := NewKeystoreService(path, passphrase)
	if err != nil {
		t.Fatalf("NewKeystoreService: %v", err)
	}

	// Verify address matches
	origAddr, _ := crypto.AddressFromPubKey(origKey.PubKey())
	if svc.GetAddress() != origAddr {
		t.Errorf("address = %s, want %s", svc.GetAddress(), origAddr)
	}

	// Verify signing works
	testData := []byte("test data for keystore signing")
	sig, err := svc.Sign(testData)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sig) != 64 {
		t.Errorf("signature length = %d, want 64", len(sig))
	}

	// Verify signature
	pubKey, _ := svc.GetPubKey()
	if !crypto.Verify(pubKey, testData, sig) {
		t.Error("signature verification failed")
	}
}

func TestKeystoreService_WrongPassphrase(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	origKey, _ := crypto.DeriveKeyFromMnemonic(mnemonic)

	path := createTestKeystore(t, origKey.Bytes(), "correct_password")

	_, err := NewKeystoreService(path, "wrong_password")
	if err == nil {
		t.Fatal("expected error for wrong passphrase")
	}
}

func TestKeystoreService_MissingFile(t *testing.T) {
	_, err := NewKeystoreService("/nonexistent/keystore.json", "password")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestKeystoreService_EmptyParams(t *testing.T) {
	_, err := NewKeystoreService("", "password")
	if err == nil {
		t.Fatal("expected error for empty path")
	}

	_, err = NewKeystoreService("/some/path", "")
	if err == nil {
		t.Fatal("expected error for empty passphrase")
	}
}

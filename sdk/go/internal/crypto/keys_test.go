package crypto

import (
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	key, err := DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("DeriveKeyFromMnemonic() error = %v", err)
	}

	testData := []byte("test transaction data for signing")

	sig, err := key.Sign(testData)
	if err != nil {
		t.Fatalf("Sign() error = %v", err)
	}

	// Signature should be 64 bytes
	if len(sig) != 64 {
		t.Errorf("signature length = %d, want 64", len(sig))
	}

	// Verify should succeed
	if !Verify(key.PubKey(), testData, sig) {
		t.Error("Verify() returned false for valid signature")
	}

	// Verify with wrong data should fail
	if Verify(key.PubKey(), []byte("wrong data"), sig) {
		t.Error("Verify() returned true for wrong data")
	}

	// Verify with tampered signature should fail
	tamperedSig := make([]byte, 64)
	copy(tamperedSig, sig)
	tamperedSig[0] ^= 0xff
	if Verify(key.PubKey(), testData, tamperedSig) {
		t.Error("Verify() returned true for tampered signature")
	}
}

func TestNewPrivateKeyFromBytes(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	key, err := DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("DeriveKeyFromMnemonic() error = %v", err)
	}

	// Reconstruct from bytes
	key2, err := NewPrivateKeyFromBytes(key.Bytes())
	if err != nil {
		t.Fatalf("NewPrivateKeyFromBytes() error = %v", err)
	}

	// Should produce same public key
	pub1 := key.PubKey()
	pub2 := key2.PubKey()

	if len(pub1) != len(pub2) {
		t.Fatalf("public key lengths differ: %d vs %d", len(pub1), len(pub2))
	}

	for i := range pub1 {
		if pub1[i] != pub2[i] {
			t.Errorf("public key byte %d differs: %02x vs %02x", i, pub1[i], pub2[i])
		}
	}
}

func TestNewPrivateKeyFromBytes_InvalidLength(t *testing.T) {
	_, err := NewPrivateKeyFromBytes([]byte{1, 2, 3})
	if err == nil {
		t.Error("expected error for invalid key length")
	}
}

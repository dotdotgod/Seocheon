package crypto

import (
	"testing"
)

func TestDeriveKeyFromMnemonic(t *testing.T) {
	// Well-known BIP39 test mnemonic (12 words)
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	key, err := DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("DeriveKeyFromMnemonic() error = %v", err)
	}

	// Verify private key is 32 bytes
	if len(key.Bytes()) != 32 {
		t.Errorf("private key length = %d, want 32", len(key.Bytes()))
	}

	// Verify public key is 33 bytes (compressed)
	pubKey := key.PubKey()
	if len(pubKey) != 33 {
		t.Errorf("public key length = %d, want 33", len(pubKey))
	}

	// Derive address
	addr, err := AddressFromPubKey(pubKey)
	if err != nil {
		t.Fatalf("AddressFromPubKey() error = %v", err)
	}

	// Address should start with "seocheon1"
	if len(addr) < 10 || addr[:9] != "seocheon1" {
		t.Errorf("address = %s, want prefix 'seocheon1'", addr)
	}

	t.Logf("Derived address: %s", addr)

	// Same mnemonic should always derive the same key
	key2, err := DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("second DeriveKeyFromMnemonic() error = %v", err)
	}

	addr2, err := AddressFromPubKey(key2.PubKey())
	if err != nil {
		t.Fatalf("second AddressFromPubKey() error = %v", err)
	}

	if addr != addr2 {
		t.Errorf("determinism: addr1=%s, addr2=%s", addr, addr2)
	}
}

func TestDeriveKeyFromMnemonic_Invalid(t *testing.T) {
	_, err := DeriveKeyFromMnemonic("invalid mnemonic phrase that is not valid")
	if err == nil {
		t.Error("expected error for invalid mnemonic, got nil")
	}
}

func TestDeriveKeyFromMnemonic_24Words(t *testing.T) {
	// 24-word mnemonic test
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"

	key, err := DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("DeriveKeyFromMnemonic() error = %v", err)
	}

	addr, err := AddressFromPubKey(key.PubKey())
	if err != nil {
		t.Fatalf("AddressFromPubKey() error = %v", err)
	}

	if len(addr) < 10 || addr[:9] != "seocheon1" {
		t.Errorf("address = %s, want prefix 'seocheon1'", addr)
	}

	t.Logf("24-word derived address: %s", addr)
}

func TestDeriveDifferentMnemonic(t *testing.T) {
	mnemonic1 := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	mnemonic2 := "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"

	key1, err := DeriveKeyFromMnemonic(mnemonic1)
	if err != nil {
		t.Fatalf("first DeriveKeyFromMnemonic() error = %v", err)
	}
	key2, err := DeriveKeyFromMnemonic(mnemonic2)
	if err != nil {
		t.Fatalf("second DeriveKeyFromMnemonic() error = %v", err)
	}

	addr1, _ := AddressFromPubKey(key1.PubKey())
	addr2, _ := AddressFromPubKey(key2.PubKey())

	if addr1 == addr2 {
		t.Error("different mnemonics should produce different addresses")
	}

	// Public keys should also differ
	pub1 := key1.PubKey()
	pub2 := key2.PubKey()
	same := true
	for i := range pub1 {
		if pub1[i] != pub2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("different mnemonics should produce different public keys")
	}
}

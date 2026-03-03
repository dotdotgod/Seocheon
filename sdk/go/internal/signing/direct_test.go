package signing

import (
	"strings"
	"testing"
)

// Cat 15: Direct signing tests (5 tests)

const testMnemonic12 = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
const testMnemonic12Alt = "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"

func TestDirectAddressHasPrefix(t *testing.T) {
	svc, err := NewDirectService(testMnemonic12)
	if err != nil {
		t.Fatalf("NewDirectService: %v", err)
	}

	addr := svc.GetAddress()
	if !strings.HasPrefix(addr, "seocheon1") {
		t.Errorf("address = %s, want prefix 'seocheon1'", addr)
	}
}

func TestDirectPubkey33Compressed(t *testing.T) {
	svc, err := NewDirectService(testMnemonic12)
	if err != nil {
		t.Fatalf("NewDirectService: %v", err)
	}

	pubKey, err := svc.GetPubKey()
	if err != nil {
		t.Fatalf("GetPubKey: %v", err)
	}
	if len(pubKey) != 33 {
		t.Errorf("pubkey length = %d, want 33 (compressed)", len(pubKey))
	}
	// First byte should be 0x02 or 0x03 (compressed point prefix)
	if pubKey[0] != 0x02 && pubKey[0] != 0x03 {
		t.Errorf("pubkey first byte = 0x%02x, want 0x02 or 0x03", pubKey[0])
	}
}

func TestDirectDeterministicAddress(t *testing.T) {
	svc1, err := NewDirectService(testMnemonic12)
	if err != nil {
		t.Fatalf("first NewDirectService: %v", err)
	}

	svc2, err := NewDirectService(testMnemonic12)
	if err != nil {
		t.Fatalf("second NewDirectService: %v", err)
	}

	if svc1.GetAddress() != svc2.GetAddress() {
		t.Errorf("same mnemonic produced different addresses: %s != %s",
			svc1.GetAddress(), svc2.GetAddress())
	}
}

func TestDirectSign64Bytes(t *testing.T) {
	svc, err := NewDirectService(testMnemonic12)
	if err != nil {
		t.Fatalf("NewDirectService: %v", err)
	}

	sig, err := svc.Sign([]byte("test transaction data"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sig) != 64 {
		t.Errorf("signature length = %d, want 64", len(sig))
	}
}

func TestDirectDifferentMnemonic(t *testing.T) {
	svc1, err := NewDirectService(testMnemonic12)
	if err != nil {
		t.Fatalf("first NewDirectService: %v", err)
	}

	svc2, err := NewDirectService(testMnemonic12Alt)
	if err != nil {
		t.Fatalf("second NewDirectService: %v", err)
	}

	if svc1.GetAddress() == svc2.GetAddress() {
		t.Error("different mnemonics should produce different addresses")
	}
}

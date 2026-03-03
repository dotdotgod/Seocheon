package signing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seocheon/sdk-go/internal/crypto"
)

func TestVaultService_MockServer(t *testing.T) {
	// Create a known key for the mock vault
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	key, err := crypto.DeriveKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("DeriveKeyFromMnemonic: %v", err)
	}

	addr, _ := crypto.AddressFromPubKey(key.PubKey())
	pubKeyHex := hex.EncodeToString(key.PubKey())

	// Create mock vault server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/v1/keys/test-key/address":
			json.NewEncoder(w).Encode(vaultAddressResponse{Address: addr})

		case r.Method == "GET" && r.URL.Path == "/v1/keys/test-key/pubkey":
			json.NewEncoder(w).Encode(vaultPubKeyResponse{PubKey: pubKeyHex})

		case r.Method == "POST" && r.URL.Path == "/v1/keys/test-key/sign":
			body, _ := io.ReadAll(r.Body)
			var req vaultSignRequest
			json.Unmarshal(body, &req)

			dataBytes, _ := hex.DecodeString(req.Data)
			// The vault signs the raw data (which is already a SignDoc)
			// The signing service calls Sign(signDocBytes), which arrives here
			// We need to hash it with SHA256 and sign (matching crypto.PrivateKey.Sign behavior)
			hash := sha256.Sum256(dataBytes)
			_ = hash // The vault would sign the hash internally

			sig, _ := key.Sign(dataBytes)
			json.NewEncoder(w).Encode(vaultSignResponse{Signature: hex.EncodeToString(sig)})

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create vault service
	svc, err := NewVaultService(server.URL, "test-key")
	if err != nil {
		t.Fatalf("NewVaultService: %v", err)
	}

	// Verify address
	if svc.GetAddress() != addr {
		t.Errorf("address = %s, want %s", svc.GetAddress(), addr)
	}

	// Verify public key
	gotPubKey, err := svc.GetPubKey()
	if err != nil {
		t.Fatalf("GetPubKey: %v", err)
	}
	if hex.EncodeToString(gotPubKey) != pubKeyHex {
		t.Errorf("pubkey mismatch")
	}

	// Sign and verify
	testData := []byte("test data for vault signing")
	sig, err := svc.Sign(testData)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sig) != 64 {
		t.Errorf("signature length = %d, want 64", len(sig))
	}

	if !crypto.Verify(gotPubKey, testData, sig) {
		t.Error("signature verification failed")
	}
}

func TestVaultService_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := NewVaultService(server.URL, "test-key")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestVaultService_EmptyParams(t *testing.T) {
	_, err := NewVaultService("", "key")
	if err == nil {
		t.Fatal("expected error for empty endpoint")
	}

	_, err = NewVaultService("http://localhost", "")
	if err == nil {
		t.Fatal("expected error for empty key name")
	}
}

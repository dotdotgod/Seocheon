package tx

import (
	"testing"
)

func TestEncodeTxBody(t *testing.T) {
	msg := &MsgSubmitActivity{
		Submitter:    "seocheon1test",
		ActivityHash: "abcd1234",
		ContentURI:   "https://example.com/report",
	}

	body := EncodeTxBody([]MessageEncoder{msg}, "", 0)
	if len(body) == 0 {
		t.Fatal("EncodeTxBody returned empty bytes")
	}

	// Should start with field 1 (messages), wire type 2 = tag 0x0a
	if body[0] != 0x0a {
		t.Errorf("TxBody first byte = 0x%02x, want 0x0a", body[0])
	}
}

func TestEncodeTxBody_WithMemo(t *testing.T) {
	msg := &MsgSubmitActivity{
		Submitter:    "seocheon1test",
		ActivityHash: "abcd1234",
		ContentURI:   "https://example.com",
	}

	body := EncodeTxBody([]MessageEncoder{msg}, "test memo", 100)
	if len(body) == 0 {
		t.Fatal("EncodeTxBody with memo returned empty bytes")
	}
}

func TestEncodeAuthInfo(t *testing.T) {
	// Fake 33-byte compressed public key
	pubKey := make([]byte, 33)
	pubKey[0] = 0x02
	for i := 1; i < 33; i++ {
		pubKey[i] = byte(i)
	}

	feeCoins := []Coin{{Denom: "uppyeo", Amount: "5000"}}
	authInfo := EncodeAuthInfo(pubKey, 0, feeCoins, 200000)

	if len(authInfo) == 0 {
		t.Fatal("EncodeAuthInfo returned empty bytes")
	}
}

func TestEncodeSignDoc(t *testing.T) {
	bodyBytes := []byte{0x0a, 0x05, 0x68, 0x65, 0x6c, 0x6c, 0x6f}
	authInfoBytes := []byte{0x12, 0x03, 0x66, 0x6f, 0x6f}

	signDoc := EncodeSignDoc(bodyBytes, authInfoBytes, "seocheon-1", 42)
	if len(signDoc) == 0 {
		t.Fatal("EncodeSignDoc returned empty bytes")
	}
}

func TestEncodeTxRaw(t *testing.T) {
	bodyBytes := []byte{1, 2, 3}
	authInfoBytes := []byte{4, 5, 6}
	signature := []byte{7, 8, 9}

	txRaw := EncodeTxRaw(bodyBytes, authInfoBytes, signature)
	if len(txRaw) == 0 {
		t.Fatal("EncodeTxRaw returned empty bytes")
	}
}

func TestFullTxRawAssembly(t *testing.T) {
	// Build a complete TX
	msg := &MsgSubmitActivity{
		Submitter:    "seocheon1abc123",
		ActivityHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		ContentURI:   "ipfs://QmTest123",
	}

	// Encode TxBody
	bodyBytes := EncodeTxBody([]MessageEncoder{msg}, "", 0)

	// Fake public key
	pubKey := make([]byte, 33)
	pubKey[0] = 0x03

	// Encode AuthInfo
	feeCoins := []Coin{{Denom: "uppyeo", Amount: "50000"}}
	authInfoBytes := EncodeAuthInfo(pubKey, 5, feeCoins, 200000)

	// Encode SignDoc
	signDoc := EncodeSignDoc(bodyBytes, authInfoBytes, "seocheon-testnet-1", 10)
	if len(signDoc) == 0 {
		t.Fatal("SignDoc is empty")
	}

	// Fake signature
	sig := make([]byte, 64)
	for i := range sig {
		sig[i] = byte(i)
	}

	// Encode TxRaw
	txRaw := EncodeTxRaw(bodyBytes, authInfoBytes, sig)
	if len(txRaw) == 0 {
		t.Fatal("TxRaw is empty")
	}

	t.Logf("TxRaw size: %d bytes", len(txRaw))
	t.Logf("TxBody size: %d bytes", len(bodyBytes))
	t.Logf("AuthInfo size: %d bytes", len(authInfoBytes))
	t.Logf("SignDoc size: %d bytes", len(signDoc))
}

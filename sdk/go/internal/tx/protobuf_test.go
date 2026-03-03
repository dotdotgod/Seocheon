package tx

import (
	"bytes"
	"testing"
)

func TestEncodeVarint(t *testing.T) {
	tests := []struct {
		input    uint64
		expected []byte
	}{
		{0, []byte{0}},
		{1, []byte{1}},
		{127, []byte{127}},
		{128, []byte{0x80, 0x01}},
		{300, []byte{0xac, 0x02}},
		{16384, []byte{0x80, 0x80, 0x01}},
	}

	for _, tt := range tests {
		result := EncodeVarint(tt.input)
		if !bytes.Equal(result, tt.expected) {
			t.Errorf("EncodeVarint(%d) = %x, want %x", tt.input, result, tt.expected)
		}
	}
}

func TestEncodeFieldString(t *testing.T) {
	result := EncodeFieldString(1, "hello")
	// field 1, wire type 2 = tag 0x0a, length 5, "hello"
	expected := []byte{0x0a, 0x05, 'h', 'e', 'l', 'l', 'o'}
	if !bytes.Equal(result, expected) {
		t.Errorf("EncodeFieldString(1, \"hello\") = %x, want %x", result, expected)
	}
}

func TestEncodeFieldString_Empty(t *testing.T) {
	result := EncodeFieldString(1, "")
	if result != nil {
		t.Errorf("EncodeFieldString(1, \"\") = %x, want nil", result)
	}
}

func TestEncodeFieldVarint_Zero(t *testing.T) {
	result := EncodeFieldVarint(1, 0)
	if result != nil {
		t.Errorf("EncodeFieldVarint(1, 0) = %x, want nil", result)
	}
}

func TestEncodeFieldVarint_NonZero(t *testing.T) {
	result := EncodeFieldVarint(1, 42)
	// field 1, wire type 0 = tag 0x08, value 42
	expected := []byte{0x08, 0x2a}
	if !bytes.Equal(result, expected) {
		t.Errorf("EncodeFieldVarint(1, 42) = %x, want %x", result, expected)
	}
}

func TestMsgSubmitActivity_Encode(t *testing.T) {
	msg := &MsgSubmitActivity{
		Submitter:    "seocheon1abc",
		ActivityHash: "deadbeef",
		ContentURI:   "ipfs://Qm123",
	}

	if msg.TypeURL() != "/seocheon.activity.v1.MsgSubmitActivity" {
		t.Errorf("TypeURL() = %s", msg.TypeURL())
	}

	encoded := msg.Encode()
	if len(encoded) == 0 {
		t.Error("Encode() returned empty bytes")
	}

	// Verify field 1 (submitter) starts with tag 0x0a
	if encoded[0] != 0x0a {
		t.Errorf("first byte = 0x%02x, want 0x0a (field 1, wire type 2)", encoded[0])
	}
}

func TestMsgSend_Encode(t *testing.T) {
	msg := &MsgSend{
		FromAddress: "seocheon1from",
		ToAddress:   "seocheon1to",
		Amount:      []Coin{{Denom: "uppyeo", Amount: "1000"}},
	}

	if msg.TypeURL() != "/cosmos.bank.v1beta1.MsgSend" {
		t.Errorf("TypeURL() = %s", msg.TypeURL())
	}

	encoded := msg.Encode()
	if len(encoded) == 0 {
		t.Error("Encode() returned empty bytes")
	}
}

func TestConcatBytes_NilHandling(t *testing.T) {
	result := ConcatBytes(nil, []byte{1}, nil, []byte{2, 3})
	expected := []byte{1, 2, 3}
	if !bytes.Equal(result, expected) {
		t.Errorf("ConcatBytes with nils = %x, want %x", result, expected)
	}
}

func TestMsgWithdrawCommissionUrlEncode(t *testing.T) {
	msg := &MsgWithdrawNodeCommission{
		Operator: "seocheon1operator",
	}

	if msg.TypeURL() != "/seocheon.node.v1.MsgWithdrawNodeCommission" {
		t.Errorf("TypeURL() = %s", msg.TypeURL())
	}

	encoded := msg.Encode()
	if len(encoded) == 0 {
		t.Error("Encode() returned empty bytes")
	}

	// Field 1 (operator) should start with tag 0x0a
	if encoded[0] != 0x0a {
		t.Errorf("first byte = 0x%02x, want 0x0a (field 1, wire type 2)", encoded[0])
	}
}

func TestMsgSendMultipleCoins(t *testing.T) {
	msg := &MsgSend{
		FromAddress: "seocheon1from",
		ToAddress:   "seocheon1to",
		Amount: []Coin{
			{Denom: "uppyeo", Amount: "1000"},
			{Denom: "kkot", Amount: "5"},
		},
	}

	encoded := msg.Encode()
	if len(encoded) == 0 {
		t.Error("Encode() returned empty bytes for multi-coin MsgSend")
	}

	// Should be longer than single-coin version
	singleCoinMsg := &MsgSend{
		FromAddress: "seocheon1from",
		ToAddress:   "seocheon1to",
		Amount:      []Coin{{Denom: "uppyeo", Amount: "1000"}},
	}
	singleEncoded := singleCoinMsg.Encode()
	if len(encoded) <= len(singleEncoded) {
		t.Error("multi-coin message should be longer than single-coin")
	}
}

func TestCoinEncode(t *testing.T) {
	coin := Coin{Denom: "uppyeo", Amount: "5000"}
	encoded := coin.Encode()
	if len(encoded) == 0 {
		t.Error("Coin.Encode() returned empty bytes")
	}

	// First field should be denom (field 1, wire type 2 = tag 0x0a)
	if encoded[0] != 0x0a {
		t.Errorf("first byte = 0x%02x, want 0x0a (field 1, wire type 2)", encoded[0])
	}
}

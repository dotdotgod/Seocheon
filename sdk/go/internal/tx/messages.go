package tx

// MessageEncoder encodes a specific message type into protobuf bytes.
type MessageEncoder interface {
	// TypeURL returns the protobuf type URL (e.g., "/seocheon.activity.v1.MsgSubmitActivity").
	TypeURL() string
	// Encode returns the protobuf-encoded message bytes.
	Encode() []byte
}

// MsgSubmitActivity encodes /seocheon.activity.v1.MsgSubmitActivity.
// Fields: submitter(1, string), activity_hash(2, string), content_uri(3, string)
type MsgSubmitActivity struct {
	Submitter    string
	ActivityHash string
	ContentURI   string
}

func (m *MsgSubmitActivity) TypeURL() string {
	return "/seocheon.activity.v1.MsgSubmitActivity"
}

func (m *MsgSubmitActivity) Encode() []byte {
	return ConcatBytes(
		EncodeFieldString(1, m.Submitter),
		EncodeFieldString(2, m.ActivityHash),
		EncodeFieldString(3, m.ContentURI),
	)
}

// MsgWithdrawNodeCommission encodes /seocheon.node.v1.MsgWithdrawNodeCommission.
// Fields: operator(1, string)
type MsgWithdrawNodeCommission struct {
	Operator string
}

func (m *MsgWithdrawNodeCommission) TypeURL() string {
	return "/seocheon.node.v1.MsgWithdrawNodeCommission"
}

func (m *MsgWithdrawNodeCommission) Encode() []byte {
	return ConcatBytes(
		EncodeFieldString(1, m.Operator),
	)
}

// Coin represents a cosmos.base.v1beta1.Coin.
// Fields: denom(1, string), amount(2, string)
type Coin struct {
	Denom  string
	Amount string
}

func (c *Coin) Encode() []byte {
	return ConcatBytes(
		EncodeFieldString(1, c.Denom),
		EncodeFieldString(2, c.Amount),
	)
}

// MsgSend encodes /cosmos.bank.v1beta1.MsgSend.
// Fields: from_address(1, string), to_address(2, string), amount(3, repeated Coin)
type MsgSend struct {
	FromAddress string
	ToAddress   string
	Amount      []Coin
}

func (m *MsgSend) TypeURL() string {
	return "/cosmos.bank.v1beta1.MsgSend"
}

func (m *MsgSend) Encode() []byte {
	parts := [][]byte{
		EncodeFieldString(1, m.FromAddress),
		EncodeFieldString(2, m.ToAddress),
	}
	for _, coin := range m.Amount {
		coinBytes := coin.Encode()
		parts = append(parts, EncodeFieldBytes(3, coinBytes))
	}
	return ConcatBytes(parts...)
}

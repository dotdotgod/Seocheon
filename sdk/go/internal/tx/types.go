package tx

import "context"

// TxRequest holds the parameters for building and broadcasting a transaction.
type TxRequest struct {
	// Message is the transaction message to encode.
	Message MessageEncoder
	// Memo is an optional memo string.
	Memo string
	// TimeoutHeight is the block height after which this TX expires (0 = no timeout).
	TimeoutHeight uint64
	// GasLimit overrides the default gas limit for this message type.
	GasLimit uint64
	// FeeAmount overrides the default fee (in uppyeo).
	FeeAmount uint64
	// FeeDenom overrides the fee denomination (default: "uppyeo").
	FeeDenom string
}

// TxResult holds the result of a successfully broadcast and confirmed transaction.
type TxResult struct {
	// TxHash is the transaction hash.
	TxHash string
	// Height is the block height at which the TX was included.
	Height int64
	// Code is the ABCI result code (0 = success).
	Code uint32
	// GasUsed is the actual gas consumed.
	GasUsed uint64
	// GasWanted is the gas limit requested.
	GasWanted uint64
	// RawLog is the raw log output.
	RawLog string
	// Events contains the transaction events.
	Events []TxEvent
}

// TxEvent represents a transaction event.
type TxEvent struct {
	Type       string
	Attributes []EventAttribute
}

// EventAttribute is a key-value pair in a TxEvent.
type EventAttribute struct {
	Key   string
	Value string
}

// Signer abstracts the signing capability needed by the TX pipeline.
type Signer interface {
	// Sign signs the given bytes (typically a SignDoc) and returns the signature.
	Sign(data []byte) ([]byte, error)
	// GetAddress returns the signer's bech32 address.
	GetAddress() string
	// GetPubKey returns the signer's compressed public key bytes.
	GetPubKey() ([]byte, error)
}

// ChainQuerier abstracts the chain queries needed by the TX pipeline.
type ChainQuerier interface {
	// GetAccountInfo returns the account number and sequence for an address.
	GetAccountInfo(ctx context.Context, address string) (accountNumber, sequence uint64, err error)
	// BroadcastTxSync broadcasts a TX and returns the hash and code.
	BroadcastTxSync(ctx context.Context, txBytes []byte) (txHash string, code uint32, rawLog string, err error)
	// GetTxResult queries a TX result by hash.
	GetTxResult(ctx context.Context, txHash string) (*TxResult, error)
}

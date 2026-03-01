// Package errors defines error types for the Seocheon SDK.
package errors

import "fmt"

// SDKError represents an error from the Seocheon SDK with an optional code.
type SDKError struct {
	Code    uint32
	Message string
	Err     error
}

func (e *SDKError) Error() string {
	if e.Code > 0 {
		return fmt.Sprintf("[%d] %s", e.Code, e.Message)
	}
	return e.Message
}

func (e *SDKError) Unwrap() error {
	return e.Err
}

// New creates a new SDKError with the given code and message.
func New(code uint32, message string) *SDKError {
	return &SDKError{Code: code, Message: message}
}

// Wrap creates a new SDKError wrapping an existing error.
func Wrap(code uint32, message string, err error) *SDKError {
	return &SDKError{Code: code, Message: message, Err: err}
}

// SDK-level errors (9000 series)
var (
	ErrNotConnected   = New(9000, "SDK is not connected to chain")
	ErrBroadcastFailed = New(9001, "transaction broadcast failed")
	ErrTxTimeout      = New(9002, "transaction confirmation timeout")
	ErrTxNotFound     = New(9003, "transaction not found")
	ErrSigningFailed  = New(9004, "transaction signing failed")
	ErrInvalidConfig  = New(9005, "invalid SDK configuration")
	ErrQueryFailed    = New(9006, "chain query failed")
	ErrInvalidAddress = New(9007, "invalid address format")
)

// x/node errors (1100 series)
var (
	ErrNodeNotFound         = New(1101, "node not found")
	ErrNodeAlreadyExists    = New(1102, "node already exists for this operator")
	ErrUnauthorizedOperator = New(1108, "unauthorized: signer is not the node operator")
	ErrUnauthorizedAgentMsg = New(1109, "agent address is not authorized for this message type")
)

// x/activity errors (1200 series)
var (
	ErrSubmitterNotRegistered = New(1200, "submitter agent address is not registered to any node")
	ErrNodeNotEligible        = New(1201, "node is not in an eligible status (REGISTERED or ACTIVE)")
	ErrDuplicateActivityHash  = New(1202, "duplicate (activity_hash, content_uri) pair already exists")
	ErrQuotaExceeded          = New(1203, "activity quota exceeded for this epoch")
	ErrInvalidActivityHash    = New(1204, "activity hash must be exactly 64 hex characters (32 bytes)")
	ErrInvalidContentURI      = New(1205, "content URI must not be empty")
)

// ABCICodeToError maps an ABCI error code to its corresponding SDKError.
func ABCICodeToError(code uint32) *SDKError {
	switch code {
	case 1101:
		return ErrNodeNotFound
	case 1102:
		return ErrNodeAlreadyExists
	case 1108:
		return ErrUnauthorizedOperator
	case 1109:
		return ErrUnauthorizedAgentMsg
	case 1200:
		return ErrSubmitterNotRegistered
	case 1201:
		return ErrNodeNotEligible
	case 1202:
		return ErrDuplicateActivityHash
	case 1203:
		return ErrQuotaExceeded
	case 1204:
		return ErrInvalidActivityHash
	case 1205:
		return ErrInvalidContentURI
	default:
		return New(code, fmt.Sprintf("chain error code %d", code))
	}
}

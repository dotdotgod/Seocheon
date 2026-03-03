package errors

import (
	"errors"
	"fmt"
	"testing"
)

// Cat 01: Error tests (6 tests)

func TestErrorWithCodeAndMessage(t *testing.T) {
	err := New(1101, "node not found")
	if err.Code != 1101 {
		t.Errorf("Code = %d, want 1101", err.Code)
	}
	if err.Message != "node not found" {
		t.Errorf("Message = %s, want 'node not found'", err.Message)
	}
	expected := "[1101] node not found"
	if err.Error() != expected {
		t.Errorf("Error() = %s, want %s", err.Error(), expected)
	}
}

func TestErrorWithoutCode(t *testing.T) {
	err := New(0, "generic error")
	if err.Code != 0 {
		t.Errorf("Code = %d, want 0", err.Code)
	}
	// Without code, Error() should return just the message
	if err.Error() != "generic error" {
		t.Errorf("Error() = %s, want 'generic error'", err.Error())
	}
}

func TestWrapErrorPreservesCause(t *testing.T) {
	cause := fmt.Errorf("original error")
	wrapped := Wrap(9001, "broadcast failed", cause)
	if wrapped.Code != 9001 {
		t.Errorf("Code = %d, want 9001", wrapped.Code)
	}
	// Unwrap should return the cause
	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestAbciCodeToErrorKnown(t *testing.T) {
	tests := []struct {
		code     uint32
		expected *SDKError
	}{
		{1101, ErrNodeNotFound},
		{1102, ErrNodeAlreadyExists},
		{1200, ErrSubmitterNotRegistered},
		{1203, ErrQuotaExceeded},
		{1204, ErrInvalidActivityHash},
	}

	for _, tt := range tests {
		result := ABCICodeToError(tt.code)
		if result != tt.expected {
			t.Errorf("ABCICodeToError(%d) = %v, want %v", tt.code, result, tt.expected)
		}
	}
}

func TestAbciCodeToErrorUnknown(t *testing.T) {
	result := ABCICodeToError(9999)
	if result.Code != 9999 {
		t.Errorf("Code = %d, want 9999", result.Code)
	}
	if result.Message != "chain error code 9999" {
		t.Errorf("Message = %s, want 'chain error code 9999'", result.Message)
	}
}

func TestPredefinedErrorCodesMatch(t *testing.T) {
	errorMap := map[uint32]*SDKError{
		9000: ErrNotConnected,
		9001: ErrBroadcastFailed,
		9002: ErrTxTimeout,
		9003: ErrTxNotFound,
		9004: ErrSigningFailed,
		9005: ErrInvalidConfig,
		9006: ErrQueryFailed,
		9007: ErrInvalidAddress,
		1101: ErrNodeNotFound,
		1102: ErrNodeAlreadyExists,
		1200: ErrSubmitterNotRegistered,
		1204: ErrInvalidActivityHash,
		1205: ErrInvalidContentURI,
	}

	for code, err := range errorMap {
		if err.Code != code {
			t.Errorf("error %s has code %d, want %d", err.Message, err.Code, code)
		}
	}
}

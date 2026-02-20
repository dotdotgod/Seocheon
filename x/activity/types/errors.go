package types

import "cosmossdk.io/errors"

// x/activity module sentinel errors.
var (
	ErrSubmitterNotRegistered = errors.Register(ModuleName, 1200, "submitter agent address is not registered to any node")
	ErrNodeNotEligible        = errors.Register(ModuleName, 1201, "node is not in an eligible status (REGISTERED or ACTIVE)")
	ErrDuplicateActivityHash  = errors.Register(ModuleName, 1202, "duplicate activity hash within the same epoch")
	ErrQuotaExceeded          = errors.Register(ModuleName, 1203, "activity quota exceeded for this epoch")
	ErrInvalidActivityHash    = errors.Register(ModuleName, 1204, "activity hash must be exactly 64 hex characters (32 bytes)")
	ErrInvalidContentURI      = errors.Register(ModuleName, 1205, "content URI must not be empty")
	ErrInvalidAuthority       = errors.Register(ModuleName, 1206, "expected gov account as only signer for proposal message")
	ErrInvalidParams          = errors.Register(ModuleName, 1207, "invalid module parameters")
	ErrActivityNotFound       = errors.Register(ModuleName, 1208, "activity record not found")
	ErrNodeNotFound           = errors.Register(ModuleName, 1209, "node not found")
)

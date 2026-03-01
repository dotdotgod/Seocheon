package types

import "cosmossdk.io/errors"

// x/randomness module sentinel errors.
var (
	ErrInvalidSigner        = errors.Register(ModuleName, 1300, "expected gov account as only signer for proposal message")
	ErrBeaconNotFound       = errors.Register(ModuleName, 1301, "beacon not found for the given round")
	ErrDuplicateBeacon      = errors.Register(ModuleName, 1302, "beacon for this round already exists")
	ErrInvalidRound         = errors.Register(ModuleName, 1303, "invalid drand round number")
	ErrInvalidRandomness    = errors.Register(ModuleName, 1304, "invalid randomness value: must be 64-char hex string")
	ErrInvalidSignature     = errors.Register(ModuleName, 1305, "invalid beacon signature")
	ErrBeaconTooOld         = errors.Register(ModuleName, 1306, "beacon round is too old")
	ErrVerificationDisabled = errors.Register(ModuleName, 1307, "beacon verification is disabled")
	ErrVerificationFailed   = errors.Register(ModuleName, 1308, "beacon signature verification failed")
	ErrNoBeaconAvailable    = errors.Register(ModuleName, 1309, "no beacon available")

	// Commit-Reveal errors (1310~1320).
	ErrCommitRevealDisabled   = errors.Register(ModuleName, 1310, "commit-reveal system is disabled")
	ErrInvalidCommitHash      = errors.Register(ModuleName, 1311, "invalid commit hash: must be 64-char hex SHA-256")
	ErrInvalidNumWords        = errors.Register(ModuleName, 1312, "invalid num_words: must be 1-10")
	ErrCallbackDataTooLong    = errors.Register(ModuleName, 1313, "callback_data exceeds 256 bytes")
	ErrInvalidCallbackData    = errors.Register(ModuleName, 1314, "invalid callback_data: must be hex-encoded")
	ErrTooManyPendingRequests = errors.Register(ModuleName, 1315, "global pending request limit exceeded")
	ErrRequestNotFound        = errors.Register(ModuleName, 1316, "randomness request not found")
	ErrRequestNotPending      = errors.Register(ModuleName, 1317, "randomness request is not in pending status")
	ErrBeaconNotYetAvailable  = errors.Register(ModuleName, 1318, "beacon for target round not yet available")
	ErrRequesterLimitExceeded = errors.Register(ModuleName, 1319, "per-requester pending request limit exceeded")
	ErrInsufficientRequestFee = errors.Register(ModuleName, 1320, "request fee is less than minimum required")
)

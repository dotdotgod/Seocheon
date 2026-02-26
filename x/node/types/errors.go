package types

import "cosmossdk.io/errors"

// x/node module sentinel errors.
var (
	ErrInvalidSigner              = errors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrNodeNotFound               = errors.Register(ModuleName, 1101, "node not found")
	ErrNodeAlreadyExists          = errors.Register(ModuleName, 1102, "node already exists for this operator")
	ErrAgentAddressAlreadyUsed    = errors.Register(ModuleName, 1103, "agent address already registered to another node")
	ErrRegistrationPoolDepleted   = errors.Register(ModuleName, 1104, "registration pool has insufficient funds")
	ErrMaxRegistrationsPerBlock   = errors.Register(ModuleName, 1105, "maximum registrations per block exceeded")
	ErrInvalidAgentShare          = errors.Register(ModuleName, 1106, "agent share must be between 0 and 100")
	ErrInvalidTags                = errors.Register(ModuleName, 1107, "invalid tags")
	ErrUnauthorizedOperator       = errors.Register(ModuleName, 1108, "unauthorized: signer is not the node operator")
	ErrUnauthorizedAgentMsg       = errors.Register(ModuleName, 1109, "agent address is not authorized for this message type")
	ErrAgentAddressChangeCooldown = errors.Register(ModuleName, 1110, "agent address change cooldown not elapsed")
	ErrNodeNotActive              = errors.Register(ModuleName, 1111, "node is not in an active state")
	ErrInvalidConsensusPubkey     = errors.Register(ModuleName, 1112, "invalid consensus public key")
	ErrAgentShareChangeExceedsMax = errors.Register(ModuleName, 1113, "agent share change exceeds max change rate")
	ErrNodeInactive               = errors.Register(ModuleName, 1116, "node is inactive or jailed")
	ErrValidatorAlreadyExists     = errors.Register(ModuleName, 1117, "validator with this consensus pubkey already exists")
	ErrAgentShareChangePending    = errors.Register(ModuleName, 1118, "a pending agent share change already exists")
)

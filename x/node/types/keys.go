package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "node"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	GovModuleName = "gov"

	// RegistrationPoolName is the module account name for the Registration Pool.
	RegistrationPoolName = "node_registration_pool"

	// FeegrantPoolName is the module account name for the Feegrant Pool.
	FeegrantPoolName = "node_feegrant_pool"
)

// Store key prefixes.
var (
	ParamsKey = collections.NewPrefix("p_node")

	// NodeKey: node_id -> Node
	NodeKey = collections.NewPrefix(1)

	// OperatorIndexKey: operator_addr -> node_id (1:1)
	OperatorIndexKey = collections.NewPrefix(2)

	// AgentIndexKey: agent_addr -> node_id (1:1)
	AgentIndexKey = collections.NewPrefix(3)

	// ValidatorIndexKey: validator_addr -> node_id (1:1)
	ValidatorIndexKey = collections.NewPrefix(4)

	// TagIndexKey: (tag, node_id) -> empty (1:N, NodesByTag query)
	TagIndexKey = collections.NewPrefix(5)

	// BlockRegistrationCountKey: block_height -> count (per-block registration tracking)
	BlockRegistrationCountKey = collections.NewPrefix(6)

	// PendingAgentShareChangeKey: node_id -> PendingAgentShareChange
	PendingAgentShareChangeKey = collections.NewPrefix(7)

	// LastAgentChangeBlockKey: node_id -> block_height (agent address change cooldown)
	LastAgentChangeBlockKey = collections.NewPrefix(8)
)

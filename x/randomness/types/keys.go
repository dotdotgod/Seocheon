package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "randomness"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	GovModuleName = "gov"
)

// Store key prefixes.
var (
	ParamsKey = collections.NewPrefix("p_rand")

	// BeaconKey: round (uint64) -> Beacon
	BeaconKey = collections.NewPrefix(1)

	// LatestRoundKey: stores the latest beacon round number.
	LatestRoundKey = collections.NewPrefix(2)

	// Commit-Reveal store prefixes (30~35).

	// RequestSequenceKey: auto-incrementing request ID sequence.
	RequestSequenceKey = collections.NewPrefix(30)

	// RequestsKey: request_id (uint64) -> RandomnessRequest
	RequestsKey = collections.NewPrefix(31)

	// PendingByRoundKey: (target_round, request_id) -> KeySet for EndBlocker lookup.
	PendingByRoundKey = collections.NewPrefix(32)

	// RequestsByRequesterKey: (requester, request_id) -> KeySet for per-requester queries.
	RequestsByRequesterKey = collections.NewPrefix(33)

	// PendingRequestCountKey: global count of pending requests.
	PendingRequestCountKey = collections.NewPrefix(34)

	// FulfilledBlockIndexKey: (fulfilled_block_height, request_id) -> KeySet for pruning.
	FulfilledBlockIndexKey = collections.NewPrefix(35)
)

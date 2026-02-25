package types

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name.
	ModuleName = "activity"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// GovModuleName duplicates the gov module's name to avoid a dependency with x/gov.
	GovModuleName = "gov"

	// ActivityRewardPoolName is the module account name for the activity reward pool.
	// Block rewards (activity_ratio fraction) and 80% of collected activity fees are accumulated here,
	// then distributed equally to eligible activity nodes at each epoch boundary.
	ActivityRewardPoolName = "activity_reward_pool"
)

// Store key prefixes.
// Note: x/node uses "p_node" + prefixes 1-8. x/activity uses "p_activity" + prefixes 10-17
// to avoid any collision.
var (
	ParamsKey = collections.NewPrefix("p_activity")

	// ActivitiesKey: (node_id, epoch, sequence) -> ActivityRecord
	ActivitiesKey = collections.NewPrefix(10)

	// HashIndexKey: (node_id, epoch, hash_hex) -> empty (duplicate detection)
	HashIndexKey = collections.NewPrefix(11)

	// EpochQuotaUsedKey: (node_id, epoch) -> count
	EpochQuotaUsedKey = collections.NewPrefix(12)

	// WindowActivityKey: (node_id, epoch, window) -> count
	WindowActivityKey = collections.NewPrefix(13)

	// EpochSummaryKey: (node_id, epoch) -> EpochActivitySummary
	EpochSummaryKey = collections.NewPrefix(14)

	// BlockActivitiesKey: (block_height, seq) -> node_id (for pruning)
	BlockActivitiesKey = collections.NewPrefix(15)

	// ActivitySequenceKey: (node_id, epoch) -> next_sequence
	ActivitySequenceKey = collections.NewPrefix(16)

	// GlobalHashIndexKey: hash_hex -> (node_id, epoch, sequence) for hash-only lookups
	GlobalHashIndexKey = collections.NewPrefix(17)

	// EpochActivityFeeKey: epoch -> activity_fee in usum (cached per epoch)
	EpochActivityFeeKey = collections.NewPrefix(18)

	// EpochEffectiveQuotaKey: epoch -> effective feegrant quota (cached per epoch)
	EpochEffectiveQuotaKey = collections.NewPrefix(19)

	// EpochCollectedFeesKey: epoch -> total collected fees in usum
	EpochCollectedFeesKey = collections.NewPrefix(20)

	// EpochActivityRewardPoolKey: epoch -> accumulated reward pool amount in usum for the epoch
	EpochActivityRewardPoolKey = collections.NewPrefix(21)
)

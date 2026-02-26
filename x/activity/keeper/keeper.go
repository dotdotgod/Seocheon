package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"seocheon/x/activity/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// authority is the address capable of executing a MsgUpdateParams message.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	// Activities stores all activity records: (node_id, epoch, sequence) -> ActivityRecord
	Activities collections.Map[collections.Triple[string, int64, uint64], types.ActivityRecord]

	// HashIndex for global duplicate detection: (activity_hash, content_uri) -> empty
	HashIndex collections.KeySet[collections.Pair[string, string]]

	// EpochQuotaUsed tracks quota usage: (node_id, epoch) -> count
	EpochQuotaUsed collections.Map[collections.Pair[string, int64], uint64]

	// WindowActivity tracks per-window activity counts: (node_id, epoch, window) -> count
	WindowActivity collections.Map[collections.Triple[string, int64, int64], uint64]

	// EpochSummary stores epoch activity summaries: (node_id, epoch) -> EpochActivitySummary
	EpochSummary collections.Map[collections.Pair[string, int64], types.EpochActivitySummary]

	// BlockActivities for pruning: (block_height, seq) -> node_id
	BlockActivities collections.Map[collections.Pair[int64, uint64], string]

	// ActivitySequence tracks the next sequence number: (node_id, epoch) -> next_seq
	ActivitySequence collections.Map[collections.Pair[string, int64], uint64]

	// EpochActivityFee: epoch -> activity_fee in usum (cached per epoch)
	EpochActivityFee collections.Map[int64, uint64]

	// EpochEffectiveQuota: epoch -> effective feegrant quota (cached per epoch)
	EpochEffectiveQuota collections.Map[int64, uint64]

	// EpochCollectedFees: epoch -> total collected fees in usum
	EpochCollectedFees collections.Map[int64, uint64]

	// EpochActivityRewardPool tracks accumulated reward pool amounts per epoch.
	EpochActivityRewardPool collections.Map[int64, uint64]

	// Keeper dependencies.
	nodeKeeper         types.NodeKeeper
	authKeeper         types.AuthKeeper
	feegrantKeeper     types.FeegrantKeeper
	stakingKeeper      types.StakingKeeper
	bankKeeper         types.BankKeeper
	distributionKeeper types.DistributionKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),

		Activities: collections.NewMap(sb, types.ActivitiesKey, "activities",
			collections.TripleKeyCodec(collections.StringKey, collections.Int64Key, collections.Uint64Key),
			codec.CollValue[types.ActivityRecord](cdc),
		),

		HashIndex: collections.NewKeySet(sb, types.HashIndexKey, "hash_index",
			collections.PairKeyCodec(collections.StringKey, collections.StringKey),
		),

		EpochQuotaUsed: collections.NewMap(sb, types.EpochQuotaUsedKey, "epoch_quota_used",
			collections.PairKeyCodec(collections.StringKey, collections.Int64Key),
			collections.Uint64Value,
		),

		WindowActivity: collections.NewMap(sb, types.WindowActivityKey, "window_activity",
			collections.TripleKeyCodec(collections.StringKey, collections.Int64Key, collections.Int64Key),
			collections.Uint64Value,
		),

		EpochSummary: collections.NewMap(sb, types.EpochSummaryKey, "epoch_summary",
			collections.PairKeyCodec(collections.StringKey, collections.Int64Key),
			codec.CollValue[types.EpochActivitySummary](cdc),
		),

		BlockActivities: collections.NewMap(sb, types.BlockActivitiesKey, "block_activities",
			collections.PairKeyCodec(collections.Int64Key, collections.Uint64Key),
			collections.StringValue,
		),

		ActivitySequence: collections.NewMap(sb, types.ActivitySequenceKey, "activity_sequence",
			collections.PairKeyCodec(collections.StringKey, collections.Int64Key),
			collections.Uint64Value,
		),

		EpochActivityFee: collections.NewMap(sb, types.EpochActivityFeeKey, "epoch_activity_fee",
			collections.Int64Key,
			collections.Uint64Value,
		),

		EpochEffectiveQuota: collections.NewMap(sb, types.EpochEffectiveQuotaKey, "epoch_effective_quota",
			collections.Int64Key,
			collections.Uint64Value,
		),

		EpochCollectedFees: collections.NewMap(sb, types.EpochCollectedFeesKey, "epoch_collected_fees",
			collections.Int64Key,
			collections.Uint64Value,
		),

		EpochActivityRewardPool: collections.NewMap(sb, types.EpochActivityRewardPoolKey, "epoch_activity_reward_pool",
			collections.Int64Key,
			collections.Uint64Value,
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// SetNodeKeeper sets the node keeper (called during module wiring).
func (k *Keeper) SetNodeKeeper(nk types.NodeKeeper) {
	k.nodeKeeper = nk
}

// SetAuthKeeper sets the auth keeper (called during module wiring).
func (k *Keeper) SetAuthKeeper(ak types.AuthKeeper) {
	k.authKeeper = ak
}

// SetFeegrantKeeper sets the feegrant keeper (called during module wiring).
func (k *Keeper) SetFeegrantKeeper(fk types.FeegrantKeeper) {
	k.feegrantKeeper = fk
}

// SetStakingKeeper sets the staking keeper for validator count queries.
func (k *Keeper) SetStakingKeeper(sk types.StakingKeeper) {
	k.stakingKeeper = sk
}

// SetBankKeeper sets the bank keeper for fee collection and distribution.
func (k *Keeper) SetBankKeeper(bk types.BankKeeper) {
	k.bankKeeper = bk
}

// SetDistributionKeeper sets the distribution keeper for community pool funding.
func (k *Keeper) SetDistributionKeeper(dk types.DistributionKeeper) {
	k.distributionKeeper = dk
}

// GetEpochLength returns the current epoch_length parameter.
// This method fulfills x/node's ActivityKeeper interface.
func (k Keeper) GetEpochLength(ctx context.Context) (int64, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return 0, err
	}
	return params.EpochLength, nil
}

// IsNodeEligibleForEpoch checks if a node met the activity threshold for a given epoch.
func (k Keeper) IsNodeEligibleForEpoch(ctx context.Context, nodeID string, epoch int64) (bool, error) {
	summary, err := k.EpochSummary.Get(ctx, collections.Join(nodeID, epoch))
	if err != nil {
		// No summary means no activity = not eligible.
		return false, nil
	}
	return summary.Eligible, nil
}

// CountEligibleEpochs counts the number of eligible epochs within the lookback window.
func (k Keeper) CountEligibleEpochs(ctx context.Context, nodeID string, currentEpoch int64, lookback int64) (int64, error) {
	var count int64
	startEpoch := currentEpoch - lookback
	if startEpoch < 0 {
		startEpoch = 0
	}

	for epoch := startEpoch; epoch < currentEpoch; epoch++ {
		eligible, err := k.IsNodeEligibleForEpoch(ctx, nodeID, epoch)
		if err != nil {
			return 0, err
		}
		if eligible {
			count++
		}
	}
	return count, nil
}

// ValidateActivityHash checks that a hex string is exactly 64 hex characters.
func ValidateActivityHash(hashHex string) bool {
	if len(hashHex) != 64 {
		return false
	}
	_, err := hex.DecodeString(hashHex)
	return err == nil
}

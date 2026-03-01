package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"seocheon/x/randomness/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// authority is the address capable of executing a MsgUpdateParams message.
	authority []byte

	// bankKeeper handles fee escrow operations for commit-reveal.
	bankKeeper types.BankKeeper

	Schema collections.Schema
	Params collections.Item[types.Params]

	// Beacons stores drand beacons indexed by round number.
	Beacons collections.Map[uint64, types.Beacon]

	// LatestRound stores the latest beacon round number.
	LatestRound collections.Item[uint64]

	// Commit-Reveal collections.

	// RequestSequence is an auto-incrementing sequence for request IDs.
	RequestSequence collections.Sequence

	// Requests stores randomness requests by request_id.
	Requests collections.Map[uint64, types.RandomnessRequest]

	// PendingByRound indexes pending requests by (target_round, request_id) for EndBlocker lookup.
	PendingByRound collections.KeySet[collections.Pair[uint64, uint64]]

	// RequestsByRequester indexes requests by (requester, request_id) for per-requester queries.
	RequestsByRequester collections.KeySet[collections.Pair[string, uint64]]

	// PendingRequestCount tracks the global number of pending requests.
	PendingRequestCount collections.Item[uint64]

	// FulfilledBlockIndex indexes fulfilled/expired requests by (block_height, request_id) for pruning.
	FulfilledBlockIndex collections.KeySet[collections.Pair[int64, uint64]]
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

		Beacons:     collections.NewMap(sb, types.BeaconKey, "beacons", collections.Uint64Key, codec.CollValue[types.Beacon](cdc)),
		LatestRound: collections.NewItem(sb, types.LatestRoundKey, "latest_round", collections.Uint64Value),

		// Commit-Reveal collections.
		RequestSequence: collections.NewSequence(sb, types.RequestSequenceKey, "request_sequence"),
		Requests:        collections.NewMap(sb, types.RequestsKey, "requests", collections.Uint64Key, codec.CollValue[types.RandomnessRequest](cdc)),
		PendingByRound: collections.NewKeySet(
			sb, types.PendingByRoundKey, "pending_by_round",
			collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key),
		),
		RequestsByRequester: collections.NewKeySet(
			sb, types.RequestsByRequesterKey, "requests_by_requester",
			collections.PairKeyCodec(collections.StringKey, collections.Uint64Key),
		),
		PendingRequestCount: collections.NewItem(sb, types.PendingRequestCountKey, "pending_request_count", collections.Uint64Value),
		FulfilledBlockIndex: collections.NewKeySet(
			sb, types.FulfilledBlockIndexKey, "fulfilled_block_index",
			collections.PairKeyCodec(collections.Int64Key, collections.Uint64Key),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// SetBankKeeper sets the bank keeper for fee escrow operations.
func (k *Keeper) SetBankKeeper(bk types.BankKeeper) {
	k.bankKeeper = bk
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// GetLatestBeacon returns the most recently stored beacon.
func (k Keeper) GetLatestBeacon(ctx context.Context) (types.Beacon, error) {
	latestRound, err := k.LatestRound.Get(ctx)
	if err != nil {
		return types.Beacon{}, types.ErrNoBeaconAvailable
	}

	beacon, err := k.Beacons.Get(ctx, latestRound)
	if err != nil {
		return types.Beacon{}, types.ErrBeaconNotFound
	}

	return beacon, nil
}

// GetBeaconByRound returns the beacon for the given round.
func (k Keeper) GetBeaconByRound(ctx context.Context, round uint64) (types.Beacon, error) {
	beacon, err := k.Beacons.Get(ctx, round)
	if err != nil {
		return types.Beacon{}, types.ErrBeaconNotFound
	}
	return beacon, nil
}

// ValidateRandomnessFormat validates that the randomness string is a valid 32-byte hex string.
func ValidateRandomnessFormat(randomness string) bool {
	if len(randomness) != 64 {
		return false
	}
	_, err := hex.DecodeString(randomness)
	return err == nil
}

// ValidateSignatureFormat validates that the signature string is a valid hex string.
func ValidateSignatureFormat(signature string) bool {
	if len(signature) == 0 {
		return false
	}
	_, err := hex.DecodeString(signature)
	return err == nil
}

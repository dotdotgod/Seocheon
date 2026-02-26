package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/activity/types"
)

// RegisterInvariants registers all module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "hash-index-consistency", HashIndexConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "quota-consistency", QuotaConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "window-activity-consistency", WindowActivityConsistencyInvariant(k))
}

// AllInvariants runs all invariants.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []sdk.Invariant{
			HashIndexConsistencyInvariant(k),
			QuotaConsistencyInvariant(k),
			WindowActivityConsistencyInvariant(k),
		} {
			msg, broken := inv(ctx)
			if broken {
				return msg, broken
			}
		}
		return "", false
	}
}

// HashIndexConsistencyInvariant checks that every activity in Activities has a corresponding HashIndex entry.
func HashIndexConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		iter, err := k.Activities.Iterate(ctx, nil)
		if err != nil {
			return fmt.Sprintf("failed to iterate activities: %v", err), true
		}
		defer iter.Close()

		for ; iter.Valid(); iter.Next() {
			kv, err := iter.KeyValue()
			if err != nil {
				return fmt.Sprintf("failed to get key-value: %v", err), true
			}
			record := kv.Value
			has, err := k.HashIndex.Has(ctx, collections.Join(record.ActivityHash, record.ContentUri))
			if err != nil {
				return fmt.Sprintf("failed to check hash index: %v", err), true
			}
			if !has {
				return fmt.Sprintf("activity %s:%d:%d missing from hash index", record.NodeId, record.Epoch, record.Sequence), true
			}
		}
		return "", false
	}
}

// QuotaConsistencyInvariant checks that EpochQuotaUsed matches actual activity count.
func QuotaConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Count actual activities per (node_id, epoch).
		counts := make(map[string]uint64) // "node_id:epoch" -> count

		iter, err := k.Activities.Iterate(ctx, nil)
		if err != nil {
			return fmt.Sprintf("failed to iterate activities: %v", err), true
		}
		defer iter.Close()

		for ; iter.Valid(); iter.Next() {
			record, err := iter.Value()
			if err != nil {
				return fmt.Sprintf("failed to get value: %v", err), true
			}
			key := fmt.Sprintf("%s:%d", record.NodeId, record.Epoch)
			counts[key]++
		}

		// Verify against EpochQuotaUsed.
		quotaIter, err := k.EpochQuotaUsed.Iterate(ctx, nil)
		if err != nil {
			return fmt.Sprintf("failed to iterate quota: %v", err), true
		}
		defer quotaIter.Close()

		for ; quotaIter.Valid(); quotaIter.Next() {
			kv, err := quotaIter.KeyValue()
			if err != nil {
				return fmt.Sprintf("failed to get quota key-value: %v", err), true
			}
			nodeID := kv.Key.K1()
			epoch := kv.Key.K2()
			quotaUsed := kv.Value

			key := fmt.Sprintf("%s:%d", nodeID, epoch)
			actual := counts[key]

			if quotaUsed != actual {
				return fmt.Sprintf("quota mismatch for %s: stored=%d, actual=%d", key, quotaUsed, actual), true
			}
		}

		return "", false
	}
}

// WindowActivityConsistencyInvariant checks that WindowActivity counts match actual activities.
func WindowActivityConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		params, err := k.Params.Get(ctx)
		if err != nil {
			return fmt.Sprintf("failed to get params: %v", err), true
		}

		// Count actual activities per (node_id, epoch, window).
		counts := make(map[string]uint64)

		iter, err := k.Activities.Iterate(ctx, nil)
		if err != nil {
			return fmt.Sprintf("failed to iterate activities: %v", err), true
		}
		defer iter.Close()

		for ; iter.Valid(); iter.Next() {
			record, err := iter.Value()
			if err != nil {
				return fmt.Sprintf("failed to get value: %v", err), true
			}
			window := GetCurrentWindow(record.BlockHeight, params)
			key := fmt.Sprintf("%s:%d:%d", record.NodeId, record.Epoch, window)
			counts[key]++
		}

		// Verify against WindowActivity.
		windowIter, err := k.WindowActivity.Iterate(ctx, nil)
		if err != nil {
			return fmt.Sprintf("failed to iterate window activity: %v", err), true
		}
		defer windowIter.Close()

		for ; windowIter.Valid(); windowIter.Next() {
			kv, err := windowIter.KeyValue()
			if err != nil {
				return fmt.Sprintf("failed to get window key-value: %v", err), true
			}
			nodeID := kv.Key.K1()
			epoch := kv.Key.K2()
			window := kv.Key.K3()
			stored := kv.Value

			key := fmt.Sprintf("%s:%d:%d", nodeID, epoch, window)
			actual := counts[key]

			if stored != actual {
				return fmt.Sprintf("window activity mismatch for %s: stored=%d, actual=%d", key, stored, actual), true
			}
		}

		return "", false
	}
}

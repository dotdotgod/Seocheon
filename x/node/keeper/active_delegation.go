package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/types"
)

// processExpiredDelegations handles expired delegation confirmations at epoch boundaries.
// It emits renewal window alerts and force-unbonds expired delegations.
func (k Keeper) processExpiredDelegations(ctx context.Context) error {
	currentEpoch := k.currentEpoch(ctx)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Skip if active delegation is disabled.
	if params.DelegationConfirmationPeriod == 0 {
		return nil
	}

	// (A) Emit renewal window alerts for delegations entering the window.
	k.emitRenewalWindowAlerts(ctx, currentEpoch, params)

	// (B) Process expired delegations: force unbond all with expiry <= currentEpoch.
	return k.forceUnbondExpired(ctx, currentEpoch)
}

// emitRenewalWindowAlerts emits events for delegations entering the renewal window.
func (k Keeper) emitRenewalWindowAlerts(ctx context.Context, currentEpoch uint64, params types.Params) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Find delegations whose expiry == currentEpoch + renewalWindow
	// (i.e., they are just entering the renewal window now).
	targetExpiry := currentEpoch + params.DelegationRenewalWindow

	// Iterate PendingExpirations where the first key component == targetExpiry.
	rng := collections.NewPrefixedTripleRange[uint64, string, string](targetExpiry)
	iter, err := k.PendingExpirations.Iterate(ctx, rng)
	if err != nil {
		return
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			continue
		}
		delegator := key.K2()
		validator := key.K3()

		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeDelegationRenewalWindowOpen,
			sdk.NewAttribute(types.AttributeKeyDelegator, delegator),
			sdk.NewAttribute(types.AttributeKeyValidatorAddress, validator),
			sdk.NewAttribute(types.AttributeKeyExpiryEpoch, fmt.Sprintf("%d", targetExpiry)),
			sdk.NewAttribute(types.AttributeKeyRenewalWindowStart, fmt.Sprintf("%d", currentEpoch)),
		))
	}
}

// forceUnbondExpired force-unbonds all delegations with expiry <= currentEpoch.
func (k Keeper) forceUnbondExpired(ctx context.Context, currentEpoch uint64) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Iterate all PendingExpirations with epoch <= currentEpoch.
	rng := new(collections.Range[collections.Triple[uint64, string, string]]).
		EndInclusive(collections.Join3(currentEpoch, "\xff", "\xff"))

	iter, err := k.PendingExpirations.Iterate(ctx, rng)
	if err != nil {
		return err
	}
	defer iter.Close()

	type expiredEntry struct {
		expiryEpoch uint64
		delegator   string
		validator   string
	}
	var toProcess []expiredEntry

	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			continue
		}
		toProcess = append(toProcess, expiredEntry{
			expiryEpoch: key.K1(),
			delegator:   key.K2(),
			validator:   key.K3(),
		})
	}

	if len(toProcess) == 0 {
		return nil
	}

	var unbondCount uint64
	for _, entry := range toProcess {
		// Check if delegation still exists (may have been manually unbonded).
		delAddr, err := sdk.AccAddressFromBech32(entry.delegator)
		if err != nil {
			k.cleanupConfirmation(ctx, entry.delegator, entry.validator, entry.expiryEpoch)
			continue
		}
		valAddr, err := sdk.ValAddressFromBech32(entry.validator)
		if err != nil {
			k.cleanupConfirmation(ctx, entry.delegator, entry.validator, entry.expiryEpoch)
			continue
		}

		delegation, err := k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
		if err != nil {
			// Delegation already removed — just clean up records.
			k.cleanupConfirmation(ctx, entry.delegator, entry.validator, entry.expiryEpoch)
			continue
		}

		// Get bond denom for the unbond message.
		bondDenom, err := k.stakingKeeper.BondDenom(ctx)
		if err != nil {
			continue
		}

		// Force unbond the full delegation amount.
		shares := delegation.GetShares()
		validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
		if err != nil {
			k.cleanupConfirmation(ctx, entry.delegator, entry.validator, entry.expiryEpoch)
			continue
		}
		unbondAmount := validator.TokensFromShares(shares).TruncateInt()

		undelegateMsg := &stakingtypes.MsgUndelegate{
			DelegatorAddress: entry.delegator,
			ValidatorAddress: entry.validator,
			Amount:           sdk.NewCoin(bondDenom, unbondAmount),
		}

		if k.stakingMsgServer != nil {
			_, err = k.stakingMsgServer.Undelegate(ctx, undelegateMsg)
			if err != nil {
				// Log failure but continue processing other expirations.
				sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
					types.EventTypeUndelegateFailed,
					sdk.NewAttribute(types.AttributeKeyDelegator, entry.delegator),
					sdk.NewAttribute(types.AttributeKeyValidatorAddress, entry.validator),
					sdk.NewAttribute(types.AttributeKeyError, err.Error()),
				))
				k.cleanupConfirmation(ctx, entry.delegator, entry.validator, entry.expiryEpoch)
				continue
			}
		}

		k.cleanupConfirmation(ctx, entry.delegator, entry.validator, entry.expiryEpoch)
		unbondCount++
	}

	if unbondCount > 0 {
		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeDelegationForceUnbonded,
			sdk.NewAttribute(types.AttributeKeyUnbondCount, fmt.Sprintf("%d", unbondCount)),
		))
	}

	return nil
}

// cleanupConfirmation removes both the DelegationConfirmations and PendingExpirations entries.
func (k Keeper) cleanupConfirmation(ctx context.Context, delegator, validator string, expiryEpoch uint64) {
	pairKey := collections.Join(delegator, validator)
	_ = k.DelegationConfirmations.Remove(ctx, pairKey)

	tripleKey := collections.Join3(expiryEpoch, delegator, validator)
	_ = k.PendingExpirations.Remove(ctx, tripleKey)
}

// setDelegationConfirmation creates or updates a delegation confirmation record.
func (k Keeper) setDelegationConfirmation(ctx context.Context, delegator, validator string, expiryEpoch uint64) error {
	pairKey := collections.Join(delegator, validator)
	if err := k.DelegationConfirmations.Set(ctx, pairKey, expiryEpoch); err != nil {
		return err
	}
	tripleKey := collections.Join3(expiryEpoch, delegator, validator)
	return k.PendingExpirations.Set(ctx, tripleKey)
}

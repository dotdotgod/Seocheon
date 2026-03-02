package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// ConfirmDelegation confirms an active delegation within the renewal window.
func (k msgServer) ConfirmDelegation(ctx context.Context, msg *types.MsgConfirmDelegation) (*types.MsgConfirmDelegationResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	// Check if active delegation is enabled.
	if params.DelegationConfirmationPeriod == 0 {
		return nil, types.ErrActiveDelegationDisabled
	}

	// Validate delegator address.
	delAddr, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid delegator address: %w", err)
	}

	// Validate validator address.
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid validator address: %w", err)
	}

	// Verify that the delegation exists in x/staking.
	_, err = k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	if err != nil {
		return nil, types.ErrDelegationNotFound.Wrapf("no delegation from %s to %s", msg.DelegatorAddress, msg.ValidatorAddress)
	}

	// Calculate current epoch.
	currentEpoch := k.currentEpoch(ctx)

	// Look up existing confirmation.
	pairKey := collections.Join(msg.DelegatorAddress, msg.ValidatorAddress)
	oldExpiry, err := k.DelegationConfirmations.Get(ctx, pairKey)
	if err != nil {
		return nil, types.ErrDelegationNotFound.Wrapf("no confirmation record for %s -> %s", msg.DelegatorAddress, msg.ValidatorAddress)
	}

	// Check if already expired.
	if currentEpoch >= oldExpiry {
		return nil, types.ErrAlreadyExpired.Wrapf("delegation expired at epoch %d, current epoch %d", oldExpiry, currentEpoch)
	}

	// Check renewal window: only accept during the last N epochs before expiry.
	renewalWindowStart := oldExpiry - params.DelegationRenewalWindow
	if currentEpoch < renewalWindowStart {
		return nil, types.ErrNotInRenewalWindow.Wrapf(
			"renewal window starts at epoch %d, current epoch %d (try again in %d epochs)",
			renewalWindowStart, currentEpoch, renewalWindowStart-currentEpoch,
		)
	}

	// Remove old expiration entry.
	tripleKey := collections.Join3(oldExpiry, msg.DelegatorAddress, msg.ValidatorAddress)
	_ = k.PendingExpirations.Remove(ctx, tripleKey)

	// Set new expiry.
	newExpiry := currentEpoch + params.DelegationConfirmationPeriod

	if err := k.DelegationConfirmations.Set(ctx, pairKey, newExpiry); err != nil {
		return nil, err
	}
	newTripleKey := collections.Join3(newExpiry, msg.DelegatorAddress, msg.ValidatorAddress)
	if err := k.PendingExpirations.Set(ctx, newTripleKey); err != nil {
		return nil, err
	}

	// Emit event.
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeDelegationConfirmed,
		sdk.NewAttribute(types.AttributeKeyDelegator, msg.DelegatorAddress),
		sdk.NewAttribute(types.AttributeKeyValidatorAddress, msg.ValidatorAddress),
		sdk.NewAttribute(types.AttributeKeyExpiryEpoch, fmt.Sprintf("%d", newExpiry)),
	))

	return &types.MsgConfirmDelegationResponse{
		ExpiryEpoch: newExpiry,
	}, nil
}

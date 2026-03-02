package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/types"
)

func TestProcessExpiredDelegations_NoExpired(t *testing.T) {
	f := initFixture(t)

	// Set block height to epoch boundary (epoch 10).
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(10 * types.DefaultEpochLength)
	f.ctx = ctx

	// No pending expirations — should complete without error.
	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify no force-unbond event was emitted.
	requireNoEvent(t, f.ctx, types.EventTypeDelegationForceUnbonded)
}

func TestProcessExpiredDelegations_SingleExpired(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	// Set up delegation in staking.
	f.stakingKeeper.delegations[delegator.String()+"/"+validator.String()] = stakingtypes.Delegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
		Shares:           math.LegacyNewDec(1000),
	}

	// Set up validator in staking (needed for TokensFromShares).
	f.stakingKeeper.validatorMap[validator.String()] = stakingtypes.Validator{
		OperatorAddress: validator.String(),
		Tokens:          math.NewInt(1000),
		DelegatorShares: math.LegacyNewDec(1000),
	}

	// Set up confirmation record with expiry at epoch 10.
	expiryEpoch := uint64(10)
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, collections.Join(delegator.String(), validator.String()), expiryEpoch)
	_ = f.keeper.PendingExpirations.Set(f.ctx, collections.Join3(expiryEpoch, delegator.String(), validator.String()))

	// Set block height to epoch 10 (expiry epoch).
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(int64(expiryEpoch) * types.DefaultEpochLength)
	f.ctx = ctx

	// Process — should force unbond.
	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("EndBlocker failed: %v", err)
	}

	// Verify confirmation record was removed.
	has, _ := f.keeper.DelegationConfirmations.Has(f.ctx, collections.Join(delegator.String(), validator.String()))
	if has {
		t.Fatal("confirmation record should have been removed")
	}

	// Verify pending expiration was removed.
	has, _ = f.keeper.PendingExpirations.Has(f.ctx, collections.Join3(expiryEpoch, delegator.String(), validator.String()))
	if has {
		t.Fatal("pending expiration should have been removed")
	}

	// Verify force unbond event was emitted.
	requireEvent(t, f.ctx, types.EventTypeDelegationForceUnbonded)
}

func TestProcessExpiredDelegations_AlreadyUnbonded(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	// NO delegation in staking (already manually unbonded).

	// Set up confirmation record with expiry at epoch 10.
	expiryEpoch := uint64(10)
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, collections.Join(delegator.String(), validator.String()), expiryEpoch)
	_ = f.keeper.PendingExpirations.Set(f.ctx, collections.Join3(expiryEpoch, delegator.String(), validator.String()))

	// Set block height to epoch 10 (expiry epoch).
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(int64(expiryEpoch) * types.DefaultEpochLength)
	f.ctx = ctx

	// Process — should clean up without error (skip force unbond since delegation doesn't exist).
	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("EndBlocker failed: %v", err)
	}

	// Verify records were cleaned up.
	has, _ := f.keeper.DelegationConfirmations.Has(f.ctx, collections.Join(delegator.String(), validator.String()))
	if has {
		t.Fatal("confirmation record should have been removed")
	}

	// Verify no force unbond event (delegation was already gone).
	requireNoEvent(t, f.ctx, types.EventTypeDelegationForceUnbonded)
}

func TestProcessExpiredDelegations_MultipleExpired(t *testing.T) {
	f := initFixture(t)

	del1 := sdk.AccAddress([]byte("delegator1__________"))
	del2 := sdk.AccAddress([]byte("delegator2__________"))
	val1 := sdk.ValAddress([]byte("validator1__________"))

	// Set up delegations for both delegators.
	for _, del := range []sdk.AccAddress{del1, del2} {
		f.stakingKeeper.delegations[del.String()+"/"+val1.String()] = stakingtypes.Delegation{
			DelegatorAddress: del.String(),
			ValidatorAddress: val1.String(),
			Shares:           math.LegacyNewDec(500),
		}
	}

	// Set up validator.
	f.stakingKeeper.validatorMap[val1.String()] = stakingtypes.Validator{
		OperatorAddress: val1.String(),
		Tokens:          math.NewInt(1000),
		DelegatorShares: math.LegacyNewDec(1000),
	}

	// Set up two confirmations, both expiring at epoch 10.
	expiryEpoch := uint64(10)
	for _, del := range []sdk.AccAddress{del1, del2} {
		_ = f.keeper.DelegationConfirmations.Set(f.ctx, collections.Join(del.String(), val1.String()), expiryEpoch)
		_ = f.keeper.PendingExpirations.Set(f.ctx, collections.Join3(expiryEpoch, del.String(), val1.String()))
	}

	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(int64(expiryEpoch) * types.DefaultEpochLength)
	f.ctx = ctx

	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("EndBlocker failed: %v", err)
	}

	// Both confirmation records should be cleaned up.
	for _, del := range []sdk.AccAddress{del1, del2} {
		has, _ := f.keeper.DelegationConfirmations.Has(f.ctx, collections.Join(del.String(), val1.String()))
		if has {
			t.Fatalf("confirmation record for %s should have been removed", del.String())
		}
	}

	// Force unbond event should be emitted.
	requireEvent(t, f.ctx, types.EventTypeDelegationForceUnbonded)
}

func TestProcessExpiredDelegations_NotAtEpochBoundary(t *testing.T) {
	f := initFixture(t)

	// Set up a pending expiration at epoch 1.
	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, collections.Join(delegator.String(), validator.String()), uint64(1))
	_ = f.keeper.PendingExpirations.Set(f.ctx, collections.Join3(uint64(1), delegator.String(), validator.String()))

	// Set block height NOT at epoch boundary.
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(types.DefaultEpochLength + 1) // epoch 1 + 1 block
	f.ctx = ctx

	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("EndBlocker failed: %v", err)
	}

	// Records should still exist (EndBlocker skips non-epoch boundaries).
	has, _ := f.keeper.DelegationConfirmations.Has(f.ctx, collections.Join(delegator.String(), validator.String()))
	if !has {
		t.Fatal("confirmation should still exist — not at epoch boundary")
	}
}

func TestEmitRenewalWindowAlerts(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	params, _ := f.keeper.Params.Get(f.ctx)
	renewalWindow := params.DelegationRenewalWindow // 7

	// Set expiry at epoch 17 (currentEpoch=10 + renewalWindow=7 = 17).
	// When EndBlocker runs at epoch 10, it checks for targetExpiry == 10 + 7 = 17.
	currentEpoch := uint64(10)
	expiryEpoch := currentEpoch + renewalWindow
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, collections.Join(delegator.String(), validator.String()), expiryEpoch)
	_ = f.keeper.PendingExpirations.Set(f.ctx, collections.Join3(expiryEpoch, delegator.String(), validator.String()))

	// Set block height to epoch 10.
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(int64(currentEpoch) * types.DefaultEpochLength)
	f.ctx = ctx

	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("EndBlocker failed: %v", err)
	}

	// Verify renewal window alert event was emitted.
	requireEvent(t, f.ctx, types.EventTypeDelegationRenewalWindowOpen)
}

func TestEmitRenewalWindowAlerts_NoAlert(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	// Set expiry at epoch 20. Current epoch=10, renewalWindow=7.
	// targetExpiry = 10 + 7 = 17 ≠ 20, so no alert should be emitted.
	expiryEpoch := uint64(20)
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, collections.Join(delegator.String(), validator.String()), expiryEpoch)
	_ = f.keeper.PendingExpirations.Set(f.ctx, collections.Join3(expiryEpoch, delegator.String(), validator.String()))

	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(10 * types.DefaultEpochLength)
	f.ctx = ctx

	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("EndBlocker failed: %v", err)
	}

	// No renewal window alert should be emitted.
	requireNoEvent(t, f.ctx, types.EventTypeDelegationRenewalWindowOpen)
}

func TestProcessExpiredDelegations_DisabledWhenPeriodZero(t *testing.T) {
	f := initFixture(t)

	// Disable active delegation by setting period to 0.
	params, _ := f.keeper.Params.Get(f.ctx)
	params.DelegationConfirmationPeriod = 0
	params.DelegationRenewalWindow = 0
	_ = f.keeper.Params.Set(f.ctx, params)

	// Set up an expired confirmation record.
	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, collections.Join(delegator.String(), validator.String()), uint64(1))
	_ = f.keeper.PendingExpirations.Set(f.ctx, collections.Join3(uint64(1), delegator.String(), validator.String()))

	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(10 * types.DefaultEpochLength)
	f.ctx = ctx

	err := f.keeper.EndBlocker(f.ctx)
	if err != nil {
		t.Fatalf("EndBlocker failed: %v", err)
	}

	// Record should still exist (disabled — not processed).
	has, _ := f.keeper.DelegationConfirmations.Has(f.ctx, collections.Join(delegator.String(), validator.String()))
	if !has {
		t.Fatal("confirmation should still exist when active delegation is disabled")
	}
}

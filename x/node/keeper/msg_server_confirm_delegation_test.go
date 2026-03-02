package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

func TestConfirmDelegation_Success(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	// Set up delegation in staking.
	f.stakingKeeper.delegations[delegator.String()+"/"+validator.String()] = stakingtypes.Delegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
		Shares:           math.LegacyNewDec(1000),
	}

	params, _ := f.keeper.Params.Get(f.ctx)
	period := params.DelegationConfirmationPeriod
	window := params.DelegationRenewalWindow

	// Set up initial confirmation record: expiry at epoch (period).
	expiryEpoch := period // e.g., 90
	pairKey := collections.Join(delegator.String(), validator.String())
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, pairKey, expiryEpoch)
	tripleKey := collections.Join3(expiryEpoch, delegator.String(), validator.String())
	_ = f.keeper.PendingExpirations.Set(f.ctx, tripleKey)

	// Set block height to within renewal window (epoch = expiryEpoch - window).
	renewalStart := expiryEpoch - window
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(int64(renewalStart) * types.DefaultEpochLength)
	f.ctx = ctx

	// Call ConfirmDelegation.
	msgServer := keeper.NewMsgServerImpl(f.keeper)
	resp, err := msgServer.ConfirmDelegation(f.ctx, &types.MsgConfirmDelegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
	})
	if err != nil {
		t.Fatalf("ConfirmDelegation failed: %v", err)
	}

	expectedNewExpiry := renewalStart + period
	if resp.ExpiryEpoch != expectedNewExpiry {
		t.Fatalf("expected expiry %d, got %d", expectedNewExpiry, resp.ExpiryEpoch)
	}

	// Verify old expiration was removed.
	has, _ := f.keeper.PendingExpirations.Has(f.ctx, tripleKey)
	if has {
		t.Fatal("old expiration entry should have been removed")
	}

	// Verify new expiration exists.
	newTriple := collections.Join3(expectedNewExpiry, delegator.String(), validator.String())
	has, _ = f.keeper.PendingExpirations.Has(f.ctx, newTriple)
	if !has {
		t.Fatal("new expiration entry should exist")
	}

	// Verify confirmation record updated.
	got, _ := f.keeper.DelegationConfirmations.Get(f.ctx, pairKey)
	if got != expectedNewExpiry {
		t.Fatalf("expected confirmation expiry %d, got %d", expectedNewExpiry, got)
	}

	// Verify event emitted.
	requireEvent(t, f.ctx, types.EventTypeDelegationConfirmed)
}

func TestConfirmDelegation_OutsideRenewalWindow(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	// Set up delegation.
	f.stakingKeeper.delegations[delegator.String()+"/"+validator.String()] = stakingtypes.Delegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
		Shares:           math.LegacyNewDec(1000),
	}

	// Set up confirmation record with expiry at epoch 90.
	expiryEpoch := uint64(90)
	pairKey := collections.Join(delegator.String(), validator.String())
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, pairKey, expiryEpoch)
	tripleKey := collections.Join3(expiryEpoch, delegator.String(), validator.String())
	_ = f.keeper.PendingExpirations.Set(f.ctx, tripleKey)

	// Set block height to epoch 50 (well before renewal window starts at epoch 83).
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(50 * types.DefaultEpochLength)
	f.ctx = ctx

	msgServer := keeper.NewMsgServerImpl(f.keeper)
	_, err := msgServer.ConfirmDelegation(f.ctx, &types.MsgConfirmDelegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
	})
	if err == nil {
		t.Fatal("expected error for outside renewal window")
	}
	if !types.ErrNotInRenewalWindow.Is(err) {
		t.Fatalf("expected ErrNotInRenewalWindow, got: %v", err)
	}
}

func TestConfirmDelegation_AlreadyExpired(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	f.stakingKeeper.delegations[delegator.String()+"/"+validator.String()] = stakingtypes.Delegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
		Shares:           math.LegacyNewDec(1000),
	}

	// Set expiry at epoch 90.
	expiryEpoch := uint64(90)
	pairKey := collections.Join(delegator.String(), validator.String())
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, pairKey, expiryEpoch)

	// Set block height to epoch 91 (after expiry).
	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(91 * types.DefaultEpochLength)
	f.ctx = ctx

	msgServer := keeper.NewMsgServerImpl(f.keeper)
	_, err := msgServer.ConfirmDelegation(f.ctx, &types.MsgConfirmDelegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
	})
	if err == nil {
		t.Fatal("expected error for expired delegation")
	}
	if !types.ErrAlreadyExpired.Is(err) {
		t.Fatalf("expected ErrAlreadyExpired, got: %v", err)
	}
}

func TestConfirmDelegation_NoDelegation(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	// No delegation in staking — but set up confirmation record.
	pairKey := collections.Join(delegator.String(), validator.String())
	_ = f.keeper.DelegationConfirmations.Set(f.ctx, pairKey, uint64(90))

	ctx := sdk.UnwrapSDKContext(f.ctx)
	ctx = ctx.WithBlockHeight(84 * types.DefaultEpochLength)
	f.ctx = ctx

	msgServer := keeper.NewMsgServerImpl(f.keeper)
	_, err := msgServer.ConfirmDelegation(f.ctx, &types.MsgConfirmDelegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
	})
	if err == nil {
		t.Fatal("expected error for missing delegation")
	}
	if !types.ErrDelegationNotFound.Is(err) {
		t.Fatalf("expected ErrDelegationNotFound, got: %v", err)
	}
}

func TestConfirmDelegation_NoConfirmationRecord(t *testing.T) {
	f := initFixture(t)

	delegator := sdk.AccAddress([]byte("delegator1__________"))
	validator := sdk.ValAddress([]byte("validator1__________"))

	// Set up delegation but no confirmation record.
	f.stakingKeeper.delegations[delegator.String()+"/"+validator.String()] = stakingtypes.Delegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
		Shares:           math.LegacyNewDec(1000),
	}

	msgServer := keeper.NewMsgServerImpl(f.keeper)
	_, err := msgServer.ConfirmDelegation(f.ctx, &types.MsgConfirmDelegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
	})
	if err == nil {
		t.Fatal("expected error for no confirmation record")
	}
}

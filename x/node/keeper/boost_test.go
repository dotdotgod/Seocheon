package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/types"
)

func TestDistributeBoostPool_NoValidators(t *testing.T) {
	f := initFixture(t)

	// Set boost pool balance.
	boostPoolAddr := authtypes.NewModuleAddress(types.BoostPoolName)
	f.bankKeeper.balances[boostPoolAddr.String()] = sdk.NewCoins(
		sdk.NewCoin("uppyeo", math.NewInt(135_000_000_000_000)), // 13,500 KKOT
	)

	// Initialize boost distributed counter.
	if err := f.keeper.BoostPoolDistributed.Set(f.ctx, math.ZeroInt()); err != nil {
		t.Fatal(err)
	}

	// No validators set — should return nil without error.
	f.stakingKeeper.validators = []stakingtypes.Validator{}

	err := f.keeper.EndBlocker(setBlockHeight(f.ctx, types.DefaultEpochLength))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No distribution should have occurred.
	distributed, err := f.keeper.BoostPoolDistributed.Get(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !distributed.IsZero() {
		t.Errorf("expected zero distributed, got %s", distributed)
	}
}

func TestDistributeBoostPool_SingleValidator(t *testing.T) {
	f := initFixture(t)

	// Set boost pool balance: 730 KKOT (small pool for test, 1 KKOT/epoch).
	boostPoolAddr := authtypes.NewModuleAddress(types.BoostPoolName)
	poolAmount := math.NewInt(7_300_000_000_000) // 730 KKOT
	f.bankKeeper.balances[boostPoolAddr.String()] = sdk.NewCoins(
		sdk.NewCoin("uppyeo", poolAmount),
	)

	if err := f.keeper.BoostPoolDistributed.Set(f.ctx, math.ZeroInt()); err != nil {
		t.Fatal(err)
	}

	// Single validator.
	valAddr := sdk.AccAddress("validator1__________")
	f.stakingKeeper.validators = []stakingtypes.Validator{
		{OperatorAddress: valAddr.String()},
	}

	err := f.keeper.EndBlocker(setBlockHeight(f.ctx, types.DefaultEpochLength))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check distribution occurred.
	distributed, err := f.keeper.BoostPoolDistributed.Get(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if distributed.IsZero() {
		t.Error("expected non-zero distributed amount")
	}

	// Verify send was recorded.
	if len(f.bankKeeper.sentFromModule) == 0 {
		t.Error("expected SendCoinsFromModuleToAccount call")
	} else {
		record := f.bankKeeper.sentFromModule[0]
		if record.SenderModule != types.BoostPoolName {
			t.Errorf("expected sender module %s, got %s", types.BoostPoolName, record.SenderModule)
		}
	}

	// Check event emission.
	requireEvent(t, f.ctx, types.EventTypeBoostDistributed)
}

func TestDistributeBoostPool_MultipleValidators(t *testing.T) {
	f := initFixture(t)

	// Set boost pool balance.
	boostPoolAddr := authtypes.NewModuleAddress(types.BoostPoolName)
	poolAmount := math.NewInt(135_000_000_000_000) // 13,500 KKOT
	f.bankKeeper.balances[boostPoolAddr.String()] = sdk.NewCoins(
		sdk.NewCoin("uppyeo", poolAmount),
	)

	if err := f.keeper.BoostPoolDistributed.Set(f.ctx, math.ZeroInt()); err != nil {
		t.Fatal(err)
	}

	// 5 validators.
	f.stakingKeeper.validators = []stakingtypes.Validator{
		{OperatorAddress: sdk.AccAddress("val1________________").String()},
		{OperatorAddress: sdk.AccAddress("val2________________").String()},
		{OperatorAddress: sdk.AccAddress("val3________________").String()},
		{OperatorAddress: sdk.AccAddress("val4________________").String()},
		{OperatorAddress: sdk.AccAddress("val5________________").String()},
	}

	err := f.keeper.EndBlocker(setBlockHeight(f.ctx, types.DefaultEpochLength))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 5 send records (one per validator).
	if len(f.bankKeeper.sentFromModule) != 5 {
		t.Errorf("expected 5 sends, got %d", len(f.bankKeeper.sentFromModule))
	}

	// All sends should have equal amounts.
	if len(f.bankKeeper.sentFromModule) >= 2 {
		firstAmount := f.bankKeeper.sentFromModule[0].Amount
		for i := 1; i < len(f.bankKeeper.sentFromModule); i++ {
			if !f.bankKeeper.sentFromModule[i].Amount.Equal(firstAmount) {
				t.Errorf("send %d amount %s differs from first %s",
					i, f.bankKeeper.sentFromModule[i].Amount, firstAmount)
			}
		}
	}

	// Check event.
	event := requireEvent(t, f.ctx, types.EventTypeBoostDistributed)
	recipients := eventAttribute(event, types.AttributeKeyBoostRecipients)
	if recipients != "5" {
		t.Errorf("expected 5 recipients in event, got %s", recipients)
	}
}

func TestDistributeBoostPool_EmptyPool(t *testing.T) {
	f := initFixture(t)

	// Empty boost pool.
	boostPoolAddr := authtypes.NewModuleAddress(types.BoostPoolName)
	f.bankKeeper.balances[boostPoolAddr.String()] = sdk.NewCoins()

	if err := f.keeper.BoostPoolDistributed.Set(f.ctx, math.ZeroInt()); err != nil {
		t.Fatal(err)
	}

	f.stakingKeeper.validators = []stakingtypes.Validator{
		{OperatorAddress: sdk.AccAddress("val1________________").String()},
	}

	err := f.keeper.EndBlocker(setBlockHeight(f.ctx, types.DefaultEpochLength))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No sends should have occurred.
	if len(f.bankKeeper.sentFromModule) != 0 {
		t.Errorf("expected 0 sends for empty pool, got %d", len(f.bankKeeper.sentFromModule))
	}

	// No event should have been emitted.
	requireNoEvent(t, f.ctx, types.EventTypeBoostDistributed)
}

func TestDistributeBoostPool_NotAtEpochBoundary(t *testing.T) {
	f := initFixture(t)

	// Set boost pool balance.
	boostPoolAddr := authtypes.NewModuleAddress(types.BoostPoolName)
	f.bankKeeper.balances[boostPoolAddr.String()] = sdk.NewCoins(
		sdk.NewCoin("uppyeo", math.NewInt(135_000_000_000_000)),
	)

	if err := f.keeper.BoostPoolDistributed.Set(f.ctx, math.ZeroInt()); err != nil {
		t.Fatal(err)
	}

	f.stakingKeeper.validators = []stakingtypes.Validator{
		{OperatorAddress: sdk.AccAddress("val1________________").String()},
	}

	// Not at epoch boundary — should not distribute.
	err := f.keeper.EndBlocker(setBlockHeight(f.ctx, types.DefaultEpochLength+1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.bankKeeper.sentFromModule) != 0 {
		t.Error("expected no sends at non-epoch boundary")
	}
}

// setBlockHeight returns a new context with the specified block height.
func setBlockHeight(ctx context.Context, height int64) context.Context {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.WithBlockHeight(height)
}

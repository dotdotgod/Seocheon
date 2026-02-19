package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// FeegrantPeriod is the period for the agent feegrant allowance (~1 epoch = ~1 day).
const FeegrantPeriod = 24 * time.Hour

// FeegrantExpiry is the expiration duration for the agent feegrant (6 months).
const FeegrantExpiry = 180 * 24 * time.Hour

// FeegrantPeriodLimit is 1 KKOT = 1,000,000 usum per epoch.
var FeegrantPeriodLimit = math.NewInt(1_000_000)

// grantAgentFeegrant creates a PeriodicAllowance feegrant from the Feegrant Pool
// to the given agent address. This is best-effort: if feegrantKeeper is not wired
// or the granting fails, a warning event is emitted but the caller is not blocked.
func (k Keeper) grantAgentFeegrant(ctx context.Context, agentAddr string) error {
	if k.feegrantKeeper == nil || agentAddr == "" {
		return nil
	}

	feegrantPoolAddr := k.authKeeper.GetModuleAddress(types.FeegrantPoolName)
	if feegrantPoolAddr == nil {
		return nil
	}

	agentAccAddr, err := sdk.AccAddressFromBech32(agentAddr)
	if err != nil {
		return errorsmod.Wrap(err, "invalid agent address for feegrant")
	}

	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return errorsmod.Wrap(err, "failed to get bond denom for feegrant")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	expiry := sdkCtx.BlockTime().Add(FeegrantExpiry)
	periodLimit := sdk.NewCoins(sdk.NewCoin(bondDenom, FeegrantPeriodLimit))

	allowance := &feegrant.PeriodicAllowance{
		Basic: feegrant.BasicAllowance{
			Expiration: &expiry,
		},
		Period:           FeegrantPeriod,
		PeriodSpendLimit: periodLimit,
		PeriodCanSpend:   periodLimit,
		PeriodReset:      sdkCtx.BlockTime().Add(FeegrantPeriod),
	}

	if err := k.feegrantKeeper.GrantAllowance(ctx, feegrantPoolAddr, agentAccAddr, allowance); err != nil {
		// Emit warning event but don't fail the operation.
		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			"feegrant_grant_failed",
			sdk.NewAttribute("agent_address", agentAddr),
			sdk.NewAttribute("error", err.Error()),
		))
		return nil
	}

	return nil
}

package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	proto "github.com/cosmos/gogoproto/proto"

	nodetypes "seocheon/x/node/types"
)

// Block-based constants (source of truth, matches document 02_core_concepts.md).
const (
	// BlockTime is the expected block production interval.
	BlockTime = 5 * time.Second

	// FeegrantPeriodBlocks is the feegrant quota reset period in blocks (1 epoch).
	FeegrantPeriodBlocks = 17_280

	// FeegrantExpiryBlocks is the feegrant expiration in blocks (~180 days).
	FeegrantExpiryBlocks = 3_110_400

	// FeegrantPeriod is derived from blocks for SDK x/feegrant PeriodicAllowance.
	FeegrantPeriod = time.Duration(FeegrantPeriodBlocks) * BlockTime

	// FeegrantExpiry is derived from blocks for SDK x/feegrant BasicAllowance.
	FeegrantExpiry = time.Duration(FeegrantExpiryBlocks) * BlockTime
)

// FeegrantPeriodLimit is 1 KKOT = 1,000,000 usum per epoch.
var FeegrantPeriodLimit = math.NewInt(1_000_000)

// grantAgentFeegrant creates a PeriodicAllowance feegrant from the Feegrant Pool
// to the given agent address. This is best-effort: if feegrantKeeper is not wired
// or the granting fails, a warning event is emitted but the caller is not blocked.
func (k Keeper) grantAgentFeegrant(ctx context.Context, agentAddr string) error {
	if k.feegrantKeeper == nil || agentAddr == "" {
		return nil
	}

	feegrantPoolAddr := k.authKeeper.GetModuleAddress(nodetypes.FeegrantPoolName)
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

	periodicAllowance := &feegrant.PeriodicAllowance{
		Basic: feegrant.BasicAllowance{
			Expiration: &expiry,
		},
		Period:           FeegrantPeriod,
		PeriodSpendLimit: periodLimit,
		PeriodCanSpend:   periodLimit,
		PeriodReset:      sdkCtx.BlockTime().Add(FeegrantPeriod),
	}

	// Wrap with AllowedMsgAllowance to restrict feegrant usage to specific message types.
	params, err := k.Params.Get(ctx)
	if err != nil {
		return errorsmod.Wrap(err, "failed to get params for feegrant allowed msg types")
	}
	allowance := &feegrant.AllowedMsgAllowance{
		Allowance:       mustPackAllowance(periodicAllowance),
		AllowedMessages: params.AgentFeegrantAllowedMsgTypes,
	}

	if err := k.feegrantKeeper.GrantAllowance(ctx, feegrantPoolAddr, agentAccAddr, allowance); err != nil {
		// Emit warning event but don't fail the operation.
		sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
			nodetypes.EventTypeFeegrantGrantFailed,
			sdk.NewAttribute(nodetypes.AttributeKeyAgentAddress, agentAddr),
			sdk.NewAttribute(nodetypes.AttributeKeyError, err.Error()),
		))
		return nil
	}

	return nil
}

// mustPackAllowance wraps a feegrant allowance into an Any for AllowedMsgAllowance.
func mustPackAllowance(allowance proto.Message) *types.Any {
	any, err := types.NewAnyWithValue(allowance)
	if err != nil {
		panic(err)
	}
	return any
}

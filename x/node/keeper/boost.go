package keeper

import (
	"context"
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// distributeBoostPool distributes tokens from the Validator Boost Pool
// to all bonded validators equally at each epoch boundary.
//
// Distribution formula: per_validator = pool_balance / target_epochs / num_validators
// The pool stops distributing when its balance reaches zero.
func (k Keeper) distributeBoostPool(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// 1. Get boost pool balance. If empty, nothing to do.
	boostPoolAddr := k.authKeeper.GetModuleAddress(types.BoostPoolName)
	if boostPoolAddr == nil {
		return nil
	}
	poolBalance := k.bankKeeper.GetAllBalances(ctx, boostPoolAddr)
	if poolBalance.IsZero() {
		return nil
	}

	// 2. Read the genesis boost pool balance and target epochs from genesis state.
	// We calculate epoch distribution as: genesis_balance / target_epochs.
	// Since we don't store genesis balance separately, we compute it from
	// cumulative distributed + current balance.
	distributed, err := k.BoostPoolDistributed.Get(ctx)
	if err != nil {
		// First distribution — no cumulative yet.
		distributed = math.ZeroInt()
	}

	// Total genesis balance = distributed + current pool balance (uppyeo).
	uppyeoBalance := poolBalance.AmountOf("uppyeo")
	if uppyeoBalance.IsZero() {
		return nil
	}
	genesisTotal := distributed.Add(uppyeoBalance)

	// Get target epochs from genesis defaults (stored in types).
	// We read boost_target_epochs from the genesis state that was passed during InitGenesis.
	// For simplicity, use the constant. This could be promoted to a governance param later.
	targetEpochs := math.NewIntFromUint64(types.DefaultBoostTargetEpochs)
	if targetEpochs.IsZero() {
		return nil
	}

	// Per-epoch amount = genesis_total / target_epochs.
	epochAmount := genesisTotal.Quo(targetEpochs)
	if epochAmount.IsZero() {
		return nil
	}

	// Cap at remaining balance.
	if epochAmount.GT(uppyeoBalance) {
		epochAmount = uppyeoBalance
	}

	// 3. Get all bonded validators.
	validators, err := k.stakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bonded validators: %w", err)
	}
	if len(validators) == 0 {
		return nil
	}

	// 4. Calculate per-validator amount (equal split).
	numValidators := math.NewInt(int64(len(validators)))
	perValidator := epochAmount.Quo(numValidators)
	if perValidator.IsZero() {
		return nil
	}

	// Actual total distributed this epoch (may be less than epochAmount due to integer division).
	actualDistributed := perValidator.Mul(numValidators)
	perValidatorCoins := sdk.NewCoins(sdk.NewCoin("uppyeo", perValidator))

	// 5. Send to each validator's operator address.
	for _, val := range validators {
		valAddr, err := sdk.AccAddressFromBech32(val.GetOperator())
		if err != nil {
			// Log and skip if operator address is invalid.
			sdkCtx.Logger().Error("boost pool: invalid validator operator address",
				"address", val.GetOperator(), "error", err)
			continue
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.BoostPoolName, valAddr, perValidatorCoins); err != nil {
			return fmt.Errorf("failed to send boost to validator %s: %w", val.GetOperator(), err)
		}
	}

	// 6. Update cumulative distributed amount.
	newDistributed := distributed.Add(actualDistributed)
	if err := k.BoostPoolDistributed.Set(ctx, newDistributed); err != nil {
		return err
	}

	// 7. Emit event.
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeBoostDistributed,
		sdk.NewAttribute(types.AttributeKeyBoostAmount, actualDistributed.String()),
		sdk.NewAttribute(types.AttributeKeyBoostRecipients, strconv.Itoa(len(validators))),
		sdk.NewAttribute(types.AttributeKeyBoostPerValidator, perValidator.String()),
	))

	return nil
}

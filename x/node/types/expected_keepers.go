package types

import (
	"context"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper defines the expected interface for the Staking module.
type StakingKeeper interface {
	ConsensusAddressCodec() address.Codec
	ValidatorByConsAddr(context.Context, sdk.ConsAddress) (stakingtypes.ValidatorI, error)
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	BondDenom(ctx context.Context) (string, error)
}

// AuthKeeper defines the expected interface for the Auth module.
type AuthKeeper interface {
	AddressCodec() address.Codec
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the expected interface for the Bank module.
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// DistributionKeeper defines the expected interface for the Distribution module.
type DistributionKeeper interface {
	WithdrawValidatorCommission(ctx context.Context, valAddr sdk.ValAddress) (sdk.Coins, error)
}

// SlashingKeeper defines the expected interface for the Slashing module.
type SlashingKeeper interface {
	IsTombstoned(ctx context.Context, consAddr sdk.ConsAddress) bool
}

// FeegrantKeeper defines the expected interface for the Feegrant module.
type FeegrantKeeper interface {
	GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error
}

// FeegrantMsgServer defines the interface for revoking feegrant allowances.
type FeegrantMsgServer interface {
	RevokeAllowance(context.Context, *feegrant.MsgRevokeAllowance) (*feegrant.MsgRevokeAllowanceResponse, error)
}

// ActivityKeeper defines the expected interface for the Activity module.
type ActivityKeeper interface {
	GetEpochLength(ctx context.Context) (int64, error)
	CountEligibleEpochs(ctx context.Context, nodeID string, currentEpoch int64, lookback int64) (int64, error)
}

// StakingMsgServer defines the interface for creating validators through x/staking.
type StakingMsgServer interface {
	CreateValidator(ctx context.Context, msg *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error)
	Undelegate(ctx context.Context, msg *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error)
}

// ValidatorAddressCodec provides conversion between validator address formats.
type ValidatorAddressCodec interface {
	StringToBytes(text string) ([]byte, error)
	BytesToString(bz []byte) (string, error)
}

// AccountAddressCodec provides conversion between account address formats.
type AccountAddressCodec interface {
	StringToBytes(text string) ([]byte, error)
	BytesToString(bz []byte) (string, error)
}

// ParamSubspace defines the expected Subspace interface for parameters (legacy).
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}

// Ensure math import is used.
var _ = math.ZeroInt

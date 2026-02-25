package types

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// NodeKeeper defines the expected interface for the x/node module.
type NodeKeeper interface {
	// GetNodeIDByAgent returns the node_id associated with the given agent address.
	GetNodeIDByAgent(ctx context.Context, agentAddr string) (string, error)

	// GetNodeStatus returns the status for the given node_id.
	// Status must be REGISTERED(1) or ACTIVE(2) for activity submission.
	GetNodeStatus(ctx context.Context, nodeID string) (int32, error)

	// GetNodeOperatorAddress returns the operator wallet address for the given node_id.
	GetNodeOperatorAddress(ctx context.Context, nodeID string) (string, error)

	// GetNodeAgentAddress returns the agent wallet address for the given node_id.
	GetNodeAgentAddress(ctx context.Context, nodeID string) (string, error)

	// GetNodeAgentShare returns the agent_share percentage (0-100) for the given node_id.
	GetNodeAgentShare(ctx context.Context, nodeID string) (math.LegacyDec, error)
}

// AuthKeeper defines the expected interface for the Auth module.
type AuthKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
}

// FeegrantKeeper defines the expected interface for the Feegrant module.
// Used to determine if a node has a feegrant allowance (quota determination).
type FeegrantKeeper interface {
	GetAllowance(ctx context.Context, granter, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error)
}

// StakingKeeper defines the expected interface for the Staking module.
// Used to get the active validator set size (N_d) for fee calculations.
type StakingKeeper interface {
	GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error)
}

// BankKeeper defines the expected interface for the Bank module.
// Used for collecting and distributing activity fees and block rewards.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// DistributionKeeper defines the expected interface for the Distribution module.
// Used for funding the community pool with the non-activity portion of collected fees.
type DistributionKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

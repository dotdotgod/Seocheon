package types

import (
	"context"

	"cosmossdk.io/x/feegrant"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NodeKeeper defines the expected interface for the x/node module.
type NodeKeeper interface {
	// GetNodeIDByAgent returns the node_id associated with the given agent address.
	GetNodeIDByAgent(ctx context.Context, agentAddr string) (string, error)

	// GetNodeStatus returns the status for the given node_id.
	// Status must be REGISTERED(1) or ACTIVE(2) for activity submission.
	GetNodeStatus(ctx context.Context, nodeID string) (int32, error)
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

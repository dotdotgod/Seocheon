package keeper

import (
	"context"

	"seocheon/x/node/types"
)

// UpdateAgentAddress changes the node's agent wallet address.
// Full implementation in Phase 0-B.
func (k msgServer) UpdateAgentAddress(ctx context.Context, msg *types.MsgUpdateAgentAddress) (*types.MsgUpdateAgentAddressResponse, error) {
	return nil, types.ErrNodeNotFound // TODO: Phase 0-B
}

package keeper

import (
	"context"

	"seocheon/x/node/types"
)

// RenewFeegrant renews an expired feegrant for an eligible node.
// Full implementation in Phase 0-D.
func (k msgServer) RenewFeegrant(ctx context.Context, msg *types.MsgRenewFeegrant) (*types.MsgRenewFeegrantResponse, error) {
	return nil, types.ErrNodeNotFound // TODO: Phase 0-D
}

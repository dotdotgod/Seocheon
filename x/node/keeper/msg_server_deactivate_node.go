package keeper

import (
	"context"

	"seocheon/x/node/types"
)

// DeactivateNode deactivates a node and begins unbonding.
// Full implementation in Phase 0-B.
func (k msgServer) DeactivateNode(ctx context.Context, msg *types.MsgDeactivateNode) (*types.MsgDeactivateNodeResponse, error) {
	return nil, types.ErrNodeNotFound // TODO: Phase 0-B
}

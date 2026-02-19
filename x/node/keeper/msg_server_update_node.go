package keeper

import (
	"context"

	"seocheon/x/node/types"
)

// UpdateNode updates node metadata (description, website, tags).
// Full implementation in Phase 0-B.
func (k msgServer) UpdateNode(ctx context.Context, msg *types.MsgUpdateNode) (*types.MsgUpdateNodeResponse, error) {
	return nil, types.ErrNodeNotFound // TODO: Phase 0-B
}

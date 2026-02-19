package keeper

import (
	"context"

	"seocheon/x/node/types"
)

// UpdateNodeAgentShare requests an agent_share change with rate limiting.
// Full implementation in Phase 0-B.
func (k msgServer) UpdateNodeAgentShare(ctx context.Context, msg *types.MsgUpdateNodeAgentShare) (*types.MsgUpdateNodeAgentShareResponse, error) {
	return nil, types.ErrNodeNotFound // TODO: Phase 0-B
}

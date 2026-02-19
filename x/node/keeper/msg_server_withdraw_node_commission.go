package keeper

import (
	"context"

	"seocheon/x/node/types"
)

// WithdrawNodeCommission withdraws commission and splits by agent_share.
// Full implementation in Phase 0-C.
func (k msgServer) WithdrawNodeCommission(ctx context.Context, msg *types.MsgWithdrawNodeCommission) (*types.MsgWithdrawNodeCommissionResponse, error) {
	return nil, types.ErrNodeNotFound // TODO: Phase 0-C
}

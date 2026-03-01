package keeper

import (
	"context"
	"fmt"

	"seocheon/x/randomness/types"
)

// UpdateParams updates the module parameters.
func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	authorityAddr, err := ms.k.addressCodec.BytesToString(ms.k.authority)
	if err != nil {
		return nil, fmt.Errorf("invalid authority address: %w", err)
	}

	if msg.Authority != authorityAddr {
		return nil, fmt.Errorf("%w: expected %s, got %s", types.ErrInvalidSigner, authorityAddr, msg.Authority)
	}

	if err := msg.Params.Validate(); err != nil {
		return nil, err
	}

	if err := ms.k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

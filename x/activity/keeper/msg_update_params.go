package keeper

import (
	"context"

	"seocheon/x/activity/types"
)

// UpdateParams handles MsgUpdateParams.
func (ms msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// Verify authority.
	authorityStr, err := ms.addressCodec.BytesToString(ms.authority)
	if err != nil {
		return nil, types.ErrInvalidAuthority.Wrap(err.Error())
	}
	if msg.Authority != authorityStr {
		return nil, types.ErrInvalidAuthority.Wrapf("expected %s, got %s", authorityStr, msg.Authority)
	}

	// Validate params.
	if err := msg.Params.Validate(); err != nil {
		return nil, types.ErrInvalidParams.Wrap(err.Error())
	}

	// Set params.
	if err := ms.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

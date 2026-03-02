package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"seocheon/x/node/types"
)

// DelegationConfirmation queries the confirmation status for a delegator-validator pair.
func (q queryServer) DelegationConfirmation(ctx context.Context, req *types.QueryDelegationConfirmationRequest) (*types.QueryDelegationConfirmationResponse, error) {
	if req == nil {
		return nil, types.ErrDelegationNotFound
	}

	pairKey := collections.Join(req.DelegatorAddress, req.ValidatorAddress)
	expiryEpoch, err := q.k.DelegationConfirmations.Get(ctx, pairKey)
	if err != nil {
		return nil, types.ErrDelegationNotFound.Wrapf("no confirmation record for %s -> %s", req.DelegatorAddress, req.ValidatorAddress)
	}

	currentEpoch := q.k.currentEpoch(ctx)

	params, err := q.k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	renewalWindowStart := uint64(0)
	inRenewalWindow := false
	if params.DelegationConfirmationPeriod > 0 && expiryEpoch > params.DelegationRenewalWindow {
		renewalWindowStart = expiryEpoch - params.DelegationRenewalWindow
		inRenewalWindow = currentEpoch >= renewalWindowStart && currentEpoch < expiryEpoch
	}

	return &types.QueryDelegationConfirmationResponse{
		ExpiryEpoch:        expiryEpoch,
		CurrentEpoch:       currentEpoch,
		InRenewalWindow:    inRenewalWindow,
		RenewalWindowStart: renewalWindowStart,
	}, nil
}

package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"

	"seocheon/x/randomness/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx context.Context, genState types.GenesisState) error {
	if err := k.Params.Set(ctx, genState.Params); err != nil {
		return err
	}

	var latestRound uint64
	for _, beacon := range genState.Beacons {
		if err := k.Beacons.Set(ctx, beacon.Round, beacon); err != nil {
			return err
		}
		if beacon.Round > latestRound {
			latestRound = beacon.Round
		}
	}

	if latestRound > 0 {
		if err := k.LatestRound.Set(ctx, latestRound); err != nil {
			return err
		}
	}

	// Initialize commit-reveal state.
	for _, req := range genState.RandomnessRequests {
		if err := k.Requests.Set(ctx, req.RequestId, req); err != nil {
			return fmt.Errorf("failed to store randomness request %d: %w", req.RequestId, err)
		}

		// Rebuild indexes.
		if err := k.RequestsByRequester.Set(ctx, collections.Join(req.Requester, req.RequestId)); err != nil {
			return fmt.Errorf("failed to set requester index for request %d: %w", req.RequestId, err)
		}

		switch req.Status {
		case types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING:
			if err := k.PendingByRound.Set(ctx, collections.Join(req.TargetRound, req.RequestId)); err != nil {
				return fmt.Errorf("failed to set pending index for request %d: %w", req.RequestId, err)
			}
		case types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_FULFILLED,
			types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_EXPIRED:
			if err := k.FulfilledBlockIndex.Set(ctx, collections.Join(req.FulfilledAt, req.RequestId)); err != nil {
				return fmt.Errorf("failed to set fulfilled index for request %d: %w", req.RequestId, err)
			}
		}
	}

	// Set pending request count.
	pendingCount := uint64(0)
	for _, req := range genState.RandomnessRequests {
		if req.Status == types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING {
			pendingCount++
		}
	}
	if err := k.PendingRequestCount.Set(ctx, pendingCount); err != nil {
		return fmt.Errorf("failed to set pending request count: %w", err)
	}

	// Set next request ID via sequence. Sequence is 0-based, we use ID = sequence + 1.
	// So to get NextRequestId = N, we advance the sequence to N-1.
	if genState.NextRequestId > 1 {
		for i := uint64(0); i < genState.NextRequestId-1; i++ {
			if _, err := k.RequestSequence.Next(ctx); err != nil {
				return fmt.Errorf("failed to advance request sequence: %w", err)
			}
		}
	}

	return nil
}

// ExportGenesis returns the module's exported genesis state.
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	var beacons []types.Beacon
	err = k.Beacons.Walk(ctx, nil, func(_ uint64, beacon types.Beacon) (bool, error) {
		beacons = append(beacons, beacon)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	var requests []types.RandomnessRequest
	err = k.Requests.Walk(ctx, nil, func(_ uint64, req types.RandomnessRequest) (bool, error) {
		requests = append(requests, req)
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	// Determine next request ID from sequence.
	nextID, err := k.RequestSequence.Peek(ctx)
	if err != nil {
		nextID = 0
	}
	nextRequestID := nextID + 1 // Sequence is 0-based, IDs are 1-based.

	return &types.GenesisState{
		Params:             params,
		Beacons:            beacons,
		RandomnessRequests: requests,
		NextRequestId:      nextRequestID,
	}, nil
}

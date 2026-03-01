package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"seocheon/x/randomness/types"
)

// RandomnessRequest queries a specific randomness request by ID.
func (qs queryServer) RandomnessRequest(ctx context.Context, req *types.QueryRandomnessRequestRequest) (*types.QueryRandomnessRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.RequestId == 0 {
		return nil, status.Error(codes.InvalidArgument, "request_id must be positive")
	}

	request, err := qs.k.Requests.Get(ctx, req.RequestId)
	if err != nil {
		return nil, types.ErrRequestNotFound
	}

	return &types.QueryRandomnessRequestResponse{Request: request}, nil
}

// PendingRequests queries all pending randomness requests with pagination.
func (qs queryServer) PendingRequests(ctx context.Context, req *types.QueryPendingRequestsRequest) (*types.QueryPendingRequestsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var requests []types.RandomnessRequest

	// Collect pending request IDs from PendingByRound index.
	var requestIDs []uint64
	err := qs.k.PendingByRound.Walk(ctx, nil, func(key collections.Pair[uint64, uint64]) (bool, error) {
		requestIDs = append(requestIDs, key.K2())
		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk pending requests: %w", err)
	}

	// Apply simple pagination.
	pageReq := req.Pagination
	offset := uint64(0)
	limit := uint64(100)
	if pageReq != nil {
		if pageReq.Offset > 0 {
			offset = pageReq.Offset
		}
		if pageReq.Limit > 0 {
			limit = pageReq.Limit
		}
	}

	total := uint64(len(requestIDs))
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	for _, id := range requestIDs[offset:end] {
		request, err := qs.k.Requests.Get(ctx, id)
		if err != nil {
			continue
		}
		if request.Status == types.RandomnessRequestStatus_RANDOMNESS_REQUEST_STATUS_PENDING {
			requests = append(requests, request)
		}
	}

	pageRes := &query.PageResponse{
		Total: total,
	}
	if end < total {
		pageRes.NextKey = []byte(fmt.Sprintf("%d", end))
	}

	return &types.QueryPendingRequestsResponse{
		Requests:   requests,
		Pagination: pageRes,
	}, nil
}

// RequestsByRequester queries randomness requests by requester address with pagination.
func (qs queryServer) RequestsByRequester(ctx context.Context, req *types.QueryRequestsByRequesterRequest) (*types.QueryRequestsByRequesterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Requester == "" {
		return nil, status.Error(codes.InvalidArgument, "requester address is required")
	}

	var requests []types.RandomnessRequest
	var requestIDs []uint64

	// Walk RequestsByRequester index for this requester.
	rng := collections.NewPrefixedPairRange[string, uint64](req.Requester)
	err := qs.k.RequestsByRequester.Walk(ctx, rng, func(key collections.Pair[string, uint64]) (bool, error) {
		requestIDs = append(requestIDs, key.K2())
		return false, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk requester requests: %w", err)
	}

	// Apply simple pagination.
	pageReq := req.Pagination
	offset := uint64(0)
	limit := uint64(100)
	if pageReq != nil {
		if pageReq.Offset > 0 {
			offset = pageReq.Offset
		}
		if pageReq.Limit > 0 {
			limit = pageReq.Limit
		}
	}

	total := uint64(len(requestIDs))
	end := offset + limit
	if end > total {
		end = total
	}
	if offset > total {
		offset = total
	}

	for _, id := range requestIDs[offset:end] {
		request, err := qs.k.Requests.Get(ctx, id)
		if err != nil {
			continue
		}
		requests = append(requests, request)
	}

	pageRes := &query.PageResponse{
		Total: total,
	}
	if end < total {
		pageRes.NextKey = []byte(fmt.Sprintf("%d", end))
	}

	return &types.QueryRequestsByRequesterResponse{
		Requests:   requests,
		Pagination: pageRes,
	}, nil
}

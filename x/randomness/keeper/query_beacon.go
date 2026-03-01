package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"seocheon/x/randomness/types"
)

// LatestBeacon queries the most recently stored beacon.
func (qs queryServer) LatestBeacon(ctx context.Context, req *types.QueryLatestBeaconRequest) (*types.QueryLatestBeaconResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	beacon, err := qs.k.GetLatestBeacon(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryLatestBeaconResponse{Beacon: beacon}, nil
}

// Beacon queries a beacon by its round number.
func (qs queryServer) Beacon(ctx context.Context, req *types.QueryBeaconRequest) (*types.QueryBeaconResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Round == 0 {
		return nil, status.Error(codes.InvalidArgument, "round must be positive")
	}

	beacon, err := qs.k.GetBeaconByRound(ctx, req.Round)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryBeaconResponse{Beacon: beacon}, nil
}

// Beacons queries beacons with pagination.
func (qs queryServer) Beacons(ctx context.Context, req *types.QueryBeaconsRequest) (*types.QueryBeaconsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	beacons, pageResp, err := query.CollectionPaginate(
		ctx,
		qs.k.Beacons,
		req.Pagination,
		func(_ uint64, beacon types.Beacon) (types.Beacon, error) {
			return beacon, nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBeaconsResponse{
		Beacons:    beacons,
		Pagination: pageResp,
	}, nil
}

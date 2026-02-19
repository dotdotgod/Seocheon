package keeper

import (
	"context"

	"cosmossdk.io/collections"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"seocheon/x/node/types"
)

func (q queryServer) Node(ctx context.Context, req *types.QueryNodeRequest) (*types.QueryNodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	node, err := q.k.Nodes.Get(ctx, req.NodeId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "node not found")
	}

	return &types.QueryNodeResponse{Node: node}, nil
}

func (q queryServer) NodeByOperator(ctx context.Context, req *types.QueryNodeByOperatorRequest) (*types.QueryNodeByOperatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Operator == "" {
		return nil, status.Error(codes.InvalidArgument, "operator is required")
	}

	nodeID, err := q.k.OperatorIndex.Get(ctx, req.Operator)
	if err != nil {
		return nil, status.Error(codes.NotFound, "node not found for operator")
	}

	node, err := q.k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get node")
	}

	return &types.QueryNodeByOperatorResponse{Node: node}, nil
}

func (q queryServer) NodeByAgentAddress(ctx context.Context, req *types.QueryNodeByAgentAddressRequest) (*types.QueryNodeByAgentAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.AgentAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "agent_address is required")
	}

	nodeID, err := q.k.AgentIndex.Get(ctx, req.AgentAddress)
	if err != nil {
		return nil, status.Error(codes.NotFound, "node not found for agent address")
	}

	node, err := q.k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get node")
	}

	return &types.QueryNodeByAgentAddressResponse{Node: node}, nil
}

func (q queryServer) NodesByTag(ctx context.Context, req *types.QueryNodesByTagRequest) (*types.QueryNodesByTagResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Tag == "" {
		return nil, status.Error(codes.InvalidArgument, "tag is required")
	}

	var nodes []types.Node

	rng := collections.NewPrefixedPairRange[string, string](req.Tag)
	err := q.k.TagIndex.Walk(ctx, rng, func(key collections.Pair[string, string]) (bool, error) {
		nodeID := key.K2()
		node, err := q.k.Nodes.Get(ctx, nodeID)
		if err != nil {
			return true, nil // skip missing nodes
		}
		nodes = append(nodes, node)
		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to query nodes by tag")
	}

	return &types.QueryNodesByTagResponse{Nodes: nodes}, nil
}

func (q queryServer) AllNodes(ctx context.Context, req *types.QueryAllNodesRequest) (*types.QueryAllNodesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var nodes []types.Node

	err := q.k.Nodes.Walk(ctx, nil, func(key string, node types.Node) (bool, error) {
		nodes = append(nodes, node)
		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list nodes")
	}

	return &types.QueryAllNodesResponse{Nodes: nodes}, nil
}

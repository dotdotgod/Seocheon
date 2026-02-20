package keeper

import (
	"context"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// UpdateNode updates node metadata (description, website, tags).
func (k msgServer) UpdateNode(ctx context.Context, msg *types.MsgUpdateNode) (*types.MsgUpdateNodeResponse, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get params")
	}

	// Look up node by operator.
	nodeID, err := k.OperatorIndex.Get(ctx, msg.Operator)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrNodeNotFound, "no node found for operator %s", msg.Operator)
	}

	node, err := k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrNodeNotFound, "node %s not found", nodeID)
	}

	// Verify the node is not inactive.
	if node.Status == types.NodeStatus_NODE_STATUS_INACTIVE {
		return nil, errorsmod.Wrap(types.ErrNodeInactive, "cannot update an inactive node")
	}

	// Validate tags.
	if uint32(len(msg.Tags)) > params.MaxTags {
		return nil, errorsmod.Wrapf(types.ErrInvalidTags, "too many tags: %d > %d", len(msg.Tags), params.MaxTags)
	}
	for _, tag := range msg.Tags {
		if uint32(len(tag)) > params.MaxTagLength {
			return nil, errorsmod.Wrapf(types.ErrInvalidTags, "tag too long: %d > %d", len(tag), params.MaxTagLength)
		}
		if tag == "" {
			return nil, errorsmod.Wrap(types.ErrInvalidTags, "empty tag not allowed")
		}
	}

	// Remove old tag indexes.
	for _, oldTag := range node.Tags {
		tagKey := collections.Join(oldTag, nodeID)
		if err := k.TagIndex.Remove(ctx, tagKey); err != nil {
			return nil, errorsmod.Wrap(err, "failed to remove old tag index")
		}
	}

	// Update node fields.
	node.Description = msg.Description
	node.Website = msg.Website
	node.Tags = msg.Tags

	if err := k.Nodes.Set(ctx, nodeID, node); err != nil {
		return nil, errorsmod.Wrap(err, "failed to update node")
	}

	// Set new tag indexes.
	for _, tag := range msg.Tags {
		tagKey := collections.Join(tag, nodeID)
		if err := k.TagIndex.Set(ctx, tagKey); err != nil {
			return nil, errorsmod.Wrap(err, "failed to set tag index")
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeNodeUpdated,
		sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
		sdk.NewAttribute(types.AttributeKeyOperator, msg.Operator),
	))

	return &types.MsgUpdateNodeResponse{}, nil
}

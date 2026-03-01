// Package activity implements the Activity module for the Seocheon SDK.
// It provides methods for submitting activities, querying activity records,
// and checking activity quotas.
package activity

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/seocheon/sdk-go/constants"
	sdkerrors "github.com/seocheon/sdk-go/errors"
	"github.com/seocheon/sdk-go/internal/chain"
	"github.com/seocheon/sdk-go/internal/signing"
	"github.com/seocheon/sdk-go/types"
)

// Module provides activity-related operations.
type Module struct {
	client  chain.Client
	signer  signing.Service
	chainID string
}

// NewModule creates a new Activity module.
func NewModule(client chain.Client, signer signing.Service, chainID string) *Module {
	return &Module{
		client:  client,
		signer:  signer,
		chainID: chainID,
	}
}

// Submit submits an activity hash to the chain.
// activityHash must be a 64-character hex string (SHA-256).
// contentURI is the off-chain location of the Activity Report.
func (m *Module) Submit(ctx context.Context, activityHash, contentURI string) (*types.SubmitActivityResponse, error) {
	if err := validateActivityHash(activityHash); err != nil {
		return nil, err
	}
	if contentURI == "" {
		return nil, sdkerrors.ErrInvalidContentURI
	}

	// Build and broadcast MsgSubmitActivity
	// The actual TX building/signing/broadcasting is delegated to the chain client
	// In a full implementation, this would:
	// 1. Create MsgSubmitActivity proto message
	// 2. Build TX with proper account info
	// 3. Sign via signing service
	// 4. Broadcast and wait for confirmation
	// 5. Compute derived fields
	_ = m.signer.GetAddress() // submitter

	// TODO: Implement full TX flow when proto dependencies are available
	return nil, fmt.Errorf("activity.Submit: TX flow requires proto message builders (not yet implemented)")
}

// GetActivities returns activity submission history for a node.
// If nodeID is empty, queries the signer's own node.
// If epochNumber is 0, queries the current epoch.
func (m *Module) GetActivities(ctx context.Context, nodeID string, epochNumber int64) (*types.GetActivitiesResponse, error) {
	effectiveNodeID := nodeID
	if effectiveNodeID == "" {
		id, err := m.resolveOwnNodeID(ctx)
		if err != nil {
			return nil, err
		}
		effectiveNodeID = id
	}

	effectiveEpoch := epochNumber
	if effectiveEpoch == 0 {
		e, err := m.computeCurrentEpoch(ctx)
		if err != nil {
			return nil, err
		}
		effectiveEpoch = e
	}

	path := fmt.Sprintf("/seocheon/activity/v1/nodes/%s/activities?epoch=%d", effectiveNodeID, effectiveEpoch)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	var result struct {
		Activities []struct {
			ActivityHash string      `json:"activity_hash"`
			ContentURI   string      `json:"content_uri"`
			BlockHeight  json.Number `json:"block_height"`
		} `json:"activities"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing activities response: %w", err)
	}

	params, err := m.getActivityParams(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]types.ActivityItem, 0, len(result.Activities))
	for _, a := range result.Activities {
		height, _ := a.BlockHeight.Int64()
		windowNum := computeWindow(height, params.epochLength, params.windowsPerEpoch)
		items = append(items, types.ActivityItem{
			ActivityHash: a.ActivityHash,
			ContentURI:   a.ContentURI,
			BlockHeight:  height,
			WindowNumber: windowNum,
		})
	}

	return &types.GetActivitiesResponse{
		Activities: items,
		TotalCount: uint64(len(items)),
	}, nil
}

// GetQuota returns the remaining activity submission quota for the current epoch.
func (m *Module) GetQuota(ctx context.Context) (*types.GetQuotaResponse, error) {
	nodeID, err := m.resolveOwnNodeID(ctx)
	if err != nil {
		return nil, err
	}

	epochNum, err := m.computeCurrentEpoch(ctx)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/seocheon/activity/v1/nodes/%s/epochs/%d", nodeID, epochNum)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	var result struct {
		QuotaUsed  json.Number `json:"quota_used"`
		QuotaLimit json.Number `json:"quota_limit"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing quota response: %w", err)
	}

	used, _ := result.QuotaUsed.Int64()
	limit, _ := result.QuotaLimit.Int64()

	// Check feegrant status
	agentAddr := m.signer.GetAddress()
	isFeegrant, feegrantExpiry := m.checkFeegrant(ctx, agentAddr)

	return &types.GetQuotaResponse{
		EpochNumber:    epochNum,
		QuotaTotal:     uint64(limit),
		QuotaUsed:      uint64(used),
		QuotaRemaining: uint64(limit - used),
		IsFeegrant:     isFeegrant,
		FeegrantExpiry: feegrantExpiry,
	}, nil
}

// resolveOwnNodeID resolves the node ID from the signer's agent address.
func (m *Module) resolveOwnNodeID(ctx context.Context) (string, error) {
	agentAddr := m.signer.GetAddress()
	path := fmt.Sprintf("/seocheon/node/v1/nodes/by-agent/%s", agentAddr)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return "", fmt.Errorf("%w: %v", sdkerrors.ErrSubmitterNotRegistered, err)
	}

	var result struct {
		Node struct {
			ID string `json:"id"`
		} `json:"node"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("parsing node response: %w", err)
	}
	if result.Node.ID == "" {
		return "", sdkerrors.ErrSubmitterNotRegistered
	}
	return result.Node.ID, nil
}

type activityParams struct {
	epochLength     int64
	windowsPerEpoch int64
}

func (m *Module) getActivityParams(ctx context.Context) (*activityParams, error) {
	data, err := m.client.QueryREST(ctx, "/seocheon/activity/v1/params")
	if err != nil {
		return nil, fmt.Errorf("querying activity params: %w", err)
	}

	var result struct {
		Params struct {
			EpochLength     json.Number `json:"epoch_length"`
			WindowsPerEpoch json.Number `json:"windows_per_epoch"`
		} `json:"params"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing params response: %w", err)
	}

	el, _ := result.Params.EpochLength.Int64()
	wpe, _ := result.Params.WindowsPerEpoch.Int64()

	if el == 0 {
		el = constants.EpochLength
	}
	if wpe == 0 {
		wpe = constants.WindowsPerEpoch
	}

	return &activityParams{
		epochLength:     el,
		windowsPerEpoch: wpe,
	}, nil
}

func (m *Module) computeCurrentEpoch(ctx context.Context) (int64, error) {
	block, err := m.client.GetLatestBlock(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting latest block: %w", err)
	}

	params, err := m.getActivityParams(ctx)
	if err != nil {
		return 0, err
	}

	return computeEpoch(block.Height, params.epochLength), nil
}

func (m *Module) checkFeegrant(ctx context.Context, agentAddr string) (bool, *int64) {
	// Query feegrant allowance - non-critical, return defaults on error
	path := fmt.Sprintf("/cosmos/feegrant/v1beta1/allowance/%s/%s", "feegrant_pool", agentAddr)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return false, nil
	}

	var result struct {
		Allowance json.RawMessage `json:"allowance"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Allowance == nil {
		return false, nil
	}

	return true, nil
}

func computeEpoch(blockHeight, epochLength int64) int64 {
	if blockHeight <= 0 || epochLength <= 0 {
		return 0
	}
	return (blockHeight - 1) / epochLength
}

func computeWindow(blockHeight, epochLength, windowsPerEpoch int64) int64 {
	if epochLength <= 0 || windowsPerEpoch <= 0 {
		return 0
	}
	windowLength := epochLength / windowsPerEpoch
	return ((blockHeight - 1) % epochLength) / windowLength
}

func validateActivityHash(hash string) error {
	if len(hash) != 64 {
		return sdkerrors.ErrInvalidActivityHash
	}
	if _, err := hex.DecodeString(hash); err != nil {
		return sdkerrors.ErrInvalidActivityHash
	}
	return nil
}

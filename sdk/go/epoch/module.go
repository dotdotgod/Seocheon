// Package epoch implements the Epoch module for the Seocheon SDK.
// It provides methods for querying epoch/window state and activity qualification.
package epoch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/seocheon/sdk-go/constants"
	sdkerrors "github.com/seocheon/sdk-go/errors"
	"github.com/seocheon/sdk-go/internal/chain"
	"github.com/seocheon/sdk-go/types"
)

// Module provides epoch-related operations.
type Module struct {
	client chain.Client
}

// NewModule creates a new Epoch module.
func NewModule(client chain.Client) *Module {
	return &Module{client: client}
}

// GetInfo returns the current epoch and window state.
func (m *Module) GetInfo(ctx context.Context) (*types.EpochInfoResponse, error) {
	data, err := m.client.QueryREST(ctx, "/seocheon/activity/v1/epoch-info")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	var proto struct {
		CurrentEpoch        json.Number `json:"current_epoch"`
		CurrentWindow       json.Number `json:"current_window"`
		EpochStartBlock     json.Number `json:"epoch_start_block"`
		BlocksUntilNextEpoch json.Number `json:"blocks_until_next_epoch"`
	}
	if err := json.Unmarshal(data, &proto); err != nil {
		return nil, fmt.Errorf("parsing epoch info response: %w", err)
	}

	params, err := m.getParams(ctx)
	if err != nil {
		return nil, err
	}

	block, err := m.client.GetLatestBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting latest block: %w", err)
	}

	currentEpoch, _ := proto.CurrentEpoch.Int64()
	currentWindow, _ := proto.CurrentWindow.Int64()
	epochStartBlock, _ := proto.EpochStartBlock.Int64()
	blocksUntilNextEpoch, _ := proto.BlocksUntilNextEpoch.Int64()

	windowLength := params.epochLength / params.windowsPerEpoch
	epochEndBlock := epochStartBlock + params.epochLength - 1
	windowStartBlock := epochStartBlock + (currentWindow * windowLength)
	windowEndBlock := windowStartBlock + windowLength - 1

	epochElapsed := block.Height - epochStartBlock + 1
	windowElapsed := block.Height - windowStartBlock + 1

	return &types.EpochInfoResponse{
		BlockHeight:           block.Height,
		EpochNumber:           currentEpoch,
		EpochStartBlock:       epochStartBlock,
		EpochEndBlock:         epochEndBlock,
		EpochProgress:         fmt.Sprintf("%d/%d", epochElapsed, params.epochLength),
		WindowNumber:          currentWindow,
		WindowStartBlock:      windowStartBlock,
		WindowEndBlock:        windowEndBlock,
		WindowProgress:        fmt.Sprintf("%d/%d", windowElapsed, windowLength),
		BlocksUntilNextWindow: windowEndBlock - block.Height,
		BlocksUntilNextEpoch:  blocksUntilNextEpoch,
	}, nil
}

// GetQualification returns the activity reward qualification status for a node.
// If nodeID is empty, queries the current signer's node via agent address.
// If epochNumber is 0, queries the current epoch.
func (m *Module) GetQualification(ctx context.Context, nodeID string, epochNumber int64) (*types.QualificationResponse, error) {
	if nodeID == "" {
		return nil, fmt.Errorf("node_id is required for qualification query")
	}

	effectiveEpoch := epochNumber
	if effectiveEpoch == 0 {
		info, err := m.GetInfo(ctx)
		if err != nil {
			return nil, err
		}
		effectiveEpoch = info.EpochNumber
	}

	// Query NodeEpochActivity
	path := fmt.Sprintf("/seocheon/activity/v1/nodes/%s/epochs/%d", nodeID, effectiveEpoch)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	var result struct {
		Summary struct {
			TotalActivities json.Number `json:"total_activities"`
			ActiveWindows   json.Number `json:"active_windows"`
			Eligible        bool        `json:"eligible"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing node epoch activity response: %w", err)
	}

	params, err := m.getParams(ctx)
	if err != nil {
		return nil, err
	}

	activeWindows, _ := result.Summary.ActiveWindows.Int64()

	// Compute elapsed windows
	var elapsedWindows int64
	epochInfo, err := m.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	if effectiveEpoch == epochInfo.EpochNumber {
		elapsedWindows = epochInfo.WindowNumber + 1
	} else {
		elapsedWindows = params.windowsPerEpoch
	}

	remainingNeeded := params.minActiveWindows - activeWindows
	if remainingNeeded < 0 {
		remainingNeeded = 0
	}

	remainingWindows := params.windowsPerEpoch - elapsedWindows
	canStillQualify := (activeWindows + remainingWindows) >= params.minActiveWindows

	// Build window detail from activities
	windowDetail := make([]types.WindowActivity, params.windowsPerEpoch)
	for w := int64(0); w < params.windowsPerEpoch; w++ {
		windowDetail[w] = types.WindowActivity{
			WindowNumber:    w,
			SubmissionCount: 0,
			HasActivity:     false,
		}
	}

	// Enrich window detail from activities query
	actPath := fmt.Sprintf("/seocheon/activity/v1/nodes/%s/activities?epoch=%d", nodeID, effectiveEpoch)
	actData, err := m.client.QueryREST(ctx, actPath)
	if err == nil {
		var actResult struct {
			Activities []struct {
				BlockHeight json.Number `json:"block_height"`
			} `json:"activities"`
		}
		if json.Unmarshal(actData, &actResult) == nil {
			windowLength := params.epochLength / params.windowsPerEpoch
			for _, act := range actResult.Activities {
				h, _ := act.BlockHeight.Int64()
				wn := ((h - 1) % params.epochLength) / windowLength
				if wn >= 0 && wn < params.windowsPerEpoch {
					windowDetail[wn].SubmissionCount++
					windowDetail[wn].HasActivity = true
				}
			}
		}
	}

	return &types.QualificationResponse{
		EpochNumber:     effectiveEpoch,
		TotalWindows:    params.windowsPerEpoch,
		ElapsedWindows:  elapsedWindows,
		ActiveWindows:   uint64(activeWindows),
		RequiredWindows: params.minActiveWindows,
		IsQualified:     result.Summary.Eligible,
		RemainingNeeded: remainingNeeded,
		CanStillQualify: canStillQualify,
		WindowDetail:    windowDetail,
	}, nil
}

type epochParams struct {
	epochLength      int64
	windowsPerEpoch  int64
	minActiveWindows int64
}

func (m *Module) getParams(ctx context.Context) (*epochParams, error) {
	data, err := m.client.QueryREST(ctx, "/seocheon/activity/v1/params")
	if err != nil {
		return nil, fmt.Errorf("querying activity params: %w", err)
	}

	var result struct {
		Params struct {
			EpochLength      json.Number `json:"epoch_length"`
			WindowsPerEpoch  json.Number `json:"windows_per_epoch"`
			MinActiveWindows json.Number `json:"min_active_windows"`
		} `json:"params"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing params: %w", err)
	}

	el, _ := result.Params.EpochLength.Int64()
	wpe, _ := result.Params.WindowsPerEpoch.Int64()
	maw, _ := result.Params.MinActiveWindows.Int64()

	if el == 0 {
		el = constants.EpochLength
	}
	if wpe == 0 {
		wpe = constants.WindowsPerEpoch
	}
	if maw == 0 {
		maw = constants.MinActiveWindows
	}

	return &epochParams{
		epochLength:      el,
		windowsPerEpoch:  wpe,
		minActiveWindows: maw,
	}, nil
}

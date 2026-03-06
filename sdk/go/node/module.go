// Package node implements the Node module for the Seocheon SDK.
// It provides methods for querying node information and searching nodes.
package node

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	sdkerrors "github.com/seocheon/sdk-go/errors"
	"github.com/seocheon/sdk-go/internal/chain"
	"github.com/seocheon/sdk-go/internal/signing"
	"github.com/seocheon/sdk-go/internal/tx"
	"github.com/seocheon/sdk-go/types"
	"github.com/seocheon/sdk-go/utility"
)

// Module provides node-related operations.
type Module struct {
	client    chain.Client
	signer    signing.Service
	txQuerier tx.ChainQuerier
	txConfig  tx.PipelineConfig
}

// NewModule creates a new Node module.
func NewModule(client chain.Client, signer signing.Service) *Module {
	return &Module{
		client:    client,
		signer:    signer,
		txQuerier: tx.NewChainClientAdapter(client),
		txConfig:  tx.DefaultPipelineConfig("seocheon-1"),
	}
}

// GetInfo returns detailed information about a node.
// If nodeID is empty, queries the signer's own node.
func (m *Module) GetInfo(ctx context.Context, nodeID string) (*types.NodeInfoResponse, error) {
	effectiveNodeID := nodeID
	if effectiveNodeID == "" {
		id, err := m.resolveOwnNodeID(ctx)
		if err != nil {
			return nil, err
		}
		effectiveNodeID = id
	}

	// Query node
	path := fmt.Sprintf("/seocheon/node/v1/nodes/%s", effectiveNodeID)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrNodeNotFound, err)
	}

	var result struct {
		Node nodeProto `json:"node"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing node response: %w", err)
	}

	n := result.Node

	// Query staking info for total_delegation and self_delegation
	totalDelegation := "0"
	selfDelegation := "0"
	commissionRate := "0"

	if n.ValidatorAddress != "" {
		totalDelegation, commissionRate = m.queryValidatorInfo(ctx, n.ValidatorAddress)
		selfDelegation = m.querySelfDelegation(ctx, n.Operator, n.ValidatorAddress)
	}

	registeredAt, _ := strconv.ParseInt(n.RegisteredAt, 10, 64)

	return &types.NodeInfoResponse{
		NodeID:           n.ID,
		Operator:         n.Operator,
		AgentAddress:     n.AgentAddress,
		Status:           string(types.NodeStatusFromString(n.Status)),
		Description:      n.Description,
		Website:          n.Website,
		Tags:             n.Tags,
		CommissionRate:   commissionRate,
		AgentShare:       n.AgentShare,
		TotalDelegation:  totalDelegation,
		SelfDelegation:   selfDelegation,
		ValidatorAddress: n.ValidatorAddress,
		RegisteredAt:     registeredAt,
	}, nil
}

// Search finds nodes matching the given criteria.
// If no parameters are specified, returns up to 20 nodes sorted by delegation.
func (m *Module) Search(ctx context.Context, tag, status string, limit uint32, orderBy string) (*types.NodeSearchResponse, error) {
	effectiveLimit := limit
	if effectiveLimit == 0 {
		effectiveLimit = 20
	}
	effectiveOrder := orderBy
	if effectiveOrder == "" {
		effectiveOrder = "delegation"
	}

	// Query nodes
	var path string
	if tag != "" {
		path = fmt.Sprintf("/seocheon/node/v1/nodes/by-tag/%s", tag)
	} else {
		path = "/seocheon/node/v1/nodes"
	}

	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	var result struct {
		Nodes []nodeProto `json:"nodes"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing nodes response: %w", err)
	}

	// Filter by status
	filtered := result.Nodes
	if status != "" {
		var statusFiltered []nodeProto
		for _, n := range filtered {
			if string(types.NodeStatusFromString(n.Status)) == status {
				statusFiltered = append(statusFiltered, n)
			}
		}
		filtered = statusFiltered
	}

	// Enrich with delegation info and build summaries
	type enrichedNode struct {
		summary    types.NodeSummary
		delegation int64
		regAt      int64
	}
	enriched := make([]enrichedNode, 0, len(filtered))
	for _, n := range filtered {
		delegation := int64(0)
		if n.ValidatorAddress != "" {
			delStr, _ := m.queryValidatorInfo(ctx, n.ValidatorAddress)
			delegation, _ = strconv.ParseInt(delStr, 10, 64)
		}
		regAt, _ := strconv.ParseInt(n.RegisteredAt, 10, 64)
		enriched = append(enriched, enrichedNode{
			summary: types.NodeSummary{
				NodeID:          n.ID,
				Status:          string(types.NodeStatusFromString(n.Status)),
				Tags:            n.Tags,
				TotalDelegation: utility.FormatKkot(delegation),
				Description:     n.Description,
			},
			delegation: delegation,
			regAt:      regAt,
		})
	}

	// Sort
	if effectiveOrder == "delegation" {
		sort.Slice(enriched, func(i, j int) bool {
			return enriched[i].delegation > enriched[j].delegation
		})
	} else {
		sort.Slice(enriched, func(i, j int) bool {
			return enriched[i].regAt > enriched[j].regAt
		})
	}

	// Apply limit
	totalCount := uint64(len(enriched))
	if uint32(len(enriched)) > effectiveLimit {
		enriched = enriched[:effectiveLimit]
	}

	summaries := make([]types.NodeSummary, len(enriched))
	for i, e := range enriched {
		summaries[i] = e.summary
	}

	return &types.NodeSearchResponse{
		Nodes:      summaries,
		TotalCount: totalCount,
	}, nil
}

// resolveOwnNodeID resolves the node ID from the signer's agent address.
func (m *Module) resolveOwnNodeID(ctx context.Context) (string, error) {
	agentAddr := m.signer.GetAddress()
	path := fmt.Sprintf("/seocheon/node/v1/nodes/by-agent/%s", agentAddr)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return "", fmt.Errorf("%w: agent address %s", sdkerrors.ErrNodeNotFound, agentAddr)
	}

	var result struct {
		Node struct {
			ID string `json:"id"`
		} `json:"node"`
	}
	if err := json.Unmarshal(data, &result); err != nil || result.Node.ID == "" {
		return "", sdkerrors.ErrNodeNotFound
	}
	return result.Node.ID, nil
}

func (m *Module) queryValidatorInfo(ctx context.Context, valAddr string) (totalDelegation, commissionRate string) {
	path := fmt.Sprintf("/cosmos/staking/v1beta1/validators/%s", valAddr)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return "0", "0"
	}

	var result struct {
		Validator struct {
			Tokens     string `json:"tokens"`
			Commission struct {
				CommissionRates struct {
					Rate string `json:"rate"`
				} `json:"commission_rates"`
			} `json:"commission"`
		} `json:"validator"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "0", "0"
	}
	return result.Validator.Tokens, result.Validator.Commission.CommissionRates.Rate
}

func (m *Module) querySelfDelegation(ctx context.Context, operator, valAddr string) string {
	path := fmt.Sprintf("/cosmos/staking/v1beta1/validators/%s/delegations/%s", valAddr, operator)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return "0"
	}

	var result struct {
		DelegationResponse struct {
			Balance struct {
				Amount string `json:"amount"`
			} `json:"balance"`
		} `json:"delegation_response"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "0"
	}
	return result.DelegationResponse.Balance.Amount
}

type nodeProto struct {
	ID               string   `json:"id"`
	Operator         string   `json:"operator"`
	AgentAddress     string   `json:"agent_address"`
	Status           string   `json:"status"`
	Description      string   `json:"description"`
	Website          string   `json:"website"`
	Tags             []string `json:"tags"`
	AgentShare       string   `json:"agent_share"`
	ValidatorAddress string   `json:"validator_address"`
	RegisteredAt     string   `json:"registered_at"`
}

// GetDelegationStatus queries the delegation confirmation status.
func (m *Module) GetDelegationStatus(ctx context.Context, delegatorAddress, validatorAddress string) (*types.DelegationStatusResponse, error) {
	path := fmt.Sprintf("/seocheon/node/v1/delegation-confirmation/%s/%s", delegatorAddress, validatorAddress)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrQueryFailed, err)
	}

	var result struct {
		ExpiryEpoch        string `json:"expiry_epoch"`
		CurrentEpoch       string `json:"current_epoch"`
		InRenewalWindow    bool   `json:"in_renewal_window"`
		RenewalWindowStart string `json:"renewal_window_start"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing delegation status: %w", err)
	}

	expiryEpoch, _ := strconv.ParseInt(result.ExpiryEpoch, 10, 64)
	currentEpoch, _ := strconv.ParseInt(result.CurrentEpoch, 10, 64)
	renewalWindowStart, _ := strconv.ParseInt(result.RenewalWindowStart, 10, 64)

	return &types.DelegationStatusResponse{
		ExpiryEpoch:        expiryEpoch,
		CurrentEpoch:       currentEpoch,
		InRenewalWindow:    result.InRenewalWindow,
		RenewalWindowStart: renewalWindowStart,
	}, nil
}

// ConfirmDelegation sends MsgConfirmDelegation for a validator.
func (m *Module) ConfirmDelegation(ctx context.Context, validatorAddress string) (*types.TxResultResponse, error) {
	msg := &tx.MsgConfirmDelegation{
		DelegatorAddress: m.signer.GetAddress(),
		ValidatorAddress: validatorAddress,
	}

	result, err := tx.ExecuteTx(ctx, m.txQuerier, m.signer, m.txConfig, tx.TxRequest{
		Message: msg,
	})
	if err != nil {
		return nil, fmt.Errorf("confirming delegation: %w", err)
	}

	return &types.TxResultResponse{
		TxHash:    result.TxHash,
		Height:    result.Height,
		Code:      result.Code,
		GasUsed:   result.GasUsed,
		GasWanted: result.GasWanted,
		RawLog:    result.RawLog,
	}, nil
}

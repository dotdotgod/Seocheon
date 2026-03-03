// Package rewards implements the Rewards module for the Seocheon SDK.
// It provides methods for querying pending rewards and withdrawing them.
package rewards

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	sdkerrors "github.com/seocheon/sdk-go/errors"
	"github.com/seocheon/sdk-go/internal/chain"
	"github.com/seocheon/sdk-go/internal/signing"
	"github.com/seocheon/sdk-go/internal/tx"
	"github.com/seocheon/sdk-go/types"
	"github.com/seocheon/sdk-go/utility"
)

// Module provides reward-related operations.
type Module struct {
	client    chain.Client
	signer    signing.Service
	txQuerier tx.ChainQuerier
	txConfig  tx.PipelineConfig
}

// NewModule creates a new Rewards module.
func NewModule(client chain.Client, signer signing.Service, chainID string) *Module {
	return &Module{
		client:    client,
		signer:    signer,
		txQuerier: tx.NewChainClientAdapter(client),
		txConfig:  tx.DefaultPipelineConfig(chainID),
	}
}

// GetPending returns pending (unwithdrawn) rewards for a node.
// If nodeID is empty, queries the signer's own node.
func (m *Module) GetPending(ctx context.Context, nodeID string) (*types.PendingRewardsResponse, error) {
	effectiveNodeID := nodeID
	if effectiveNodeID == "" {
		id, err := m.resolveOwnNodeID(ctx)
		if err != nil {
			return nil, err
		}
		effectiveNodeID = id
	}

	// Get node info for validator_address and agent_share
	nodeInfo, err := m.getNode(ctx, effectiveNodeID)
	if err != nil {
		return nil, err
	}

	// Query outstanding rewards
	delegationReward := int64(0)
	if nodeInfo.ValidatorAddress != "" {
		delegationReward = m.queryOutstandingRewards(ctx, nodeInfo.ValidatorAddress)
	}

	// Query commission
	commissionTotal := int64(0)
	if nodeInfo.ValidatorAddress != "" {
		commissionTotal = m.queryCommission(ctx, nodeInfo.ValidatorAddress)
	}

	// Activity reward estimation (placeholder)
	activityReward := int64(0)

	totalReward := delegationReward + activityReward

	// Compute operator/agent shares
	agentShareRatio := parseAgentShare(nodeInfo.AgentShare)
	operatorShare := int64(float64(commissionTotal) * (1.0 - agentShareRatio))
	agentShare := commissionTotal - operatorShare

	return &types.PendingRewardsResponse{
		DelegationReward: utility.FormatKkot(delegationReward),
		ActivityReward:   utility.FormatKkot(activityReward),
		TotalReward:      utility.FormatKkot(totalReward),
		CommissionTotal:  utility.FormatKkot(commissionTotal),
		OperatorShare:    utility.FormatKkot(operatorShare),
		AgentShare:       utility.FormatKkot(agentShare),
	}, nil
}

// Withdraw withdraws all pending rewards. Requires operator signing key.
func (m *Module) Withdraw(ctx context.Context) (*types.WithdrawRewardsResponse, error) {
	operator := m.signer.GetAddress()

	msg := &tx.MsgWithdrawNodeCommission{
		Operator: operator,
	}

	result, err := tx.ExecuteTx(ctx, m.txQuerier, m.signer, m.txConfig, tx.TxRequest{
		Message: msg,
	})
	if err != nil {
		if result != nil && result.Code != 0 {
			return nil, sdkerrors.ABCICodeToError(result.Code)
		}
		return nil, fmt.Errorf("withdrawing rewards: %w", err)
	}

	// Parse withdrawn amounts from TX events
	withdrawnTotal := int64(0)
	toOperator := int64(0)
	toAgent := int64(0)

	for _, evt := range result.Events {
		if evt.Type == "withdraw_commission" || evt.Type == "withdraw_node_commission" {
			for _, attr := range evt.Attributes {
				switch attr.Key {
				case "amount":
					withdrawnTotal = parseAmount(attr.Value)
				case "operator_share":
					toOperator = parseAmount(attr.Value)
				case "agent_share":
					toAgent = parseAmount(attr.Value)
				}
			}
		}
	}

	// If event parsing didn't yield operator/agent split, estimate
	if withdrawnTotal > 0 && toOperator == 0 && toAgent == 0 {
		toOperator = withdrawnTotal * 80 / 100
		toAgent = withdrawnTotal - toOperator
	}

	return &types.WithdrawRewardsResponse{
		TxHash:         result.TxHash,
		WithdrawnTotal: utility.FormatKkot(withdrawnTotal),
		ToOperator:     utility.FormatKkot(toOperator),
		ToAgent:        utility.FormatKkot(toAgent),
	}, nil
}

func parseAmount(s string) int64 {
	// Parse amount string like "1000uppyeo" or "1000"
	var result int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		} else {
			break
		}
	}
	return result
}

type nodeInfo struct {
	ID               string `json:"id"`
	Operator         string `json:"operator"`
	AgentShare       string `json:"agent_share"`
	ValidatorAddress string `json:"validator_address"`
}

func (m *Module) getNode(ctx context.Context, nodeID string) (*nodeInfo, error) {
	path := fmt.Sprintf("/seocheon/node/v1/nodes/%s", nodeID)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrNodeNotFound, err)
	}

	var result struct {
		Node nodeInfo `json:"node"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parsing node response: %w", err)
	}
	return &result.Node, nil
}

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

func (m *Module) queryOutstandingRewards(ctx context.Context, valAddr string) int64 {
	path := fmt.Sprintf("/cosmos/distribution/v1beta1/validators/%s/outstanding_rewards", valAddr)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return 0
	}

	var result struct {
		Rewards struct {
			Rewards []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"rewards"`
		} `json:"rewards"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return 0
	}

	for _, r := range result.Rewards.Rewards {
		if r.Denom == "uppyeo" {
			amount, _ := strconv.ParseFloat(r.Amount, 64)
			return int64(amount)
		}
	}
	return 0
}

func (m *Module) queryCommission(ctx context.Context, valAddr string) int64 {
	path := fmt.Sprintf("/cosmos/distribution/v1beta1/validators/%s/commission", valAddr)
	data, err := m.client.QueryREST(ctx, path)
	if err != nil {
		return 0
	}

	var result struct {
		Commission struct {
			Commission []struct {
				Denom  string `json:"denom"`
				Amount string `json:"amount"`
			} `json:"commission"`
		} `json:"commission"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return 0
	}

	for _, c := range result.Commission.Commission {
		if c.Denom == "uppyeo" {
			amount, _ := strconv.ParseFloat(c.Amount, 64)
			return int64(amount)
		}
	}
	return 0
}

func parseAgentShare(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.2 // default 20%
	}
	if val > 1 {
		return val / 100.0
	}
	return val
}

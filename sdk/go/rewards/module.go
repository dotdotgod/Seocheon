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
	"github.com/seocheon/sdk-go/types"
)

// Module provides reward-related operations.
type Module struct {
	client chain.Client
	signer signing.Service
}

// NewModule creates a new Rewards module.
func NewModule(client chain.Client, signer signing.Service) *Module {
	return &Module{
		client: client,
		signer: signer,
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
		DelegationReward: formatKkot(delegationReward),
		ActivityReward:   formatKkot(activityReward),
		TotalReward:      formatKkot(totalReward),
		CommissionTotal:  formatKkot(commissionTotal),
		OperatorShare:    formatKkot(operatorShare),
		AgentShare:       formatKkot(agentShare),
	}, nil
}

// Withdraw withdraws all pending rewards. Requires operator signing key.
func (m *Module) Withdraw(ctx context.Context) (*types.WithdrawRewardsResponse, error) {
	// Build MsgWithdrawNodeCommission
	// This requires the full TX building/signing/broadcasting flow
	_ = m.signer.GetAddress() // operator

	// TODO: Implement full TX flow when proto message builders are available
	return nil, fmt.Errorf("rewards.Withdraw: TX flow requires proto message builders (not yet implemented)")
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

func formatKkot(uppyeo int64) string {
	intPart := uppyeo / 10000000000
	decPart := uppyeo % 10000000000
	if decPart < 0 {
		decPart = -decPart
	}
	return fmt.Sprintf("%d.%010d", intPart, decPart)
}

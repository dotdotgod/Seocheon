package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// WithdrawNodeCommission withdraws validator commission and splits it between
// operator and agent according to agent_share.
//
// Flow:
// 1. Look up node by operator
// 2. Call distributionKeeper.WithdrawValidatorCommission
// 3. Calculate agent_amount = total * (agent_share / 100)
// 4. Send agent_amount from operator to agent_address
func (k msgServer) WithdrawNodeCommission(ctx context.Context, msg *types.MsgWithdrawNodeCommission) (*types.MsgWithdrawNodeCommissionResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Look up node by operator.
	nodeID, err := k.OperatorIndex.Get(ctx, msg.Operator)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrNodeNotFound, "no node found for operator %s", msg.Operator)
	}

	node, err := k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return nil, errorsmod.Wrapf(types.ErrNodeNotFound, "node %s not found", nodeID)
	}

	// Node must not be inactive.
	if node.Status == types.NodeStatus_NODE_STATUS_INACTIVE {
		return nil, errorsmod.Wrap(types.ErrNodeInactive, "cannot withdraw commission for an inactive node")
	}

	// Withdraw validator commission via distribution module.
	valAddrBytes, err := sdk.GetFromBech32(node.ValidatorAddress, "seocheonvaloper")
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid validator address")
	}
	valAddr := sdk.ValAddress(valAddrBytes)

	totalCoins, err := k.distributionKeeper.WithdrawValidatorCommission(ctx, valAddr)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to withdraw validator commission")
	}

	// If no commission, return zero amounts.
	if totalCoins.IsZero() {
		return &types.MsgWithdrawNodeCommissionResponse{
			OperatorAmount: sdk.NewCoins().String(),
			AgentAmount:    sdk.NewCoins().String(),
		}, nil
	}

	// Calculate agent split.
	hundred := math.LegacyNewDec(100)
	agentShareRatio := node.AgentShare.Quo(hundred)

	var agentCoins sdk.Coins
	var operatorCoins sdk.Coins

	if node.AgentAddress != "" && node.AgentShare.IsPositive() {
		// Calculate agent_amount for each denom.
		for _, coin := range totalCoins {
			agentAmt := agentShareRatio.MulInt(coin.Amount).TruncateInt()
			operatorAmt := coin.Amount.Sub(agentAmt)

			if agentAmt.IsPositive() {
				agentCoins = agentCoins.Add(sdk.NewCoin(coin.Denom, agentAmt))
			}
			if operatorAmt.IsPositive() {
				operatorCoins = operatorCoins.Add(sdk.NewCoin(coin.Denom, operatorAmt))
			}
		}

		// Send agent's share from operator to agent.
		if agentCoins.IsAllPositive() {
			operatorAddr, addrErr := sdk.AccAddressFromBech32(msg.Operator)
			if addrErr != nil {
				return nil, errorsmod.Wrap(addrErr, "invalid operator address")
			}
			agentAddr, addrErr := sdk.AccAddressFromBech32(node.AgentAddress)
			if addrErr != nil {
				return nil, errorsmod.Wrap(addrErr, "invalid agent address")
			}

			if err := k.bankKeeper.SendCoins(ctx, operatorAddr, agentAddr, agentCoins); err != nil {
				return nil, errorsmod.Wrap(err, "failed to send agent share")
			}
		}
	} else {
		// No agent or zero share — operator gets everything.
		operatorCoins = totalCoins
	}

	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		types.EventTypeCommissionWithdrawn,
		sdk.NewAttribute(types.AttributeKeyNodeID, nodeID),
		sdk.NewAttribute(types.AttributeKeyOperatorAmount, operatorCoins.String()),
		sdk.NewAttribute(types.AttributeKeyAgentAmount, agentCoins.String()),
	))

	return &types.MsgWithdrawNodeCommissionResponse{
		OperatorAmount: operatorCoins.String(),
		AgentAmount:    agentCoins.String(),
	}, nil
}

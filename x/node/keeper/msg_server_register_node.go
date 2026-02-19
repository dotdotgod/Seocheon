package keeper

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/x/node/types"
)

func (k msgServer) RegisterNode(ctx context.Context, msg *types.MsgRegisterNode) (*types.MsgRegisterNodeResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := k.Params.Get(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get params")
	}

	// [1] Input validation.

	// Validate agent_share range (0 <= x <= 100).
	hundred := math.LegacyNewDec(100)
	if msg.AgentShare.IsNegative() || msg.AgentShare.GT(hundred) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAgentShare, "agent_share %s out of range [0, 100]", msg.AgentShare)
	}
	if msg.MaxAgentShareChangeRate.IsNegative() || msg.MaxAgentShareChangeRate.GT(hundred) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAgentShare, "max_agent_share_change_rate %s out of range [0, 100]", msg.MaxAgentShareChangeRate)
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

	// Check 1 operator = 1 node.
	if has, _ := k.OperatorIndex.Has(ctx, msg.Operator); has {
		return nil, errorsmod.Wrapf(types.ErrNodeAlreadyExists, "operator %s already has a node", msg.Operator)
	}

	// Check agent_address uniqueness.
	if msg.AgentAddress != "" {
		if has, _ := k.AgentIndex.Has(ctx, msg.AgentAddress); has {
			return nil, errorsmod.Wrapf(types.ErrAgentAddressAlreadyUsed, "agent address %s already registered", msg.AgentAddress)
		}
	}

	// Validate consensus_pubkey is provided.
	if msg.ConsensusPubkey == nil {
		return nil, errorsmod.Wrap(types.ErrInvalidConsensusPubkey, "consensus pubkey is required")
	}

	// [2] Check per-block registration limit.
	blockHeight := sdkCtx.BlockHeight()
	regCount, err := k.BlockRegistrationCount.Get(ctx, blockHeight)
	if err != nil {
		regCount = 0
	}
	if regCount >= params.MaxRegistrationsPerBlock {
		return nil, errorsmod.Wrapf(types.ErrMaxRegistrationsPerBlock, "block %d already has %d registrations", blockHeight, regCount)
	}

	// [3] Check Registration Pool balance.
	regPoolAddr := k.authKeeper.GetModuleAddress(types.RegistrationPoolName)
	bondDenom, err := k.stakingKeeper.BondDenom(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get bond denom")
	}
	regPoolBalance := k.bankKeeper.GetBalance(ctx, regPoolAddr, bondDenom)
	oneUsum := sdk.NewCoin(bondDenom, math.NewInt(1))
	if regPoolBalance.IsLT(oneUsum) {
		return nil, errorsmod.Wrap(types.ErrRegistrationPoolDepleted, "insufficient registration pool balance")
	}

	// [4] Transfer 1 usum from Registration Pool to operator.
	operatorAddr, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid operator address")
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.RegistrationPoolName, operatorAddr, sdk.NewCoins(oneUsum))
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to transfer from registration pool")
	}

	// [5] Create validator via x/staking (self-delegation of 1 usum).
	valAddr := sdk.ValAddress(operatorAddr)
	valAddrStr, err := sdk.Bech32ifyAddressBytes("seocheonvaloper", valAddr)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create validator address")
	}

	if k.stakingMsgServer != nil {
		createValMsg := &stakingtypes.MsgCreateValidator{
			Description: stakingtypes.Description{
				Moniker: msg.Description,
			},
			Commission: stakingtypes.CommissionRates{
				Rate:          math.LegacyZeroDec(),
				MaxRate:       math.LegacyOneDec(),
				MaxChangeRate: math.LegacyOneDec(),
			},
			MinSelfDelegation: math.OneInt(),
			DelegatorAddress:  msg.Operator,
			ValidatorAddress:  valAddrStr,
			Pubkey:            msg.ConsensusPubkey,
			Value:             oneUsum,
		}
		if _, err := k.stakingMsgServer.CreateValidator(ctx, createValMsg); err != nil {
			return nil, errorsmod.Wrap(err, "failed to create validator")
		}
	}

	// [6] Generate deterministic node ID and store node state.
	nodeID := generateNodeID(msg.Operator)

	node := types.Node{
		Id:                      nodeID,
		Operator:                msg.Operator,
		AgentAddress:            msg.AgentAddress,
		AgentShare:              msg.AgentShare,
		MaxAgentShareChangeRate: msg.MaxAgentShareChangeRate,
		Description:             msg.Description,
		Website:                 msg.Website,
		Tags:                    msg.Tags,
		ValidatorAddress:        valAddrStr,
		Status:                  types.NodeStatus_NODE_STATUS_REGISTERED,
		RegisteredAt:            blockHeight,
	}

	if err := k.Nodes.Set(ctx, nodeID, node); err != nil {
		return nil, errorsmod.Wrap(err, "failed to store node")
	}

	// Set indexes.
	if err := k.OperatorIndex.Set(ctx, msg.Operator, nodeID); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set operator index")
	}
	if msg.AgentAddress != "" {
		if err := k.AgentIndex.Set(ctx, msg.AgentAddress, nodeID); err != nil {
			return nil, errorsmod.Wrap(err, "failed to set agent index")
		}
	}
	if err := k.ValidatorIndex.Set(ctx, valAddrStr, nodeID); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set validator index")
	}

	// Set tag indexes.
	for _, tag := range msg.Tags {
		tagKey := collections.Join(tag, nodeID)
		if err := k.TagIndex.Set(ctx, tagKey); err != nil {
			return nil, errorsmod.Wrap(err, "failed to set tag index")
		}
	}

	// Update block registration count.
	if err := k.BlockRegistrationCount.Set(ctx, blockHeight, regCount+1); err != nil {
		return nil, errorsmod.Wrap(err, "failed to update registration count")
	}

	// [7] Grant feegrant to agent_address (best-effort).
	if msg.AgentAddress != "" {
		_ = k.grantAgentFeegrant(ctx, msg.AgentAddress)
	}

	// [8] Emit event.
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent(
		"node_registered",
		sdk.NewAttribute("node_id", nodeID),
		sdk.NewAttribute("operator", msg.Operator),
		sdk.NewAttribute("agent_address", msg.AgentAddress),
		sdk.NewAttribute("validator_address", valAddrStr),
		sdk.NewAttribute("block_height", fmt.Sprintf("%d", blockHeight)),
	))

	return &types.MsgRegisterNodeResponse{
		NodeId:           nodeID,
		ValidatorAddress: valAddrStr,
	}, nil
}

// generateNodeID creates a deterministic node ID from the operator address.
func generateNodeID(operator string) string {
	hash := sha256.Sum256([]byte("seocheon-node:" + operator))
	return hex.EncodeToString(hash[:16])
}

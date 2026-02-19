package ante

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// NodeKeeper defines the interface for the node keeper methods needed by ante decorators.
type NodeKeeper interface {
	IsRegisteredAgent(ctx context.Context, address string) bool
	GetAllowedAgentMsgTypes(ctx context.Context) ([]string, error)
}

// AgentPermissionDecorator ensures that agent addresses can only execute whitelisted message types.
// If the fee payer of a transaction is a registered agent_address, all messages in the TX must be
// in the agent_allowed_msg_types parameter list. Non-agent addresses are not restricted.
type AgentPermissionDecorator struct {
	nodeKeeper NodeKeeper
}

func NewAgentPermissionDecorator(nk NodeKeeper) AgentPermissionDecorator {
	return AgentPermissionDecorator{nodeKeeper: nk}
}

func (d AgentPermissionDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return next(ctx, tx, simulate)
	}

	feePayerBytes := feeTx.FeePayer()
	if feePayerBytes == nil {
		return next(ctx, tx, simulate)
	}

	feePayerAddr := sdk.AccAddress(feePayerBytes).String()

	if !d.nodeKeeper.IsRegisteredAgent(ctx, feePayerAddr) {
		// Not an agent address — no restrictions.
		return next(ctx, tx, simulate)
	}

	// Fee payer is a registered agent. Validate all messages against the whitelist.
	allowedTypes, err := d.nodeKeeper.GetAllowedAgentMsgTypes(ctx)
	if err != nil {
		return ctx, errorsmod.Wrap(err, "failed to get allowed agent msg types")
	}

	allowed := make(map[string]bool, len(allowedTypes))
	for _, t := range allowedTypes {
		allowed[t] = true
	}

	for _, msg := range tx.GetMsgs() {
		msgTypeURL := sdk.MsgTypeURL(msg)
		if !allowed[msgTypeURL] {
			return ctx, errorsmod.Wrapf(
				types.ErrUnauthorizedAgentMsg,
				"agent %s is not authorized to execute %s",
				feePayerAddr, msgTypeURL,
			)
		}
	}

	return next(ctx, tx, simulate)
}

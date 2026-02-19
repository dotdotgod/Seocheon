package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// RegistrationFeeDecorator allows registration transactions (MsgRegisterNode, MsgRenewFeegrant)
// to be processed with zero fees. This enables 0-cost participation in the network.
// It must be placed before the DeductFeeDecorator in the ante handler chain.
type RegistrationFeeDecorator struct{}

func NewRegistrationFeeDecorator() RegistrationFeeDecorator {
	return RegistrationFeeDecorator{}
}

func (d RegistrationFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if isRegistrationOnlyTx(tx.GetMsgs()) {
		// Set min gas prices to zero, allowing this TX to pass fee validation with 0 fees.
		// Registration TXs are subsidized by the network to enable 0-cost participation.
		ctx = ctx.WithMinGasPrices(sdk.DecCoins{})
	}

	return next(ctx, tx, simulate)
}

// isRegistrationOnlyTx checks if all messages in the transaction are registration-related.
// Mixed transactions (containing both registration and non-registration messages) are not eligible.
func isRegistrationOnlyTx(msgs []sdk.Msg) bool {
	if len(msgs) == 0 {
		return false
	}

	for _, msg := range msgs {
		switch msg.(type) {
		case *types.MsgRegisterNode, *types.MsgRenewFeegrant:
			continue
		default:
			return false
		}
	}

	return true
}

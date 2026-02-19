package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	nodeante "seocheon/x/node/ante"
)

// newAnteHandler creates the application's ante handler chain.
// It prepends Seocheon-specific decorators (RegistrationFeeDecorator, AgentPermissionDecorator)
// before the standard Cosmos SDK ante handler chain.
func newAnteHandler(app *App) (sdk.AnteHandler, error) {
	// Build the standard SDK ante handler.
	standardHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   app.AuthKeeper,
		BankKeeper:      app.BankKeeper,
		SignModeHandler: app.txConfig.SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		// FeegrantKeeper is nil — Seocheon uses a custom feegrant pool mechanism
		// via the RegistrationFeeDecorator instead of the standard feegrant fee deduction.
	})
	if err != nil {
		return nil, err
	}

	// Chain custom decorators before the standard handler.
	return chainBeforeHandler(
		standardHandler,
		nodeante.NewRegistrationFeeDecorator(),
		nodeante.NewAgentPermissionDecorator(&app.NodeKeeper),
	), nil
}

// chainBeforeHandler prepends ante decorators before an existing AnteHandler.
// Each decorator calls next, and the last decorator's next is the base handler.
func chainBeforeHandler(base sdk.AnteHandler, decorators ...sdk.AnteDecorator) sdk.AnteHandler {
	if len(decorators) == 0 {
		return base
	}

	handler := base
	for i := len(decorators) - 1; i >= 0; i-- {
		decorator := decorators[i]
		next := handler
		handler = func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
			return decorator.AnteHandle(ctx, tx, simulate, next)
		}
	}

	return handler
}


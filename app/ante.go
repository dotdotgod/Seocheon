package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	nodeante "seocheon/x/node/ante"
)

func newAnteHandler(app *App) (sdk.AnteHandler, error) {
	standardHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   app.AuthKeeper,
		BankKeeper:      app.BankKeeper,
		SignModeHandler: app.txConfig.SignModeHandler(),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	})
	if err != nil {
		return nil, err
	}

	return chainBeforeHandler(
		standardHandler,
		nodeante.NewRegistrationFeeDecorator(),
		nodeante.NewAgentPermissionDecorator(app.NodeKeeper),
	), nil
}

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

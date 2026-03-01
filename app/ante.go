package app

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

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

	// Convert *uint64 to *storetypes.Gas for wasm simulation gas limit.
	var simGasLimit *storetypes.Gas
	if app.wasmNodeConfig.SimulationGasLimit != nil {
		limit := storetypes.Gas(*app.wasmNodeConfig.SimulationGasLimit)
		simGasLimit = &limit
	}

	return chainBeforeHandler(
		standardHandler,
		// Seocheon-specific decorators.
		nodeante.NewRegistrationFeeDecorator(),
		nodeante.NewAgentPermissionDecorator(app.NodeKeeper),
		// CosmWasm decorators.
		wasmkeeper.NewLimitSimulationGasDecorator(simGasLimit),
		wasmkeeper.NewCountTXDecorator(runtime.NewKVStoreService(app.GetKey(wasmtypes.StoreKey))),
		wasmkeeper.NewGasRegisterDecorator(app.WasmKeeper.GetGasRegister()),
		wasmkeeper.NewTxContractsDecorator(),
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

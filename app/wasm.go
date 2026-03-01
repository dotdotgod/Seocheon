package app

import (
	"fmt"
	"path/filepath"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/spf13/cast"

	wasm "github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// registerWasmModule creates the CosmWasm (x/wasm) keeper and registers the wasm app module.
// Must be called AFTER registerIBCKeepers (needs IBCKeeper.ChannelKeeper)
// and BEFORE registerIBCRouterAndModules (which adds the wasm IBC route).
func (app *App) registerWasmModule(appOpts servertypes.AppOptions) error {
	// Register wasm store key.
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(wasmtypes.StoreKey),
	); err != nil {
		return err
	}

	// Register wasm params subspace.
	app.ParamsKeeper.Subspace(wasmtypes.ModuleName)

	// Read wasm node configuration from app options.
	homePath := cast.ToString(appOpts.Get(flags.FlagHome))
	wasmDir := filepath.Join(homePath, "wasm")
	nodeConfig, err := wasm.ReadNodeConfig(appOpts)
	if err != nil {
		return fmt.Errorf("error reading wasm config: %w", err)
	}

	// Store simulation gas limit for ante handler use.
	app.wasmNodeConfig = nodeConfig

	govModuleAddr, _ := app.AuthKeeper.AddressCodec().BytesToString(
		authtypes.NewModuleAddress(govtypes.ModuleName),
	)

	// Create WasmKeeper.
	app.WasmKeeper = wasmkeeper.NewKeeper(
		app.appCodec,
		runtime.NewKVStoreService(app.GetKey(wasmtypes.StoreKey)),
		app.AuthKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,  // ICS4Wrapper
		app.IBCKeeper.ChannelKeeper,  // ChannelKeeper
		app.IBCKeeper.ChannelKeeperV2,
		app.TransferKeeper,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		wasmDir,
		nodeConfig,
		wasmtypes.VMConfig{},
		wasmkeeper.BuiltInCapabilities(),
		govModuleAddr,
	)

	// Register the wasm AppModule.
	if err := app.RegisterModules(
		wasm.NewAppModule(
			app.appCodec,
			&app.WasmKeeper,
			app.StakingKeeper,
			app.AuthKeeper,
			app.BankKeeper,
			app.MsgServiceRouter(),
			app.GetSubspace(wasmtypes.ModuleName),
		),
	); err != nil {
		return err
	}

	return nil
}

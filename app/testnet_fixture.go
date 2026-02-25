package app

import (
	dbm "github.com/cosmos/cosmos-db"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

// NewTestNetworkFixture returns a TestFixture factory for testutil/network.
// The AppConstructor uses app.New() so that post-injection keeper wiring
// (feegrant bankKeeper, node distributionKeeper, etc.) is included.
func NewTestNetworkFixture() network.TestFixture {
	app := New(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		simtestutil.NewAppOptionsWithFlagHome(DefaultNodeHome),
	)

	return network.TestFixture{
		AppConstructor: func(val network.ValidatorI) servertypes.Application {
			return New(
				val.GetCtx().Logger,
				dbm.NewMemDB(),
				nil,
				true,
				simtestutil.AppOptionsMap{flags.FlagHome: val.GetCtx().Config.RootDir},
				baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
				baseapp.SetChainID(val.GetCtx().Viper.GetString(flags.FlagChainID)),
			)
		},
		GenesisState:   app.DefaultGenesis(),
		EncodingConfig: makeEncodingConfig(app),
	}
}

func makeEncodingConfig(app *App) moduletestutil.TestEncodingConfig {
	return moduletestutil.TestEncodingConfig{
		InterfaceRegistry: app.InterfaceRegistry(),
		Codec:             app.AppCodec(),
		TxConfig:          app.TxConfig(),
		Amino:             app.LegacyAmino(),
	}
}

// Note: DefaultGenesis() is inherited from runtime.App via embedding.

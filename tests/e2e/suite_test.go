package e2e_test

import (
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"seocheon/app"
	"seocheon/testutil"
	activitytypes "seocheon/x/activity/types"
	nodetypes "seocheon/x/node/types"
)

// E2ESuite is the shared test suite for all Seocheon E2E tests.
type E2ESuite struct {
	suite.Suite
	cfg     network.Config
	network *network.Network
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ESuite))
}

func (s *E2ESuite) SetupSuite() {
	s.T().Log("setting up E2E test suite")

	cfg := network.DefaultConfig(app.NewTestNetworkFixture)

	// Override for fast epoch progression.
	cfg.NumValidators = 2
	cfg.BondDenom = "usum"
	cfg.MinGasPrices = "0usum"
	cfg.TimeoutCommit = 500 * time.Millisecond

	// Increase validator staking tokens for sufficient delegation.
	cfg.StakingTokens = sdkmath.NewInt(1_000_000_000)
	cfg.BondedTokens = sdkmath.NewInt(100_000_000)
	cfg.AccountTokens = sdkmath.NewInt(10_000_000_000)

	// Patch genesis: x/staking bond_denom = usum.
	s.patchStakingGenesis(&cfg)

	// Patch genesis: x/activity fast testnet params.
	s.patchActivityGenesis(&cfg)

	// Patch genesis: x/node registration pool funding.
	s.patchNodeGenesis(&cfg)

	// Patch genesis: bank balances for module accounts.
	s.patchBankGenesis(&cfg)

	// Patch genesis: x/gov fast voting for governance tests.
	s.patchGovGenesis(&cfg)

	s.cfg = cfg

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err)

	// Wait for the network to produce a block.
	s.Require().NoError(s.network.WaitForNextBlock())
}

func (s *E2ESuite) TearDownSuite() {
	s.T().Log("tearing down E2E test suite")
	s.network.Cleanup()
}

// patchStakingGenesis sets the bond denom to "usum".
func (s *E2ESuite) patchStakingGenesis(cfg *network.Config) {
	var stakingGen stakingtypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(cfg.GenesisState[stakingtypes.ModuleName], &stakingGen))
	stakingGen.Params.BondDenom = "usum"
	bz, err := cfg.Codec.MarshalJSON(&stakingGen)
	s.Require().NoError(err)
	cfg.GenesisState[stakingtypes.ModuleName] = bz
}

// patchActivityGenesis injects fast testnet params.
func (s *E2ESuite) patchActivityGenesis(cfg *network.Config) {
	params := testutil.FastTestnetActivityParams()

	// Use raw JSON patching since activity genesis may vary.
	raw := cfg.GenesisState[activitytypes.ModuleName]
	var gen map[string]json.RawMessage
	if raw != nil {
		s.Require().NoError(json.Unmarshal(raw, &gen))
	} else {
		gen = make(map[string]json.RawMessage)
	}

	paramsBz, err := cfg.Codec.MarshalJSON(&params)
	s.Require().NoError(err)
	gen["params"] = paramsBz

	bz, err := json.Marshal(gen)
	s.Require().NoError(err)
	cfg.GenesisState[activitytypes.ModuleName] = bz
}

// patchNodeGenesis funds the registration pool so that MsgRegisterNode can succeed.
func (s *E2ESuite) patchNodeGenesis(cfg *network.Config) {
	raw := cfg.GenesisState[nodetypes.ModuleName]
	var gen map[string]json.RawMessage
	if raw != nil {
		s.Require().NoError(json.Unmarshal(raw, &gen))
	} else {
		gen = make(map[string]json.RawMessage)
	}

	// Patch params if needed (defaults are fine for E2E).
	bz, err := json.Marshal(gen)
	s.Require().NoError(err)
	cfg.GenesisState[nodetypes.ModuleName] = bz
}

// patchBankGenesis adds balances for registration_pool and feegrant_pool module accounts.
func (s *E2ESuite) patchBankGenesis(cfg *network.Config) {
	var bankGen banktypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(cfg.GenesisState[banktypes.ModuleName], &bankGen))

	// Fund registration_pool with 1000 usum (enough for many test registrations).
	regPoolAddr := authtypes_ModuleAddress(nodetypes.RegistrationPoolName)
	regPoolCoins := sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(1000)))

	// Fund feegrant_pool with 100000 usum.
	fgPoolAddr := authtypes_ModuleAddress(nodetypes.FeegrantPoolName)
	fgPoolCoins := sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(100000)))

	// Fund activity_reward_pool with 1,000,000 usum for reward distribution tests.
	arPoolAddr := authtypes_ModuleAddress(activitytypes.ActivityRewardPoolName)
	arPoolCoins := sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(1_000_000)))

	bankGen.Balances = append(bankGen.Balances,
		banktypes.Balance{Address: regPoolAddr, Coins: regPoolCoins},
		banktypes.Balance{Address: fgPoolAddr, Coins: fgPoolCoins},
		banktypes.Balance{Address: arPoolAddr, Coins: arPoolCoins},
	)

	// Clear supply so InitGenesis auto-calculates from balances.
	// The network util adds validator balances after this, making a pre-set supply incorrect.
	bankGen.Supply = sdk.Coins{}

	bz, err := cfg.Codec.MarshalJSON(&bankGen)
	s.Require().NoError(err)
	cfg.GenesisState[banktypes.ModuleName] = bz
}

// patchGovGenesis configures fast voting period for governance tests.
func (s *E2ESuite) patchGovGenesis(cfg *network.Config) {
	var govGen govv1.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(cfg.GenesisState[govtypes.ModuleName], &govGen))

	votingPeriod := 10 * time.Second
	depositPeriod := 10 * time.Second

	if govGen.Params == nil {
		govGen.Params = &govv1.Params{}
	}
	govGen.Params.VotingPeriod = &votingPeriod
	govGen.Params.MaxDepositPeriod = &depositPeriod
	govGen.Params.MinDeposit = sdk.NewCoins(sdk.NewCoin("usum", sdkmath.NewInt(10)))
	govGen.Params.Quorum = "0.000001"
	govGen.Params.Threshold = "0.5"

	bz, err := cfg.Codec.MarshalJSON(&govGen)
	s.Require().NoError(err)
	cfg.GenesisState[govtypes.ModuleName] = bz
}

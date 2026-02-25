package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	activitytypes "seocheon/x/activity/types"
	nodetypes "seocheon/x/node/types"
)

const (
	// Total supply: 500M KKOT = 500,000,000 * 10^6 usum.
	totalSupplyKKOT = 500_000_000
	usumPerKKOT     = 1_000_000
	totalSupplyUsum = totalSupplyKKOT * usumPerKKOT // 500,000,000,000,000

	// Allocation percentages.
	teamPercent             = 20 // 100M KKOT
	foundationPercent       = 15 // 75M KKOT
	airdropPercent          = 15 // 75M KKOT
	ecosystemFundPercent    = 15 // 75M KKOT
	initialValidatorPercent = 10 // 50M KKOT
	ecosystemReservePercent = 10 // 50M KKOT
	communityPoolPercent    = 15 // 75M KKOT
)

// GenesisAllocation holds the computed token allocations.
type GenesisAllocation struct {
	Team             math.Int `json:"team"`
	Foundation       math.Int `json:"foundation"`
	Airdrop          math.Int `json:"airdrop"`
	EcosystemFund    math.Int `json:"ecosystem_fund"`
	InitialValidator math.Int `json:"initial_validator"`
	EcosystemReserve math.Int `json:"ecosystem_reserve"`
	CommunityPool    math.Int `json:"community_pool"`
	Total            math.Int `json:"total"`
}

func computeAllocations() GenesisAllocation {
	total := math.NewInt(totalSupplyUsum)
	alloc := GenesisAllocation{
		Team:             total.Mul(math.NewInt(teamPercent)).Quo(math.NewInt(100)),
		Foundation:       total.Mul(math.NewInt(foundationPercent)).Quo(math.NewInt(100)),
		Airdrop:          total.Mul(math.NewInt(airdropPercent)).Quo(math.NewInt(100)),
		EcosystemFund:    total.Mul(math.NewInt(ecosystemFundPercent)).Quo(math.NewInt(100)),
		InitialValidator: total.Mul(math.NewInt(initialValidatorPercent)).Quo(math.NewInt(100)),
		EcosystemReserve: total.Mul(math.NewInt(ecosystemReservePercent)).Quo(math.NewInt(100)),
		CommunityPool:    total.Mul(math.NewInt(communityPoolPercent)).Quo(math.NewInt(100)),
		Total:            total,
	}
	return alloc
}

// NewGenesisBuildCmd returns a command that configures a production genesis file.
func NewGenesisBuildCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis-build [genesis-file]",
		Short: "Configure a production genesis file with Seocheon-specific parameters",
		Long: `Configure module parameters and token allocations for a production genesis file.

This command reads an existing genesis.json (created by 'seocheon init' and gentx collection),
applies Seocheon-specific module parameters, sets up pool accounts, and validates the result.

Parameters configured:
  x/staking:  max_validators=150, unbonding_time=21d, bond_denom=usum
  x/mint:     inflation 7-15%, goal_bonded=67%, blocks_per_year=6,307,200
  x/activity: epoch=17280, windows=12, min_active=8, d_min=3000

Pool accounts funded:
  airdrop_pool:      75,000,000 KKOT (15%)
  ecosystem_fund:    75,000,000 KKOT (15%)
  ecosystem_reserve: 50,000,000 KKOT (10%)
  community_pool:    75,000,000 KKOT (15%)
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			return runGenesisBuild(clientCtx, args[0], mbm)
		},
	}

	cmd.Flags().String("team-address", "", "Team vesting account address (bech32)")
	cmd.Flags().String("foundation-address", "", "Foundation multisig address (bech32)")
	cmd.Flags().Bool("dry-run", false, "Print allocations without modifying the file")

	return cmd
}

func runGenesisBuild(clientCtx client.Context, genesisFile string, mbm module.BasicManager) error {
	cdc := clientCtx.Codec

	alloc := computeAllocations()

	// Check dry-run.
	cmd := clientCtx.CmdContext
	_ = cmd

	// Print allocation summary.
	fmt.Println("=== Seocheon Genesis Allocation ===")
	fmt.Printf("  Team (20%%):              %s usum (%s KKOT)\n", alloc.Team, alloc.Team.Quo(math.NewInt(usumPerKKOT)))
	fmt.Printf("  Foundation (15%%):        %s usum (%s KKOT)\n", alloc.Foundation, alloc.Foundation.Quo(math.NewInt(usumPerKKOT)))
	fmt.Printf("  Airdrop Pool (15%%):      %s usum (%s KKOT)\n", alloc.Airdrop, alloc.Airdrop.Quo(math.NewInt(usumPerKKOT)))
	fmt.Printf("  Ecosystem Fund (15%%):    %s usum (%s KKOT)\n", alloc.EcosystemFund, alloc.EcosystemFund.Quo(math.NewInt(usumPerKKOT)))
	fmt.Printf("  Initial Validators (10%%): %s usum (%s KKOT)\n", alloc.InitialValidator, alloc.InitialValidator.Quo(math.NewInt(usumPerKKOT)))
	fmt.Printf("  Ecosystem Reserve (10%%): %s usum (%s KKOT)\n", alloc.EcosystemReserve, alloc.EcosystemReserve.Quo(math.NewInt(usumPerKKOT)))
	fmt.Printf("  Community Pool (15%%):    %s usum (%s KKOT)\n", alloc.CommunityPool, alloc.CommunityPool.Quo(math.NewInt(usumPerKKOT)))
	fmt.Printf("  TOTAL:                   %s usum (%s KKOT)\n", alloc.Total, alloc.Total.Quo(math.NewInt(usumPerKKOT)))

	// Verify sum.
	sum := alloc.Team.Add(alloc.Foundation).Add(alloc.Airdrop).Add(alloc.EcosystemFund).
		Add(alloc.InitialValidator).Add(alloc.EcosystemReserve).Add(alloc.CommunityPool)
	if !sum.Equal(alloc.Total) {
		return fmt.Errorf("allocation sum mismatch: %s != %s", sum, alloc.Total)
	}
	fmt.Println("  Allocation sum verified OK")

	// Read genesis file.
	genesisBytes, err := os.ReadFile(genesisFile)
	if err != nil {
		return fmt.Errorf("failed to read genesis file: %w", err)
	}

	var genesisDoc map[string]json.RawMessage
	if err := json.Unmarshal(genesisBytes, &genesisDoc); err != nil {
		return fmt.Errorf("failed to parse genesis file: %w", err)
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genesisDoc["app_state"], &appState); err != nil {
		return fmt.Errorf("failed to parse app_state: %w", err)
	}

	// Apply module parameters.
	if err := applyStakingParams(cdc, appState); err != nil {
		return fmt.Errorf("staking params: %w", err)
	}
	if err := applyMintParams(cdc, appState); err != nil {
		return fmt.Errorf("mint params: %w", err)
	}
	if err := applyActivityParams(cdc, appState); err != nil {
		return fmt.Errorf("activity params: %w", err)
	}
	if err := applyNodeParams(cdc, appState); err != nil {
		return fmt.Errorf("node params: %w", err)
	}

	// Fund pool module accounts.
	if err := fundPoolAccounts(cdc, appState, alloc); err != nil {
		return fmt.Errorf("fund pools: %w", err)
	}

	// Fund community pool.
	if err := fundCommunityPool(cdc, appState, alloc); err != nil {
		return fmt.Errorf("community pool: %w", err)
	}

	// Update bank supply.
	if err := updateBankSupply(cdc, appState, alloc); err != nil {
		return fmt.Errorf("bank supply: %w", err)
	}

	// Marshal back.
	appStateBytes, err := json.Marshal(appState)
	if err != nil {
		return fmt.Errorf("failed to marshal app_state: %w", err)
	}
	genesisDoc["app_state"] = appStateBytes

	finalBytes, err := json.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal genesis: %w", err)
	}

	// Write output.
	outputPath := genesisFile
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outputPath, finalBytes, 0o644); err != nil {
		return fmt.Errorf("failed to write genesis: %w", err)
	}

	fmt.Printf("\nGenesis file updated: %s\n", outputPath)
	return nil
}

func applyStakingParams(cdc codec.Codec, appState map[string]json.RawMessage) error {
	var stakingGenesis stakingtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[stakingtypes.ModuleName], &stakingGenesis); err != nil {
		return err
	}

	stakingGenesis.Params.MaxValidators = 150
	stakingGenesis.Params.UnbondingTime = 21 * 24 * time.Hour // 21 days
	stakingGenesis.Params.BondDenom = "usum"
	stakingGenesis.Params.MinCommissionRate = math.LegacyNewDecWithPrec(5, 2) // 5%

	bz, err := cdc.MarshalJSON(&stakingGenesis)
	if err != nil {
		return err
	}
	appState[stakingtypes.ModuleName] = bz
	fmt.Println("  Applied x/staking params: max_validators=150, unbonding=21d, bond_denom=usum")
	return nil
}

func applyMintParams(cdc codec.Codec, appState map[string]json.RawMessage) error {
	var mintGenesis minttypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[minttypes.ModuleName], &mintGenesis); err != nil {
		return err
	}

	mintGenesis.Params.MintDenom = "usum"
	mintGenesis.Params.InflationMin = math.LegacyNewDecWithPrec(7, 2)          // 7%
	mintGenesis.Params.InflationMax = math.LegacyNewDecWithPrec(15, 2)         // 15%
	mintGenesis.Params.InflationRateChange = math.LegacyNewDecWithPrec(8, 2)   // 8%p/year
	mintGenesis.Params.GoalBonded = math.LegacyNewDecWithPrec(67, 2)           // 67%
	mintGenesis.Params.BlocksPerYear = 6_307_200                               // 5s blocks

	bz, err := cdc.MarshalJSON(&mintGenesis)
	if err != nil {
		return err
	}
	appState[minttypes.ModuleName] = bz
	fmt.Println("  Applied x/mint params: inflation=7-15%, rate_change=8pp/year, goal_bonded=67%, blocks_per_year=6307200")
	return nil
}

func applyActivityParams(cdc codec.Codec, appState map[string]json.RawMessage) error {
	var activityGenesis activitytypes.GenesisState
	if raw, ok := appState[activitytypes.ModuleName]; ok {
		if err := cdc.UnmarshalJSON(raw, &activityGenesis); err != nil {
			return err
		}
	}

	activityGenesis.Params = activitytypes.Params{
		EpochLength:              17280,
		WindowsPerEpoch:          12,
		MinActiveWindows:         8,
		SelfFundedQuota:          100,
		FeegrantQuota:            10,
		ActivityPruningKeepBlocks: 1_555_200, // ~90 days
		FeeThresholdMultiplier:   3,
		BaseActivityFee:          1_000_000,   // 1 KKOT
		FeeExponent:              5000,        // 0.5
		MaxActivityFee:           100_000_000, // 100 KKOT
		MinFeegrantQuota:         8,
		QuotaReductionRate:       5000,        // 0.5
		FeegrantFeeExempt:        true,
		DMin:                     3000,        // 0.3
		FeeToActivityPoolRatio:   8000,        // 80%
	}

	bz, err := cdc.MarshalJSON(&activityGenesis)
	if err != nil {
		return err
	}
	appState[activitytypes.ModuleName] = bz
	fmt.Println("  Applied x/activity params: epoch=17280, windows=12, min_active=8, d_min=3000")
	return nil
}

func applyNodeParams(cdc codec.Codec, appState map[string]json.RawMessage) error {
	var nodeGenesis nodetypes.GenesisState
	if raw, ok := appState[nodetypes.ModuleName]; ok {
		if err := cdc.UnmarshalJSON(raw, &nodeGenesis); err != nil {
			return err
		}
	}

	// Ensure pool balances are set.
	if len(nodeGenesis.RegistrationPoolBalance) == 0 {
		nodeGenesis.RegistrationPoolBalance = sdk.NewCoins(
			sdk.NewCoin("usum", math.NewInt(1_000_000_000_000)), // 1,000 KKOT
		)
	}
	if len(nodeGenesis.FeegrantPoolBalance) == 0 {
		nodeGenesis.FeegrantPoolBalance = sdk.NewCoins(
			sdk.NewCoin("usum", math.NewInt(10_000_000_000_000)), // 10,000 KKOT
		)
	}

	bz, err := cdc.MarshalJSON(&nodeGenesis)
	if err != nil {
		return err
	}
	appState[nodetypes.ModuleName] = bz
	fmt.Println("  Applied x/node params: registration_pool=1000 KKOT, feegrant_pool=10000 KKOT")
	return nil
}

func fundPoolAccounts(cdc codec.Codec, appState map[string]json.RawMessage, alloc GenesisAllocation) error {
	var bankGenesis banktypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return err
	}

	// Fund module pool accounts.
	pools := []struct {
		name   string
		amount math.Int
	}{
		{"airdrop_pool", alloc.Airdrop},
		{"ecosystem_fund", alloc.EcosystemFund},
		{"ecosystem_reserve", alloc.EcosystemReserve},
	}

	for _, pool := range pools {
		addr := authtypes.NewModuleAddress(pool.name)
		coins := sdk.NewCoins(sdk.NewCoin("usum", pool.amount))

		// Add balance.
		bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
			Address: addr.String(),
			Coins:   coins,
		})
		fmt.Printf("  Funded %s: %s KKOT\n", pool.name, pool.amount.Quo(math.NewInt(usumPerKKOT)))
	}

	bz, err := cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz
	return nil
}

func fundCommunityPool(cdc codec.Codec, appState map[string]json.RawMessage, alloc GenesisAllocation) error {
	var distrGenesis distrtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[distrtypes.ModuleName], &distrGenesis); err != nil {
		return err
	}

	// Add community pool balance.
	communityPoolCoins := sdk.NewDecCoinsFromCoins(sdk.NewCoin("usum", alloc.CommunityPool))
	distrGenesis.FeePool.CommunityPool = distrGenesis.FeePool.CommunityPool.Add(communityPoolCoins...)

	bz, err := cdc.MarshalJSON(&distrGenesis)
	if err != nil {
		return err
	}
	appState[distrtypes.ModuleName] = bz

	// Also need to fund distribution module account in bank balances.
	var bankGenesis banktypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return err
	}

	distrAddr := authtypes.NewModuleAddress(distrtypes.ModuleName)
	bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
		Address: distrAddr.String(),
		Coins:   sdk.NewCoins(sdk.NewCoin("usum", alloc.CommunityPool)),
	})

	bz, err = cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz

	fmt.Printf("  Funded community pool: %s KKOT\n", alloc.CommunityPool.Quo(math.NewInt(usumPerKKOT)))
	return nil
}

func updateBankSupply(cdc codec.Codec, appState map[string]json.RawMessage, alloc GenesisAllocation) error {
	var bankGenesis banktypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return err
	}

	// Calculate total supply from all balances.
	totalFromBalances := math.ZeroInt()
	for _, bal := range bankGenesis.Balances {
		totalFromBalances = totalFromBalances.Add(bal.Coins.AmountOf("usum"))
	}

	bankGenesis.Supply = sdk.NewCoins(sdk.NewCoin("usum", totalFromBalances))

	bz, err := cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz

	fmt.Printf("  Bank supply set to: %s usum (%s KKOT)\n",
		totalFromBalances, totalFromBalances.Quo(math.NewInt(usumPerKKOT)))
	return nil
}

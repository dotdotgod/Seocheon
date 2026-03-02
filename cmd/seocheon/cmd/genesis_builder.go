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
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	activitytypes "seocheon/x/activity/types"
	nodetypes "seocheon/x/node/types"
)

const (
	// Total supply: 50,000 KKOT = 50,000 * 10^10 uppyeo = 5×10^14 uppyeo.
	totalSupplyKKOT   = 50_000
	uppyeoPerKKOT     = 10_000_000_000
	totalSupplyUppyeo = totalSupplyKKOT * uppyeoPerKKOT // 500,000,000,000,000

	// 3-Pool allocation percentages.
	builderPercent       = 33 // 16,500 KKOT — 4년 베스팅, 1년 클리프
	boostPoolPercent     = 27 // 13,500 KKOT — 에포크마다 점진 분배, ~2년 소진
	communityPoolPercent = 40 // 20,000 KKOT — 거버넌스로 사용처 결정

	// Builder vesting schedule.
	builderCliffDuration  = 365 * 24 * time.Hour // 1년 클리프
	builderVestingMonths  = 36                    // 클리프 후 36개월 균등 해제
	builderVestingPeriod  = 30 * 24 * time.Hour   // ~1개월
)

// GenesisAllocation holds the computed token allocations for the 3-pool model.
type GenesisAllocation struct {
	Builder       math.Int `json:"builder"`
	BoostPool     math.Int `json:"boost_pool"`
	CommunityPool math.Int `json:"community_pool"`
	Total         math.Int `json:"total"`
}

func computeAllocations() GenesisAllocation {
	total := math.NewInt(totalSupplyUppyeo)
	return GenesisAllocation{
		Builder:       total.Mul(math.NewInt(builderPercent)).Quo(math.NewInt(100)),
		BoostPool:     total.Mul(math.NewInt(boostPoolPercent)).Quo(math.NewInt(100)),
		CommunityPool: total.Mul(math.NewInt(communityPoolPercent)).Quo(math.NewInt(100)),
		Total:         total,
	}
}

// NewGenesisBuildCmd returns a command that configures a production genesis file.
func NewGenesisBuildCmd(mbm module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis-build [genesis-file]",
		Short: "Configure a production genesis file with Seocheon-specific parameters",
		Long: `Configure module parameters and token allocations for a production genesis file.

This command reads an existing genesis.json (created by 'seocheon init' and gentx collection),
applies Seocheon-specific module parameters, sets up the 3-pool allocation, and validates the result.

3-Pool Genesis Allocation (50,000 KKOT total):
  Builder (33%):               16,500 KKOT — 4년 베스팅 (1년 클리프 + 36개월 균등 해제)
  Validator Boost Pool (27%):  13,500 KKOT — 에포크마다 active validator에게 균등 분배 (~2년 소진)
  Community Pool (40%):        20,000 KKOT — 거버넌스로 사용처 결정

Parameters configured:
  x/staking:  max_validators=150, unbonding_time=21d, bond_denom=uppyeo
  x/mint:     inflation 7-15%, goal_bonded=67%, blocks_per_year=6,307,200
  x/activity: epoch=17280, windows=12, min_active=8, d_min=3000
  x/node:     registration_pool=0.1 KKOT, feegrant_pool=1 KKOT, boost_pool=13,500 KKOT
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			return runGenesisBuild(clientCtx, args[0], cmd, mbm)
		},
	}

	cmd.Flags().String("builder-address", "", "Builder vesting account address (bech32)")
	cmd.Flags().Bool("dry-run", false, "Print allocations without modifying the file")

	return cmd
}

func runGenesisBuild(clientCtx client.Context, genesisFile string, cmd *cobra.Command, mbm module.BasicManager) error {
	cdc := clientCtx.Codec

	alloc := computeAllocations()

	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// Print allocation summary.
	fmt.Println("=== Seocheon Genesis Allocation (3-Pool Model) ===")
	fmt.Printf("  Builder (33%%):             %s uppyeo (%s KKOT)\n", alloc.Builder, alloc.Builder.Quo(math.NewInt(uppyeoPerKKOT)))
	fmt.Printf("  Boost Pool (27%%):          %s uppyeo (%s KKOT)\n", alloc.BoostPool, alloc.BoostPool.Quo(math.NewInt(uppyeoPerKKOT)))
	fmt.Printf("  Community Pool (40%%):      %s uppyeo (%s KKOT)\n", alloc.CommunityPool, alloc.CommunityPool.Quo(math.NewInt(uppyeoPerKKOT)))
	fmt.Printf("  TOTAL:                     %s uppyeo (%s KKOT)\n", alloc.Total, alloc.Total.Quo(math.NewInt(uppyeoPerKKOT)))

	// Verify sum.
	sum := alloc.Builder.Add(alloc.BoostPool).Add(alloc.CommunityPool)
	if !sum.Equal(alloc.Total) {
		return fmt.Errorf("allocation sum mismatch: %s != %s", sum, alloc.Total)
	}
	fmt.Println("  Allocation sum verified OK")

	if dryRun {
		fmt.Println("\n  [dry-run] No changes written.")
		return nil
	}

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
	if err := applyNodeParams(cdc, appState, alloc); err != nil {
		return fmt.Errorf("node params: %w", err)
	}

	// Create builder vesting account.
	builderAddr, _ := cmd.Flags().GetString("builder-address")
	if builderAddr != "" {
		if err := createBuilderVestingAccount(cdc, appState, alloc, builderAddr); err != nil {
			return fmt.Errorf("builder vesting: %w", err)
		}
	} else {
		fmt.Println("  [WARN] --builder-address not set, skipping builder vesting account")
	}

	// Fund community pool.
	if err := fundCommunityPool(cdc, appState, alloc); err != nil {
		return fmt.Errorf("community pool: %w", err)
	}

	// Register denomination metadata.
	if err := registerDenomMetadata(cdc, appState); err != nil {
		return fmt.Errorf("denom metadata: %w", err)
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
	stakingGenesis.Params.BondDenom = "uppyeo"
	stakingGenesis.Params.MinCommissionRate = math.LegacyNewDecWithPrec(5, 2) // 5%

	bz, err := cdc.MarshalJSON(&stakingGenesis)
	if err != nil {
		return err
	}
	appState[stakingtypes.ModuleName] = bz
	fmt.Println("  Applied x/staking params: max_validators=150, unbonding=21d, bond_denom=uppyeo")
	return nil
}

func applyMintParams(cdc codec.Codec, appState map[string]json.RawMessage) error {
	var mintGenesis minttypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[minttypes.ModuleName], &mintGenesis); err != nil {
		return err
	}

	mintGenesis.Params.MintDenom = "uppyeo"
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
		BaseActivityFee:          10_000_000_000,     // 1 KKOT
		FeeExponent:              5000,              // 0.5
		MaxActivityFee:           1_000_000_000_000, // 100 KKOT
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

func applyNodeParams(cdc codec.Codec, appState map[string]json.RawMessage, alloc GenesisAllocation) error {
	var nodeGenesis nodetypes.GenesisState
	if raw, ok := appState[nodetypes.ModuleName]; ok {
		if err := cdc.UnmarshalJSON(raw, &nodeGenesis); err != nil {
			return err
		}
	}

	// Set pool balances.
	nodeGenesis.RegistrationPoolBalance = nodetypes.DefaultRegistrationPoolBalance // 0.1 KKOT
	nodeGenesis.FeegrantPoolBalance = nodetypes.DefaultFeegrantPoolBalance         // 1 KKOT
	nodeGenesis.BoostPoolBalance = sdk.NewCoins(
		sdk.NewCoin("uppyeo", alloc.BoostPool), // 13,500 KKOT
	)
	nodeGenesis.BoostTargetEpochs = nodetypes.DefaultBoostTargetEpochs // 730

	bz, err := cdc.MarshalJSON(&nodeGenesis)
	if err != nil {
		return err
	}
	appState[nodetypes.ModuleName] = bz
	fmt.Printf("  Applied x/node params: reg_pool=0.1 KKOT, feegrant_pool=1 KKOT, boost_pool=%s KKOT, target_epochs=%d\n",
		alloc.BoostPool.Quo(math.NewInt(uppyeoPerKKOT)), nodetypes.DefaultBoostTargetEpochs)
	return nil
}

func createBuilderVestingAccount(cdc codec.Codec, appState map[string]json.RawMessage, alloc GenesisAllocation, builderAddr string) error {
	// Validate address.
	_, err := sdk.AccAddressFromBech32(builderAddr)
	if err != nil {
		return fmt.Errorf("invalid builder address: %w", err)
	}

	// Create PeriodicVestingAccount: 1-year cliff + 36 months of equal unlocks.
	builderCoins := sdk.NewCoins(sdk.NewCoin("uppyeo", alloc.Builder))

	// Monthly unlock amount after cliff.
	monthlyUnlock := alloc.Builder.Quo(math.NewInt(builderVestingMonths))
	// Last period gets the remainder to ensure exact total.
	lastMonthUnlock := alloc.Builder.Sub(monthlyUnlock.Mul(math.NewInt(builderVestingMonths - 1)))

	periods := make([]vestingtypes.Period, 0, builderVestingMonths+1)

	// First period: 1-year cliff (0 unlock).
	periods = append(periods, vestingtypes.Period{
		Length: int64(builderCliffDuration.Seconds()),
		Amount: sdk.NewCoins(), // Nothing unlocks at cliff itself.
	})

	// 36 monthly periods after cliff.
	for i := 0; i < builderVestingMonths; i++ {
		amount := monthlyUnlock
		if i == builderVestingMonths-1 {
			amount = lastMonthUnlock
		}
		periods = append(periods, vestingtypes.Period{
			Length: int64(builderVestingPeriod.Seconds()),
			Amount: sdk.NewCoins(sdk.NewCoin("uppyeo", amount)),
		})
	}

	// Add builder account to auth genesis.
	var authGenesis authtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[authtypes.ModuleName], &authGenesis); err != nil {
		return err
	}

	genesisTime := time.Now() // Will be overridden by genesis_time in the doc.
	baseAccount := authtypes.NewBaseAccount(sdk.MustAccAddressFromBech32(builderAddr), nil, 0, 0)
	vestingAccount, err := vestingtypes.NewPeriodicVestingAccount(
		baseAccount,
		builderCoins,
		genesisTime.Unix(),
		periods,
	)
	if err != nil {
		return fmt.Errorf("failed to create vesting account: %w", err)
	}

	any, err := codectypes.NewAnyWithValue(vestingAccount)
	if err != nil {
		return fmt.Errorf("failed to pack vesting account: %w", err)
	}
	authGenesis.Accounts = append(authGenesis.Accounts, any)

	bz, err := cdc.MarshalJSON(&authGenesis)
	if err != nil {
		return err
	}
	appState[authtypes.ModuleName] = bz

	// Add builder balance to bank genesis.
	var bankGenesis banktypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return err
	}

	bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
		Address: builderAddr,
		Coins:   builderCoins,
	})

	bz, err = cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz

	fmt.Printf("  Created builder vesting account: %s (%s KKOT, 1yr cliff + 36mo unlock)\n",
		builderAddr, alloc.Builder.Quo(math.NewInt(uppyeoPerKKOT)))
	return nil
}

func fundCommunityPool(cdc codec.Codec, appState map[string]json.RawMessage, alloc GenesisAllocation) error {
	var distrGenesis distrtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[distrtypes.ModuleName], &distrGenesis); err != nil {
		return err
	}

	// Add community pool balance.
	communityPoolCoins := sdk.NewDecCoinsFromCoins(sdk.NewCoin("uppyeo", alloc.CommunityPool))
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
		Coins:   sdk.NewCoins(sdk.NewCoin("uppyeo", alloc.CommunityPool)),
	})

	bz, err = cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz

	fmt.Printf("  Funded community pool: %s KKOT\n", alloc.CommunityPool.Quo(math.NewInt(uppyeoPerKKOT)))
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
		totalFromBalances = totalFromBalances.Add(bal.Coins.AmountOf("uppyeo"))
	}

	bankGenesis.Supply = sdk.NewCoins(sdk.NewCoin("uppyeo", totalFromBalances))

	bz, err := cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz

	fmt.Printf("  Bank supply set to: %s uppyeo (%s KKOT)\n",
		totalFromBalances, totalFromBalances.Quo(math.NewInt(uppyeoPerKKOT)))
	return nil
}

func registerDenomMetadata(cdc codec.Codec, appState map[string]json.RawMessage) error {
	var bankGenesis banktypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return err
	}

	bankGenesis.DenomMetadata = []banktypes.Metadata{
		{
			Description: "The native token of the Seocheon network — 서천꽃밭(이공본풀이) 환생 순서: 뼈→살→피→숨→혼→꽃",
			DenomUnits: []*banktypes.DenomUnit{
				{Denom: "uppyeo", Exponent: 0, Aliases: []string{"뼈"}},
				{Denom: "sal", Exponent: 2, Aliases: []string{"살"}},
				{Denom: "pi", Exponent: 4, Aliases: []string{"피"}},
				{Denom: "sum", Exponent: 6, Aliases: []string{"숨"}},
				{Denom: "hon", Exponent: 8, Aliases: []string{"혼"}},
				{Denom: "kkot", Exponent: 10},
			},
			Base:    "uppyeo",
			Display: "kkot",
			Name:    "KKOT",
			Symbol:  "KKOT",
		},
	}

	bz, err := cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz
	fmt.Println("  Registered denom metadata: uppyeo (뼈, base) → sal (살, 10^2) → pi (피, 10^4) → sum (숨, 10^6) → hon (혼, 10^8) → kkot (꽃, display, 10^10)")
	return nil
}

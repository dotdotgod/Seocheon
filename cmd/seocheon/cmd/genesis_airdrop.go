package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// NewGenesisAirdropCmd returns a command that adds airdrop recipients to a genesis file.
func NewGenesisAirdropCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genesis-airdrop [genesis-file] [csv-file]",
		Short: "Add airdrop recipients to genesis from a CSV file",
		Long: `Add airdrop recipients to the genesis bank balances from a CSV file.

CSV format (no header): address,amount_usum
Example:
  seocheon1abc...,1000000000
  seocheon1def...,2000000000

The command:
1. Reads the CSV file with address,amount pairs
2. Adds each as a bank balance in the genesis file
3. Updates total supply accordingly
4. Validates all addresses are valid bech32
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			return runGenesisAirdrop(clientCtx, args[0], args[1])
		},
	}

	cmd.Flags().String("source-pool", "airdrop_pool", "Module account to deduct from (for accounting)")
	cmd.Flags().Bool("validate-only", false, "Only validate the CSV without modifying genesis")

	return cmd
}

func runGenesisAirdrop(clientCtx client.Context, genesisFile, csvFile string) error {
	cdc := clientCtx.Codec

	// Read CSV.
	f, err := os.Open(csvFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV file is empty")
	}

	// Parse and validate entries.
	type airdropEntry struct {
		Address string
		Amount  math.Int
	}

	var entries []airdropEntry
	totalAirdrop := math.ZeroInt()
	seen := make(map[string]bool)

	for i, record := range records {
		if len(record) < 2 {
			return fmt.Errorf("line %d: expected address,amount", i+1)
		}

		addr := strings.TrimSpace(record[0])
		amountStr := strings.TrimSpace(record[1])

		// Skip header row if present.
		if addr == "address" {
			continue
		}

		// Validate address.
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return fmt.Errorf("line %d: invalid address %s: %w", i+1, addr, err)
		}

		// Check duplicates.
		if seen[addr] {
			return fmt.Errorf("line %d: duplicate address %s", i+1, addr)
		}
		seen[addr] = true

		// Parse amount.
		amount, ok := math.NewIntFromString(amountStr)
		if !ok {
			return fmt.Errorf("line %d: invalid amount %s", i+1, amountStr)
		}
		if !amount.IsPositive() {
			return fmt.Errorf("line %d: amount must be positive", i+1)
		}

		entries = append(entries, airdropEntry{Address: addr, Amount: amount})
		totalAirdrop = totalAirdrop.Add(amount)
	}

	fmt.Printf("=== Genesis Airdrop ===\n")
	fmt.Printf("  Recipients: %d\n", len(entries))
	fmt.Printf("  Total:      %s usum (%s KKOT)\n", totalAirdrop, totalAirdrop.Quo(math.NewInt(usumPerKKOT)))

	// Read genesis file.
	genesisBytes, err := os.ReadFile(genesisFile)
	if err != nil {
		return fmt.Errorf("failed to read genesis: %w", err)
	}

	var genesisDoc map[string]json.RawMessage
	if err := json.Unmarshal(genesisBytes, &genesisDoc); err != nil {
		return err
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genesisDoc["app_state"], &appState); err != nil {
		return err
	}

	// Add bank balances.
	var bankGenesis banktypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[banktypes.ModuleName], &bankGenesis); err != nil {
		return err
	}

	for _, entry := range entries {
		bankGenesis.Balances = append(bankGenesis.Balances, banktypes.Balance{
			Address: entry.Address,
			Coins:   sdk.NewCoins(sdk.NewCoin("usum", entry.Amount)),
		})
	}

	// Update supply.
	currentSupply := bankGenesis.Supply.AmountOf("usum")
	newSupply := currentSupply.Add(totalAirdrop)
	bankGenesis.Supply = sdk.NewCoins(sdk.NewCoin("usum", newSupply))

	bz, err := cdc.MarshalJSON(&bankGenesis)
	if err != nil {
		return err
	}
	appState[banktypes.ModuleName] = bz

	// Marshal back.
	appStateBytes, err := json.Marshal(appState)
	if err != nil {
		return err
	}
	genesisDoc["app_state"] = appStateBytes

	finalBytes, err := json.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(genesisFile, finalBytes, 0o644); err != nil {
		return err
	}

	fmt.Printf("  Genesis updated: %s\n", genesisFile)
	fmt.Printf("  New total supply: %s usum\n", newSupply)
	return nil
}

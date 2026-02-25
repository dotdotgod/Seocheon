package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	activitytypes "seocheon/x/activity/types"
)

// NewSimulateActivityCmd returns a command that submits synthetic activities
// in configurable patterns for testnet validation.
func NewSimulateActivityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simulate-activity",
		Short: "Submit synthetic activities for testnet testing",
		Long: `Submit synthetic MsgSubmitActivity transactions to a running testnet
in configurable patterns to validate epoch/window mechanics and reward distribution.

Patterns:
  normal   - Submit 1 activity per window for 8+ windows (meets eligibility)
  sparse   - Submit in only 4 windows (does NOT meet eligibility)
  burst    - Submit max activities in a single window (spam defense test)
  mixed    - Alternate between normal and sparse across epochs

This command requires a funded account with an active feegrant or sufficient balance.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			return runSimulateActivity(clientCtx, cmd)
		},
	}

	cmd.Flags().String("pattern", "normal", "Activity pattern: normal|sparse|burst|mixed")
	cmd.Flags().Int("count", 10, "Number of activities to submit")
	cmd.Flags().Duration("interval", 2*time.Second, "Interval between submissions")
	cmd.Flags().String("content-uri-prefix", "ipfs://test/", "URI prefix for simulated content")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func runSimulateActivity(clientCtx client.Context, cmd *cobra.Command) error {
	pattern, _ := cmd.Flags().GetString("pattern")
	count, _ := cmd.Flags().GetInt("count")
	interval, _ := cmd.Flags().GetDuration("interval")
	uriPrefix, _ := cmd.Flags().GetString("content-uri-prefix")

	submitter := clientCtx.GetFromAddress().String()
	if submitter == "" {
		return fmt.Errorf("--from flag is required")
	}

	fmt.Printf("=== Activity Simulator ===\n")
	fmt.Printf("  Pattern:   %s\n", pattern)
	fmt.Printf("  Count:     %d\n", count)
	fmt.Printf("  Submitter: %s\n", submitter)
	fmt.Printf("  Interval:  %s\n", interval)
	fmt.Println()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < count; i++ {
		// Determine whether to submit based on pattern.
		if shouldSkip(pattern, i, rng) {
			fmt.Printf("  [%d/%d] Skipped (pattern: %s)\n", i+1, count, pattern)
			time.Sleep(interval)
			continue
		}

		// Generate deterministic content hash.
		content := fmt.Sprintf("testnet-activity-%s-%d-%d", submitter, time.Now().UnixNano(), i)
		hash := sha256.Sum256([]byte(content))
		activityHash := hex.EncodeToString(hash[:])
		contentURI := fmt.Sprintf("%s%s", uriPrefix, activityHash[:16])

		msg := &activitytypes.MsgSubmitActivity{
			Submitter:    submitter,
			ActivityHash: activityHash,
			ContentUri:   contentURI,
		}

		txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
		if err != nil {
			return fmt.Errorf("failed to create tx factory: %w", err)
		}
		txf = txf.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

		if err := tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg); err != nil {
			fmt.Printf("  [%d/%d] ERROR: %v\n", i+1, count, err)
		} else {
			fmt.Printf("  [%d/%d] Submitted: %s...\n", i+1, count, activityHash[:16])
		}

		if i < count-1 {
			time.Sleep(interval)
		}
	}

	fmt.Println()
	fmt.Println("=== Simulation complete ===")
	return nil
}

// shouldSkip determines if a submission should be skipped based on the pattern.
func shouldSkip(pattern string, index int, rng *rand.Rand) bool {
	switch pattern {
	case "sparse":
		// Only submit every 3rd activity — will NOT meet 8/12 window threshold.
		return index%3 != 0
	case "burst":
		// Never skip — submit all at once (tests max_activities_per_window).
		return false
	case "mixed":
		// Alternate: first half normal, second half sparse.
		if rng.Float64() < 0.3 {
			return true
		}
		return false
	default: // "normal"
		return false
	}
}

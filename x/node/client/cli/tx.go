package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetTxCmd returns the transaction commands for the node module.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        "node",
		Short:                      "Node module transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(NewRegisterNodeCmd())

	return txCmd
}

// NewRegisterNodeCmd returns a CLI command to register a new node.
func NewRegisterNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-node",
		Short: "Register a new node",
		Long: `Register a new node on the Seocheon network. This creates a validator
with 1 usum self-delegation funded from the Registration Pool.

The consensus pubkey can be obtained from CometBFT:
  seocheon comet show-validator

Example:
  seocheon tx node register-node \
    --pubkey $(seocheon comet show-validator) \
    --agent-address seocheon1... \
    --agent-share 30 \
    --description "My Node" \
    --commission-rate 0.10 \
    --commission-max-rate 0.20 \
    --commission-max-change-rate 0.01 \
    --from mykey
`,
		RunE: runRegisterNode,
	}

	cmd.Flags().String("pubkey", "", "CometBFT consensus pubkey JSON (required, use: seocheon comet show-validator)")
	cmd.Flags().String("agent-address", "", "Agent wallet address (optional)")
	cmd.Flags().String("agent-share", "0", "Agent share percentage (0-100)")
	cmd.Flags().String("max-agent-share-change-rate", "5", "Max agent share change per epoch (0-100)")
	cmd.Flags().String("description", "", "Node description")
	cmd.Flags().String("website", "", "Node website URL")
	cmd.Flags().StringSlice("tags", nil, "Node tags (comma-separated)")
	cmd.Flags().String("commission-rate", "0.10", "Initial validator commission rate")
	cmd.Flags().String("commission-max-rate", "0.20", "Maximum validator commission rate")
	cmd.Flags().String("commission-max-change-rate", "0.01", "Maximum daily commission change rate")

	flags.AddTxFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired("pubkey")

	return cmd
}

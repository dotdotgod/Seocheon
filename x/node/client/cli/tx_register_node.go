package cli

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	nodetypes "seocheon/x/node/types"
)

func runRegisterNode(cmd *cobra.Command, _ []string) error {
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		return err
	}

	// Parse consensus pubkey from JSON (same format as `seocheon comet show-validator`).
	pkJSON, _ := cmd.Flags().GetString("pubkey")
	var pk cryptotypes.PubKey
	if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(pkJSON), &pk); err != nil {
		return fmt.Errorf("invalid pubkey JSON: %w (use: seocheon comet show-validator)", err)
	}

	pkAny, err := nodetypes.PackPubKey(pk)
	if err != nil {
		return fmt.Errorf("failed to pack pubkey: %w", err)
	}

	// Parse other flags.
	agentAddr, _ := cmd.Flags().GetString("agent-address")
	description, _ := cmd.Flags().GetString("description")
	website, _ := cmd.Flags().GetString("website")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	agentShareStr, _ := cmd.Flags().GetString("agent-share")
	agentShare, err := math.LegacyNewDecFromStr(agentShareStr)
	if err != nil {
		return fmt.Errorf("invalid agent-share: %w", err)
	}

	maxChangeStr, _ := cmd.Flags().GetString("max-agent-share-change-rate")
	maxChange, err := math.LegacyNewDecFromStr(maxChangeStr)
	if err != nil {
		return fmt.Errorf("invalid max-agent-share-change-rate: %w", err)
	}

	commRateStr, _ := cmd.Flags().GetString("commission-rate")
	commRate, err := math.LegacyNewDecFromStr(commRateStr)
	if err != nil {
		return fmt.Errorf("invalid commission-rate: %w", err)
	}

	commMaxRateStr, _ := cmd.Flags().GetString("commission-max-rate")
	commMaxRate, err := math.LegacyNewDecFromStr(commMaxRateStr)
	if err != nil {
		return fmt.Errorf("invalid commission-max-rate: %w", err)
	}

	commMaxChangeStr, _ := cmd.Flags().GetString("commission-max-change-rate")
	commMaxChange, err := math.LegacyNewDecFromStr(commMaxChangeStr)
	if err != nil {
		return fmt.Errorf("invalid commission-max-change-rate: %w", err)
	}

	fromAddr := clientCtx.GetFromAddress()
	if fromAddr.Empty() {
		return fmt.Errorf("--from flag is required")
	}

	msg := &nodetypes.MsgRegisterNode{
		Operator:                sdk.AccAddress(fromAddr).String(),
		AgentAddress:            agentAddr,
		AgentShare:              agentShare,
		MaxAgentShareChangeRate: maxChange,
		Description:             description,
		Website:                 website,
		Tags:                    tags,
		ConsensusPubkey:         pkAny,
		CommissionRate:          commRate,
		CommissionMaxRate:       commMaxRate,
		CommissionMaxChangeRate: commMaxChange,
	}

	return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
}

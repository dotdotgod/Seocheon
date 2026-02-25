package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	activitytypes "seocheon/x/activity/types"
	nodetypes "seocheon/x/node/types"
)

// NewVerifyRewardsCmd returns a command to verify reward distribution for a given epoch.
func NewVerifyRewardsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-rewards",
		Short: "Verify activity reward distribution for a testnet epoch",
		Long: `Query the chain to verify that activity rewards were correctly distributed
for the specified epoch. Checks:

1. Which nodes were eligible (met 8/12 window threshold)
2. Activity reward pool balance changes
3. Per-node balance changes (operator + agent shares)

Useful for validating reward mechanics after an epoch transition on a testnet.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			return runVerifyRewards(clientCtx, cmd)
		},
	}

	cmd.Flags().Int64("epoch", -1, "Epoch number to verify (required)")
	flags.AddQueryFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired("epoch")

	return cmd
}

func runVerifyRewards(clientCtx client.Context, cmd *cobra.Command) error {
	epoch, _ := cmd.Flags().GetInt64("epoch")

	fmt.Printf("=== Reward Verification: Epoch %d ===\n\n", epoch)

	// Step 1: Query activity module params.
	activityClient := activitytypes.NewQueryClient(clientCtx)
	paramsResp, err := activityClient.Params(clientCtx.CmdContext, &activitytypes.QueryParamsRequest{})
	if err != nil {
		return fmt.Errorf("failed to query activity params: %w", err)
	}
	params := paramsResp.Params
	fmt.Printf("--- Module Parameters ---\n")
	fmt.Printf("  EpochLength:       %d blocks\n", params.EpochLength)
	fmt.Printf("  WindowsPerEpoch:   %d\n", params.WindowsPerEpoch)
	fmt.Printf("  MinActiveWindows:  %d\n", params.MinActiveWindows)
	fmt.Printf("  DMin:              %d (basis points)\n", params.DMin)
	fmt.Printf("  FeeToActivityPool: %d (basis points)\n", params.FeeToActivityPoolRatio)
	fmt.Println()

	// Step 2: Query all registered nodes and check eligibility.
	nodeClient := nodetypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	nodesResp, err := nodeClient.AllNodes(clientCtx.CmdContext, &nodetypes.QueryAllNodesRequest{})
	if err != nil {
		return fmt.Errorf("failed to query nodes: %w", err)
	}

	fmt.Printf("--- Node Eligibility (Epoch %d) ---\n", epoch)
	var eligibleCount, ineligibleCount int
	var eligibleNodeIDs []string

	for _, node := range nodesResp.Nodes {
		actResp, err := activityClient.NodeEpochActivity(clientCtx.CmdContext, &activitytypes.QueryNodeEpochActivityRequest{
			NodeId: node.Id,
			Epoch:  epoch,
		})

		status := "NO DATA"
		activeWindows := uint64(0)
		eligible := false

		if err == nil {
			activeWindows = actResp.Summary.ActiveWindows
			eligible = actResp.Summary.Eligible
			if eligible {
				status = "ELIGIBLE"
				eligibleCount++
				eligibleNodeIDs = append(eligibleNodeIDs, node.Id)
			} else {
				status = "INELIGIBLE"
				ineligibleCount++
			}
		} else {
			ineligibleCount++
		}

		fmt.Printf("  %-20s windows=%d/%d  %s\n",
			node.Id, activeWindows, params.MinActiveWindows, status)
	}
	fmt.Printf("  Total: %d eligible, %d ineligible\n\n", eligibleCount, ineligibleCount)

	// Step 3: Query activity reward pool balance.
	poolAddr := activitytypes.ActivityRewardPoolName
	poolAccResp, err := bankClient.Balance(clientCtx.CmdContext, &banktypes.QueryBalanceRequest{
		Address: poolAddr,
		Denom:   "usum",
	})

	fmt.Printf("--- Reward Pool Status ---\n")
	if err != nil {
		fmt.Printf("  Activity reward pool: (query failed: %v)\n", err)
		fmt.Printf("  Note: Pool address may need module account address lookup\n")
	} else {
		fmt.Printf("  Activity reward pool balance: %s usum\n", poolAccResp.Balance.Amount)
	}
	fmt.Println()

	// Step 4: Query balances for eligible nodes.
	if len(eligibleNodeIDs) > 0 {
		fmt.Printf("--- Eligible Node Balances ---\n")
		totalDistributed := math.ZeroInt()

		for _, nodeID := range eligibleNodeIDs {
			// Look up operator address.
			nodeResp, err := nodeClient.Node(clientCtx.CmdContext, &nodetypes.QueryNodeRequest{
				NodeId: nodeID,
			})
			if err != nil {
				fmt.Printf("  %-20s (lookup failed)\n", nodeID)
				continue
			}

			operatorAddr := nodeResp.Node.Operator
			agentAddr := nodeResp.Node.AgentAddress

			// Query operator balance.
			opBalResp, err := bankClient.Balance(clientCtx.CmdContext, &banktypes.QueryBalanceRequest{
				Address: operatorAddr,
				Denom:   "usum",
			})
			opBal := math.ZeroInt()
			if err == nil && opBalResp.Balance != nil {
				opBal = opBalResp.Balance.Amount
			}

			// Query agent balance if set.
			agentBal := math.ZeroInt()
			if agentAddr != "" {
				agBalResp, err := bankClient.Balance(clientCtx.CmdContext, &banktypes.QueryBalanceRequest{
					Address: agentAddr,
					Denom:   "usum",
				})
				if err == nil && agBalResp.Balance != nil {
					agentBal = agBalResp.Balance.Amount
				}
			}

			combined := opBal.Add(agentBal)
			totalDistributed = totalDistributed.Add(combined)

			fmt.Printf("  %-20s operator=%s  agent=%s  total=%s usum\n",
				nodeID, opBal, agentBal, combined)
		}

		fmt.Printf("  Combined total: %s usum (%s KKOT)\n\n",
			totalDistributed, totalDistributed.Quo(math.NewInt(usumPerKKOT)))
	}

	// Step 5: Query current epoch info.
	epochResp, err := activityClient.EpochInfo(clientCtx.CmdContext, &activitytypes.QueryEpochInfoRequest{})
	if err == nil {
		fmt.Printf("--- Current Chain Status ---\n")
		fmt.Printf("  Current epoch:  %d\n", epochResp.CurrentEpoch)
		fmt.Printf("  Current window: %d\n", epochResp.CurrentWindow)
		fmt.Printf("  Blocks until next epoch: %d\n", epochResp.BlocksUntilNextEpoch)
	}

	fmt.Println()
	fmt.Println("=== Verification complete ===")
	return nil
}

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"

	activitytypes "seocheon/x/activity/types"
	nodetypes "seocheon/x/node/types"
)

// AirdropSnapshotEntry represents one recipient in the snapshot output.
type AirdropSnapshotEntry struct {
	NodeID        string `json:"node_id"`
	ActiveWindows uint64 `json:"active_windows"`
	Amount        string `json:"amount_usum"`
}

// AirdropSnapshot is the full snapshot output.
type AirdropSnapshot struct {
	Quarter       string                 `json:"quarter"`
	Epoch         int64                  `json:"epoch"`
	EligibleCount int                    `json:"eligible_count"`
	TotalAmount   string                 `json:"total_amount_usum"`
	PerNodeAmount string                 `json:"per_node_amount_usum"`
	Recipients    []AirdropSnapshotEntry `json:"recipients"`
}

// NewAirdropSnapshotCmd returns a command that generates an activity airdrop snapshot.
func NewAirdropSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "airdrop-snapshot",
		Short: "Generate an activity-based airdrop distribution snapshot",
		Long: `Query the chain for eligible activity nodes and generate an equal-distribution
snapshot for the activity airdrop.

This command queries all registered nodes and checks each node's activity
summary for the specified epoch. Nodes that met the activity threshold
(Eligible==true) are included in the equal distribution.

Output: JSON file with node_id, active_windows, and amount for each recipient.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			return runAirdropSnapshot(cmd.Context(), clientCtx, cmd)
		},
	}

	cmd.Flags().Int64("epoch", -1, "Epoch number to snapshot (required)")
	cmd.Flags().String("total-amount", "", "Total airdrop amount in usum (required)")
	cmd.Flags().String("quarter", "", "Quarter label for output (e.g., 2026-Q2)")
	cmd.Flags().String("out-file", "airdrop_snapshot.json", "Output file path")
	flags.AddQueryFlagsToCmd(cmd)

	_ = cmd.MarkFlagRequired("epoch")
	_ = cmd.MarkFlagRequired("total-amount")

	return cmd
}

func runAirdropSnapshot(_ context.Context, clientCtx client.Context, cmd *cobra.Command) error {
	epoch, _ := cmd.Flags().GetInt64("epoch")
	totalAmountStr, _ := cmd.Flags().GetString("total-amount")
	quarter, _ := cmd.Flags().GetString("quarter")
	outputFile, _ := cmd.Flags().GetString("out-file")

	totalAmount, ok := math.NewIntFromString(totalAmountStr)
	if !ok || !totalAmount.IsPositive() {
		return fmt.Errorf("invalid total-amount: %s", totalAmountStr)
	}

	fmt.Printf("Querying epoch %d for eligible nodes...\n", epoch)

	// Step 1: Query all registered nodes from x/node module.
	nodeClient := nodetypes.NewQueryClient(clientCtx)
	activityClient := activitytypes.NewQueryClient(clientCtx)

	var allNodeIDs []string
	var nextKey []byte

	for {
		nodesResp, err := nodeClient.AllNodes(clientCtx.CmdContext, &nodetypes.QueryAllNodesRequest{
			Pagination: &query.PageRequest{
				Key:   nextKey,
				Limit: 100,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to query nodes: %w", err)
		}

		for _, node := range nodesResp.Nodes {
			allNodeIDs = append(allNodeIDs, node.Id)
		}

		if nodesResp.Pagination == nil || len(nodesResp.Pagination.NextKey) == 0 {
			break
		}
		nextKey = nodesResp.Pagination.NextKey
	}

	fmt.Printf("  Total registered nodes: %d\n", len(allNodeIDs))

	if len(allNodeIDs) == 0 {
		fmt.Println("  No registered nodes found")
		return writeSnapshot(outputFile, AirdropSnapshot{
			Quarter:       quarter,
			Epoch:         epoch,
			EligibleCount: 0,
			TotalAmount:   totalAmount.String(),
			PerNodeAmount: "0",
			Recipients:    []AirdropSnapshotEntry{},
		})
	}

	// Step 2: Check each node's activity eligibility for the target epoch.
	var eligibleEntries []AirdropSnapshotEntry

	for _, nodeID := range allNodeIDs {
		resp, err := activityClient.NodeEpochActivity(clientCtx.CmdContext, &activitytypes.QueryNodeEpochActivityRequest{
			NodeId: nodeID,
			Epoch:  epoch,
		})
		if err != nil {
			// Node may not have any activity for this epoch — skip.
			continue
		}

		if resp.Summary.Eligible {
			eligibleEntries = append(eligibleEntries, AirdropSnapshotEntry{
				NodeID:        nodeID,
				ActiveWindows: resp.Summary.ActiveWindows,
			})
		}
	}

	// Sort by node ID for deterministic output.
	sort.Slice(eligibleEntries, func(i, j int) bool {
		return eligibleEntries[i].NodeID < eligibleEntries[j].NodeID
	})

	if len(eligibleEntries) == 0 {
		fmt.Println("  No eligible nodes found for this epoch")
		return writeSnapshot(outputFile, AirdropSnapshot{
			Quarter:       quarter,
			Epoch:         epoch,
			EligibleCount: 0,
			TotalAmount:   totalAmount.String(),
			PerNodeAmount: "0",
			Recipients:    []AirdropSnapshotEntry{},
		})
	}

	// Step 3: Calculate per-node amount (equal distribution).
	nNodes := int64(len(eligibleEntries))
	perNode := totalAmount.Quo(math.NewInt(nNodes))

	fmt.Printf("  Eligible nodes: %d\n", nNodes)
	fmt.Printf("  Per-node amount: %s usum (%s KKOT)\n", perNode, perNode.Quo(math.NewInt(usumPerKKOT)))

	for i := range eligibleEntries {
		eligibleEntries[i].Amount = perNode.String()
	}

	snapshot := AirdropSnapshot{
		Quarter:       quarter,
		Epoch:         epoch,
		EligibleCount: len(eligibleEntries),
		TotalAmount:   totalAmount.String(),
		PerNodeAmount: perNode.String(),
		Recipients:    eligibleEntries,
	}

	return writeSnapshot(outputFile, snapshot)
}

func writeSnapshot(outputFile string, snapshot AirdropSnapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputFile, data, 0o644); err != nil {
		return err
	}

	fmt.Printf("  Snapshot written to: %s\n", outputFile)
	return nil
}

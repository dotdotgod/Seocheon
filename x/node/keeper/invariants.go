package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"seocheon/x/node/types"
)

// RegisterInvariants registers all node module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-accounts", ModuleAccountInvariant(k))
	ir.RegisterRoute(types.ModuleName, "index-consistency", IndexConsistencyInvariant(k))
	ir.RegisterRoute(types.ModuleName, "node-status", NodeStatusInvariant(k))
}

// AllInvariants runs all invariants of the node module.
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		res, stop := ModuleAccountInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		res, stop = IndexConsistencyInvariant(k)(ctx)
		if stop {
			return res, stop
		}
		return NodeStatusInvariant(k)(ctx)
	}
}

// ModuleAccountInvariant checks that Registration Pool and Feegrant Pool
// module accounts exist and have non-negative balances.
func ModuleAccountInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		regPoolAddr := k.authKeeper.GetModuleAddress(types.RegistrationPoolName)
		if regPoolAddr == nil {
			return sdk.FormatInvariant(types.ModuleName, "module-accounts",
				"registration pool module account does not exist"), true
		}
		regBalance := k.bankKeeper.GetAllBalances(ctx, regPoolAddr)
		if regBalance.IsAnyNegative() {
			return sdk.FormatInvariant(types.ModuleName, "module-accounts",
				fmt.Sprintf("registration pool has negative balance: %s", regBalance)), true
		}

		fgPoolAddr := k.authKeeper.GetModuleAddress(types.FeegrantPoolName)
		if fgPoolAddr == nil {
			return sdk.FormatInvariant(types.ModuleName, "module-accounts",
				"feegrant pool module account does not exist"), true
		}
		fgBalance := k.bankKeeper.GetAllBalances(ctx, fgPoolAddr)
		if fgBalance.IsAnyNegative() {
			return sdk.FormatInvariant(types.ModuleName, "module-accounts",
				fmt.Sprintf("feegrant pool has negative balance: %s", fgBalance)), true
		}

		return "", false
	}
}

// IndexConsistencyInvariant verifies that all node indexes (Operator, Agent, Validator)
// are bidirectionally consistent with the Nodes store.
func IndexConsistencyInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var broken string

		// Forward check: every node has correct index entries.
		nodeCount := 0
		err := k.Nodes.Walk(ctx, nil, func(nodeID string, node types.Node) (bool, error) {
			nodeCount++

			// OperatorIndex must map back to this node.
			indexedID, err := k.OperatorIndex.Get(ctx, node.Operator)
			if err != nil {
				broken = fmt.Sprintf("node %s: operator index missing for %s", nodeID, node.Operator)
				return true, nil
			}
			if indexedID != nodeID {
				broken = fmt.Sprintf("node %s: operator index maps to %s instead", nodeID, indexedID)
				return true, nil
			}

			// AgentIndex (if agent is set).
			if node.AgentAddress != "" {
				indexedID, err := k.AgentIndex.Get(ctx, node.AgentAddress)
				if err != nil {
					broken = fmt.Sprintf("node %s: agent index missing for %s", nodeID, node.AgentAddress)
					return true, nil
				}
				if indexedID != nodeID {
					broken = fmt.Sprintf("node %s: agent index maps to %s instead", nodeID, indexedID)
					return true, nil
				}
			}

			// ValidatorIndex (if validator is set).
			if node.ValidatorAddress != "" {
				indexedID, err := k.ValidatorIndex.Get(ctx, node.ValidatorAddress)
				if err != nil {
					broken = fmt.Sprintf("node %s: validator index missing for %s", nodeID, node.ValidatorAddress)
					return true, nil
				}
				if indexedID != nodeID {
					broken = fmt.Sprintf("node %s: validator index maps to %s instead", nodeID, indexedID)
					return true, nil
				}
			}

			return false, nil
		})

		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, "index-consistency",
				fmt.Sprintf("error walking nodes: %v", err)), true
		}
		if broken != "" {
			return sdk.FormatInvariant(types.ModuleName, "index-consistency", broken), true
		}

		// Reverse check: every OperatorIndex entry points to an existing node.
		operatorCount := 0
		err = k.OperatorIndex.Walk(ctx, nil, func(operator string, nodeID string) (bool, error) {
			operatorCount++
			if _, err := k.Nodes.Get(ctx, nodeID); err != nil {
				broken = fmt.Sprintf("operator index %s points to non-existent node %s", operator, nodeID)
				return true, nil
			}
			return false, nil
		})
		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, "index-consistency",
				fmt.Sprintf("error walking operator index: %v", err)), true
		}
		if broken != "" {
			return sdk.FormatInvariant(types.ModuleName, "index-consistency", broken), true
		}

		// Node count must equal operator index count (1:1 mapping).
		if nodeCount != operatorCount {
			return sdk.FormatInvariant(types.ModuleName, "index-consistency",
				fmt.Sprintf("node count (%d) != operator index count (%d)", nodeCount, operatorCount)), true
		}

		return "", false
	}
}

// NodeStatusInvariant checks that no node has an UNSPECIFIED status.
func NodeStatusInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var broken string
		err := k.Nodes.Walk(ctx, nil, func(nodeID string, node types.Node) (bool, error) {
			if node.Status == types.NodeStatus_NODE_STATUS_UNSPECIFIED {
				broken = fmt.Sprintf("node %s has UNSPECIFIED status", nodeID)
				return true, nil
			}
			if node.Operator == "" {
				broken = fmt.Sprintf("node %s has empty operator", nodeID)
				return true, nil
			}
			return false, nil
		})

		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, "node-status",
				fmt.Sprintf("error walking nodes: %v", err)), true
		}
		if broken != "" {
			return sdk.FormatInvariant(types.ModuleName, "node-status", broken), true
		}

		return "", false
	}
}

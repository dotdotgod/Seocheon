package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/types"
)

// setupNodeWithValidator creates a node in the Nodes store with the given status and validator address,
// and sets up the ValidatorIndex mapping. Returns the node ID.
func setupNodeWithValidator(t *testing.T, f *fixture, operatorName string, status types.NodeStatus) (string, sdk.ValAddress) {
	t.Helper()

	operator := testAddr(operatorName)
	operatorStr := operator.String()
	nodeID := expectedNodeID(operatorStr)

	valAddr := sdk.ValAddress(operator)
	valAddrStr, err := sdk.Bech32ifyAddressBytes("seocheonvaloper", valAddr)
	require.NoError(t, err)

	node := types.Node{
		Id:                      nodeID,
		Operator:                operatorStr,
		AgentShare:              math.LegacyNewDec(30),
		MaxAgentShareChangeRate: math.LegacyNewDec(5),
		ValidatorAddress:        valAddrStr,
		Status:                  status,
	}

	err = f.keeper.Nodes.Set(f.ctx, nodeID, node)
	require.NoError(t, err)

	err = f.keeper.ValidatorIndex.Set(f.ctx, valAddrStr, nodeID)
	require.NoError(t, err)

	return nodeID, valAddr
}

func TestAfterValidatorBonded(t *testing.T) {
	t.Run("REGISTERED node transitions to ACTIVE", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		nodeID, valAddr := setupNodeWithValidator(t, f, "hook_op_bonded____", types.NodeStatus_NODE_STATUS_REGISTERED)

		consAddr := sdk.ConsAddress(valAddr)
		err := hooks.AfterValidatorBonded(f.ctx, consAddr, valAddr)
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_ACTIVE, node.Status)
	})

	t.Run("already ACTIVE node stays ACTIVE", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		nodeID, valAddr := setupNodeWithValidator(t, f, "hook_op_active____", types.NodeStatus_NODE_STATUS_ACTIVE)

		consAddr := sdk.ConsAddress(valAddr)
		err := hooks.AfterValidatorBonded(f.ctx, consAddr, valAddr)
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_ACTIVE, node.Status) // unchanged
	})

	t.Run("INACTIVE node stays INACTIVE", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		nodeID, valAddr := setupNodeWithValidator(t, f, "hook_op_inactive__", types.NodeStatus_NODE_STATUS_INACTIVE)

		consAddr := sdk.ConsAddress(valAddr)
		err := hooks.AfterValidatorBonded(f.ctx, consAddr, valAddr)
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_INACTIVE, node.Status) // unchanged
	})

	t.Run("unknown validator silently skipped", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		unknownAddr := testAddr("unknown_validator_")
		valAddr := sdk.ValAddress(unknownAddr)
		consAddr := sdk.ConsAddress(unknownAddr)

		err := hooks.AfterValidatorBonded(f.ctx, consAddr, valAddr)
		require.NoError(t, err) // no error, silently skipped
	})
}

func TestAfterValidatorBeginUnbonding(t *testing.T) {
	t.Run("ACTIVE node transitions to REGISTERED", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		nodeID, valAddr := setupNodeWithValidator(t, f, "hook_unbond_act___", types.NodeStatus_NODE_STATUS_ACTIVE)

		consAddr := sdk.ConsAddress(valAddr)
		err := hooks.AfterValidatorBeginUnbonding(f.ctx, consAddr, valAddr)
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_REGISTERED, node.Status)
	})

	t.Run("INACTIVE node stays INACTIVE", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		nodeID, valAddr := setupNodeWithValidator(t, f, "hook_unbond_inact_", types.NodeStatus_NODE_STATUS_INACTIVE)

		consAddr := sdk.ConsAddress(valAddr)
		err := hooks.AfterValidatorBeginUnbonding(f.ctx, consAddr, valAddr)
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_INACTIVE, node.Status) // unchanged
	})

	t.Run("REGISTERED node stays REGISTERED", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		nodeID, valAddr := setupNodeWithValidator(t, f, "hook_unbond_reg___", types.NodeStatus_NODE_STATUS_REGISTERED)

		consAddr := sdk.ConsAddress(valAddr)
		err := hooks.AfterValidatorBeginUnbonding(f.ctx, consAddr, valAddr)
		require.NoError(t, err)

		node, err := f.keeper.Nodes.Get(f.ctx, nodeID)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_REGISTERED, node.Status) // unchanged
	})

	t.Run("unknown validator silently skipped", func(t *testing.T) {
		f := initFixture(t)
		hooks := f.keeper.Hooks()

		unknownAddr := testAddr("unknown_unbond____")
		valAddr := sdk.ValAddress(unknownAddr)
		consAddr := sdk.ConsAddress(unknownAddr)

		err := hooks.AfterValidatorBeginUnbonding(f.ctx, consAddr, valAddr)
		require.NoError(t, err) // no error, silently skipped
	})
}

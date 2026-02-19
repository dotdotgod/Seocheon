package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/types"
)

func TestGenesis(t *testing.T) {
	t.Run("init and export with default genesis", func(t *testing.T) {
		f := initFixture(t)

		genesisState := *types.DefaultGenesis()

		err := f.keeper.InitGenesis(f.ctx, genesisState)
		require.NoError(t, err)

		got, err := f.keeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.EqualExportedValues(t, genesisState.Params, got.Params)
		require.Empty(t, got.Nodes)
	})

	t.Run("init genesis with nodes and export round-trip", func(t *testing.T) {
		f := initFixture(t)

		node1 := types.Node{
			Id:                      "genesis-node-1",
			Operator:                "seocheon1genesisop1",
			AgentAddress:            "seocheon1genesisag1",
			AgentShare:              math.LegacyNewDec(30),
			MaxAgentShareChangeRate: math.LegacyNewDec(5),
			ValidatorAddress:        "seocheonvaloper1genesisval1",
			Tags:                    []string{"ai", "nlp"},
			Status:                  types.NodeStatus_NODE_STATUS_REGISTERED,
			RegisteredAt:            100,
		}

		node2 := types.Node{
			Id:                      "genesis-node-2",
			Operator:                "seocheon1genesisop2",
			AgentAddress:            "",
			AgentShare:              math.LegacyNewDec(0),
			MaxAgentShareChangeRate: math.LegacyNewDec(0),
			ValidatorAddress:        "seocheonvaloper1genesisval2",
			Tags:                    nil,
			Status:                  types.NodeStatus_NODE_STATUS_ACTIVE,
			RegisteredAt:            200,
		}

		genesisState := types.GenesisState{
			Params:                  types.DefaultParams(),
			Nodes:                   []types.Node{node1, node2},
			RegistrationPoolBalance: sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(500))),
			FeegrantPoolBalance:     sdk.NewCoins(sdk.NewCoin("usum", math.NewInt(50))),
		}

		err := f.keeper.InitGenesis(f.ctx, genesisState)
		require.NoError(t, err)

		// Verify nodes are stored correctly.
		storedNode1, err := f.keeper.Nodes.Get(f.ctx, "genesis-node-1")
		require.NoError(t, err)
		require.Equal(t, node1.Operator, storedNode1.Operator)
		require.Equal(t, node1.Tags, storedNode1.Tags)

		storedNode2, err := f.keeper.Nodes.Get(f.ctx, "genesis-node-2")
		require.NoError(t, err)
		require.Equal(t, node2.Operator, storedNode2.Operator)
		require.Equal(t, types.NodeStatus_NODE_STATUS_ACTIVE, storedNode2.Status)

		// Verify indexes.
		nodeID, err := f.keeper.OperatorIndex.Get(f.ctx, "seocheon1genesisop1")
		require.NoError(t, err)
		require.Equal(t, "genesis-node-1", nodeID)

		nodeID, err = f.keeper.AgentIndex.Get(f.ctx, "seocheon1genesisag1")
		require.NoError(t, err)
		require.Equal(t, "genesis-node-1", nodeID)

		nodeID, err = f.keeper.ValidatorIndex.Get(f.ctx, "seocheonvaloper1genesisval1")
		require.NoError(t, err)
		require.Equal(t, "genesis-node-1", nodeID)

		// node2 has no agent, so AgentIndex should not exist for it.
		has, err := f.keeper.AgentIndex.Has(f.ctx, "")
		require.NoError(t, err)
		require.False(t, has)

		// ExportGenesis round-trip.
		got, err := f.keeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.EqualExportedValues(t, genesisState.Params, got.Params)
		require.Len(t, got.Nodes, 2)

		// Pool balances should come from the mock bank keeper.
		// The mock has Registration Pool with 1000 usum and Feegrant Pool with 100 usum.
		require.NotNil(t, got.RegistrationPoolBalance)
		require.NotNil(t, got.FeegrantPoolBalance)
	})

	t.Run("export genesis with no nodes returns empty list", func(t *testing.T) {
		f := initFixture(t)

		got, err := f.keeper.ExportGenesis(f.ctx)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Empty(t, got.Nodes)
	})
}

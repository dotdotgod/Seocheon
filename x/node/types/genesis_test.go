package types_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/types"
)

func TestGenesisState_Validate(t *testing.T) {
	validNode := types.Node{
		Id:                      "abc123",
		Operator:                "seocheon1operator1",
		AgentAddress:            "seocheon1agent1",
		AgentShare:              math.LegacyNewDec(30),
		MaxAgentShareChangeRate: math.LegacyNewDec(5),
		ValidatorAddress:        "seocheonvaloper1val1",
		Status:                  types.NodeStatus_NODE_STATUS_REGISTERED,
	}

	validNode2 := types.Node{
		Id:                      "def456",
		Operator:                "seocheon1operator2",
		AgentAddress:            "seocheon1agent2",
		AgentShare:              math.LegacyNewDec(20),
		MaxAgentShareChangeRate: math.LegacyNewDec(3),
		ValidatorAddress:        "seocheonvaloper1val2",
		Status:                  types.NodeStatus_NODE_STATUS_ACTIVE,
	}

	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state with proper params",
			genState: &types.GenesisState{
				Params:                  types.DefaultParams(),
				Nodes:                   []types.Node{},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: true,
		},
		{
			desc: "invalid params: zero max_registrations_per_block",
			genState: &types.GenesisState{
				Params:                  types.Params{},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: false,
		},
		{
			desc: "valid genesis with nodes",
			genState: &types.GenesisState{
				Params:                  types.DefaultParams(),
				Nodes:                   []types.Node{validNode, validNode2},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: true,
		},
		{
			desc: "duplicate node IDs should fail",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Nodes: []types.Node{
					validNode,
					{
						Id:                      "abc123", // same ID as validNode
						Operator:                "seocheon1operator_dup",
						AgentShare:              math.LegacyNewDec(10),
						MaxAgentShareChangeRate: math.LegacyNewDec(2),
					},
				},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: false,
		},
		{
			desc: "duplicate operators should fail",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Nodes: []types.Node{
					validNode,
					{
						Id:                      "xyz789",
						Operator:                "seocheon1operator1", // same operator as validNode
						AgentAddress:            "seocheon1agent_other",
						AgentShare:              math.LegacyNewDec(10),
						MaxAgentShareChangeRate: math.LegacyNewDec(2),
					},
				},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: false,
		},
		{
			desc: "duplicate agent addresses should fail",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Nodes: []types.Node{
					validNode,
					{
						Id:                      "xyz789",
						Operator:                "seocheon1operator_other",
						AgentAddress:            "seocheon1agent1", // same agent as validNode
						AgentShare:              math.LegacyNewDec(10),
						MaxAgentShareChangeRate: math.LegacyNewDec(2),
					},
				},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: false,
		},
		{
			desc: "empty node ID should fail",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Nodes: []types.Node{
					{
						Id:                      "",
						Operator:                "seocheon1op",
						AgentShare:              math.LegacyNewDec(10),
						MaxAgentShareChangeRate: math.LegacyNewDec(2),
					},
				},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: false,
		},
		{
			desc: "empty operator should fail",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Nodes: []types.Node{
					{
						Id:                      "node1",
						Operator:                "",
						AgentShare:              math.LegacyNewDec(10),
						MaxAgentShareChangeRate: math.LegacyNewDec(2),
					},
				},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: false,
		},
		{
			desc: "empty agent address is valid (no agent)",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				Nodes: []types.Node{
					{
						Id:                      "node1",
						Operator:                "seocheon1op1",
						AgentAddress:            "", // no agent
						AgentShare:              math.LegacyNewDec(0),
						MaxAgentShareChangeRate: math.LegacyNewDec(0),
					},
					{
						Id:                      "node2",
						Operator:                "seocheon1op2",
						AgentAddress:            "", // also no agent - should not conflict
						AgentShare:              math.LegacyNewDec(0),
						MaxAgentShareChangeRate: math.LegacyNewDec(0),
					},
				},
				RegistrationPoolBalance: sdk.NewCoins(),
				FeegrantPoolBalance:     sdk.NewCoins(),
			},
			valid: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

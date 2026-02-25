package node

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"seocheon/x/node/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "Node",
					Use:       "node [node-id]",
					Short:     "Query a node by ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "node_id"},
					},
				},
				{
					RpcMethod: "NodeByOperator",
					Use:       "node-by-operator [operator]",
					Short:     "Query a node by operator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "operator"},
					},
				},
				{
					RpcMethod: "NodeByAgentAddress",
					Use:       "node-by-agent [agent-address]",
					Short:     "Query a node by agent address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "agent_address"},
					},
				},
				{
					RpcMethod: "NodesByTag",
					Use:       "nodes-by-tag [tag]",
					Short:     "Query nodes by tag",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "tag"},
					},
				},
				{
					RpcMethod: "AllNodes",
					Use:       "all-nodes",
					Short:     "List all registered nodes",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // authority gated
				},
				{
					RpcMethod: "RegisterNode",
					Skip:      true, // custom CLI: consensus_pubkey (proto.Any) parsing
				},
				{
					RpcMethod: "UpdateNode",
					Use:       "update-node",
					Short:     "Update node metadata",
				},
				{
					RpcMethod: "UpdateNodeAgentShare",
					Use:       "update-agent-share",
					Short:     "Update the node agent share",
				},
				{
					RpcMethod: "UpdateAgentAddress",
					Use:       "update-agent-address",
					Short:     "Update the agent wallet address",
				},
				{
					RpcMethod: "DeactivateNode",
					Use:       "deactivate-node",
					Short:     "Deactivate a node",
				},
				{
					RpcMethod: "WithdrawNodeCommission",
					Use:       "withdraw-commission",
					Short:     "Withdraw node commission",
				},
				{
					RpcMethod: "RenewFeegrant",
					Use:       "renew-feegrant",
					Short:     "Renew an expired feegrant",
				},
			},
		},
	}
}

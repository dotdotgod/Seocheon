package activity

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"seocheon/x/activity/types"
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
					RpcMethod: "Activity",
					Use:       "activity [activity-hash]",
					Short:     "Query an activity by hash",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "activity_hash"},
					},
				},
				{
					RpcMethod: "ActivitiesByNode",
					Use:       "activities-by-node [node-id]",
					Short:     "Query activities by node ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "node_id"},
					},
				},
				{
					RpcMethod: "ActivitiesByBlock",
					Use:       "activities-by-block [block-height]",
					Short:     "Query activities by block height",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "block_height"},
					},
				},
				{
					RpcMethod: "EpochInfo",
					Use:       "epoch-info",
					Short:     "Query current epoch and window information",
				},
				{
					RpcMethod: "NodeEpochActivity",
					Use:       "node-epoch-activity [node-id] [epoch]",
					Short:     "Query a node's activity summary for a given epoch",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "node_id"},
						{ProtoField: "epoch"},
					},
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
					RpcMethod: "SubmitActivity",
					Use:       "submit-activity [activity-hash] [content-uri]",
					Short:     "Submit an activity record",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "activity_hash"},
						{ProtoField: "content_uri"},
					},
				},
			},
		},
	}
}

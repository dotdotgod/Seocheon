package randomness

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"seocheon/x/randomness/types"
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
					RpcMethod: "LatestBeacon",
					Use:       "latest-beacon",
					Short:     "Query the most recently stored drand beacon",
				},
				{
					RpcMethod: "Beacon",
					Use:       "beacon [round]",
					Short:     "Query a drand beacon by round number",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "round"},
					},
				},
				{
					RpcMethod: "Beacons",
					Use:       "beacons",
					Short:     "List all stored drand beacons",
				},
				{
					RpcMethod: "RandomnessRequest",
					Use:       "randomness-request [request-id]",
					Short:     "Query a randomness request by ID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "request_id"},
					},
				},
				{
					RpcMethod: "PendingRequests",
					Use:       "pending-requests",
					Short:     "List all pending randomness requests",
				},
				{
					RpcMethod: "RequestsByRequester",
					Use:       "requests-by-requester [requester]",
					Short:     "List randomness requests by requester address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "requester"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: types.Msg_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // authority gated
				},
				{
					RpcMethod: "SubmitBeacon",
					Use:       "submit-beacon [round] [randomness] [signature]",
					Short:     "Submit a drand randomness beacon",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "round"},
						{ProtoField: "randomness"},
						{ProtoField: "signature"},
					},
				},
				{
					RpcMethod: "RequestRandomness",
					Use:       "request-randomness [commit-hash] [num-words]",
					Short:     "Submit a commit-reveal randomness request",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "commit_hash"},
						{ProtoField: "num_words"},
					},
				},
			},
		},
	}
}

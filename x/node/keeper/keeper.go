package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"seocheon/x/node/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// authority is the address capable of executing a MsgUpdateParams message.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	// Nodes stores all registered nodes indexed by node_id.
	Nodes collections.Map[string, types.Node]

	// OperatorIndex maps operator_address -> node_id (1:1).
	OperatorIndex collections.Map[string, string]

	// AgentIndex maps agent_address -> node_id (1:1).
	AgentIndex collections.Map[string, string]

	// ValidatorIndex maps validator_address -> node_id (1:1).
	ValidatorIndex collections.Map[string, string]

	// TagIndex maps (tag, node_id) -> empty value for tag-based queries.
	TagIndex collections.KeySet[collections.Pair[string, string]]

	// BlockRegistrationCount tracks registration count per block.
	BlockRegistrationCount collections.Map[int64, uint64]

	// PendingAgentShareChanges stores pending agent_share change requests.
	PendingAgentShareChanges collections.Map[string, types.PendingAgentShareChange]

	// LastAgentChangeBlock stores last agent address change block per node.
	LastAgentChangeBlock collections.Map[string, int64]

	// Keeper dependencies.
	authKeeper         types.AuthKeeper
	bankKeeper         types.BankKeeper
	stakingKeeper      types.StakingKeeper
	stakingMsgServer   types.StakingMsgServer
	distributionKeeper types.DistributionKeeper
	slashingKeeper     types.SlashingKeeper
	feegrantKeeper     types.FeegrantKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	bankKeeper types.BankKeeper,
	stakingKeeper types.StakingKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,

		Params: collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),

		Nodes:          collections.NewMap(sb, types.NodeKey, "nodes", collections.StringKey, codec.CollValue[types.Node](cdc)),
		OperatorIndex:  collections.NewMap(sb, types.OperatorIndexKey, "operator_index", collections.StringKey, collections.StringValue),
		AgentIndex:     collections.NewMap(sb, types.AgentIndexKey, "agent_index", collections.StringKey, collections.StringValue),
		ValidatorIndex: collections.NewMap(sb, types.ValidatorIndexKey, "validator_index", collections.StringKey, collections.StringValue),

		TagIndex: collections.NewKeySet(sb, types.TagIndexKey, "tag_index", collections.PairKeyCodec(collections.StringKey, collections.StringKey)),

		BlockRegistrationCount: collections.NewMap(sb, types.BlockRegistrationCountKey, "block_reg_count", collections.Int64Key, collections.Uint64Value),

		PendingAgentShareChanges: collections.NewMap(sb, types.PendingAgentShareChangeKey, "pending_agent_share", collections.StringKey, codec.CollValue[types.PendingAgentShareChange](cdc)),

		LastAgentChangeBlock: collections.NewMap(sb, types.LastAgentChangeBlockKey, "last_agent_change", collections.StringKey, collections.Int64Value),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// SetAuthKeeper sets the auth keeper (called during module wiring).
func (k *Keeper) SetAuthKeeper(ak types.AuthKeeper) {
	k.authKeeper = ak
}

// SetDistributionKeeper sets the distribution keeper.
func (k *Keeper) SetDistributionKeeper(dk types.DistributionKeeper) {
	k.distributionKeeper = dk
}

// SetSlashingKeeper sets the slashing keeper.
func (k *Keeper) SetSlashingKeeper(sk types.SlashingKeeper) {
	k.slashingKeeper = sk
}

// SetFeegrantKeeper sets the feegrant keeper.
func (k *Keeper) SetFeegrantKeeper(fk types.FeegrantKeeper) {
	k.feegrantKeeper = fk
}

// SetStakingMsgServer sets the staking module's MsgServer for CreateValidator/Undelegate.
func (k *Keeper) SetStakingMsgServer(sms types.StakingMsgServer) {
	k.stakingMsgServer = sms
}

// GetNodeIDByAgent returns the node_id for the given agent address.
// Used by x/activity to resolve agent → node.
func (k Keeper) GetNodeIDByAgent(ctx context.Context, agentAddr string) (string, error) {
	nodeID, err := k.AgentIndex.Get(ctx, agentAddr)
	if err != nil {
		return "", fmt.Errorf("agent address %s not registered: %w", agentAddr, err)
	}
	return nodeID, nil
}

// GetNodeStatus returns the status for the given node_id.
func (k Keeper) GetNodeStatus(ctx context.Context, nodeID string) (int32, error) {
	node, err := k.Nodes.Get(ctx, nodeID)
	if err != nil {
		return 0, fmt.Errorf("node %s not found: %w", nodeID, err)
	}
	return int32(node.Status), nil
}


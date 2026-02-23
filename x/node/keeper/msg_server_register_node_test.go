package keeper_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"seocheon/x/node/keeper"
	"seocheon/x/node/types"
)

// testPubKey creates a valid ed25519 consensus pubkey Any for testing.
func testPubKey(seed byte) *codectypes.Any {
	key := make([]byte, ed25519.PubKeySize)
	key[0] = seed
	pk := &ed25519.PubKey{Key: key}
	any, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		panic(err)
	}
	return any
}

// testAddr creates a deterministic test address (20 bytes padded).
func testAddr(name string) sdk.AccAddress {
	padded := make([]byte, 20)
	copy(padded, name)
	return sdk.AccAddress(padded)
}

// expectedNodeID computes the expected deterministic node ID for a given operator bech32 address.
func expectedNodeID(operator string) string {
	hash := sha256.Sum256([]byte("seocheon-node:" + operator))
	return hex.EncodeToString(hash[:16])
}

func TestMsgRegisterNode(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	operator1 := testAddr("operator1_________")
	operator1Str := operator1.String()
	agent1 := testAddr("agent1____________")
	agent1Str := agent1.String()

	operator2 := testAddr("operator2_________")
	operator2Str := operator2.String()
	agent2 := testAddr("agent2____________")
	agent2Str := agent2.String()

	t.Run("success: valid registration", func(t *testing.T) {
		msg := &types.MsgRegisterNode{
			Operator:                operator1Str,
			AgentAddress:            agent1Str,
			AgentShare:              math.LegacyNewDec(30),
			MaxAgentShareChangeRate: math.LegacyNewDec(5),
			Description:             "Test node",
			Website:                 "https://test.example.com",
			Tags:                    []string{"ai", "nlp"},
			ConsensusPubkey:         testPubKey(1),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		resp, err := ms.RegisterNode(f.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Node ID should be deterministic.
		expID := expectedNodeID(operator1Str)
		require.Equal(t, expID, resp.NodeId)
		require.NotEmpty(t, resp.ValidatorAddress)

		// Verify node is stored with REGISTERED status.
		node, err := f.keeper.Nodes.Get(f.ctx, resp.NodeId)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_REGISTERED, node.Status)
		require.Equal(t, operator1Str, node.Operator)
		require.Equal(t, agent1Str, node.AgentAddress)
		require.Equal(t, math.LegacyNewDec(30), node.AgentShare)
		require.Equal(t, []string{"ai", "nlp"}, node.Tags)

		// Verify operator index.
		nodeID, err := f.keeper.OperatorIndex.Get(f.ctx, operator1Str)
		require.NoError(t, err)
		require.Equal(t, resp.NodeId, nodeID)

		// Verify agent index.
		nodeID, err = f.keeper.AgentIndex.Get(f.ctx, agent1Str)
		require.NoError(t, err)
		require.Equal(t, resp.NodeId, nodeID)

		// Verify validator index.
		nodeID, err = f.keeper.ValidatorIndex.Get(f.ctx, resp.ValidatorAddress)
		require.NoError(t, err)
		require.Equal(t, resp.NodeId, nodeID)

		// Verify tag indexes.
		has, err := f.keeper.TagIndex.Has(f.ctx, collections.Join("ai", resp.NodeId))
		require.NoError(t, err)
		require.True(t, has)
		has, err = f.keeper.TagIndex.Has(f.ctx, collections.Join("nlp", resp.NodeId))
		require.NoError(t, err)
		require.True(t, has)
	})

	t.Run("error: duplicate operator", func(t *testing.T) {
		msg := &types.MsgRegisterNode{
			Operator:                operator1Str, // same operator as above
			AgentAddress:            agent2Str,
			AgentShare:              math.LegacyNewDec(20),
			MaxAgentShareChangeRate: math.LegacyNewDec(3),
			ConsensusPubkey:         testPubKey(2),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := ms.RegisterNode(f.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "already has a node")
	})

	t.Run("error: duplicate agent_address", func(t *testing.T) {
		msg := &types.MsgRegisterNode{
			Operator:                operator2Str, // different operator
			AgentAddress:            agent1Str,    // same agent as operator1's node
			AgentShare:              math.LegacyNewDec(20),
			MaxAgentShareChangeRate: math.LegacyNewDec(3),
			ConsensusPubkey:         testPubKey(3),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := ms.RegisterNode(f.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "already registered")
	})

	t.Run("error: registration pool depleted", func(t *testing.T) {
		// Use fresh fixture to avoid contamination.
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		// Drain the registration pool.
		regPoolAddr := ff.authKeeper.GetModuleAddress(types.RegistrationPoolName)
		ff.bankKeeper.balances[regPoolAddr.String()] = sdk.NewCoins()

		addr := testAddr("op_pool_empty_____")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(20),
			MaxAgentShareChangeRate: math.LegacyNewDec(3),
			ConsensusPubkey:         testPubKey(4),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "insufficient registration pool balance")
	})

	t.Run("error: per-block registration limit", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		params, err := ff.keeper.Params.Get(ff.ctx)
		require.NoError(t, err)

		// Register max_registrations_per_block nodes.
		for i := uint64(0); i < params.MaxRegistrationsPerBlock; i++ {
			addr := testAddr(fmt.Sprintf("blkop%d_____________", i))
			agentAddr := testAddr(fmt.Sprintf("blkag%d_____________", i))
			msg := &types.MsgRegisterNode{
				Operator:                addr.String(),
				AgentAddress:            agentAddr.String(),
				AgentShare:              math.LegacyNewDec(10),
				MaxAgentShareChangeRate: math.LegacyNewDec(2),
				ConsensusPubkey:         testPubKey(byte(10 + i)),
				CommissionRate:          math.LegacyNewDec(10),
				CommissionMaxRate:       math.LegacyNewDec(20),
				CommissionMaxChangeRate: math.LegacyNewDec(1),
			}
			_, regErr := msLocal.RegisterNode(ff.ctx, msg)
			require.NoError(t, regErr, "registration %d should succeed", i)
		}

		// Next one should fail.
		overAddr := testAddr("blkop_over________")
		overAgentAddr := testAddr("blkag_over________")
		msg := &types.MsgRegisterNode{
			Operator:                overAddr.String(),
			AgentAddress:            overAgentAddr.String(),
			AgentShare:              math.LegacyNewDec(10),
			MaxAgentShareChangeRate: math.LegacyNewDec(2),
			ConsensusPubkey:         testPubKey(99),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}
		_, err = msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "registrations")
	})

	t.Run("error: invalid agent_share negative", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_neg_share______")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(-1),
			MaxAgentShareChangeRate: math.LegacyNewDec(5),
			ConsensusPubkey:         testPubKey(50),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "out of range")
	})

	t.Run("error: invalid agent_share over 100", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_over_share_____")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(101),
			MaxAgentShareChangeRate: math.LegacyNewDec(5),
			ConsensusPubkey:         testPubKey(51),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "out of range")
	})

	t.Run("error: invalid max_agent_share_change_rate over 100", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_over_rate______")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(50),
			MaxAgentShareChangeRate: math.LegacyNewDec(101),
			ConsensusPubkey:         testPubKey(54),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "out of range")
	})

	t.Run("error: too many tags", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		params, err := ff.keeper.Params.Get(ff.ctx)
		require.NoError(t, err)

		tags := make([]string, params.MaxTags+1)
		for i := range tags {
			tags[i] = fmt.Sprintf("tag%d", i)
		}

		addr := testAddr("op_many_tags______")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(10),
			MaxAgentShareChangeRate: math.LegacyNewDec(2),
			Tags:                    tags,
			ConsensusPubkey:         testPubKey(52),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err = msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "too many tags")
	})

	t.Run("error: empty tag", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_empty_tag______")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(10),
			MaxAgentShareChangeRate: math.LegacyNewDec(2),
			Tags:                    []string{"valid", ""},
			ConsensusPubkey:         testPubKey(53),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "empty tag")
	})

	t.Run("error: nil consensus pubkey", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_no_pubkey______")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(10),
			MaxAgentShareChangeRate: math.LegacyNewDec(2),
			ConsensusPubkey:         nil,
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		_, err := msLocal.RegisterNode(ff.ctx, msg)
		require.Error(t, err)
		require.ErrorContains(t, err, "consensus pubkey")
	})

	t.Run("success: nil stakingMsgServer skips validator creation", func(t *testing.T) {
		ff := initFixture(t)
		// Clear staking msg server.
		ff.keeper.SetStakingMsgServer(nil)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_no_sms_________")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentAddress:            testAddr("ag_no_sms_________").String(),
			AgentShare:              math.LegacyNewDec(10),
			MaxAgentShareChangeRate: math.LegacyNewDec(2),
			ConsensusPubkey:         testPubKey(70),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		resp, err := msLocal.RegisterNode(ff.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotEmpty(t, resp.NodeId)

		// Node should be stored correctly.
		node, err := ff.keeper.Nodes.Get(ff.ctx, resp.NodeId)
		require.NoError(t, err)
		require.Equal(t, types.NodeStatus_NODE_STATUS_REGISTERED, node.Status)
	})

	t.Run("success: emits node_registered event", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_evt____________")
		agentAddr := testAddr("ag_evt____________")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentAddress:            agentAddr.String(),
			AgentShare:              math.LegacyNewDec(20),
			MaxAgentShareChangeRate: math.LegacyNewDec(5),
			Description:             "event test",
			ConsensusPubkey:         testPubKey(71),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		resp, err := msLocal.RegisterNode(ff.ctx, msg)
		require.NoError(t, err)

		evt := requireEvent(t, ff.ctx, types.EventTypeNodeRegistered)
		require.Equal(t, resp.NodeId, eventAttribute(evt, types.AttributeKeyNodeID))
		require.Equal(t, addr.String(), eventAttribute(evt, types.AttributeKeyOperator))
		require.Equal(t, agentAddr.String(), eventAttribute(evt, types.AttributeKeyAgentAddress))
		require.NotEmpty(t, eventAttribute(evt, types.AttributeKeyValidatorAddress))
	})

	t.Run("success: nil Dec fields default to zero", func(t *testing.T) {
		// This test covers the panic that occurred on a live chain when
		// AgentShare and MaxAgentShareChangeRate were not provided in the CLI.
		// Zero-value LegacyDec has nil *big.Int, causing IsNegative() to panic.
		// The fix adds IsNil() checks that default nil Decs to LegacyZeroDec().
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_nil_dec________")
		msg := &types.MsgRegisterNode{
			Operator:        addr.String(),
			ConsensusPubkey: testPubKey(70),
			Description:     "nil dec test",
			// AgentShare and MaxAgentShareChangeRate intentionally omitted (nil Dec).
			// CommissionRate, CommissionMaxRate, CommissionMaxChangeRate also omitted.
		}

		resp, err := msLocal.RegisterNode(ff.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify that nil Decs were defaulted to zero.
		node, err := ff.keeper.Nodes.Get(ff.ctx, resp.NodeId)
		require.NoError(t, err)
		require.True(t, node.AgentShare.IsZero())
		require.True(t, node.MaxAgentShareChangeRate.IsZero())
	})

	t.Run("success: empty agent_address (no agent)", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_no_agent_______")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentAddress:            "",
			AgentShare:              math.LegacyNewDec(0),
			MaxAgentShareChangeRate: math.LegacyNewDec(0),
			ConsensusPubkey:         testPubKey(60),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		resp, err := msLocal.RegisterNode(ff.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Agent index should not be set for empty agent_address.
		has, err := ff.keeper.AgentIndex.Has(ff.ctx, "")
		require.NoError(t, err)
		require.False(t, has)
	})

	t.Run("success: no tags", func(t *testing.T) {
		ff := initFixture(t)
		msLocal := keeper.NewMsgServerImpl(ff.keeper)

		addr := testAddr("op_no_tags________")
		msg := &types.MsgRegisterNode{
			Operator:                addr.String(),
			AgentShare:              math.LegacyNewDec(10),
			MaxAgentShareChangeRate: math.LegacyNewDec(2),
			Tags:                    []string{},
			ConsensusPubkey:         testPubKey(61),
			CommissionRate:          math.LegacyNewDec(10),
			CommissionMaxRate:       math.LegacyNewDec(20),
			CommissionMaxChangeRate: math.LegacyNewDec(1),
		}

		resp, err := msLocal.RegisterNode(ff.ctx, msg)
		require.NoError(t, err)
		require.NotNil(t, resp)

		node, err := ff.keeper.Nodes.Get(ff.ctx, resp.NodeId)
		require.NoError(t, err)
		require.Empty(t, node.Tags)
	})
}

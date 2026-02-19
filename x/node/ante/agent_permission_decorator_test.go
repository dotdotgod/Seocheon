package ante_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"seocheon/x/node/ante"
	"seocheon/x/node/types"
)

// mockNodeKeeper implements ante.NodeKeeper for testing.
type mockNodeKeeper struct {
	registeredAgents map[string]bool
	allowedMsgTypes  []string
}

func (m *mockNodeKeeper) IsRegisteredAgent(_ context.Context, address string) bool {
	return m.registeredAgents[address]
}

func (m *mockNodeKeeper) GetAllowedAgentMsgTypes(_ context.Context) ([]string, error) {
	return m.allowedMsgTypes, nil
}

// mockFeeTx implements sdk.Tx and sdk.FeeTx.
type mockFeeTx struct {
	msgs     []sdk.Msg
	feePayer sdk.AccAddress
	fee      sdk.Coins
}

func (m mockFeeTx) GetMsgs() []sdk.Msg                    { return m.msgs }
func (m mockFeeTx) GetMsgsV2() ([]protov2.Message, error)  { return nil, nil }
func (m mockFeeTx) GetGas() uint64                         { return 200000 }
func (m mockFeeTx) GetFee() sdk.Coins                      { return m.fee }
func (m mockFeeTx) FeePayer() []byte                        { return m.feePayer }
func (m mockFeeTx) FeeGranter() []byte                      { return nil }

func TestAgentPermissionDecorator(t *testing.T) {
	agentAddr := sdk.AccAddress([]byte("agent_address_______"))
	nonAgentAddr := sdk.AccAddress([]byte("normal_address______"))

	nk := &mockNodeKeeper{
		registeredAgents: map[string]bool{
			agentAddr.String(): true,
		},
		allowedMsgTypes: []string{
			"/seocheon.activity.v1.MsgSubmitActivity",
			"/cosmos.bank.v1beta1.MsgSend",
		},
	}

	decorator := ante.NewAgentPermissionDecorator(nk)

	passthrough := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		return ctx, nil
	}

	t.Run("non-agent address passes without restriction", func(t *testing.T) {
		tx := mockFeeTx{
			msgs:     []sdk.Msg{&types.MsgRegisterNode{}},
			feePayer: nonAgentAddr,
		}

		ctx := sdk.Context{}
		_, err := decorator.AnteHandle(ctx, tx, false, passthrough)
		require.NoError(t, err)
	})

	t.Run("agent with disallowed msg type fails", func(t *testing.T) {
		// MsgRegisterNode has type URL "/seocheon.node.v1.MsgRegisterNode"
		// which is NOT in the allowed list.
		tx := mockFeeTx{
			msgs:     []sdk.Msg{&types.MsgRegisterNode{}},
			feePayer: agentAddr,
		}

		ctx := sdk.Context{}
		_, err := decorator.AnteHandle(ctx, tx, false, passthrough)
		require.Error(t, err)
		require.ErrorContains(t, err, "not authorized")
	})

	t.Run("nil fee payer passes through", func(t *testing.T) {
		tx := mockFeeTx{
			msgs:     []sdk.Msg{&types.MsgRegisterNode{}},
			feePayer: nil,
		}

		ctx := sdk.Context{}
		_, err := decorator.AnteHandle(ctx, tx, false, passthrough)
		require.NoError(t, err)
	})

	t.Run("non-fee TX passes through", func(t *testing.T) {
		// mockTx does not implement FeeTx.
		tx := mockTx{
			msgs: []sdk.Msg{&types.MsgRegisterNode{}},
		}

		ctx := sdk.Context{}
		_, err := decorator.AnteHandle(ctx, tx, false, passthrough)
		require.NoError(t, err)
	})
}

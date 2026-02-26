package ante_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	protov2 "google.golang.org/protobuf/proto"

	"seocheon/x/node/ante"
	"seocheon/x/node/types"
)

// mockMsg wraps an sdk.Msg for testing.
type mockMsg struct {
	sdk.Msg
}

// mockTx implements sdk.Tx with configurable messages.
type mockTx struct {
	msgs []sdk.Msg
}

func (m mockTx) GetMsgs() []sdk.Msg                       { return m.msgs }
func (m mockTx) GetMsgsV2() ([]protov2.Message, error)     { return nil, nil }

func TestRegistrationFeeDecorator(t *testing.T) {
	decorator := ante.NewRegistrationFeeDecorator()

	var capturedCtx sdk.Context
	nextHandler := func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
		capturedCtx = ctx
		return ctx, nil
	}

	t.Run("registration-only TX sets min gas prices to zero", func(t *testing.T) {
		tx := mockTx{
			msgs: []sdk.Msg{
				&types.MsgRegisterNode{},
			},
		}

		ctx := sdk.Context{}.WithMinGasPrices(sdk.DecCoins{sdk.NewDecCoin("usum", math.NewInt(1))})
		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
		require.Empty(t, capturedCtx.MinGasPrices())
	})

	t.Run("mixed TX keeps min gas prices unchanged", func(t *testing.T) {
		tx := mockTx{
			msgs: []sdk.Msg{
				&types.MsgRegisterNode{},
				&mockMsg{}, // non-registration msg
			},
		}

		originalPrices := sdk.DecCoins{sdk.NewDecCoin("usum", math.NewInt(1))}
		ctx := sdk.Context{}.WithMinGasPrices(originalPrices)
		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
		require.Equal(t, originalPrices, capturedCtx.MinGasPrices())
	})

	t.Run("empty TX keeps min gas prices unchanged", func(t *testing.T) {
		tx := mockTx{
			msgs: []sdk.Msg{},
		}

		originalPrices := sdk.DecCoins{sdk.NewDecCoin("usum", math.NewInt(1))}
		ctx := sdk.Context{}.WithMinGasPrices(originalPrices)
		_, err := decorator.AnteHandle(ctx, tx, false, nextHandler)
		require.NoError(t, err)
		require.Equal(t, originalPrices, capturedCtx.MinGasPrices())
	})
}

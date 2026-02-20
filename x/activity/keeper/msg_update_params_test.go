package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"seocheon/x/activity/keeper"
	"seocheon/x/activity/types"
)

func TestUpdateParams_Success(t *testing.T) {
	f := initFixture(t)

	authority := authtypes.NewModuleAddress(types.GovModuleName)
	authorityStr, _ := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), authority)

	msgServer := keeper.NewMsgServerImpl(f.keeper)
	newParams := types.DefaultParams()
	newParams.EpochLength = 8640

	_, err := msgServer.UpdateParams(f.ctx, &types.MsgUpdateParams{
		Authority: authorityStr,
		Params:    newParams,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify updated params.
	got, err := f.keeper.Params.Get(f.ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got.EpochLength != 8640 {
		t.Errorf("expected epoch_length 8640, got %d", got.EpochLength)
	}
}

func TestUpdateParams_InvalidAuthority(t *testing.T) {
	f := initFixture(t)

	randomAddr := sdk.AccAddress([]byte("random_______________")).String()

	msgServer := keeper.NewMsgServerImpl(f.keeper)
	_, err := msgServer.UpdateParams(f.ctx, &types.MsgUpdateParams{
		Authority: randomAddr,
		Params:    types.DefaultParams(),
	})
	if err == nil {
		t.Fatal("expected error for invalid authority")
	}
}

func TestUpdateParams_InvalidParams(t *testing.T) {
	f := initFixture(t)

	authority := authtypes.NewModuleAddress(types.GovModuleName)
	authorityStr, _ := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), authority)

	msgServer := keeper.NewMsgServerImpl(f.keeper)
	badParams := types.DefaultParams()
	badParams.EpochLength = 0 // invalid

	_, err := msgServer.UpdateParams(f.ctx, &types.MsgUpdateParams{
		Authority: authorityStr,
		Params:    badParams,
	})
	if err == nil {
		t.Fatal("expected error for invalid params")
	}
}

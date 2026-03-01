package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"seocheon/x/randomness/keeper"
	"seocheon/x/randomness/types"
)

func TestValidateRandomnessFormat(t *testing.T) {
	tests := []struct {
		name       string
		randomness string
		valid      bool
	}{
		{"valid 32-byte hex", "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", true},
		{"valid all zeros", "0000000000000000000000000000000000000000000000000000000000000000", true},
		{"too short", "a1b2c3d4", false},
		{"too long", "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2ff", false},
		{"invalid hex chars", "g1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := keeper.ValidateRandomnessFormat(tc.randomness)
			require.Equal(t, tc.valid, result)
		})
	}
}

func TestValidateSignatureFormat(t *testing.T) {
	tests := []struct {
		name      string
		signature string
		valid     bool
	}{
		{"valid hex", "a1b2c3d4e5f6", true},
		{"valid long hex", "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", true},
		{"invalid hex", "xyz123", false},
		{"empty", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := keeper.ValidateSignatureFormat(tc.signature)
			require.Equal(t, tc.valid, result)
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()

	require.False(t, params.BeaconVerificationEnabled)
	require.Equal(t, uint64(17280), params.MaxBeaconAgeBlocks)
	require.NotEmpty(t, params.DrandChainHash)
	require.NoError(t, params.Validate())
}

func TestGenesisValidation(t *testing.T) {
	tests := []struct {
		name    string
		genesis types.GenesisState
		valid   bool
	}{
		{
			"default genesis",
			*types.DefaultGenesis(),
			true,
		},
		{
			"genesis with beacons",
			types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.Beacon{
					{Round: 1, Randomness: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", Signature: "sig1", SubmittedAt: 100, Submitter: "addr1"},
					{Round: 2, Randomness: "b1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", Signature: "sig2", SubmittedAt: 200, Submitter: "addr2"},
				},
				NextRequestId: 1,
			},
			true,
		},
		{
			"duplicate round",
			types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.Beacon{
					{Round: 1, Randomness: "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", Signature: "sig1"},
					{Round: 1, Randomness: "b1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2", Signature: "sig2"},
				},
			},
			false,
		},
		{
			"zero round",
			types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.Beacon{
					{Round: 0, Randomness: "abc", Signature: "sig"},
				},
			},
			false,
		},
		{
			"empty randomness",
			types.GenesisState{
				Params: types.DefaultParams(),
				Beacons: []types.Beacon{
					{Round: 1, Randomness: "", Signature: "sig"},
				},
			},
			false,
		},
		{
			"invalid params",
			types.GenesisState{
				Params:  types.Params{MaxBeaconAgeBlocks: 0, DrandPeriodSeconds: 30},
				Beacons: []types.Beacon{},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

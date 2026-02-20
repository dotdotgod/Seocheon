package types_test

import (
	"strings"
	"testing"

	"seocheon/x/activity/types"
)

func TestDefaultGenesis(t *testing.T) {
	gs := types.DefaultGenesis()
	if err := gs.Validate(); err != nil {
		t.Fatalf("DefaultGenesis should be valid: %v", err)
	}
	if len(gs.Activities) != 0 {
		t.Errorf("expected empty activities, got %d", len(gs.Activities))
	}
}

func TestGenesisValidate(t *testing.T) {
	validHash := strings.Repeat("ab", 32) // 64 hex chars

	tests := []struct {
		name    string
		gs      types.GenesisState
		wantErr bool
	}{
		{
			name:    "valid default",
			gs:      *types.DefaultGenesis(),
			wantErr: false,
		},
		{
			name: "invalid params",
			gs: types.GenesisState{
				Params:     types.Params{EpochLength: 0},
				Activities: []types.ActivityRecord{},
			},
			wantErr: true,
		},
		{
			name: "valid with activities",
			gs: types.GenesisState{
				Params: types.DefaultParams(),
				Activities: []types.ActivityRecord{
					{
						NodeId:       "node1",
						Epoch:        1,
						Sequence:     0,
						Submitter:    "addr1",
						ActivityHash: validHash,
						ContentUri:   "ipfs://Qm123",
						BlockHeight:  100,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty node_id",
			gs: types.GenesisState{
				Params: types.DefaultParams(),
				Activities: []types.ActivityRecord{
					{
						NodeId:       "",
						ActivityHash: validHash,
						ContentUri:   "ipfs://Qm123",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty activity_hash",
			gs: types.GenesisState{
				Params: types.DefaultParams(),
				Activities: []types.ActivityRecord{
					{
						NodeId:       "node1",
						ActivityHash: "",
						ContentUri:   "ipfs://Qm123",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid activity_hash length",
			gs: types.GenesisState{
				Params: types.DefaultParams(),
				Activities: []types.ActivityRecord{
					{
						NodeId:       "node1",
						ActivityHash: "abcd",
						ContentUri:   "ipfs://Qm123",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty content_uri",
			gs: types.GenesisState{
				Params: types.DefaultParams(),
				Activities: []types.ActivityRecord{
					{
						NodeId:       "node1",
						ActivityHash: validHash,
						ContentUri:   "",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate hash in same epoch",
			gs: types.GenesisState{
				Params: types.DefaultParams(),
				Activities: []types.ActivityRecord{
					{
						NodeId:       "node1",
						Epoch:        1,
						Sequence:     0,
						ActivityHash: validHash,
						ContentUri:   "ipfs://Qm1",
					},
					{
						NodeId:       "node1",
						Epoch:        1,
						Sequence:     1,
						ActivityHash: validHash,
						ContentUri:   "ipfs://Qm2",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "same hash different epochs is valid",
			gs: types.GenesisState{
				Params: types.DefaultParams(),
				Activities: []types.ActivityRecord{
					{
						NodeId:       "node1",
						Epoch:        1,
						Sequence:     0,
						ActivityHash: validHash,
						ContentUri:   "ipfs://Qm1",
					},
					{
						NodeId:       "node1",
						Epoch:        2,
						Sequence:     0,
						ActivityHash: validHash,
						ContentUri:   "ipfs://Qm2",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.gs.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

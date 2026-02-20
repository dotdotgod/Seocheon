package types_test

import (
	"testing"

	"seocheon/x/activity/types"
)

func TestDefaultParams(t *testing.T) {
	p := types.DefaultParams()
	if err := p.Validate(); err != nil {
		t.Fatalf("DefaultParams should be valid: %v", err)
	}
}

func TestParamsValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*types.Params)
		wantErr bool
	}{
		{
			name:    "valid default",
			modify:  func(p *types.Params) {},
			wantErr: false,
		},
		{
			name:    "zero epoch_length",
			modify:  func(p *types.Params) { p.EpochLength = 0 },
			wantErr: true,
		},
		{
			name:    "negative epoch_length",
			modify:  func(p *types.Params) { p.EpochLength = -1 },
			wantErr: true,
		},
		{
			name:    "zero windows_per_epoch",
			modify:  func(p *types.Params) { p.WindowsPerEpoch = 0 },
			wantErr: true,
		},
		{
			name:    "zero min_active_windows",
			modify:  func(p *types.Params) { p.MinActiveWindows = 0 },
			wantErr: true,
		},
		{
			name:    "min_active_windows exceeds windows_per_epoch",
			modify:  func(p *types.Params) { p.MinActiveWindows = 13 },
			wantErr: true,
		},
		{
			name:    "epoch_length not divisible by windows_per_epoch",
			modify:  func(p *types.Params) { p.EpochLength = 17281 },
			wantErr: true,
		},
		{
			name:    "zero self_funded_quota",
			modify:  func(p *types.Params) { p.SelfFundedQuota = 0 },
			wantErr: true,
		},
		{
			name:    "zero feegrant_quota",
			modify:  func(p *types.Params) { p.FeegrantQuota = 0 },
			wantErr: true,
		},
		{
			name:    "zero pruning_keep_blocks",
			modify:  func(p *types.Params) { p.ActivityPruningKeepBlocks = 0 },
			wantErr: true,
		},
		{
			name:    "negative pruning_keep_blocks",
			modify:  func(p *types.Params) { p.ActivityPruningKeepBlocks = -1 },
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := types.DefaultParams()
			tc.modify(&p)
			err := p.Validate()
			if tc.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

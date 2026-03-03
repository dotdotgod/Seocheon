package constants

import (
	"testing"
)

// Constants validation tests for spec alignment

func TestEpochWindowRelationship(t *testing.T) {
	// WindowLength should equal EpochLength / WindowsPerEpoch
	expected := EpochLength / WindowsPerEpoch
	if WindowLength != expected {
		t.Errorf("WindowLength = %d, want %d (EpochLength/WindowsPerEpoch)", WindowLength, expected)
	}
}

func TestDenominationFactors(t *testing.T) {
	// Verify denomination chain: each is 100x the previous
	if UppyeoPerSal != 100 {
		t.Errorf("UppyeoPerSal = %d, want 100", UppyeoPerSal)
	}
	if UppyeoPerPi != 10000 {
		t.Errorf("UppyeoPerPi = %d, want 10000", UppyeoPerPi)
	}
	if UppyeoPerSum != 1000000 {
		t.Errorf("UppyeoPerSum = %d, want 1000000", UppyeoPerSum)
	}
	if UppyeoPerHon != 100000000 {
		t.Errorf("UppyeoPerHon = %d, want 100000000", UppyeoPerHon)
	}
	if UppyeoPerKkot != 10000000000 {
		t.Errorf("UppyeoPerKkot = %d, want 10000000000", UppyeoPerKkot)
	}
}

func TestMinActiveWindowsLessThanTotal(t *testing.T) {
	if MinActiveWindows >= WindowsPerEpoch {
		t.Errorf("MinActiveWindows(%d) should be < WindowsPerEpoch(%d)",
			MinActiveWindows, WindowsPerEpoch)
	}
}

func TestDenomStrings(t *testing.T) {
	if TokenBaseDenom != "uppyeo" {
		t.Errorf("TokenBaseDenom = %s, want uppyeo", TokenBaseDenom)
	}
	if TokenDisplayDenom != "kkot" {
		t.Errorf("TokenDisplayDenom = %s, want kkot", TokenDisplayDenom)
	}
}

func TestQuotaValues(t *testing.T) {
	if SelfFundedQuota != 100 {
		t.Errorf("SelfFundedQuota = %d, want 100", SelfFundedQuota)
	}
	if FeegrantQuota != 10 {
		t.Errorf("FeegrantQuota = %d, want 10", FeegrantQuota)
	}
	if SelfFundedQuota <= FeegrantQuota {
		t.Error("SelfFundedQuota should be greater than FeegrantQuota")
	}
}

func TestAgentAllowedMsgTypes(t *testing.T) {
	if len(AgentAllowedMsgTypes) != 2 {
		t.Errorf("AgentAllowedMsgTypes has %d entries, want 2", len(AgentAllowedMsgTypes))
	}
	// Should include MsgSubmitActivity
	found := false
	for _, msg := range AgentAllowedMsgTypes {
		if msg == "/seocheon.activity.v1.MsgSubmitActivity" {
			found = true
			break
		}
	}
	if !found {
		t.Error("AgentAllowedMsgTypes should include MsgSubmitActivity")
	}
}

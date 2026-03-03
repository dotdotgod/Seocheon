package tx

import (
	"testing"
)

// Additional gas/fee tests for spec alignment

func TestDefaultGasSubmitActivity(t *testing.T) {
	gas := DefaultGasForMessage("/seocheon.activity.v1.MsgSubmitActivity")
	if gas != 200000 {
		t.Errorf("gas for MsgSubmitActivity = %d, want 200000", gas)
	}
}

func TestDefaultGasMsgSend(t *testing.T) {
	gas := DefaultGasForMessage("/cosmos.bank.v1beta1.MsgSend")
	if gas != 100000 {
		t.Errorf("gas for MsgSend = %d, want 100000", gas)
	}
}

func TestDefaultGasWithdraw(t *testing.T) {
	gas := DefaultGasForMessage("/seocheon.node.v1.MsgWithdrawNodeCommission")
	if gas != 300000 {
		t.Errorf("gas for MsgWithdrawNodeCommission = %d, want 300000", gas)
	}
}

func TestDefaultGasUnknown(t *testing.T) {
	gas := DefaultGasForMessage("/unknown.MsgType")
	if gas != 200000 {
		t.Errorf("gas for unknown message = %d, want 200000 (fallback)", gas)
	}
}

func TestCalculateFeeNonZero(t *testing.T) {
	fee := CalculateFee(100000, 250)
	expected := uint64(25000000)
	if fee != expected {
		t.Errorf("CalculateFee(100000, 250) = %d, want %d", fee, expected)
	}
}

func TestCalculateFeeLargeValues(t *testing.T) {
	fee := CalculateFee(300000, 250)
	expected := uint64(75000000)
	if fee != expected {
		t.Errorf("CalculateFee(300000, 250) = %d, want %d", fee, expected)
	}
}

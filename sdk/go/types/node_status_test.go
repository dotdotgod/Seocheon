package types

import (
	"testing"
)

func TestNodeStatusFromInt(t *testing.T) {
	tests := []struct {
		input    int
		expected NodeStatus
	}{
		{0, NodeStatusUnspecified},
		{1, NodeStatusRegistered},
		{2, NodeStatusActive},
		{3, NodeStatusInactive},
		{4, NodeStatusJailed},
		{99, NodeStatusUnspecified},
	}

	for _, tt := range tests {
		result := NodeStatusFromInt(tt.input)
		if result != tt.expected {
			t.Errorf("NodeStatusFromInt(%d) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

package utility

import (
	"testing"

	"cosmossdk.io/math"
)

func TestConvertDenom_UppyeoToKkot(t *testing.T) {
	amount := math.NewInt(10000000000)
	result, err := ConvertDenom(amount, "uppyeo", "kkot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(1)) {
		t.Errorf("expected 1 kkot, got %s", result.String())
	}
}

func TestConvertDenom_KkotToUppyeo(t *testing.T) {
	amount := math.NewInt(1)
	result, err := ConvertDenom(amount, "kkot", "uppyeo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(10000000000)) {
		t.Errorf("expected 10000000000 uppyeo, got %s", result.String())
	}
}

func TestConvertDenom_UppyeoToHon(t *testing.T) {
	amount := math.NewInt(100000000)
	result, err := ConvertDenom(amount, "uppyeo", "hon")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(1)) {
		t.Errorf("expected 1 hon, got %s", result.String())
	}
}

func TestConvertDenom_HonToKkot(t *testing.T) {
	amount := math.NewInt(100)
	result, err := ConvertDenom(amount, "hon", "kkot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(1)) {
		t.Errorf("expected 1 kkot, got %s", result.String())
	}
}

func TestConvertDenom_UppyeoToSal(t *testing.T) {
	amount := math.NewInt(100)
	result, err := ConvertDenom(amount, "uppyeo", "sal")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(1)) {
		t.Errorf("expected 1 sal, got %s", result.String())
	}
}

func TestConvertDenom_UppyeoToPi(t *testing.T) {
	amount := math.NewInt(10000)
	result, err := ConvertDenom(amount, "uppyeo", "pi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(1)) {
		t.Errorf("expected 1 pi, got %s", result.String())
	}
}

func TestConvertDenom_UppyeoToSum(t *testing.T) {
	amount := math.NewInt(1000000)
	result, err := ConvertDenom(amount, "uppyeo", "sum")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(1)) {
		t.Errorf("expected 1 sum, got %s", result.String())
	}
}

func TestConvertDenom_SameDenom(t *testing.T) {
	amount := math.NewInt(42)
	result, err := ConvertDenom(amount, "uppyeo", "uppyeo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Equal(math.NewInt(42)) {
		t.Errorf("expected 42, got %s", result.String())
	}
}

func TestConvertDenom_InvalidDenom(t *testing.T) {
	amount := math.NewInt(100)
	_, err := ConvertDenom(amount, "invalid", "uppyeo")
	if err == nil {
		t.Fatal("expected error for invalid denomination")
	}
}

func TestConvertDenom_LargeAmount(t *testing.T) {
	// 50,000 KKOT = 500,000,000,000,000 uppyeo (5×10^14)
	amount := math.NewInt(50000)
	result, err := ConvertDenom(amount, "kkot", "uppyeo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := math.NewInt(500_000_000_000_000)
	if !result.Equal(expected) {
		t.Errorf("expected %s uppyeo, got %s", expected.String(), result.String())
	}
}

func TestFormatKkot(t *testing.T) {
	tests := []struct {
		uppyeo   int64
		expected string
	}{
		{0, "0.0000000000"},
		{1, "0.0000000001"},
		{10000000000, "1.0000000000"},
		{15000000000, "1.5000000000"},
		{500_000_000_000_000, "50000.0000000000"},
		{1234567890, "0.1234567890"},
	}

	for _, tt := range tests {
		result := FormatKkot(tt.uppyeo)
		if result != tt.expected {
			t.Errorf("FormatKkot(%d) = %s, want %s", tt.uppyeo, result, tt.expected)
		}
	}
}

func TestParseKkot(t *testing.T) {
	tests := []struct {
		kkot     string
		expected int64
		wantErr  bool
	}{
		{"0.0000000000", 0, false},
		{"1.0000000000", 10000000000, false},
		{"1.5", 15000000000, false},
		{"0.1234567890", 1234567890, false},
		{"0.0000000001", 1, false},
		{"50000.0000000000", 500_000_000_000_000, false},
		{"invalid", 0, true},
		{"1.2.3", 0, true},
	}

	for _, tt := range tests {
		result, err := ParseKkot(tt.kkot)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseKkot(%s) expected error, got %d", tt.kkot, result)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseKkot(%s) unexpected error: %v", tt.kkot, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("ParseKkot(%s) = %d, want %d", tt.kkot, result, tt.expected)
		}
	}
}

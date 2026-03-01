// Package utility provides helper functions for the Seocheon SDK,
// including denomination conversion and activity hash verification.
package utility

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"

	"github.com/seocheon/sdk-go/constants"
)

// ConvertDenom converts an amount between Seocheon token denominations.
// Supported denominations: "uppyeo" (base), "sal", "pi", "sum", "hon", "kkot" (display).
func ConvertDenom(amount math.Int, from, to string) (math.Int, error) {
	fromFactor, err := denomFactor(from)
	if err != nil {
		return math.Int{}, err
	}
	toFactor, err := denomFactor(to)
	if err != nil {
		return math.Int{}, err
	}

	if from == to {
		return amount, nil
	}

	// Convert to base (uppyeo) first, then to target
	baseAmount := amount.Mul(math.NewInt(fromFactor))
	if toFactor == 1 {
		return baseAmount, nil
	}

	result := baseAmount.Quo(math.NewInt(toFactor))
	return result, nil
}

// FormatKkot converts an uppyeo amount to a human-readable KKOT string.
// Example: 10000000000 uppyeo → "1.0000000000"
func FormatKkot(uppyeoAmount int64) string {
	intPart := uppyeoAmount / constants.UppyeoPerKkot
	decPart := uppyeoAmount % constants.UppyeoPerKkot
	if decPart < 0 {
		decPart = -decPart
	}
	return fmt.Sprintf("%d.%010d", intPart, decPart)
}

// ParseKkot parses a KKOT string to uppyeo amount.
// Example: "1.0000000000" → 10000000000
func ParseKkot(kkot string) (int64, error) {
	parts := strings.Split(kkot, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("invalid kkot format: %s", kkot)
	}

	intPart := int64(0)
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid character in kkot integer part: %c", c)
		}
		intPart = intPart*10 + int64(c-'0')
	}

	decPart := int64(0)
	if len(parts) == 2 {
		dec := parts[1]
		// Pad or truncate to 10 decimal places
		for len(dec) < 10 {
			dec += "0"
		}
		dec = dec[:10]
		for _, c := range dec {
			if c < '0' || c > '9' {
				return 0, fmt.Errorf("invalid character in kkot decimal part: %c", c)
			}
			decPart = decPart*10 + int64(c-'0')
		}
	}

	return intPart*constants.UppyeoPerKkot + decPart, nil
}

// denomFactor returns the base-unit multiplier for the given denomination.
func denomFactor(denom string) (int64, error) {
	switch denom {
	case constants.TokenBaseDenom: // "uppyeo"
		return 1, nil
	case constants.TokenSalDenom: // "sal"
		return constants.UppyeoPerSal, nil
	case constants.TokenPiDenom: // "pi"
		return constants.UppyeoPerPi, nil
	case constants.TokenSumDenom: // "sum"
		return constants.UppyeoPerSum, nil
	case constants.TokenHonDenom: // "hon"
		return constants.UppyeoPerHon, nil
	case constants.TokenDisplayDenom: // "kkot"
		return constants.UppyeoPerKkot, nil
	default:
		return 0, fmt.Errorf("unknown denomination: %s (supported: uppyeo, sal, pi, sum, hon, kkot)", denom)
	}
}

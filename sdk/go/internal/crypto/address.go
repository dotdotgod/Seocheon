package crypto

import (
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/ripemd160"
)

const (
	// Bech32PrefixAccAddr is the Bech32 prefix for account addresses.
	Bech32PrefixAccAddr = "seocheon1"
)

// AddressFromPubKey derives a Cosmos bech32 address from a compressed public key.
func AddressFromPubKey(pubKey []byte) (string, error) {
	if len(pubKey) != 33 {
		return "", fmt.Errorf("public key must be 33 bytes (compressed), got %d", len(pubKey))
	}

	// SHA256 hash
	sha := sha256.Sum256(pubKey)

	// RIPEMD160 hash
	rip := ripemd160.New()
	rip.Write(sha[:])
	addrBytes := rip.Sum(nil) // 20 bytes

	// Bech32 encode with "seocheon" HRP (produces "seocheon1..." address)
	bech32Addr, err := bech32Encode("seocheon", addrBytes)
	if err != nil {
		return "", fmt.Errorf("bech32 encoding: %w", err)
	}

	return bech32Addr, nil
}

// bech32Encode encodes data to bech32 format.
// This is a minimal implementation for Cosmos address encoding.
func bech32Encode(hrp string, data []byte) (string, error) {
	// Convert 8-bit data to 5-bit groups
	converted, err := convertBits(data, 8, 5, true)
	if err != nil {
		return "", fmt.Errorf("converting bits: %w", err)
	}

	// Calculate checksum
	values := append(converted, 0, 0, 0, 0, 0, 0)
	polymod := bech32Polymod(expandHRP(hrp, values)) ^ 1
	for i := 0; i < 6; i++ {
		values[len(converted)+i] = byte((polymod >> uint(5*(5-i))) & 31)
	}

	// Encode to string
	charset := "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	result := hrp + "1"
	for _, v := range values {
		result += string(charset[v])
	}

	return result, nil
}

// convertBits converts a byte slice from one bit-group size to another.
func convertBits(data []byte, fromBits, toBits uint, pad bool) ([]byte, error) {
	acc := uint32(0)
	bits := uint(0)
	var result []byte
	maxV := uint32((1 << toBits) - 1)

	for _, b := range data {
		acc = (acc << fromBits) | uint32(b)
		bits += fromBits
		for bits >= toBits {
			bits -= toBits
			result = append(result, byte((acc>>bits)&maxV))
		}
	}

	if pad {
		if bits > 0 {
			result = append(result, byte((acc<<(toBits-bits))&maxV))
		}
	} else if bits >= fromBits {
		return nil, fmt.Errorf("illegal zero padding")
	} else if ((acc << (toBits - bits)) & maxV) != 0 {
		return nil, fmt.Errorf("non-zero padding")
	}

	return result, nil
}

// expandHRP expands the human-readable part for checksum computation.
func expandHRP(hrp string, values []byte) []byte {
	result := make([]byte, 0, len(hrp)*2+1+len(values))
	for _, c := range hrp {
		result = append(result, byte(c>>5))
	}
	result = append(result, 0)
	for _, c := range hrp {
		result = append(result, byte(c&31))
	}
	result = append(result, values...)
	return result
}

// bech32Polymod computes the bech32 checksum.
func bech32Polymod(values []byte) uint32 {
	gen := [5]uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := uint32(1)
	for _, v := range values {
		top := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(v)
		for i := 0; i < 5; i++ {
			if (top>>uint(i))&1 == 1 {
				chk ^= gen[i]
			}
		}
	}
	return chk
}

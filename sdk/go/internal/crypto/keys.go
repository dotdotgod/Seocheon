// Package crypto provides secp256k1 key management, BIP44 derivation,
// and Cosmos-compatible address encoding for the Seocheon SDK.
package crypto

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
)

// PrivateKey wraps a secp256k1 private key.
type PrivateKey struct {
	key *btcec.PrivateKey
}

// NewPrivateKeyFromBtcec creates a PrivateKey from a btcec private key.
func NewPrivateKeyFromBtcec(key *btcec.PrivateKey) *PrivateKey {
	return &PrivateKey{key: key}
}

// NewPrivateKeyFromBytes creates a PrivateKey from raw 32-byte private key.
func NewPrivateKeyFromBytes(privKeyBytes []byte) (*PrivateKey, error) {
	if len(privKeyBytes) != 32 {
		return nil, fmt.Errorf("private key must be 32 bytes, got %d", len(privKeyBytes))
	}
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)
	return &PrivateKey{key: privKey}, nil
}

// PubKey returns the compressed 33-byte public key.
func (pk *PrivateKey) PubKey() []byte {
	return pk.key.PubKey().SerializeCompressed()
}

// Bytes returns the raw 32-byte private key.
func (pk *PrivateKey) Bytes() []byte {
	return pk.key.Serialize()
}

// Sign signs the given data using secp256k1. It hashes the data with SHA-256
// first and returns a 64-byte compact signature (R || S).
func (pk *PrivateKey) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)

	// SignCompact returns [recoveryFlag, R (32 bytes), S (32 bytes)] = 65 bytes
	compactSig := ecdsa.SignCompact(pk.key, hash[:], false)

	// Strip recovery flag (first byte), return 64-byte R || S
	return compactSig[1:], nil
}

// Verify verifies a 64-byte compact signature against data using the public key.
func Verify(pubKeyBytes, data, sigBytes []byte) bool {
	if len(sigBytes) != 64 {
		return false
	}

	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return false
	}

	hash := sha256.Sum256(data)

	// Parse R and S from compact format into a DER signature for verification.
	// Build a DER-encoded signature from the 64-byte compact R || S.
	r := sigBytes[:32]
	s := sigBytes[32:]

	derSig := marshalDER(r, s)
	sig, err := ecdsa.ParseDERSignature(derSig)
	if err != nil {
		return false
	}
	return sig.Verify(hash[:], pubKey)
}

// marshalDER encodes R and S (each 32 bytes, big-endian) into DER format.
func marshalDER(r, s []byte) []byte {
	rEnc := derInteger(r)
	sEnc := derInteger(s)

	// SEQUENCE { rEnc, sEnc }
	seq := append(rEnc, sEnc...)
	return append([]byte{0x30, byte(len(seq))}, seq...)
}

// derInteger encodes a big-endian unsigned integer as a DER INTEGER.
func derInteger(b []byte) []byte {
	// Strip leading zero bytes
	for len(b) > 1 && b[0] == 0 {
		b = b[1:]
	}
	// If the high bit is set, prepend a zero byte
	if len(b) > 0 && b[0]&0x80 != 0 {
		b = append([]byte{0}, b...)
	}
	return append([]byte{0x02, byte(len(b))}, b...)
}

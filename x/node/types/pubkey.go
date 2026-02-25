package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// PackPubKey wraps a cryptotypes.PubKey into a google.protobuf.Any.
func PackPubKey(pk cryptotypes.PubKey) (*codectypes.Any, error) {
	return codectypes.NewAnyWithValue(pk)
}

// Ensure MsgRegisterNode implements UnpackInterfacesMessage so that
// the consensus_pubkey Any field is properly unpacked during deserialization.
// Without this, GetCachedValue() on the Any returns nil, causing staking's
// MsgCreateValidator to fail with "Expecting cryptotypes.PubKey, got <nil>".
var _ codectypes.UnpackInterfacesMessage = (*MsgRegisterNode)(nil)

// UnpackInterfaces implements codectypes.UnpackInterfacesMessage.
func (msg *MsgRegisterNode) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	if msg.ConsensusPubkey == nil {
		return nil
	}
	var pk cryptotypes.PubKey
	return unpacker.UnpackAny(msg.ConsensusPubkey, &pk)
}

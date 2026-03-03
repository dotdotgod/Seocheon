package tx

// google.protobuf.Any encoding
// Fields: type_url(1, string), value(2, bytes)
func encodeAny(typeURL string, value []byte) []byte {
	return ConcatBytes(
		EncodeFieldString(1, typeURL),
		EncodeFieldBytes(2, value),
	)
}

// cosmos.crypto.secp256k1.PubKey encoding
// type_url: "/cosmos.crypto.secp256k1.PubKey"
// Fields: key(1, bytes)
func encodePubKeyAny(pubKey []byte) []byte {
	innerMsg := EncodeFieldBytes(1, pubKey)
	return encodeAny("/cosmos.crypto.secp256k1.PubKey", innerMsg)
}

// ModeInfo for SIGN_MODE_DIRECT
// ModeInfo { single { mode: SIGN_MODE_DIRECT (1) } }
// ModeInfo: field 1 = Single (message)
// Single: field 1 = mode (varint, SIGN_MODE_DIRECT = 1)
func encodeModeInfoDirect() []byte {
	modeField := EncodeFieldVarint(1, 1) // SIGN_MODE_DIRECT = 1
	single := EncodeFieldBytes(1, modeField)
	return single
}

// SignerInfo encoding
// Fields: public_key(1, Any), mode_info(2, ModeInfo), sequence(3, uint64)
func encodeSignerInfo(pubKey []byte, sequence uint64) []byte {
	pubKeyAny := encodePubKeyAny(pubKey)
	modeInfo := encodeModeInfoDirect()

	return ConcatBytes(
		EncodeFieldBytes(1, pubKeyAny),
		EncodeFieldBytes(2, modeInfo),
		EncodeFieldVarint(3, sequence),
	)
}

// Fee encoding
// Fields: amount(1, repeated Coin), gas_limit(2, uint64)
func encodeFee(coins []Coin, gasLimit uint64) []byte {
	parts := make([][]byte, 0, len(coins)+1)
	for _, coin := range coins {
		coinBytes := coin.Encode()
		parts = append(parts, EncodeFieldBytes(1, coinBytes))
	}
	parts = append(parts, EncodeFieldVarint(2, gasLimit))
	return ConcatBytes(parts...)
}

// EncodeTxBody encodes a TxBody.
// Fields: messages(1, repeated Any), memo(2, string), timeout_height(3, uint64)
func EncodeTxBody(messages []MessageEncoder, memo string, timeoutHeight uint64) []byte {
	parts := make([][]byte, 0, len(messages)+2)
	for _, msg := range messages {
		anyBytes := encodeAny(msg.TypeURL(), msg.Encode())
		parts = append(parts, EncodeFieldBytes(1, anyBytes))
	}
	if memo != "" {
		parts = append(parts, EncodeFieldString(2, memo))
	}
	if timeoutHeight > 0 {
		parts = append(parts, EncodeFieldVarint(3, timeoutHeight))
	}
	return ConcatBytes(parts...)
}

// EncodeAuthInfo encodes an AuthInfo.
// Fields: signer_infos(1, repeated SignerInfo), fee(2, Fee)
func EncodeAuthInfo(pubKey []byte, sequence uint64, feeCoins []Coin, gasLimit uint64) []byte {
	signerInfo := encodeSignerInfo(pubKey, sequence)
	fee := encodeFee(feeCoins, gasLimit)

	return ConcatBytes(
		EncodeFieldBytes(1, signerInfo),
		EncodeFieldBytes(2, fee),
	)
}

// EncodeSignDoc encodes a SignDoc for SIGN_MODE_DIRECT.
// Fields: body_bytes(1, bytes), auth_info_bytes(2, bytes), chain_id(3, string), account_number(4, uint64)
func EncodeSignDoc(bodyBytes, authInfoBytes []byte, chainID string, accountNumber uint64) []byte {
	return ConcatBytes(
		EncodeFieldBytes(1, bodyBytes),
		EncodeFieldBytes(2, authInfoBytes),
		EncodeFieldString(3, chainID),
		EncodeFieldVarint(4, accountNumber),
	)
}

// EncodeTxRaw encodes a TxRaw for broadcast.
// Fields: body_bytes(1, bytes), auth_info_bytes(2, bytes), signatures(3, repeated bytes)
func EncodeTxRaw(bodyBytes, authInfoBytes []byte, signatures ...[]byte) []byte {
	parts := [][]byte{
		EncodeFieldBytes(1, bodyBytes),
		EncodeFieldBytes(2, authInfoBytes),
	}
	for _, sig := range signatures {
		parts = append(parts, EncodeFieldBytes(3, sig))
	}
	return ConcatBytes(parts...)
}

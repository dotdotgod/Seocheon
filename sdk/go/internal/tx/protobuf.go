package tx

// Minimal protobuf wire format encoder.
// Only supports the types needed for Cosmos TX encoding:
// - varint (wire type 0)
// - length-delimited (wire type 2)

// EncodeVarint encodes a uint64 as a protobuf varint.
func EncodeVarint(v uint64) []byte {
	var buf [10]byte
	n := 0
	for v >= 0x80 {
		buf[n] = byte(v) | 0x80
		v >>= 7
		n++
	}
	buf[n] = byte(v)
	return buf[:n+1]
}

// EncodeFieldVarint encodes a varint field (wire type 0).
func EncodeFieldVarint(fieldNumber int, value uint64) []byte {
	if value == 0 {
		return nil // protobuf default: omit zero
	}
	tag := EncodeVarint(uint64(fieldNumber<<3 | 0))
	return append(tag, EncodeVarint(value)...)
}

// EncodeFieldBytes encodes a length-delimited field (wire type 2).
func EncodeFieldBytes(fieldNumber int, data []byte) []byte {
	if len(data) == 0 {
		return nil // protobuf default: omit empty
	}
	tag := EncodeVarint(uint64(fieldNumber<<3 | 2))
	length := EncodeVarint(uint64(len(data)))
	result := make([]byte, 0, len(tag)+len(length)+len(data))
	result = append(result, tag...)
	result = append(result, length...)
	result = append(result, data...)
	return result
}

// EncodeFieldString encodes a string field (wire type 2).
func EncodeFieldString(fieldNumber int, value string) []byte {
	if value == "" {
		return nil
	}
	return EncodeFieldBytes(fieldNumber, []byte(value))
}

// EncodeSignedVarint encodes an int64 using zigzag encoding for sint64.
// Note: Cosmos uses regular varint for int64 fields, not zigzag.
// This function encodes a signed int64 as an unsigned varint (standard proto int64 behavior).
func EncodeSignedVarint(v int64) []byte {
	return EncodeVarint(uint64(v))
}

// ConcatBytes concatenates multiple byte slices, skipping nils.
func ConcatBytes(parts ...[]byte) []byte {
	size := 0
	for _, p := range parts {
		size += len(p)
	}
	result := make([]byte, 0, size)
	for _, p := range parts {
		result = append(result, p...)
	}
	return result
}

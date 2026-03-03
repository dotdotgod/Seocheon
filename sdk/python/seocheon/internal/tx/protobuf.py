"""Minimal protobuf wire format encoder.

Only supports the types needed for Cosmos TX encoding:
- varint (wire type 0)
- length-delimited (wire type 2)
"""


def encode_varint(v: int) -> bytes:
    """Encode a uint64 as a protobuf varint."""
    buf = bytearray()
    while v >= 0x80:
        buf.append((v & 0x7F) | 0x80)
        v >>= 7
    buf.append(v & 0x7F)
    return bytes(buf)


def encode_field_varint(field_number: int, value: int) -> bytes:
    """Encode a varint field (wire type 0). Omits if value is 0."""
    if value == 0:
        return b""
    tag = encode_varint((field_number << 3) | 0)
    return tag + encode_varint(value)


def encode_field_bytes(field_number: int, data: bytes) -> bytes:
    """Encode a length-delimited field (wire type 2). Omits if data is empty."""
    if not data:
        return b""
    tag = encode_varint((field_number << 3) | 2)
    length = encode_varint(len(data))
    return tag + length + data


def encode_field_string(field_number: int, value: str) -> bytes:
    """Encode a string field (wire type 2). Omits if value is empty."""
    if not value:
        return b""
    return encode_field_bytes(field_number, value.encode("utf-8"))


def concat_bytes(*parts: bytes) -> bytes:
    """Concatenate multiple byte sequences, skipping empty ones."""
    return b"".join(p for p in parts if p)

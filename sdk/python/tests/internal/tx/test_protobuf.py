"""Tests for minimal protobuf encoder."""

from seocheon.internal.tx.protobuf import (
    concat_bytes,
    encode_field_string,
    encode_field_varint,
    encode_varint,
)


def test_encode_varint_zero():
    assert encode_varint(0) == b"\x00"


def test_encode_varint_one():
    assert encode_varint(1) == b"\x01"


def test_encode_varint_127():
    assert encode_varint(127) == b"\x7f"


def test_encode_varint_128():
    assert encode_varint(128) == b"\x80\x01"


def test_encode_field_varint_zero_omitted():
    assert encode_field_varint(1, 0) == b""


def test_encode_field_varint_nonzero():
    result = encode_field_varint(1, 1)
    # tag = (1 << 3 | 0) = 8, value = 1
    assert result == b"\x08\x01"


def test_encode_field_string_empty_omitted():
    assert encode_field_string(1, "") == b""


def test_encode_field_string_hello():
    result = encode_field_string(1, "hello")
    # tag = 10, length = 5, "hello" = 68 65 6c 6c 6f
    assert result == b"\x0a\x05hello"


def test_concat_bytes_skips_empty():
    assert concat_bytes(b"\x01", b"", b"\x03") == b"\x01\x03"

"""Tests for secp256k1 key management."""

from seocheon.internal.crypto.keys import PrivateKey, verify


def test_create_from_bytes():
    priv_bytes = bytes.fromhex("0000000000000000000000000000000000000000000000000000000000000001")
    key = PrivateKey.from_bytes(priv_bytes)
    assert len(key.raw_bytes()) == 32
    assert len(key.pub_key()) == 33


def test_pub_key_is_compressed():
    priv_bytes = bytes.fromhex("0000000000000000000000000000000000000000000000000000000000000001")
    key = PrivateKey.from_bytes(priv_bytes)
    pub = key.pub_key()
    assert pub[0] in (0x02, 0x03)  # compressed prefix


def test_sign_produces_64_bytes():
    priv_bytes = bytes.fromhex("0000000000000000000000000000000000000000000000000000000000000001")
    key = PrivateKey.from_bytes(priv_bytes)
    sig = key.sign(b"hello world")
    assert len(sig) == 64


def test_sign_and_verify():
    priv_bytes = bytes.fromhex("0000000000000000000000000000000000000000000000000000000000000001")
    key = PrivateKey.from_bytes(priv_bytes)
    data = b"test data for signing"
    sig = key.sign(data)
    assert verify(key.pub_key(), data, sig) is True


def test_verify_rejects_wrong_data():
    priv_bytes = bytes.fromhex("0000000000000000000000000000000000000000000000000000000000000001")
    key = PrivateKey.from_bytes(priv_bytes)
    data = b"test data"
    # Wrong signature bytes
    assert verify(key.pub_key(), data, b"\x00" * 64) is False
    # Invalid signature length
    assert verify(key.pub_key(), b"data", b"\x00" * 32) is False


def test_invalid_key_length():
    import pytest
    with pytest.raises(ValueError, match="32 bytes"):
        PrivateKey.from_bytes(b"\x00" * 16)

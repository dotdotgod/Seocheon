"""Tests for BIP44 key derivation."""

from seocheon.internal.crypto.bip44 import derive_key_from_mnemonic


STANDARD_MNEMONIC = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"


def test_derive_key_produces_32_bytes():
    key = derive_key_from_mnemonic(STANDARD_MNEMONIC)
    assert len(key.raw_bytes()) == 32


def test_derive_key_produces_compressed_pubkey():
    key = derive_key_from_mnemonic(STANDARD_MNEMONIC)
    pub = key.pub_key()
    assert len(pub) == 33
    assert pub[0] in (0x02, 0x03)


def test_derive_key_deterministic():
    key1 = derive_key_from_mnemonic(STANDARD_MNEMONIC)
    key2 = derive_key_from_mnemonic(STANDARD_MNEMONIC)
    assert key1.raw_bytes() == key2.raw_bytes()


def test_different_mnemonic_different_key():
    mnemonic2 = "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"
    key1 = derive_key_from_mnemonic(STANDARD_MNEMONIC)
    key2 = derive_key_from_mnemonic(mnemonic2)
    assert key1.raw_bytes() != key2.raw_bytes()


def test_derive_invalid_mnemonic():
    import pytest

    with pytest.raises(Exception):
        derive_key_from_mnemonic("invalid garbage words that are not a real mnemonic phrase")

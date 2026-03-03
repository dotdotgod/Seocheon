"""Tests for bech32 address encoding."""

from seocheon.internal.crypto.address import address_from_pubkey
from seocheon.internal.crypto.bip44 import derive_key_from_mnemonic


STANDARD_MNEMONIC = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"


def test_address_from_pubkey_format():
    key = derive_key_from_mnemonic(STANDARD_MNEMONIC)
    addr = address_from_pubkey(key.pub_key())
    assert addr.startswith("seocheon1")

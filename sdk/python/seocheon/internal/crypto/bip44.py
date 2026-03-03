"""BIP44 key derivation for the Seocheon SDK."""

from __future__ import annotations

from bip_utils import (
    Bip39SeedGenerator,
    Bip39WordsNum,
    Bip44,
    Bip44Changes,
    Bip44Coins,
)

from seocheon.internal.crypto.keys import PrivateKey


def derive_key_from_mnemonic(mnemonic: str) -> PrivateKey:
    """Derive a secp256k1 private key from a BIP39 mnemonic.

    Uses the Cosmos BIP44 path: m/44'/118'/0'/0/0
    """
    seed = Bip39SeedGenerator(mnemonic).Generate()
    bip44_ctx = Bip44.FromSeed(seed, Bip44Coins.COSMOS)
    key = (
        bip44_ctx
        .Purpose()
        .Coin()
        .Account(0)
        .Change(Bip44Changes.CHAIN_EXT)
        .AddressIndex(0)
    )
    priv_key_bytes = key.PrivateKey().Raw().ToBytes()
    return PrivateKey.from_bytes(priv_key_bytes)

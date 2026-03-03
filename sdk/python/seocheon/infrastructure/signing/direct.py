"""DirectSigningService: signs transactions using a mnemonic directly (test only)."""

from __future__ import annotations

from seocheon.internal.crypto.address import address_from_pubkey
from seocheon.internal.crypto.bip44 import derive_key_from_mnemonic
from seocheon.internal.crypto.keys import PrivateKey


class DirectSigningService:
    """Signs transactions using a mnemonic directly (test only)."""

    def __init__(self, mnemonic: str) -> None:
        if not mnemonic:
            raise ValueError("mnemonic is required")
        self._priv_key: PrivateKey = derive_key_from_mnemonic(mnemonic)
        self._pub_key: bytes = self._priv_key.pub_key()
        self._address: str = address_from_pubkey(self._pub_key)

    def sign(self, tx_bytes: bytes) -> bytes:
        return self._priv_key.sign(tx_bytes)

    def get_address(self) -> str:
        return self._address

    def get_pub_key(self) -> bytes:
        return self._pub_key

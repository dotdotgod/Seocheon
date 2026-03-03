"""SigningService protocol definition."""

from __future__ import annotations

from typing import Protocol


class SigningService(Protocol):
    """Defines the interface for transaction signing."""

    def sign(self, tx_bytes: bytes) -> bytes:
        """Sign the given transaction bytes and return the signature."""
        ...

    def get_address(self) -> str:
        """Return the signer's address."""
        ...

    def get_pub_key(self) -> bytes:
        """Return the signer's public key bytes."""
        ...

"""TX types and protocol interfaces."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Protocol

from seocheon.internal.tx.messages import MessageEncoder


@dataclass
class TxEventAttr:
    """A key-value pair in a TxEvent."""

    key: str
    value: str


@dataclass
class TxEventData:
    """Represents a transaction event."""

    type: str
    attributes: list[TxEventAttr] = field(default_factory=list)


@dataclass
class TxResult:
    """Holds the result of a broadcast and confirmed transaction."""

    tx_hash: str = ""
    height: int = 0
    code: int = 0
    gas_used: int = 0
    gas_wanted: int = 0
    raw_log: str = ""
    events: list[TxEventData] = field(default_factory=list)


@dataclass
class TxRequest:
    """Holds the parameters for building and broadcasting a transaction."""

    message: MessageEncoder
    memo: str = ""
    timeout_height: int = 0
    gas_limit: int = 0
    fee_amount: int = 0
    fee_denom: str = ""


class Signer(Protocol):
    """Abstracts the signing capability needed by the TX pipeline."""

    def sign(self, data: bytes) -> bytes:
        """Sign the given bytes and return the signature."""
        ...

    def get_address(self) -> str:
        """Return the signer's bech32 address."""
        ...

    def get_pub_key(self) -> bytes:
        """Return the signer's compressed public key bytes."""
        ...


class ChainQuerier(Protocol):
    """Abstracts the chain queries needed by the TX pipeline."""

    async def get_account_info(self, address: str) -> tuple[int, int]:
        """Return (account_number, sequence) for an address."""
        ...

    async def broadcast_tx_sync(self, tx_bytes: bytes) -> tuple[str, int, str]:
        """Broadcast a TX and return (tx_hash, code, raw_log)."""
        ...

    async def get_tx_result(self, tx_hash: str) -> TxResult:
        """Query a TX result by hash."""
        ...

"""ChainClient to ChainQuerier adapter."""

from __future__ import annotations

from typing import TYPE_CHECKING

from seocheon.internal.tx.types import TxEventAttr, TxEventData, TxResult

if TYPE_CHECKING:
    from seocheon.infrastructure.chain_client import ChainClient


class ChainClientAdapter:
    """Adapts ChainClient to the ChainQuerier interface."""

    def __init__(self, client: ChainClient) -> None:
        self._client = client

    async def get_account_info(self, address: str) -> tuple[int, int]:
        info = await self._client.get_account_info(address)
        return info["account_number"], info["sequence"]

    async def broadcast_tx_sync(self, tx_bytes: bytes) -> tuple[str, int, str]:
        resp = await self._client.broadcast_tx(tx_bytes, "sync")
        return resp["tx_hash"], resp["code"], resp["raw_log"]

    async def get_tx_result(self, tx_hash: str) -> TxResult:
        resp = await self._client.get_tx(tx_hash)
        events = [
            TxEventData(
                type=e["type"],
                attributes=[TxEventAttr(key=a["key"], value=a["value"]) for a in e.get("attributes", [])],
            )
            for e in resp.get("events", [])
        ]
        return TxResult(
            tx_hash=resp.get("tx_hash", ""),
            height=resp.get("height", 0),
            code=resp.get("code", 0),
            gas_used=resp.get("gas_used", 0),
            gas_wanted=resp.get("gas_wanted", 0),
            raw_log=resp.get("raw_log", ""),
            events=events,
        )

"""Chain client interface and HTTP implementation."""

from __future__ import annotations

import base64
import json
from typing import Any, Protocol

import aiohttp


class ChainClient(Protocol):
    """Defines the interface for chain communication."""

    async def connect(self) -> None: ...
    async def disconnect(self) -> None: ...
    def is_connected(self) -> bool: ...
    async def query_rest(self, path: str) -> Any: ...
    async def broadcast_tx(self, tx_bytes: bytes, mode: str) -> dict[str, Any]: ...
    async def get_latest_block(self) -> dict[str, Any]: ...
    async def get_tx(self, tx_hash: str) -> dict[str, Any]: ...
    async def get_account_info(self, address: str) -> dict[str, Any]: ...


class HTTPChainClient:
    """HTTP-based implementation of ChainClient."""

    def __init__(self, rpc_endpoint: str, grpc_endpoint: str) -> None:
        self._rpc_endpoint = rpc_endpoint.rstrip("/")
        self._grpc_endpoint = grpc_endpoint.rstrip("/")
        self._session: aiohttp.ClientSession | None = None
        self._connected = False

    async def connect(self) -> None:
        """Test the connection to the chain node."""
        self._session = aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=30))
        await self.get_latest_block()
        self._connected = True

    async def disconnect(self) -> None:
        """Close the HTTP client."""
        self._connected = False
        if self._session:
            await self._session.close()
            self._session = None

    def is_connected(self) -> bool:
        return self._connected

    def _get_session(self) -> aiohttp.ClientSession:
        if self._session is None:
            self._session = aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=30))
        return self._session

    async def query_rest(self, path: str) -> Any:
        """Perform a GET request to the REST endpoint."""
        session = self._get_session()
        url = self._grpc_endpoint + path
        async with session.get(url) as resp:
            body = await resp.text()
            if resp.status != 200:
                raise RuntimeError(f"query failed with status {resp.status}: {body}")
            return json.loads(body)

    async def broadcast_tx(self, tx_bytes: bytes, mode: str) -> dict[str, Any]:
        """Broadcast a signed transaction."""
        session = self._get_session()
        proto_mode = "BROADCAST_MODE_ASYNC" if mode == "async" else "BROADCAST_MODE_SYNC"
        payload = {
            "tx_bytes": base64.b64encode(tx_bytes).decode("ascii"),
            "mode": proto_mode,
        }

        url = self._grpc_endpoint + "/cosmos/tx/v1beta1/txs"
        async with session.post(url, json=payload) as resp:
            body = await resp.text()
            data = json.loads(body)

        tx_response = data.get("tx_response", {})
        return {
            "tx_hash": tx_response.get("txhash", ""),
            "code": tx_response.get("code", 0),
            "raw_log": tx_response.get("raw_log", ""),
        }

    async def get_latest_block(self) -> dict[str, Any]:
        """Return the latest block information."""
        data = await self.query_rest("/cosmos/base/tendermint/v1beta1/blocks/latest")
        block = data.get("block", {})
        header = block.get("header", {})
        txs = block.get("data", {}).get("txs", [])
        return {
            "height": int(header.get("height", 0)),
            "time": header.get("time", ""),
            "chain_id": header.get("chain_id", ""),
            "num_txs": len(txs),
        }

    async def get_tx(self, tx_hash: str) -> dict[str, Any]:
        """Query a transaction by hash."""
        data = await self.query_rest(f"/cosmos/tx/v1beta1/txs/{tx_hash}")
        tx_response = data.get("tx_response", {})
        return {
            "tx_hash": tx_response.get("txhash", ""),
            "height": int(tx_response.get("height", 0)),
            "code": tx_response.get("code", 0),
            "gas_used": int(tx_response.get("gas_used", 0)),
            "gas_wanted": int(tx_response.get("gas_wanted", 0)),
            "raw_log": tx_response.get("raw_log", ""),
            "events": [
                {
                    "type": e.get("type", ""),
                    "attributes": [
                        {"key": a.get("key", ""), "value": a.get("value", "")}
                        for a in e.get("attributes", [])
                    ],
                }
                for e in tx_response.get("events", [])
            ],
        }

    async def get_account_info(self, address: str) -> dict[str, Any]:
        """Return the account number and sequence for an address."""
        data = await self.query_rest(f"/cosmos/auth/v1beta1/accounts/{address}")
        account = data.get("account", {})
        return {
            "account_number": int(account.get("account_number", 0)),
            "sequence": int(account.get("sequence", 0)),
        }

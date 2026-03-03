"""VaultSigningService: signs transactions via an external vault server (production)."""

from __future__ import annotations

import aiohttp


class VaultSigningService:
    """Signs transactions via an external vault server."""

    def __init__(self, endpoint: str, key_name: str) -> None:
        if not endpoint or not key_name:
            raise ValueError("vault endpoint and key name are required")
        self._endpoint = endpoint.rstrip("/")
        self._key_name = key_name
        self._address: str = ""
        self._pub_key: bytes = b""

    async def initialize(self) -> None:
        """Connect to the vault server to fetch address and public key."""
        timeout = aiohttp.ClientTimeout(total=10)
        async with aiohttp.ClientSession(timeout=timeout) as session:
            # Fetch address
            addr_url = f"{self._endpoint}/v1/keys/{self._key_name}/address"
            async with session.get(addr_url) as resp:
                if resp.status != 200:
                    body = await resp.text()
                    raise RuntimeError(f"vault address request failed: {body}")
                data = await resp.json()
                self._address = data["address"]

            # Fetch public key
            pk_url = f"{self._endpoint}/v1/keys/{self._key_name}/pubkey"
            async with session.get(pk_url) as resp:
                if resp.status != 200:
                    body = await resp.text()
                    raise RuntimeError(f"vault pubkey request failed: {body}")
                data = await resp.json()
                self._pub_key = bytes.fromhex(data["pubkey"])

    def sign(self, tx_bytes: bytes) -> bytes:
        """Sign data via vault. Must call initialize() first in an async context.

        For synchronous use, call sign_sync. For async, call sign_async.
        """
        raise NotImplementedError("Use sign_async for vault signing")

    async def sign_async(self, tx_bytes: bytes) -> bytes:
        """Sign data via the vault server asynchronously."""
        timeout = aiohttp.ClientTimeout(total=10)
        url = f"{self._endpoint}/v1/keys/{self._key_name}/sign"
        payload = {"data": tx_bytes.hex()}

        async with aiohttp.ClientSession(timeout=timeout) as session:
            async with session.post(url, json=payload) as resp:
                if resp.status != 200:
                    body = await resp.text()
                    raise RuntimeError(f"vault sign failed with status {resp.status}: {body}")
                data = await resp.json()
                return bytes.fromhex(data["signature"])

    def get_address(self) -> str:
        return self._address

    def get_pub_key(self) -> bytes:
        return self._pub_key

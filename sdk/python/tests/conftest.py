"""Common test fixtures."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest


@pytest.fixture
def mock_chain_client():
    """Create a mock chain client."""
    client = MagicMock()
    client.connect = AsyncMock()
    client.disconnect = AsyncMock()
    client.is_connected = MagicMock(return_value=True)
    client.query_rest = AsyncMock()
    client.broadcast_tx = AsyncMock()
    client.get_latest_block = AsyncMock(return_value={
        "height": 17281,
        "time": "2026-01-01T00:00:00Z",
        "chain_id": "seocheon-testnet-1",
        "num_txs": 5,
    })
    client.get_tx = AsyncMock()
    client.get_account_info = AsyncMock(return_value={
        "account_number": 1,
        "sequence": 5,
    })
    return client


@pytest.fixture
def mock_signer():
    """Create a mock signing service."""
    signer = MagicMock()
    signer.sign = MagicMock(return_value=b"\x00" * 64)
    signer.get_address = MagicMock(return_value="seocheon1testaddress")
    signer.get_pub_key = MagicMock(return_value=b"\x02" + b"\x00" * 32)
    return signer


TEST_MNEMONIC = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

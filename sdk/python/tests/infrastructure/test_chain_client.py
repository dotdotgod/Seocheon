"""Tests for HTTPChainClient."""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from seocheon.infrastructure.chain_client import HTTPChainClient


@pytest.fixture
def client():
    return HTTPChainClient("http://localhost:26657", "http://localhost:1317")


def test_init(client):
    assert client._rpc_endpoint == "http://localhost:26657"
    assert client._grpc_endpoint == "http://localhost:1317"
    assert client.is_connected() is False


def test_init_strips_trailing_slash():
    c = HTTPChainClient("http://localhost:26657/", "http://localhost:1317/")
    assert c._rpc_endpoint == "http://localhost:26657"
    assert c._grpc_endpoint == "http://localhost:1317"


async def test_disconnect(client):
    client._session = MagicMock()
    client._session.close = AsyncMock()
    client._connected = True
    await client.disconnect()
    assert client.is_connected() is False
    assert client._session is None


def test_get_session_creates_session(client):
    assert client._session is None
    session = client._get_session()
    assert session is not None

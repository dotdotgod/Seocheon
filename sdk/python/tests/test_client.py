"""Tests for the SeocheonSDK client."""

from unittest.mock import AsyncMock, patch

import pytest

from seocheon.client import SeocheonSDK
from seocheon.config import ChainConfig, SDKConfig, SigningConfig, SigningMode, TxConfig
from seocheon.errors.errors import ErrInvalidConfig


TEST_MNEMONIC = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"


def _make_config(mnemonic: str = TEST_MNEMONIC) -> SDKConfig:
    return SDKConfig(
        chain=ChainConfig(
            chain_id="seocheon-test-1",
            rpc_endpoint="http://localhost:26657",
            grpc_endpoint="http://localhost:1317",
        ),
        signing=SigningConfig(
            mode=SigningMode.DIRECT,
            mnemonic=mnemonic,
        ),
    )


def test_create_sdk():
    sdk = SeocheonSDK(_make_config())
    assert sdk.is_connected is False
    assert sdk.config.chain.chain_id == "seocheon-test-1"


def test_invalid_config_raises():
    bad_config = SDKConfig(
        chain=ChainConfig(chain_id="", rpc_endpoint="", grpc_endpoint=""),
        signing=SigningConfig(mode=SigningMode.DIRECT, mnemonic="test"),
    )
    with pytest.raises(type(ErrInvalidConfig)):
        SeocheonSDK(bad_config)


def test_modules_initialized():
    sdk = SeocheonSDK(_make_config())
    assert hasattr(sdk, "activity")
    assert hasattr(sdk, "epoch")
    assert hasattr(sdk, "node")
    assert hasattr(sdk, "rewards")
    assert hasattr(sdk, "cosmos")


async def test_connect_disconnect():
    sdk = SeocheonSDK(_make_config())
    with patch.object(sdk._client, "connect", new_callable=AsyncMock):
        await sdk.connect()
        assert sdk.is_connected is True

    with patch.object(sdk._client, "disconnect", new_callable=AsyncMock):
        await sdk.disconnect()
        assert sdk.is_connected is False

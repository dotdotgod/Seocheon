"""Tests for TX pipeline."""

from unittest.mock import AsyncMock, MagicMock

import pytest

from seocheon.internal.tx.pipeline import PipelineConfig, execute_tx
from seocheon.internal.tx.messages import MsgSubmitActivity
from seocheon.internal.tx.types import TxRequest, TxResult


@pytest.fixture
def mock_querier():
    querier = MagicMock()
    querier.get_account_info = AsyncMock(return_value=(1, 5))
    querier.broadcast_tx_sync = AsyncMock(return_value=("ABCDEF1234567890", 0, ""))
    querier.get_tx_result = AsyncMock(return_value=TxResult(
        tx_hash="ABCDEF1234567890",
        height=17300,
        code=0,
        gas_used=150000,
        gas_wanted=200000,
    ))
    return querier


@pytest.fixture
def mock_signer():
    signer = MagicMock()
    signer.get_address = MagicMock(return_value="seocheon1testaddr")
    signer.get_pub_key = MagicMock(return_value=b"\x02" + b"\x00" * 32)
    signer.sign = MagicMock(return_value=b"\x00" * 64)
    return signer


@pytest.fixture
def pipeline_config():
    return PipelineConfig(chain_id="seocheon-test-1", confirm_timeout=2.0, poll_interval=0.1)


async def test_execute_tx_success(mock_querier, mock_signer, pipeline_config):
    msg = MsgSubmitActivity("seocheon1testaddr", "a" * 64, "https://example.com")
    req = TxRequest(message=msg)

    result = await execute_tx(mock_querier, mock_signer, pipeline_config, req)

    assert result.tx_hash == "ABCDEF1234567890"
    assert result.height == 17300
    assert result.code == 0
    mock_querier.get_account_info.assert_called_once_with("seocheon1testaddr")
    mock_signer.sign.assert_called_once()
    mock_querier.broadcast_tx_sync.assert_called_once()


async def test_execute_tx_broadcast_failure(mock_querier, mock_signer, pipeline_config):
    mock_querier.broadcast_tx_sync = AsyncMock(return_value=("HASH", 5, "error msg"))

    msg = MsgSubmitActivity("seocheon1testaddr", "a" * 64, "https://example.com")
    req = TxRequest(message=msg)

    result = await execute_tx(mock_querier, mock_signer, pipeline_config, req)
    assert result.code == 5
    assert result.raw_log == "error msg"


async def test_execute_tx_custom_gas():
    querier = MagicMock()
    querier.get_account_info = AsyncMock(return_value=(0, 0))
    querier.broadcast_tx_sync = AsyncMock(return_value=("HASH", 0, ""))
    querier.get_tx_result = AsyncMock(return_value=TxResult(tx_hash="HASH", height=100))

    signer = MagicMock()
    signer.get_address = MagicMock(return_value="addr")
    signer.get_pub_key = MagicMock(return_value=b"\x02" + b"\x00" * 32)
    signer.sign = MagicMock(return_value=b"\x00" * 64)

    cfg = PipelineConfig(chain_id="test", confirm_timeout=2.0, poll_interval=0.1)
    msg = MsgSubmitActivity("addr", "a" * 64, "https://example.com")
    req = TxRequest(message=msg, gas_limit=500000)

    result = await execute_tx(querier, signer, cfg, req)
    assert result.tx_hash == "HASH"


async def test_execute_tx_with_polling(mock_querier, mock_signer, pipeline_config):
    """Successful broadcast then poll for confirmation."""
    msg = MsgSubmitActivity("seocheon1testaddr", "b" * 64, "https://example.com")
    req = TxRequest(message=msg)

    result = await execute_tx(mock_querier, mock_signer, pipeline_config, req)
    assert result.tx_hash == "ABCDEF1234567890"
    assert result.height == 17300
    mock_querier.get_tx_result.assert_called_once_with("ABCDEF1234567890")


async def test_execute_tx_account_error(mock_signer, pipeline_config):
    """Account query failure raises."""
    querier = MagicMock()
    querier.get_account_info = AsyncMock(side_effect=Exception("account not found"))

    msg = MsgSubmitActivity("seocheon1testaddr", "a" * 64, "https://example.com")
    req = TxRequest(message=msg)

    with pytest.raises(Exception, match="account not found"):
        await execute_tx(querier, mock_signer, pipeline_config, req)


async def test_execute_tx_signing_error(mock_querier, pipeline_config):
    """Signing failure raises."""
    signer = MagicMock()
    signer.get_address = MagicMock(return_value="seocheon1testaddr")
    signer.get_pub_key = MagicMock(return_value=b"\x02" + b"\x00" * 32)
    signer.sign = MagicMock(side_effect=RuntimeError("HSM unavailable"))

    msg = MsgSubmitActivity("seocheon1testaddr", "a" * 64, "https://example.com")
    req = TxRequest(message=msg)

    with pytest.raises(RuntimeError, match="HSM unavailable"):
        await execute_tx(mock_querier, signer, pipeline_config, req)


async def test_poll_tx_confirmation_timeout():
    """Poll always returns not-found, expect timeout."""
    from seocheon.internal.tx.confirm import poll_tx_confirmation

    querier = MagicMock()
    querier.get_tx_result = AsyncMock(side_effect=Exception("tx not found"))

    with pytest.raises(TimeoutError, match="timeout"):
        await poll_tx_confirmation(querier, "DEADBEEF", timeout=0.3, poll_interval=0.05)

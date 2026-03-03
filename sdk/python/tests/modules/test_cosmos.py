"""Tests for the Cosmos module."""

from unittest.mock import AsyncMock, patch

import pytest

from seocheon.errors.errors import ErrInvalidAddress, ErrTxNotFound
from seocheon.internal.tx.types import TxResult
from seocheon.modules.cosmos import CosmosModule


@pytest.fixture
def cosmos_module(mock_chain_client, mock_signer):
    return CosmosModule(mock_chain_client, mock_signer, "seocheon-test-1")


async def test_get_balance(cosmos_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(return_value={
        "balance": {"denom": "uppyeo", "amount": "10000000000"},
    })

    result = await cosmos_module.get_balance()
    assert result.address == "seocheon1testaddress"
    assert result.balance == "10000000000"
    assert result.balance_kkot == "1.0000000000"


async def test_send_tokens(cosmos_module):
    with patch("seocheon.modules.cosmos.execute_tx", new_callable=AsyncMock) as mock_tx:
        mock_tx.return_value = TxResult(tx_hash="TXHASH123", height=17300, code=0)

        result = await cosmos_module.send_tokens("seocheon1recipient", "1000000")
        assert result.tx_hash == "TXHASH123"
        assert result.block_height == 17300


async def test_send_tokens_empty_address(cosmos_module):
    with pytest.raises(type(ErrInvalidAddress)):
        await cosmos_module.send_tokens("", "1000")


async def test_get_block_info(cosmos_module, mock_chain_client):
    result = await cosmos_module.get_block_info()
    assert result.block_height == 17281
    assert result.chain_id == "seocheon-testnet-1"
    assert result.num_txs == 5


async def test_get_tx_result_not_found(cosmos_module, mock_chain_client):
    mock_chain_client.get_tx = AsyncMock(side_effect=Exception("not found"))

    with pytest.raises(type(ErrTxNotFound)):
        await cosmos_module.get_tx_result("NONEXISTENT")

"""Tests for the Activity module."""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from seocheon.errors.errors import ErrInvalidActivityHash, ErrInvalidContentURI, ErrSubmitterNotRegistered
from seocheon.internal.tx.types import TxResult
from seocheon.modules.activity import ActivityModule


@pytest.fixture
def activity_module(mock_chain_client, mock_signer):
    return ActivityModule(mock_chain_client, mock_signer, "seocheon-test-1")


async def test_get_activities(activity_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=[
        # _resolve_own_node_id
        {"node": {"id": "node1"}},
        # get_latest_block is already mocked
        # _get_activity_params
        {"params": {"epoch_length": 17280, "windows_per_epoch": 12}},
        # get_activities query
        {"activities": [
            {"activity_hash": "a" * 64, "content_uri": "https://example.com", "block_height": 17300},
        ]},
        # _get_activity_params (called again for window calc)
        {"params": {"epoch_length": 17280, "windows_per_epoch": 12}},
    ])

    result = await activity_module.get_activities()
    assert result.total_count == 1
    assert result.activities[0].activity_hash == "a" * 64


async def test_get_quota(activity_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=[
        # _resolve_own_node_id
        {"node": {"id": "node1"}},
        # _compute_current_epoch -> get_latest_block (already mocked) + _get_activity_params
        {"params": {"epoch_length": 17280, "windows_per_epoch": 12}},
        # get_quota query
        {"quota_used": 3, "quota_limit": 10},
        # _check_feegrant
        {"allowance": {"some": "data"}},
    ])

    result = await activity_module.get_quota()
    assert result.quota_total == 10
    assert result.quota_used == 3
    assert result.quota_remaining == 7
    assert result.is_feegrant is True


async def test_submit_invalid_hash(activity_module):
    with pytest.raises(type(ErrInvalidActivityHash)):
        await activity_module.submit("invalid-hash", "https://example.com")


async def test_submit_empty_uri(activity_module):
    with pytest.raises(type(ErrInvalidContentURI)):
        await activity_module.submit("a" * 64, "")


async def test_resolve_own_node_id_not_registered(activity_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(return_value={"node": {}})

    with pytest.raises(type(ErrSubmitterNotRegistered)):
        await activity_module.get_activities()

"""Tests for the Epoch module."""

from unittest.mock import AsyncMock

import pytest

from seocheon.modules.epoch import EpochModule


@pytest.fixture
def epoch_module(mock_chain_client):
    return EpochModule(mock_chain_client)


async def test_get_info(epoch_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=[
        # get_info epoch-info query
        {
            "current_epoch": 1,
            "current_window": 0,
            "epoch_start_block": 17281,
            "blocks_until_next_epoch": 17279,
        },
        # _get_params
        {"params": {"epoch_length": 17280, "windows_per_epoch": 12, "min_active_windows": 8}},
    ])

    result = await epoch_module.get_info()
    assert result.epoch_number == 1
    assert result.window_number == 0
    assert result.block_height == 17281
    assert result.epoch_start_block == 17281
    assert result.epoch_end_block == 17281 + 17280 - 1


async def test_get_qualification(epoch_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=[
        # get_qualification query
        {"summary": {"active_windows": 5, "eligible": False}},
        # _get_params for qualification
        {"params": {"epoch_length": 17280, "windows_per_epoch": 12, "min_active_windows": 8}},
        # get_info -> epoch-info
        {
            "current_epoch": 1,
            "current_window": 6,
            "epoch_start_block": 17281,
            "blocks_until_next_epoch": 8640,
        },
        # _get_params for get_info
        {"params": {"epoch_length": 17280, "windows_per_epoch": 12, "min_active_windows": 8}},
        # activities query for window detail
        {"activities": []},
    ])

    result = await epoch_module.get_qualification("node1", epoch_number=1)
    assert result.epoch_number == 1
    assert result.active_windows == 5
    assert result.required_windows == 8
    assert result.is_qualified is False
    assert result.remaining_needed == 3
    assert result.total_windows == 12
    assert len(result.window_detail) == 12


async def test_get_qualification_requires_node_id(epoch_module):
    with pytest.raises(ValueError, match="node_id"):
        await epoch_module.get_qualification("")

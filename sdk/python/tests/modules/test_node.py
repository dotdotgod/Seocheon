"""Tests for the Node module."""

from unittest.mock import AsyncMock

import pytest

from seocheon.errors.errors import ErrNodeNotFound
from seocheon.modules.node import NodeModule


@pytest.fixture
def node_module(mock_chain_client, mock_signer):
    return NodeModule(mock_chain_client, mock_signer)


async def test_get_info_by_id(node_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=[
        # get node
        {
            "node": {
                "id": "node1",
                "operator": "seocheon1operator",
                "agent_address": "seocheon1agent",
                "status": 1,
                "description": "test node",
                "website": "",
                "tags": ["ai"],
                "validator_address": "",
                "agent_share": "0.2",
                "registered_at": 100,
            }
        },
    ])

    result = await node_module.get_info("node1")
    assert result.node_id == "node1"
    assert result.operator == "seocheon1operator"
    assert result.status == "REGISTERED"
    assert result.tags == ["ai"]


async def test_get_info_own_node(node_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=[
        # _resolve_own_node_id
        {"node": {"id": "mynode"}},
        # get_info
        {
            "node": {
                "id": "mynode",
                "operator": "seocheon1op",
                "agent_address": "seocheon1testaddress",
                "status": 2,
                "description": "",
                "website": "",
                "tags": [],
                "validator_address": "",
                "agent_share": "0.2",
                "registered_at": 200,
            }
        },
    ])

    result = await node_module.get_info()
    assert result.node_id == "mynode"
    assert result.status == "ACTIVE"


async def test_search_nodes(node_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(return_value={
        "nodes": [
            {"id": "n1", "status": 1, "tags": ["ai"], "description": "Node 1", "validator_address": "", "registered_at": 100},
            {"id": "n2", "status": 2, "tags": ["ml"], "description": "Node 2", "validator_address": "", "registered_at": 200},
        ]
    })

    result = await node_module.search()
    assert result.total_count == 2
    assert len(result.nodes) == 2


async def test_get_info_not_found(node_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=Exception("not found"))

    with pytest.raises(type(ErrNodeNotFound)):
        await node_module.get_info("nonexistent")

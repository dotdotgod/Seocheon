"""Tests for the Rewards module."""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest

from seocheon.errors.errors import ErrNodeNotFound
from seocheon.internal.tx.types import TxEventAttr, TxEventData, TxResult
from seocheon.modules.rewards import RewardsModule, _parse_amount, _parse_agent_share


@pytest.fixture
def rewards_module(mock_chain_client, mock_signer):
    return RewardsModule(mock_chain_client, mock_signer, "seocheon-test-1")


async def test_get_pending(rewards_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(side_effect=[
        # _resolve_own_node_id
        {"node": {"id": "node1"}},
        # _get_node
        {"node": {"id": "node1", "validator_address": "valaddr1", "agent_share": "0.2"}},
        # _query_outstanding_rewards
        {"rewards": {"rewards": [{"denom": "uppyeo", "amount": "1000000"}]}},
        # _query_commission
        {"commission": {"commission": [{"denom": "uppyeo", "amount": "500000"}]}},
    ])

    result = await rewards_module.get_pending()
    assert result.delegation_reward != ""
    assert result.commission_total != ""


async def test_get_pending_not_registered(rewards_module, mock_chain_client):
    mock_chain_client.query_rest = AsyncMock(return_value={"node": {}})

    with pytest.raises(type(ErrNodeNotFound)):
        await rewards_module.get_pending()


async def test_withdraw(rewards_module, mock_chain_client):
    with patch("seocheon.modules.rewards.execute_tx", new_callable=AsyncMock) as mock_tx:
        mock_tx.return_value = TxResult(
            tx_hash="ABCDEF",
            height=17300,
            code=0,
            events=[
                TxEventData(
                    type="withdraw_commission",
                    attributes=[
                        TxEventAttr(key="amount", value="100000uppyeo"),
                        TxEventAttr(key="operator_share", value="80000uppyeo"),
                        TxEventAttr(key="agent_share", value="20000uppyeo"),
                    ],
                )
            ],
        )

        result = await rewards_module.withdraw()
        assert result.tx_hash == "ABCDEF"


def test_parse_amount():
    assert _parse_amount("100000uppyeo") == 100000
    assert _parse_amount("0") == 0
    assert _parse_amount("") == 0
    assert _parse_amount("500") == 500


def test_parse_agent_share():
    assert _parse_agent_share("0.2") == 0.2
    assert _parse_agent_share("20") == 0.2
    assert _parse_agent_share("invalid") == 0.2

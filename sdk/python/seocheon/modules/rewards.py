"""Rewards module for the Seocheon SDK."""

from __future__ import annotations

from seocheon.errors.errors import ErrNodeNotFound, ErrQueryFailed, abci_code_to_error
from seocheon.infrastructure.chain_client import ChainClient
from seocheon.infrastructure.signing.service import SigningService
from seocheon.internal.tx.chain_adapter import ChainClientAdapter
from seocheon.internal.tx.messages import MsgWithdrawNodeCommission
from seocheon.internal.tx.pipeline import PipelineConfig, default_pipeline_config, execute_tx
from seocheon.internal.tx.types import TxRequest
from seocheon.types.responses import PendingRewardsResponse, WithdrawRewardsResponse
from seocheon.utils.convert import format_kkot


class RewardsModule:
    """Provides reward-related operations."""

    def __init__(self, client: ChainClient, signer: SigningService, chain_id: str) -> None:
        self._client = client
        self._signer = signer
        self._tx_querier = ChainClientAdapter(client)
        self._tx_config: PipelineConfig = default_pipeline_config(chain_id)

    async def get_pending(self, node_id: str = "") -> PendingRewardsResponse:
        """Return pending (unwithdrawn) rewards for a node."""
        effective_node_id = node_id or await self._resolve_own_node_id()
        node_info = await self._get_node(effective_node_id)

        delegation_reward = 0
        commission_total = 0
        val_addr = node_info.get("validator_address", "")
        if val_addr:
            delegation_reward = await self._query_outstanding_rewards(val_addr)
            commission_total = await self._query_commission(val_addr)

        activity_reward = 0  # placeholder
        total_reward = delegation_reward + activity_reward

        agent_share_ratio = _parse_agent_share(node_info.get("agent_share", "0.2"))
        operator_share = int(commission_total * (1.0 - agent_share_ratio))
        agent_share = commission_total - operator_share

        return PendingRewardsResponse(
            delegation_reward=format_kkot(delegation_reward),
            activity_reward=format_kkot(activity_reward),
            total_reward=format_kkot(total_reward),
            commission_total=format_kkot(commission_total),
            operator_share=format_kkot(operator_share),
            agent_share=format_kkot(agent_share),
        )

    async def withdraw(self) -> WithdrawRewardsResponse:
        """Withdraw all pending rewards. Requires operator signing key."""
        operator = self._signer.get_address()
        msg = MsgWithdrawNodeCommission(operator=operator)

        result = await execute_tx(
            self._tx_querier, self._signer, self._tx_config, TxRequest(message=msg)
        )

        if result.code != 0:
            raise abci_code_to_error(result.code)

        withdrawn_total = 0
        to_operator = 0
        to_agent = 0

        for evt in result.events:
            if evt.type in ("withdraw_commission", "withdraw_node_commission"):
                for attr in evt.attributes:
                    if attr.key == "amount":
                        withdrawn_total = _parse_amount(attr.value)
                    elif attr.key == "operator_share":
                        to_operator = _parse_amount(attr.value)
                    elif attr.key == "agent_share":
                        to_agent = _parse_amount(attr.value)

        if withdrawn_total > 0 and to_operator == 0 and to_agent == 0:
            to_operator = withdrawn_total * 80 // 100
            to_agent = withdrawn_total - to_operator

        return WithdrawRewardsResponse(
            tx_hash=result.tx_hash,
            withdrawn_total=format_kkot(withdrawn_total),
            to_operator=format_kkot(to_operator),
            to_agent=format_kkot(to_agent),
        )

    async def _resolve_own_node_id(self) -> str:
        agent_addr = self._signer.get_address()
        path = f"/seocheon/node/v1/nodes/by-agent/{agent_addr}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrNodeNotFound from e
        node_id = data.get("node", {}).get("id", "")
        if not node_id:
            raise ErrNodeNotFound
        return node_id

    async def _get_node(self, node_id: str) -> dict:
        path = f"/seocheon/node/v1/nodes/{node_id}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrNodeNotFound from e
        return data.get("node", {})

    async def _query_outstanding_rewards(self, val_addr: str) -> int:
        try:
            data = await self._client.query_rest(
                f"/cosmos/distribution/v1beta1/validators/{val_addr}/outstanding_rewards"
            )
            for r in data.get("rewards", {}).get("rewards", []):
                if r.get("denom") == "uppyeo":
                    return int(float(r.get("amount", "0")))
        except Exception:
            pass
        return 0

    async def _query_commission(self, val_addr: str) -> int:
        try:
            data = await self._client.query_rest(
                f"/cosmos/distribution/v1beta1/validators/{val_addr}/commission"
            )
            for c in data.get("commission", {}).get("commission", []):
                if c.get("denom") == "uppyeo":
                    return int(float(c.get("amount", "0")))
        except Exception:
            pass
        return 0


def _parse_amount(s: str) -> int:
    result = 0
    for c in s:
        if c.isdigit():
            result = result * 10 + int(c)
        else:
            break
    return result


def _parse_agent_share(s: str) -> float:
    try:
        val = float(s)
    except ValueError:
        return 0.2
    if val > 1:
        return val / 100.0
    return val

"""Node module for the Seocheon SDK."""

from __future__ import annotations

from seocheon.errors.errors import ErrNodeNotFound, ErrQueryFailed
from seocheon.infrastructure.chain_client import ChainClient
from seocheon.infrastructure.signing.service import SigningService
from seocheon.types.enums import NodeStatus
from seocheon.types.responses import NodeInfoResponse, NodeSearchResponse, NodeSummary
from seocheon.utils.convert import format_kkot


class NodeModule:
    """Provides node-related operations."""

    def __init__(self, client: ChainClient, signer: SigningService) -> None:
        self._client = client
        self._signer = signer

    async def get_info(self, node_id: str = "") -> NodeInfoResponse:
        """Return detailed information about a node."""
        effective_node_id = node_id or await self._resolve_own_node_id()

        path = f"/seocheon/node/v1/nodes/{effective_node_id}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrNodeNotFound from e

        n = data.get("node", {})
        total_delegation = "0"
        self_delegation = "0"
        commission_rate = "0"

        val_addr = n.get("validator_address", "")
        if val_addr:
            total_delegation, commission_rate = await self._query_validator_info(val_addr)
            self_delegation = await self._query_self_delegation(n.get("operator", ""), val_addr)

        return NodeInfoResponse(
            node_id=n.get("id", ""),
            operator=n.get("operator", ""),
            agent_address=n.get("agent_address", ""),
            status=NodeStatus.from_int(n.get("status", 0)),
            description=n.get("description", ""),
            website=n.get("website", ""),
            tags=n.get("tags", []),
            commission_rate=commission_rate,
            agent_share=n.get("agent_share", "0.2"),
            total_delegation=total_delegation,
            self_delegation=self_delegation,
            validator_address=val_addr,
            registered_at=int(n.get("registered_at", 0)),
        )

    async def search(
        self, tag: str = "", status: str = "", limit: int = 20, order_by: str = "delegation",
    ) -> NodeSearchResponse:
        """Find nodes matching the given criteria."""
        if limit <= 0:
            limit = 20
        if not order_by:
            order_by = "delegation"

        path = f"/seocheon/node/v1/nodes/by-tag/{tag}" if tag else "/seocheon/node/v1/nodes"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrQueryFailed from e

        nodes = data.get("nodes", [])

        # Filter by status
        if status:
            nodes = [n for n in nodes if NodeStatus.from_int(n.get("status", 0)) == status]

        # Enrich with delegation info
        enriched = []
        for n in nodes:
            delegation = 0
            val_addr = n.get("validator_address", "")
            if val_addr:
                del_str, _ = await self._query_validator_info(val_addr)
                try:
                    delegation = int(del_str)
                except ValueError:
                    pass
            reg_at = int(n.get("registered_at", 0))
            enriched.append({
                "summary": NodeSummary(
                    node_id=n.get("id", ""),
                    status=NodeStatus.from_int(n.get("status", 0)),
                    tags=n.get("tags", []),
                    total_delegation=format_kkot(delegation),
                    description=n.get("description", ""),
                ),
                "delegation": delegation,
                "reg_at": reg_at,
            })

        # Sort
        if order_by == "delegation":
            enriched.sort(key=lambda x: x["delegation"], reverse=True)
        else:
            enriched.sort(key=lambda x: x["reg_at"], reverse=True)

        total_count = len(enriched)
        enriched = enriched[:limit]

        return NodeSearchResponse(
            nodes=[e["summary"] for e in enriched],
            total_count=total_count,
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

    async def _query_validator_info(self, val_addr: str) -> tuple[str, str]:
        try:
            data = await self._client.query_rest(f"/cosmos/staking/v1beta1/validators/{val_addr}")
            validator = data.get("validator", {})
            tokens = validator.get("tokens", "0")
            rate = validator.get("commission", {}).get("commission_rates", {}).get("rate", "0")
            return tokens, rate
        except Exception:
            return "0", "0"

    async def _query_self_delegation(self, operator: str, val_addr: str) -> str:
        try:
            data = await self._client.query_rest(
                f"/cosmos/staking/v1beta1/validators/{val_addr}/delegations/{operator}"
            )
            return data.get("delegation_response", {}).get("balance", {}).get("amount", "0")
        except Exception:
            return "0"

"""Activity module for the Seocheon SDK."""

from __future__ import annotations

from seocheon.constants.chain import EPOCH_LENGTH, WINDOWS_PER_EPOCH
from seocheon.errors.errors import (
    ErrInvalidActivityHash,
    ErrInvalidContentURI,
    ErrQueryFailed,
    ErrSubmitterNotRegistered,
    abci_code_to_error,
)
from seocheon.infrastructure.chain_client import ChainClient
from seocheon.infrastructure.signing.service import SigningService
from seocheon.internal.tx.chain_adapter import ChainClientAdapter
from seocheon.internal.tx.messages import MsgSubmitActivity
from seocheon.internal.tx.pipeline import PipelineConfig, default_pipeline_config, execute_tx
from seocheon.internal.tx.types import TxRequest
from seocheon.types.responses import (
    ActivityItem,
    GetActivitiesResponse,
    GetQuotaResponse,
    SubmitActivityResponse,
)
from seocheon.utils.epoch import compute_epoch, compute_window
from seocheon.utils.hash import verify_activity_hash


class ActivityModule:
    """Provides activity-related operations."""

    def __init__(self, client: ChainClient, signer: SigningService, chain_id: str) -> None:
        self._client = client
        self._signer = signer
        self._tx_querier = ChainClientAdapter(client)
        self._tx_config: PipelineConfig = default_pipeline_config(chain_id)

    async def submit(self, activity_hash: str, content_uri: str) -> SubmitActivityResponse:
        """Submit an activity hash to the chain."""
        if not verify_activity_hash(activity_hash):
            raise ErrInvalidActivityHash
        if not content_uri:
            raise ErrInvalidContentURI

        msg = MsgSubmitActivity(
            submitter=self._signer.get_address(),
            activity_hash=activity_hash,
            content_uri=content_uri,
        )

        result = await execute_tx(self._tx_querier, self._signer, self._tx_config, TxRequest(message=msg))

        if result.code != 0:
            raise abci_code_to_error(result.code)

        params = await self._get_activity_params()
        epoch_number = compute_epoch(result.height, params["epoch_length"])
        window_number = compute_window(result.height, params["epoch_length"], params["windows_per_epoch"])

        quota_remaining = 0
        try:
            quota = await self.get_quota()
            quota_remaining = quota.quota_remaining
        except Exception:
            pass

        return SubmitActivityResponse(
            tx_hash=result.tx_hash,
            block_height=result.height,
            epoch_number=epoch_number,
            window_number=window_number,
            quota_remaining=quota_remaining,
        )

    async def get_activities(self, node_id: str = "", epoch_number: int = 0) -> GetActivitiesResponse:
        """Return activity submission history for a node."""
        effective_node_id = node_id or await self._resolve_own_node_id()
        effective_epoch = epoch_number or await self._compute_current_epoch()

        path = f"/seocheon/activity/v1/nodes/{effective_node_id}/activities?epoch={effective_epoch}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrQueryFailed from e

        params = await self._get_activity_params()
        items = []
        for a in data.get("activities", []):
            height = int(a.get("block_height", 0))
            wn = compute_window(height, params["epoch_length"], params["windows_per_epoch"])
            items.append(ActivityItem(
                activity_hash=a.get("activity_hash", ""),
                content_uri=a.get("content_uri", ""),
                block_height=height,
                window_number=wn,
            ))

        return GetActivitiesResponse(activities=items, total_count=len(items))

    async def get_quota(self) -> GetQuotaResponse:
        """Return the remaining activity submission quota for the current epoch."""
        node_id = await self._resolve_own_node_id()
        epoch_num = await self._compute_current_epoch()

        path = f"/seocheon/activity/v1/nodes/{node_id}/epochs/{epoch_num}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrQueryFailed from e

        used = int(data.get("quota_used", 0))
        limit = int(data.get("quota_limit", 0))

        agent_addr = self._signer.get_address()
        is_feegrant = await self._check_feegrant(agent_addr)

        return GetQuotaResponse(
            epoch_number=epoch_num,
            quota_total=limit,
            quota_used=used,
            quota_remaining=limit - used,
            is_feegrant=is_feegrant,
        )

    async def _resolve_own_node_id(self) -> str:
        agent_addr = self._signer.get_address()
        path = f"/seocheon/node/v1/nodes/by-agent/{agent_addr}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrSubmitterNotRegistered from e
        node_id = data.get("node", {}).get("id", "")
        if not node_id:
            raise ErrSubmitterNotRegistered
        return node_id

    async def _get_activity_params(self) -> dict:
        try:
            data = await self._client.query_rest("/seocheon/activity/v1/params")
            params = data.get("params", {})
            el = int(params.get("epoch_length", 0)) or EPOCH_LENGTH
            wpe = int(params.get("windows_per_epoch", 0)) or WINDOWS_PER_EPOCH
            return {"epoch_length": el, "windows_per_epoch": wpe}
        except Exception:
            return {"epoch_length": EPOCH_LENGTH, "windows_per_epoch": WINDOWS_PER_EPOCH}

    async def _compute_current_epoch(self) -> int:
        block = await self._client.get_latest_block()
        params = await self._get_activity_params()
        return compute_epoch(block["height"], params["epoch_length"])

    async def _check_feegrant(self, agent_addr: str) -> bool:
        try:
            path = f"/cosmos/feegrant/v1beta1/allowance/feegrant_pool/{agent_addr}"
            data = await self._client.query_rest(path)
            return data.get("allowance") is not None
        except Exception:
            return False

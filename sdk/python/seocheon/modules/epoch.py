"""Epoch module for the Seocheon SDK."""

from __future__ import annotations

from seocheon.constants.chain import EPOCH_LENGTH, MIN_ACTIVE_WINDOWS, WINDOWS_PER_EPOCH
from seocheon.errors.errors import ErrQueryFailed
from seocheon.infrastructure.chain_client import ChainClient
from seocheon.types.responses import (
    EpochInfoResponse,
    QualificationResponse,
    WindowActivity,
)


class EpochModule:
    """Provides epoch-related operations."""

    def __init__(self, client: ChainClient) -> None:
        self._client = client

    async def get_info(self) -> EpochInfoResponse:
        """Return the current epoch and window state."""
        try:
            data = await self._client.query_rest("/seocheon/activity/v1/epoch-info")
        except Exception as e:
            raise ErrQueryFailed from e

        params = await self._get_params()
        block = await self._client.get_latest_block()

        current_epoch = int(data.get("current_epoch", 0))
        current_window = int(data.get("current_window", 0))
        epoch_start_block = int(data.get("epoch_start_block", 0))
        blocks_until_next_epoch = int(data.get("blocks_until_next_epoch", 0))

        window_length = params["epoch_length"] // params["windows_per_epoch"]
        epoch_end_block = epoch_start_block + params["epoch_length"] - 1
        window_start_block = epoch_start_block + (current_window * window_length)
        window_end_block = window_start_block + window_length - 1

        epoch_elapsed = block["height"] - epoch_start_block + 1
        window_elapsed = block["height"] - window_start_block + 1

        return EpochInfoResponse(
            block_height=block["height"],
            epoch_number=current_epoch,
            epoch_start_block=epoch_start_block,
            epoch_end_block=epoch_end_block,
            epoch_progress=f"{epoch_elapsed}/{params['epoch_length']}",
            window_number=current_window,
            window_start_block=window_start_block,
            window_end_block=window_end_block,
            window_progress=f"{window_elapsed}/{window_length}",
            blocks_until_next_window=window_end_block - block["height"],
            blocks_until_next_epoch=blocks_until_next_epoch,
        )

    async def get_qualification(self, node_id: str, epoch_number: int = 0) -> QualificationResponse:
        """Return the activity reward qualification status for a node."""
        if not node_id:
            raise ValueError("node_id is required for qualification query")

        effective_epoch = epoch_number
        if effective_epoch == 0:
            info = await self.get_info()
            effective_epoch = info.epoch_number

        path = f"/seocheon/activity/v1/nodes/{node_id}/epochs/{effective_epoch}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrQueryFailed from e

        params = await self._get_params()
        summary = data.get("summary", {})
        active_windows = int(summary.get("active_windows", 0))

        epoch_info = await self.get_info()
        if effective_epoch == epoch_info.epoch_number:
            elapsed_windows = epoch_info.window_number + 1
        else:
            elapsed_windows = params["windows_per_epoch"]

        remaining_needed = max(0, params["min_active_windows"] - active_windows)
        remaining_windows = params["windows_per_epoch"] - elapsed_windows
        can_still_qualify = (active_windows + remaining_windows) >= params["min_active_windows"]

        # Build window detail
        window_detail = [
            WindowActivity(window_number=w, submission_count=0, has_activity=False)
            for w in range(params["windows_per_epoch"])
        ]

        # Enrich from activities query
        try:
            act_path = f"/seocheon/activity/v1/nodes/{node_id}/activities?epoch={effective_epoch}"
            act_data = await self._client.query_rest(act_path)
            window_length = params["epoch_length"] // params["windows_per_epoch"]
            for act in act_data.get("activities", []):
                h = int(act.get("block_height", 0))
                wn = ((h - 1) % params["epoch_length"]) // window_length
                if 0 <= wn < params["windows_per_epoch"]:
                    window_detail[wn].submission_count += 1
                    window_detail[wn].has_activity = True
        except Exception:
            pass

        return QualificationResponse(
            epoch_number=effective_epoch,
            total_windows=params["windows_per_epoch"],
            elapsed_windows=elapsed_windows,
            active_windows=active_windows,
            required_windows=params["min_active_windows"],
            is_qualified=summary.get("eligible", False),
            remaining_needed=remaining_needed,
            can_still_qualify=can_still_qualify,
            window_detail=window_detail,
        )

    async def _get_params(self) -> dict:
        try:
            data = await self._client.query_rest("/seocheon/activity/v1/params")
            params = data.get("params", {})
            el = int(params.get("epoch_length", 0)) or EPOCH_LENGTH
            wpe = int(params.get("windows_per_epoch", 0)) or WINDOWS_PER_EPOCH
            maw = int(params.get("min_active_windows", 0)) or MIN_ACTIVE_WINDOWS
            return {"epoch_length": el, "windows_per_epoch": wpe, "min_active_windows": maw}
        except Exception:
            return {"epoch_length": EPOCH_LENGTH, "windows_per_epoch": WINDOWS_PER_EPOCH, "min_active_windows": MIN_ACTIVE_WINDOWS}

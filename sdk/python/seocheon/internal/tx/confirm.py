"""TX confirmation polling."""

from __future__ import annotations

import asyncio
import time

from seocheon.internal.tx.types import ChainQuerier, TxResult

DEFAULT_CONFIRM_TIMEOUT = 30.0  # seconds
DEFAULT_POLL_INTERVAL = 1.0  # seconds


async def poll_tx_confirmation(
    querier: ChainQuerier,
    tx_hash: str,
    timeout: float = DEFAULT_CONFIRM_TIMEOUT,
    poll_interval: float = DEFAULT_POLL_INTERVAL,
) -> TxResult:
    """Poll for a transaction result until confirmed or timeout.

    Raises TimeoutError if the TX is not confirmed within the timeout.
    """
    if timeout <= 0:
        timeout = DEFAULT_CONFIRM_TIMEOUT
    if poll_interval <= 0:
        poll_interval = DEFAULT_POLL_INTERVAL

    deadline = time.monotonic() + timeout

    while True:
        if time.monotonic() >= deadline:
            raise TimeoutError(f"timeout waiting for tx {tx_hash} after {timeout}s")

        try:
            result = await querier.get_tx_result(tx_hash)
            return result
        except Exception:
            # TX not yet indexed, continue polling
            remaining = deadline - time.monotonic()
            if remaining <= 0:
                raise TimeoutError(f"timeout waiting for tx {tx_hash} after {timeout}s")
            await asyncio.sleep(min(poll_interval, remaining))

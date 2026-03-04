"""E2E integration tests for the Seocheon Python SDK.

Skip conditions:
  - SEOCHEON_GRPC not set
  - SEOCHEON_MNEMONIC not set

Run with:
  uv run pytest e2e/ -m e2e -v
"""

from __future__ import annotations

import hashlib
import os
import time

import pytest

from seocheon.client import SeocheonSDK
from seocheon.config import ChainConfig, SDKConfig, SigningConfig, SigningMode

# ---------------------------------------------------------------------------
# Environment
# ---------------------------------------------------------------------------

GRPC = os.environ.get("SEOCHEON_GRPC", "")
MNEMONIC = os.environ.get("SEOCHEON_MNEMONIC", "")
RPC = os.environ.get("SEOCHEON_RPC", "http://localhost:26657")
CHAIN_ID = os.environ.get("SEOCHEON_CHAIN_ID", "seocheon-e2e")

pytestmark = pytest.mark.e2e


def _skip_if_missing() -> None:
    if not GRPC or not MNEMONIC:
        pytest.skip("E2E 스킵: SEOCHEON_GRPC 또는 SEOCHEON_MNEMONIC 미설정")


def _build_config() -> SDKConfig:
    return SDKConfig(
        chain=ChainConfig(
            chain_id=CHAIN_ID,
            rpc_endpoint=RPC,
            grpc_endpoint=GRPC,
            gas_price="250uppyeo",
        ),
        signing=SigningConfig(
            mode=SigningMode.DIRECT,
            mnemonic=MNEMONIC,
        ),
    )


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture(scope="module")
async def sdk() -> SeocheonSDK:  # type: ignore[misc]
    _skip_if_missing()
    s = SeocheonSDK(_build_config())
    await s.connect()
    yield s
    await s.disconnect()


# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------


@pytest.mark.asyncio
async def test_connect(sdk: SeocheonSDK) -> None:
    """connect() 후 is_connected가 True여야 한다."""
    assert sdk.is_connected, "Connect 후 is_connected = False"


@pytest.mark.asyncio
async def test_get_latest_block(sdk: SeocheonSDK) -> None:
    """Cosmos.get_block_info가 양수 block_height를 반환해야 한다."""
    block = await sdk.cosmos.get_block_info()
    assert block.block_height > 0, f"블록 높이가 양수여야 함: {block.block_height}"
    print(f"\n최신 블록: height={block.block_height} chain_id={block.chain_id}")


@pytest.mark.asyncio
async def test_query_node_module(sdk: SeocheonSDK) -> None:
    """Node.search가 x/node 엔드포인트에 응답해야 한다."""
    resp = await sdk.node.search(limit=10)
    assert resp is not None
    print(f"\nx/node 조회 성공: total={resp.total_count}")


@pytest.mark.asyncio
async def test_query_epoch_info(sdk: SeocheonSDK) -> None:
    """Epoch.get_info가 올바른 에포크 정보를 반환해야 한다."""
    info = await sdk.epoch.get_info()
    assert info.block_height > 0, f"에포크 블록 높이가 양수여야 함: {info.block_height}"
    print(
        f"\n에포크 정보: epoch={info.epoch_number} window={info.window_number} height={info.block_height}"
    )


@pytest.mark.asyncio
async def test_query_balance(sdk: SeocheonSDK) -> None:
    """Cosmos.get_balance가 에이전트 잔액을 반환해야 한다."""
    resp = await sdk.cosmos.get_balance()
    assert resp.address, "주소가 비어 있어서는 안 됨"
    print(f"\n잔액: {resp.balance} uppyeo ({resp.balance_kkot} kkot) 주소={resp.address}")


@pytest.mark.asyncio
async def test_submit_activity(sdk: SeocheonSDK) -> None:
    """Activity.submit이 MsgSubmitActivity를 브로드캐스트하고 TxHash를 반환해야 한다.

    사전 조건: 서명자가 등록된 노드의 agent_address여야 한다.
    """
    payload = f"e2e-activity-{time.time_ns()}"
    activity_hash = hashlib.sha256(payload.encode()).hexdigest()
    content_uri = "https://example.com/e2e-activity"

    resp = await sdk.activity.submit(activity_hash, content_uri)
    assert resp.tx_hash, "TxHash가 비어 있어서는 안 됨"
    print(
        f"\n활동 제출 성공: txhash={resp.tx_hash} height={resp.block_height} "
        f"epoch={resp.epoch_number} window={resp.window_number}"
    )

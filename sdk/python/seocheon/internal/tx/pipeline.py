"""4-phase TX pipeline: Assembly -> Signing -> Broadcast -> Confirmation."""

from __future__ import annotations

from dataclasses import dataclass

from seocheon.internal.tx.confirm import poll_tx_confirmation
from seocheon.internal.tx.envelope import (
    encode_auth_info,
    encode_sign_doc,
    encode_tx_body,
    encode_tx_raw,
)
from seocheon.internal.tx.gas import (
    DEFAULT_FEE_DENOM,
    DEFAULT_GAS_PRICE,
    calculate_fee,
    default_gas_for_message,
)
from seocheon.internal.tx.messages import Coin
from seocheon.internal.tx.types import ChainQuerier, Signer, TxRequest, TxResult


@dataclass
class PipelineConfig:
    """Configuration for the TX pipeline."""

    chain_id: str
    gas_price: int = DEFAULT_GAS_PRICE
    confirm_timeout: float = 30.0  # seconds
    poll_interval: float = 1.0  # seconds


def default_pipeline_config(chain_id: str) -> PipelineConfig:
    """Return a PipelineConfig with sensible defaults."""
    return PipelineConfig(chain_id=chain_id)


async def execute_tx(
    querier: ChainQuerier,
    signer: Signer,
    cfg: PipelineConfig,
    req: TxRequest,
) -> TxResult:
    """Execute the full 4-phase TX pipeline.

    Phase 1 - Assembly: query account, build TxBody + AuthInfo + SignDoc
    Phase 2 - Signing: sign the SignDoc
    Phase 3 - Broadcast: encode TxRaw and broadcast
    Phase 4 - Confirmation: poll for TX inclusion
    """
    # Phase 1: Assembly
    address = signer.get_address()
    pub_key = signer.get_pub_key()

    account_number, sequence = await querier.get_account_info(address)

    # Determine gas limit
    gas_limit = req.gas_limit
    if gas_limit == 0:
        gas_limit = default_gas_for_message(req.message.type_url())

    # Determine fee
    fee_amount = req.fee_amount
    if fee_amount == 0:
        gas_price = cfg.gas_price if cfg.gas_price else DEFAULT_GAS_PRICE
        fee_amount = calculate_fee(gas_limit, gas_price)

    fee_denom = req.fee_denom or DEFAULT_FEE_DENOM

    # Encode TxBody
    body_bytes = encode_tx_body([req.message], req.memo, req.timeout_height)

    # Encode AuthInfo
    fee_coins = [Coin(denom=fee_denom, amount=str(fee_amount))]
    auth_info_bytes = encode_auth_info(pub_key, sequence, fee_coins, gas_limit)

    # Encode SignDoc
    sign_doc_bytes = encode_sign_doc(body_bytes, auth_info_bytes, cfg.chain_id, account_number)

    # Phase 2: Signing
    signature = signer.sign(sign_doc_bytes)

    # Phase 3: Broadcast
    tx_raw_bytes = encode_tx_raw(body_bytes, auth_info_bytes, signature)
    tx_hash, code, raw_log = await querier.broadcast_tx_sync(tx_raw_bytes)

    # Check broadcast result code
    if code != 0:
        return TxResult(tx_hash=tx_hash, code=code, raw_log=raw_log)

    # Phase 4: Confirmation
    result = await poll_tx_confirmation(
        querier, tx_hash, cfg.confirm_timeout, cfg.poll_interval
    )
    return result

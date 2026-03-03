"""Cosmos module for the Seocheon SDK."""

from __future__ import annotations

from seocheon.errors.errors import (
    ErrInvalidAddress,
    ErrQueryFailed,
    ErrTxNotFound,
    abci_code_to_error,
)
from seocheon.infrastructure.chain_client import ChainClient
from seocheon.infrastructure.signing.service import SigningService
from seocheon.internal.tx.chain_adapter import ChainClientAdapter
from seocheon.internal.tx.messages import Coin, MsgSend
from seocheon.internal.tx.pipeline import PipelineConfig, default_pipeline_config, execute_tx
from seocheon.internal.tx.types import TxRequest
from seocheon.types.responses import (
    BalanceResponse,
    BlockInfoResponse,
    EventAttribute,
    SendTokensResponse,
    TxEvent,
    TxResultResponse,
)
from seocheon.utils.convert import format_kkot


class CosmosModule:
    """Provides standard Cosmos operations."""

    def __init__(self, client: ChainClient, signer: SigningService, chain_id: str) -> None:
        self._client = client
        self._signer = signer
        self._tx_querier = ChainClientAdapter(client)
        self._tx_config: PipelineConfig = default_pipeline_config(chain_id)

    async def get_balance(self, address: str = "", denom: str = "") -> BalanceResponse:
        """Return the token balance for an address."""
        effective_addr = address or self._signer.get_address()
        effective_denom = denom or "uppyeo"

        path = f"/cosmos/bank/v1beta1/balances/{effective_addr}/by_denom?denom={effective_denom}"
        try:
            data = await self._client.query_rest(path)
        except Exception as e:
            raise ErrQueryFailed from e

        balance = data.get("balance", {})
        amount_str = balance.get("amount", "0")
        balance_uppyeo = _parse_int_safe(amount_str)

        return BalanceResponse(
            address=effective_addr,
            balance=amount_str,
            balance_kkot=format_kkot(balance_uppyeo),
        )

    async def send_tokens(self, to_address: str, amount: str, denom: str = "") -> SendTokensResponse:
        """Send tokens to the specified address."""
        if not to_address:
            raise ErrInvalidAddress
        if not amount:
            raise ValueError("amount is required")

        effective_denom = denom or "uppyeo"

        msg = MsgSend(
            from_address=self._signer.get_address(),
            to_address=to_address,
            amount=[Coin(denom=effective_denom, amount=amount)],
        )

        result = await execute_tx(
            self._tx_querier, self._signer, self._tx_config, TxRequest(message=msg)
        )

        if result.code != 0:
            raise abci_code_to_error(result.code)

        return SendTokensResponse(tx_hash=result.tx_hash, block_height=result.height)

    async def get_block_info(self) -> BlockInfoResponse:
        """Return the latest block information."""
        try:
            block = await self._client.get_latest_block()
        except Exception as e:
            raise ErrQueryFailed from e

        return BlockInfoResponse(
            block_height=block["height"],
            block_time=block["time"],
            chain_id=block["chain_id"],
            num_txs=block["num_txs"],
        )

    async def get_tx_result(self, tx_hash: str) -> TxResultResponse:
        """Return the result of a transaction by its hash."""
        if not tx_hash:
            raise ErrTxNotFound

        try:
            tx = await self._client.get_tx(tx_hash)
        except Exception as e:
            raise ErrTxNotFound from e

        events = [
            TxEvent(
                type=e["type"],
                attributes=[EventAttribute(key=a["key"], value=a["value"]) for a in e.get("attributes", [])],
            )
            for e in tx.get("events", [])
        ]

        return TxResultResponse(
            tx_hash=tx.get("tx_hash", ""),
            height=tx.get("height", 0),
            code=tx.get("code", 0),
            gas_used=tx.get("gas_used", 0),
            gas_wanted=tx.get("gas_wanted", 0),
            raw_log=tx.get("raw_log", ""),
            events=events,
        )


def _parse_int_safe(s: str) -> int:
    result = 0
    for c in s:
        if c.isdigit():
            result = result * 10 + int(c)
        else:
            break
    return result

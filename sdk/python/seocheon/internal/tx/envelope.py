"""TX envelope encoding: TxBody, AuthInfo, SignDoc, TxRaw."""

from __future__ import annotations

from seocheon.internal.tx.messages import Coin, MessageEncoder
from seocheon.internal.tx.protobuf import (
    concat_bytes,
    encode_field_bytes,
    encode_field_string,
    encode_field_varint,
)


def _encode_any(type_url: str, value: bytes) -> bytes:
    """Encode a google.protobuf.Any message."""
    return concat_bytes(
        encode_field_string(1, type_url),
        encode_field_bytes(2, value),
    )


def _encode_pub_key_any(pub_key: bytes) -> bytes:
    """Encode cosmos.crypto.secp256k1.PubKey as Any."""
    inner_msg = encode_field_bytes(1, pub_key)
    return _encode_any("/cosmos.crypto.secp256k1.PubKey", inner_msg)


def _encode_mode_info_direct() -> bytes:
    """Encode ModeInfo for SIGN_MODE_DIRECT."""
    mode_field = encode_field_varint(1, 1)  # SIGN_MODE_DIRECT = 1
    return encode_field_bytes(1, mode_field)


def _encode_signer_info(pub_key: bytes, sequence: int) -> bytes:
    """Encode SignerInfo."""
    pub_key_any = _encode_pub_key_any(pub_key)
    mode_info = _encode_mode_info_direct()
    return concat_bytes(
        encode_field_bytes(1, pub_key_any),
        encode_field_bytes(2, mode_info),
        encode_field_varint(3, sequence),
    )


def _encode_fee(coins: list[Coin], gas_limit: int) -> bytes:
    """Encode Fee message."""
    parts: list[bytes] = []
    for coin in coins:
        coin_bytes = coin.encode()
        parts.append(encode_field_bytes(1, coin_bytes))
    parts.append(encode_field_varint(2, gas_limit))
    return concat_bytes(*parts)


def encode_tx_body(
    messages: list[MessageEncoder],
    memo: str = "",
    timeout_height: int = 0,
) -> bytes:
    """Encode a TxBody.

    Fields: messages(1, repeated Any), memo(2, string), timeout_height(3, uint64)
    """
    parts: list[bytes] = []
    for msg in messages:
        any_bytes = _encode_any(msg.type_url(), msg.encode())
        parts.append(encode_field_bytes(1, any_bytes))
    if memo:
        parts.append(encode_field_string(2, memo))
    if timeout_height > 0:
        parts.append(encode_field_varint(3, timeout_height))
    return concat_bytes(*parts)


def encode_auth_info(
    pub_key: bytes,
    sequence: int,
    fee_coins: list[Coin],
    gas_limit: int,
) -> bytes:
    """Encode an AuthInfo.

    Fields: signer_infos(1, repeated SignerInfo), fee(2, Fee)
    """
    signer_info = _encode_signer_info(pub_key, sequence)
    fee = _encode_fee(fee_coins, gas_limit)
    return concat_bytes(
        encode_field_bytes(1, signer_info),
        encode_field_bytes(2, fee),
    )


def encode_sign_doc(
    body_bytes: bytes,
    auth_info_bytes: bytes,
    chain_id: str,
    account_number: int,
) -> bytes:
    """Encode a SignDoc for SIGN_MODE_DIRECT.

    Fields: body_bytes(1), auth_info_bytes(2), chain_id(3), account_number(4)
    """
    return concat_bytes(
        encode_field_bytes(1, body_bytes),
        encode_field_bytes(2, auth_info_bytes),
        encode_field_string(3, chain_id),
        encode_field_varint(4, account_number),
    )


def encode_tx_raw(
    body_bytes: bytes,
    auth_info_bytes: bytes,
    *signatures: bytes,
) -> bytes:
    """Encode a TxRaw for broadcast.

    Fields: body_bytes(1), auth_info_bytes(2), signatures(3, repeated bytes)
    """
    parts = [
        encode_field_bytes(1, body_bytes),
        encode_field_bytes(2, auth_info_bytes),
    ]
    for sig in signatures:
        parts.append(encode_field_bytes(3, sig))
    return concat_bytes(*parts)

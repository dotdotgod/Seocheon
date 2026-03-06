"""Message encoders for Cosmos TX messages."""

from __future__ import annotations

from typing import Protocol

from seocheon.internal.tx.protobuf import concat_bytes, encode_field_bytes, encode_field_string


class MessageEncoder(Protocol):
    """Protocol for encoding a specific message type into protobuf bytes."""

    def type_url(self) -> str:
        """Return the protobuf type URL."""
        ...

    def encode(self) -> bytes:
        """Return the protobuf-encoded message bytes."""
        ...


class MsgSubmitActivity:
    """Encodes /seocheon.activity.v1.MsgSubmitActivity.

    Fields: submitter(1, string), activity_hash(2, string), content_uri(3, string)
    """

    def __init__(self, submitter: str, activity_hash: str, content_uri: str) -> None:
        self.submitter = submitter
        self.activity_hash = activity_hash
        self.content_uri = content_uri

    def type_url(self) -> str:
        return "/seocheon.activity.v1.MsgSubmitActivity"

    def encode(self) -> bytes:
        return concat_bytes(
            encode_field_string(1, self.submitter),
            encode_field_string(2, self.activity_hash),
            encode_field_string(3, self.content_uri),
        )


class MsgWithdrawNodeCommission:
    """Encodes /seocheon.node.v1.MsgWithdrawNodeCommission.

    Fields: operator(1, string)
    """

    def __init__(self, operator: str) -> None:
        self.operator = operator

    def type_url(self) -> str:
        return "/seocheon.node.v1.MsgWithdrawNodeCommission"

    def encode(self) -> bytes:
        return concat_bytes(
            encode_field_string(1, self.operator),
        )


class MsgConfirmDelegation:
    """Encodes /seocheon.node.v1.MsgConfirmDelegation.

    Fields: delegator_address(1, string), validator_address(2, string)
    """

    def __init__(self, delegator_address: str, validator_address: str) -> None:
        self.delegator_address = delegator_address
        self.validator_address = validator_address

    def type_url(self) -> str:
        return "/seocheon.node.v1.MsgConfirmDelegation"

    def encode(self) -> bytes:
        return concat_bytes(
            encode_field_string(1, self.delegator_address),
            encode_field_string(2, self.validator_address),
        )


class Coin:
    """Represents a cosmos.base.v1beta1.Coin.

    Fields: denom(1, string), amount(2, string)
    """

    def __init__(self, denom: str, amount: str) -> None:
        self.denom = denom
        self.amount = amount

    def encode(self) -> bytes:
        return concat_bytes(
            encode_field_string(1, self.denom),
            encode_field_string(2, self.amount),
        )


class MsgSend:
    """Encodes /cosmos.bank.v1beta1.MsgSend.

    Fields: from_address(1, string), to_address(2, string), amount(3, repeated Coin)
    """

    def __init__(self, from_address: str, to_address: str, amount: list[Coin]) -> None:
        self.from_address = from_address
        self.to_address = to_address
        self.amount = amount

    def type_url(self) -> str:
        return "/cosmos.bank.v1beta1.MsgSend"

    def encode(self) -> bytes:
        parts = [
            encode_field_string(1, self.from_address),
            encode_field_string(2, self.to_address),
        ]
        for coin in self.amount:
            coin_bytes = coin.encode()
            parts.append(encode_field_bytes(3, coin_bytes))
        return concat_bytes(*parts)

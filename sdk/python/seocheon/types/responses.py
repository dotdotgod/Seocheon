"""Response dataclasses for the Seocheon SDK."""

from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class SubmitActivityResponse:
    """Returned after submitting an activity."""

    tx_hash: str
    block_height: int
    window_number: int
    epoch_number: int
    quota_remaining: int


@dataclass
class ActivityItem:
    """Represents a single activity record."""

    activity_hash: str
    content_uri: str
    block_height: int
    window_number: int
    tx_hash: str = ""


@dataclass
class GetActivitiesResponse:
    """Returned when querying activities."""

    activities: list[ActivityItem]
    total_count: int


@dataclass
class GetQuotaResponse:
    """Returned when querying activity quota."""

    epoch_number: int
    quota_total: int
    quota_used: int
    quota_remaining: int
    is_feegrant: bool
    feegrant_expiry: int | None = None


@dataclass
class EpochInfoResponse:
    """Returned when querying epoch information."""

    block_height: int
    epoch_number: int
    epoch_start_block: int
    epoch_end_block: int
    epoch_progress: str
    window_number: int
    window_start_block: int
    window_end_block: int
    window_progress: str
    blocks_until_next_window: int
    blocks_until_next_epoch: int


@dataclass
class WindowActivity:
    """Represents activity within a single window."""

    window_number: int
    submission_count: int
    has_activity: bool


@dataclass
class QualificationResponse:
    """Returned when querying reward qualification."""

    epoch_number: int
    total_windows: int
    elapsed_windows: int
    active_windows: int
    required_windows: int
    is_qualified: bool
    remaining_needed: int
    can_still_qualify: bool
    window_detail: list[WindowActivity] = field(default_factory=list)


@dataclass
class NodeInfoResponse:
    """Returned when querying node information."""

    node_id: str
    operator: str
    agent_address: str
    status: str
    description: str
    website: str
    tags: list[str]
    commission_rate: str
    agent_share: str
    total_delegation: str
    self_delegation: str
    validator_address: str
    registered_at: int


@dataclass
class NodeSummary:
    """A condensed view of a node."""

    node_id: str
    status: str
    tags: list[str]
    total_delegation: str
    description: str


@dataclass
class NodeSearchResponse:
    """Returned from a node search."""

    nodes: list[NodeSummary]
    total_count: int


@dataclass
class PendingRewardsResponse:
    """Returned when querying pending rewards."""

    delegation_reward: str
    activity_reward: str
    total_reward: str
    commission_total: str
    operator_share: str
    agent_share: str


@dataclass
class WithdrawRewardsResponse:
    """Returned after withdrawing rewards."""

    tx_hash: str
    withdrawn_total: str
    to_operator: str
    to_agent: str


@dataclass
class BalanceResponse:
    """Returned when querying balance."""

    address: str
    balance: str
    balance_kkot: str


@dataclass
class SendTokensResponse:
    """Returned after sending tokens."""

    tx_hash: str
    block_height: int


@dataclass
class BlockInfoResponse:
    """Returned when querying block information."""

    block_height: int
    block_time: str
    chain_id: str
    num_txs: int


@dataclass
class EventAttribute:
    """A key-value pair within a TxEvent."""

    key: str
    value: str


@dataclass
class TxEvent:
    """Represents an event emitted by a transaction."""

    type: str
    attributes: list[EventAttribute]


@dataclass
class TxResultResponse:
    """Returned when querying a transaction result."""

    tx_hash: str
    height: int
    code: int
    gas_used: int
    gas_wanted: int
    raw_log: str
    events: list[TxEvent]

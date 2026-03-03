"""Enumerations for the Seocheon SDK."""

from enum import StrEnum


class NodeStatus(StrEnum):
    """Represents the status of a registered node."""

    UNSPECIFIED = "UNSPECIFIED"
    REGISTERED = "REGISTERED"
    ACTIVE = "ACTIVE"
    INACTIVE = "INACTIVE"
    JAILED = "JAILED"

    @classmethod
    def from_int(cls, status: int) -> "NodeStatus":
        """Convert a proto enum integer to NodeStatus."""
        _map = {
            1: cls.REGISTERED,
            2: cls.ACTIVE,
            3: cls.INACTIVE,
            4: cls.JAILED,
        }
        return _map.get(status, cls.UNSPECIFIED)

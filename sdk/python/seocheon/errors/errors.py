"""Error types for the Seocheon SDK."""

from __future__ import annotations


class SDKError(Exception):
    """Represents an error from the Seocheon SDK with an optional code."""

    def __init__(self, code: int, message: str, cause: Exception | None = None) -> None:
        self.code = code
        self.message = message
        self.cause = cause
        super().__init__(str(self))

    def __str__(self) -> str:
        if self.code > 0:
            return f"[{self.code}] {self.message}"
        return self.message


def new(code: int, message: str) -> SDKError:
    """Create a new SDKError with the given code and message."""
    return SDKError(code, message)


def wrap(code: int, message: str, err: Exception) -> SDKError:
    """Create a new SDKError wrapping an existing error."""
    return SDKError(code, message, cause=err)


# SDK-level errors (9000 series)
ErrNotConnected = new(9000, "SDK is not connected to chain")
ErrBroadcastFailed = new(9001, "transaction broadcast failed")
ErrTxTimeout = new(9002, "transaction confirmation timeout")
ErrTxNotFound = new(9003, "transaction not found")
ErrSigningFailed = new(9004, "transaction signing failed")
ErrInvalidConfig = new(9005, "invalid SDK configuration")
ErrQueryFailed = new(9006, "chain query failed")
ErrInvalidAddress = new(9007, "invalid address format")

# x/node errors (1100 series)
ErrNodeNotFound = new(1101, "node not found")
ErrNodeAlreadyExists = new(1102, "node already exists for this operator")
ErrUnauthorizedOperator = new(1108, "unauthorized: signer is not the node operator")
ErrUnauthorizedAgentMsg = new(1109, "agent address is not authorized for this message type")

# x/activity errors (1200 series)
ErrSubmitterNotRegistered = new(1200, "submitter agent address is not registered to any node")
ErrNodeNotEligible = new(1201, "node is not in an eligible status (REGISTERED or ACTIVE)")
ErrDuplicateActivityHash = new(1202, "duplicate (activity_hash, content_uri) pair already exists")
ErrQuotaExceeded = new(1203, "activity quota exceeded for this epoch")
ErrInvalidActivityHash = new(1204, "activity hash must be exactly 64 hex characters (32 bytes)")
ErrInvalidContentURI = new(1205, "content URI must not be empty")

# ABCI code to error mapping
_ABCI_ERROR_MAP: dict[int, SDKError] = {
    1101: ErrNodeNotFound,
    1102: ErrNodeAlreadyExists,
    1108: ErrUnauthorizedOperator,
    1109: ErrUnauthorizedAgentMsg,
    1200: ErrSubmitterNotRegistered,
    1201: ErrNodeNotEligible,
    1202: ErrDuplicateActivityHash,
    1203: ErrQuotaExceeded,
    1204: ErrInvalidActivityHash,
    1205: ErrInvalidContentURI,
}


def abci_code_to_error(code: int) -> SDKError:
    """Map an ABCI error code to its corresponding SDKError."""
    if code in _ABCI_ERROR_MAP:
        return _ABCI_ERROR_MAP[code]
    return new(code, f"chain error code {code}")

"""Activity hash utilities for the Seocheon SDK."""

import hashlib


def verify_activity_hash(hash_str: str) -> bool:
    """Validate that the given string is a valid activity hash.

    A valid activity hash is exactly 64 hex characters (representing 32 bytes / SHA-256).
    """
    if len(hash_str) != 64:
        return False
    try:
        bytes.fromhex(hash_str)
        return True
    except ValueError:
        return False


def compute_activity_hash(data: bytes) -> str:
    """Compute a SHA-256 hash of the given data.

    Returns a 64-character hex string suitable for use as an activity_hash.
    """
    return hashlib.sha256(data).hexdigest()

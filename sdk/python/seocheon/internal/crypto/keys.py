"""secp256k1 key management for the Seocheon SDK."""

from __future__ import annotations

import hashlib

from coincurve import PrivateKey as _CPrivateKey
from coincurve import PublicKey as _CPublicKey


class PrivateKey:
    """Wraps a secp256k1 private key using coincurve."""

    def __init__(self, key: _CPrivateKey) -> None:
        self._key = key

    @classmethod
    def from_bytes(cls, priv_key_bytes: bytes) -> PrivateKey:
        """Create a PrivateKey from raw 32-byte private key."""
        if len(priv_key_bytes) != 32:
            raise ValueError(f"private key must be 32 bytes, got {len(priv_key_bytes)}")
        return cls(_CPrivateKey(priv_key_bytes))

    def pub_key(self) -> bytes:
        """Return the compressed 33-byte public key."""
        return self._key.public_key.format(compressed=True)

    def raw_bytes(self) -> bytes:
        """Return the raw 32-byte private key."""
        return self._key.secret

    def sign(self, data: bytes) -> bytes:
        """Sign the given data using secp256k1.

        Hashes the data with SHA-256 first and returns a 64-byte compact signature (R || S).
        """
        hash_bytes = hashlib.sha256(data).digest()
        # sign_recoverable returns 65 bytes: R (32) + S (32) + recovery_id (1)
        raw_sig = self._key.sign_recoverable(hash_bytes, hasher=None)
        # Return first 64 bytes (R || S), strip recovery byte
        return raw_sig[:64]


def verify(pub_key_bytes: bytes, data: bytes, sig_bytes: bytes) -> bool:
    """Verify a 64-byte compact signature against data using the public key."""
    if len(sig_bytes) != 64:
        return False

    try:
        pub_key = _CPublicKey(pub_key_bytes)
    except Exception:
        return False

    hash_bytes = hashlib.sha256(data).digest()

    # Convert compact R||S to DER format for verification
    r = sig_bytes[:32]
    s = sig_bytes[32:]
    der_sig = _marshal_der(r, s)

    try:
        return pub_key.verify(der_sig, hash_bytes, hasher=None)
    except Exception:
        return False


def _der_integer(b: bytes) -> bytes:
    """Encode a big-endian unsigned integer as a DER INTEGER."""
    # Strip leading zero bytes
    data = b.lstrip(b"\x00") or b"\x00"
    # If the high bit is set, prepend a zero byte
    if data[0] & 0x80:
        data = b"\x00" + data
    return b"\x02" + bytes([len(data)]) + data


def _marshal_der(r: bytes, s: bytes) -> bytes:
    """Encode R and S (each 32 bytes, big-endian) into DER format."""
    r_enc = _der_integer(r)
    s_enc = _der_integer(s)
    seq = r_enc + s_enc
    return b"\x30" + bytes([len(seq)]) + seq

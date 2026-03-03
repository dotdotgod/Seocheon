"""Bech32 address encoding for the Seocheon SDK."""

from __future__ import annotations

import hashlib

BECH32_PREFIX = "seocheon"
BECH32_CHARSET = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"


def address_from_pubkey(pub_key: bytes) -> str:
    """Derive a Cosmos bech32 address from a compressed public key."""
    if len(pub_key) != 33:
        raise ValueError(f"public key must be 33 bytes (compressed), got {len(pub_key)}")

    # SHA256 hash
    sha = hashlib.sha256(pub_key).digest()

    # RIPEMD160 hash
    rip = hashlib.new("ripemd160", sha).digest()  # 20 bytes

    # Bech32 encode with "seocheon" HRP
    return bech32_encode(BECH32_PREFIX, rip)


def bech32_encode(hrp: str, data: bytes) -> str:
    """Encode data to bech32 format."""
    converted = _convert_bits(data, 8, 5, pad=True)

    # Calculate checksum
    values = converted + [0, 0, 0, 0, 0, 0]
    polymod = _bech32_polymod(_expand_hrp(hrp) + values) ^ 1
    for i in range(6):
        values[len(converted) + i] = (polymod >> (5 * (5 - i))) & 31

    return hrp + "1" + "".join(BECH32_CHARSET[v] for v in values)


def _convert_bits(data: bytes, from_bits: int, to_bits: int, pad: bool) -> list[int]:
    """Convert a byte slice from one bit-group size to another."""
    acc = 0
    bits = 0
    result: list[int] = []
    max_v = (1 << to_bits) - 1

    for b in data:
        acc = (acc << from_bits) | b
        bits += from_bits
        while bits >= to_bits:
            bits -= to_bits
            result.append((acc >> bits) & max_v)

    if pad:
        if bits > 0:
            result.append((acc << (to_bits - bits)) & max_v)
    elif bits >= from_bits:
        raise ValueError("illegal zero padding")
    elif ((acc << (to_bits - bits)) & max_v) != 0:
        raise ValueError("non-zero padding")

    return result


def _expand_hrp(hrp: str) -> list[int]:
    """Expand the human-readable part for checksum computation."""
    result = [ord(c) >> 5 for c in hrp]
    result.append(0)
    result.extend(ord(c) & 31 for c in hrp)
    return result


def _bech32_polymod(values: list[int]) -> int:
    """Compute the bech32 checksum."""
    gen = [0x3B6A57B2, 0x26508E6D, 0x1EA119FA, 0x3D4233DD, 0x2A1462B3]
    chk = 1
    for v in values:
        top = chk >> 25
        chk = ((chk & 0x1FFFFFF) << 5) ^ v
        for i in range(5):
            if (top >> i) & 1:
                chk ^= gen[i]
    return chk

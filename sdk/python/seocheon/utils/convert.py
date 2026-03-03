"""Token denomination conversion utilities for the Seocheon SDK."""

from seocheon.constants.chain import DENOM_FACTORS, UPPYEO_PER_KKOT


def convert_denom(amount: int, from_denom: str, to_denom: str) -> int:
    """Convert an amount between Seocheon token denominations.

    Supported denominations: "uppyeo" (base), "sal", "pi", "sum", "hon", "kkot" (display).
    """
    if from_denom not in DENOM_FACTORS:
        raise ValueError(
            f"unknown denomination: {from_denom} "
            f"(supported: {', '.join(DENOM_FACTORS)})"
        )
    if to_denom not in DENOM_FACTORS:
        raise ValueError(
            f"unknown denomination: {to_denom} "
            f"(supported: {', '.join(DENOM_FACTORS)})"
        )

    if from_denom == to_denom:
        return amount

    from_factor = DENOM_FACTORS[from_denom]
    to_factor = DENOM_FACTORS[to_denom]

    # Convert to base (uppyeo) first, then to target
    base_amount = amount * from_factor
    return base_amount // to_factor


def format_kkot(uppyeo_amount: int) -> str:
    """Convert an uppyeo amount to a human-readable KKOT string.

    Example: 10000000000 uppyeo -> "1.0000000000"
    """
    int_part = uppyeo_amount // UPPYEO_PER_KKOT
    dec_part = uppyeo_amount % UPPYEO_PER_KKOT
    if dec_part < 0:
        dec_part = -dec_part
    return f"{int_part}.{dec_part:010d}"


def parse_kkot(kkot: str) -> int:
    """Parse a KKOT string to uppyeo amount.

    Example: "1.0000000000" -> 10000000000
    """
    parts = kkot.split(".")
    if len(parts) > 2:
        raise ValueError(f"invalid kkot format: {kkot}")

    int_part = 0
    for c in parts[0]:
        if not c.isdigit():
            raise ValueError(f"invalid character in kkot integer part: {c}")
        int_part = int_part * 10 + int(c)

    dec_part = 0
    if len(parts) == 2:
        dec = parts[1]
        # Pad or truncate to 10 decimal places
        dec = dec.ljust(10, "0")[:10]
        for c in dec:
            if not c.isdigit():
                raise ValueError(f"invalid character in kkot decimal part: {c}")
            dec_part = dec_part * 10 + int(c)

    return int_part * UPPYEO_PER_KKOT + dec_part

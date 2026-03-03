"""Tests for denomination conversion utilities."""

import pytest

from seocheon.utils.convert import convert_denom, format_kkot, parse_kkot


def test_format_kkot_various():
    assert format_kkot(0) == "0.0000000000"
    assert format_kkot(10_000_000_000) == "1.0000000000"
    assert format_kkot(15_000_000_000) == "1.5000000000"
    assert format_kkot(100_000_000_000) == "10.0000000000"
    assert format_kkot(1) == "0.0000000001"


def test_parse_kkot_int_and_frac():
    assert parse_kkot("1.0000000000") == 10_000_000_000
    assert parse_kkot("1.5000000000") == 15_000_000_000
    assert parse_kkot("5") == 50_000_000_000
    assert parse_kkot("1.5") == 15_000_000_000
    # roundtrip
    original = 12_345_678_901
    assert parse_kkot(format_kkot(original)) == original


def test_parse_kkot_invalid_format():
    with pytest.raises(ValueError):
        parse_kkot("1.2.3")


def test_convert_denom_same():
    assert convert_denom(100, "uppyeo", "uppyeo") == 100


def test_convert_denom_uppyeo_to_kkot():
    assert convert_denom(10_000_000_000, "uppyeo", "kkot") == 1


def test_convert_denom_kkot_to_uppyeo():
    assert convert_denom(1, "kkot", "uppyeo") == 10_000_000_000


def test_convert_denom_sal_to_pi():
    assert convert_denom(100, "sal", "pi") == 1


def test_convert_large_genesis_amount():
    # 50,000 KKOT = 5×10^14 uppyeo
    assert convert_denom(50_000, "kkot", "uppyeo") == 500_000_000_000_000


def test_convert_denom_unknown():
    with pytest.raises(ValueError, match="unknown denomination"):
        convert_denom(1, "invalid", "kkot")

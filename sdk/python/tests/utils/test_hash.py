"""Tests for hash utilities."""

from seocheon.utils.hash import compute_activity_hash, verify_activity_hash


def test_verify_valid_hash():
    h = "a" * 64
    assert verify_activity_hash(h) is True


def test_verify_valid_hash_mixed():
    h = "abcdef0123456789" * 4
    assert verify_activity_hash(h) is True


def test_verify_short_hash():
    assert verify_activity_hash("abc") is False


def test_verify_long_hash():
    assert verify_activity_hash("a" * 65) is False


def test_verify_invalid_hex():
    assert verify_activity_hash("g" * 64) is False


def test_verify_empty():
    assert verify_activity_hash("") is False


def test_compute_activity_hash():
    h = compute_activity_hash(b"hello world")
    assert len(h) == 64
    assert verify_activity_hash(h)
    # Known SHA-256 of "hello world"
    assert h == "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"


def test_verify_valid_uppercase_hash():
    h = "A" * 64
    assert verify_activity_hash(h) is True


def test_compute_empty():
    h = compute_activity_hash(b"")
    assert len(h) == 64
    assert verify_activity_hash(h)

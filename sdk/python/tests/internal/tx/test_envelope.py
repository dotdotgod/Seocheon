"""Tests for TX envelope encoding."""

from seocheon.internal.tx.envelope import (
    encode_auth_info,
    encode_sign_doc,
    encode_tx_body,
    encode_tx_raw,
)
from seocheon.internal.tx.messages import Coin, MsgSubmitActivity


def test_encode_tx_body_single_message():
    msg = MsgSubmitActivity("addr", "a" * 64, "https://example.com")
    body = encode_tx_body([msg])
    assert isinstance(body, bytes)
    assert len(body) > 0


def test_encode_tx_body_with_memo():
    msg = MsgSubmitActivity("addr", "a" * 64, "https://example.com")
    body = encode_tx_body([msg], memo="test memo")
    assert b"test memo" in body


def test_encode_auth_info():
    pub_key = b"\x02" + b"\x00" * 32
    fee_coins = [Coin("uppyeo", "50000000")]
    auth_info = encode_auth_info(pub_key, 5, fee_coins, 200000)
    assert isinstance(auth_info, bytes)
    assert len(auth_info) > 0


def test_encode_sign_doc():
    body_bytes = b"\x01\x02\x03"
    auth_info_bytes = b"\x04\x05\x06"
    sign_doc = encode_sign_doc(body_bytes, auth_info_bytes, "seocheon-1", 1)
    assert isinstance(sign_doc, bytes)
    assert b"seocheon-1" in sign_doc


def test_encode_tx_raw():
    body_bytes = b"\x01\x02\x03"
    auth_info_bytes = b"\x04\x05\x06"
    signature = b"\x07" * 64
    tx_raw = encode_tx_raw(body_bytes, auth_info_bytes, signature)
    assert isinstance(tx_raw, bytes)
    assert len(tx_raw) > 0


def test_encode_tx_raw_multiple_signatures():
    body_bytes = b"\x01\x02"
    auth_info_bytes = b"\x03\x04"
    sig1 = b"\x05" * 64
    sig2 = b"\x06" * 64
    tx_raw = encode_tx_raw(body_bytes, auth_info_bytes, sig1, sig2)
    assert sig1 in tx_raw
    assert sig2 in tx_raw


def test_full_tx_assembly_pipeline():
    """End-to-end: create msg -> build body -> build auth_info -> sign_doc -> raw tx."""
    msg = MsgSubmitActivity("seocheon1addr", "b" * 64, "https://example.com/report")
    # Phase 1: encode body
    body = encode_tx_body([msg])
    assert isinstance(body, bytes) and len(body) > 0
    # Phase 2: encode auth_info
    pub_key = b"\x02" + b"\x01" * 32
    fee_coins = [Coin("uppyeo", "50000000")]
    auth_info = encode_auth_info(pub_key, 0, fee_coins, 200000)
    assert isinstance(auth_info, bytes) and len(auth_info) > 0
    # Phase 3: encode sign_doc
    sign_doc = encode_sign_doc(body, auth_info, "seocheon-test-1", 1)
    assert isinstance(sign_doc, bytes) and len(sign_doc) > 0
    # Phase 4: encode raw tx with a dummy signature
    signature = b"\xaa" * 64
    tx_raw = encode_tx_raw(body, auth_info, signature)
    assert isinstance(tx_raw, bytes) and len(tx_raw) > 0
    assert body[0:4] in tx_raw or len(tx_raw) > len(body)

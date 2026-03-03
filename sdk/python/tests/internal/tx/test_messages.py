"""Tests for message encoders."""

from seocheon.internal.tx.messages import Coin, MsgSend, MsgSubmitActivity, MsgWithdrawNodeCommission


def test_msg_submit_activity_type_url():
    msg = MsgSubmitActivity("addr", "hash", "uri")
    assert msg.type_url() == "/seocheon.activity.v1.MsgSubmitActivity"


def test_msg_submit_activity_encode():
    msg = MsgSubmitActivity("addr", "hash", "uri")
    encoded = msg.encode()
    assert isinstance(encoded, bytes)
    assert len(encoded) > 0
    # Should contain "addr", "hash", "uri" as encoded strings
    assert b"addr" in encoded
    assert b"hash" in encoded
    assert b"uri" in encoded


def test_msg_withdraw_node_commission_type_url():
    msg = MsgWithdrawNodeCommission("operator")
    assert msg.type_url() == "/seocheon.node.v1.MsgWithdrawNodeCommission"


def test_msg_withdraw_node_commission_encode():
    msg = MsgWithdrawNodeCommission("operator")
    encoded = msg.encode()
    assert b"operator" in encoded


def test_coin_encode():
    coin = Coin("uppyeo", "1000")
    encoded = coin.encode()
    assert b"uppyeo" in encoded
    assert b"1000" in encoded


def test_msg_send_type_url():
    msg = MsgSend("from", "to", [Coin("uppyeo", "1000")])
    assert msg.type_url() == "/cosmos.bank.v1beta1.MsgSend"


def test_msg_send_encode():
    msg = MsgSend("from_addr", "to_addr", [Coin("uppyeo", "1000")])
    encoded = msg.encode()
    assert b"from_addr" in encoded
    assert b"to_addr" in encoded
    assert b"uppyeo" in encoded
    assert b"1000" in encoded


def test_msg_send_multiple_coins():
    msg = MsgSend("from", "to", [Coin("uppyeo", "100"), Coin("sal", "200")])
    encoded = msg.encode()
    assert b"uppyeo" in encoded
    assert b"sal" in encoded

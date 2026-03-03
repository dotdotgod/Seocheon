"""Tests for gas constants and calculation."""

from seocheon.internal.tx.gas import (
    DEFAULT_GAS_FALLBACK,
    DEFAULT_GAS_SEND,
    DEFAULT_GAS_SUBMIT_ACTIVITY,
    DEFAULT_GAS_WITHDRAW_NODE_COMMISSION,
    calculate_fee,
    default_gas_for_message,
)


def test_default_gas_submit_activity():
    gas = default_gas_for_message("/seocheon.activity.v1.MsgSubmitActivity")
    assert gas == DEFAULT_GAS_SUBMIT_ACTIVITY
    assert gas == 200000


def test_default_gas_withdraw():
    gas = default_gas_for_message("/seocheon.node.v1.MsgWithdrawNodeCommission")
    assert gas == DEFAULT_GAS_WITHDRAW_NODE_COMMISSION
    assert gas == 300000


def test_default_gas_send():
    gas = default_gas_for_message("/cosmos.bank.v1beta1.MsgSend")
    assert gas == DEFAULT_GAS_SEND
    assert gas == 100000


def test_default_gas_unknown():
    gas = default_gas_for_message("/unknown.MsgType")
    assert gas == DEFAULT_GAS_FALLBACK
    assert gas == 200000


def test_calculate_fee():
    assert calculate_fee(200000, 250) == 50000000


def test_calculate_fee_zero():
    assert calculate_fee(0, 250) == 0

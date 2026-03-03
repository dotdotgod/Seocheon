"""Tests for SDK errors."""

from seocheon.errors.errors import (
    ErrDuplicateActivityHash,
    ErrNodeNotFound,
    ErrNotConnected,
    SDKError,
    abci_code_to_error,
    new,
    wrap,
)


def test_sdk_error_with_code():
    err = new(9000, "test error")
    assert str(err) == "[9000] test error"
    assert err.code == 9000
    assert err.message == "test error"


def test_sdk_error_without_code():
    err = SDKError(0, "no code")
    assert str(err) == "no code"


def test_wrap_error():
    cause = ValueError("original")
    err = wrap(9001, "wrapped", cause)
    assert err.code == 9001
    assert err.cause is cause


def test_abci_code_to_error_known():
    err = abci_code_to_error(1101)
    assert err is ErrNodeNotFound

    err = abci_code_to_error(1202)
    assert err is ErrDuplicateActivityHash


def test_abci_code_to_error_unknown():
    err = abci_code_to_error(9999)
    assert err.code == 9999
    assert "chain error code 9999" in err.message


def test_predefined_errors():
    assert ErrNotConnected.code == 9000
    assert ErrNodeNotFound.code == 1101

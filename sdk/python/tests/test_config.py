"""Tests for SDK configuration."""

import pytest

from seocheon.config import ChainConfig, SDKConfig, SigningConfig, SigningMode, TxConfig


def test_valid_direct_config():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="seocheon-1", rpc_endpoint="http://localhost:26657", grpc_endpoint="http://localhost:1317"),
        signing=SigningConfig(mode=SigningMode.DIRECT, mnemonic="test mnemonic words"),
    )
    cfg.validate()


def test_missing_chain_id():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="", rpc_endpoint="http://localhost:26657", grpc_endpoint="http://localhost:1317"),
        signing=SigningConfig(mode=SigningMode.DIRECT, mnemonic="test"),
    )
    with pytest.raises(ValueError, match="chain_id is required"):
        cfg.validate()


def test_missing_rpc_endpoint():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="seocheon-1", rpc_endpoint="", grpc_endpoint="http://localhost:1317"),
        signing=SigningConfig(mode=SigningMode.DIRECT, mnemonic="test"),
    )
    with pytest.raises(ValueError, match="rpc_endpoint is required"):
        cfg.validate()


def test_missing_grpc_endpoint():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="seocheon-1", rpc_endpoint="http://localhost:26657", grpc_endpoint=""),
        signing=SigningConfig(mode=SigningMode.DIRECT, mnemonic="test"),
    )
    with pytest.raises(ValueError, match="grpc_endpoint is required"):
        cfg.validate()


def test_direct_mode_requires_mnemonic():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="seocheon-1", rpc_endpoint="http://localhost:26657", grpc_endpoint="http://localhost:1317"),
        signing=SigningConfig(mode=SigningMode.DIRECT, mnemonic=""),
    )
    with pytest.raises(ValueError, match="direct mode requires mnemonic"):
        cfg.validate()


def test_vault_requires_endpoint_and_key():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="seocheon-1", rpc_endpoint="http://localhost:26657", grpc_endpoint="http://localhost:1317"),
        signing=SigningConfig(mode=SigningMode.VAULT, vault_endpoint="", key_name="test"),
    )
    with pytest.raises(ValueError, match="vault mode requires"):
        cfg.validate()


def test_keystore_mode_requires_path():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="seocheon-1", rpc_endpoint="http://localhost:26657", grpc_endpoint="http://localhost:1317"),
        signing=SigningConfig(mode=SigningMode.KEYSTORE, keystore_path="", passphrase_env=""),
    )
    with pytest.raises(ValueError, match="keystore mode requires"):
        cfg.validate()


def test_invalid_signing_mode():
    cfg = SDKConfig(
        chain=ChainConfig(chain_id="seocheon-1", rpc_endpoint="http://localhost:26657", grpc_endpoint="http://localhost:1317"),
        signing=SigningConfig(mode="invalid_mode", mnemonic="test"),
    )
    with pytest.raises(ValueError, match="invalid signing mode"):
        cfg.validate()


def test_default_tx_config():
    tx = TxConfig()
    assert tx.broadcast_mode == "sync"
    assert tx.confirm_timeout_ms == 30000
    assert tx.confirm_poll_interval_ms == 1000

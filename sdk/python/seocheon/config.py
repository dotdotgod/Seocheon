"""Configuration dataclasses for the Seocheon SDK."""

from __future__ import annotations

from dataclasses import dataclass, field
from enum import StrEnum

from seocheon.constants.chain import (
    DEFAULT_BROADCAST_MODE,
    DEFAULT_CONFIRM_POLL_MS,
    DEFAULT_CONFIRM_TIMEOUT_MS,
    DEFAULT_GAS_ADJUSTMENT,
    DEFAULT_GAS_PRICE,
)


class SigningMode(StrEnum):
    """Specifies how transactions are signed."""

    VAULT = "vault"
    KEYSTORE = "keystore"
    DIRECT = "direct"


@dataclass
class ChainConfig:
    """Blockchain connection settings."""

    chain_id: str
    rpc_endpoint: str
    grpc_endpoint: str
    gas_price: str = DEFAULT_GAS_PRICE
    gas_adjustment: float = DEFAULT_GAS_ADJUSTMENT


@dataclass
class SigningConfig:
    """Transaction signing settings."""

    mode: SigningMode
    vault_endpoint: str = ""
    key_name: str = ""
    keystore_path: str = ""
    passphrase_env: str = ""
    mnemonic: str = ""


@dataclass
class TxConfig:
    """Transaction broadcast settings."""

    broadcast_mode: str = DEFAULT_BROADCAST_MODE
    confirm_timeout_ms: int = DEFAULT_CONFIRM_TIMEOUT_MS
    confirm_poll_interval_ms: int = DEFAULT_CONFIRM_POLL_MS


@dataclass
class SDKConfig:
    """Top-level configuration for SeocheonSDK."""

    chain: ChainConfig
    signing: SigningConfig
    tx: TxConfig = field(default_factory=TxConfig)

    def validate(self) -> None:
        """Check that the SDKConfig is valid. Raises ValueError on invalid config."""
        if not self.chain.chain_id:
            raise ValueError("chain_id is required")
        if not self.chain.rpc_endpoint:
            raise ValueError("rpc_endpoint is required")
        if not self.chain.grpc_endpoint:
            raise ValueError("grpc_endpoint is required")

        match self.signing.mode:
            case SigningMode.VAULT:
                if not self.signing.vault_endpoint or not self.signing.key_name:
                    raise ValueError("vault mode requires vault_endpoint and key_name")
            case SigningMode.KEYSTORE:
                if not self.signing.keystore_path or not self.signing.passphrase_env:
                    raise ValueError("keystore mode requires keystore_path and passphrase_env")
            case SigningMode.DIRECT:
                if not self.signing.mnemonic:
                    raise ValueError("direct mode requires mnemonic")
            case _:
                raise ValueError(f"invalid signing mode: {self.signing.mode}")

"""SeocheonSDK main entry point."""

from __future__ import annotations

from seocheon.config import SDKConfig, SigningMode
from seocheon.errors.errors import ErrInvalidConfig
from seocheon.infrastructure.chain_client import HTTPChainClient
from seocheon.infrastructure.signing.direct import DirectSigningService
from seocheon.infrastructure.signing.keystore import KeystoreSigningService
from seocheon.infrastructure.signing.vault import VaultSigningService
from seocheon.modules.activity import ActivityModule
from seocheon.modules.cosmos import CosmosModule
from seocheon.modules.epoch import EpochModule
from seocheon.modules.node import NodeModule
from seocheon.modules.rewards import RewardsModule


class SeocheonSDK:
    """Main entry point for the Seocheon blockchain SDK."""

    def __init__(self, config: SDKConfig) -> None:
        try:
            config.validate()
        except ValueError as e:
            raise ErrInvalidConfig from e

        self._config = config
        self._client = HTTPChainClient(config.chain.rpc_endpoint, config.chain.grpc_endpoint)
        self._signer = _create_signer(config)
        self._connected = False

        chain_id = config.chain.chain_id

        self.activity = ActivityModule(self._client, self._signer, chain_id)
        self.epoch = EpochModule(self._client)
        self.node = NodeModule(self._client, self._signer)
        self.rewards = RewardsModule(self._client, self._signer, chain_id)
        self.cosmos = CosmosModule(self._client, self._signer, chain_id)

    async def connect(self) -> None:
        """Establish a connection to the blockchain node."""
        await self._client.connect()
        self._connected = True

    async def disconnect(self) -> None:
        """Close the connection to the blockchain node."""
        self._connected = False
        await self._client.disconnect()

    @property
    def is_connected(self) -> bool:
        return self._connected

    @property
    def config(self) -> SDKConfig:
        return self._config


def _create_signer(config: SDKConfig):
    """Create the appropriate signing service based on config."""
    match config.signing.mode:
        case SigningMode.VAULT:
            return VaultSigningService(config.signing.vault_endpoint, config.signing.key_name)
        case SigningMode.KEYSTORE:
            return KeystoreSigningService(config.signing.keystore_path, config.signing.passphrase_env)
        case SigningMode.DIRECT:
            return DirectSigningService(config.signing.mnemonic)
        case _:
            raise ValueError(f"unsupported signing mode: {config.signing.mode}")

"""Tests for signing services."""

import json
import os
import tempfile
from unittest.mock import MagicMock

import pytest

from seocheon.infrastructure.signing.direct import DirectSigningService
from seocheon.infrastructure.signing.keystore import KeystoreSigningService


TEST_MNEMONIC = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"


class TestDirectSigningService:
    def test_create_from_mnemonic(self):
        svc = DirectSigningService(TEST_MNEMONIC)
        assert svc.get_address().startswith("seocheon1")
        assert len(svc.get_pub_key()) == 33
        assert svc.get_pub_key()[0] in (0x02, 0x03)

    def test_sign_returns_64_bytes(self):
        svc = DirectSigningService(TEST_MNEMONIC)
        sig = svc.sign(b"test data")
        assert isinstance(sig, bytes)
        assert len(sig) == 64

    def test_empty_mnemonic_raises(self):
        with pytest.raises(ValueError, match="mnemonic"):
            DirectSigningService("")

    def test_deterministic_address(self):
        svc1 = DirectSigningService(TEST_MNEMONIC)
        svc2 = DirectSigningService(TEST_MNEMONIC)
        assert svc1.get_address() == svc2.get_address()
        assert svc1.get_pub_key() == svc2.get_pub_key()

    def test_different_mnemonic_different_address(self):
        other = "zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"
        svc1 = DirectSigningService(TEST_MNEMONIC)
        svc2 = DirectSigningService(other)
        assert svc1.get_address() != svc2.get_address()


class TestVaultSigningService:
    def test_vault_requires_endpoint_and_key(self):
        from seocheon.infrastructure.signing.vault import VaultSigningService

        with pytest.raises(ValueError, match="vault endpoint"):
            VaultSigningService("", "key")
        with pytest.raises(ValueError, match="vault endpoint"):
            VaultSigningService("http://vault", "")

    def test_vault_sync_sign_raises(self):
        from seocheon.infrastructure.signing.vault import VaultSigningService

        svc = VaultSigningService("http://vault:8080", "mykey")
        with pytest.raises(NotImplementedError):
            svc.sign(b"data")


    async def test_vault_mock_sign_verify(self):
        from unittest.mock import AsyncMock, patch

        from seocheon.infrastructure.signing.vault import VaultSigningService

        svc = VaultSigningService("http://vault:8080", "mykey")
        # Mock initialize to set address and pub_key
        svc._address = "seocheon1vaultaddr"
        svc._pub_key = b"\x02" + b"\x01" * 32

        dummy_sig = b"\xab" * 64
        mock_resp = MagicMock()
        mock_resp.status = 200
        mock_resp.json = AsyncMock(return_value={"signature": dummy_sig.hex()})
        mock_resp.__aenter__ = AsyncMock(return_value=mock_resp)
        mock_resp.__aexit__ = AsyncMock(return_value=False)

        mock_session = MagicMock()
        mock_session.post = MagicMock(return_value=mock_resp)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=False)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            result = await svc.sign_async(b"test data")
            assert result == dummy_sig
            assert len(result) == 64


class TestKeystoreSigningService:
    def _make_keystore(self, priv_key_hex: str, passphrase: str) -> str:
        """Create a minimal keystore file for testing."""
        import hashlib

        from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes

        priv_key_bytes = bytes.fromhex(priv_key_hex)
        salt = os.urandom(32)
        iv = os.urandom(16)

        derived_key = hashlib.scrypt(
            passphrase.encode("utf-8"),
            salt=salt,
            n=8192,
            r=8,
            p=1,
            dklen=32,
        )

        aes_key = derived_key[:16]
        cipher = Cipher(algorithms.AES(aes_key), modes.CTR(iv))
        encryptor = cipher.encryptor()
        ciphertext = encryptor.update(priv_key_bytes) + encryptor.finalize()

        mac_input = derived_key[16:32] + ciphertext
        mac = hashlib.sha256(mac_input).digest()

        ks = {
            "crypto": {
                "cipher": "aes-128-ctr",
                "ciphertext": ciphertext.hex(),
                "cipherparams": {"iv": iv.hex()},
                "kdf": "scrypt",
                "kdfparams": {
                    "dklen": 32,
                    "n": 8192,
                    "r": 8,
                    "p": 1,
                    "salt": salt.hex(),
                },
                "mac": mac.hex(),
            }
        }

        fd, path = tempfile.mkstemp(suffix=".json")
        with os.fdopen(fd, "w") as f:
            json.dump(ks, f)
        return path

    def test_keystore_decrypt_and_sign(self):
        from seocheon.internal.crypto.bip44 import derive_key_from_mnemonic

        pk = derive_key_from_mnemonic(TEST_MNEMONIC)
        priv_hex = pk._key.secret.hex()

        passphrase = "test-password-123"
        path = self._make_keystore(priv_hex, passphrase)
        try:
            svc = KeystoreSigningService(path, passphrase)
            assert svc.get_address().startswith("seocheon1")
            sig = svc.sign(b"test data")
            assert len(sig) == 64
        finally:
            os.unlink(path)

    def test_keystore_wrong_passphrase(self):
        from seocheon.internal.crypto.bip44 import derive_key_from_mnemonic

        pk = derive_key_from_mnemonic(TEST_MNEMONIC)
        priv_hex = pk._key.secret.hex()

        path = self._make_keystore(priv_hex, "correct-pass")
        try:
            with pytest.raises(ValueError, match="MAC verification failed"):
                KeystoreSigningService(path, "wrong-pass")
        finally:
            os.unlink(path)

    def test_keystore_requires_path_and_passphrase(self):
        with pytest.raises(ValueError, match="keystore path"):
            KeystoreSigningService("", "pass")
        with pytest.raises(ValueError, match="keystore path"):
            KeystoreSigningService("/path/to/ks", "")

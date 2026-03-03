"""KeystoreSigningService: signs transactions using a local encrypted keystore file."""

from __future__ import annotations

import hashlib
import hmac
import json
import os

from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes

from seocheon.internal.crypto.address import address_from_pubkey
from seocheon.internal.crypto.keys import PrivateKey


class KeystoreSigningService:
    """Signs transactions using a local encrypted keystore file."""

    def __init__(self, keystore_path: str, passphrase_env: str) -> None:
        if not keystore_path or not passphrase_env:
            raise ValueError("keystore path and passphrase are required")
        self._priv_key = _load_and_decrypt_keystore(keystore_path, passphrase_env)
        self._pub_key = self._priv_key.pub_key()
        self._address = address_from_pubkey(self._pub_key)

    def sign(self, tx_bytes: bytes) -> bytes:
        return self._priv_key.sign(tx_bytes)

    def get_address(self) -> str:
        return self._address

    def get_pub_key(self) -> bytes:
        return self._pub_key


def _load_and_decrypt_keystore(path: str, passphrase: str) -> PrivateKey:
    """Read a keystore file, derive the decryption key via scrypt,
    and decrypt the private key using AES-128-CTR.
    """
    with open(path) as f:
        ks = json.load(f)

    crypto = ks["crypto"]
    if crypto["kdf"] != "scrypt":
        raise ValueError(f"unsupported KDF: {crypto['kdf']} (only scrypt is supported)")
    if crypto["cipher"] != "aes-128-ctr":
        raise ValueError(f"unsupported cipher: {crypto['cipher']} (only aes-128-ctr is supported)")

    kdf_params = crypto["kdfparams"]
    salt = bytes.fromhex(kdf_params["salt"])
    iv = bytes.fromhex(crypto["cipherparams"]["iv"])
    ciphertext = bytes.fromhex(crypto["ciphertext"])
    mac = bytes.fromhex(crypto["mac"])

    dk_len = kdf_params.get("dklen", 32)
    n = kdf_params["n"]
    r = kdf_params["r"]
    p = kdf_params["p"]

    # Derive key using scrypt
    derived_key = hashlib.scrypt(
        passphrase.encode("utf-8"),
        salt=salt,
        n=n,
        r=r,
        p=p,
        dklen=dk_len,
    )

    # Verify MAC: SHA256(derivedKey[16:32] + cipherText)
    mac_input = derived_key[16:32] + ciphertext
    calculated_mac = hashlib.sha256(mac_input).digest()
    if not hmac.compare_digest(calculated_mac, mac):
        raise ValueError("MAC verification failed: incorrect passphrase or corrupted keystore")

    # Decrypt using AES-128-CTR
    aes_key = derived_key[:16]
    cipher = Cipher(algorithms.AES(aes_key), modes.CTR(iv))
    decryptor = cipher.decryptor()
    priv_key_bytes = decryptor.update(ciphertext) + decryptor.finalize()

    return PrivateKey.from_bytes(priv_key_bytes)

import {
  Bip39,
  EnglishMnemonic,
  Slip10,
  Slip10Curve,
  stringToPath,
  Secp256k1,
  sha256,
  Ripemd160,
} from "@cosmjs/crypto";
import { toBech32, toHex, fromHex } from "@cosmjs/encoding";

export interface SigningService {
  sign(txBytes: Uint8Array): Promise<Uint8Array>;
  getAddress(): Promise<string>;
  getPubKey(): Promise<Uint8Array>;
}

// BIP44 path for Cosmos: m/44'/118'/0'/0/0
const COSMOS_HD_PATH = "m/44'/118'/0'/0/0";
const ADDRESS_PREFIX = "seocheon";

export class DirectSigningService implements SigningService {
  private readonly mnemonic: string;
  private privKey: Uint8Array | null = null;
  private pubKey: Uint8Array | null = null;
  private address: string | null = null;
  private initPromise: Promise<void> | null = null;

  constructor(mnemonic: string) {
    this.mnemonic = mnemonic;
  }

  private async init(): Promise<void> {
    if (this.privKey) return;
    if (this.initPromise) return this.initPromise;

    this.initPromise = (async () => {
      const mnemonicObj = new EnglishMnemonic(this.mnemonic);
      const seed = await Bip39.mnemonicToSeed(mnemonicObj);
      const { privkey } = Slip10.derivePath(
        Slip10Curve.Secp256k1,
        seed,
        stringToPath(COSMOS_HD_PATH),
      );
      const keypair = await Secp256k1.makeKeypair(privkey);
      this.privKey = privkey;
      this.pubKey = Secp256k1.compressPubkey(keypair.pubkey);
      // Address derivation: SHA256 -> RIPEMD160 -> bech32
      const hash = sha256(this.pubKey);
      const rawAddress = new Ripemd160().update(hash).digest();
      this.address = toBech32(ADDRESS_PREFIX, rawAddress);
    })();

    return this.initPromise;
  }

  async sign(txBytes: Uint8Array): Promise<Uint8Array> {
    await this.init();
    const hash = sha256(txBytes);
    const sig = await Secp256k1.createSignature(hash, this.privKey!);
    // Return 64-byte R||S compact signature (no recovery flag)
    return new Uint8Array([...sig.r(32), ...sig.s(32)]);
  }

  async getAddress(): Promise<string> {
    await this.init();
    return this.address!;
  }

  async getPubKey(): Promise<Uint8Array> {
    await this.init();
    return this.pubKey!;
  }

  getMnemonic(): string {
    return this.mnemonic;
  }
}

export class VaultSigningService implements SigningService {
  private readonly vaultEndpoint: string;
  private readonly keyName: string;
  private cachedAddress: string | null = null;
  private cachedPubKey: Uint8Array | null = null;

  constructor(vaultEndpoint: string, keyName: string) {
    this.vaultEndpoint = vaultEndpoint.replace(/\/+$/, "");
    this.keyName = keyName;
  }

  private async fetchInit(): Promise<void> {
    if (this.cachedAddress) return;

    // Fetch address
    const addrResp = await fetch(
      `${this.vaultEndpoint}/v1/keys/${this.keyName}/address`,
    );
    if (!addrResp.ok) {
      throw new Error(`vault address fetch failed: ${addrResp.status}`);
    }
    const addrResult = (await addrResp.json()) as { address: string };
    this.cachedAddress = addrResult.address;

    // Fetch public key
    const pubKeyResp = await fetch(
      `${this.vaultEndpoint}/v1/keys/${this.keyName}/pubkey`,
    );
    if (!pubKeyResp.ok) {
      throw new Error(`vault pubkey fetch failed: ${pubKeyResp.status}`);
    }
    const pubKeyResult = (await pubKeyResp.json()) as { pubkey: string };
    this.cachedPubKey = fromHex(pubKeyResult.pubkey);
  }

  async sign(txBytes: Uint8Array): Promise<Uint8Array> {
    const resp = await fetch(
      `${this.vaultEndpoint}/v1/keys/${this.keyName}/sign`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ data: toHex(txBytes) }),
      },
    );
    if (!resp.ok) {
      const body = await resp.text();
      throw new Error(`vault sign failed: ${resp.status}: ${body}`);
    }
    const result = (await resp.json()) as { signature: string };
    return fromHex(result.signature);
  }

  async getAddress(): Promise<string> {
    await this.fetchInit();
    return this.cachedAddress!;
  }

  async getPubKey(): Promise<Uint8Array> {
    await this.fetchInit();
    return this.cachedPubKey!;
  }

  getVaultEndpoint(): string {
    return this.vaultEndpoint;
  }

  getKeyName(): string {
    return this.keyName;
  }
}

export class KeystoreSigningService implements SigningService {
  private readonly keystorePath: string;
  private readonly passphrase: string;
  private privKey: Uint8Array | null = null;
  private pubKey: Uint8Array | null = null;
  private address: string | null = null;
  private initPromise: Promise<void> | null = null;

  constructor(keystorePath: string, passphrase: string) {
    this.keystorePath = keystorePath;
    this.passphrase = passphrase;
  }

  private async init(): Promise<void> {
    if (this.privKey) return;
    if (this.initPromise) return this.initPromise;

    this.initPromise = (async () => {
      // Node.js crypto and fs modules
      const { readFile } = await import("node:fs/promises");
      const { scryptSync, createDecipheriv, createHash } = await import(
        "node:crypto"
      );

      const data = await readFile(this.keystorePath, "utf-8");
      const ks = JSON.parse(data) as {
        crypto: {
          cipher: string;
          ciphertext: string;
          cipherparams: { iv: string };
          kdf: string;
          kdfparams: {
            dklen: number;
            n: number;
            r: number;
            p: number;
            salt: string;
          };
          mac: string;
        };
      };

      if (ks.crypto.kdf !== "scrypt") {
        throw new Error(`unsupported KDF: ${ks.crypto.kdf}`);
      }
      if (ks.crypto.cipher !== "aes-128-ctr") {
        throw new Error(`unsupported cipher: ${ks.crypto.cipher}`);
      }

      const salt = Buffer.from(ks.crypto.kdfparams.salt, "hex");
      const iv = Buffer.from(ks.crypto.cipherparams.iv, "hex");
      const cipherText = Buffer.from(ks.crypto.ciphertext, "hex");
      const mac = Buffer.from(ks.crypto.mac, "hex");

      // Derive key using scrypt
      const dkLen = ks.crypto.kdfparams.dklen || 32;
      const derivedKey = scryptSync(this.passphrase, salt, dkLen, {
        N: ks.crypto.kdfparams.n,
        r: ks.crypto.kdfparams.r,
        p: ks.crypto.kdfparams.p,
      });

      // Verify MAC: SHA256(derivedKey[16:32] + cipherText)
      const macInput = Buffer.concat([
        derivedKey.subarray(16, 32),
        cipherText,
      ]);
      const calculatedMAC = createHash("sha256").update(macInput).digest();
      if (!calculatedMAC.equals(mac)) {
        throw new Error(
          "MAC verification failed: incorrect passphrase or corrupted keystore",
        );
      }

      // Decrypt using AES-128-CTR
      const decipher = createDecipheriv(
        "aes-128-ctr",
        derivedKey.subarray(0, 16),
        iv,
      );
      const privKeyBytes = Buffer.concat([
        decipher.update(cipherText),
        decipher.final(),
      ]);

      this.privKey = new Uint8Array(privKeyBytes);

      // Derive pubkey and address using same crypto as DirectSigningService
      const keypair = await Secp256k1.makeKeypair(this.privKey);
      this.pubKey = Secp256k1.compressPubkey(keypair.pubkey);
      const hash = sha256(this.pubKey);
      const rawAddress = new Ripemd160().update(hash).digest();
      this.address = toBech32(ADDRESS_PREFIX, rawAddress);
    })();

    return this.initPromise;
  }

  async sign(txBytes: Uint8Array): Promise<Uint8Array> {
    await this.init();
    const hash = sha256(txBytes);
    const sig = await Secp256k1.createSignature(hash, this.privKey!);
    return new Uint8Array([...sig.r(32), ...sig.s(32)]);
  }

  async getAddress(): Promise<string> {
    await this.init();
    return this.address!;
  }

  async getPubKey(): Promise<Uint8Array> {
    await this.init();
    return this.pubKey!;
  }

  getKeystorePath(): string {
    return this.keystorePath;
  }
}

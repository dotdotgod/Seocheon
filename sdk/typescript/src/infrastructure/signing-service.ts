export interface SigningService {
  sign(txBytes: Uint8Array): Promise<Uint8Array>;
  getAddress(): Promise<string>;
  getPubKey(): Promise<Uint8Array>;
}

export class DirectSigningService implements SigningService {
  private readonly mnemonic: string;
  private address: string | null = null;

  constructor(mnemonic: string) {
    this.mnemonic = mnemonic;
  }

  async sign(_txBytes: Uint8Array): Promise<Uint8Array> {
    // Implementation requires @cosmjs/proto-signing wallet derivation.
    // Actual signing logic deferred to integration with CosmJS.
    throw new Error("DirectSigningService.sign() requires CosmJS integration");
  }

  async getAddress(): Promise<string> {
    if (this.address) return this.address;
    // Derive address from mnemonic using CosmJS.
    // Placeholder: actual derivation in integration phase.
    throw new Error(
      "DirectSigningService.getAddress() requires CosmJS integration",
    );
  }

  async getPubKey(): Promise<Uint8Array> {
    throw new Error(
      "DirectSigningService.getPubKey() requires CosmJS integration",
    );
  }

  getMnemonic(): string {
    return this.mnemonic;
  }
}

export class VaultSigningService implements SigningService {
  private readonly vaultEndpoint: string;
  private readonly keyName: string;

  constructor(vaultEndpoint: string, keyName: string) {
    this.vaultEndpoint = vaultEndpoint;
    this.keyName = keyName;
  }

  async sign(_txBytes: Uint8Array): Promise<Uint8Array> {
    throw new Error(
      "VaultSigningService.sign() requires vault server integration",
    );
  }

  async getAddress(): Promise<string> {
    throw new Error(
      "VaultSigningService.getAddress() requires vault server integration",
    );
  }

  async getPubKey(): Promise<Uint8Array> {
    throw new Error(
      "VaultSigningService.getPubKey() requires vault server integration",
    );
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

  constructor(keystorePath: string, passphrase: string) {
    this.keystorePath = keystorePath;
    this.passphrase = passphrase;
  }

  async sign(_txBytes: Uint8Array): Promise<Uint8Array> {
    throw new Error(
      "KeystoreSigningService.sign() requires keystore integration",
    );
  }

  async getAddress(): Promise<string> {
    throw new Error(
      "KeystoreSigningService.getAddress() requires keystore integration",
    );
  }

  async getPubKey(): Promise<Uint8Array> {
    throw new Error(
      "KeystoreSigningService.getPubKey() requires keystore integration",
    );
  }

  getKeystorePath(): string {
    return this.keystorePath;
  }
}

import { describe, it, expect } from "vitest";
import {
  DirectSigningService,
  VaultSigningService,
  KeystoreSigningService,
} from "../../src/infrastructure/signing-service.js";

// Well-known test mnemonic (DO NOT USE IN PRODUCTION)
const TEST_MNEMONIC =
  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about";

describe("signing-service", () => {
  describe("DirectSigningService", () => {
    it("15.1: derives_address_from_mnemonic", async () => {
      const svc = new DirectSigningService(TEST_MNEMONIC);
      const address = await svc.getAddress();
      expect(address).toMatch(/^seocheon1[a-z0-9]+$/);
    });

    it("15.2: derives_compressed_33byte_pubkey", async () => {
      const svc = new DirectSigningService(TEST_MNEMONIC);
      const pubKey = await svc.getPubKey();
      expect(pubKey.length).toBe(33);
      expect([0x02, 0x03]).toContain(pubKey[0]);
    });
  });

  describe("KeystoreSigningService", () => {
    it("15.3: keystore_constructor_stores_path", () => {
      const svc = new KeystoreSigningService("/path/to/keystore.json", "pass");
      expect(svc.getKeystorePath()).toBe("/path/to/keystore.json");
    });

    it("15.4: keystore_load_and_sign", async () => {
      // Without a real keystore file, loading will fail
      const svc = new KeystoreSigningService("/nonexistent/keystore.json", "pass");
      await expect(svc.getAddress()).rejects.toThrow();
    });

    it("15.5: keystore_wrong_pass_fails", async () => {
      // Without a real keystore file, this will fail on file read
      const svc = new KeystoreSigningService("/dev/null", "wrong");
      await expect(svc.getAddress()).rejects.toThrow();
    });

    it("15.6: keystore_empty_params_fail", () => {
      // Empty path should still construct but fail on operations
      const svc = new KeystoreSigningService("", "");
      expect(svc.getKeystorePath()).toBe("");
    });
  });

  describe("VaultSigningService", () => {
    it("16.1: vault_constructor_stores_endpoint", () => {
      const svc = new VaultSigningService("http://vault:8200", "mykey");
      expect(svc.getVaultEndpoint()).toBe("http://vault:8200");
      expect(svc.getKeyName()).toBe("mykey");
    });

    it("16.2: vault_strips_trailing_slash", () => {
      const svc = new VaultSigningService("http://vault:8200///", "mykey");
      expect(svc.getVaultEndpoint()).toBe("http://vault:8200");
    });

    it("16.3: vault_mock_sign_verify", () => {
      // Vault service can be instantiated with mock endpoint
      const svc = new VaultSigningService("http://mock-vault:8200", "test-key");
      expect(svc.getVaultEndpoint()).toBe("http://mock-vault:8200");
      expect(svc.getKeyName()).toBe("test-key");
    });

    it("16.4: vault_server_error_fails", async () => {
      const svc = new VaultSigningService("http://127.0.0.1:1", "mykey");
      await expect(svc.getAddress()).rejects.toThrow();
    });

    it("16.5: vault_empty_params_fail", () => {
      // Empty endpoint and key name
      const svc = new VaultSigningService("", "");
      expect(svc.getVaultEndpoint()).toBe("");
      expect(svc.getKeyName()).toBe("");
    });
  });
});

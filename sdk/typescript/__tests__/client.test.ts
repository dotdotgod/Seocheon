import { describe, it, expect } from "vitest";
import { SeocheonSDK } from "../src/client.js";
import { ValidationError } from "../src/errors/errors.js";

const validConfig = {
  chain: {
    chain_id: "seocheon-1",
    rpc_endpoint: "http://localhost:26657",
    grpc_endpoint: "http://localhost:9090",
    gas_price: "250uppyeo",
    gas_adjustment: 1.3,
  },
  signing: {
    mode: "direct" as const,
    mnemonic: "test test test test test test test test test test test junk",
  },
  tx: {
    broadcast_mode: "sync" as const,
    confirm_timeout_ms: 30000,
    confirm_poll_interval_ms: 1000,
  },
};

describe("SeocheonSDK", () => {
  describe("constructor", () => {
    it("02.1: creates_sdk_with_valid_config", () => {
      const sdk = new SeocheonSDK(validConfig);
      expect(sdk).toBeDefined();
      expect(sdk.activity).toBeDefined();
      expect(sdk.epoch).toBeDefined();
      expect(sdk.node).toBeDefined();
      expect(sdk.rewards).toBeDefined();
      expect(sdk.cosmos).toBeDefined();
    });

    it("02.2: throws_on_missing_chain_id", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: { ...validConfig.chain, chain_id: "" },
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("02.3: throws_on_missing_rpc_endpoint", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: { ...validConfig.chain, rpc_endpoint: "" },
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("02.4: throws_on_missing_grpc_endpoint", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: { ...validConfig.chain, grpc_endpoint: "" },
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("02.5: throws_on_invalid_signing_mode", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: { mode: "invalid" as "direct" },
          }),
      ).toThrow(ValidationError);
    });

    it("02.6: throws_on_direct_without_mnemonic", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: { mode: "direct" },
          }),
      ).toThrow(ValidationError);
    });

    it("02.7: throws_on_vault_without_endpoint", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: { mode: "vault" },
          }),
      ).toThrow(ValidationError);
    });

    it("02.8: throws_on_keystore_without_path", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: { mode: "keystore" },
          }),
      ).toThrow(ValidationError);
    });
  });

  describe("connection", () => {
    it("03.1: starts_disconnected", () => {
      const sdk = new SeocheonSDK(validConfig);
      expect(sdk.isConnected()).toBe(false);
    });

    it("03.2: connects_and_disconnects", async () => {
      const sdk = new SeocheonSDK(validConfig);
      await sdk.connect();
      expect(sdk.isConnected()).toBe(true);
      await sdk.disconnect();
      expect(sdk.isConnected()).toBe(false);
    });
  });

  describe("config", () => {
    it("03.3: returns_config_copy", () => {
      const sdk = new SeocheonSDK(validConfig);
      const config = sdk.getConfig();
      expect(config.chain.chain_id).toBe("seocheon-1");
      expect(config.signing.mode).toBe("direct");
    });

    it("03.4: applies_default_tx_config", () => {
      const sdk = new SeocheonSDK({
        chain: validConfig.chain,
        signing: validConfig.signing,
      });
      const config = sdk.getConfig();
      expect(config.tx).toBeUndefined();
    });

    it("03.5: invalid_config_returns_error", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: { ...validConfig.chain, chain_id: "", rpc_endpoint: "" },
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });
  });
});

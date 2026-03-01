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
    it("creates SDK with valid config", () => {
      const sdk = new SeocheonSDK(validConfig);
      expect(sdk).toBeDefined();
      expect(sdk.activity).toBeDefined();
      expect(sdk.epoch).toBeDefined();
      expect(sdk.node).toBeDefined();
      expect(sdk.rewards).toBeDefined();
      expect(sdk.cosmos).toBeDefined();
    });

    it("throws on missing chain config", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: undefined as unknown as typeof validConfig.chain,
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("throws on missing chain_id", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: { ...validConfig.chain, chain_id: "" },
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("throws on missing rpc_endpoint", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: { ...validConfig.chain, rpc_endpoint: "" },
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("throws on missing grpc_endpoint", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: { ...validConfig.chain, grpc_endpoint: "" },
            signing: validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("throws on missing signing config", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: undefined as unknown as typeof validConfig.signing,
          }),
      ).toThrow(ValidationError);
    });

    it("throws on invalid signing mode", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: { mode: "invalid" as "direct" },
          }),
      ).toThrow(ValidationError);
    });

    it("throws on direct mode without mnemonic", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: { mode: "direct" },
          }),
      ).toThrow(ValidationError);
    });

    it("throws on vault mode without endpoint", () => {
      expect(
        () =>
          new SeocheonSDK({
            chain: validConfig.chain,
            signing: { mode: "vault" },
          }),
      ).toThrow(ValidationError);
    });

    it("throws on keystore mode without path", () => {
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
    it("starts disconnected", () => {
      const sdk = new SeocheonSDK(validConfig);
      expect(sdk.isConnected()).toBe(false);
    });

    it("connects and disconnects", async () => {
      const sdk = new SeocheonSDK(validConfig);
      await sdk.connect();
      expect(sdk.isConnected()).toBe(true);
      await sdk.disconnect();
      expect(sdk.isConnected()).toBe(false);
    });
  });

  describe("getConfig", () => {
    it("returns a copy of the config", () => {
      const sdk = new SeocheonSDK(validConfig);
      const config = sdk.getConfig();
      expect(config.chain.chain_id).toBe("seocheon-1");
      expect(config.signing.mode).toBe("direct");
    });
  });

  describe("tx config defaults", () => {
    it("applies default tx config when not provided", () => {
      const sdk = new SeocheonSDK({
        chain: validConfig.chain,
        signing: validConfig.signing,
      });
      const config = sdk.getConfig();
      expect(config.tx).toBeUndefined();
    });
  });
});

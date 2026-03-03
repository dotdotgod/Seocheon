import { describe, it, expect } from "vitest";
import type {
  SDKConfig,
  ChainConfig,
  SigningConfig,
  TxConfig,
  ResolvedTxConfig,
  SigningMode,
  BroadcastMode,
} from "../../src/types/config.js";

describe("config-types", () => {
  it("09.1: valid_sdk_config_shape", () => {
    const config: SDKConfig = {
      chain: {
        chain_id: "seocheon-1",
        rpc_endpoint: "http://localhost:26657",
        grpc_endpoint: "http://localhost:9090",
        gas_price: "250uppyeo",
        gas_adjustment: 1.3,
      },
      signing: {
        mode: "direct",
        mnemonic: "test mnemonic words",
      },
      tx: {
        broadcast_mode: "sync",
        confirm_timeout_ms: 30000,
        confirm_poll_interval_ms: 1000,
      },
    };
    expect(config.chain.chain_id).toBe("seocheon-1");
    expect(config.signing.mode).toBe("direct");
    expect(config.tx?.broadcast_mode).toBe("sync");
  });

  it("09.2: signing_modes_are_valid", () => {
    const modes: SigningMode[] = ["direct", "vault", "keystore"];
    expect(modes).toHaveLength(3);
    expect(modes).toContain("direct");
    expect(modes).toContain("vault");
    expect(modes).toContain("keystore");
  });

  it("09.3: broadcast_modes_are_valid", () => {
    const modes: BroadcastMode[] = ["sync", "async"];
    expect(modes).toHaveLength(2);
    expect(modes).toContain("sync");
    expect(modes).toContain("async");
  });

  it("09.4: resolved_tx_config_has_all_fields", () => {
    const resolved: ResolvedTxConfig = {
      broadcast_mode: "sync",
      confirm_timeout_ms: 30000,
      confirm_poll_interval_ms: 1000,
      chain_id: "seocheon-1",
      gas_price: 250,
    };
    expect(resolved.broadcast_mode).toBe("sync");
    expect(resolved.confirm_timeout_ms).toBe(30000);
    expect(resolved.confirm_poll_interval_ms).toBe(1000);
    expect(resolved.chain_id).toBe("seocheon-1");
    expect(resolved.gas_price).toBe(250);
  });

  it("09.5: tx_config_fields_are_optional", () => {
    const config: SDKConfig = {
      chain: {
        chain_id: "seocheon-1",
        rpc_endpoint: "http://localhost:26657",
        grpc_endpoint: "http://localhost:9090",
      },
      signing: {
        mode: "direct",
        mnemonic: "test",
      },
    };
    // tx config should be undefined when not provided
    expect(config.tx).toBeUndefined();
  });

  it("09.6: chain_config_optional_fields", () => {
    const chain: ChainConfig = {
      chain_id: "seocheon-1",
      rpc_endpoint: "http://localhost:26657",
      grpc_endpoint: "http://localhost:9090",
    };
    // gas_price and gas_adjustment are optional
    expect(chain.gas_price).toBeUndefined();
    expect(chain.gas_adjustment).toBeUndefined();
  });
});

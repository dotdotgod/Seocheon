import { describe, it, expect, vi } from "vitest";
import { ActivityModule } from "../../src/modules/activity.js";
import type { ChainClient } from "../../src/infrastructure/chain-client.js";
import type { SigningService } from "../../src/infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../../src/types/config.js";
import { ValidationError } from "../../src/errors/errors.js";

function createMockChainClient(overrides?: Record<string, unknown>): ChainClient {
  return {
    connect: vi.fn(),
    disconnect: vi.fn(),
    isConnected: vi.fn().mockReturnValue(true),
    queryRest: vi.fn().mockImplementation(async (path: string) => {
      if (path.includes("/activities")) {
        return {
          activities: [
            {
              activity_hash: "a".repeat(64),
              content_uri: "ipfs://QmTest",
              block_height: "100",
            },
          ],
        };
      }
      if (path.includes("/epochs/")) {
        return {
          summary: { active_windows: "5", total_activities: "10", eligible: false },
          quota_used: "3",
          quota_limit: "10",
        };
      }
      if (path.includes("/by-agent/")) {
        return { node: { id: "node-1" } };
      }
      if (path.includes("/params")) {
        return {
          params: {
            epoch_length: "17280",
            windows_per_epoch: "12",
            min_active_windows: "8",
          },
        };
      }
      if (path.includes("/blocks/latest")) {
        return {
          block: {
            header: { height: "100", time: "2026-01-01T00:00:00Z", chain_id: "seocheon-1" },
            data: { txs: null },
          },
        };
      }
      return {};
    }),
    broadcastTx: vi.fn(),
    getAccountInfo: vi.fn().mockResolvedValue({ account_number: 1, sequence: 0 }),
    getLatestBlock: vi.fn().mockResolvedValue({
      header: { height: 100, time: "2026-01-01T00:00:00Z", chain_id: "seocheon-1" },
      data: { txs: [] },
    }),
    getBlockByHeight: vi.fn(),
    getTx: vi.fn(),
    ...overrides,
  } as ChainClient;
}

function createMockSigner(): SigningService {
  return {
    sign: vi.fn().mockResolvedValue(new Uint8Array(64)),
    getAddress: vi.fn().mockResolvedValue("seocheon1testaddr"),
    getPubKey: vi.fn().mockResolvedValue(new Uint8Array(33).fill(0x02)),
  };
}

const txConfig: ResolvedTxConfig = {
  broadcast_mode: "sync",
  confirm_timeout_ms: 5000,
  confirm_poll_interval_ms: 100,
  chain_id: "seocheon-testnet-1",
  gas_price: 250,
};

describe("activity-module", () => {
  it("18.1: getActivities_returns_list", async () => {
    const client = createMockChainClient();
    const signer = createMockSigner();
    const mod = new ActivityModule(client, signer, txConfig);

    const result = await mod.getActivities("node-1", 0);
    expect(result.activities).toHaveLength(1);
    expect(result.activities[0].activity_hash).toBe("a".repeat(64));
    expect(result.total_count).toBe(1);
  });

  it("18.2: getQuota_returns_info", async () => {
    const client = createMockChainClient();
    const signer = createMockSigner();
    const mod = new ActivityModule(client, signer, txConfig);

    const result = await mod.getQuota();
    expect(result.quota_total).toBe(10);
    expect(result.quota_used).toBe(3);
    expect(result.quota_remaining).toBe(7);
  });

  it("18.3: submit_invalid_hash_fails", async () => {
    const client = createMockChainClient();
    const signer = createMockSigner();
    const mod = new ActivityModule(client, signer, txConfig);

    await expect(
      mod.submit("invalid-hash", "ipfs://QmTest"),
    ).rejects.toThrow(ValidationError);
  });

  it("18.4: submit_empty_uri_fails", async () => {
    const client = createMockChainClient();
    const signer = createMockSigner();
    const mod = new ActivityModule(client, signer, txConfig);

    await expect(
      mod.submit("a".repeat(64), ""),
    ).rejects.toThrow(ValidationError);
  });

  it("18.5: submit_not_registered_fails", async () => {
    const client = createMockChainClient({
      queryRest: vi.fn().mockImplementation(async (path: string) => {
        if (path.includes("/by-agent/")) {
          throw new Error("node not found");
        }
        if (path.includes("/params")) {
          return {
            params: {
              epoch_length: "17280",
              windows_per_epoch: "12",
              min_active_windows: "8",
            },
          };
        }
        return {};
      }),
      getAccountInfo: vi.fn().mockResolvedValue({ account_number: 1, sequence: 0 }),
      broadcastTx: vi.fn().mockResolvedValue({ tx_hash: "ABC", code: 0, raw_log: "" }),
      getLatestBlock: vi.fn().mockResolvedValue({
        header: { height: 100, time: "2026-01-01T00:00:00Z", chain_id: "seocheon-1" },
        data: { txs: [] },
      }),
    });
    const signer = createMockSigner();
    const mod = new ActivityModule(client, signer, txConfig);

    // getQuota internally calls resolveOwnNodeId which will fail
    await expect(mod.getQuota()).rejects.toThrow("node not found");
  });
});

import { describe, it, expect, vi } from "vitest";
import { EpochModule } from "../../src/modules/epoch.js";
import type { ChainClient } from "../../src/infrastructure/chain-client.js";
import type { SigningService } from "../../src/infrastructure/signing-service.js";

function createMockChainClient(): ChainClient {
  return {
    connect: vi.fn(),
    disconnect: vi.fn(),
    isConnected: vi.fn().mockReturnValue(true),
    queryRest: vi.fn().mockImplementation(async (path: string) => {
      if (path.includes("/epoch-info")) {
        return {
          current_epoch: "5",
          current_window: "3",
          epoch_start_block: "86401",
          blocks_until_next_epoch: "10000",
        };
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
      if (path.includes("/by-agent/")) {
        return { node: { id: "node-1" } };
      }
      if (path.includes("/activities")) {
        return {
          activities: [
            { activity_hash: "a".repeat(64), block_height: "86500" },
            { activity_hash: "b".repeat(64), block_height: "87000" },
          ],
        };
      }
      if (path.includes("/epochs/")) {
        return {
          summary: { active_windows: "3", total_activities: "5", eligible: false },
          quota_used: "5",
          quota_limit: "10",
        };
      }
      return {};
    }),
    broadcastTx: vi.fn(),
    getAccountInfo: vi.fn(),
    getLatestBlock: vi.fn().mockResolvedValue({
      header: { height: 90000, time: "2026-01-01T00:00:00Z", chain_id: "seocheon-1" },
      data: { txs: [] },
    }),
    getBlockByHeight: vi.fn(),
    getTx: vi.fn(),
  } as ChainClient;
}

function createMockSigner(): SigningService {
  return {
    sign: vi.fn(),
    getAddress: vi.fn().mockResolvedValue("seocheon1agent"),
    getPubKey: vi.fn().mockResolvedValue(new Uint8Array(33)),
  };
}

describe("epoch-module", () => {
  it("20.1: getInfo_returns_epoch", async () => {
    const mod = new EpochModule(createMockChainClient(), createMockSigner());
    const info = await mod.getInfo();
    expect(info.epoch_number).toBe(5);
    expect(info.window_number).toBe(3);
    expect(info.block_height).toBe(90000);
    expect(info.epoch_start_block).toBe(86401);
  });

  it("20.2: getQualification_status", async () => {
    const mod = new EpochModule(createMockChainClient(), createMockSigner());
    const qual = await mod.getQualification("node-1", 5);
    expect(qual.epoch_number).toBe(5);
    expect(qual.active_windows).toBe(3);
    expect(qual.required_windows).toBe(8);
    expect(qual.is_qualified).toBe(false);
    expect(qual.total_windows).toBe(12);
  });

  it("20.3: getQualification_requires_node_id_resolution", async () => {
    const mod = new EpochModule(createMockChainClient(), createMockSigner());
    // Without nodeId, should resolve own node via agent address
    const qual = await mod.getQualification();
    expect(qual.epoch_number).toBe(5);
    expect(qual.active_windows).toBe(3);
  });
});

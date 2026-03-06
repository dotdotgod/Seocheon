import { describe, it, expect, vi } from "vitest";
import { NodeModule } from "../../src/modules/node.js";
import type { ChainClient } from "../../src/infrastructure/chain-client.js";
import type { SigningService } from "../../src/infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../../src/types/config.js";

function createMockChainClient(): ChainClient {
  return {
    connect: vi.fn(),
    disconnect: vi.fn(),
    isConnected: vi.fn().mockReturnValue(true),
    queryRest: vi.fn().mockImplementation(async (path: string) => {
      if (path.includes("/by-agent/")) {
        return { node: { id: "node-1" } };
      }
      if (path.includes("/by-tag/")) {
        return {
          nodes: [
            {
              id: "node-1",
              status: 2,
              tags: ["ai"],
              description: "AI Node",
              validator_address: "seocheonvaloper1abc",
              registered_at: "100",
            },
          ],
        };
      }
      if (path.match(/\/nodes\/node-1$/)) {
        return {
          node: {
            id: "node-1",
            operator: "seocheon1op",
            agent_address: "seocheon1agent",
            status: 2,
            description: "Test Node",
            website: "https://test.com",
            tags: ["ai", "agent"],
            agent_share: "20",
            validator_address: "seocheonvaloper1abc",
            registered_at: "50",
          },
        };
      }
      if (path.match(/\/nodes\/nonexistent$/)) {
        throw new Error("node not found");
      }
      if (path.includes("/nodes")) {
        return {
          nodes: [
            {
              id: "node-1",
              status: 2,
              tags: ["ai"],
              description: "AI Node",
              validator_address: "seocheonvaloper1abc",
              registered_at: "100",
            },
          ],
        };
      }
      if (path.includes("/validators/")) {
        if (path.includes("/delegations/")) {
          return { delegation_response: { balance: { amount: "5000000000" } } };
        }
        return {
          validator: {
            tokens: "10000000000",
            commission: { commission_rates: { rate: "0.100000" } },
          },
        };
      }
      return {};
    }),
    broadcastTx: vi.fn(),
    getAccountInfo: vi.fn(),
    getLatestBlock: vi.fn(),
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

describe("node-module", () => {
  it("19.1: getInfo_by_id", async () => {
    const mod = new NodeModule(createMockChainClient(), createMockSigner());
    const info = await mod.getInfo("node-1");
    expect(info.node_id).toBe("node-1");
    expect(info.operator).toBe("seocheon1op");
    expect(info.status).toBe("ACTIVE");
    expect(info.tags).toContain("ai");
  });

  it("19.2: getInfo_own_node", async () => {
    const mod = new NodeModule(createMockChainClient(), createMockSigner());
    // When no nodeId is passed, resolves own node via agent address
    const info = await mod.getInfo();
    expect(info.node_id).toBe("node-1");
  });

  it("19.3: search_nodes", async () => {
    const mod = new NodeModule(createMockChainClient(), createMockSigner());
    const result = await mod.search("ai");
    expect(result.nodes.length).toBeGreaterThan(0);
    expect(result.nodes[0].node_id).toBe("node-1");
  });

  it("19.4: not_found_fails", async () => {
    const mod = new NodeModule(createMockChainClient(), createMockSigner());
    await expect(mod.getInfo("nonexistent")).rejects.toThrow("node not found");
  });

  it("19.5: getDelegationStatus", async () => {
    const client = createMockChainClient();
    (client.queryRest as ReturnType<typeof vi.fn>).mockImplementation(async (path: string) => {
      if (path.includes("/delegation-confirmation/")) {
        return {
          expiry_epoch: "90",
          current_epoch: "5",
          in_renewal_window: false,
          renewal_window_start: "83",
        };
      }
      return {};
    });
    const mod = new NodeModule(client, createMockSigner());
    const status = await mod.getDelegationStatus("seocheon1del", "seocheonvaloper1val");
    expect(status.expiry_epoch).toBe(90);
    expect(status.current_epoch).toBe(5);
    expect(status.in_renewal_window).toBe(false);
    expect(status.renewal_window_start).toBe(83);
  });

  it("19.5b: search_all_nodes_no_tag", async () => {
    const mod = new NodeModule(createMockChainClient(), createMockSigner());
    const result = await mod.search();
    expect(result.nodes.length).toBeGreaterThan(0);
    expect(result.total_count).toBeGreaterThanOrEqual(1);
  });

  it("19.5c: search_with_status_filter", async () => {
    const client = createMockChainClient();
    (client.queryRest as ReturnType<typeof vi.fn>).mockImplementation(async (path: string) => {
      if (path.includes("/nodes")) {
        return {
          nodes: [
            { id: "node-1", status: 2, tags: ["ai"], description: "Active", validator_address: "seocheonvaloper1a", registered_at: "100" },
            { id: "node-2", status: 1, tags: ["data"], description: "Registered", validator_address: "seocheonvaloper1b", registered_at: "200" },
          ],
        };
      }
      if (path.includes("/validators/")) {
        return { validator: { tokens: "5000" } };
      }
      return {};
    });
    const mod = new NodeModule(client, createMockSigner());
    const result = await mod.search(undefined, "ACTIVE");
    expect(result.nodes.every((n) => n.status === "ACTIVE")).toBe(true);
  });

  it("19.5d: confirmDelegation_without_txConfig_throws", async () => {
    const mod = new NodeModule(createMockChainClient(), createMockSigner());
    await expect(mod.confirmDelegation("seocheonvaloper1val")).rejects.toThrow(
      "txConfig is required",
    );
  });

  it("19.6: confirmDelegation", async () => {
    const client = createMockChainClient();
    (client.getAccountInfo as ReturnType<typeof vi.fn>).mockResolvedValue({ account_number: 1, sequence: 0 });
    (client.broadcastTx as ReturnType<typeof vi.fn>).mockResolvedValue({ tx_hash: "ABCDEF", code: 0, raw_log: "" });
    (client.getTx as ReturnType<typeof vi.fn>).mockResolvedValue({
      tx_hash: "ABCDEF",
      height: 100,
      tx_result: { code: 0, gas_used: 50000, gas_wanted: 200000, log: "", events: [] },
    });
    const txConfig: ResolvedTxConfig = {
      broadcast_mode: "sync",
      confirm_timeout_ms: 5000,
      confirm_poll_interval_ms: 100,
      chain_id: "seocheon-testnet-1",
      gas_price: 250,
    };
    const signer = {
      sign: vi.fn().mockResolvedValue(new Uint8Array(64)),
      getAddress: vi.fn().mockResolvedValue("seocheon1agent"),
      getPubKey: vi.fn().mockResolvedValue(new Uint8Array(33).fill(0x02)),
    };
    const mod = new NodeModule(client, signer, txConfig);
    const result = await mod.confirmDelegation("seocheonvaloper1val");
    expect(result.tx_hash).toBe("ABCDEF");
    expect(result.code).toBe(0);
  });
});

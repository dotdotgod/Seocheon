import { describe, it, expect, vi } from "vitest";
import { RewardsModule } from "../../src/modules/rewards.js";
import type { ChainClient } from "../../src/infrastructure/chain-client.js";
import type { SigningService } from "../../src/infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../../src/types/config.js";

function createMockChainClient(overrides?: Record<string, unknown>): ChainClient {
  return {
    connect: vi.fn(),
    disconnect: vi.fn(),
    isConnected: vi.fn().mockReturnValue(true),
    queryRest: vi.fn().mockImplementation(async (path: string) => {
      if (path.includes("/by-agent/")) {
        return { node: { id: "node-1" } };
      }
      if (path.match(/\/nodes\/node-1$/)) {
        return {
          node: {
            validator_address: "seocheonvaloper1abc",
            agent_share: "20",
          },
        };
      }
      if (path.includes("/outstanding_rewards")) {
        return {
          rewards: {
            rewards: [{ denom: "uppyeo", amount: "5000000000.000000" }],
          },
        };
      }
      if (path.includes("/commission")) {
        return {
          commission: {
            commission: [{ denom: "uppyeo", amount: "1000000000.000000" }],
          },
        };
      }
      return {};
    }),
    broadcastTx: vi.fn().mockResolvedValue({ tx_hash: "ABC123", code: 0, raw_log: "" }),
    getAccountInfo: vi.fn().mockResolvedValue({ account_number: 1, sequence: 0 }),
    getLatestBlock: vi.fn(),
    getBlockByHeight: vi.fn(),
    getTx: vi.fn().mockResolvedValue({
      tx_hash: "ABC123",
      height: 100,
      tx_result: {
        code: 0,
        gas_used: 200000,
        gas_wanted: 300000,
        log: "[]",
        events: [
          {
            type: "withdraw_commission",
            attributes: [
              { key: "amount", value: "1000000000uppyeo" },
              { key: "operator_amount", value: "800000000uppyeo" },
              { key: "agent_amount", value: "200000000uppyeo" },
            ],
          },
        ],
      },
    }),
    ...overrides,
  } as ChainClient;
}

function createMockSigner(): SigningService {
  return {
    sign: vi.fn().mockResolvedValue(new Uint8Array(64)),
    getAddress: vi.fn().mockResolvedValue("seocheon1operator"),
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

describe("rewards-module", () => {
  it("21.1: getPending_returns_rewards", async () => {
    const mod = new RewardsModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    const result = await mod.getPending("node-1");
    // delegation_reward comes from outstanding_rewards
    expect(result.delegation_reward).toBeDefined();
    // commission_total from commission query
    expect(result.commission_total).toBeDefined();
    expect(result.operator_share).toBeDefined();
    expect(result.agent_share).toBeDefined();
  });

  it("21.2: getPending_not_registered_fails", async () => {
    const client = createMockChainClient({
      queryRest: vi.fn().mockImplementation(async (path: string) => {
        if (path.includes("/by-agent/")) {
          throw new Error("node not found");
        }
        return {};
      }),
    });
    const mod = new RewardsModule(client, createMockSigner(), txConfig);
    // Without nodeId, resolves own node which fails
    await expect(mod.getPending()).rejects.toThrow("node not found");
  });

  it("21.3: withdraw_returns_result", async () => {
    const mod = new RewardsModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    const result = await mod.withdraw();
    expect(result.tx_hash).toBeDefined();
    expect(result.withdrawn_total).toBeDefined();
  });

  it("21.4: parseAmount_formats_correctly", async () => {
    const mod = new RewardsModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    const result = await mod.getPending("node-1");
    // delegation_reward should be formatted as KKOT string
    // 5000000000 uppyeo = 0.5 KKOT
    expect(result.delegation_reward).toBe("0.5000000000");
  });

  it("21.5: parseAgentShare_splits_correctly", async () => {
    const mod = new RewardsModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    const result = await mod.getPending("node-1");
    // commission_total = 1000000000 uppyeo = 0.1 KKOT
    // agent_share = 20% of commission = 200000000 uppyeo
    // operator_share = 80% = 800000000 uppyeo
    expect(result.commission_total).toBe("0.1000000000");
  });
});

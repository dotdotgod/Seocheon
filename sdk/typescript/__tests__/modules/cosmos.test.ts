import { describe, it, expect, vi } from "vitest";
import { CosmosModule } from "../../src/modules/cosmos.js";
import type { ChainClient } from "../../src/infrastructure/chain-client.js";
import type { SigningService } from "../../src/infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../../src/types/config.js";
import { QueryError } from "../../src/errors/errors.js";

function createMockChainClient(): ChainClient {
  return {
    connect: vi.fn(),
    disconnect: vi.fn(),
    isConnected: vi.fn().mockReturnValue(true),
    queryRest: vi.fn().mockImplementation(async (path: string) => {
      if (path.includes("/balances/")) {
        return {
          balance: { denom: "uppyeo", amount: "50000000000" },
        };
      }
      return {};
    }),
    broadcastTx: vi.fn().mockResolvedValue({ tx_hash: "TX123", code: 0, raw_log: "" }),
    getAccountInfo: vi.fn().mockResolvedValue({ account_number: 1, sequence: 0 }),
    getLatestBlock: vi.fn().mockResolvedValue({
      header: { height: 12345, time: "2026-01-01T12:00:00Z", chain_id: "seocheon-1" },
      data: { txs: [] },
    }),
    getBlockByHeight: vi.fn(),
    getTx: vi.fn().mockImplementation(async (hash: string) => {
      if (hash === "NOTFOUND") return null;
      return {
        tx_hash: hash,
        height: 100,
        tx_result: {
          code: 0,
          gas_used: 150000,
          gas_wanted: 200000,
          log: "[]",
          events: [
            {
              type: "transfer",
              attributes: [{ key: "amount", value: "1000uppyeo" }],
            },
          ],
        },
      };
    }),
  } as ChainClient;
}

function createMockSigner(): SigningService {
  return {
    sign: vi.fn().mockResolvedValue(new Uint8Array(64)),
    getAddress: vi.fn().mockResolvedValue("seocheon1sender"),
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

describe("cosmos-module", () => {
  it("22.1: getBalance", async () => {
    const mod = new CosmosModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    const result = await mod.getBalance("seocheon1test");
    expect(result.address).toBe("seocheon1test");
    expect(result.balance).toBe("50000000000");
    expect(result.balance_kkot).toBe("5.0000000000");
  });

  it("22.2: sendTokens_success", async () => {
    const mod = new CosmosModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    const result = await mod.sendTokens("seocheon1receiver", "1000000000");
    expect(result.tx_hash).toBeDefined();
    expect(result.block_height).toBeDefined();
  });

  it("22.3: sendTokens_empty_address_fails", async () => {
    const client = createMockChainClient();
    (client.getAccountInfo as ReturnType<typeof vi.fn>).mockRejectedValue(
      new Error("invalid address"),
    );
    const mod = new CosmosModule(client, createMockSigner(), txConfig);
    await expect(
      mod.sendTokens("", "1000000000"),
    ).rejects.toThrow();
  });

  it("22.4: getBlockInfo", async () => {
    const mod = new CosmosModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    const result = await mod.getBlockInfo();
    expect(result.block_height).toBe(12345);
    expect(result.chain_id).toBe("seocheon-1");
    expect(result.num_txs).toBe(0);
  });

  it("22.5: getTxResult_not_found", async () => {
    const mod = new CosmosModule(
      createMockChainClient(),
      createMockSigner(),
      txConfig,
    );
    await expect(mod.getTxResult("NOTFOUND")).rejects.toThrow(QueryError);
  });
});

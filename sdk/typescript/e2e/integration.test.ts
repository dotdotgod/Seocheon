import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { createHash } from "node:crypto";
import { SeocheonSDK } from "../src/client.js";
import type { SDKConfig } from "../src/types/config.js";

// ---------------------------------------------------------------------------
// Skip helpers
// ---------------------------------------------------------------------------

const GRPC = process.env.SEOCHEON_GRPC ?? "";
const MNEMONIC = process.env.SEOCHEON_MNEMONIC ?? "";
const RPC = process.env.SEOCHEON_RPC ?? "http://localhost:26657";
const CHAIN_ID = process.env.SEOCHEON_CHAIN_ID ?? "seocheon-e2e";

const shouldSkip = !GRPC || !MNEMONIC;

// ---------------------------------------------------------------------------
// SDK factory
// ---------------------------------------------------------------------------

function buildConfig(): SDKConfig {
  return {
    chain: {
      chain_id: CHAIN_ID,
      rpc_endpoint: RPC,
      grpc_endpoint: GRPC,
      gas_price: "250uppyeo",
    },
    signing: {
      mode: "direct",
      mnemonic: MNEMONIC,
    },
  };
}

// ---------------------------------------------------------------------------
// Suites
// ---------------------------------------------------------------------------

describe.skipIf(shouldSkip)("E2E: 연결 및 기본 쿼리", () => {
  let sdk: SeocheonSDK;

  beforeAll(async () => {
    sdk = new SeocheonSDK(buildConfig());
    await sdk.connect();
  });

  afterAll(async () => {
    await sdk.disconnect();
  });

  it("Connect — isConnected()가 true를 반환해야 한다", () => {
    expect(sdk.isConnected()).toBe(true);
  });

  it("GetBlockInfo — 양수 blockHeight를 반환해야 한다", async () => {
    const block = await sdk.cosmos.getBlockInfo();
    expect(block.block_height).toBeGreaterThan(0);
    console.log(`최신 블록: height=${block.block_height} chainId=${block.chain_id}`);
  });

  it("Node.search — x/node 엔드포인트가 응답해야 한다", async () => {
    const resp = await sdk.node.search(undefined, undefined, 10, "asc");
    expect(resp).toBeDefined();
    console.log(`x/node nodes: total=${resp.total_count}`);
  });

  it("Epoch.getInfo — 에포크 정보를 반환해야 한다", async () => {
    const info = await sdk.epoch.getInfo();
    expect(info.block_height).toBeGreaterThan(0);
    console.log(`에포크: epoch=${info.epoch_number} window=${info.window_number} height=${info.block_height}`);
  });

  it("Cosmos.getBalance — 에이전트 잔액을 반환해야 한다", async () => {
    const resp = await sdk.cosmos.getBalance();
    expect(resp.address).toBeTruthy();
    console.log(`잔액: ${resp.balance} uppyeo (${resp.balance_kkot} kkot) 주소=${resp.address}`);
  });
});

describe.skipIf(shouldSkip)("E2E: TX 제출 — MsgSubmitActivity", () => {
  let sdk: SeocheonSDK;

  beforeAll(async () => {
    sdk = new SeocheonSDK(buildConfig());
    await sdk.connect();
  });

  afterAll(async () => {
    await sdk.disconnect();
  });

  it("Activity.submit — TxHash를 반환해야 한다", async () => {
    const hash = createHash("sha256")
      .update(`e2e-activity-${Date.now()}`)
      .digest("hex");
    const contentUri = "https://example.com/e2e-activity";

    const resp = await sdk.activity.submit(hash, contentUri);
    expect(resp.tx_hash).toBeTruthy();
    console.log(
      `활동 제출 성공: txHash=${resp.tx_hash} height=${resp.block_height} epoch=${resp.epoch_number} window=${resp.window_number}`,
    );
  }, 90_000);
});

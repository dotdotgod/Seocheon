import { describe, it, expect, vi } from "vitest";
import {
  executeTx,
  pollTxConfirmation,
} from "../../src/infrastructure/tx-pipeline.js";
import type {
  PipelineConfig,
  Signer,
  ChainQuerier,
  TxResult,
} from "../../src/infrastructure/tx-pipeline.js";
import { MsgSubmitActivity } from "../../src/infrastructure/messages.js";

function createMockSigner(
  address: string = "seocheon1test",
  pubKey: Uint8Array = new Uint8Array(33).fill(0x02),
): Signer {
  return {
    sign: vi.fn().mockResolvedValue(new Uint8Array(64).fill(0xab)),
    getAddress: vi.fn().mockResolvedValue(address),
    getPubKey: vi.fn().mockResolvedValue(pubKey),
  };
}

function createMockQuerier(overrides?: {
  accountNumber?: number;
  sequence?: number;
  broadcastCode?: number;
  txResult?: TxResult;
  getTxFails?: number;
  accountFails?: boolean;
}): ChainQuerier {
  const opts = {
    accountNumber: 42,
    sequence: 5,
    broadcastCode: 0,
    getTxFails: 0,
    accountFails: false,
    ...overrides,
  };

  let getTxCallCount = 0;
  const txResult: TxResult = opts.txResult ?? {
    txHash: "ABCDEF1234567890",
    height: 100,
    code: 0,
    gasUsed: 180000,
    gasWanted: 200000,
    rawLog: "[]",
    events: [
      {
        type: "message",
        attributes: [{ key: "action", value: "submit_activity" }],
      },
    ],
  };

  return {
    getAccountInfo: opts.accountFails
      ? vi.fn().mockRejectedValue(new Error("account not found"))
      : vi.fn().mockResolvedValue({
          accountNumber: opts.accountNumber,
          sequence: opts.sequence,
        }),
    broadcastTxSync: vi.fn().mockResolvedValue({
      txHash: txResult.txHash,
      code: opts.broadcastCode,
      rawLog: opts.broadcastCode === 0 ? "" : "error",
    }),
    getTxResult: vi.fn().mockImplementation(async () => {
      getTxCallCount++;
      if (getTxCallCount <= opts.getTxFails) {
        throw new Error("tx not found");
      }
      return txResult;
    }),
  };
}

const defaultConfig: PipelineConfig = {
  chainId: "seocheon-testnet-1",
  gasPrice: 250,
  confirmTimeoutMs: 5000,
  pollIntervalMs: 100,
};

describe("tx-pipeline", () => {
  it("14.1: execute_full_pipeline", async () => {
    const signer = createMockSigner();
    const querier = createMockQuerier();
    const msg = new MsgSubmitActivity(
      "seocheon1test",
      "a".repeat(64),
      "ipfs://test",
    );

    const result = await executeTx(querier, signer, defaultConfig, {
      message: msg,
    });

    expect(result.txHash).toBe("ABCDEF1234567890");
    expect(result.height).toBe(100);
    expect(result.code).toBe(0);
    expect(signer.sign).toHaveBeenCalledOnce();
    expect(querier.broadcastTxSync).toHaveBeenCalledOnce();
  });

  it("14.2: execute_custom_gas_and_fee", async () => {
    const signer = createMockSigner();
    const querier = createMockQuerier();
    const msg = new MsgSubmitActivity("s", "a".repeat(64), "u");

    await executeTx(querier, signer, defaultConfig, {
      message: msg,
      gasLimit: 500000,
      feeAmount: 100000000,
    });

    expect(signer.sign).toHaveBeenCalledOnce();
  });

  it("14.3: execute_with_memo", async () => {
    const signer = createMockSigner();
    const querier = createMockQuerier();
    const msg = new MsgSubmitActivity("s", "a".repeat(64), "u");

    await executeTx(querier, signer, defaultConfig, {
      message: msg,
      memo: "test memo",
    });

    expect(signer.sign).toHaveBeenCalledOnce();
  });

  it("14.4: broadcast_failure_returns_error_code", async () => {
    const signer = createMockSigner();
    const querier = createMockQuerier({ broadcastCode: 4 });
    const msg = new MsgSubmitActivity("s", "a".repeat(64), "u");

    const result = await executeTx(querier, signer, defaultConfig, {
      message: msg,
    });

    expect(result.code).toBe(4);
    expect(querier.getTxResult).not.toHaveBeenCalled();
  });

  it("14.5: poll_tx_confirmation_with_retries_and_timeout", async () => {
    // Polls until found
    let callCount = 0;
    const expectedResult: TxResult = {
      txHash: "DEF",
      height: 75,
      code: 0,
      gasUsed: 150000,
      gasWanted: 200000,
      rawLog: "[]",
      events: [],
    };
    const querier: ChainQuerier = {
      getAccountInfo: vi.fn(),
      broadcastTxSync: vi.fn(),
      getTxResult: vi.fn().mockImplementation(async () => {
        callCount++;
        if (callCount < 3) throw new Error("not found");
        return expectedResult;
      }),
    };

    const result = await pollTxConfirmation(querier, "DEF", 5000, 50);
    expect(result).toEqual(expectedResult);
    expect(callCount).toBe(3);

    // Timeout case
    const failQuerier: ChainQuerier = {
      getAccountInfo: vi.fn(),
      broadcastTxSync: vi.fn(),
      getTxResult: vi.fn().mockRejectedValue(new Error("not found")),
    };
    await expect(
      pollTxConfirmation(failQuerier, "TIMEOUT", 200, 50),
    ).rejects.toThrow("timeout");
  });

  it("14.6: account_info_error_propagates", async () => {
    const signer = createMockSigner();
    const querier = createMockQuerier({ accountFails: true });
    const msg = new MsgSubmitActivity("s", "a".repeat(64), "u");

    await expect(
      executeTx(querier, signer, defaultConfig, { message: msg }),
    ).rejects.toThrow("account not found");
  });

  it("14.7: signing_error_propagates", async () => {
    const signer = createMockSigner();
    (signer.sign as ReturnType<typeof vi.fn>).mockRejectedValue(
      new Error("signing failed"),
    );
    const querier = createMockQuerier();
    const msg = new MsgSubmitActivity("s", "a".repeat(64), "u");

    await expect(
      executeTx(querier, signer, defaultConfig, { message: msg }),
    ).rejects.toThrow("signing failed");
  });
});

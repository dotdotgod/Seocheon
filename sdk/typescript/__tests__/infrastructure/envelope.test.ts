import { describe, it, expect } from "vitest";
import {
  encodeTxBody,
  encodeAuthInfo,
  encodeSignDoc,
  encodeTxRaw,
} from "../../src/infrastructure/envelope.js";
import {
  MsgSubmitActivity,
  Coin,
} from "../../src/infrastructure/messages.js";

describe("envelope", () => {
  it("12.1: encode_tx_body_single_msg", () => {
    const msg = new MsgSubmitActivity(
      "seocheon1abc",
      "a".repeat(64),
      "ipfs://QmTest",
    );
    const result = encodeTxBody([msg], "", 0);
    expect(result.length).toBeGreaterThan(0);
    // Should contain field 1 tag (0x0a) for the Any message
    expect(result[0]).toBe(0x0a);
  });

  it("12.2: encode_tx_body_with_memo", () => {
    const msg = new MsgSubmitActivity(
      "seocheon1abc",
      "a".repeat(64),
      "ipfs://QmTest",
    );
    const withMemo = encodeTxBody([msg], "test memo", 0);
    const withoutMemo = encodeTxBody([msg], "", 0);
    expect(withMemo.length).toBeGreaterThan(withoutMemo.length);
  });

  it("12.3: encode_auth_info", () => {
    const pubKey = new Uint8Array(33).fill(0x02);
    const feeCoins = [new Coin("uppyeo", "50000000")];
    const result = encodeAuthInfo(pubKey, 5, feeCoins, 200000);
    expect(result.length).toBeGreaterThan(0);
  });

  it("12.4: encode_sign_doc", () => {
    const bodyBytes = new Uint8Array([1, 2, 3]);
    const authInfoBytes = new Uint8Array([4, 5, 6]);
    const result = encodeSignDoc(
      bodyBytes,
      authInfoBytes,
      "seocheon-testnet-1",
      42,
    );
    expect(result.length).toBeGreaterThan(0);
    // Should contain chain ID string
    const chainIdBytes = new TextEncoder().encode("seocheon-testnet-1");
    const resultStr = Array.from(result)
      .map((b) => String.fromCharCode(b))
      .join("");
    const chainIdStr = Array.from(chainIdBytes)
      .map((b) => String.fromCharCode(b))
      .join("");
    expect(resultStr).toContain(chainIdStr);
  });

  it("12.5: encode_tx_raw", () => {
    const bodyBytes = new Uint8Array([1, 2, 3]);
    const authInfoBytes = new Uint8Array([4, 5, 6]);
    const signature = new Uint8Array(64).fill(0xab);
    const result = encodeTxRaw(bodyBytes, authInfoBytes, signature);
    expect(result.length).toBeGreaterThan(0);
  });

  it("12.6: full_pipeline_activity_submission", () => {
    const msg = new MsgSubmitActivity(
      "seocheon1test",
      "abcd".repeat(16),
      "ipfs://QmTest",
    );
    const bodyBytes = encodeTxBody([msg], "", 0);
    const pubKey = new Uint8Array(33).fill(0x02);
    const feeCoins = [new Coin("uppyeo", "50000000")];
    const authInfoBytes = encodeAuthInfo(pubKey, 0, feeCoins, 200000);
    const signDoc = encodeSignDoc(
      bodyBytes,
      authInfoBytes,
      "seocheon-testnet-1",
      1,
    );
    const signature = new Uint8Array(64).fill(0x00);
    const txRaw = encodeTxRaw(bodyBytes, authInfoBytes, signature);

    expect(signDoc.length).toBeGreaterThan(0);
    expect(txRaw.length).toBeGreaterThan(0);
  });
});

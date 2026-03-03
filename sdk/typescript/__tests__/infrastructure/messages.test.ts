import { describe, it, expect } from "vitest";
import {
  MsgSubmitActivity,
  MsgWithdrawNodeCommission,
  MsgSend,
  Coin,
} from "../../src/infrastructure/messages.js";

describe("messages", () => {
  it("11.1: msg_submit_activity_type_url", () => {
    const msg = new MsgSubmitActivity(
      "seocheon1abc",
      "a".repeat(64),
      "ipfs://QmTest",
    );
    expect(msg.typeUrl()).toBe(
      "/seocheon.activity.v1.MsgSubmitActivity",
    );
  });

  it("11.2: msg_submit_activity_encodes_fields", () => {
    const msg = new MsgSubmitActivity(
      "seocheon1abc",
      "a".repeat(64),
      "ipfs://QmTest",
    );
    const encoded = msg.encode();
    expect(encoded.length).toBeGreaterThan(0);
    // Field 1 tag (0x0a) is present
    expect(encoded[0]).toBe(0x0a);
  });

  it("11.3: msg_withdraw_commission_encodes", () => {
    const msg = new MsgWithdrawNodeCommission("seocheon1operator");
    expect(msg.typeUrl()).toBe(
      "/seocheon.node.v1.MsgWithdrawNodeCommission",
    );
    const encoded = msg.encode();
    expect(encoded.length).toBeGreaterThan(0);
    expect(encoded[0]).toBe(0x0a);
  });

  it("11.4: coin_encodes_fields", () => {
    const coin = new Coin("uppyeo", "50000000");
    const encoded = coin.encode();
    expect(encoded.length).toBeGreaterThan(0);
    expect(encoded[0]).toBe(0x0a);
  });

  it("11.5: msg_send_type_url", () => {
    const msg = new MsgSend("seocheon1from", "seocheon1to", [
      new Coin("uppyeo", "1000"),
    ]);
    expect(msg.typeUrl()).toBe("/cosmos.bank.v1beta1.MsgSend");
  });

  it("11.6: msg_send_encodes_fields", () => {
    const msg = new MsgSend("seocheon1from", "seocheon1to", [
      new Coin("uppyeo", "1000"),
      new Coin("sal", "500"),
    ]);
    const encoded = msg.encode();
    expect(encoded.length).toBeGreaterThan(0);
    expect(encoded[0]).toBe(0x0a);
    // Multiple coin encodings: field 3 tag (0x1a) should appear multiple times
    let count = 0;
    for (let i = 0; i < encoded.length; i++) {
      if (encoded[i] === 0x1a) count++;
    }
    expect(count).toBeGreaterThanOrEqual(2);
  });
});

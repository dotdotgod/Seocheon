import { describe, it, expect } from "vitest";
import {
  defaultGasForMessage,
  calculateFee,
} from "../../src/infrastructure/gas.js";

describe("gas", () => {
  it("13.1: default_gas_submit_activity", () => {
    expect(
      defaultGasForMessage("/seocheon.activity.v1.MsgSubmitActivity"),
    ).toBe(200000);
  });

  it("13.2: default_gas_withdraw_commission", () => {
    expect(
      defaultGasForMessage("/seocheon.node.v1.MsgWithdrawNodeCommission"),
    ).toBe(300000);
  });

  it("13.3: default_gas_send", () => {
    expect(
      defaultGasForMessage("/cosmos.bank.v1beta1.MsgSend"),
    ).toBe(100000);
  });

  it("13.4: default_gas_unknown_fallback", () => {
    expect(defaultGasForMessage("/unknown.v1.MsgUnknown")).toBe(200000);
  });

  it("13.5: calculate_fee_known_values", () => {
    // 200000 * 250 = 50000000 (activity)
    expect(calculateFee(200000, 250)).toBe(50000000);
    // 300000 * 250 = 75000000 (withdraw)
    expect(calculateFee(300000, 250)).toBe(75000000);
    // 100000 * 250 = 25000000 (send)
    expect(calculateFee(100000, 250)).toBe(25000000);
  });

  it("13.6: calculate_fee_zero_inputs", () => {
    expect(calculateFee(200000, 0)).toBe(0);
    expect(calculateFee(0, 250)).toBe(0);
    expect(calculateFee(0, 0)).toBe(0);
  });
});

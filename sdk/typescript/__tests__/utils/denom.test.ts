import { describe, it, expect } from "vitest";
import {
  convertDenom,
  uppyeoToKkot,
  kkotToUppyeo,
  formatKkot,
} from "../../src/utils/denom.js";

describe("denom", () => {
  it("06.1: convert_uppyeo_to_kkot", () => {
    expect(convertDenom(10000000000n, "uppyeo", "kkot")).toBe(1n);
    expect(convertDenom(50000000000n, "uppyeo", "kkot")).toBe(5n);
    expect(convertDenom(5000000000n, "uppyeo", "kkot")).toBe(0n); // truncated
  });

  it("06.2: convert_kkot_to_uppyeo", () => {
    expect(convertDenom(1n, "kkot", "uppyeo")).toBe(10000000000n);
    expect(convertDenom(100n, "kkot", "uppyeo")).toBe(1000000000000n);
  });

  it("06.3: convert_intermediate_denoms", () => {
    // uppyeo -> hon
    expect(convertDenom(100000000n, "uppyeo", "hon")).toBe(1n);
    // uppyeo -> sal
    expect(convertDenom(100n, "uppyeo", "sal")).toBe(1n);
    // uppyeo -> pi
    expect(convertDenom(10000n, "uppyeo", "pi")).toBe(1n);
    // uppyeo -> sum
    expect(convertDenom(1000000n, "uppyeo", "sum")).toBe(1n);
    // hon -> kkot
    expect(convertDenom(100n, "hon", "kkot")).toBe(1n);
  });

  it("06.4: convert_same_denom_returns_same", () => {
    expect(convertDenom(42n, "uppyeo", "uppyeo")).toBe(42n);
    expect(convertDenom(42n, "kkot", "kkot")).toBe(42n);
  });

  it("06.5: convert_invalid_denom_fails", () => {
    expect(() =>
      convertDenom(1n, "invalid" as "uppyeo", "kkot"),
    ).toThrow("Unknown denomination");
  });

  it("06.6: convert_large_genesis_amount", () => {
    // 50,000 KKOT = 500,000,000,000,000 uppyeo (5e14)
    const genesisUppyeo = 500000000000000n;
    expect(convertDenom(genesisUppyeo, "uppyeo", "kkot")).toBe(50000n);
    expect(convertDenom(50000n, "kkot", "uppyeo")).toBe(genesisUppyeo);
  });

  it("06.7: uppyeo_to_kkot_formatting", () => {
    expect(uppyeoToKkot(10000000000n)).toBe("1.0000000000");
    expect(uppyeoToKkot(0n)).toBe("0.0000000000");
    expect(uppyeoToKkot(1n)).toBe("0.0000000001");
    expect(uppyeoToKkot(500000000000000n)).toBe("50000.0000000000");
  });

  it("06.8: format_kkot_from_string", () => {
    expect(formatKkot("10000000000")).toBe("1.0000000000");
    expect(formatKkot("0")).toBe("0.0000000000");
    expect(formatKkot("500000000000000")).toBe("50000.0000000000");
  });

  it("06.9: parse_kkot_invalid_format", () => {
    // Non-numeric string should throw
    expect(() => kkotToUppyeo("abc")).toThrow();
  });

  it("06.10: parse_kkot_roundtrip", () => {
    const original = 12345678901234n;
    const kkotStr = uppyeoToKkot(original);
    const restored = kkotToUppyeo(kkotStr);
    expect(restored).toBe(original);
  });
});

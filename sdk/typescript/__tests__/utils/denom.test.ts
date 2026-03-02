import { describe, it, expect } from "vitest";
import {
  convertDenom,
  uppyeoToKkot,
  kkotToUppyeo,
  formatKkot,
} from "../../src/utils/denom.js";

describe("convertDenom", () => {
  it("uppyeo -> kkot", () => {
    expect(convertDenom(10000000000n, "uppyeo", "kkot")).toBe(1n);
    expect(convertDenom(50000000000n, "uppyeo", "kkot")).toBe(5n);
    expect(convertDenom(5000000000n, "uppyeo", "kkot")).toBe(0n); // truncated
  });

  it("kkot -> uppyeo", () => {
    expect(convertDenom(1n, "kkot", "uppyeo")).toBe(10000000000n);
    expect(convertDenom(100n, "kkot", "uppyeo")).toBe(1000000000000n);
  });

  it("uppyeo -> hon", () => {
    expect(convertDenom(100000000n, "uppyeo", "hon")).toBe(1n);
    expect(convertDenom(300000000n, "uppyeo", "hon")).toBe(3n);
  });

  it("hon -> uppyeo", () => {
    expect(convertDenom(1n, "hon", "uppyeo")).toBe(100000000n);
    expect(convertDenom(500n, "hon", "uppyeo")).toBe(50000000000n);
  });

  it("hon -> kkot", () => {
    expect(convertDenom(100n, "hon", "kkot")).toBe(1n);
    expect(convertDenom(50n, "hon", "kkot")).toBe(0n); // truncated
  });

  it("kkot -> hon", () => {
    expect(convertDenom(1n, "kkot", "hon")).toBe(100n);
    expect(convertDenom(5n, "kkot", "hon")).toBe(500n);
  });

  it("uppyeo -> sal", () => {
    expect(convertDenom(100n, "uppyeo", "sal")).toBe(1n);
    expect(convertDenom(300n, "uppyeo", "sal")).toBe(3n);
  });

  it("uppyeo -> pi", () => {
    expect(convertDenom(10000n, "uppyeo", "pi")).toBe(1n);
    expect(convertDenom(50000n, "uppyeo", "pi")).toBe(5n);
  });

  it("uppyeo -> sum", () => {
    expect(convertDenom(1000000n, "uppyeo", "sum")).toBe(1n);
    expect(convertDenom(5000000n, "uppyeo", "sum")).toBe(5n);
  });

  it("same denomination returns same amount", () => {
    expect(convertDenom(42n, "uppyeo", "uppyeo")).toBe(42n);
    expect(convertDenom(42n, "hon", "hon")).toBe(42n);
    expect(convertDenom(42n, "kkot", "kkot")).toBe(42n);
  });
});

describe("uppyeoToKkot", () => {
  it("converts whole numbers", () => {
    expect(uppyeoToKkot(10000000000n)).toBe("1.0000000000");
    expect(uppyeoToKkot(50000000000n)).toBe("5.0000000000");
    expect(uppyeoToKkot(0n)).toBe("0.0000000000");
  });

  it("converts fractional amounts", () => {
    expect(uppyeoToKkot(15000000000n)).toBe("1.5000000000");
    expect(uppyeoToKkot(1234567890n)).toBe("0.1234567890");
    expect(uppyeoToKkot(1n)).toBe("0.0000000001");
    expect(uppyeoToKkot(9999999999n)).toBe("0.9999999999");
  });

  it("converts large amounts", () => {
    expect(uppyeoToKkot(500000000000000n)).toBe("50000.0000000000");
    expect(uppyeoToKkot(12345678901234n)).toBe("1234.5678901234");
  });
});

describe("kkotToUppyeo", () => {
  it("converts whole numbers", () => {
    expect(kkotToUppyeo("1.0000000000")).toBe(10000000000n);
    expect(kkotToUppyeo("5.0000000000")).toBe(50000000000n);
  });

  it("converts fractional amounts", () => {
    expect(kkotToUppyeo("1.5")).toBe(15000000000n);
    expect(kkotToUppyeo("0.1234567890")).toBe(1234567890n);
    expect(kkotToUppyeo("0.0000000001")).toBe(1n);
  });

  it("converts integer strings", () => {
    expect(kkotToUppyeo("1")).toBe(10000000000n);
    expect(kkotToUppyeo("100")).toBe(1000000000000n);
  });

  it("handles zero", () => {
    expect(kkotToUppyeo("0")).toBe(0n);
    expect(kkotToUppyeo("0.0000000000")).toBe(0n);
  });
});

describe("formatKkot", () => {
  it("formats uppyeo string to kkot string", () => {
    expect(formatKkot("10000000000")).toBe("1.0000000000");
    expect(formatKkot("0")).toBe("0.0000000000");
    expect(formatKkot("1234567890")).toBe("0.1234567890");
    expect(formatKkot("500000000000000")).toBe("50000.0000000000");
  });
});

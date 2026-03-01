import { describe, it, expect } from "vitest";
import {
  isValidActivityHash,
  verifyActivityHash,
  isValidContentUri,
} from "../../src/utils/hash.js";

describe("isValidActivityHash", () => {
  it("accepts valid 64-character hex hash", () => {
    const validHash =
      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";
    expect(isValidActivityHash(validHash)).toBe(true);
  });

  it("accepts uppercase hex", () => {
    const hash =
      "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2";
    expect(isValidActivityHash(hash)).toBe(true);
  });

  it("accepts mixed case hex", () => {
    const hash =
      "a1B2c3D4e5F6a1B2c3D4e5F6a1B2c3D4e5F6a1B2c3D4e5F6a1B2c3D4e5F6a1B2";
    expect(isValidActivityHash(hash)).toBe(true);
  });

  it("rejects too short hash", () => {
    expect(isValidActivityHash("a1b2c3d4")).toBe(false);
  });

  it("rejects too long hash", () => {
    const tooLong =
      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2ff";
    expect(isValidActivityHash(tooLong)).toBe(false);
  });

  it("rejects non-hex characters", () => {
    const nonHex =
      "g1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";
    expect(isValidActivityHash(nonHex)).toBe(false);
  });

  it("rejects empty string", () => {
    expect(isValidActivityHash("")).toBe(false);
  });

  it("rejects hash with spaces", () => {
    const withSpaces =
      "a1b2c3d4 5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";
    expect(isValidActivityHash(withSpaces)).toBe(false);
  });
});

describe("verifyActivityHash", () => {
  it("is an alias for isValidActivityHash", () => {
    const validHash =
      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";
    expect(verifyActivityHash(validHash)).toBe(true);
    expect(verifyActivityHash("short")).toBe(false);
  });
});

describe("isValidContentUri", () => {
  it("accepts non-empty strings", () => {
    expect(isValidContentUri("https://example.com/report.json")).toBe(true);
    expect(isValidContentUri("ipfs://QmHash")).toBe(true);
    expect(isValidContentUri("a")).toBe(true);
  });

  it("rejects empty string", () => {
    expect(isValidContentUri("")).toBe(false);
  });
});

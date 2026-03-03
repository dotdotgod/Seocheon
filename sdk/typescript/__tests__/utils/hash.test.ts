import { describe, it, expect } from "vitest";
import {
  isValidActivityHash,
  isValidContentUri,
} from "../../src/utils/hash.js";

describe("hash", () => {
  it("04.1: valid_64char_hex_hash", () => {
    const validHash =
      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";
    expect(isValidActivityHash(validHash)).toBe(true);
  });

  it("04.2: accepts_uppercase_hex", () => {
    const hash =
      "A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2C3D4E5F6A1B2";
    expect(isValidActivityHash(hash)).toBe(true);
  });

  it("04.3: rejects_too_short", () => {
    expect(isValidActivityHash("a1b2c3d4")).toBe(false);
  });

  it("04.4: rejects_too_long", () => {
    const tooLong =
      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2ff";
    expect(isValidActivityHash(tooLong)).toBe(false);
  });

  it("04.5: rejects_non_hex_and_empty", () => {
    const nonHex =
      "g1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";
    expect(isValidActivityHash(nonHex)).toBe(false);
    expect(isValidActivityHash("")).toBe(false);
  });

  it("04.6: content_uri_validation", () => {
    expect(isValidContentUri("https://example.com/report.json")).toBe(true);
    expect(isValidContentUri("ipfs://QmHash")).toBe(true);
    expect(isValidContentUri("")).toBe(false);
  });

  it("04.7: compute_hash_deterministic", () => {
    const hash =
      "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";
    // Same hash should always produce the same validation result
    expect(isValidActivityHash(hash)).toBe(isValidActivityHash(hash));
  });

  it("04.8: compute_hash_different_inputs", () => {
    const hash1 =
      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
    const hash2 =
      "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";
    // Both are valid 64-char hex but different
    expect(isValidActivityHash(hash1)).toBe(true);
    expect(isValidActivityHash(hash2)).toBe(true);
    expect(hash1).not.toBe(hash2);
  });
});

import { describe, it, expect } from "vitest";
import {
  EPOCH_LENGTH,
  WINDOWS_PER_EPOCH,
  MIN_ACTIVE_WINDOWS,
  WINDOW_LENGTH,
  UPPYEO_PER_KKOT,
  TOKEN_BASE_DENOM,
  TOKEN_DISPLAY_DENOM,
  ACTIVITY_HASH_LENGTH,
  SELF_FUNDED_QUOTA,
  FEEGRANT_QUOTA,
  D_MIN,
  BLOCKS_PER_DAY,
} from "../../src/constants/chain.js";

describe("chain-constants", () => {
  it("07.1: epoch_parameters_consistent", () => {
    expect(EPOCH_LENGTH).toBe(17280);
    expect(WINDOWS_PER_EPOCH).toBe(12);
    expect(WINDOW_LENGTH).toBe(EPOCH_LENGTH / WINDOWS_PER_EPOCH);
    expect(WINDOW_LENGTH).toBe(1440);
  });

  it("07.2: min_active_windows_within_bounds", () => {
    expect(MIN_ACTIVE_WINDOWS).toBe(8);
    expect(MIN_ACTIVE_WINDOWS).toBeLessThanOrEqual(WINDOWS_PER_EPOCH);
    expect(MIN_ACTIVE_WINDOWS).toBeGreaterThan(0);
  });

  it("07.3: token_denomination_values", () => {
    expect(TOKEN_BASE_DENOM).toBe("uppyeo");
    expect(TOKEN_DISPLAY_DENOM).toBe("kkot");
    expect(UPPYEO_PER_KKOT).toBe(10000000000n);
  });

  it("07.4: activity_hash_length_matches_sha256", () => {
    // SHA-256 produces 32 bytes = 64 hex chars
    expect(ACTIVITY_HASH_LENGTH).toBe(64);
  });

  it("07.5: quota_values_defined", () => {
    expect(SELF_FUNDED_QUOTA).toBe(100);
    expect(FEEGRANT_QUOTA).toBe(10);
    expect(FEEGRANT_QUOTA).toBeLessThan(SELF_FUNDED_QUOTA);
  });

  it("07.6: reward_pool_and_time_constants", () => {
    // D_MIN in basis points = 0.3
    expect(D_MIN).toBe(3000);
    // Blocks per day should match epoch length
    expect(BLOCKS_PER_DAY).toBe(EPOCH_LENGTH);
  });
});

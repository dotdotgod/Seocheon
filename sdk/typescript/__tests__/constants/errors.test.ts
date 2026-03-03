import { describe, it, expect } from "vitest";
import {
  ERR_NODE_NOT_FOUND,
  ERR_SUBMITTER_NOT_REGISTERED,
  ERR_QUOTA_EXCEEDED,
  ERR_INVALID_ACTIVITY_HASH,
  ERR_DUPLICATE_ACTIVITY_HASH,
  ERR_NOT_CONNECTED,
  ERR_BROADCAST_FAILED,
  ERR_TX_TIMEOUT,
  ERR_TX_NOT_FOUND,
  ERR_SIGNING_FAILED,
  ERR_INVALID_CONFIG,
  ERR_QUERY_FAILED,
} from "../../src/constants/errors.js";

describe("error-constants", () => {
  it("08.1: node_errors_in_1100_range", () => {
    expect(ERR_NODE_NOT_FOUND).toBe(1101);
    expect(ERR_NODE_NOT_FOUND).toBeGreaterThanOrEqual(1100);
    expect(ERR_NODE_NOT_FOUND).toBeLessThan(1200);
  });

  it("08.2: activity_errors_in_1200_range", () => {
    expect(ERR_SUBMITTER_NOT_REGISTERED).toBe(1200);
    expect(ERR_QUOTA_EXCEEDED).toBe(1203);
    expect(ERR_INVALID_ACTIVITY_HASH).toBe(1204);
    expect(ERR_DUPLICATE_ACTIVITY_HASH).toBe(1202);
  });

  it("08.3: sdk_errors_in_9000_range", () => {
    expect(ERR_NOT_CONNECTED).toBe(9000);
    expect(ERR_BROADCAST_FAILED).toBe(9001);
    expect(ERR_TX_TIMEOUT).toBe(9002);
    expect(ERR_TX_NOT_FOUND).toBe(9003);
    expect(ERR_SIGNING_FAILED).toBe(9004);
    expect(ERR_INVALID_CONFIG).toBe(9005);
    expect(ERR_QUERY_FAILED).toBe(9006);
  });

  it("08.4: no_code_overlap_between_ranges", () => {
    // Node errors (1100s) should not overlap with activity errors (1200s)
    const nodeErrors = [ERR_NODE_NOT_FOUND];
    const activityErrors = [
      ERR_SUBMITTER_NOT_REGISTERED,
      ERR_QUOTA_EXCEEDED,
      ERR_INVALID_ACTIVITY_HASH,
      ERR_DUPLICATE_ACTIVITY_HASH,
    ];
    const sdkErrors = [
      ERR_NOT_CONNECTED,
      ERR_BROADCAST_FAILED,
      ERR_TX_TIMEOUT,
      ERR_TX_NOT_FOUND,
      ERR_SIGNING_FAILED,
      ERR_INVALID_CONFIG,
      ERR_QUERY_FAILED,
    ];
    const allCodes = [...nodeErrors, ...activityErrors, ...sdkErrors];
    const uniqueCodes = new Set(allCodes);
    expect(uniqueCodes.size).toBe(allCodes.length);
  });

  it("08.5: all_error_codes_are_positive_integers", () => {
    const codes = [
      ERR_NODE_NOT_FOUND,
      ERR_SUBMITTER_NOT_REGISTERED,
      ERR_QUOTA_EXCEEDED,
      ERR_NOT_CONNECTED,
      ERR_BROADCAST_FAILED,
      ERR_INVALID_CONFIG,
    ];
    for (const code of codes) {
      expect(Number.isInteger(code)).toBe(true);
      expect(code).toBeGreaterThan(0);
    }
  });
});

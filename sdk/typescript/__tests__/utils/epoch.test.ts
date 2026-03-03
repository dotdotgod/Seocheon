import { describe, it, expect } from "vitest";
import {
  computeEpoch,
  computeWindow,
  epochStartBlock,
  epochEndBlock,
  windowStartBlock,
  windowEndBlock,
} from "../../src/utils/epoch.js";
import {
  EPOCH_LENGTH,
  WINDOW_LENGTH,
} from "../../src/constants/chain.js";

describe("epoch", () => {
  it("05.1: compute_epoch_block1_is_epoch0", () => {
    expect(computeEpoch(1)).toBe(0);
  });

  it("05.2: compute_epoch_last_block_of_epoch0", () => {
    expect(computeEpoch(EPOCH_LENGTH)).toBe(0);
  });

  it("05.3: compute_epoch_first_block_of_epoch1", () => {
    expect(computeEpoch(EPOCH_LENGTH + 1)).toBe(1);
  });

  it("05.4: compute_epoch_edge_zero_height", () => {
    // height 0 is below valid range (blocks start at 1)
    // Math.floor((0-1)/17280) = -1
    expect(computeEpoch(0)).toBe(-1);
  });

  it("05.5: compute_window_block1_is_window0", () => {
    expect(computeWindow(1)).toBe(0);
  });

  it("05.6: compute_window_boundary", () => {
    // block 1440 is still window 0
    expect(computeWindow(1440)).toBe(0);
    // block 1441 is window 1
    expect(computeWindow(1441)).toBe(1);
  });

  it("05.7: compute_window_last_block_is_window11", () => {
    expect(computeWindow(EPOCH_LENGTH)).toBe(11);
  });

  it("05.8: epoch_start_block", () => {
    expect(epochStartBlock(0)).toBe(1);
    expect(epochStartBlock(1)).toBe(EPOCH_LENGTH + 1);
  });

  it("05.9: epoch_end_block", () => {
    expect(epochEndBlock(0)).toBe(EPOCH_LENGTH);
    expect(epochEndBlock(1)).toBe(2 * EPOCH_LENGTH);
  });

  it("05.10: window_start_end_block", () => {
    expect(windowStartBlock(1, 0)).toBe(1);
    expect(windowEndBlock(1, 0)).toBe(WINDOW_LENGTH);
    expect(windowStartBlock(1, 1)).toBe(WINDOW_LENGTH + 1);
    expect(windowEndBlock(1, 11)).toBe(EPOCH_LENGTH);
  });

  it("05.11: compute_epoch_edge_zero_length", () => {
    // zero epoch length causes division by zero edge case
    // Math.floor((1-1)/0) = NaN or Infinity
    const result = computeEpoch(1, 0);
    expect(result).toBeNaN();
  });
});

import { describe, it, expect } from "vitest";
import {
  computeEpoch,
  computeWindow,
  epochStartBlock,
  epochEndBlock,
  windowStartBlock,
  windowEndBlock,
  formatProgress,
} from "../../src/utils/epoch.js";
import {
  EPOCH_LENGTH,
  WINDOWS_PER_EPOCH,
  WINDOW_LENGTH,
} from "../../src/constants/chain.js";

describe("computeEpoch", () => {
  it("block 1 is epoch 0", () => {
    expect(computeEpoch(1)).toBe(0);
  });

  it("last block of epoch 0", () => {
    expect(computeEpoch(EPOCH_LENGTH)).toBe(0);
  });

  it("first block of epoch 1", () => {
    expect(computeEpoch(EPOCH_LENGTH + 1)).toBe(1);
  });

  it("block 34560 is epoch 1", () => {
    expect(computeEpoch(34560)).toBe(1);
  });

  it("block 34561 is epoch 2", () => {
    expect(computeEpoch(34561)).toBe(2);
  });

  it("supports custom epoch length", () => {
    expect(computeEpoch(101, 100)).toBe(1);
    expect(computeEpoch(200, 100)).toBe(1);
    expect(computeEpoch(201, 100)).toBe(2);
  });
});

describe("computeWindow", () => {
  it("block 1 is window 0", () => {
    expect(computeWindow(1)).toBe(0);
  });

  it("block 1440 is still window 0", () => {
    expect(computeWindow(1440)).toBe(0);
  });

  it("block 1441 is window 1", () => {
    expect(computeWindow(1441)).toBe(1);
  });

  it("block 2880 is window 1", () => {
    expect(computeWindow(2880)).toBe(1);
  });

  it("block 2881 is window 2", () => {
    expect(computeWindow(2881)).toBe(2);
  });

  it("last block of epoch is window 11", () => {
    expect(computeWindow(EPOCH_LENGTH)).toBe(11);
  });

  it("first block of epoch 1 is window 0", () => {
    expect(computeWindow(EPOCH_LENGTH + 1)).toBe(0);
  });

  it("supports custom parameters", () => {
    expect(computeWindow(1, 100, 10)).toBe(0);
    expect(computeWindow(11, 100, 10)).toBe(1);
    expect(computeWindow(20, 100, 10)).toBe(1);
    expect(computeWindow(21, 100, 10)).toBe(2);
  });
});

describe("epochStartBlock", () => {
  it("epoch 0 starts at block 1", () => {
    expect(epochStartBlock(0)).toBe(1);
  });

  it("epoch 1 starts at 17281", () => {
    expect(epochStartBlock(1)).toBe(EPOCH_LENGTH + 1);
  });

  it("epoch 10 starts at 172801", () => {
    expect(epochStartBlock(10)).toBe(10 * EPOCH_LENGTH + 1);
  });
});

describe("epochEndBlock", () => {
  it("epoch 0 ends at 17280", () => {
    expect(epochEndBlock(0)).toBe(EPOCH_LENGTH);
  });

  it("epoch 1 ends at 34560", () => {
    expect(epochEndBlock(1)).toBe(2 * EPOCH_LENGTH);
  });
});

describe("windowStartBlock / windowEndBlock", () => {
  it("window 0 of epoch 0 starts at 1", () => {
    expect(windowStartBlock(1, 0)).toBe(1);
  });

  it("window 0 of epoch 0 ends at 1440", () => {
    expect(windowEndBlock(1, 0)).toBe(1440);
  });

  it("window 1 of epoch 0 starts at 1441", () => {
    expect(windowStartBlock(1, 1)).toBe(1441);
  });

  it("window 1 of epoch 0 ends at 2880", () => {
    expect(windowEndBlock(1, 1)).toBe(2880);
  });

  it("window 11 (last) ends at epoch end", () => {
    expect(windowEndBlock(1, 11)).toBe(EPOCH_LENGTH);
  });
});

describe("formatProgress", () => {
  it("formats progress string", () => {
    expect(formatProgress(720, 1440)).toBe("720/1440");
    expect(formatProgress(0, 17280)).toBe("0/17280");
    expect(formatProgress(17280, 17280)).toBe("17280/17280");
  });
});

import { EPOCH_LENGTH, WINDOWS_PER_EPOCH, WINDOW_LENGTH } from "../constants/chain.js";

export function computeEpoch(
  blockHeight: number,
  epochLength: number = EPOCH_LENGTH,
): number {
  return Math.floor((blockHeight - 1) / epochLength);
}

export function computeWindow(
  blockHeight: number,
  epochLength: number = EPOCH_LENGTH,
  windowsPerEpoch: number = WINDOWS_PER_EPOCH,
): number {
  const windowLength = Math.floor(epochLength / windowsPerEpoch);
  return Math.floor(((blockHeight - 1) % epochLength) / windowLength);
}

export function epochStartBlock(epochNumber: number, epochLength: number = EPOCH_LENGTH): number {
  return epochNumber * epochLength + 1;
}

export function epochEndBlock(epochNumber: number, epochLength: number = EPOCH_LENGTH): number {
  return (epochNumber + 1) * epochLength;
}

export function windowStartBlock(
  epochStartBlockHeight: number,
  windowIndex: number,
  windowLength: number = WINDOW_LENGTH,
): number {
  return epochStartBlockHeight + windowIndex * windowLength;
}

export function windowEndBlock(
  epochStartBlockHeight: number,
  windowIndex: number,
  windowLength: number = WINDOW_LENGTH,
): number {
  return epochStartBlockHeight + windowIndex * windowLength + windowLength - 1;
}

export function formatProgress(elapsed: number, total: number): string {
  return `${elapsed}/${total}`;
}

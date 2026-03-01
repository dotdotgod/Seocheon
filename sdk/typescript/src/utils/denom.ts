import {
  UPPYEO_PER_SAL,
  UPPYEO_PER_PI,
  UPPYEO_PER_SUM,
  UPPYEO_PER_HON,
  UPPYEO_PER_KKOT,
  TOKEN_BASE_DENOM,
  TOKEN_SAL_DENOM,
  TOKEN_PI_DENOM,
  TOKEN_SUM_DENOM,
  TOKEN_HON_DENOM,
  TOKEN_DISPLAY_DENOM,
} from "../constants/chain.js";

export type Denom =
  | typeof TOKEN_BASE_DENOM
  | typeof TOKEN_SAL_DENOM
  | typeof TOKEN_PI_DENOM
  | typeof TOKEN_SUM_DENOM
  | typeof TOKEN_HON_DENOM
  | typeof TOKEN_DISPLAY_DENOM;

export function convertDenom(amount: bigint, from: Denom, to: Denom): bigint {
  if (from === to) return amount;

  const uppyeoAmount = toUppyeo(amount, from);
  return fromUppyeo(uppyeoAmount, to);
}

function toUppyeo(amount: bigint, denom: Denom): bigint {
  switch (denom) {
    case TOKEN_BASE_DENOM:
      return amount;
    case TOKEN_SAL_DENOM:
      return amount * UPPYEO_PER_SAL;
    case TOKEN_PI_DENOM:
      return amount * UPPYEO_PER_PI;
    case TOKEN_SUM_DENOM:
      return amount * UPPYEO_PER_SUM;
    case TOKEN_HON_DENOM:
      return amount * UPPYEO_PER_HON;
    case TOKEN_DISPLAY_DENOM:
      return amount * UPPYEO_PER_KKOT;
    default:
      throw new Error(`Unknown denomination: ${denom}`);
  }
}

function fromUppyeo(uppyeo: bigint, denom: Denom): bigint {
  switch (denom) {
    case TOKEN_BASE_DENOM:
      return uppyeo;
    case TOKEN_SAL_DENOM:
      return uppyeo / UPPYEO_PER_SAL;
    case TOKEN_PI_DENOM:
      return uppyeo / UPPYEO_PER_PI;
    case TOKEN_SUM_DENOM:
      return uppyeo / UPPYEO_PER_SUM;
    case TOKEN_HON_DENOM:
      return uppyeo / UPPYEO_PER_HON;
    case TOKEN_DISPLAY_DENOM:
      return uppyeo / UPPYEO_PER_KKOT;
    default:
      throw new Error(`Unknown denomination: ${denom}`);
  }
}

export function uppyeoToKkot(uppyeo: bigint): string {
  const integerPart = uppyeo / UPPYEO_PER_KKOT;
  const decimalPart = uppyeo % UPPYEO_PER_KKOT;
  const decimalStr = decimalPart.toString().padStart(10, "0");
  return `${integerPart}.${decimalStr}`;
}

export function kkotToUppyeo(kkot: string): bigint {
  const parts = kkot.split(".");
  const integerPart = BigInt(parts[0] ?? "0");
  const decimalStr = (parts[1] ?? "").padEnd(10, "0").slice(0, 10);
  const decimalPart = BigInt(decimalStr);
  return integerPart * UPPYEO_PER_KKOT + decimalPart;
}

export function formatKkot(uppyeoStr: string): string {
  return uppyeoToKkot(BigInt(uppyeoStr));
}

export {
  UPPYEO_PER_SAL,
  UPPYEO_PER_PI,
  UPPYEO_PER_SUM,
  UPPYEO_PER_HON,
  UPPYEO_PER_KKOT,
};

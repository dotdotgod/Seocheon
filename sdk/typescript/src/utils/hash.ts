import { ACTIVITY_HASH_LENGTH } from "../constants/chain.js";

const HEX_REGEX = /^[0-9a-fA-F]+$/;

export function isValidActivityHash(hash: string): boolean {
  return hash.length === ACTIVITY_HASH_LENGTH && HEX_REGEX.test(hash);
}

export function verifyActivityHash(hash: string): boolean {
  return isValidActivityHash(hash);
}

export function isValidContentUri(uri: string): boolean {
  return uri.length > 0;
}

// TX envelope encoding: TxBody, AuthInfo, SignDoc, TxRaw.
// Matches the Go SDK's envelope.go byte-for-byte.

import {
  encodeVarint,
  encodeFieldVarint,
  encodeFieldBytes,
  encodeFieldString,
  concatBytes,
} from "./protobuf.js";
import type { MessageEncoder } from "./messages.js";
import { Coin } from "./messages.js";

// google.protobuf.Any encoding
// Fields: type_url(1, string), value(2, bytes)
function encodeAny(typeURL: string, value: Uint8Array): Uint8Array {
  return concatBytes(
    encodeFieldString(1, typeURL),
    encodeFieldBytes(2, value),
  );
}

// cosmos.crypto.secp256k1.PubKey encoding
// type_url: "/cosmos.crypto.secp256k1.PubKey"
// Fields: key(1, bytes)
function encodePubKeyAny(pubKey: Uint8Array): Uint8Array {
  const innerMsg = encodeFieldBytes(1, pubKey);
  return encodeAny("/cosmos.crypto.secp256k1.PubKey", innerMsg);
}

// ModeInfo for SIGN_MODE_DIRECT
// ModeInfo { single { mode: SIGN_MODE_DIRECT (1) } }
function encodeModeInfoDirect(): Uint8Array {
  const modeField = encodeFieldVarint(1, 1); // SIGN_MODE_DIRECT = 1
  const single = encodeFieldBytes(1, modeField);
  return single;
}

// SignerInfo encoding
// Fields: public_key(1, Any), mode_info(2, ModeInfo), sequence(3, uint64)
function encodeSignerInfo(pubKey: Uint8Array, sequence: number): Uint8Array {
  const pubKeyAny = encodePubKeyAny(pubKey);
  const modeInfo = encodeModeInfoDirect();
  return concatBytes(
    encodeFieldBytes(1, pubKeyAny),
    encodeFieldBytes(2, modeInfo),
    encodeFieldVarint(3, sequence),
  );
}

// Fee encoding
// Fields: amount(1, repeated Coin), gas_limit(2, uint64)
function encodeFee(coins: Coin[], gasLimit: number): Uint8Array {
  const parts: Uint8Array[] = [];
  for (const coin of coins) {
    parts.push(encodeFieldBytes(1, coin.encode()));
  }
  parts.push(encodeFieldVarint(2, gasLimit));
  return concatBytes(...parts);
}

/**
 * Encodes a TxBody.
 * Fields: messages(1, repeated Any), memo(2, string), timeout_height(3, uint64)
 */
export function encodeTxBody(
  messages: MessageEncoder[],
  memo: string,
  timeoutHeight: number,
): Uint8Array {
  const parts: Uint8Array[] = [];
  for (const msg of messages) {
    const anyBytes = encodeAny(msg.typeUrl(), msg.encode());
    parts.push(encodeFieldBytes(1, anyBytes));
  }
  if (memo !== "") {
    parts.push(encodeFieldString(2, memo));
  }
  if (timeoutHeight > 0) {
    parts.push(encodeFieldVarint(3, timeoutHeight));
  }
  return concatBytes(...parts);
}

/**
 * Encodes an AuthInfo.
 * Fields: signer_infos(1, repeated SignerInfo), fee(2, Fee)
 */
export function encodeAuthInfo(
  pubKey: Uint8Array,
  sequence: number,
  feeCoins: Coin[],
  gasLimit: number,
): Uint8Array {
  const signerInfo = encodeSignerInfo(pubKey, sequence);
  const fee = encodeFee(feeCoins, gasLimit);
  return concatBytes(
    encodeFieldBytes(1, signerInfo),
    encodeFieldBytes(2, fee),
  );
}

/**
 * Encodes a SignDoc for SIGN_MODE_DIRECT.
 * Fields: body_bytes(1, bytes), auth_info_bytes(2, bytes), chain_id(3, string), account_number(4, uint64)
 */
export function encodeSignDoc(
  bodyBytes: Uint8Array,
  authInfoBytes: Uint8Array,
  chainId: string,
  accountNumber: number,
): Uint8Array {
  return concatBytes(
    encodeFieldBytes(1, bodyBytes),
    encodeFieldBytes(2, authInfoBytes),
    encodeFieldString(3, chainId),
    encodeFieldVarint(4, accountNumber),
  );
}

/**
 * Encodes a TxRaw for broadcast.
 * Fields: body_bytes(1, bytes), auth_info_bytes(2, bytes), signatures(3, repeated bytes)
 */
export function encodeTxRaw(
  bodyBytes: Uint8Array,
  authInfoBytes: Uint8Array,
  ...signatures: Uint8Array[]
): Uint8Array {
  const parts: Uint8Array[] = [
    encodeFieldBytes(1, bodyBytes),
    encodeFieldBytes(2, authInfoBytes),
  ];
  for (const sig of signatures) {
    parts.push(encodeFieldBytes(3, sig));
  }
  return concatBytes(...parts);
}

// Minimal protobuf wire format encoder.
// Only supports the types needed for Cosmos TX encoding:
// - varint (wire type 0)
// - length-delimited (wire type 2)

/**
 * Encodes a non-negative integer as a protobuf varint.
 * Supports values up to Number.MAX_SAFE_INTEGER (2^53 - 1).
 */
export function encodeVarint(v: number): Uint8Array {
  if (v === 0) return new Uint8Array([0]);
  const buf: number[] = [];
  while (v >= 0x80) {
    buf.push((v & 0x7f) | 0x80);
    v = Math.floor(v / 128); // avoid bitwise ops for values > 2^31
  }
  buf.push(v & 0x7f);
  return new Uint8Array(buf);
}

/**
 * Encodes a varint field (wire type 0).
 * Returns empty if value is 0 (protobuf default: omit zero).
 */
export function encodeFieldVarint(
  fieldNumber: number,
  value: number,
): Uint8Array {
  if (value === 0) return new Uint8Array(0);
  const tag = encodeVarint((fieldNumber << 3) | 0);
  const val = encodeVarint(value);
  return concatBytes(tag, val);
}

/**
 * Encodes a length-delimited field (wire type 2).
 * Returns empty if data is empty (protobuf default: omit empty).
 */
export function encodeFieldBytes(
  fieldNumber: number,
  data: Uint8Array,
): Uint8Array {
  if (data.length === 0) return new Uint8Array(0);
  const tag = encodeVarint((fieldNumber << 3) | 2);
  const length = encodeVarint(data.length);
  return concatBytes(tag, length, data);
}

/**
 * Encodes a string field (wire type 2).
 * Returns empty if value is empty (protobuf default: omit empty).
 */
export function encodeFieldString(
  fieldNumber: number,
  value: string,
): Uint8Array {
  if (value === "") return new Uint8Array(0);
  return encodeFieldBytes(fieldNumber, new TextEncoder().encode(value));
}

/**
 * Concatenates multiple Uint8Arrays, skipping nulls and undefineds.
 */
export function concatBytes(...parts: (Uint8Array | null | undefined)[]): Uint8Array {
  let size = 0;
  for (const p of parts) {
    if (p) size += p.length;
  }
  const result = new Uint8Array(size);
  let offset = 0;
  for (const p of parts) {
    if (p && p.length > 0) {
      result.set(p, offset);
      offset += p.length;
    }
  }
  return result;
}

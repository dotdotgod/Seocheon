import { describe, it, expect } from "vitest";
import {
  encodeVarint,
  encodeFieldVarint,
  encodeFieldString,
  concatBytes,
} from "../../src/infrastructure/protobuf.js";

describe("protobuf", () => {
  it("10.1: encode_varint_known_values", () => {
    expect(encodeVarint(0)).toEqual(new Uint8Array([0x00]));
    expect(encodeVarint(1)).toEqual(new Uint8Array([0x01]));
    expect(encodeVarint(128)).toEqual(new Uint8Array([0x80, 0x01]));
    expect(encodeVarint(300)).toEqual(new Uint8Array([0xac, 0x02]));
  });

  it("10.2: encode_field_varint_zero_empty", () => {
    expect(encodeFieldVarint(1, 0).length).toBe(0);
  });

  it("10.3: encode_field_varint_nonzero", () => {
    // field 1, value 1 (SIGN_MODE_DIRECT): tag (1<<3|0)=0x08, value 0x01
    expect(encodeFieldVarint(1, 1)).toEqual(new Uint8Array([0x08, 0x01]));
    // field 3, value 5: tag (3<<3|0)=0x18, value 0x05
    expect(encodeFieldVarint(3, 5)).toEqual(new Uint8Array([0x18, 0x05]));
  });

  it("10.4: encode_field_string_empty", () => {
    expect(encodeFieldString(1, "").length).toBe(0);
  });

  it("10.5: encode_field_string_data", () => {
    const result = encodeFieldString(1, "abc");
    // tag: 0x0a, length: 3, "abc" bytes
    expect(result).toEqual(new Uint8Array([0x0a, 0x03, 0x61, 0x62, 0x63]));
  });

  it("10.6: concat_bytes_nil_empty", () => {
    // empty input
    expect(concatBytes()).toEqual(new Uint8Array(0));
    // null/undefined skipped
    expect(concatBytes(null, undefined, null)).toEqual(new Uint8Array(0));
    // single array passthrough
    expect(concatBytes(new Uint8Array([1, 2]))).toEqual(
      new Uint8Array([1, 2]),
    );
    // mixed concatenation
    const a = new Uint8Array([1]);
    const b = new Uint8Array([2]);
    expect(concatBytes(a, null, b, undefined)).toEqual(
      new Uint8Array([1, 2]),
    );
  });
});

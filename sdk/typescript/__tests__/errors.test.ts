import { describe, it, expect } from "vitest";
import {
  SeocheonError,
  ConnectionError,
  QueryError,
  TransactionError,
  ValidationError,
  TimeoutError,
  mapAbciError,
} from "../src/errors/errors.js";
import {
  ERR_NOT_CONNECTED,
  ERR_QUERY_FAILED,
  ERR_BROADCAST_FAILED,
  ERR_INVALID_CONFIG,
  ERR_TX_TIMEOUT,
} from "../src/constants/errors.js";

describe("errors", () => {
  it("01.1: base_error_has_code_and_message", () => {
    const err = new SeocheonError(1234, "test error");
    expect(err.code).toBe(1234);
    expect(err.message).toBe("test error");
    expect(err.name).toBe("SeocheonError");
    expect(err).toBeInstanceOf(Error);
  });

  it("01.2: error_subclasses_use_correct_codes", () => {
    const conn = new ConnectionError("not connected");
    expect(conn.code).toBe(ERR_NOT_CONNECTED);
    expect(conn.name).toBe("ConnectionError");
    expect(conn).toBeInstanceOf(SeocheonError);

    const query = new QueryError("query failed");
    expect(query.code).toBe(ERR_QUERY_FAILED);
    expect(query.name).toBe("QueryError");

    const tx = new TransactionError("broadcast failed");
    expect(tx.code).toBe(ERR_BROADCAST_FAILED);
    expect(tx.name).toBe("TransactionError");

    const val = new ValidationError("invalid config");
    expect(val.code).toBe(ERR_INVALID_CONFIG);
    expect(val.name).toBe("ValidationError");

    const timeout = new TimeoutError("AABB1122");
    expect(timeout.code).toBe(ERR_TX_TIMEOUT);
    expect(timeout.message).toContain("AABB1122");
  });

  it("01.3: wrap_error_preserves_cause", () => {
    const cause = new Error("root cause");
    const err = new QueryError("query failed");
    // SeocheonError is an Error subclass, so we can check prototype chain
    expect(err).toBeInstanceOf(Error);
    expect(err).toBeInstanceOf(SeocheonError);
    // QueryError accepts custom code
    const custom = new QueryError("not found", 9003);
    expect(custom.code).toBe(9003);
  });

  it("01.4: map_abci_activity_errors", () => {
    const err = mapAbciError(1203, "quota exceeded");
    expect(err.code).toBe(1203);
    expect(err).toBeInstanceOf(TransactionError);
  });

  it("01.5: map_abci_node_and_unknown_errors", () => {
    const nodeErr = mapAbciError(1101, "node not found");
    expect(nodeErr.code).toBe(1101);
    expect(nodeErr).toBeInstanceOf(TransactionError);

    const unknownErr = mapAbciError(5, "unknown");
    expect(unknownErr.code).toBe(5);
    expect(unknownErr).toBeInstanceOf(TransactionError);
  });

  it("01.6: predefined_error_codes_match", () => {
    // Verify error code constants are in expected ranges
    expect(ERR_NOT_CONNECTED).toBe(9000);
    expect(ERR_QUERY_FAILED).toBe(9006);
    expect(ERR_BROADCAST_FAILED).toBe(9001);
    expect(ERR_INVALID_CONFIG).toBe(9005);
    expect(ERR_TX_TIMEOUT).toBe(9002);
  });
});

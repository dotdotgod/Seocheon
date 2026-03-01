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

describe("SeocheonError", () => {
  it("has code and message", () => {
    const err = new SeocheonError(1234, "test error");
    expect(err.code).toBe(1234);
    expect(err.message).toBe("test error");
    expect(err.name).toBe("SeocheonError");
    expect(err).toBeInstanceOf(Error);
  });
});

describe("ConnectionError", () => {
  it("uses ERR_NOT_CONNECTED code", () => {
    const err = new ConnectionError("not connected");
    expect(err.code).toBe(ERR_NOT_CONNECTED);
    expect(err.name).toBe("ConnectionError");
    expect(err).toBeInstanceOf(SeocheonError);
  });
});

describe("QueryError", () => {
  it("uses ERR_QUERY_FAILED by default", () => {
    const err = new QueryError("query failed");
    expect(err.code).toBe(ERR_QUERY_FAILED);
    expect(err.name).toBe("QueryError");
  });

  it("accepts custom code", () => {
    const err = new QueryError("not found", 9003);
    expect(err.code).toBe(9003);
  });
});

describe("TransactionError", () => {
  it("uses ERR_BROADCAST_FAILED by default", () => {
    const err = new TransactionError("broadcast failed");
    expect(err.code).toBe(ERR_BROADCAST_FAILED);
    expect(err.name).toBe("TransactionError");
  });
});

describe("ValidationError", () => {
  it("uses ERR_INVALID_CONFIG by default", () => {
    const err = new ValidationError("invalid config");
    expect(err.code).toBe(ERR_INVALID_CONFIG);
    expect(err.name).toBe("ValidationError");
  });
});

describe("TimeoutError", () => {
  it("includes tx hash in message", () => {
    const err = new TimeoutError("AABB1122");
    expect(err.code).toBe(ERR_TX_TIMEOUT);
    expect(err.message).toContain("AABB1122");
    expect(err.name).toBe("TimeoutError");
  });
});

describe("mapAbciError", () => {
  it("maps x/activity errors", () => {
    const err = mapAbciError(1203, "quota exceeded");
    expect(err.code).toBe(1203);
    expect(err).toBeInstanceOf(TransactionError);
  });

  it("maps x/node errors", () => {
    const err = mapAbciError(1101, "node not found");
    expect(err.code).toBe(1101);
    expect(err).toBeInstanceOf(TransactionError);
  });

  it("maps unknown errors", () => {
    const err = mapAbciError(5, "unknown");
    expect(err.code).toBe(5);
    expect(err).toBeInstanceOf(TransactionError);
  });
});

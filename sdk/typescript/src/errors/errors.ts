import * as codes from "../constants/errors.js";

export class SeocheonError extends Error {
  readonly code: number;

  constructor(code: number, message: string) {
    super(message);
    this.name = "SeocheonError";
    this.code = code;
  }
}

export class ConnectionError extends SeocheonError {
  constructor(message: string) {
    super(codes.ERR_NOT_CONNECTED, message);
    this.name = "ConnectionError";
  }
}

export class QueryError extends SeocheonError {
  constructor(message: string, code: number = codes.ERR_QUERY_FAILED) {
    super(code, message);
    this.name = "QueryError";
  }
}

export class TransactionError extends SeocheonError {
  constructor(message: string, code: number = codes.ERR_BROADCAST_FAILED) {
    super(code, message);
    this.name = "TransactionError";
  }
}

export class ValidationError extends SeocheonError {
  constructor(message: string, code: number = codes.ERR_INVALID_CONFIG) {
    super(code, message);
    this.name = "ValidationError";
  }
}

export class TimeoutError extends SeocheonError {
  constructor(txHash: string) {
    super(codes.ERR_TX_TIMEOUT, `TX confirmation timeout: ${txHash}`);
    this.name = "TimeoutError";
  }
}

export function mapAbciError(code: number, log: string): SeocheonError {
  if (code >= 1200 && code < 1300) {
    return new TransactionError(log, code);
  }
  if (code >= 1100 && code < 1200) {
    return new TransactionError(log, code);
  }
  return new TransactionError(log, code);
}

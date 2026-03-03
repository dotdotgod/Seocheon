// 4-phase TX pipeline: Assembly -> Signing -> Broadcast -> Confirmation.
// Matches the Go SDK's pipeline.go architecture.

import type { MessageEncoder } from "./messages.js";
import { Coin } from "./messages.js";
import { encodeTxBody, encodeAuthInfo, encodeSignDoc, encodeTxRaw } from "./envelope.js";
import { defaultGasForMessage, calculateFee, DEFAULT_GAS_PRICE, DEFAULT_FEE_DENOM } from "./gas.js";

/** Pipeline configuration. */
export interface PipelineConfig {
  chainId: string;
  gasPrice: number;
  confirmTimeoutMs: number;
  pollIntervalMs: number;
}

/** TX build/broadcast request. */
export interface TxRequest {
  message: MessageEncoder;
  memo?: string;
  timeoutHeight?: number;
  gasLimit?: number;
  feeAmount?: number;
  feeDenom?: string;
}

/** TX event attribute. */
export interface TxEventAttribute {
  key: string;
  value: string;
}

/** TX event. */
export interface TxEvent {
  type: string;
  attributes: TxEventAttribute[];
}

/** TX result after confirmation. */
export interface TxResult {
  txHash: string;
  height: number;
  code: number;
  gasUsed: number;
  gasWanted: number;
  rawLog: string;
  events: TxEvent[];
}

/** Signer abstraction for the TX pipeline. */
export interface Signer {
  sign(data: Uint8Array): Promise<Uint8Array>;
  getAddress(): Promise<string>;
  getPubKey(): Promise<Uint8Array>;
}

/** Chain querier abstraction for the TX pipeline. */
export interface ChainQuerier {
  getAccountInfo(address: string): Promise<{ accountNumber: number; sequence: number }>;
  broadcastTxSync(txBytes: Uint8Array): Promise<{ txHash: string; code: number; rawLog: string }>;
  getTxResult(txHash: string): Promise<TxResult>;
}

/**
 * Executes the full 4-phase TX pipeline:
 *   Phase 1 - Assembly: query account, build TxBody + AuthInfo + SignDoc
 *   Phase 2 - Signing: sign the SignDoc
 *   Phase 3 - Broadcast: encode TxRaw and broadcast
 *   Phase 4 - Confirmation: poll for TX inclusion
 */
export async function executeTx(
  querier: ChainQuerier,
  signer: Signer,
  cfg: PipelineConfig,
  req: TxRequest,
): Promise<TxResult> {
  // Phase 1: Assembly
  const address = await signer.getAddress();
  const pubKey = await signer.getPubKey();
  const { accountNumber, sequence } = await querier.getAccountInfo(address);

  // Determine gas limit
  let gasLimit = req.gasLimit ?? 0;
  if (gasLimit === 0) {
    gasLimit = defaultGasForMessage(req.message.typeUrl());
  }

  // Determine fee
  let feeAmount = req.feeAmount ?? 0;
  if (feeAmount === 0) {
    const gasPrice = cfg.gasPrice || DEFAULT_GAS_PRICE;
    feeAmount = calculateFee(gasLimit, gasPrice);
  }

  const feeDenom = req.feeDenom ?? DEFAULT_FEE_DENOM;

  // Encode TxBody
  const bodyBytes = encodeTxBody(
    [req.message],
    req.memo ?? "",
    req.timeoutHeight ?? 0,
  );

  // Encode AuthInfo
  const feeCoins = [new Coin(feeDenom, String(feeAmount))];
  const authInfoBytes = encodeAuthInfo(pubKey, sequence, feeCoins, gasLimit);

  // Encode SignDoc
  const signDocBytes = encodeSignDoc(bodyBytes, authInfoBytes, cfg.chainId, accountNumber);

  // Phase 2: Signing
  const signature = await signer.sign(signDocBytes);

  // Phase 3: Broadcast
  const txRawBytes = encodeTxRaw(bodyBytes, authInfoBytes, signature);

  const broadcastResult = await querier.broadcastTxSync(txRawBytes);

  // Check broadcast result code
  if (broadcastResult.code !== 0) {
    return {
      txHash: broadcastResult.txHash,
      height: 0,
      code: broadcastResult.code,
      gasUsed: 0,
      gasWanted: 0,
      rawLog: broadcastResult.rawLog,
      events: [],
    };
  }

  // Phase 4: Confirmation
  return pollTxConfirmation(
    querier,
    broadcastResult.txHash,
    cfg.confirmTimeoutMs,
    cfg.pollIntervalMs,
  );
}

/**
 * Polls for a transaction result until confirmed or timeout.
 */
export async function pollTxConfirmation(
  querier: ChainQuerier,
  txHash: string,
  timeoutMs: number,
  pollIntervalMs: number,
): Promise<TxResult> {
  const deadline = Date.now() + timeoutMs;

  while (Date.now() < deadline) {
    try {
      const result = await querier.getTxResult(txHash);
      return result;
    } catch {
      // TX not yet indexed, continue polling
    }
    await sleep(pollIntervalMs);
  }

  // Timeout: return partial result
  throw new Error(`timeout waiting for tx ${txHash} after ${timeoutMs}ms`);
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Adapter that wraps a ChainClient as a ChainQuerier for the TX pipeline.
 */
export class ChainClientAdapter implements ChainQuerier {
  constructor(
    private readonly client: {
      getAccountInfo(address: string): Promise<{ account_number: number; sequence: number }>;
      broadcastTx(txBytes: Uint8Array, mode: string): Promise<{ tx_hash: string; code: number; raw_log: string }>;
      getTx(txHash: string): Promise<{
        tx_hash: string;
        height: number;
        tx_result: {
          code: number;
          gas_used: number;
          gas_wanted: number;
          log: string;
          events: Array<{ type: string; attributes: Array<{ key: string; value: string }> }>;
        };
      } | null>;
    },
  ) {}

  async getAccountInfo(address: string): Promise<{ accountNumber: number; sequence: number }> {
    const info = await this.client.getAccountInfo(address);
    return { accountNumber: info.account_number, sequence: info.sequence };
  }

  async broadcastTxSync(txBytes: Uint8Array): Promise<{ txHash: string; code: number; rawLog: string }> {
    const resp = await this.client.broadcastTx(txBytes, "sync");
    return { txHash: resp.tx_hash, code: resp.code, rawLog: resp.raw_log };
  }

  async getTxResult(txHash: string): Promise<TxResult> {
    const resp = await this.client.getTx(txHash);
    if (!resp) {
      throw new Error(`tx not found: ${txHash}`);
    }
    return {
      txHash: resp.tx_hash,
      height: resp.height,
      code: resp.tx_result.code,
      gasUsed: resp.tx_result.gas_used,
      gasWanted: resp.tx_result.gas_wanted,
      rawLog: resp.tx_result.log,
      events: resp.tx_result.events.map((e) => ({
        type: e.type,
        attributes: e.attributes.map((a) => ({ key: a.key, value: a.value })),
      })),
    };
  }
}

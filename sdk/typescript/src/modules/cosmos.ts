import type { ChainClient } from "../infrastructure/chain-client.js";
import type { SigningService } from "../infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../types/config.js";
import type {
  BalanceResponse,
  SendTokensResponse,
  BlockInfoResponse,
  TxResultResponse,
} from "../types/responses.js";
import { MsgSend, Coin } from "../infrastructure/messages.js";
import { ChainClientAdapter, executeTx } from "../infrastructure/tx-pipeline.js";
import type { PipelineConfig } from "../infrastructure/tx-pipeline.js";
import { formatKkot } from "../utils/denom.js";
import { TransactionError, QueryError } from "../errors/errors.js";
import { ERR_BROADCAST_FAILED, ERR_TX_NOT_FOUND } from "../constants/errors.js";

export class CosmosModule {
  constructor(
    private readonly chainClient: ChainClient,
    private readonly signingService: SigningService,
    private readonly txConfig: ResolvedTxConfig,
  ) {}

  async getBalance(address?: string, denom?: string): Promise<BalanceResponse> {
    const effectiveAddress =
      address ?? (await this.signingService.getAddress());
    const effectiveDenom = denom ?? "uppyeo";

    const response = await this.chainClient.queryRest<{
      balance: { denom: string; amount: string };
    }>(
      `/cosmos/bank/v1beta1/balances/${effectiveAddress}/by_denom?denom=${effectiveDenom}`,
    );

    const balanceAmount = response.balance?.amount ?? "0";

    return {
      address: effectiveAddress,
      balance: balanceAmount,
      balance_kkot: formatKkot(balanceAmount),
    };
  }

  async sendTokens(
    toAddress: string,
    amount: string,
    denom?: string,
  ): Promise<SendTokensResponse> {
    const effectiveDenom = denom ?? "uppyeo";
    const fromAddress = await this.signingService.getAddress();

    const msg = new MsgSend(fromAddress, toAddress, [
      new Coin(effectiveDenom, amount),
    ]);

    const querier = new ChainClientAdapter(this.chainClient);
    const pipelineCfg: PipelineConfig = {
      chainId: this.txConfig.chain_id,
      gasPrice: this.txConfig.gas_price,
      confirmTimeoutMs: this.txConfig.confirm_timeout_ms,
      pollIntervalMs: this.txConfig.confirm_poll_interval_ms,
    };

    const result = await executeTx(querier, this.signingService, pipelineCfg, {
      message: msg,
    });

    if (result.code !== 0) {
      throw new TransactionError(
        `sendTokens failed with code ${result.code}: ${result.rawLog}`,
        ERR_BROADCAST_FAILED,
      );
    }

    return {
      tx_hash: result.txHash,
      block_height: result.height,
    };
  }

  async getBlockInfo(): Promise<BlockInfoResponse> {
    const block = await this.chainClient.getLatestBlock();

    return {
      block_height: block.header.height,
      block_time: block.header.time,
      chain_id: block.header.chain_id,
      num_txs: block.data.txs.length,
    };
  }

  async getTxResult(txHash: string): Promise<TxResultResponse> {
    const tx = await this.chainClient.getTx(txHash);
    if (!tx) {
      throw new QueryError(
        `transaction not found: ${txHash}`,
        ERR_TX_NOT_FOUND,
      );
    }

    return {
      tx_hash: txHash,
      height: tx.height,
      code: tx.tx_result.code,
      gas_used: tx.tx_result.gas_used,
      gas_wanted: tx.tx_result.gas_wanted,
      raw_log: tx.tx_result.log,
      events: tx.tx_result.events.map((e) => ({
        type: e.type,
        attributes: e.attributes.map((a) => ({
          key: a.key,
          value: a.value,
        })),
      })),
    };
  }
}

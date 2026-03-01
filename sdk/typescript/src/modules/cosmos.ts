import type { ChainClient } from "../infrastructure/chain-client.js";
import type { SigningService } from "../infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../types/config.js";
import type {
  BalanceResponse,
  SendTokensResponse,
  BlockInfoResponse,
  TxResultResponse,
} from "../types/responses.js";
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

    const _msg = {
      typeUrl: "/cosmos.bank.v1beta1.MsgSend",
      value: {
        from_address: fromAddress,
        to_address: toAddress,
        amount: [{ denom: effectiveDenom, amount }],
      },
    };

    // Placeholder: actual TX broadcast requires CosmJS integration
    throw new TransactionError(
      "cosmos.sendTokens() requires full CosmJS TX pipeline integration",
      ERR_BROADCAST_FAILED,
    );
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

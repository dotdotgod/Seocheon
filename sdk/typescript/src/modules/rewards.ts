import type { ChainClient } from "../infrastructure/chain-client.js";
import type { SigningService } from "../infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../types/config.js";
import type {
  PendingRewardsResponse,
  WithdrawRewardsResponse,
} from "../types/responses.js";
import { MsgWithdrawNodeCommission } from "../infrastructure/messages.js";
import { ChainClientAdapter, executeTx } from "../infrastructure/tx-pipeline.js";
import type { PipelineConfig } from "../infrastructure/tx-pipeline.js";
import { formatKkot } from "../utils/denom.js";
import { TransactionError } from "../errors/errors.js";
import { ERR_BROADCAST_FAILED } from "../constants/errors.js";

export class RewardsModule {
  constructor(
    private readonly chainClient: ChainClient,
    private readonly signingService: SigningService,
    private readonly txConfig: ResolvedTxConfig,
  ) {}

  async getPending(nodeId?: string): Promise<PendingRewardsResponse> {
    const effectiveNodeId = nodeId ?? (await this.resolveOwnNodeId());

    const nodeResponse = await this.chainClient.queryRest<{
      node: {
        validator_address: string;
        agent_share: string;
      };
    }>(`/seocheon/node/v1/nodes/${effectiveNodeId}`);

    const validatorAddress = nodeResponse.node.validator_address;
    const agentSharePercent = parseFloat(nodeResponse.node.agent_share);

    let delegationReward = "0";
    try {
      const outstanding = await this.chainClient.queryRest<{
        rewards: { rewards: Array<{ denom: string; amount: string }> };
      }>(
        `/cosmos/distribution/v1beta1/validators/${validatorAddress}/outstanding_rewards`,
      );
      const uppyeoReward = outstanding.rewards.rewards?.find(
        (r) => r.denom === "uppyeo",
      );
      if (uppyeoReward) {
        delegationReward = uppyeoReward.amount.split(".")[0] ?? "0";
      }
    } catch {
      // Validator may not have rewards yet
    }

    let commissionTotal = "0";
    try {
      const commission = await this.chainClient.queryRest<{
        commission: { commission: Array<{ denom: string; amount: string }> };
      }>(
        `/cosmos/distribution/v1beta1/validators/${validatorAddress}/commission`,
      );
      const uppyeoCommission = commission.commission.commission?.find(
        (r) => r.denom === "uppyeo",
      );
      if (uppyeoCommission) {
        commissionTotal = uppyeoCommission.amount.split(".")[0] ?? "0";
      }
    } catch {
      // No commission yet
    }

    const activityReward = "0";

    const totalReward = (
      BigInt(delegationReward) + BigInt(activityReward)
    ).toString();

    const agentRatio = agentSharePercent / 100;
    const commissionBigInt = BigInt(commissionTotal);
    const agentShareAmount = (
      (commissionBigInt * BigInt(Math.round(agentRatio * 10000))) /
      10000n
    ).toString();
    const operatorShare = (
      commissionBigInt - BigInt(agentShareAmount)
    ).toString();

    return {
      delegation_reward: formatKkot(delegationReward),
      activity_reward: formatKkot(activityReward),
      total_reward: formatKkot(totalReward),
      commission_total: formatKkot(commissionTotal),
      operator_share: formatKkot(operatorShare),
      agent_share: formatKkot(agentShareAmount),
    };
  }

  async withdraw(): Promise<WithdrawRewardsResponse> {
    const operator = await this.signingService.getAddress();
    const msg = new MsgWithdrawNodeCommission(operator);

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
        `withdraw failed with code ${result.code}: ${result.rawLog}`,
        ERR_BROADCAST_FAILED,
      );
    }

    // Parse withdrawn amounts from events
    let withdrawnTotal = "0";
    let toOperator = "0";
    let toAgent = "0";
    for (const event of result.events) {
      if (event.type === "withdraw_commission" || event.type === "transfer") {
        for (const attr of event.attributes) {
          if (attr.key === "amount" && attr.value.includes("uppyeo")) {
            withdrawnTotal = attr.value.replace("uppyeo", "");
          }
          if (attr.key === "operator_amount" && attr.value.includes("uppyeo")) {
            toOperator = attr.value.replace("uppyeo", "");
          }
          if (attr.key === "agent_amount" && attr.value.includes("uppyeo")) {
            toAgent = attr.value.replace("uppyeo", "");
          }
        }
      }
    }

    return {
      tx_hash: result.txHash,
      withdrawn_total: formatKkot(withdrawnTotal),
      to_operator: formatKkot(toOperator),
      to_agent: formatKkot(toAgent),
    };
  }

  private async resolveOwnNodeId(): Promise<string> {
    const address = await this.signingService.getAddress();
    const response = await this.chainClient.queryRest<{
      node: { id: string };
    }>(`/seocheon/node/v1/nodes/by-agent/${address}`);
    return response.node.id;
  }
}

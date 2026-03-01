import type { ChainClient } from "../infrastructure/chain-client.js";
import type { SigningService } from "../infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../types/config.js";
import type {
  PendingRewardsResponse,
  WithdrawRewardsResponse,
} from "../types/responses.js";
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

    // Get node info for validator_address and agent_share
    const nodeResponse = await this.chainClient.queryRest<{
      node: {
        validator_address: string;
        agent_share: string;
      };
    }>(`/seocheon/node/v1/nodes/${effectiveNodeId}`);

    const validatorAddress = nodeResponse.node.validator_address;
    const agentSharePercent = parseFloat(nodeResponse.node.agent_share);

    // Query outstanding rewards
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

    // Query commission
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

    // Activity reward is estimated (would need activity pool calculation)
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
    // Build MsgWithdrawNodeCommission
    const _msg = {
      typeUrl: "/seocheon.node.v1.MsgWithdrawNodeCommission",
      value: { operator: await this.signingService.getAddress() },
    };

    // Placeholder: actual TX broadcast requires CosmJS integration
    throw new TransactionError(
      "rewards.withdraw() requires full CosmJS TX pipeline integration",
      ERR_BROADCAST_FAILED,
    );
  }

  private async resolveOwnNodeId(): Promise<string> {
    const address = await this.signingService.getAddress();
    const response = await this.chainClient.queryRest<{
      node: { id: string };
    }>(`/seocheon/node/v1/nodes/by-agent/${address}`);
    return response.node.id;
  }
}

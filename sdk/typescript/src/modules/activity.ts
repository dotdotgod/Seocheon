import type { ChainClient } from "../infrastructure/chain-client.js";
import type { SigningService } from "../infrastructure/signing-service.js";
import type { ResolvedTxConfig } from "../types/config.js";
import type {
  SubmitActivityResponse,
  GetActivitiesResponse,
  ActivityItem,
  GetQuotaResponse,
} from "../types/responses.js";
import { MsgSubmitActivity } from "../infrastructure/messages.js";
import { ChainClientAdapter, executeTx } from "../infrastructure/tx-pipeline.js";
import type { PipelineConfig } from "../infrastructure/tx-pipeline.js";
import { isValidActivityHash, isValidContentUri } from "../utils/hash.js";
import { computeEpoch, computeWindow } from "../utils/epoch.js";
import { ValidationError, TransactionError } from "../errors/errors.js";
import {
  ERR_INVALID_ACTIVITY_HASH,
  ERR_INVALID_CONTENT_URI,
  ERR_BROADCAST_FAILED,
} from "../constants/errors.js";

export class ActivityModule {
  constructor(
    private readonly chainClient: ChainClient,
    private readonly signingService: SigningService,
    private readonly txConfig: ResolvedTxConfig,
  ) {}

  async submit(
    activityHash: string,
    contentUri: string,
  ): Promise<SubmitActivityResponse> {
    if (!isValidActivityHash(activityHash)) {
      throw new ValidationError(
        "activity hash must be exactly 64 hex characters (32 bytes)",
        ERR_INVALID_ACTIVITY_HASH,
      );
    }
    if (!isValidContentUri(contentUri)) {
      throw new ValidationError(
        "content URI must not be empty",
        ERR_INVALID_CONTENT_URI,
      );
    }

    const submitter = await this.signingService.getAddress();
    const msg = new MsgSubmitActivity(submitter, activityHash, contentUri);

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
        `activity submit failed with code ${result.code}: ${result.rawLog}`,
        ERR_BROADCAST_FAILED,
      );
    }

    // Compute epoch/window from block height
    const params = await this.getActivityParams();
    const epoch = computeEpoch(
      result.height,
      params.epoch_length,
    );
    const window = computeWindow(
      result.height,
      params.epoch_length,
      params.windows_per_epoch,
    );

    // Query remaining quota
    let quotaRemaining = 0;
    try {
      const quota = await this.getQuota();
      quotaRemaining = quota.quota_remaining;
    } catch {
      // Quota query failure is non-fatal
    }

    return {
      tx_hash: result.txHash,
      block_height: result.height,
      epoch_number: epoch,
      window_number: window,
      quota_remaining: quotaRemaining,
    };
  }

  async getActivities(
    nodeId?: string,
    epochNumber?: number,
  ): Promise<GetActivitiesResponse> {
    const effectiveNodeId = nodeId ?? (await this.resolveOwnNodeId());
    const effectiveEpoch = epochNumber ?? await this.computeCurrentEpoch();

    const response = await this.chainClient.queryRest<{
      activities: Array<{
        activity_hash: string;
        content_uri: string;
        block_height: string;
      }>;
    }>(
      `/seocheon/activity/v1/nodes/${effectiveNodeId}/activities?epoch=${effectiveEpoch}`,
    );

    const params = await this.getActivityParams();
    const activities: ActivityItem[] = (response.activities ?? []).map(
      (record) => {
        const blockHeight = parseInt(record.block_height, 10);
        return {
          activity_hash: record.activity_hash,
          content_uri: record.content_uri,
          block_height: blockHeight,
          window_number: computeWindow(
            blockHeight,
            params.epoch_length,
            params.windows_per_epoch,
          ),
          tx_hash: "",
        };
      },
    );

    return {
      activities,
      total_count: activities.length,
    };
  }

  async getQuota(): Promise<GetQuotaResponse> {
    const nodeId = await this.resolveOwnNodeId();
    const epochNumber = await this.computeCurrentEpoch();

    const response = await this.chainClient.queryRest<{
      summary: { active_windows: string; total_activities: string; eligible: boolean };
      quota_used: string;
      quota_limit: string;
    }>(`/seocheon/activity/v1/nodes/${nodeId}/epochs/${epochNumber}`);

    const quotaUsed = parseInt(response.quota_used, 10);
    const quotaLimit = parseInt(response.quota_limit, 10);

    return {
      epoch_number: epochNumber,
      quota_total: quotaLimit,
      quota_used: quotaUsed,
      quota_remaining: quotaLimit - quotaUsed,
      is_feegrant: false,
      feegrant_expiry: null,
    };
  }

  private async resolveOwnNodeId(): Promise<string> {
    const address = await this.signingService.getAddress();
    const response = await this.chainClient.queryRest<{
      node: { id: string };
    }>(`/seocheon/node/v1/nodes/by-agent/${address}`);
    return response.node.id;
  }

  private async computeCurrentEpoch(): Promise<number> {
    try {
      const block = await this.chainClient.getLatestBlock();
      const params = await this.getActivityParams();
      return computeEpoch(block.header.height, params.epoch_length);
    } catch {
      return 0;
    }
  }

  private async getActivityParams(): Promise<{
    epoch_length: number;
    windows_per_epoch: number;
    min_active_windows: number;
  }> {
    const response = await this.chainClient.queryRest<{
      params: {
        epoch_length: string;
        windows_per_epoch: string;
        min_active_windows: string;
      };
    }>("/seocheon/activity/v1/params");
    return {
      epoch_length: parseInt(response.params.epoch_length, 10),
      windows_per_epoch: parseInt(response.params.windows_per_epoch, 10),
      min_active_windows: parseInt(response.params.min_active_windows, 10),
    };
  }
}

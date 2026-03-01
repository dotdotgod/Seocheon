import type { ChainClient } from "../infrastructure/chain-client.js";
import type {
  EpochInfoResponse,
  QualificationResponse,
  WindowActivity,
} from "../types/responses.js";
import type { SigningService } from "../infrastructure/signing-service.js";
import {
  computeWindow,
  formatProgress,
} from "../utils/epoch.js";

export class EpochModule {
  constructor(
    private readonly chainClient: ChainClient,
    private readonly signingService: SigningService,
  ) {}

  async getInfo(): Promise<EpochInfoResponse> {
    // Query proto response (4 fields)
    const proto = await this.chainClient.queryRest<{
      current_epoch: string;
      current_window: string;
      epoch_start_block: string;
      blocks_until_next_epoch: string;
    }>("/seocheon/activity/v1/epoch-info");

    const params = await this.getActivityParams();
    const windowLength = Math.floor(
      params.epoch_length / params.windows_per_epoch,
    );

    const block = await this.chainClient.getLatestBlock();
    const blockHeight = block.header.height;

    const currentEpoch = parseInt(proto.current_epoch, 10);
    const currentWindow = parseInt(proto.current_window, 10);
    const epochStartBlock = parseInt(proto.epoch_start_block, 10);
    const blocksUntilNextEpoch = parseInt(proto.blocks_until_next_epoch, 10);

    const epochEndBlockHeight = epochStartBlock + params.epoch_length - 1;
    const epochElapsed = blockHeight - epochStartBlock + 1;

    const winStartBlock = epochStartBlock + currentWindow * windowLength;
    const winEndBlock = winStartBlock + windowLength - 1;
    const windowElapsed = blockHeight - winStartBlock + 1;

    return {
      block_height: blockHeight,
      epoch_number: currentEpoch,
      epoch_start_block: epochStartBlock,
      epoch_end_block: epochEndBlockHeight,
      epoch_progress: formatProgress(epochElapsed, params.epoch_length),
      window_number: currentWindow,
      window_start_block: winStartBlock,
      window_end_block: winEndBlock,
      window_progress: formatProgress(windowElapsed, windowLength),
      blocks_until_next_window: winEndBlock - blockHeight,
      blocks_until_next_epoch: blocksUntilNextEpoch,
    };
  }

  async getQualification(
    nodeId?: string,
    epochNumber?: number,
  ): Promise<QualificationResponse> {
    const effectiveNodeId = nodeId ?? (await this.resolveOwnNodeId());

    // Get epoch info to determine current epoch
    const epochInfo = await this.chainClient.queryRest<{
      current_epoch: string;
      current_window: string;
      epoch_start_block: string;
      blocks_until_next_epoch: string;
    }>("/seocheon/activity/v1/epoch-info");

    const currentEpoch = parseInt(epochInfo.current_epoch, 10);
    const currentWindow = parseInt(epochInfo.current_window, 10);
    const effectiveEpoch = epochNumber ?? currentEpoch;

    const params = await this.getActivityParams();

    // Query node epoch activity
    const nodeEpoch = await this.chainClient.queryRest<{
      summary: {
        active_windows: string;
        total_activities: string;
        eligible: boolean;
      };
      quota_used: string;
      quota_limit: string;
    }>(
      `/seocheon/activity/v1/nodes/${effectiveNodeId}/epochs/${effectiveEpoch}`,
    );

    const activeWindows = parseInt(nodeEpoch.summary.active_windows, 10);

    // Calculate elapsed windows
    const elapsedWindows =
      effectiveEpoch === currentEpoch
        ? currentWindow + 1
        : params.windows_per_epoch;

    const remainingNeeded = Math.max(
      0,
      params.min_active_windows - activeWindows,
    );
    const remainingWindows = params.windows_per_epoch - elapsedWindows;
    const canStillQualify =
      activeWindows + remainingWindows >= params.min_active_windows;

    // Build window detail from activities
    const activitiesResponse = await this.chainClient.queryRest<{
      activities: Array<{
        activity_hash: string;
        block_height: string;
      }>;
    }>(
      `/seocheon/activity/v1/nodes/${effectiveNodeId}/activities?epoch=${effectiveEpoch}`,
    );

    const windowCounts = new Map<number, number>();
    for (const record of activitiesResponse.activities ?? []) {
      const wn = computeWindow(
        parseInt(record.block_height, 10),
        params.epoch_length,
        params.windows_per_epoch,
      );
      windowCounts.set(wn, (windowCounts.get(wn) ?? 0) + 1);
    }

    const windowDetail: WindowActivity[] = [];
    for (let w = 0; w < params.windows_per_epoch; w++) {
      const count = windowCounts.get(w) ?? 0;
      windowDetail.push({
        window_number: w,
        submission_count: count,
        has_activity: count > 0,
      });
    }

    return {
      epoch_number: effectiveEpoch,
      total_windows: params.windows_per_epoch,
      elapsed_windows: elapsedWindows,
      active_windows: activeWindows,
      required_windows: params.min_active_windows,
      is_qualified: nodeEpoch.summary.eligible,
      remaining_needed: remainingNeeded,
      can_still_qualify: canStillQualify,
      window_detail: windowDetail,
    };
  }

  private async resolveOwnNodeId(): Promise<string> {
    const address = await this.signingService.getAddress();
    const response = await this.chainClient.queryRest<{
      node: { id: string };
    }>(`/seocheon/node/v1/nodes/by-agent/${address}`);
    return response.node.id;
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

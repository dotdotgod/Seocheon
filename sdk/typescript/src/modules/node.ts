import type { ChainClient } from "../infrastructure/chain-client.js";
import type { SigningService } from "../infrastructure/signing-service.js";
import type {
  NodeInfoResponse,
  NodeSearchResponse,
  NodeSummary,
  NodeStatus,
} from "../types/responses.js";
import { formatKkot } from "../utils/denom.js";

const STATUS_MAP: Record<number, NodeStatus> = {
  0: "UNSPECIFIED",
  1: "REGISTERED",
  2: "ACTIVE",
  3: "INACTIVE",
  4: "JAILED",
};

function toNodeStatus(status: number | string): NodeStatus {
  if (typeof status === "number") {
    return STATUS_MAP[status] ?? "UNSPECIFIED";
  }
  const upper = status.toUpperCase();
  if (
    upper === "REGISTERED" ||
    upper === "ACTIVE" ||
    upper === "INACTIVE" ||
    upper === "JAILED" ||
    upper === "UNSPECIFIED"
  ) {
    return upper as NodeStatus;
  }
  // Try parsing enum prefix like "NODE_STATUS_ACTIVE"
  const stripped = upper.replace("NODE_STATUS_", "");
  return (STATUS_MAP[
    Object.entries(STATUS_MAP).find(([, v]) => v === stripped)?.[0] as unknown as number
  ] ?? "UNSPECIFIED");
}

export class NodeModule {
  constructor(
    private readonly chainClient: ChainClient,
    private readonly signingService: SigningService,
  ) {}

  async getInfo(nodeId?: string): Promise<NodeInfoResponse> {
    const effectiveNodeId = nodeId ?? (await this.resolveOwnNodeId());

    const response = await this.chainClient.queryRest<{
      node: {
        id: string;
        operator: string;
        agent_address: string;
        status: number | string;
        description: string;
        website: string;
        tags: string[];
        agent_share: string;
        validator_address: string;
        registered_at: string;
      };
    }>(`/seocheon/node/v1/nodes/${effectiveNodeId}`);

    const node = response.node;

    // Fetch staking info for delegation data
    let totalDelegation = "0";
    let selfDelegation = "0";
    let commissionRate = "0";

    try {
      const validator = await this.chainClient.queryRest<{
        validator: {
          tokens: string;
          commission: {
            commission_rates: { rate: string };
          };
        };
      }>(
        `/cosmos/staking/v1beta1/validators/${node.validator_address}`,
      );
      totalDelegation = validator.validator.tokens;
      commissionRate = validator.validator.commission.commission_rates.rate;

      const selfDel = await this.chainClient.queryRest<{
        delegation_response: {
          balance: { amount: string };
        };
      }>(
        `/cosmos/staking/v1beta1/validators/${node.validator_address}/delegations/${node.operator}`,
      );
      selfDelegation = selfDel.delegation_response.balance.amount;
    } catch {
      // Validator may not exist yet for REGISTERED nodes
    }

    return {
      node_id: node.id,
      operator: node.operator,
      agent_address: node.agent_address,
      status: toNodeStatus(node.status),
      description: node.description ?? "",
      website: node.website ?? "",
      tags: node.tags ?? [],
      commission_rate: commissionRate,
      agent_share: node.agent_share,
      total_delegation: formatKkot(totalDelegation),
      self_delegation: formatKkot(selfDelegation),
      validator_address: node.validator_address,
      registered_at: parseInt(node.registered_at, 10),
    };
  }

  async search(
    tag?: string,
    status?: string,
    limit?: number,
    orderBy?: string,
  ): Promise<NodeSearchResponse> {
    const effectiveLimit = limit ?? 20;
    const effectiveOrder = orderBy ?? "delegation";

    // Query nodes
    let rawNodes: Array<{
      id: string;
      status: number | string;
      tags: string[];
      description: string;
      validator_address: string;
      registered_at: string;
    }>;

    if (tag) {
      const response = await this.chainClient.queryRest<{
        nodes: typeof rawNodes;
      }>(`/seocheon/node/v1/nodes/by-tag/${tag}`);
      rawNodes = response.nodes ?? [];
    } else {
      const response = await this.chainClient.queryRest<{
        nodes: typeof rawNodes;
      }>("/seocheon/node/v1/nodes");
      rawNodes = response.nodes ?? [];
    }

    // SDK-level status filter
    if (status) {
      const targetStatus = status.toUpperCase();
      rawNodes = rawNodes.filter(
        (n) => toNodeStatus(n.status) === targetStatus,
      );
    }

    // Fetch delegation for sorting
    const nodesWithDelegation = await Promise.all(
      rawNodes.map(async (n) => {
        let delegation = "0";
        try {
          const validator = await this.chainClient.queryRest<{
            validator: { tokens: string };
          }>(
            `/cosmos/staking/v1beta1/validators/${n.validator_address}`,
          );
          delegation = validator.validator.tokens;
        } catch {
          // Validator may not exist
        }
        return { ...n, _delegation: BigInt(delegation) };
      }),
    );

    // Sort
    if (effectiveOrder === "delegation") {
      nodesWithDelegation.sort((a, b) =>
        a._delegation > b._delegation ? -1 : a._delegation < b._delegation ? 1 : 0,
      );
    } else {
      nodesWithDelegation.sort(
        (a, b) => parseInt(b.registered_at, 10) - parseInt(a.registered_at, 10),
      );
    }

    const totalCount = nodesWithDelegation.length;
    const limited = nodesWithDelegation.slice(0, effectiveLimit);

    const nodes: NodeSummary[] = limited.map((n) => ({
      node_id: n.id,
      status: toNodeStatus(n.status),
      tags: n.tags ?? [],
      total_delegation: formatKkot(n._delegation.toString()),
      description: n.description ?? "",
    }));

    return { nodes, total_count: totalCount };
  }

  private async resolveOwnNodeId(): Promise<string> {
    const address = await this.signingService.getAddress();
    const response = await this.chainClient.queryRest<{
      node: { id: string };
    }>(`/seocheon/node/v1/nodes/by-agent/${address}`);
    return response.node.id;
  }
}

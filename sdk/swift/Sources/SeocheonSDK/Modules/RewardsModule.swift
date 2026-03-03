import Foundation

/// Provides reward-related operations.
public final class RewardsModule: @unchecked Sendable {
    private let client: ChainClient
    private let signer: SigningService
    private let txQuerier: ChainQuerier
    private let txConfig: PipelineConfig
    private let txSigner: TxSigner

    internal init(client: ChainClient, signer: SigningService, chainID: String) {
        self.client = client
        self.signer = signer
        self.txQuerier = ChainClientAdapter(client: client)
        self.txConfig = PipelineConfig.defaultConfig(chainID: chainID)
        self.txSigner = SigningServiceAdapter(service: signer)
    }

    /// Returns pending (unwithdrawn) rewards for a node.
    public func getPending(nodeID: String = "") async throws -> PendingRewardsResponse {
        let effectiveNodeID = nodeID.isEmpty ? try await resolveOwnNodeID() : nodeID

        let nodeInfo = try await getNode(effectiveNodeID)

        var delegationReward: Int64 = 0
        var commissionTotal: Int64 = 0
        if !nodeInfo.validatorAddress.isEmpty {
            delegationReward = await queryOutstandingRewards(nodeInfo.validatorAddress)
            commissionTotal = await queryCommission(nodeInfo.validatorAddress)
        }

        let activityReward: Int64 = 0
        let totalReward = delegationReward + activityReward

        let agentShareRatio = parseAgentShare(nodeInfo.agentShare)
        let operatorShareAmount = Int64(Double(commissionTotal) * (1.0 - agentShareRatio))
        let agentShareAmount = commissionTotal - operatorShareAmount

        return PendingRewardsResponse(
            delegationReward: ConvertUtils.formatKkot(delegationReward),
            activityReward: ConvertUtils.formatKkot(activityReward),
            totalReward: ConvertUtils.formatKkot(totalReward),
            commissionTotal: ConvertUtils.formatKkot(commissionTotal),
            operatorShare: ConvertUtils.formatKkot(operatorShareAmount),
            agentShare: ConvertUtils.formatKkot(agentShareAmount)
        )
    }

    /// Withdraws all pending rewards.
    public func withdraw() async throws -> WithdrawRewardsResponse {
        let msg = MsgWithdrawNodeCommission(operator: signer.getAddress())

        let result = try await TxPipeline.executeTx(querier: txQuerier, signer: txSigner, config: txConfig, request: TxRequest(message: msg))

        if result.code != 0 {
            throw SDKError.fromABCICode(result.code)
        }

        var withdrawnTotal: Int64 = 0
        var toOperator: Int64 = 0
        var toAgent: Int64 = 0

        for evt in result.events {
            if evt.type == "withdraw_commission" || evt.type == "withdraw_node_commission" {
                for attr in evt.attributes {
                    switch attr.key {
                    case "amount": withdrawnTotal = parseAmount(attr.value)
                    case "operator_share": toOperator = parseAmount(attr.value)
                    case "agent_share": toAgent = parseAmount(attr.value)
                    default: break
                    }
                }
            }
        }

        if withdrawnTotal > 0 && toOperator == 0 && toAgent == 0 {
            toOperator = withdrawnTotal * 80 / 100
            toAgent = withdrawnTotal - toOperator
        }

        return WithdrawRewardsResponse(
            txHash: result.txHash,
            withdrawnTotal: ConvertUtils.formatKkot(withdrawnTotal),
            toOperator: ConvertUtils.formatKkot(toOperator),
            toAgent: ConvertUtils.formatKkot(toAgent)
        )
    }

    // MARK: - Private

    private struct NodeInfo {
        let validatorAddress: String
        let agentShare: String
    }

    private func getNode(_ nodeID: String) async throws -> NodeInfo {
        let path = "/seocheon/node/v1/nodes/\(nodeID)"
        let data = try await client.queryREST(path: path)

        struct NodeResult: Codable {
            struct Node: Codable {
                let validator_address: String?
                let agent_share: String?
            }
            let node: Node?
        }

        let result = try JSONDecoder().decode(NodeResult.self, from: data)
        guard let n = result.node else { throw SDKError.nodeNotFound }
        return NodeInfo(validatorAddress: n.validator_address ?? "", agentShare: n.agent_share ?? "0.2")
    }

    private func resolveOwnNodeID() async throws -> String {
        let path = "/seocheon/node/v1/nodes/by-agent/\(signer.getAddress())"
        let data = try await client.queryREST(path: path)

        struct NodeResult: Codable { struct Node: Codable { let id: String? }; let node: Node? }
        let result = try JSONDecoder().decode(NodeResult.self, from: data)
        guard let id = result.node?.id, !id.isEmpty else { throw SDKError.nodeNotFound }
        return id
    }

    private func queryOutstandingRewards(_ valAddr: String) async -> Int64 {
        guard let data = try? await client.queryREST(path: "/cosmos/distribution/v1beta1/validators/\(valAddr)/outstanding_rewards") else { return 0 }

        struct RewardsResult: Codable {
            struct Rewards: Codable {
                struct Coin: Codable { let denom: String?; let amount: String? }
                let rewards: [Coin]?
            }
            let rewards: Rewards?
        }

        guard let result = try? JSONDecoder().decode(RewardsResult.self, from: data) else { return 0 }
        for r in result.rewards?.rewards ?? [] {
            if r.denom == "uppyeo" {
                return Int64(Double(r.amount ?? "0") ?? 0)
            }
        }
        return 0
    }

    private func queryCommission(_ valAddr: String) async -> Int64 {
        guard let data = try? await client.queryREST(path: "/cosmos/distribution/v1beta1/validators/\(valAddr)/commission") else { return 0 }

        struct CommissionResult: Codable {
            struct Commission: Codable {
                struct Coin: Codable { let denom: String?; let amount: String? }
                let commission: [Coin]?
            }
            let commission: Commission?
        }

        guard let result = try? JSONDecoder().decode(CommissionResult.self, from: data) else { return 0 }
        for c in result.commission?.commission ?? [] {
            if c.denom == "uppyeo" {
                return Int64(Double(c.amount ?? "0") ?? 0)
            }
        }
        return 0
    }

    private func parseAgentShare(_ s: String) -> Double {
        guard let val = Double(s) else { return 0.2 }
        return val > 1 ? val / 100.0 : val
    }

    private func parseAmount(_ s: String) -> Int64 {
        var result: Int64 = 0
        for c in s {
            guard c.isNumber else { break }
            result = result * 10 + Int64(c.asciiValue! - Character("0").asciiValue!)
        }
        return result
    }
}

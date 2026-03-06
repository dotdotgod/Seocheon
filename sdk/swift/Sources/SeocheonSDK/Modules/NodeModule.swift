import Foundation

/// Provides node-related operations.
public final class NodeModule: @unchecked Sendable {
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

    /// Returns detailed information about a node.
    public func getInfo(nodeID: String = "") async throws -> NodeInfoResponse {
        let effectiveNodeID = nodeID.isEmpty ? try await resolveOwnNodeID() : nodeID

        let path = "/seocheon/node/v1/nodes/\(effectiveNodeID)"
        let data = try await client.queryREST(path: path)

        struct NodeProto: Codable {
            let id: String?; let `operator`: String?; let agent_address: String?
            let status: String?; let description: String?; let website: String?
            let tags: [String]?; let agent_share: String?
            let validator_address: String?; let registered_at: String?
        }
        struct NodeResult: Codable { let node: NodeProto? }

        let result = try JSONDecoder().decode(NodeResult.self, from: data)
        guard let n = result.node else { throw SDKError.nodeNotFound }

        var totalDelegation = "0", selfDelegation = "0", commissionRate = "0"
        if let valAddr = n.validator_address, !valAddr.isEmpty {
            (totalDelegation, commissionRate) = await queryValidatorInfo(valAddr)
            selfDelegation = await querySelfDelegation(operator: n.operator ?? "", valAddr: valAddr)
        }

        return NodeInfoResponse(
            nodeId: n.id ?? "",
            operator: n.operator ?? "",
            agentAddress: n.agent_address ?? "",
            status: NodeStatus.fromString(n.status ?? "").rawValue,
            description: n.description ?? "",
            website: n.website ?? "",
            tags: n.tags ?? [],
            commissionRate: commissionRate,
            agentShare: n.agent_share ?? "0.2",
            totalDelegation: totalDelegation,
            selfDelegation: selfDelegation,
            validatorAddress: n.validator_address ?? "",
            registeredAt: Int64(n.registered_at ?? "0") ?? 0
        )
    }

    /// Finds nodes matching the given criteria.
    public func search(tag: String = "", status: String = "", limit: UInt32 = 20, orderBy: String = "delegation") async throws -> NodeSearchResponse {
        let path = tag.isEmpty ? "/seocheon/node/v1/nodes" : "/seocheon/node/v1/nodes/by-tag/\(tag)"
        let data = try await client.queryREST(path: path)

        struct NodeProto: Codable {
            let id: String?; let status: String?; let tags: [String]?
            let description: String?; let validator_address: String?
        }
        struct NodesResult: Codable { let nodes: [NodeProto]? }

        let result = try JSONDecoder().decode(NodesResult.self, from: data)
        var filtered = result.nodes ?? []

        if !status.isEmpty {
            filtered = filtered.filter { NodeStatus.fromString($0.status ?? "").rawValue == status }
        }

        let summaries: [NodeSummary] = filtered.prefix(Int(limit)).map { n in
            NodeSummary(
                nodeId: n.id ?? "",
                status: NodeStatus.fromString(n.status ?? "").rawValue,
                tags: n.tags ?? [],
                totalDelegation: "0",
                description: n.description ?? ""
            )
        }

        return NodeSearchResponse(nodes: summaries, totalCount: UInt64(filtered.count))
    }

    /// Queries delegation confirmation status.
    public func getDelegationStatus(
        delegatorAddress: String,
        validatorAddress: String
    ) async throws -> DelegationStatusResponse {
        let path = "/seocheon/node/v1/delegation-confirmation/\(delegatorAddress)/\(validatorAddress)"
        let data = try await client.queryREST(path: path)

        struct Response: Codable {
            let expiry_epoch: String?
            let current_epoch: String?
            let in_renewal_window: Bool?
            let renewal_window_start: String?
        }

        let result = try JSONDecoder().decode(Response.self, from: data)
        return DelegationStatusResponse(
            expiryEpoch: Int64(result.expiry_epoch ?? "0") ?? 0,
            currentEpoch: Int64(result.current_epoch ?? "0") ?? 0,
            inRenewalWindow: result.in_renewal_window ?? false,
            renewalWindowStart: Int64(result.renewal_window_start ?? "0") ?? 0
        )
    }

    /// Confirms delegation for a validator.
    public func confirmDelegation(validatorAddress: String) async throws -> TxResultResponse {
        let msg = MsgConfirmDelegation(
            delegatorAddress: signer.getAddress(),
            validatorAddress: validatorAddress
        )
        let result = try await TxPipeline.executeTx(
            querier: txQuerier,
            signer: txSigner,
            config: txConfig,
            request: TxRequest(message: msg)
        )

        if result.code != 0 {
            throw SDKError.fromABCICode(result.code)
        }

        let events = result.events.map { e in
            TxEvent(type: e.type, attributes: e.attributes.map { EventAttribute(key: $0.key, value: $0.value) })
        }
        return TxResultResponse(
            txHash: result.txHash,
            height: result.height,
            code: result.code,
            gasUsed: result.gasUsed,
            gasWanted: result.gasWanted,
            rawLog: result.rawLog,
            events: events
        )
    }

    // MARK: - Private

    private func resolveOwnNodeID() async throws -> String {
        let agentAddr = signer.getAddress()
        let path = "/seocheon/node/v1/nodes/by-agent/\(agentAddr)"
        let data = try await client.queryREST(path: path)

        struct NodeResult: Codable { struct Node: Codable { let id: String? }; let node: Node? }
        let result = try JSONDecoder().decode(NodeResult.self, from: data)
        guard let id = result.node?.id, !id.isEmpty else { throw SDKError.nodeNotFound }
        return id
    }

    private func queryValidatorInfo(_ valAddr: String) async -> (String, String) {
        guard let data = try? await client.queryREST(path: "/cosmos/staking/v1beta1/validators/\(valAddr)") else {
            return ("0", "0")
        }
        struct ValResult: Codable {
            struct Validator: Codable {
                let tokens: String?
                struct Commission: Codable {
                    struct Rates: Codable { let rate: String? }
                    let commission_rates: Rates?
                }
                let commission: Commission?
            }
            let validator: Validator?
        }
        guard let result = try? JSONDecoder().decode(ValResult.self, from: data) else { return ("0", "0") }
        return (result.validator?.tokens ?? "0", result.validator?.commission?.commission_rates?.rate ?? "0")
    }

    private func querySelfDelegation(operator opAddr: String, valAddr: String) async -> String {
        guard let data = try? await client.queryREST(path: "/cosmos/staking/v1beta1/validators/\(valAddr)/delegations/\(opAddr)") else {
            return "0"
        }
        struct DelResult: Codable {
            struct DelegationResponse: Codable {
                struct Balance: Codable { let amount: String? }
                let balance: Balance?
            }
            let delegation_response: DelegationResponse?
        }
        guard let result = try? JSONDecoder().decode(DelResult.self, from: data) else { return "0" }
        return result.delegation_response?.balance?.amount ?? "0"
    }
}

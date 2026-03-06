import Foundation

/// Provides activity-related operations.
public final class ActivityModule: @unchecked Sendable {
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

    /// Submits an activity hash to the chain.
    public func submit(activityHash: String, contentURI: String) async throws -> SubmitActivityResponse {
        guard HashUtils.verifyActivityHash(activityHash) else {
            throw SDKError.invalidActivityHash
        }
        guard !contentURI.isEmpty else {
            throw SDKError.invalidContentURI
        }

        let msg = MsgSubmitActivity(submitter: signer.getAddress(), activityHash: activityHash, contentURI: contentURI)
        let result = try await TxPipeline.executeTx(querier: txQuerier, signer: txSigner, config: txConfig, request: TxRequest(message: msg))

        if result.code != 0 {
            throw SDKError.chainError(code: result.code, message: result.rawLog.isEmpty ? "chain error code \(result.code)" : result.rawLog)
        }

        let epochNumber = EpochUtils.computeEpoch(blockHeight: result.height)
        let windowNumber = EpochUtils.computeWindow(blockHeight: result.height)

        var quotaRemaining: UInt64 = 0
        if let quota = try? await getQuota() {
            quotaRemaining = quota.quotaRemaining
        }

        return SubmitActivityResponse(
            txHash: result.txHash,
            blockHeight: result.height,
            windowNumber: windowNumber,
            epochNumber: epochNumber,
            quotaRemaining: quotaRemaining
        )
    }

    /// Returns activity submission history for a node.
    public func getActivities(nodeID: String = "", epochNumber: Int64 = 0) async throws -> GetActivitiesResponse {
        let effectiveNodeID = nodeID.isEmpty ? try await resolveOwnNodeID() : nodeID
        let effectiveEpoch = epochNumber == 0 ? try await computeCurrentEpoch() : epochNumber

        let path = "/seocheon/activity/v1/nodes/\(effectiveNodeID)/activities?epoch=\(effectiveEpoch)"
        let data = try await client.queryREST(path: path)

        struct ActivitiesResult: Codable {
            struct Activity: Codable {
                let activity_hash: String?
                let content_uri: String?
                let block_height: String?
            }
            let activities: [Activity]?
        }

        let result = try JSONDecoder().decode(ActivitiesResult.self, from: data)
        let items = (result.activities ?? []).map { a in
            let height = Int64(a.block_height ?? "0") ?? 0
            let windowNum = EpochUtils.computeWindow(blockHeight: height)
            return ActivityItem(
                activityHash: a.activity_hash ?? "",
                contentUri: a.content_uri ?? "",
                blockHeight: height,
                windowNumber: windowNum,
                txHash: ""
            )
        }

        return GetActivitiesResponse(activities: items, totalCount: UInt64(items.count))
    }

    /// Returns the remaining activity submission quota.
    public func getQuota() async throws -> GetQuotaResponse {
        let nodeID = try await resolveOwnNodeID()
        let epochNum = try await computeCurrentEpoch()

        let path = "/seocheon/activity/v1/nodes/\(nodeID)/epochs/\(epochNum)"
        let data = try await client.queryREST(path: path)

        struct QuotaResult: Codable {
            let quota_used: String?
            let quota_limit: String?
        }

        let result = try JSONDecoder().decode(QuotaResult.self, from: data)
        let used = UInt64(result.quota_used ?? "0") ?? 0
        let limit = UInt64(result.quota_limit ?? "0") ?? 0

        return GetQuotaResponse(
            epochNumber: epochNum,
            quotaTotal: limit,
            quotaUsed: used,
            quotaRemaining: limit - used,
            isFeegrant: false,
            feegrantExpiry: nil
        )
    }

    // MARK: - Private

    private func resolveOwnNodeID() async throws -> String {
        let agentAddr = signer.getAddress()
        let path = "/seocheon/node/v1/nodes/by-agent/\(agentAddr)"
        let data = try await client.queryREST(path: path)

        struct NodeResult: Codable {
            struct Node: Codable {
                let id: String?
            }
            let node: Node?
        }

        let result = try JSONDecoder().decode(NodeResult.self, from: data)
        guard let id = result.node?.id, !id.isEmpty else {
            throw SDKError.submitterNotRegistered
        }
        return id
    }

    private func computeCurrentEpoch() async throws -> Int64 {
        let block = try await client.getLatestBlock()
        return EpochUtils.computeEpoch(blockHeight: block.height)
    }
}

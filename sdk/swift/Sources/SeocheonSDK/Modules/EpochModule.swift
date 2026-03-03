import Foundation

/// Provides epoch-related operations.
public final class EpochModule: @unchecked Sendable {
    private let client: ChainClient

    internal init(client: ChainClient) {
        self.client = client
    }

    /// Returns the current epoch and window state.
    public func getInfo() async throws -> EpochInfoResponse {
        let data = try await client.queryREST(path: "/seocheon/activity/v1/epoch-info")

        struct EpochProto: Codable {
            let current_epoch: String?
            let current_window: String?
            let epoch_start_block: String?
            let blocks_until_next_epoch: String?
        }

        let proto = try JSONDecoder().decode(EpochProto.self, from: data)
        let block = try await client.getLatestBlock()

        let epochLength = ChainConstants.epochLength
        let windowsPerEpoch = ChainConstants.windowsPerEpoch
        let windowLength = epochLength / windowsPerEpoch

        let currentEpoch = Int64(proto.current_epoch ?? "0") ?? 0
        let currentWindow = Int64(proto.current_window ?? "0") ?? 0
        let epochStartBlock = Int64(proto.epoch_start_block ?? "0") ?? 0
        let blocksUntilNextEpoch = Int64(proto.blocks_until_next_epoch ?? "0") ?? 0

        let epochEndBlock = epochStartBlock + epochLength - 1
        let windowStartBlock = epochStartBlock + (currentWindow * windowLength)
        let windowEndBlock = windowStartBlock + windowLength - 1
        let epochElapsed = block.height - epochStartBlock + 1
        let windowElapsed = block.height - windowStartBlock + 1

        return EpochInfoResponse(
            blockHeight: block.height,
            epochNumber: currentEpoch,
            epochStartBlock: epochStartBlock,
            epochEndBlock: epochEndBlock,
            epochProgress: "\(epochElapsed)/\(epochLength)",
            windowNumber: currentWindow,
            windowStartBlock: windowStartBlock,
            windowEndBlock: windowEndBlock,
            windowProgress: "\(windowElapsed)/\(windowLength)",
            blocksUntilNextWindow: windowEndBlock - block.height,
            blocksUntilNextEpoch: blocksUntilNextEpoch
        )
    }

    /// Returns the activity reward qualification status for a node.
    public func getQualification(nodeID: String, epochNumber: Int64 = 0) async throws -> QualificationResponse {
        guard !nodeID.isEmpty else {
            throw SDKError.invalidConfig("node_id is required for qualification query")
        }

        let effectiveEpoch: Int64
        if epochNumber == 0 {
            let info = try await getInfo()
            effectiveEpoch = info.epochNumber
        } else {
            effectiveEpoch = epochNumber
        }

        let path = "/seocheon/activity/v1/nodes/\(nodeID)/epochs/\(effectiveEpoch)"
        let data = try await client.queryREST(path: path)

        struct SummaryResult: Codable {
            struct Summary: Codable {
                let active_windows: String?
                let eligible: Bool?
            }
            let summary: Summary?
        }

        let result = try JSONDecoder().decode(SummaryResult.self, from: data)
        let activeWindows = Int64(result.summary?.active_windows ?? "0") ?? 0
        let minActiveWindows = ChainConstants.minActiveWindows
        let windowsPerEpoch = ChainConstants.windowsPerEpoch

        let epochInfo = try await getInfo()
        let elapsedWindows: Int64
        if effectiveEpoch == epochInfo.epochNumber {
            elapsedWindows = epochInfo.windowNumber + 1
        } else {
            elapsedWindows = windowsPerEpoch
        }

        var remainingNeeded = minActiveWindows - activeWindows
        if remainingNeeded < 0 { remainingNeeded = 0 }
        let remainingWindows = windowsPerEpoch - elapsedWindows
        let canStillQualify = (activeWindows + remainingWindows) >= minActiveWindows

        let windowDetail = (0..<windowsPerEpoch).map { w in
            WindowActivity(windowNumber: w, submissionCount: 0, hasActivity: false)
        }

        return QualificationResponse(
            epochNumber: effectiveEpoch,
            totalWindows: windowsPerEpoch,
            elapsedWindows: elapsedWindows,
            activeWindows: UInt64(activeWindows),
            requiredWindows: minActiveWindows,
            isQualified: result.summary?.eligible ?? false,
            remainingNeeded: remainingNeeded,
            canStillQualify: canStillQualify,
            windowDetail: windowDetail
        )
    }
}

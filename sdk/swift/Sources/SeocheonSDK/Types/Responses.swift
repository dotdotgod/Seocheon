import Foundation

/// Returned after submitting an activity.
public struct SubmitActivityResponse: Codable, Sendable {
    public let txHash: String
    public let blockHeight: Int64
    public let windowNumber: Int64
    public let epochNumber: Int64
    public let quotaRemaining: UInt64
}

/// A single activity record.
public struct ActivityItem: Codable, Sendable {
    public let activityHash: String
    public let contentUri: String
    public let blockHeight: Int64
    public let windowNumber: Int64
    public let txHash: String
}

/// Returned when querying activities.
public struct GetActivitiesResponse: Codable, Sendable {
    public let activities: [ActivityItem]
    public let totalCount: UInt64
}

/// Returned when querying activity quota.
public struct GetQuotaResponse: Codable, Sendable {
    public let epochNumber: Int64
    public let quotaTotal: UInt64
    public let quotaUsed: UInt64
    public let quotaRemaining: UInt64
    public let isFeegrant: Bool
    public let feegrantExpiry: Int64?
}

/// Returned when querying epoch information.
public struct EpochInfoResponse: Codable, Sendable {
    public let blockHeight: Int64
    public let epochNumber: Int64
    public let epochStartBlock: Int64
    public let epochEndBlock: Int64
    public let epochProgress: String
    public let windowNumber: Int64
    public let windowStartBlock: Int64
    public let windowEndBlock: Int64
    public let windowProgress: String
    public let blocksUntilNextWindow: Int64
    public let blocksUntilNextEpoch: Int64
}

/// Activity within a single window.
public struct WindowActivity: Codable, Sendable {
    public let windowNumber: Int64
    public var submissionCount: UInt64
    public var hasActivity: Bool
}

/// Returned when querying reward qualification.
public struct QualificationResponse: Codable, Sendable {
    public let epochNumber: Int64
    public let totalWindows: Int64
    public let elapsedWindows: Int64
    public let activeWindows: UInt64
    public let requiredWindows: Int64
    public let isQualified: Bool
    public let remainingNeeded: Int64
    public let canStillQualify: Bool
    public let windowDetail: [WindowActivity]
}

/// Returned when querying node information.
public struct NodeInfoResponse: Codable, Sendable {
    public let nodeId: String
    public let `operator`: String
    public let agentAddress: String
    public let status: String
    public let description: String
    public let website: String
    public let tags: [String]
    public let commissionRate: String
    public let agentShare: String
    public let totalDelegation: String
    public let selfDelegation: String
    public let validatorAddress: String
    public let registeredAt: Int64
}

/// A condensed view of a node.
public struct NodeSummary: Codable, Sendable {
    public let nodeId: String
    public let status: String
    public let tags: [String]
    public let totalDelegation: String
    public let description: String
}

/// Returned from a node search.
public struct NodeSearchResponse: Codable, Sendable {
    public let nodes: [NodeSummary]
    public let totalCount: UInt64
}

/// Returned when querying pending rewards.
public struct PendingRewardsResponse: Codable, Sendable {
    public let delegationReward: String
    public let activityReward: String
    public let totalReward: String
    public let commissionTotal: String
    public let operatorShare: String
    public let agentShare: String
}

/// Returned after withdrawing rewards.
public struct WithdrawRewardsResponse: Codable, Sendable {
    public let txHash: String
    public let withdrawnTotal: String
    public let toOperator: String
    public let toAgent: String
}

/// Returned when querying balance.
public struct BalanceResponse: Codable, Sendable {
    public let address: String
    public let balance: String
    public let balanceKkot: String
}

/// Returned after sending tokens.
public struct SendTokensResponse: Codable, Sendable {
    public let txHash: String
    public let blockHeight: Int64
}

/// Returned when querying block information.
public struct BlockInfoResponse: Codable, Sendable {
    public let blockHeight: Int64
    public let blockTime: String
    public let chainId: String
    public let numTxs: UInt64
}

/// An event emitted by a transaction.
public struct TxEvent: Codable, Sendable {
    public let type: String
    public let attributes: [EventAttribute]
}

/// A key-value pair within a TxEvent.
public struct EventAttribute: Codable, Sendable {
    public let key: String
    public let value: String
}

/// Returned when querying a transaction result.
public struct TxResultResponse: Codable, Sendable {
    public let txHash: String
    public let height: Int64
    public let code: UInt32
    public let gasUsed: UInt64
    public let gasWanted: UInt64
    public let rawLog: String
    public let events: [TxEvent]
}

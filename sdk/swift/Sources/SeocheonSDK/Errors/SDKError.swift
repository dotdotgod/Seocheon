import Foundation

/// Error type for the Seocheon SDK.
public enum SDKError: Error, LocalizedError, Equatable {
    // SDK-level errors (9000 series)
    case notConnected
    case broadcastFailed(String)
    case txTimeout
    case txNotFound
    case signingFailed(String)
    case invalidConfig(String)
    case queryFailed(String)
    case invalidAddress

    // x/node errors (1100 series)
    case nodeNotFound
    case nodeAlreadyExists
    case unauthorizedOperator
    case unauthorizedAgentMsg

    // x/activity errors (1200 series)
    case submitterNotRegistered
    case nodeNotEligible
    case duplicateActivityHash
    case quotaExceeded
    case invalidActivityHash
    case invalidContentURI

    // Chain error with raw code
    case chainError(code: UInt32, message: String)

    /// Numeric error code matching Go SDK conventions.
    public var code: UInt32 {
        switch self {
        case .notConnected: return 9000
        case .broadcastFailed: return 9001
        case .txTimeout: return 9002
        case .txNotFound: return 9003
        case .signingFailed: return 9004
        case .invalidConfig: return 9005
        case .queryFailed: return 9006
        case .invalidAddress: return 9007
        case .nodeNotFound: return 1101
        case .nodeAlreadyExists: return 1102
        case .unauthorizedOperator: return 1108
        case .unauthorizedAgentMsg: return 1109
        case .submitterNotRegistered: return 1200
        case .nodeNotEligible: return 1201
        case .duplicateActivityHash: return 1202
        case .quotaExceeded: return 1203
        case .invalidActivityHash: return 1204
        case .invalidContentURI: return 1205
        case .chainError(let code, _): return code
        }
    }

    public var errorDescription: String? {
        switch self {
        case .notConnected: return "[9000] SDK is not connected to chain"
        case .broadcastFailed(let msg): return "[9001] transaction broadcast failed: \(msg)"
        case .txTimeout: return "[9002] transaction confirmation timeout"
        case .txNotFound: return "[9003] transaction not found"
        case .signingFailed(let msg): return "[9004] transaction signing failed: \(msg)"
        case .invalidConfig(let msg): return "[9005] invalid SDK configuration: \(msg)"
        case .queryFailed(let msg): return "[9006] chain query failed: \(msg)"
        case .invalidAddress: return "[9007] invalid address format"
        case .nodeNotFound: return "[1101] node not found"
        case .nodeAlreadyExists: return "[1102] node already exists for this operator"
        case .unauthorizedOperator: return "[1108] unauthorized: signer is not the node operator"
        case .unauthorizedAgentMsg: return "[1109] agent address is not authorized for this message type"
        case .submitterNotRegistered: return "[1200] submitter agent address is not registered to any node"
        case .nodeNotEligible: return "[1201] node is not in an eligible status (REGISTERED or ACTIVE)"
        case .duplicateActivityHash: return "[1202] duplicate (activity_hash, content_uri) pair already exists"
        case .quotaExceeded: return "[1203] activity quota exceeded for this epoch"
        case .invalidActivityHash: return "[1204] activity hash must be exactly 64 hex characters (32 bytes)"
        case .invalidContentURI: return "[1205] content URI must not be empty"
        case .chainError(let code, let msg): return "[\(code)] \(msg)"
        }
    }

    /// Maps an ABCI error code to its corresponding SDKError.
    public static func fromABCICode(_ code: UInt32) -> SDKError {
        switch code {
        case 1101: return .nodeNotFound
        case 1102: return .nodeAlreadyExists
        case 1108: return .unauthorizedOperator
        case 1109: return .unauthorizedAgentMsg
        case 1200: return .submitterNotRegistered
        case 1201: return .nodeNotEligible
        case 1202: return .duplicateActivityHash
        case 1203: return .quotaExceeded
        case 1204: return .invalidActivityHash
        case 1205: return .invalidContentURI
        default: return .chainError(code: code, message: "chain error code \(code)")
        }
    }

    public static func == (lhs: SDKError, rhs: SDKError) -> Bool {
        return lhs.code == rhs.code
    }
}

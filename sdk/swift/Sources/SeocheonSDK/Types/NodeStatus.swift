import Foundation

/// Represents the status of a registered node.
public enum NodeStatus: String, Codable, Sendable {
    case unspecified = "UNSPECIFIED"
    case registered = "REGISTERED"
    case active = "ACTIVE"
    case inactive = "INACTIVE"
    case jailed = "JAILED"

    /// Converts a proto enum integer to NodeStatus.
    public static func fromInt(_ value: Int) -> NodeStatus {
        switch value {
        case 1: return .registered
        case 2: return .active
        case 3: return .inactive
        case 4: return .jailed
        default: return .unspecified
        }
    }

    /// Converts a protobuf JSON status string to NodeStatus.
    /// Accepts both "NODE_STATUS_REGISTERED" and "REGISTERED" formats.
    public static func fromString(_ value: String) -> NodeStatus {
        let s = value.hasPrefix("NODE_STATUS_")
            ? String(value.dropFirst("NODE_STATUS_".count))
            : value
        switch s {
        case "REGISTERED": return .registered
        case "ACTIVE":     return .active
        case "INACTIVE":   return .inactive
        case "JAILED":     return .jailed
        default:           return .unspecified
        }
    }
}

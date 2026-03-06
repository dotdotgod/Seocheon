// Package types defines data types for the Seocheon SDK.
package types

import "strings"

// NodeStatus represents the status of a registered node.
type NodeStatus string

const (
	NodeStatusUnspecified NodeStatus = "UNSPECIFIED"
	NodeStatusRegistered  NodeStatus = "REGISTERED"
	NodeStatusActive      NodeStatus = "ACTIVE"
	NodeStatusInactive    NodeStatus = "INACTIVE"
	NodeStatusJailed      NodeStatus = "JAILED"
)

// NodeStatusFromInt converts a proto enum integer to NodeStatus string.
func NodeStatusFromInt(status int) NodeStatus {
	switch status {
	case 1:
		return NodeStatusRegistered
	case 2:
		return NodeStatusActive
	case 3:
		return NodeStatusInactive
	case 4:
		return NodeStatusJailed
	default:
		return NodeStatusUnspecified
	}
}

// NodeStatusFromString converts a protobuf JSON status string to NodeStatus.
// Accepts both "NODE_STATUS_REGISTERED" and "REGISTERED" formats.
func NodeStatusFromString(status string) NodeStatus {
	s := strings.TrimPrefix(status, "NODE_STATUS_")
	switch s {
	case "REGISTERED":
		return NodeStatusRegistered
	case "ACTIVE":
		return NodeStatusActive
	case "INACTIVE":
		return NodeStatusInactive
	case "JAILED":
		return NodeStatusJailed
	default:
		return NodeStatusUnspecified
	}
}

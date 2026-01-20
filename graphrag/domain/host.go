package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// Host represents a network host in the knowledge graph.
// A host is identified by its IP address and may optionally include hostname, state, and OS information.
// Host is a root-level node with no parent relationships.
//
// Example:
//
//	host := &Host{
//	    IP:       "192.168.1.1",
//	    Hostname: "web-server.example.com",
//	    State:    "up",
//	    OS:       "Linux Ubuntu 22.04",
//	}
//
// Identifying Properties:
//   - ip (required): The IP address of the host
//
// Relationships:
//   - None (root node)
//   - Children: Port nodes (via HAS_PORT relationship)
type Host struct {
	// IP is the IP address of the host (IPv4 or IPv6).
	// This is the identifying property and is required.
	// Example: "192.168.1.1", "2001:db8::1"
	IP string

	// Hostname is the DNS hostname or FQDN of the host.
	// Optional. Example: "web-server.example.com"
	Hostname string

	// State represents the host's current state.
	// Optional. Common values: "up", "down", "unknown"
	State string

	// OS is the operating system detected on the host.
	// Optional. Example: "Linux Ubuntu 22.04", "Windows Server 2019"
	OS string
}

// NodeType returns the canonical node type for Host nodes.
// Implements GraphNode interface.
func (h *Host) NodeType() string {
	return graphrag.NodeTypeHost
}

// IdentifyingProperties returns the properties that uniquely identify this host.
// For Host nodes, only the IP address is identifying.
// Implements GraphNode interface.
func (h *Host) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropIP: h.IP,
	}
}

// Properties returns all properties to set on the host node.
// Includes both identifying properties (ip) and optional descriptive properties
// (hostname, state, os). Properties with empty string values are still included.
// Implements GraphNode interface.
func (h *Host) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropIP: h.IP,
	}

	// Add optional properties if they are set
	if h.Hostname != "" {
		props["hostname"] = h.Hostname
	}
	if h.State != "" {
		props[graphrag.PropState] = h.State
	}
	if h.OS != "" {
		props["os"] = h.OS
	}

	return props
}

// ParentRef returns nil because Host is a root node with no parent.
// Implements GraphNode interface.
func (h *Host) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Host is a root node.
// Implements GraphNode interface.
func (h *Host) RelationshipType() string {
	return ""
}

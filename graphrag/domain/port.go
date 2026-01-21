package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// Port represents a network port on a host in the knowledge graph.
// A port is identified by its host ID (IP address), port number, and protocol.
// Port nodes are children of Host nodes via the HAS_PORT relationship.
//
// Example (legacy):
//
//	port := &Port{
//	    HostID:   "192.168.1.1",
//	    Number:   80,
//	    Protocol: "tcp",
//	    State:    "open",
//	}
//
// Example (new BelongsTo pattern):
//
//	host := &Host{IP: "192.168.1.1"}
//	port := NewPort(80, "tcp").BelongsTo(host)
//	port.State = "open"
//
// Identifying Properties:
//   - host_id (required): IP address of the host this port belongs to
//   - number (required): Port number (1-65535)
//   - protocol (required): Protocol (tcp, udp, sctp)
//
// Relationships:
//   - Parent: Host node (via HAS_PORT relationship)
//   - Children: Service nodes (via RUNS_SERVICE relationship)
type Port struct {
	// HostID is the IP address of the host this port belongs to.
	// This is an identifying property and is required.
	// Must match the IP of an existing or to-be-created Host node.
	// Example: "192.168.1.1"
	HostID string `json:"host_id"`

	// Number is the port number.
	// This is an identifying property and is required.
	// Valid range: 1-65535
	Number int `json:"number"`

	// Protocol is the transport protocol for this port.
	// This is an identifying property and is required.
	// Common values: "tcp", "udp", "sctp"
	Protocol string `json:"protocol"`

	// State represents the port's current state.
	// Optional. Common values: "open", "closed", "filtered"
	State string `json:"state,omitempty"`

	// parent is the internal parent reference set via BelongsTo().
	// This takes precedence over HostID for ParentRef() if set.
	parent *NodeRef
}

// NewPort creates a new Port with the required identifying properties.
// This is the recommended way to create Port nodes using the builder pattern.
//
// Example:
//
//	host := &Host{IP: "192.168.1.1"}
//	port := NewPort(443, "tcp").BelongsTo(host)
//	port.State = "open"
func NewPort(number int, protocol string) *Port {
	return &Port{
		Number:   number,
		Protocol: protocol,
	}
}

// BelongsTo sets the parent host for this port.
// This method should be called before storing the port to establish the parent relationship.
// Returns the port instance for method chaining.
//
// Example:
//
//	host := &Host{IP: "192.168.1.1"}
//	port := NewPort(443, "tcp").BelongsTo(host)
//
// Note: If you set HostID directly (legacy pattern), BelongsTo takes precedence.
func (p *Port) BelongsTo(host *Host) *Port {
	if host == nil {
		panic("Port.BelongsTo: host cannot be nil")
	}
	if host.IP == "" {
		panic("Port.BelongsTo: host.IP cannot be empty")
	}

	// Set the internal parent reference
	p.parent = &NodeRef{
		NodeType: graphrag.NodeTypeHost,
		Properties: map[string]any{
			graphrag.PropIP: host.IP,
		},
	}

	// Also set HostID for backward compatibility with code that reads HostID directly
	p.HostID = host.IP

	return p
}

// NodeType returns the canonical node type for Port nodes.
// Implements GraphNode interface.
func (p *Port) NodeType() string {
	return graphrag.NodeTypePort
}

// IdentifyingProperties returns the properties that uniquely identify this port.
// For Port nodes, host_id, number, and protocol are all identifying properties.
// Implements GraphNode interface.
func (p *Port) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropHostID:   p.HostID,
		graphrag.PropNumber:   p.Number,
		graphrag.PropProtocol: p.Protocol,
	}
}

// Properties returns all properties to set on the port node.
// Includes both identifying properties (host_id, number, protocol) and
// optional descriptive properties (state).
// Implements GraphNode interface.
func (p *Port) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropHostID:   p.HostID,
		graphrag.PropNumber:   p.Number,
		graphrag.PropProtocol: p.Protocol,
	}

	// Add optional properties if they are set
	if p.State != "" {
		props[graphrag.PropState] = p.State
	}

	return props
}

// ParentRef returns a reference to the parent Host node.
// If BelongsTo() was called, returns the internal parent reference.
// Otherwise, falls back to using HostID for backward compatibility.
// Implements GraphNode interface.
func (p *Port) ParentRef() *NodeRef {
	// Use internal parent if set via BelongsTo()
	if p.parent != nil {
		return p.parent
	}

	// Fall back to HostID-based reference for backward compatibility
	if p.HostID == "" {
		return nil
	}

	return &NodeRef{
		NodeType: graphrag.NodeTypeHost,
		Properties: map[string]any{
			graphrag.PropIP: p.HostID,
		},
	}
}

// RelationshipType returns the relationship type to the parent Host node.
// Ports are connected to hosts via the HAS_PORT relationship.
// Implements GraphNode interface.
func (p *Port) RelationshipType() string {
	return graphrag.RelTypeHasPort
}

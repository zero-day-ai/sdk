package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// Port represents a network port on a host in the knowledge graph.
// A port is identified by its host ID (IP address), port number, and protocol.
// Port nodes are children of Host nodes via the HAS_PORT relationship.
//
// Example:
//
//	port := &Port{
//	    HostID:   "192.168.1.1",
//	    Number:   80,
//	    Protocol: "tcp",
//	    State:    "open",
//	}
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
	HostID string

	// Number is the port number.
	// This is an identifying property and is required.
	// Valid range: 1-65535
	Number int

	// Protocol is the transport protocol for this port.
	// This is an identifying property and is required.
	// Common values: "tcp", "udp", "sctp"
	Protocol string

	// State represents the port's current state.
	// Optional. Common values: "open", "closed", "filtered"
	State string
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
// The parent is identified by the IP address (HostID).
// Implements GraphNode interface.
func (p *Port) ParentRef() *NodeRef {
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

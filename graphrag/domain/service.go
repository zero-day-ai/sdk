package domain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zero-day-ai/sdk/graphrag"
)

// Service represents a service running on a network port in the knowledge graph.
// A service is identified by its port ID (composite format) and service name.
// Service nodes are children of Port nodes via the RUNS_SERVICE relationship.
//
// Example:
//
//	service := &Service{
//	    PortID:  "192.168.1.1:80:tcp",
//	    Name:    "http",
//	    Version: "Apache 2.4.51",
//	    Banner:  "Apache/2.4.51 (Ubuntu)",
//	}
//
// Identifying Properties:
//   - port_id (required): Composite identifier in format "{host_id}:{number}:{protocol}"
//   - name (required): Service name (e.g., "http", "ssh", "mysql")
//
// Relationships:
//   - Parent: Port node (via RUNS_SERVICE relationship)
//   - Children: Endpoint nodes (via HAS_ENDPOINT relationship)
type Service struct {
	// PortID is the composite identifier for the port this service runs on.
	// This is an identifying property and is required.
	// Format: "{host_id}:{number}:{protocol}"
	// Example: "192.168.1.1:80:tcp"
	PortID string `json:"port_id"`

	// Name is the service name or protocol identifier.
	// This is an identifying property and is required.
	// Examples: "http", "https", "ssh", "mysql", "smtp"
	Name string `json:"name"`

	// Version is the detected version of the service.
	// Optional. Example: "Apache 2.4.51", "OpenSSH 8.2"
	Version string `json:"version,omitempty"`

	// Banner is the service banner or identification string.
	// Optional. Example: "Apache/2.4.51 (Ubuntu)", "SSH-2.0-OpenSSH_8.2p1"
	Banner string `json:"banner,omitempty"`
}

// NodeType returns the canonical node type for Service nodes.
// Implements GraphNode interface.
func (s *Service) NodeType() string {
	return graphrag.NodeTypeService
}

// IdentifyingProperties returns the properties that uniquely identify this service.
// For Service nodes, port_id and name are both identifying properties.
// Implements GraphNode interface.
func (s *Service) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropPortID: s.PortID,
		graphrag.PropName:   s.Name,
	}
}

// Properties returns all properties to set on the service node.
// Includes both identifying properties (port_id, name) and optional
// descriptive properties (version, banner).
// Implements GraphNode interface.
func (s *Service) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropPortID: s.PortID,
		graphrag.PropName:   s.Name,
	}

	// Add optional properties if they are set
	if s.Version != "" {
		props["version"] = s.Version
	}
	if s.Banner != "" {
		props["banner"] = s.Banner
	}

	return props
}

// ParentRef returns a reference to the parent Port node.
// The PortID is parsed to extract host_id, number, and protocol for the parent port.
// Returns nil if PortID cannot be parsed (invalid format).
// Implements GraphNode interface.
func (s *Service) ParentRef() *NodeRef {
	// Parse PortID format: "{host_id}:{number}:{protocol}"
	hostID, portNumber, protocol, err := parsePortID(s.PortID)
	if err != nil {
		// Invalid PortID format - return nil
		// In a production system, this should be validated earlier
		return nil
	}

	return &NodeRef{
		NodeType: graphrag.NodeTypePort,
		Properties: map[string]any{
			graphrag.PropHostID:   hostID,
			graphrag.PropNumber:   portNumber,
			graphrag.PropProtocol: protocol,
		},
	}
}

// RelationshipType returns the relationship type to the parent Port node.
// Services are connected to ports via the RUNS_SERVICE relationship.
// Implements GraphNode interface.
func (s *Service) RelationshipType() string {
	return graphrag.RelTypeRunsService
}

// parsePortID parses a composite port ID in the format "{host_id}:{number}:{protocol}".
// Returns the individual components or an error if the format is invalid.
//
// Example:
//
//	hostID, port, proto, err := parsePortID("192.168.1.1:80:tcp")
//	// hostID = "192.168.1.1", port = 80, proto = "tcp"
func parsePortID(portID string) (hostID string, portNumber int, protocol string, err error) {
	// Split on ':' - expecting at least 3 parts (host:port:protocol)
	// Note: IPv6 addresses may contain colons, so we take the last two parts as port:protocol
	parts := strings.Split(portID, ":")
	if len(parts) < 3 {
		return "", 0, "", fmt.Errorf("invalid port_id format: expected '{host_id}:{number}:{protocol}', got '%s'", portID)
	}

	// Last part is protocol
	protocol = parts[len(parts)-1]

	// Second-to-last part is port number
	portStr := parts[len(parts)-2]
	portNumber, err = strconv.Atoi(portStr)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid port number in port_id '%s': %w", portID, err)
	}

	// Everything before the last two parts is the host ID
	// This handles both IPv4 (e.g., "192.168.1.1") and IPv6 (e.g., "2001:db8::1")
	hostID = strings.Join(parts[:len(parts)-2], ":")

	return hostID, portNumber, protocol, nil
}

// Package domain provides strongly-typed domain objects for GraphRAG nodes.
// These types implement the GraphNode interface, allowing agents to create and
// manage knowledge graph nodes with compile-time type safety.
package domain

// GraphNode represents a strongly-typed domain object that can be stored in the GraphRAG knowledge graph.
// Implementations provide structured data that the GraphRAG system uses to create Neo4j nodes
// and relationships.
//
// The interface follows the builder pattern for graph construction:
//  1. NodeType() - Returns the canonical node type from the taxonomy (e.g., "host", "port")
//  2. IdentifyingProperties() - Returns properties that uniquely identify this node (natural key)
//  3. Properties() - Returns all properties to set on the node
//  4. ParentRef() - Returns reference to parent node for hierarchical relationships (nil for root nodes)
//  5. RelationshipType() - Returns the relationship type to the parent node
//
// Example usage:
//
//	host := &Host{IP: "192.168.1.1", Hostname: "web-server", State: "up"}
//	nodeType := host.NodeType()                   // "host"
//	idProps := host.IdentifyingProperties()       // {"ip": "192.168.1.1"}
//	allProps := host.Properties()                 // {"ip": "192.168.1.1", "hostname": "web-server", "state": "up"}
//	parent := host.ParentRef()                    // nil (host is a root node)
//
//	port := &Port{HostID: "192.168.1.1", Number: 80, Protocol: "tcp", State: "open"}
//	parent = port.ParentRef()                     // &NodeRef{NodeType: "host", Properties: {"ip": "192.168.1.1"}}
//	relType := port.RelationshipType()            // "HAS_PORT"
type GraphNode interface {
	// NodeType returns the canonical node type from the GraphRAG taxonomy.
	// This must be one of the NodeType* constants from taxonomy_generated.go
	// (e.g., "host", "port", "service", "finding").
	//
	// Example:
	//  host.NodeType() // Returns "host" (NodeTypeHost constant)
	NodeType() string

	// IdentifyingProperties returns the properties that uniquely identify this node.
	// These properties form the natural key for the node type and are used for:
	//  - Deterministic ID generation
	//  - Deduplication (preventing duplicate nodes)
	//  - Node lookups and queries
	//
	// The returned map must contain all identifying properties defined in the registry
	// for this node type. Missing properties will cause validation errors.
	//
	// Example:
	//  host.IdentifyingProperties()     // {"ip": "192.168.1.1"}
	//  port.IdentifyingProperties()     // {"host_id": "192.168.1.1", "number": 80, "protocol": "tcp"}
	//  service.IdentifyingProperties()  // {"port_id": "192.168.1.1:80:tcp", "name": "http"}
	IdentifyingProperties() map[string]any

	// Properties returns all properties to set on the node in the knowledge graph.
	// This includes both identifying properties and additional descriptive properties.
	//
	// Properties with nil or empty values may be filtered out by the GraphRAG system.
	//
	// Example:
	//  host.Properties() // {"ip": "192.168.1.1", "hostname": "web-server", "state": "up", "os": "Linux"}
	Properties() map[string]any

	// ParentRef returns a reference to the parent node for creating hierarchical relationships.
	// Returns nil if this is a root node (no parent relationship).
	//
	// The NodeRef contains:
	//  - NodeType: The parent node's type (e.g., "host" for a port's parent)
	//  - Properties: The parent's identifying properties used for lookup
	//
	// Example:
	//  host.ParentRef()     // nil (root node)
	//  port.ParentRef()     // &NodeRef{NodeType: "host", Properties: {"ip": "192.168.1.1"}}
	//  service.ParentRef()  // &NodeRef{NodeType: "port", Properties: {"host_id": "...", "number": 80, "protocol": "tcp"}}
	ParentRef() *NodeRef

	// RelationshipType returns the type of relationship to the parent node.
	// Returns empty string if this is a root node (no parent relationship).
	//
	// This must be one of the RelType* constants from taxonomy_generated.go
	// (e.g., "HAS_PORT", "RUNS_SERVICE", "HAS_SUBDOMAIN").
	//
	// Example:
	//  port.RelationshipType()     // "HAS_PORT" (RelTypeHasPort constant)
	//  service.RelationshipType()  // "RUNS_SERVICE" (RelTypeRunsService constant)
	RelationshipType() string
}

// NodeRef represents a reference to a parent node in the knowledge graph.
// It contains the parent node's type and identifying properties needed for lookup.
//
// NodeRef is used by GraphNode implementations to specify parent relationships,
// enabling the GraphRAG system to create hierarchical structures like:
//   - Host -> Port -> Service -> Endpoint
//   - Domain -> Subdomain -> Host
//   - Mission -> AgentRun -> ToolExecution
//
// Example:
//
//	// Reference to a Host node with IP "192.168.1.1"
//	hostRef := &NodeRef{
//	    NodeType: "host",
//	    Properties: map[string]any{"ip": "192.168.1.1"},
//	}
//
//	// Reference to a Port node
//	portRef := &NodeRef{
//	    NodeType: "port",
//	    Properties: map[string]any{
//	        "host_id": "192.168.1.1",
//	        "number": 80,
//	        "protocol": "tcp",
//	    },
//	}
type NodeRef struct {
	// NodeType is the canonical node type from the taxonomy (e.g., "host", "port").
	// Must be one of the NodeType* constants from taxonomy_generated.go.
	NodeType string

	// Properties contains the identifying properties for the parent node.
	// These must be sufficient to uniquely identify the parent node
	// (i.e., must contain all identifying properties from the registry).
	Properties map[string]any
}

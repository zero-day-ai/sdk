package domain

import "fmt"

// CustomEntity is a base type for agent-defined custom graph nodes.
// It enables agents to extend the graph schema with domain-specific types
// (e.g., "k8s:pod", "aws:security_group", "custom:vulnerability") without
// modifying the SDK or Gibson core.
//
// CustomEntity uses a namespace prefix to prevent conflicts with canonical types
// and to organize custom types by their domain or source agent.
//
// Hierarchy: Flexible - can be root node or have parent references
//
// Identifying Properties: Defined by IDProps map
// Parent: Optional - defined by Parent field
//
// Example (Kubernetes Pod):
//
//	pod := &CustomEntity{
//	    Namespace: "k8s",
//	    Type:      "pod",
//	    IDProps: map[string]any{
//	        "namespace": "default",
//	        "name":      "web-server-abc123",
//	    },
//	    AllProps: map[string]any{
//	        "namespace": "default",
//	        "name":      "web-server-abc123",
//	        "status":    "Running",
//	        "image":     "nginx:1.21",
//	        "node":      "node-01",
//	    },
//	}
//	// pod.NodeType() returns "k8s:pod"
//
// Example (AWS Security Group with parent):
//
//	securityGroup := &CustomEntity{
//	    Namespace: "aws",
//	    Type:      "security_group",
//	    IDProps: map[string]any{
//	        "id": "sg-0123456789abcdef0",
//	    },
//	    AllProps: map[string]any{
//	        "id":          "sg-0123456789abcdef0",
//	        "name":        "web-server-sg",
//	        "description": "Security group for web servers",
//	        "vpc_id":      "vpc-abc123",
//	    },
//	    Parent: &NodeRef{
//	        NodeType: "aws:vpc",
//	        Properties: map[string]any{
//	            "id": "vpc-abc123",
//	        },
//	    },
//	    ParentRel: "BELONGS_TO",
//	}
type CustomEntity struct {
	// Namespace is the namespace prefix for this custom type (e.g., "k8s", "aws", "custom").
	// The namespace helps organize custom types and prevent naming conflicts.
	// Required field.
	Namespace string

	// Type is the entity type within the namespace (e.g., "pod", "security_group").
	// Required field.
	Type string

	// IDProps contains the identifying properties that uniquely identify this entity.
	// These properties are used for MERGE operations and deduplication.
	// The keys are property names, and values are the property values.
	// Required field (must have at least one identifying property).
	IDProps map[string]any

	// AllProps contains all properties to set on the node, including identifying properties.
	// This should be a superset of IDProps.
	// Optional - if nil, IDProps will be used for all properties.
	AllProps map[string]any

	// Parent is an optional reference to the parent node.
	// If set, a relationship will be created to the parent using ParentRel as the type.
	// Optional field.
	Parent *NodeRef

	// ParentRel is the relationship type to the parent node.
	// Only used if Parent is non-nil.
	// Example: "BELONGS_TO", "PART_OF", "DEPLOYED_ON"
	// Optional field.
	ParentRel string
}

// NodeType returns the namespaced node type in the format "{Namespace}:{Type}".
// Example: "k8s:pod", "aws:security_group", "custom:vulnerability"
func (c *CustomEntity) NodeType() string {
	return fmt.Sprintf("%s:%s", c.Namespace, c.Type)
}

// IdentifyingProperties returns the identifying properties for this custom entity.
// These are the properties in the IDProps map that uniquely identify the entity.
func (c *CustomEntity) IdentifyingProperties() map[string]any {
	if c.IDProps == nil {
		return make(map[string]any)
	}

	// Return a copy to prevent external modification
	props := make(map[string]any, len(c.IDProps))
	for k, v := range c.IDProps {
		props[k] = v
	}
	return props
}

// Properties returns all properties to set on the custom entity node.
// If AllProps is set, it returns AllProps. Otherwise, it returns IDProps.
func (c *CustomEntity) Properties() map[string]any {
	var sourceProps map[string]any
	if c.AllProps != nil {
		sourceProps = c.AllProps
	} else {
		sourceProps = c.IDProps
	}

	if sourceProps == nil {
		return make(map[string]any)
	}

	// Return a copy to prevent external modification
	props := make(map[string]any, len(sourceProps))
	for k, v := range sourceProps {
		props[k] = v
	}
	return props
}

// ParentRef returns a reference to the parent node, or nil if no parent is set.
func (c *CustomEntity) ParentRef() *NodeRef {
	return c.Parent
}

// RelationshipType returns the relationship type to the parent node.
// Returns empty string if no parent is set.
func (c *CustomEntity) RelationshipType() string {
	if c.Parent == nil {
		return ""
	}
	return c.ParentRel
}

// NewCustomEntity is a convenience constructor for creating custom entities.
// It validates that required fields are set and provides a fluent API.
//
// Example:
//
//	pod := NewCustomEntity("k8s", "pod").
//	    WithIDProps(map[string]any{
//	        "namespace": "default",
//	        "name":      "web-server-abc123",
//	    }).
//	    WithAllProps(map[string]any{
//	        "namespace": "default",
//	        "name":      "web-server-abc123",
//	        "status":    "Running",
//	        "image":     "nginx:1.21",
//	    })
func NewCustomEntity(namespace, entityType string) *CustomEntity {
	return &CustomEntity{
		Namespace: namespace,
		Type:      entityType,
	}
}

// WithIDProps sets the identifying properties for the custom entity.
// These properties uniquely identify the entity and are used for deduplication.
func (c *CustomEntity) WithIDProps(props map[string]any) *CustomEntity {
	c.IDProps = props
	return c
}

// WithAllProps sets all properties for the custom entity.
// This should include both identifying and descriptive properties.
func (c *CustomEntity) WithAllProps(props map[string]any) *CustomEntity {
	c.AllProps = props
	return c
}

// WithParent sets the parent reference and relationship type.
// This creates a hierarchical relationship in the graph.
func (c *CustomEntity) WithParent(parent *NodeRef, relationshipType string) *CustomEntity {
	c.Parent = parent
	c.ParentRel = relationshipType
	return c
}

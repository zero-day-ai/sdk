package schema

// TaxonomyMapping defines how to map data into a graph node with relationships.
// It specifies the node type, property mappings (including identifying properties),
// and relationships to create when processing data according to a taxonomy.
type TaxonomyMapping struct {
	// NodeType is the type of node to create in the graph (e.g., "Asset", "Vulnerability")
	NodeType string `json:"node_type"`

	// IdentifyingProperties maps property names to JSONPath expressions.
	// These properties uniquely identify the node and are used for deterministic ID generation.
	// Example: {"hostname": "$.host", "ip": "$.ip_address"}
	IdentifyingProperties map[string]string `json:"identifying_properties"`

	// Properties maps source data fields to node properties (non-identifying)
	Properties []PropertyMapping `json:"properties,omitempty"`

	// Relationships defines edges to create to/from this node
	Relationships []RelationshipMapping `json:"relationships,omitempty"`
}

// PropertyMapping defines how to map a source field to a target node property.
// It supports default values and transformation functions.
type PropertyMapping struct {
	// Source is the field name in the source data
	Source string `json:"source"`

	// Target is the property name in the target node
	Target string `json:"target"`

	// Default is the default value if source is missing or empty
	Default any `json:"default,omitempty"`

	// Transform is a transformation function to apply (e.g., "lowercase", "uppercase", "trim")
	Transform string `json:"transform,omitempty"`
}

// NodeReference identifies a node by type and property mappings.
// It can reference either the current node being mapped ("self") or another node
// by specifying its type and the properties that identify it.
type NodeReference struct {
	// Type is the node type (e.g., "host", "port", "service").
	// Use "self" to reference the node currently being mapped.
	Type string `json:"type" yaml:"type"`

	// Properties maps identifying property names to JSONPath expressions.
	// This field is required when Type != "self" and specifies how to locate
	// the target node by extracting values from the source data.
	// Example: {"hostname": "$.target.host", "port": "$.target.port"}
	Properties map[string]string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

// RelationshipMapping defines a relationship to create between nodes.
// It supports conditional relationships and property mappings on the edge itself.
type RelationshipMapping struct {
	// Type is the relationship type (e.g., "HAS_VULNERABILITY", "AFFECTS")
	Type string `json:"type"`

	// From identifies the source node of the relationship.
	// Use Type="self" to reference the current node being mapped.
	From NodeReference `json:"from"`

	// To identifies the target node of the relationship.
	// Specify the node type and identifying properties to locate it.
	To NodeReference `json:"to"`

	// Condition is an optional condition for creating this relationship
	// Example: "{{.severity}} == 'critical'"
	Condition string `json:"condition,omitempty"`

	// Properties are property mappings for the relationship edge
	Properties []PropertyMapping `json:"properties,omitempty"`
}

// Fluent API helpers for building taxonomy mappings

// PropMap creates a simple property mapping from source to target.
func PropMap(source, target string) PropertyMapping {
	return PropertyMapping{
		Source: source,
		Target: target,
	}
}

// PropMapWithDefault creates a property mapping with a default value.
func PropMapWithDefault(source, target string, def any) PropertyMapping {
	return PropertyMapping{
		Source:  source,
		Target:  target,
		Default: def,
	}
}

// PropMapWithTransform creates a property mapping with a transformation function.
func PropMapWithTransform(source, target, transform string) PropertyMapping {
	return PropertyMapping{
		Source:    source,
		Target:    target,
		Transform: transform,
	}
}

// SelfNode creates a NodeReference for the current node being mapped.
func SelfNode() NodeReference {
	return NodeReference{
		Type: "self",
	}
}

// Node creates a NodeReference with the specified type and identifying properties.
func Node(nodeType string, properties map[string]string) NodeReference {
	return NodeReference{
		Type:       nodeType,
		Properties: properties,
	}
}

// Rel creates a simple relationship mapping between two node references.
func Rel(relType string, from, to NodeReference) RelationshipMapping {
	return RelationshipMapping{
		Type: relType,
		From: from,
		To:   to,
	}
}

// RelWithCondition creates a relationship mapping with a condition.
func RelWithCondition(relType string, from, to NodeReference, condition string) RelationshipMapping {
	return RelationshipMapping{
		Type:      relType,
		From:      from,
		To:        to,
		Condition: condition,
	}
}

// RelWithProps creates a relationship mapping with property mappings on the edge.
func RelWithProps(relType string, from, to NodeReference, props ...PropertyMapping) RelationshipMapping {
	return RelationshipMapping{
		Type:       relType,
		From:       from,
		To:         to,
		Properties: props,
	}
}

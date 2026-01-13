package schema

// TaxonomyMapping defines how to map data into a graph node with relationships.
// It specifies the node type, ID generation, property mappings, and relationships
// to create when processing data according to a taxonomy.
type TaxonomyMapping struct {
	// NodeType is the type of node to create in the graph (e.g., "Asset", "Vulnerability")
	NodeType string `json:"node_type"`

	// IDTemplate is a template string for generating node IDs
	// Example: "asset:{{.hostname}}" or "vuln:{{.cve_id}}"
	IDTemplate string `json:"id_template"`

	// Properties maps source data fields to node properties
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

// RelationshipMapping defines a relationship to create between nodes.
// It supports conditional relationships and property mappings on the edge itself.
type RelationshipMapping struct {
	// Type is the relationship type (e.g., "HAS_VULNERABILITY", "AFFECTS")
	Type string `json:"type"`

	// FromTemplate is a template for the source node ID
	// Example: "asset:{{.hostname}}"
	FromTemplate string `json:"from_template"`

	// ToTemplate is a template for the target node ID
	// Example: "vuln:{{.cve_id}}"
	ToTemplate string `json:"to_template"`

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

// Rel creates a simple relationship mapping.
func Rel(relType, from, to string) RelationshipMapping {
	return RelationshipMapping{
		Type:         relType,
		FromTemplate: from,
		ToTemplate:   to,
	}
}

// RelWithCondition creates a relationship mapping with a condition.
func RelWithCondition(relType, from, to, condition string) RelationshipMapping {
	return RelationshipMapping{
		Type:         relType,
		FromTemplate: from,
		ToTemplate:   to,
		Condition:    condition,
	}
}

// RelWithProps creates a relationship mapping with property mappings on the edge.
func RelWithProps(relType, from, to string, props ...PropertyMapping) RelationshipMapping {
	return RelationshipMapping{
		Type:         relType,
		FromTemplate: from,
		ToTemplate:   to,
		Properties:   props,
	}
}

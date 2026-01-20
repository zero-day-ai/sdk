package domain

import "github.com/zero-day-ai/sdk/graphrag"

// Technology represents a software, framework, or technology stack component detected on an asset.
// Technologies can be associated with services, endpoints, or other assets via USES_TECHNOLOGY relationships.
//
// Hierarchy: Technology is a root node (no parent)
//
// Identifying Properties: name, version
// Parent: None (root node)
//
// Example:
//
//	technology := &Technology{
//	    Name:     "nginx",
//	    Version:  "1.18.0",
//	    Category: "web-server",
//	    Vendor:   "Nginx Inc.",
//	}
type Technology struct {
	// Name is the technology name (e.g., "nginx", "Apache", "PostgreSQL").
	// This is an identifying property.
	Name string `json:"name"`

	// Version is the technology version (e.g., "1.18.0", "2.4.41").
	// This is an identifying property.
	Version string `json:"version,omitempty"`

	// Category is the technology category (e.g., "web-server", "database", "framework") (optional).
	Category string `json:"category,omitempty"`

	// Vendor is the technology vendor or maintainer (optional).
	Vendor string `json:"vendor,omitempty"`

	// CPE is the Common Platform Enumeration identifier (optional).
	CPE string `json:"cpe,omitempty"`

	// License is the software license (optional).
	License string `json:"license,omitempty"`

	// EOL is the end-of-life date (optional).
	EOL string `json:"eol,omitempty"`
}

// NodeType returns the canonical node type for technologies.
func (t *Technology) NodeType() string {
	return graphrag.NodeTypeTechnology
}

// IdentifyingProperties returns the properties that uniquely identify this technology.
// A technology is identified by its name and version.
func (t *Technology) IdentifyingProperties() map[string]any {
	return map[string]any{
		"name":    t.Name,
		"version": t.Version,
	}
}

// Properties returns all properties to set on the technology node.
func (t *Technology) Properties() map[string]any {
	props := map[string]any{
		"name":    t.Name,
		"version": t.Version,
	}

	// Add optional properties if present
	if t.Category != "" {
		props["category"] = t.Category
	}
	if t.Vendor != "" {
		props["vendor"] = t.Vendor
	}
	if t.CPE != "" {
		props["cpe"] = t.CPE
	}
	if t.License != "" {
		props["license"] = t.License
	}
	if t.EOL != "" {
		props["eol"] = t.EOL
	}

	return props
}

// ParentRef returns nil because Technology is a root node.
func (t *Technology) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Technology has no parent.
func (t *Technology) RelationshipType() string {
	return ""
}

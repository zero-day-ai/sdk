package domain

import "github.com/zero-day-ai/sdk/graphrag"

// Domain represents a root domain entity representing a registered domain name.
// Domains are top-level entities that can have subdomains and resolve to hosts.
//
// Hierarchy: Domain is a root node (no parent)
//
// Identifying Properties: name
// Parent: None (root node)
//
// Example:
//
//	domain := &Domain{
//	    Name:       "example.com",
//	    Registrar:  "GoDaddy",
//	    CreatedAt:  time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC),
//	    ExpiresAt:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
//	}
type Domain struct {
	// Name is the domain name (e.g., "example.com").
	// This is an identifying property.
	Name string

	// Registrar is the domain registrar (optional).
	Registrar string

	// CreatedAt is the domain registration date (optional).
	CreatedAt string

	// ExpiresAt is the domain expiration date (optional).
	ExpiresAt string

	// Nameservers is the list of nameservers for this domain (optional).
	Nameservers []string

	// Status is the domain status (e.g., "active", "expired") (optional).
	Status string
}

// NodeType returns the canonical node type for domains.
func (d *Domain) NodeType() string {
	return graphrag.NodeTypeDomain
}

// IdentifyingProperties returns the properties that uniquely identify this domain.
// A domain is identified by its name.
func (d *Domain) IdentifyingProperties() map[string]any {
	return map[string]any{
		"name": d.Name,
	}
}

// Properties returns all properties to set on the domain node.
func (d *Domain) Properties() map[string]any {
	props := map[string]any{
		"name": d.Name,
	}

	// Add optional properties if present
	if d.Registrar != "" {
		props["registrar"] = d.Registrar
	}
	if d.CreatedAt != "" {
		props["created_at"] = d.CreatedAt
	}
	if d.ExpiresAt != "" {
		props["expires_at"] = d.ExpiresAt
	}
	if d.Nameservers != nil && len(d.Nameservers) > 0 {
		props["nameservers"] = d.Nameservers
	}
	if d.Status != "" {
		props["status"] = d.Status
	}

	return props
}

// ParentRef returns nil because Domain is a root node.
func (d *Domain) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Domain has no parent.
func (d *Domain) RelationshipType() string {
	return ""
}

package domain

import "github.com/zero-day-ai/sdk/graphrag"

// Subdomain represents a subdomain discovered under a root domain.
// Subdomains are hierarchical entities that belong to a parent domain.
//
// Hierarchy: Domain -> Subdomain
//
// Identifying Properties: parent_domain, name
// Parent: Domain (via HAS_SUBDOMAIN relationship)
//
// Example:
//
//	subdomain := &Subdomain{
//	    ParentDomain: "example.com",
//	    Name:         "api.example.com",
//	    RecordType:   "A",
//	    RecordValue:  "192.168.1.1",
//	}
type Subdomain struct {
	// ParentDomain is the name of the parent domain (e.g., "example.com").
	// This is an identifying property.
	ParentDomain string `json:"parent_domain"`

	// Name is the full subdomain name (e.g., "api.example.com").
	// This is an identifying property.
	Name string `json:"name"`

	// RecordType is the DNS record type (A, AAAA, CNAME, etc.) (optional).
	RecordType string `json:"record_type,omitempty"`

	// RecordValue is the DNS record value (IP address, CNAME target, etc.) (optional).
	RecordValue string `json:"record_value,omitempty"`

	// TTL is the DNS time-to-live value (optional).
	TTL int `json:"ttl,omitempty"`

	// Status is the subdomain status (e.g., "active", "inactive") (optional).
	Status string `json:"status,omitempty"`
}

// NodeType returns the canonical node type for subdomains.
func (s *Subdomain) NodeType() string {
	return graphrag.NodeTypeSubdomain
}

// IdentifyingProperties returns the properties that uniquely identify this subdomain.
// A subdomain is identified by its parent domain and name.
func (s *Subdomain) IdentifyingProperties() map[string]any {
	return map[string]any{
		"parent_domain": s.ParentDomain,
		"name":          s.Name,
	}
}

// Properties returns all properties to set on the subdomain node.
func (s *Subdomain) Properties() map[string]any {
	props := map[string]any{
		"parent_domain": s.ParentDomain,
		"name":          s.Name,
	}

	// Add optional properties if present
	if s.RecordType != "" {
		props["record_type"] = s.RecordType
	}
	if s.RecordValue != "" {
		props["record_value"] = s.RecordValue
	}
	if s.TTL != 0 {
		props["ttl"] = s.TTL
	}
	if s.Status != "" {
		props["status"] = s.Status
	}

	return props
}

// ParentRef returns a reference to the parent Domain node.
func (s *Subdomain) ParentRef() *NodeRef {
	if s.ParentDomain == "" {
		return nil
	}

	return &NodeRef{
		NodeType: graphrag.NodeTypeDomain,
		Properties: map[string]any{
			"name": s.ParentDomain,
		},
	}
}

// RelationshipType returns the relationship type to the parent domain.
func (s *Subdomain) RelationshipType() string {
	return graphrag.RelTypeHasSubdomain
}

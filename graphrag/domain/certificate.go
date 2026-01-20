package domain

import "github.com/zero-day-ai/sdk/graphrag"

// Certificate represents a TLS/SSL certificate discovered on a host or service.
// Certificates can be associated with hosts via SERVES_CERTIFICATE relationships.
//
// Hierarchy: Certificate is a root node (no parent)
//
// Identifying Properties: fingerprint
// Parent: None (root node)
//
// Example:
//
//	certificate := &Certificate{
//	    Fingerprint: "SHA256:1234567890abcdef...",
//	    Subject:     "CN=example.com",
//	    Issuer:      "CN=Let's Encrypt Authority X3",
//	    NotBefore:   "2024-01-01T00:00:00Z",
//	    NotAfter:    "2025-01-01T00:00:00Z",
//	}
type Certificate struct {
	// Fingerprint is the certificate fingerprint (SHA256 hash).
	// This is an identifying property.
	Fingerprint string

	// Subject is the certificate subject (DN) (optional).
	Subject string

	// Issuer is the certificate issuer (DN) (optional).
	Issuer string

	// NotBefore is the certificate valid start date (optional).
	NotBefore string

	// NotAfter is the certificate expiration date (optional).
	NotAfter string

	// SerialNumber is the certificate serial number (optional).
	SerialNumber string

	// SubjectAltNames is the list of subject alternative names (optional).
	SubjectAltNames []string

	// SignatureAlgorithm is the signature algorithm (e.g., "SHA256-RSA") (optional).
	SignatureAlgorithm string

	// KeySize is the public key size in bits (optional).
	KeySize int

	// SelfSigned indicates if the certificate is self-signed (optional).
	SelfSigned bool
}

// NodeType returns the canonical node type for certificates.
func (c *Certificate) NodeType() string {
	return graphrag.NodeTypeCertificate
}

// IdentifyingProperties returns the properties that uniquely identify this certificate.
// A certificate is identified by its fingerprint.
func (c *Certificate) IdentifyingProperties() map[string]any {
	return map[string]any{
		"fingerprint": c.Fingerprint,
	}
}

// Properties returns all properties to set on the certificate node.
func (c *Certificate) Properties() map[string]any {
	props := map[string]any{
		"fingerprint": c.Fingerprint,
	}

	// Add optional properties if present
	if c.Subject != "" {
		props["subject"] = c.Subject
	}
	if c.Issuer != "" {
		props["issuer"] = c.Issuer
	}
	if c.NotBefore != "" {
		props["not_before"] = c.NotBefore
	}
	if c.NotAfter != "" {
		props["not_after"] = c.NotAfter
	}
	if c.SerialNumber != "" {
		props["serial_number"] = c.SerialNumber
	}
	if c.SubjectAltNames != nil && len(c.SubjectAltNames) > 0 {
		props["subject_alt_names"] = c.SubjectAltNames
	}
	if c.SignatureAlgorithm != "" {
		props["signature_algorithm"] = c.SignatureAlgorithm
	}
	if c.KeySize != 0 {
		props["key_size"] = c.KeySize
	}
	props["self_signed"] = c.SelfSigned

	return props
}

// ParentRef returns nil because Certificate is a root node.
func (c *Certificate) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Certificate has no parent.
func (c *Certificate) RelationshipType() string {
	return ""
}

// Package domain provides domain types for the graphrag knowledge graph.
package domain

import (
	"github.com/zero-day-ai/sdk/api/gen/graphragpb"
	"github.com/zero-day-ai/sdk/api/gen/taxonomypb"
)

// DiscoveryResult is a wrapper around the proto DiscoveryResult that provides
// convenience methods for accessing discovered entities.
type DiscoveryResult struct {
	// Proto is the underlying proto message.
	Proto *graphragpb.DiscoveryResult

	// Additional nodes collected (domain types implementing GraphNode).
	nodes []GraphNode
}

// NewDiscoveryResult creates a new DiscoveryResult from a proto message.
func NewDiscoveryResult(proto *graphragpb.DiscoveryResult) *DiscoveryResult {
	if proto == nil {
		return &DiscoveryResult{}
	}
	return &DiscoveryResult{Proto: proto}
}

// IsEmpty returns true if the discovery result contains no entities.
func (d *DiscoveryResult) IsEmpty() bool {
	if d == nil || d.Proto == nil {
		return len(d.nodes) == 0
	}
	return d.NodeCount() == 0
}

// NodeCount returns the total count of all discovered entities.
func (d *DiscoveryResult) NodeCount() int {
	if d == nil || d.Proto == nil {
		return len(d.nodes)
	}

	count := len(d.Proto.Hosts) +
		len(d.Proto.Ports) +
		len(d.Proto.Services) +
		len(d.Proto.Endpoints) +
		len(d.Proto.Domains) +
		len(d.Proto.Subdomains) +
		len(d.Proto.Technologies) +
		len(d.Proto.Certificates) +
		len(d.Proto.Findings) +
		len(d.Proto.Evidence) +
		len(d.Proto.CustomNodes)

	return count + len(d.nodes)
}

// AddNode adds a domain node to the discovery result.
func (d *DiscoveryResult) AddNode(node GraphNode) {
	d.nodes = append(d.nodes, node)
}

// AllNodes returns all discovered entities as GraphNode instances,
// ordered so that parent nodes come before their children.
func (d *DiscoveryResult) AllNodes() []GraphNode {
	var nodes []GraphNode

	if d.Proto != nil {
		// Add hosts first (root nodes)
		for _, h := range d.Proto.Hosts {
			nodes = append(nodes, hostFromProto(h))
		}

		// Add domains (root nodes)
		for _, dom := range d.Proto.Domains {
			nodes = append(nodes, domainFromProto(dom))
		}

		// Add ports (children of hosts)
		for _, p := range d.Proto.Ports {
			nodes = append(nodes, portFromProto(p))
		}

		// Add subdomains (children of domains)
		for _, s := range d.Proto.Subdomains {
			nodes = append(nodes, subdomainFromProto(s))
		}

		// Add services (children of ports)
		for _, svc := range d.Proto.Services {
			nodes = append(nodes, serviceFromProto(svc))
		}

		// Add endpoints (children of services)
		for _, ep := range d.Proto.Endpoints {
			nodes = append(nodes, endpointFromProto(ep))
		}

		// Add technologies (can be attached to various parents)
		for _, t := range d.Proto.Technologies {
			nodes = append(nodes, technologyFromProto(t))
		}

		// Add certificates
		for _, c := range d.Proto.Certificates {
			nodes = append(nodes, certificateFromProto(c))
		}

		// Add findings
		for _, f := range d.Proto.Findings {
			nodes = append(nodes, findingFromProto(f))
		}

		// Add evidence (children of findings)
		for _, e := range d.Proto.Evidence {
			nodes = append(nodes, evidenceFromProto(e))
		}

		// Add custom nodes
		for _, cn := range d.Proto.CustomNodes {
			nodes = append(nodes, customNodeFromProto(cn))
		}
	}

	// Add any additional domain nodes
	nodes = append(nodes, d.nodes...)

	return nodes
}

// Helper functions to convert proto messages to domain types.
// Note: These create nodes without parent references set. The loader
// uses the host_id, port_id etc. fields from the proto to create relationships.

func hostFromProto(h *graphragpb.Host) GraphNode {
	host := NewHost().SetIp(h.Ip)
	if h.Hostname != nil {
		host.SetHostname(*h.Hostname)
	}
	if h.State != nil {
		host.SetState(*h.State)
	}
	if h.Os != nil {
		host.SetOs(*h.Os)
	}
	if h.OsVersion != nil {
		host.SetOsVersion(*h.OsVersion)
	}
	if h.MacAddress != nil {
		host.SetMacAddress(*h.MacAddress)
	}
	return host
}

func domainFromProto(d *graphragpb.Domain) GraphNode {
	dom := NewDomain(d.Name)
	if d.Registrar != nil {
		dom.SetRegistrar(*d.Registrar)
	}
	if d.CreatedDate != nil {
		dom.SetCreatedDate(*d.CreatedDate)
	}
	if d.ExpiryDate != nil {
		dom.SetExpiryDate(*d.ExpiryDate)
	}
	return dom
}

func portFromProto(p *graphragpb.Port) GraphNode {
	port := NewPort(p.Number, p.Protocol)
	if p.State != nil {
		port.SetState(*p.State)
	}
	if p.Reason != nil {
		port.SetReason(*p.Reason)
	}
	// Note: host_id is stored in the Properties() via parent reference handling
	// The loader creates the HAS_PORT relationship using the host_id
	return &protoPort{port: port, hostId: p.HostId}
}

func subdomainFromProto(s *graphragpb.Subdomain) GraphNode {
	sub := NewSubdomain(s.Name)
	if s.FullName != nil {
		sub.SetFullName(*s.FullName)
	}
	// Note: parent_domain is stored via parent reference
	return &protoSubdomain{subdomain: sub, parentDomain: s.ParentDomain}
}

func serviceFromProto(svc *graphragpb.Service) GraphNode {
	service := NewService(svc.Name)
	if svc.Product != nil {
		service.SetProduct(*svc.Product)
	}
	if svc.Version != nil {
		service.SetVersion(*svc.Version)
	}
	if svc.ExtraInfo != nil {
		service.SetExtraInfo(*svc.ExtraInfo)
	}
	if svc.Banner != nil {
		service.SetBanner(*svc.Banner)
	}
	if svc.Cpe != nil {
		service.SetCpe(*svc.Cpe)
	}
	return &protoService{service: service, portId: svc.PortId}
}

func endpointFromProto(ep *graphragpb.Endpoint) GraphNode {
	endpoint := NewEndpoint(ep.Url)
	if ep.Method != nil {
		endpoint.SetMethod(*ep.Method)
	}
	if ep.StatusCode != nil {
		endpoint.SetStatusCode(*ep.StatusCode)
	}
	if ep.ContentType != nil {
		endpoint.SetContentType(*ep.ContentType)
	}
	if ep.ContentLength != nil {
		endpoint.SetContentLength(*ep.ContentLength)
	}
	if ep.Title != nil {
		endpoint.SetTitle(*ep.Title)
	}
	return &protoEndpoint{endpoint: endpoint, serviceId: ep.ServiceId}
}

func technologyFromProto(t *graphragpb.Technology) GraphNode {
	tech := NewTechnology(t.Name)
	if t.Version != nil {
		tech.SetVersion(*t.Version)
	}
	if t.Category != nil {
		tech.SetCategory(*t.Category)
	}
	if t.Confidence != nil {
		tech.SetConfidence(*t.Confidence)
	}
	if t.Cpe != nil {
		tech.SetCpe(*t.Cpe)
	}
	return tech
}

func certificateFromProto(c *graphragpb.Certificate) GraphNode {
	cert := NewCertificate()
	if c.Subject != nil {
		cert.SetSubject(*c.Subject)
	}
	if c.Issuer != nil {
		cert.SetIssuer(*c.Issuer)
	}
	if c.SerialNumber != nil {
		cert.SetSerialNumber(*c.SerialNumber)
	}
	if c.NotBefore != nil {
		cert.SetNotBefore(*c.NotBefore)
	}
	if c.NotAfter != nil {
		cert.SetNotAfter(*c.NotAfter)
	}
	if c.FingerprintSha256 != nil {
		cert.SetFingerprintSha256(*c.FingerprintSha256)
	}
	if c.San != nil {
		cert.SetSan(*c.San)
	}
	return cert
}

func findingFromProto(f *graphragpb.Finding) GraphNode {
	finding := NewFinding(f.Title, f.Severity)
	if f.Description != nil {
		finding.SetDescription(*f.Description)
	}
	if f.Confidence != nil {
		finding.SetConfidence(*f.Confidence)
	}
	if f.Category != nil {
		finding.SetCategory(*f.Category)
	}
	if f.Remediation != nil {
		finding.SetRemediation(*f.Remediation)
	}
	if f.CvssScore != nil {
		finding.SetCvssScore(*f.CvssScore)
	}
	if f.CveIds != nil {
		finding.SetCveIds(*f.CveIds)
	}
	return finding
}

func evidenceFromProto(e *graphragpb.Evidence) GraphNode {
	evidence := NewEvidence(e.Type)
	if e.Content != nil {
		evidence.SetContent(*e.Content)
	}
	if e.Url != nil {
		evidence.SetUrl(*e.Url)
	}
	return &protoEvidence{evidence: evidence, findingId: e.FindingId}
}

func customNodeFromProto(cn *graphragpb.CustomNode) GraphNode {
	props := make(map[string]any)

	// Copy ID properties
	for k, v := range cn.IdProperties {
		props[k] = v
	}

	// Copy additional properties
	for k, v := range cn.Properties {
		props[k] = v
	}

	var parentRef *NodeRef
	if cn.ParentType != nil && len(cn.ParentId) > 0 {
		parentProps := make(map[string]any)
		for k, v := range cn.ParentId {
			parentProps[k] = v
		}
		relType := "BELONGS_TO"
		if cn.RelationshipType != nil {
			relType = *cn.RelationshipType
		}
		parentRef = &NodeRef{
			NodeType:     *cn.ParentType,
			Properties:   parentProps,
			Relationship: relType,
		}
	}

	return &CustomNode{
		nodeType:  cn.NodeType,
		props:     props,
		parentRef: parentRef,
	}
}

// ==================== WRAPPER TYPES ====================
// These types wrap domain types to add parent reference information from proto.

// protoPort wraps Port with host_id for parent relationship.
type protoPort struct {
	port   *Port
	hostId string
}

func (p *protoPort) NodeType() string              { return p.port.NodeType() }
func (p *protoPort) Properties() map[string]any    { return p.port.Properties() }
func (p *protoPort) IdentifyingProperties() map[string]any { return p.port.IdentifyingProperties() }
func (p *protoPort) ParentRef() *NodeRef {
	if p.hostId == "" {
		return nil
	}
	return &NodeRef{
		NodeType:     "host",
		Properties:   map[string]any{"ip": p.hostId},
		Relationship: "HAS_PORT",
	}
}
func (p *protoPort) Validate() error                     { return nil } // Skip validation for proto-sourced nodes
func (p *protoPort) ToProto() *taxonomypb.GraphNode      { return p.port.ToProto() }
func (p *protoPort) ID() string                          { return p.port.ID() }
func (p *protoPort) SetID(id string)                     { p.port.SetID(id) }

// protoSubdomain wraps Subdomain with parent_domain for parent relationship.
type protoSubdomain struct {
	subdomain    *Subdomain
	parentDomain string
}

func (s *protoSubdomain) NodeType() string              { return s.subdomain.NodeType() }
func (s *protoSubdomain) Properties() map[string]any    { return s.subdomain.Properties() }
func (s *protoSubdomain) IdentifyingProperties() map[string]any { return s.subdomain.IdentifyingProperties() }
func (s *protoSubdomain) ParentRef() *NodeRef {
	if s.parentDomain == "" {
		return nil
	}
	return &NodeRef{
		NodeType:     "domain",
		Properties:   map[string]any{"name": s.parentDomain},
		Relationship: "HAS_SUBDOMAIN",
	}
}
func (s *protoSubdomain) Validate() error                     { return nil }
func (s *protoSubdomain) ToProto() *taxonomypb.GraphNode      { return s.subdomain.ToProto() }
func (s *protoSubdomain) ID() string                          { return s.subdomain.ID() }
func (s *protoSubdomain) SetID(id string)                     { s.subdomain.SetID(id) }

// protoService wraps Service with port_id for parent relationship.
type protoService struct {
	service *Service
	portId  string
}

func (s *protoService) NodeType() string              { return s.service.NodeType() }
func (s *protoService) Properties() map[string]any    { return s.service.Properties() }
func (s *protoService) IdentifyingProperties() map[string]any { return s.service.IdentifyingProperties() }
func (s *protoService) ParentRef() *NodeRef {
	if s.portId == "" {
		return nil
	}
	return &NodeRef{
		NodeType:     "port",
		Properties:   map[string]any{"id": s.portId},
		Relationship: "RUNS_SERVICE",
	}
}
func (s *protoService) Validate() error                     { return nil }
func (s *protoService) ToProto() *taxonomypb.GraphNode      { return s.service.ToProto() }
func (s *protoService) ID() string                          { return s.service.ID() }
func (s *protoService) SetID(id string)                     { s.service.SetID(id) }

// protoEndpoint wraps Endpoint with service_id for parent relationship.
type protoEndpoint struct {
	endpoint  *Endpoint
	serviceId string
}

func (e *protoEndpoint) NodeType() string              { return e.endpoint.NodeType() }
func (e *protoEndpoint) Properties() map[string]any    { return e.endpoint.Properties() }
func (e *protoEndpoint) IdentifyingProperties() map[string]any { return e.endpoint.IdentifyingProperties() }
func (e *protoEndpoint) ParentRef() *NodeRef {
	if e.serviceId == "" {
		return nil
	}
	return &NodeRef{
		NodeType:     "service",
		Properties:   map[string]any{"id": e.serviceId},
		Relationship: "HAS_ENDPOINT",
	}
}
func (e *protoEndpoint) Validate() error                     { return nil }
func (e *protoEndpoint) ToProto() *taxonomypb.GraphNode      { return e.endpoint.ToProto() }
func (e *protoEndpoint) ID() string                          { return e.endpoint.ID() }
func (e *protoEndpoint) SetID(id string)                     { e.endpoint.SetID(id) }

// protoEvidence wraps Evidence with finding_id for parent relationship.
type protoEvidence struct {
	evidence  *Evidence
	findingId string
}

func (e *protoEvidence) NodeType() string              { return e.evidence.NodeType() }
func (e *protoEvidence) Properties() map[string]any    { return e.evidence.Properties() }
func (e *protoEvidence) IdentifyingProperties() map[string]any { return e.evidence.IdentifyingProperties() }
func (e *protoEvidence) ParentRef() *NodeRef {
	if e.findingId == "" {
		return nil
	}
	return &NodeRef{
		NodeType:     "finding",
		Properties:   map[string]any{"id": e.findingId},
		Relationship: "HAS_EVIDENCE",
	}
}
func (e *protoEvidence) Validate() error                     { return nil }
func (e *protoEvidence) ToProto() *taxonomypb.GraphNode      { return e.evidence.ToProto() }
func (e *protoEvidence) ID() string                          { return e.evidence.ID() }
func (e *protoEvidence) SetID(id string)                     { e.evidence.SetID(id) }

// CustomNode represents a node type not in the core taxonomy.
type CustomNode struct {
	id        string
	nodeType  string
	props     map[string]any
	parentRef *NodeRef
}

// NodeType implements GraphNode.
func (n *CustomNode) NodeType() string { return n.nodeType }

// Properties implements GraphNode.
func (n *CustomNode) Properties() map[string]any { return n.props }

// IdentifyingProperties implements GraphNode.
func (n *CustomNode) IdentifyingProperties() map[string]any { return n.props }

// ParentRef implements GraphNode.
func (n *CustomNode) ParentRef() *NodeRef { return n.parentRef }

// Validate implements GraphNode. Custom nodes skip validation.
func (n *CustomNode) Validate() error { return nil }

// ToProto implements GraphNode.
func (n *CustomNode) ToProto() *taxonomypb.GraphNode {
	return &taxonomypb.GraphNode{
		Id:         n.id,
		Type:       n.nodeType,
		Properties: propsToValueMap(n.props),
	}
}

// ID implements GraphNode.
func (n *CustomNode) ID() string { return n.id }

// SetID implements GraphNode.
func (n *CustomNode) SetID(id string) { n.id = id }

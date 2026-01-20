package domain

// DiscoveryResult is a standardized container for tool and agent output.
// It provides typed slices for each canonical domain type, plus a flexible
// Custom slice for agent-defined types.
//
// Tools should populate the appropriate slices with discovered assets,
// and the GraphRAG loader will automatically create nodes and relationships
// in the knowledge graph.
//
// Example (Network Reconnaissance Tool):
//
//	result := &DiscoveryResult{
//	    Hosts: []*Host{
//	        {IP: "192.168.1.1", Hostname: "gateway", State: "up"},
//	        {IP: "192.168.1.10", Hostname: "web-server", State: "up"},
//	    },
//	    Ports: []*Port{
//	        {HostID: "192.168.1.10", Number: 80, Protocol: "tcp", State: "open"},
//	        {HostID: "192.168.1.10", Number: 443, Protocol: "tcp", State: "open"},
//	    },
//	    Services: []*Service{
//	        {PortID: "192.168.1.10:80:tcp", Name: "http", Version: "nginx/1.18.0"},
//	        {PortID: "192.168.1.10:443:tcp", Name: "https", Version: "nginx/1.18.0"},
//	    },
//	}
//
// Example (Custom Kubernetes Agent):
//
//	result := &DiscoveryResult{
//	    Custom: []GraphNode{
//	        NewCustomEntity("k8s", "pod").
//	            WithIDProps(map[string]any{"namespace": "default", "name": "web-01"}),
//	        NewCustomEntity("k8s", "service").
//	            WithIDProps(map[string]any{"namespace": "default", "name": "web-svc"}),
//	    },
//	}
type DiscoveryResult struct {
	// Hosts contains discovered network hosts (IP addresses).
	Hosts []*Host

	// Ports contains discovered network ports on hosts.
	Ports []*Port

	// Services contains discovered services running on ports.
	Services []*Service

	// Endpoints contains discovered web endpoints or URLs.
	Endpoints []*Endpoint

	// Domains contains discovered root domain names.
	Domains []*Domain

	// Subdomains contains discovered subdomains under root domains.
	Subdomains []*Subdomain

	// Technologies contains discovered technologies, frameworks, or software.
	Technologies []*Technology

	// Certificates contains discovered TLS/SSL certificates.
	Certificates []*Certificate

	// CloudAssets contains discovered cloud infrastructure resources.
	CloudAssets []*CloudAsset

	// APIs contains discovered web API services.
	APIs []*API

	// Custom contains custom agent-defined graph nodes.
	// Use this for domain-specific types like "k8s:pod" or "aws:security_group".
	Custom []GraphNode
}

// AllNodes returns all discovered nodes as a flattened slice of GraphNode interfaces.
// This is the primary method used by the GraphRAG loader to process all discoveries.
//
// Nodes are returned in dependency order:
//  1. Root nodes (Hosts, Domains, Technologies, Certificates, CloudAssets, APIs)
//  2. Dependent nodes (Ports, Subdomains)
//  3. Further dependent nodes (Services)
//  4. Leaf nodes (Endpoints)
//  5. Custom nodes (order preserved as added)
//
// This ordering ensures parent nodes are created before child nodes reference them.
func (d *DiscoveryResult) AllNodes() []GraphNode {
	var nodes []GraphNode

	// Add root nodes first (no parent dependencies)
	for _, host := range d.Hosts {
		nodes = append(nodes, host)
	}
	for _, domain := range d.Domains {
		nodes = append(nodes, domain)
	}
	for _, tech := range d.Technologies {
		nodes = append(nodes, tech)
	}
	for _, cert := range d.Certificates {
		nodes = append(nodes, cert)
	}
	for _, cloud := range d.CloudAssets {
		nodes = append(nodes, cloud)
	}
	for _, api := range d.APIs {
		nodes = append(nodes, api)
	}

	// Add first-level dependent nodes
	for _, port := range d.Ports {
		nodes = append(nodes, port)
	}
	for _, subdomain := range d.Subdomains {
		nodes = append(nodes, subdomain)
	}

	// Add second-level dependent nodes
	for _, service := range d.Services {
		nodes = append(nodes, service)
	}

	// Add leaf nodes
	for _, endpoint := range d.Endpoints {
		nodes = append(nodes, endpoint)
	}

	// Add custom nodes (preserve order)
	nodes = append(nodes, d.Custom...)

	return nodes
}

// IsEmpty returns true if the discovery result contains no nodes.
func (d *DiscoveryResult) IsEmpty() bool {
	return len(d.Hosts) == 0 &&
		len(d.Ports) == 0 &&
		len(d.Services) == 0 &&
		len(d.Endpoints) == 0 &&
		len(d.Domains) == 0 &&
		len(d.Subdomains) == 0 &&
		len(d.Technologies) == 0 &&
		len(d.Certificates) == 0 &&
		len(d.CloudAssets) == 0 &&
		len(d.APIs) == 0 &&
		len(d.Custom) == 0
}

// NodeCount returns the total number of nodes in this discovery result.
func (d *DiscoveryResult) NodeCount() int {
	return len(d.Hosts) +
		len(d.Ports) +
		len(d.Services) +
		len(d.Endpoints) +
		len(d.Domains) +
		len(d.Subdomains) +
		len(d.Technologies) +
		len(d.Certificates) +
		len(d.CloudAssets) +
		len(d.APIs) +
		len(d.Custom)
}

// NewDiscoveryResult creates an empty DiscoveryResult with all slices initialized.
func NewDiscoveryResult() *DiscoveryResult {
	return &DiscoveryResult{
		Hosts:        make([]*Host, 0),
		Ports:        make([]*Port, 0),
		Services:     make([]*Service, 0),
		Endpoints:    make([]*Endpoint, 0),
		Domains:      make([]*Domain, 0),
		Subdomains:   make([]*Subdomain, 0),
		Technologies: make([]*Technology, 0),
		Certificates: make([]*Certificate, 0),
		CloudAssets:  make([]*CloudAsset, 0),
		APIs:         make([]*API, 0),
		Custom:       make([]GraphNode, 0),
	}
}

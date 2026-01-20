package domain

import "github.com/zero-day-ai/sdk/graphrag"

// DNSRecord represents a DNS record for a domain.
// DNS records provide the mapping between domain names and various resource types.
//
// Example:
//
//	record := &DNSRecord{
//	    Domain:     "example.com",
//	    RecordType: "A",
//	    Value:      "192.168.1.1",
//	    TTL:        3600,
//	}
//
// Identifying Properties:
//   - domain (required): The domain name this record belongs to
//   - record_type (required): Type of DNS record (A, AAAA, CNAME, MX, TXT, NS, SOA, PTR, SRV)
//   - value (required): The value/target of the DNS record
//
// Relationships:
//   - None (root node)
type DNSRecord struct {
	// Domain is the domain name this record belongs to (e.g., "example.com").
	// This is an identifying property and is required.
	Domain string

	// RecordType is the type of DNS record.
	// This is an identifying property and is required.
	// Common values: "A", "AAAA", "CNAME", "MX", "TXT", "NS", "SOA", "PTR", "SRV"
	RecordType string

	// Value is the target/value of the DNS record.
	// This is an identifying property and is required.
	// For A records: IP address
	// For CNAME records: target domain
	// For MX records: mail server
	// For TXT records: text value
	Value string

	// TTL is the time-to-live in seconds for this record.
	// Optional. Default is typically 3600 (1 hour).
	TTL int

	// Priority is the priority for MX and SRV records.
	// Optional. Lower values indicate higher priority.
	Priority int
}

func (d *DNSRecord) NodeType() string { return "dns_record" }

func (d *DNSRecord) IdentifyingProperties() map[string]any {
	return map[string]any{
		"domain":      d.Domain,
		"record_type": d.RecordType,
		"value":       d.Value,
	}
}

func (d *DNSRecord) Properties() map[string]any {
	props := d.IdentifyingProperties()
	if d.TTL > 0 {
		props["ttl"] = d.TTL
	}
	if d.Priority > 0 {
		props["priority"] = d.Priority
	}
	return props
}

func (d *DNSRecord) ParentRef() *NodeRef      { return nil }
func (d *DNSRecord) RelationshipType() string { return "" }

// Firewall represents a network firewall device or service.
// Firewalls control network traffic based on security rules.
//
// Example:
//
//	firewall := &Firewall{
//	    Name:   "prod-firewall",
//	    Type:   "hardware",
//	    Vendor: "Palo Alto",
//	    Model:  "PA-5220",
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the firewall
//
// Relationships:
//   - None (root node)
//   - Children: FirewallRule nodes (via HAS_RULE relationship)
type Firewall struct {
	// Name is the unique identifier for this firewall.
	// This is an identifying property and is required.
	Name string

	// Type is the type of firewall.
	// Optional. Common values: "hardware", "software", "cloud", "virtual"
	Type string

	// Vendor is the manufacturer or provider of the firewall.
	// Optional. Example: "Palo Alto", "Cisco", "Fortinet", "AWS"
	Vendor string

	// Model is the specific model or SKU of the firewall.
	// Optional. Example: "PA-5220", "ASA 5525-X", "FortiGate 100F"
	Model string

	// Version is the firmware or software version.
	// Optional. Example: "10.1.3"
	Version string
}

func (f *Firewall) NodeType() string { return "firewall" }

func (f *Firewall) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: f.Name,
	}
}

func (f *Firewall) Properties() map[string]any {
	props := f.IdentifyingProperties()
	if f.Type != "" {
		props["type"] = f.Type
	}
	if f.Vendor != "" {
		props["vendor"] = f.Vendor
	}
	if f.Model != "" {
		props["model"] = f.Model
	}
	if f.Version != "" {
		props["version"] = f.Version
	}
	return props
}

func (f *Firewall) ParentRef() *NodeRef      { return nil }
func (f *Firewall) RelationshipType() string { return "" }

// FirewallRule represents a single rule in a firewall configuration.
// Rules define allow/deny decisions for network traffic.
//
// Example:
//
//	rule := &FirewallRule{
//	    FirewallID: "prod-firewall",
//	    Name:       "allow-web-traffic",
//	    Action:     "allow",
//	    Source:     "0.0.0.0/0",
//	    Destination: "10.0.1.0/24",
//	    Port:       "80,443",
//	    Protocol:   "tcp",
//	}
//
// Identifying Properties:
//   - firewall_id (required): The firewall this rule belongs to
//   - name (required): Unique name of the rule within the firewall
//
// Relationships:
//   - Parent: Firewall node (via HAS_RULE relationship)
type FirewallRule struct {
	// FirewallID is the identifier of the parent firewall.
	// This is an identifying property and is required.
	FirewallID string

	// Name is the unique name of this rule within the firewall.
	// This is an identifying property and is required.
	Name string

	// Action is the action to take when traffic matches this rule.
	// Optional. Common values: "allow", "deny", "drop", "reject"
	Action string

	// Source is the source IP address, network, or zone.
	// Optional. Example: "192.168.1.0/24", "any", "internal-zone"
	Source string

	// Destination is the destination IP address, network, or zone.
	// Optional. Example: "10.0.0.0/8", "any", "dmz-zone"
	Destination string

	// Port is the port or port range.
	// Optional. Example: "80", "443", "8000-9000", "any"
	Port string

	// Protocol is the network protocol.
	// Optional. Common values: "tcp", "udp", "icmp", "any"
	Protocol string

	// Priority is the rule priority or order.
	// Optional. Lower numbers typically mean higher priority.
	Priority int

	// Enabled indicates if the rule is active.
	// Optional. Default: true
	Enabled bool
}

func (f *FirewallRule) NodeType() string { return "firewall_rule" }

func (f *FirewallRule) IdentifyingProperties() map[string]any {
	return map[string]any{
		"firewall_id":     f.FirewallID,
		graphrag.PropName: f.Name,
	}
}

func (f *FirewallRule) Properties() map[string]any {
	props := f.IdentifyingProperties()
	if f.Action != "" {
		props["action"] = f.Action
	}
	if f.Source != "" {
		props["source"] = f.Source
	}
	if f.Destination != "" {
		props["destination"] = f.Destination
	}
	if f.Port != "" {
		props[graphrag.PropPort] = f.Port
	}
	if f.Protocol != "" {
		props[graphrag.PropProtocol] = f.Protocol
	}
	if f.Priority > 0 {
		props["priority"] = f.Priority
	}
	props["enabled"] = f.Enabled
	return props
}

func (f *FirewallRule) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "firewall",
		Properties: map[string]any{
			graphrag.PropName: f.FirewallID,
		},
	}
}

func (f *FirewallRule) RelationshipType() string { return "HAS_RULE" }

// Router represents a network router device.
// Routers forward packets between networks based on routing tables.
//
// Example:
//
//	router := &Router{
//	    Name:   "core-router-01",
//	    Type:   "enterprise",
//	    Vendor: "Cisco",
//	    Model:  "Catalyst 9500",
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the router
//
// Relationships:
//   - None (root node)
//   - Children: Route nodes (via HAS_ROUTE relationship)
type Router struct {
	// Name is the unique identifier for this router.
	// This is an identifying property and is required.
	Name string

	// Type is the type of router.
	// Optional. Common values: "enterprise", "edge", "core", "distribution"
	Type string

	// Vendor is the manufacturer of the router.
	// Optional. Example: "Cisco", "Juniper", "Arista"
	Vendor string

	// Model is the specific model or SKU.
	// Optional. Example: "Catalyst 9500", "MX480"
	Model string

	// ManagementIP is the IP address used to manage this router.
	// Optional. Example: "10.0.0.1"
	ManagementIP string
}

func (r *Router) NodeType() string { return "router" }

func (r *Router) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: r.Name,
	}
}

func (r *Router) Properties() map[string]any {
	props := r.IdentifyingProperties()
	if r.Type != "" {
		props["type"] = r.Type
	}
	if r.Vendor != "" {
		props["vendor"] = r.Vendor
	}
	if r.Model != "" {
		props["model"] = r.Model
	}
	if r.ManagementIP != "" {
		props["management_ip"] = r.ManagementIP
	}
	return props
}

func (r *Router) ParentRef() *NodeRef      { return nil }
func (r *Router) RelationshipType() string { return "" }

// Route represents a routing table entry in a router.
// Routes define how packets should be forwarded to reach different networks.
//
// Example:
//
//	route := &Route{
//	    RouterID:    "core-router-01",
//	    Destination: "10.0.0.0/8",
//	    NextHop:     "192.168.1.1",
//	    Metric:      100,
//	    Protocol:    "static",
//	}
//
// Identifying Properties:
//   - router_id (required): The router this route belongs to
//   - destination (required): Destination network in CIDR notation
//
// Relationships:
//   - Parent: Router node (via HAS_ROUTE relationship)
type Route struct {
	// RouterID is the identifier of the parent router.
	// This is an identifying property and is required.
	RouterID string

	// Destination is the destination network in CIDR notation.
	// This is an identifying property and is required.
	// Example: "10.0.0.0/8", "0.0.0.0/0" (default route)
	Destination string

	// NextHop is the next hop IP address.
	// Optional. Example: "192.168.1.1"
	NextHop string

	// Interface is the outbound network interface.
	// Optional. Example: "eth0", "GigabitEthernet0/0/1"
	Interface string

	// Metric is the routing metric or cost.
	// Optional. Lower values are preferred.
	Metric int

	// Protocol is the routing protocol that learned this route.
	// Optional. Common values: "static", "bgp", "ospf", "rip", "connected"
	Protocol string

	// AdminDistance is the administrative distance for this route.
	// Optional. Lower values are more trusted.
	AdminDistance int
}

func (r *Route) NodeType() string { return "route" }

func (r *Route) IdentifyingProperties() map[string]any {
	return map[string]any{
		"router_id":   r.RouterID,
		"destination": r.Destination,
	}
}

func (r *Route) Properties() map[string]any {
	props := r.IdentifyingProperties()
	if r.NextHop != "" {
		props["next_hop"] = r.NextHop
	}
	if r.Interface != "" {
		props["interface"] = r.Interface
	}
	if r.Metric > 0 {
		props["metric"] = r.Metric
	}
	if r.Protocol != "" {
		props[graphrag.PropProtocol] = r.Protocol
	}
	if r.AdminDistance > 0 {
		props["admin_distance"] = r.AdminDistance
	}
	return props
}

func (r *Route) ParentRef() *NodeRef {
	return &NodeRef{
		NodeType: "router",
		Properties: map[string]any{
			graphrag.PropName: r.RouterID,
		},
	}
}

func (r *Route) RelationshipType() string { return "HAS_ROUTE" }

// LoadBalancer represents a load balancer device or service.
// Load balancers distribute traffic across multiple backend servers.
//
// Example:
//
//	lb := &LoadBalancer{
//	    Name:      "prod-lb-01",
//	    Type:      "application",
//	    Provider:  "AWS",
//	    Algorithm: "round-robin",
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the load balancer
//
// Relationships:
//   - None (root node)
type LoadBalancer struct {
	// Name is the unique identifier for this load balancer.
	// This is an identifying property and is required.
	Name string

	// Type is the type of load balancer.
	// Optional. Common values: "application", "network", "classic"
	Type string

	// Provider is the cloud provider or vendor.
	// Optional. Example: "AWS", "GCP", "Azure", "F5", "HAProxy"
	Provider string

	// Algorithm is the load balancing algorithm.
	// Optional. Common values: "round-robin", "least-connections", "ip-hash"
	Algorithm string

	// VIP is the virtual IP address clients connect to.
	// Optional. Example: "203.0.113.1"
	VIP string

	// State is the current operational state.
	// Optional. Common values: "active", "inactive", "provisioning"
	State string
}

func (l *LoadBalancer) NodeType() string { return "load_balancer" }

func (l *LoadBalancer) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: l.Name,
	}
}

func (l *LoadBalancer) Properties() map[string]any {
	props := l.IdentifyingProperties()
	if l.Type != "" {
		props["type"] = l.Type
	}
	if l.Provider != "" {
		props["provider"] = l.Provider
	}
	if l.Algorithm != "" {
		props["algorithm"] = l.Algorithm
	}
	if l.VIP != "" {
		props["vip"] = l.VIP
	}
	if l.State != "" {
		props[graphrag.PropState] = l.State
	}
	return props
}

func (l *LoadBalancer) ParentRef() *NodeRef      { return nil }
func (l *LoadBalancer) RelationshipType() string { return "" }

// Proxy represents a network proxy server.
// Proxies act as intermediaries for requests from clients seeking resources.
//
// Example:
//
//	proxy := &Proxy{
//	    Name:     "squid-proxy-01",
//	    Type:     "forward",
//	    Protocol: "http",
//	    Port:     3128,
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the proxy
//
// Relationships:
//   - None (root node)
type Proxy struct {
	// Name is the unique identifier for this proxy.
	// This is an identifying property and is required.
	Name string

	// Type is the type of proxy.
	// Optional. Common values: "forward", "reverse", "transparent", "socks"
	Type string

	// Protocol is the proxy protocol.
	// Optional. Common values: "http", "https", "socks4", "socks5"
	Protocol string

	// Host is the hostname or IP address of the proxy.
	// Optional. Example: "proxy.example.com", "10.0.0.50"
	Host string

	// Port is the port the proxy listens on.
	// Optional. Common values: 3128 (Squid), 8080, 1080 (SOCKS)
	Port int

	// AuthRequired indicates if authentication is required.
	// Optional. Default: false
	AuthRequired bool
}

func (p *Proxy) NodeType() string { return "proxy" }

func (p *Proxy) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: p.Name,
	}
}

func (p *Proxy) Properties() map[string]any {
	props := p.IdentifyingProperties()
	if p.Type != "" {
		props["type"] = p.Type
	}
	if p.Protocol != "" {
		props[graphrag.PropProtocol] = p.Protocol
	}
	if p.Host != "" {
		props["host"] = p.Host
	}
	if p.Port > 0 {
		props[graphrag.PropPort] = p.Port
	}
	props["auth_required"] = p.AuthRequired
	return props
}

func (p *Proxy) ParentRef() *NodeRef      { return nil }
func (p *Proxy) RelationshipType() string { return "" }

// VPN represents a Virtual Private Network service or tunnel.
// VPNs provide secure, encrypted connections over untrusted networks.
//
// Example:
//
//	vpn := &VPN{
//	    Name:     "remote-access-vpn",
//	    Type:     "site-to-site",
//	    Protocol: "ipsec",
//	    Endpoint: "vpn.example.com",
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the VPN
//
// Relationships:
//   - None (root node)
type VPN struct {
	// Name is the unique identifier for this VPN.
	// This is an identifying property and is required.
	Name string

	// Type is the type of VPN.
	// Optional. Common values: "site-to-site", "remote-access", "client-to-site"
	Type string

	// Protocol is the VPN protocol.
	// Optional. Common values: "ipsec", "openvpn", "wireguard", "l2tp", "pptp"
	Protocol string

	// Endpoint is the VPN gateway address.
	// Optional. Example: "vpn.example.com", "203.0.113.1"
	Endpoint string

	// Encryption is the encryption algorithm.
	// Optional. Example: "AES-256-GCM", "ChaCha20-Poly1305"
	Encryption string

	// State is the current connection state.
	// Optional. Common values: "connected", "disconnected", "connecting"
	State string
}

func (v *VPN) NodeType() string { return "vpn" }

func (v *VPN) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: v.Name,
	}
}

func (v *VPN) Properties() map[string]any {
	props := v.IdentifyingProperties()
	if v.Type != "" {
		props["type"] = v.Type
	}
	if v.Protocol != "" {
		props[graphrag.PropProtocol] = v.Protocol
	}
	if v.Endpoint != "" {
		props["endpoint"] = v.Endpoint
	}
	if v.Encryption != "" {
		props["encryption"] = v.Encryption
	}
	if v.State != "" {
		props[graphrag.PropState] = v.State
	}
	return props
}

func (v *VPN) ParentRef() *NodeRef      { return nil }
func (v *VPN) RelationshipType() string { return "" }

// Network represents a network segment or subnet.
// Networks define logical groupings of IP addresses.
//
// Example:
//
//	network := &Network{
//	    CIDR:    "10.0.1.0/24",
//	    Name:    "prod-web-subnet",
//	    Type:    "private",
//	    Gateway: "10.0.1.1",
//	}
//
// Identifying Properties:
//   - cidr (required): Network address in CIDR notation
//
// Relationships:
//   - None (root node)
type Network struct {
	// CIDR is the network address in CIDR notation.
	// This is an identifying property and is required.
	// Example: "10.0.1.0/24", "192.168.0.0/16"
	CIDR string

	// Name is the descriptive name of the network.
	// Optional. Example: "prod-web-subnet", "dmz-network"
	Name string

	// Type is the type of network.
	// Optional. Common values: "private", "public", "dmz", "management"
	Type string

	// Gateway is the default gateway IP address.
	// Optional. Example: "10.0.1.1"
	Gateway string

	// VLAN is the VLAN ID if applicable.
	// Optional. Example: 100
	VLAN int

	// Location is the physical or cloud region location.
	// Optional. Example: "datacenter-1", "us-east-1"
	Location string
}

func (n *Network) NodeType() string { return "network" }

func (n *Network) IdentifyingProperties() map[string]any {
	return map[string]any{
		"cidr": n.CIDR,
	}
}

func (n *Network) Properties() map[string]any {
	props := n.IdentifyingProperties()
	if n.Name != "" {
		props[graphrag.PropName] = n.Name
	}
	if n.Type != "" {
		props["type"] = n.Type
	}
	if n.Gateway != "" {
		props["gateway"] = n.Gateway
	}
	if n.VLAN > 0 {
		props["vlan"] = n.VLAN
	}
	if n.Location != "" {
		props["location"] = n.Location
	}
	return props
}

func (n *Network) ParentRef() *NodeRef      { return nil }
func (n *Network) RelationshipType() string { return "" }

// VLAN represents a Virtual Local Area Network.
// VLANs logically segment network traffic at Layer 2.
//
// Example:
//
//	vlan := &VLAN{
//	    ID:   100,
//	    Name: "prod-web-vlan",
//	    Type: "ethernet",
//	}
//
// Identifying Properties:
//   - id (required): VLAN ID (1-4094)
//
// Relationships:
//   - None (root node)
type VLAN struct {
	// ID is the VLAN identifier (1-4094).
	// This is an identifying property and is required.
	ID int

	// Name is the descriptive name of the VLAN.
	// Optional. Example: "prod-web-vlan", "guest-wifi"
	Name string

	// Type is the VLAN type.
	// Optional. Common values: "ethernet", "token-ring"
	Type string

	// Description is a detailed description of the VLAN purpose.
	// Optional.
	Description string
}

func (v *VLAN) NodeType() string { return "vlan" }

func (v *VLAN) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": v.ID,
	}
}

func (v *VLAN) Properties() map[string]any {
	props := v.IdentifyingProperties()
	if v.Name != "" {
		props[graphrag.PropName] = v.Name
	}
	if v.Type != "" {
		props["type"] = v.Type
	}
	if v.Description != "" {
		props[graphrag.PropDescription] = v.Description
	}
	return props
}

func (v *VLAN) ParentRef() *NodeRef      { return nil }
func (v *VLAN) RelationshipType() string { return "" }

// NetworkInterface represents a network interface card or virtual interface.
// Interfaces connect hosts to networks and have MAC addresses.
//
// Example:
//
//	iface := &NetworkInterface{
//	    Name:       "eth0",
//	    MAC:        "00:1A:2B:3C:4D:5E",
//	    HostID:     "192.168.1.100",
//	    IPAddress:  "192.168.1.100",
//	    State:      "up",
//	}
//
// Identifying Properties:
//   - name (required): Interface name
//   - mac (required): MAC address
//
// Relationships:
//   - Parent: Host node (via HAS_INTERFACE relationship)
type NetworkInterface struct {
	// Name is the interface name.
	// This is an identifying property and is required.
	// Example: "eth0", "ens33", "GigabitEthernet0/0"
	Name string

	// MAC is the MAC (hardware) address.
	// This is an identifying property and is required.
	// Example: "00:1A:2B:3C:4D:5E"
	MAC string

	// HostID is the identifier of the host this interface belongs to.
	// Optional but recommended for parent relationship.
	HostID string

	// IPAddress is the IP address assigned to this interface.
	// Optional. Example: "192.168.1.100"
	IPAddress string

	// Netmask is the subnet mask or prefix length.
	// Optional. Example: "255.255.255.0", "/24"
	Netmask string

	// State is the operational state of the interface.
	// Optional. Common values: "up", "down", "unknown"
	State string

	// Speed is the interface speed.
	// Optional. Example: "1000", "10000" (Mbps)
	Speed int

	// MTU is the Maximum Transmission Unit.
	// Optional. Default: 1500
	MTU int
}

func (n *NetworkInterface) NodeType() string { return "network_interface" }

func (n *NetworkInterface) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: n.Name,
		"mac":             n.MAC,
	}
}

func (n *NetworkInterface) Properties() map[string]any {
	props := n.IdentifyingProperties()
	if n.HostID != "" {
		props[graphrag.PropHostID] = n.HostID
	}
	if n.IPAddress != "" {
		props["ip_address"] = n.IPAddress
	}
	if n.Netmask != "" {
		props["netmask"] = n.Netmask
	}
	if n.State != "" {
		props[graphrag.PropState] = n.State
	}
	if n.Speed > 0 {
		props["speed"] = n.Speed
	}
	if n.MTU > 0 {
		props["mtu"] = n.MTU
	}
	return props
}

func (n *NetworkInterface) ParentRef() *NodeRef {
	if n.HostID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeHost,
		Properties: map[string]any{
			graphrag.PropIP: n.HostID,
		},
	}
}

func (n *NetworkInterface) RelationshipType() string {
	if n.HostID == "" {
		return ""
	}
	return "HAS_INTERFACE"
}

// NetworkZone represents a logical security zone or network segment.
// Zones group assets with similar security requirements.
//
// Example:
//
//	zone := &NetworkZone{
//	    Name:        "dmz",
//	    Description: "Demilitarized zone for public-facing services",
//	    TrustLevel:  "untrusted",
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the zone
//
// Relationships:
//   - None (root node)
type NetworkZone struct {
	// Name is the unique identifier for this zone.
	// This is an identifying property and is required.
	// Example: "dmz", "internal", "external", "trust", "untrust"
	Name string

	// Description is a detailed description of the zone.
	// Optional.
	Description string

	// TrustLevel indicates the security trust level.
	// Optional. Common values: "trusted", "untrusted", "neutral"
	TrustLevel string

	// Type is the type of zone.
	// Optional. Common values: "security", "functional", "compliance"
	Type string
}

func (n *NetworkZone) NodeType() string { return "network_zone" }

func (n *NetworkZone) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: n.Name,
	}
}

func (n *NetworkZone) Properties() map[string]any {
	props := n.IdentifyingProperties()
	if n.Description != "" {
		props[graphrag.PropDescription] = n.Description
	}
	if n.TrustLevel != "" {
		props["trust_level"] = n.TrustLevel
	}
	if n.Type != "" {
		props["type"] = n.Type
	}
	return props
}

func (n *NetworkZone) ParentRef() *NodeRef      { return nil }
func (n *NetworkZone) RelationshipType() string { return "" }

// NetworkACL represents a Network Access Control List.
// ACLs define rules for allowing or denying network traffic.
//
// Example:
//
//	acl := &NetworkACL{
//	    Name:        "subnet-acl-01",
//	    Type:        "stateless",
//	    Scope:       "subnet",
//	    DefaultRule: "deny",
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the ACL
//
// Relationships:
//   - None (root node)
type NetworkACL struct {
	// Name is the unique identifier for this ACL.
	// This is an identifying property and is required.
	Name string

	// Type is the type of ACL.
	// Optional. Common values: "stateless", "stateful"
	Type string

	// Scope is the scope of the ACL.
	// Optional. Common values: "subnet", "interface", "global"
	Scope string

	// DefaultRule is the default action when no rule matches.
	// Optional. Common values: "allow", "deny"
	DefaultRule string

	// Description is a detailed description of the ACL.
	// Optional.
	Description string
}

func (n *NetworkACL) NodeType() string { return "network_acl" }

func (n *NetworkACL) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: n.Name,
	}
}

func (n *NetworkACL) Properties() map[string]any {
	props := n.IdentifyingProperties()
	if n.Type != "" {
		props["type"] = n.Type
	}
	if n.Scope != "" {
		props["scope"] = n.Scope
	}
	if n.DefaultRule != "" {
		props["default_rule"] = n.DefaultRule
	}
	if n.Description != "" {
		props[graphrag.PropDescription] = n.Description
	}
	return props
}

func (n *NetworkACL) ParentRef() *NodeRef      { return nil }
func (n *NetworkACL) RelationshipType() string { return "" }

// NATGateway represents a Network Address Translation gateway.
// NAT gateways enable private networks to access the internet.
//
// Example:
//
//	nat := &NATGateway{
//	    Name:      "nat-gateway-01",
//	    Type:      "public",
//	    PublicIP:  "203.0.113.50",
//	    PrivateIP: "10.0.1.5",
//	}
//
// Identifying Properties:
//   - name (required): Unique name of the NAT gateway
//
// Relationships:
//   - None (root node)
type NATGateway struct {
	// Name is the unique identifier for this NAT gateway.
	// This is an identifying property and is required.
	Name string

	// Type is the type of NAT.
	// Optional. Common values: "public", "private", "static", "dynamic"
	Type string

	// PublicIP is the public IP address.
	// Optional. Example: "203.0.113.50"
	PublicIP string

	// PrivateIP is the private IP address.
	// Optional. Example: "10.0.1.5"
	PrivateIP string

	// State is the current state.
	// Optional. Common values: "available", "unavailable", "deleting"
	State string

	// Provider is the cloud provider.
	// Optional. Example: "AWS", "GCP", "Azure"
	Provider string
}

func (n *NATGateway) NodeType() string { return "nat_gateway" }

func (n *NATGateway) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: n.Name,
	}
}

func (n *NATGateway) Properties() map[string]any {
	props := n.IdentifyingProperties()
	if n.Type != "" {
		props["type"] = n.Type
	}
	if n.PublicIP != "" {
		props["public_ip"] = n.PublicIP
	}
	if n.PrivateIP != "" {
		props["private_ip"] = n.PrivateIP
	}
	if n.State != "" {
		props[graphrag.PropState] = n.State
	}
	if n.Provider != "" {
		props["provider"] = n.Provider
	}
	return props
}

func (n *NATGateway) ParentRef() *NodeRef      { return nil }
func (n *NATGateway) RelationshipType() string { return "" }

// BGPPeer represents a Border Gateway Protocol peer relationship.
// BGP peers exchange routing information between autonomous systems.
//
// Example:
//
//	peer := &BGPPeer{
//	    ASN:         65000,
//	    IP:          "192.0.2.1",
//	    RemoteASN:   65001,
//	    State:       "established",
//	}
//
// Identifying Properties:
//   - asn (required): Local Autonomous System Number
//   - ip (required): Peer IP address
//
// Relationships:
//   - None (root node)
type BGPPeer struct {
	// ASN is the local Autonomous System Number.
	// This is an identifying property and is required.
	ASN int

	// IP is the peer IP address.
	// This is an identifying property and is required.
	IP string

	// RemoteASN is the peer's Autonomous System Number.
	// Optional.
	RemoteASN int

	// State is the BGP session state.
	// Optional. Common values: "established", "active", "idle", "connect"
	State string

	// LocalIP is the local IP address used for peering.
	// Optional.
	LocalIP string

	// Description is a description of this peer.
	// Optional.
	Description string
}

func (b *BGPPeer) NodeType() string { return "bgp_peer" }

func (b *BGPPeer) IdentifyingProperties() map[string]any {
	return map[string]any{
		"asn":           b.ASN,
		graphrag.PropIP: b.IP,
	}
}

func (b *BGPPeer) Properties() map[string]any {
	props := b.IdentifyingProperties()
	if b.RemoteASN > 0 {
		props["remote_asn"] = b.RemoteASN
	}
	if b.State != "" {
		props[graphrag.PropState] = b.State
	}
	if b.LocalIP != "" {
		props["local_ip"] = b.LocalIP
	}
	if b.Description != "" {
		props[graphrag.PropDescription] = b.Description
	}
	return props
}

func (b *BGPPeer) ParentRef() *NodeRef      { return nil }
func (b *BGPPeer) RelationshipType() string { return "" }

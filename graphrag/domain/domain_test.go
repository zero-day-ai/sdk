package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/graphrag"
)

// TestHost_GraphNodeInterface tests that Host implements GraphNode correctly.
func TestHost_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		host     *Host
		wantType string
		wantID   map[string]any
		wantAll  map[string]any
	}{
		{
			name: "minimal host - only IP",
			host: &Host{
				IP: "192.168.1.1",
			},
			wantType: graphrag.NodeTypeHost,
			wantID: map[string]any{
				graphrag.PropIP: "192.168.1.1",
			},
			wantAll: map[string]any{
				graphrag.PropIP: "192.168.1.1",
			},
		},
		{
			name: "full host with all fields",
			host: &Host{
				IP:       "192.168.1.10",
				Hostname: "web-server.example.com",
				State:    "up",
				OS:       "Linux Ubuntu 22.04",
			},
			wantType: graphrag.NodeTypeHost,
			wantID: map[string]any{
				graphrag.PropIP: "192.168.1.10",
			},
			wantAll: map[string]any{
				graphrag.PropIP:    "192.168.1.10",
				"hostname":         "web-server.example.com",
				graphrag.PropState: "up",
				"os":               "Linux Ubuntu 22.04",
			},
		},
		{
			name: "IPv6 host",
			host: &Host{
				IP:       "2001:db8::1",
				Hostname: "ipv6-server",
				State:    "up",
			},
			wantType: graphrag.NodeTypeHost,
			wantID: map[string]any{
				graphrag.PropIP: "2001:db8::1",
			},
			wantAll: map[string]any{
				graphrag.PropIP:    "2001:db8::1",
				"hostname":         "ipv6-server",
				graphrag.PropState: "up",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.host.NodeType())
			assert.Equal(t, tt.wantID, tt.host.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.host.Properties())
			assert.Nil(t, tt.host.ParentRef(), "Host should be a root node")
			assert.Empty(t, tt.host.RelationshipType(), "Host should have no parent relationship")
		})
	}
}

// TestPort_GraphNodeInterface tests that Port implements GraphNode correctly.
func TestPort_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name       string
		port       *Port
		wantType   string
		wantID     map[string]any
		wantAll    map[string]any
		wantParent *NodeRef
		wantRel    string
	}{
		{
			name: "minimal port - open TCP 80",
			port: &Port{
				HostID:   "192.168.1.10",
				Number:   80,
				Protocol: "tcp",
			},
			wantType: graphrag.NodeTypePort,
			wantID: map[string]any{
				graphrag.PropHostID:   "192.168.1.10",
				graphrag.PropNumber:   80,
				graphrag.PropProtocol: "tcp",
			},
			wantAll: map[string]any{
				graphrag.PropHostID:   "192.168.1.10",
				graphrag.PropNumber:   80,
				graphrag.PropProtocol: "tcp",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeHost,
				Properties: map[string]any{
					graphrag.PropIP: "192.168.1.10",
				},
			},
			wantRel: graphrag.RelTypeHasPort,
		},
		{
			name: "full port with state",
			port: &Port{
				HostID:   "192.168.1.10",
				Number:   443,
				Protocol: "tcp",
				State:    "open",
			},
			wantType: graphrag.NodeTypePort,
			wantID: map[string]any{
				graphrag.PropHostID:   "192.168.1.10",
				graphrag.PropNumber:   443,
				graphrag.PropProtocol: "tcp",
			},
			wantAll: map[string]any{
				graphrag.PropHostID:   "192.168.1.10",
				graphrag.PropNumber:   443,
				graphrag.PropProtocol: "tcp",
				graphrag.PropState:    "open",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeHost,
				Properties: map[string]any{
					graphrag.PropIP: "192.168.1.10",
				},
			},
			wantRel: graphrag.RelTypeHasPort,
		},
		{
			name: "UDP port",
			port: &Port{
				HostID:   "10.0.0.5",
				Number:   53,
				Protocol: "udp",
				State:    "open|filtered",
			},
			wantType: graphrag.NodeTypePort,
			wantID: map[string]any{
				graphrag.PropHostID:   "10.0.0.5",
				graphrag.PropNumber:   53,
				graphrag.PropProtocol: "udp",
			},
			wantAll: map[string]any{
				graphrag.PropHostID:   "10.0.0.5",
				graphrag.PropNumber:   53,
				graphrag.PropProtocol: "udp",
				graphrag.PropState:    "open|filtered",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeHost,
				Properties: map[string]any{
					graphrag.PropIP: "10.0.0.5",
				},
			},
			wantRel: graphrag.RelTypeHasPort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.port.NodeType())
			assert.Equal(t, tt.wantID, tt.port.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.port.Properties())
			assert.Equal(t, tt.wantParent, tt.port.ParentRef())
			assert.Equal(t, tt.wantRel, tt.port.RelationshipType())
		})
	}
}

// TestService_GraphNodeInterface tests that Service implements GraphNode correctly.
func TestService_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name       string
		service    *Service
		wantType   string
		wantID     map[string]any
		wantAll    map[string]any
		wantParent *NodeRef
		wantRel    string
	}{
		{
			name: "minimal service - HTTP on IPv4",
			service: &Service{
				PortID: "192.168.1.10:80:tcp",
				Name:   "http",
			},
			wantType: graphrag.NodeTypeService,
			wantID: map[string]any{
				graphrag.PropPortID: "192.168.1.10:80:tcp",
				graphrag.PropName:   "http",
			},
			wantAll: map[string]any{
				graphrag.PropPortID: "192.168.1.10:80:tcp",
				graphrag.PropName:   "http",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypePort,
				Properties: map[string]any{
					graphrag.PropHostID:   "192.168.1.10",
					graphrag.PropNumber:   80,
					graphrag.PropProtocol: "tcp",
				},
			},
			wantRel: graphrag.RelTypeRunsService,
		},
		{
			name: "full service with version and banner",
			service: &Service{
				PortID:  "192.168.1.10:443:tcp",
				Name:    "https",
				Version: "nginx/1.18.0",
				Banner:  "nginx/1.18.0 (Ubuntu)",
			},
			wantType: graphrag.NodeTypeService,
			wantID: map[string]any{
				graphrag.PropPortID: "192.168.1.10:443:tcp",
				graphrag.PropName:   "https",
			},
			wantAll: map[string]any{
				graphrag.PropPortID: "192.168.1.10:443:tcp",
				graphrag.PropName:   "https",
				"version":           "nginx/1.18.0",
				"banner":            "nginx/1.18.0 (Ubuntu)",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypePort,
				Properties: map[string]any{
					graphrag.PropHostID:   "192.168.1.10",
					graphrag.PropNumber:   443,
					graphrag.PropProtocol: "tcp",
				},
			},
			wantRel: graphrag.RelTypeRunsService,
		},
		{
			name: "service on IPv6 host",
			service: &Service{
				PortID:  "2001:db8::1:8080:tcp",
				Name:    "http-alt",
				Version: "Apache 2.4.51",
			},
			wantType: graphrag.NodeTypeService,
			wantID: map[string]any{
				graphrag.PropPortID: "2001:db8::1:8080:tcp",
				graphrag.PropName:   "http-alt",
			},
			wantAll: map[string]any{
				graphrag.PropPortID: "2001:db8::1:8080:tcp",
				graphrag.PropName:   "http-alt",
				"version":           "Apache 2.4.51",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypePort,
				Properties: map[string]any{
					graphrag.PropHostID:   "2001:db8::1",
					graphrag.PropNumber:   8080,
					graphrag.PropProtocol: "tcp",
				},
			},
			wantRel: graphrag.RelTypeRunsService,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.service.NodeType())
			assert.Equal(t, tt.wantID, tt.service.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.service.Properties())
			assert.Equal(t, tt.wantParent, tt.service.ParentRef())
			assert.Equal(t, tt.wantRel, tt.service.RelationshipType())
		})
	}
}

// TestService_ParentRef_InvalidPortID tests Service.ParentRef() with invalid PortID formats.
func TestService_ParentRef_InvalidPortID(t *testing.T) {
	tests := []struct {
		name   string
		portID string
	}{
		{name: "empty port ID", portID: ""},
		{name: "missing protocol", portID: "192.168.1.1:80"},
		{name: "missing port number", portID: "192.168.1.1"},
		{name: "invalid port number", portID: "192.168.1.1:abc:tcp"},
		{name: "single colon", portID: "192.168.1.1:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				PortID: tt.portID,
				Name:   "test",
			}
			assert.Nil(t, service.ParentRef(), "Invalid PortID should return nil parent")
		})
	}
}

// TestParsePortID tests the parsePortID function with various inputs.
func TestParsePortID(t *testing.T) {
	tests := []struct {
		name         string
		portID       string
		wantHostID   string
		wantPort     int
		wantProtocol string
		wantErr      bool
	}{
		{
			name:         "IPv4 standard port",
			portID:       "192.168.1.1:80:tcp",
			wantHostID:   "192.168.1.1",
			wantPort:     80,
			wantProtocol: "tcp",
			wantErr:      false,
		},
		{
			name:         "IPv4 high port",
			portID:       "10.0.0.5:8443:tcp",
			wantHostID:   "10.0.0.5",
			wantPort:     8443,
			wantProtocol: "tcp",
			wantErr:      false,
		},
		{
			name:         "IPv6 with colons",
			portID:       "2001:db8::1:443:tcp",
			wantHostID:   "2001:db8::1",
			wantPort:     443,
			wantProtocol: "tcp",
			wantErr:      false,
		},
		{
			name:         "IPv6 full address",
			portID:       "2001:0db8:85a3:0000:0000:8a2e:0370:7334:8080:tcp",
			wantHostID:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantPort:     8080,
			wantProtocol: "tcp",
			wantErr:      false,
		},
		{
			name:         "UDP protocol",
			portID:       "192.168.1.1:53:udp",
			wantHostID:   "192.168.1.1",
			wantPort:     53,
			wantProtocol: "udp",
			wantErr:      false,
		},
		{
			name:    "missing protocol",
			portID:  "192.168.1.1:80",
			wantErr: true,
		},
		{
			name:    "missing port",
			portID:  "192.168.1.1",
			wantErr: true,
		},
		{
			name:    "invalid port number",
			portID:  "192.168.1.1:abc:tcp",
			wantErr: true,
		},
		{
			name:    "empty string",
			portID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostID, port, protocol, err := parsePortID(tt.portID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantHostID, hostID)
				assert.Equal(t, tt.wantPort, port)
				assert.Equal(t, tt.wantProtocol, protocol)
			}
		})
	}
}

// TestEndpoint_GraphNodeInterface tests that Endpoint implements GraphNode correctly.
func TestEndpoint_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name       string
		endpoint   *Endpoint
		wantType   string
		wantID     map[string]any
		wantAll    map[string]any
		wantParent *NodeRef
		wantRel    string
	}{
		{
			name: "minimal endpoint",
			endpoint: &Endpoint{
				ServiceID: "192.168.1.10:443:tcp:https",
				URL:       "/api/users",
				Method:    "GET",
			},
			wantType: graphrag.NodeTypeEndpoint,
			wantID: map[string]any{
				"service_id": "192.168.1.10:443:tcp:https",
				"url":        "/api/users",
				"method":     "GET",
			},
			wantAll: map[string]any{
				"service_id": "192.168.1.10:443:tcp:https",
				"url":        "/api/users",
				"method":     "GET",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeService,
				Properties: map[string]any{
					"service_id": "192.168.1.10:443:tcp:https",
				},
			},
			wantRel: graphrag.RelTypeHasEndpoint,
		},
		{
			name: "full endpoint with response data",
			endpoint: &Endpoint{
				ServiceID:     "192.168.1.10:443:tcp:https",
				URL:           "/api/login",
				Method:        "POST",
				StatusCode:    200,
				Headers:       map[string]string{"Content-Type": "application/json"},
				ResponseTime:  145,
				ContentType:   "application/json",
				ContentLength: 1024,
			},
			wantType: graphrag.NodeTypeEndpoint,
			wantID: map[string]any{
				"service_id": "192.168.1.10:443:tcp:https",
				"url":        "/api/login",
				"method":     "POST",
			},
			wantAll: map[string]any{
				"service_id":     "192.168.1.10:443:tcp:https",
				"url":            "/api/login",
				"method":         "POST",
				"status_code":    200,
				"headers":        map[string]string{"Content-Type": "application/json"},
				"response_time":  int64(145),
				"content_type":   "application/json",
				"content_length": int64(1024),
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeService,
				Properties: map[string]any{
					"service_id": "192.168.1.10:443:tcp:https",
				},
			},
			wantRel: graphrag.RelTypeHasEndpoint,
		},
		{
			name: "endpoint with empty ServiceID returns nil parent",
			endpoint: &Endpoint{
				ServiceID: "",
				URL:       "/api/test",
				Method:    "GET",
			},
			wantType: graphrag.NodeTypeEndpoint,
			wantID: map[string]any{
				"service_id": "",
				"url":        "/api/test",
				"method":     "GET",
			},
			wantAll: map[string]any{
				"service_id": "",
				"url":        "/api/test",
				"method":     "GET",
			},
			wantParent: nil,
			wantRel:    graphrag.RelTypeHasEndpoint,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.endpoint.NodeType())
			assert.Equal(t, tt.wantID, tt.endpoint.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.endpoint.Properties())
			assert.Equal(t, tt.wantParent, tt.endpoint.ParentRef())
			assert.Equal(t, tt.wantRel, tt.endpoint.RelationshipType())
		})
	}
}

// TestDomain_GraphNodeInterface tests that Domain implements GraphNode correctly.
func TestDomain_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		domain   *Domain
		wantType string
		wantID   map[string]any
		wantAll  map[string]any
	}{
		{
			name: "minimal domain - only name",
			domain: &Domain{
				Name: "example.com",
			},
			wantType: graphrag.NodeTypeDomain,
			wantID: map[string]any{
				"name": "example.com",
			},
			wantAll: map[string]any{
				"name": "example.com",
			},
		},
		{
			name: "full domain with all fields",
			domain: &Domain{
				Name:        "example.com",
				Registrar:   "GoDaddy",
				CreatedAt:   "2010-01-01T00:00:00Z",
				ExpiresAt:   "2025-01-01T00:00:00Z",
				Nameservers: []string{"ns1.example.com", "ns2.example.com"},
				Status:      "active",
			},
			wantType: graphrag.NodeTypeDomain,
			wantID: map[string]any{
				"name": "example.com",
			},
			wantAll: map[string]any{
				"name":        "example.com",
				"registrar":   "GoDaddy",
				"created_at":  "2010-01-01T00:00:00Z",
				"expires_at":  "2025-01-01T00:00:00Z",
				"nameservers": []string{"ns1.example.com", "ns2.example.com"},
				"status":      "active",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.domain.NodeType())
			assert.Equal(t, tt.wantID, tt.domain.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.domain.Properties())
			assert.Nil(t, tt.domain.ParentRef(), "Domain should be a root node")
			assert.Empty(t, tt.domain.RelationshipType(), "Domain should have no parent relationship")
		})
	}
}

// TestSubdomain_GraphNodeInterface tests that Subdomain implements GraphNode correctly.
func TestSubdomain_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name       string
		subdomain  *Subdomain
		wantType   string
		wantID     map[string]any
		wantAll    map[string]any
		wantParent *NodeRef
		wantRel    string
	}{
		{
			name: "minimal subdomain",
			subdomain: &Subdomain{
				ParentDomain: "example.com",
				Name:         "api.example.com",
			},
			wantType: graphrag.NodeTypeSubdomain,
			wantID: map[string]any{
				"parent_domain": "example.com",
				"name":          "api.example.com",
			},
			wantAll: map[string]any{
				"parent_domain": "example.com",
				"name":          "api.example.com",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeDomain,
				Properties: map[string]any{
					"name": "example.com",
				},
			},
			wantRel: graphrag.RelTypeHasSubdomain,
		},
		{
			name: "full subdomain with DNS records",
			subdomain: &Subdomain{
				ParentDomain: "example.com",
				Name:         "www.example.com",
				RecordType:   "A",
				RecordValue:  "192.168.1.10",
				TTL:          3600,
				Status:       "active",
			},
			wantType: graphrag.NodeTypeSubdomain,
			wantID: map[string]any{
				"parent_domain": "example.com",
				"name":          "www.example.com",
			},
			wantAll: map[string]any{
				"parent_domain": "example.com",
				"name":          "www.example.com",
				"record_type":   "A",
				"record_value":  "192.168.1.10",
				"ttl":           3600,
				"status":        "active",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeDomain,
				Properties: map[string]any{
					"name": "example.com",
				},
			},
			wantRel: graphrag.RelTypeHasSubdomain,
		},
		{
			name: "subdomain with empty parent returns nil",
			subdomain: &Subdomain{
				ParentDomain: "",
				Name:         "orphan.com",
			},
			wantType: graphrag.NodeTypeSubdomain,
			wantID: map[string]any{
				"parent_domain": "",
				"name":          "orphan.com",
			},
			wantAll: map[string]any{
				"parent_domain": "",
				"name":          "orphan.com",
			},
			wantParent: nil,
			wantRel:    graphrag.RelTypeHasSubdomain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.subdomain.NodeType())
			assert.Equal(t, tt.wantID, tt.subdomain.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.subdomain.Properties())
			assert.Equal(t, tt.wantParent, tt.subdomain.ParentRef())
			assert.Equal(t, tt.wantRel, tt.subdomain.RelationshipType())
		})
	}
}

// TestTechnology_GraphNodeInterface tests that Technology implements GraphNode correctly.
func TestTechnology_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		tech     *Technology
		wantType string
		wantID   map[string]any
		wantAll  map[string]any
	}{
		{
			name: "minimal technology - name and version",
			tech: &Technology{
				Name:    "nginx",
				Version: "1.18.0",
			},
			wantType: graphrag.NodeTypeTechnology,
			wantID: map[string]any{
				"name":    "nginx",
				"version": "1.18.0",
			},
			wantAll: map[string]any{
				"name":    "nginx",
				"version": "1.18.0",
			},
		},
		{
			name: "full technology with all fields",
			tech: &Technology{
				Name:     "PostgreSQL",
				Version:  "14.5",
				Category: "database",
				Vendor:   "PostgreSQL Global Development Group",
				CPE:      "cpe:/a:postgresql:postgresql:14.5",
				License:  "PostgreSQL License",
				EOL:      "2026-11-12",
			},
			wantType: graphrag.NodeTypeTechnology,
			wantID: map[string]any{
				"name":    "PostgreSQL",
				"version": "14.5",
			},
			wantAll: map[string]any{
				"name":     "PostgreSQL",
				"version":  "14.5",
				"category": "database",
				"vendor":   "PostgreSQL Global Development Group",
				"cpe":      "cpe:/a:postgresql:postgresql:14.5",
				"license":  "PostgreSQL License",
				"eol":      "2026-11-12",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.tech.NodeType())
			assert.Equal(t, tt.wantID, tt.tech.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.tech.Properties())
			assert.Nil(t, tt.tech.ParentRef(), "Technology should be a root node")
			assert.Empty(t, tt.tech.RelationshipType(), "Technology should have no parent relationship")
		})
	}
}

// TestCertificate_GraphNodeInterface tests that Certificate implements GraphNode correctly.
func TestCertificate_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		cert     *Certificate
		wantType string
		wantID   map[string]any
		wantAll  map[string]any
	}{
		{
			name: "minimal certificate - only fingerprint",
			cert: &Certificate{
				Fingerprint: "SHA256:1234567890abcdef",
			},
			wantType: graphrag.NodeTypeCertificate,
			wantID: map[string]any{
				"fingerprint": "SHA256:1234567890abcdef",
			},
			wantAll: map[string]any{
				"fingerprint": "SHA256:1234567890abcdef",
				"self_signed": false,
			},
		},
		{
			name: "full certificate with all fields",
			cert: &Certificate{
				Fingerprint:        "SHA256:abcdef1234567890",
				Subject:            "CN=example.com",
				Issuer:             "CN=Let's Encrypt Authority X3",
				NotBefore:          "2024-01-01T00:00:00Z",
				NotAfter:           "2025-01-01T00:00:00Z",
				SerialNumber:       "1234567890",
				SubjectAltNames:    []string{"example.com", "www.example.com"},
				SignatureAlgorithm: "SHA256-RSA",
				KeySize:            2048,
				SelfSigned:         false,
			},
			wantType: graphrag.NodeTypeCertificate,
			wantID: map[string]any{
				"fingerprint": "SHA256:abcdef1234567890",
			},
			wantAll: map[string]any{
				"fingerprint":         "SHA256:abcdef1234567890",
				"subject":             "CN=example.com",
				"issuer":              "CN=Let's Encrypt Authority X3",
				"not_before":          "2024-01-01T00:00:00Z",
				"not_after":           "2025-01-01T00:00:00Z",
				"serial_number":       "1234567890",
				"subject_alt_names":   []string{"example.com", "www.example.com"},
				"signature_algorithm": "SHA256-RSA",
				"key_size":            2048,
				"self_signed":         false,
			},
		},
		{
			name: "self-signed certificate",
			cert: &Certificate{
				Fingerprint: "SHA256:self123",
				Subject:     "CN=localhost",
				SelfSigned:  true,
			},
			wantType: graphrag.NodeTypeCertificate,
			wantID: map[string]any{
				"fingerprint": "SHA256:self123",
			},
			wantAll: map[string]any{
				"fingerprint": "SHA256:self123",
				"subject":     "CN=localhost",
				"self_signed": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.cert.NodeType())
			assert.Equal(t, tt.wantID, tt.cert.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.cert.Properties())
			assert.Nil(t, tt.cert.ParentRef(), "Certificate should be a root node")
			assert.Empty(t, tt.cert.RelationshipType(), "Certificate should have no parent relationship")
		})
	}
}

// TestCloudAsset_GraphNodeInterface tests that CloudAsset implements GraphNode correctly.
func TestCloudAsset_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		asset    *CloudAsset
		wantType string
		wantID   map[string]any
		wantAll  map[string]any
	}{
		{
			name: "minimal cloud asset - AWS EC2",
			asset: &CloudAsset{
				Provider:   "aws",
				ResourceID: "i-0123456789abcdef0",
			},
			wantType: graphrag.NodeTypeCloudAsset,
			wantID: map[string]any{
				"provider":    "aws",
				"resource_id": "i-0123456789abcdef0",
			},
			wantAll: map[string]any{
				"provider":    "aws",
				"resource_id": "i-0123456789abcdef0",
			},
		},
		{
			name: "full cloud asset with all fields",
			asset: &CloudAsset{
				Provider:       "aws",
				ResourceID:     "i-0123456789abcdef0",
				Region:         "us-east-1",
				Type:           "ec2-instance",
				Name:           "web-server-01",
				AccountID:      "123456789012",
				VPC:            "vpc-abc123",
				SubnetID:       "subnet-xyz789",
				SecurityGroups: []string{"sg-web", "sg-ssh"},
				Tags:           map[string]string{"Environment": "production", "Team": "security"},
				State:          "running",
			},
			wantType: graphrag.NodeTypeCloudAsset,
			wantID: map[string]any{
				"provider":    "aws",
				"resource_id": "i-0123456789abcdef0",
			},
			wantAll: map[string]any{
				"provider":        "aws",
				"resource_id":     "i-0123456789abcdef0",
				"region":          "us-east-1",
				"type":            "ec2-instance",
				"name":            "web-server-01",
				"account_id":      "123456789012",
				"vpc":             "vpc-abc123",
				"subnet_id":       "subnet-xyz789",
				"security_groups": []string{"sg-web", "sg-ssh"},
				"tags":            map[string]string{"Environment": "production", "Team": "security"},
				"state":           "running",
			},
		},
		{
			name: "GCP compute instance",
			asset: &CloudAsset{
				Provider:   "gcp",
				ResourceID: "projects/my-project/zones/us-central1-a/instances/vm-1",
				Region:     "us-central1",
				Type:       "compute-instance",
			},
			wantType: graphrag.NodeTypeCloudAsset,
			wantID: map[string]any{
				"provider":    "gcp",
				"resource_id": "projects/my-project/zones/us-central1-a/instances/vm-1",
			},
			wantAll: map[string]any{
				"provider":    "gcp",
				"resource_id": "projects/my-project/zones/us-central1-a/instances/vm-1",
				"region":      "us-central1",
				"type":        "compute-instance",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.asset.NodeType())
			assert.Equal(t, tt.wantID, tt.asset.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.asset.Properties())
			assert.Nil(t, tt.asset.ParentRef(), "CloudAsset should be a root node")
			assert.Empty(t, tt.asset.RelationshipType(), "CloudAsset should have no parent relationship")
		})
	}
}

// TestAPI_GraphNodeInterface tests that API implements GraphNode correctly.
func TestAPI_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		api      *API
		wantType string
		wantID   map[string]any
		wantAll  map[string]any
	}{
		{
			name: "minimal API - only base URL",
			api: &API{
				BaseURL: "https://api.example.com",
			},
			wantType: graphrag.NodeTypeApi,
			wantID: map[string]any{
				"base_url": "https://api.example.com",
			},
			wantAll: map[string]any{
				"base_url": "https://api.example.com",
			},
		},
		{
			name: "full API with all fields",
			api: &API{
				BaseURL:     "https://api.example.com",
				Name:        "Example API",
				Version:     "v1",
				Description: "RESTful API for example service",
				SwaggerURL:  "https://api.example.com/swagger.json",
				AuthType:    "bearer",
				RateLimit:   "1000 requests/hour",
				Status:      "active",
			},
			wantType: graphrag.NodeTypeApi,
			wantID: map[string]any{
				"base_url": "https://api.example.com",
			},
			wantAll: map[string]any{
				"base_url":    "https://api.example.com",
				"name":        "Example API",
				"version":     "v1",
				"description": "RESTful API for example service",
				"swagger_url": "https://api.example.com/swagger.json",
				"auth_type":   "bearer",
				"rate_limit":  "1000 requests/hour",
				"status":      "active",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.api.NodeType())
			assert.Equal(t, tt.wantID, tt.api.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.api.Properties())
			assert.Nil(t, tt.api.ParentRef(), "API should be a root node")
			assert.Empty(t, tt.api.RelationshipType(), "API should have no parent relationship")
		})
	}
}

// TestAgentRun_GraphNodeInterface tests that AgentRun implements GraphNode correctly.
func TestAgentRun_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		run      *AgentRun
		wantType string
		wantID   map[string]any
		wantAll  map[string]any
	}{
		{
			name: "minimal agent run",
			run: &AgentRun{
				ID:        "run-123",
				MissionID: "mission-456",
				AgentName: "network-recon",
				RunNumber: 1,
			},
			wantType: graphrag.NodeTypeAgentRun,
			wantID: map[string]any{
				"id":         "run-123",
				"mission_id": "mission-456",
				"agent_name": "network-recon",
				"run_number": 1,
			},
			wantAll: map[string]any{
				"id":         "run-123",
				"mission_id": "mission-456",
				"agent_name": "network-recon",
				"run_number": 1,
			},
		},
		{
			name: "full agent run with all fields",
			run: &AgentRun{
				ID:        "run-123",
				MissionID: "mission-456",
				AgentName: "network-recon",
				RunNumber: 1,
				StartTime: "2024-01-20T10:00:00Z",
				EndTime:   "2024-01-20T10:05:30Z",
				Status:    "completed",
				Error:     "",
				Duration:  330.5,
			},
			wantType: graphrag.NodeTypeAgentRun,
			wantID: map[string]any{
				"id":         "run-123",
				"mission_id": "mission-456",
				"agent_name": "network-recon",
				"run_number": 1,
			},
			wantAll: map[string]any{
				"id":         "run-123",
				"mission_id": "mission-456",
				"agent_name": "network-recon",
				"run_number": 1,
				"start_time": "2024-01-20T10:00:00Z",
				"end_time":   "2024-01-20T10:05:30Z",
				"status":     "completed",
				"duration":   330.5,
			},
		},
		{
			name: "failed agent run with error",
			run: &AgentRun{
				ID:        "run-789",
				MissionID: "mission-456",
				AgentName: "sql-injection",
				RunNumber: 2,
				Status:    "failed",
				Error:     "connection timeout",
			},
			wantType: graphrag.NodeTypeAgentRun,
			wantID: map[string]any{
				"id":         "run-789",
				"mission_id": "mission-456",
				"agent_name": "sql-injection",
				"run_number": 2,
			},
			wantAll: map[string]any{
				"id":         "run-789",
				"mission_id": "mission-456",
				"agent_name": "sql-injection",
				"run_number": 2,
				"status":     "failed",
				"error":      "connection timeout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.run.NodeType())
			assert.Equal(t, tt.wantID, tt.run.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.run.Properties())
			assert.Nil(t, tt.run.ParentRef(), "AgentRun should be a root node")
			assert.Empty(t, tt.run.RelationshipType(), "AgentRun should have no parent relationship")
		})
	}
}

// TestToolExecution_GraphNodeInterface tests that ToolExecution implements GraphNode correctly.
func TestToolExecution_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name       string
		exec       *ToolExecution
		wantType   string
		wantID     map[string]any
		wantAll    map[string]any
		wantParent *NodeRef
		wantRel    string
	}{
		{
			name: "minimal tool execution",
			exec: &ToolExecution{
				ID:         "exec-789",
				AgentRunID: "run-123",
				ToolName:   "nmap",
				Sequence:   1,
			},
			wantType: graphrag.NodeTypeToolExecution,
			wantID: map[string]any{
				"id":           "exec-789",
				"agent_run_id": "run-123",
				"tool_name":    "nmap",
				"sequence":     1,
			},
			wantAll: map[string]any{
				"id":           "exec-789",
				"agent_run_id": "run-123",
				"tool_name":    "nmap",
				"sequence":     1,
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeAgentRun,
				Properties: map[string]any{
					"id": "run-123",
				},
			},
			wantRel: graphrag.RelTypePartOf,
		},
		{
			name: "full tool execution with all fields",
			exec: &ToolExecution{
				ID:         "exec-789",
				AgentRunID: "run-123",
				ToolName:   "nmap",
				Sequence:   1,
				StartTime:  "2024-01-20T10:05:00Z",
				EndTime:    "2024-01-20T10:10:12Z",
				Duration:   312.5,
				Status:     "success",
				Error:      "",
				ExitCode:   0,
				Command:    "nmap -sS -p- 192.168.1.0/24",
			},
			wantType: graphrag.NodeTypeToolExecution,
			wantID: map[string]any{
				"id":           "exec-789",
				"agent_run_id": "run-123",
				"tool_name":    "nmap",
				"sequence":     1,
			},
			wantAll: map[string]any{
				"id":           "exec-789",
				"agent_run_id": "run-123",
				"tool_name":    "nmap",
				"sequence":     1,
				"start_time":   "2024-01-20T10:05:00Z",
				"end_time":     "2024-01-20T10:10:12Z",
				"duration":     312.5,
				"status":       "success",
				"command":      "nmap -sS -p- 192.168.1.0/24",
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeAgentRun,
				Properties: map[string]any{
					"id": "run-123",
				},
			},
			wantRel: graphrag.RelTypePartOf,
		},
		{
			name: "failed tool execution with error",
			exec: &ToolExecution{
				ID:         "exec-999",
				AgentRunID: "run-456",
				ToolName:   "sqlmap",
				Sequence:   3,
				Status:     "failed",
				Error:      "target unreachable",
				ExitCode:   1,
			},
			wantType: graphrag.NodeTypeToolExecution,
			wantID: map[string]any{
				"id":           "exec-999",
				"agent_run_id": "run-456",
				"tool_name":    "sqlmap",
				"sequence":     3,
			},
			wantAll: map[string]any{
				"id":           "exec-999",
				"agent_run_id": "run-456",
				"tool_name":    "sqlmap",
				"sequence":     3,
				"status":       "failed",
				"error":        "target unreachable",
				"exit_code":    1,
			},
			wantParent: &NodeRef{
				NodeType: graphrag.NodeTypeAgentRun,
				Properties: map[string]any{
					"id": "run-456",
				},
			},
			wantRel: graphrag.RelTypePartOf,
		},
		{
			name: "tool execution with empty AgentRunID returns nil parent",
			exec: &ToolExecution{
				ID:         "exec-orphan",
				AgentRunID: "",
				ToolName:   "test",
				Sequence:   1,
			},
			wantType: graphrag.NodeTypeToolExecution,
			wantID: map[string]any{
				"id":           "exec-orphan",
				"agent_run_id": "",
				"tool_name":    "test",
				"sequence":     1,
			},
			wantAll: map[string]any{
				"id":           "exec-orphan",
				"agent_run_id": "",
				"tool_name":    "test",
				"sequence":     1,
			},
			wantParent: nil,
			wantRel:    graphrag.RelTypePartOf,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.exec.NodeType())
			assert.Equal(t, tt.wantID, tt.exec.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.exec.Properties())
			assert.Equal(t, tt.wantParent, tt.exec.ParentRef())
			assert.Equal(t, tt.wantRel, tt.exec.RelationshipType())
		})
	}
}

// TestCustomEntity_GraphNodeInterface tests that CustomEntity implements GraphNode correctly.
func TestCustomEntity_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name       string
		entity     *CustomEntity
		wantType   string
		wantID     map[string]any
		wantAll    map[string]any
		wantParent *NodeRef
		wantRel    string
	}{
		{
			name: "Kubernetes pod - minimal",
			entity: &CustomEntity{
				Namespace: "k8s",
				Type:      "pod",
				IDProps: map[string]any{
					"namespace": "default",
					"name":      "web-server-abc123",
				},
			},
			wantType: "k8s:pod",
			wantID: map[string]any{
				"namespace": "default",
				"name":      "web-server-abc123",
			},
			wantAll: map[string]any{
				"namespace": "default",
				"name":      "web-server-abc123",
			},
			wantParent: nil,
			wantRel:    "",
		},
		{
			name: "Kubernetes pod - with all properties",
			entity: &CustomEntity{
				Namespace: "k8s",
				Type:      "pod",
				IDProps: map[string]any{
					"namespace": "default",
					"name":      "web-server-abc123",
				},
				AllProps: map[string]any{
					"namespace": "default",
					"name":      "web-server-abc123",
					"status":    "Running",
					"image":     "nginx:1.21",
					"node":      "node-01",
				},
			},
			wantType: "k8s:pod",
			wantID: map[string]any{
				"namespace": "default",
				"name":      "web-server-abc123",
			},
			wantAll: map[string]any{
				"namespace": "default",
				"name":      "web-server-abc123",
				"status":    "Running",
				"image":     "nginx:1.21",
				"node":      "node-01",
			},
			wantParent: nil,
			wantRel:    "",
		},
		{
			name: "AWS security group - with parent",
			entity: &CustomEntity{
				Namespace: "aws",
				Type:      "security_group",
				IDProps: map[string]any{
					"id": "sg-0123456789abcdef0",
				},
				AllProps: map[string]any{
					"id":          "sg-0123456789abcdef0",
					"name":        "web-server-sg",
					"description": "Security group for web servers",
					"vpc_id":      "vpc-abc123",
				},
				Parent: &NodeRef{
					NodeType: "aws:vpc",
					Properties: map[string]any{
						"id": "vpc-abc123",
					},
				},
				ParentRel: "BELONGS_TO",
			},
			wantType: "aws:security_group",
			wantID: map[string]any{
				"id": "sg-0123456789abcdef0",
			},
			wantAll: map[string]any{
				"id":          "sg-0123456789abcdef0",
				"name":        "web-server-sg",
				"description": "Security group for web servers",
				"vpc_id":      "vpc-abc123",
			},
			wantParent: &NodeRef{
				NodeType: "aws:vpc",
				Properties: map[string]any{
					"id": "vpc-abc123",
				},
			},
			wantRel: "BELONGS_TO",
		},
		{
			name: "custom entity with empty IDProps",
			entity: &CustomEntity{
				Namespace: "custom",
				Type:      "test",
				IDProps:   nil,
			},
			wantType:   "custom:test",
			wantID:     map[string]any{},
			wantAll:    map[string]any{},
			wantParent: nil,
			wantRel:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.entity.NodeType())
			assert.Equal(t, tt.wantID, tt.entity.IdentifyingProperties())
			assert.Equal(t, tt.wantAll, tt.entity.Properties())
			assert.Equal(t, tt.wantParent, tt.entity.ParentRef())
			assert.Equal(t, tt.wantRel, tt.entity.RelationshipType())
		})
	}
}

// TestCustomEntity_FluentAPI tests the fluent builder API for CustomEntity.
func TestCustomEntity_FluentAPI(t *testing.T) {
	// Test basic construction
	entity := NewCustomEntity("k8s", "pod").
		WithIDProps(map[string]any{
			"namespace": "default",
			"name":      "web-01",
		}).
		WithAllProps(map[string]any{
			"namespace": "default",
			"name":      "web-01",
			"status":    "Running",
		})

	assert.Equal(t, "k8s:pod", entity.NodeType())
	assert.Equal(t, map[string]any{"namespace": "default", "name": "web-01"}, entity.IdentifyingProperties())
	assert.Equal(t, map[string]any{"namespace": "default", "name": "web-01", "status": "Running"}, entity.Properties())
	assert.Nil(t, entity.ParentRef())

	// Test with parent
	entityWithParent := NewCustomEntity("aws", "subnet").
		WithIDProps(map[string]any{"id": "subnet-123"}).
		WithParent(&NodeRef{
			NodeType:   "aws:vpc",
			Properties: map[string]any{"id": "vpc-456"},
		}, "PART_OF")

	assert.NotNil(t, entityWithParent.ParentRef())
	assert.Equal(t, "PART_OF", entityWithParent.RelationshipType())
}

// TestDiscoveryResult_AllNodes tests DiscoveryResult.AllNodes() returns nodes in correct order.
func TestDiscoveryResult_AllNodes(t *testing.T) {
	result := &DiscoveryResult{
		Hosts: []*Host{
			{IP: "192.168.1.1"},
		},
		Ports: []*Port{
			{HostID: "192.168.1.1", Number: 80, Protocol: "tcp"},
		},
		Services: []*Service{
			{PortID: "192.168.1.1:80:tcp", Name: "http"},
		},
		Endpoints: []*Endpoint{
			{ServiceID: "service-1", URL: "/api", Method: "GET"},
		},
		Domains: []*Domain{
			{Name: "example.com"},
		},
		Subdomains: []*Subdomain{
			{ParentDomain: "example.com", Name: "api.example.com"},
		},
		Technologies: []*Technology{
			{Name: "nginx", Version: "1.18.0"},
		},
		Certificates: []*Certificate{
			{Fingerprint: "SHA256:abc123"},
		},
		CloudAssets: []*CloudAsset{
			{Provider: "aws", ResourceID: "i-123"},
		},
		APIs: []*API{
			{BaseURL: "https://api.example.com"},
		},
		Custom: []GraphNode{
			NewCustomEntity("k8s", "pod").WithIDProps(map[string]any{"name": "pod-1"}),
		},
	}

	nodes := result.AllNodes()
	assert.Len(t, nodes, 11)

	// Verify order: new dependency-based ordering
	// 1. Compute: Hosts
	assert.Equal(t, graphrag.NodeTypeHost, nodes[0].NodeType())
	// 2. Network details: Ports
	assert.Equal(t, graphrag.NodeTypePort, nodes[1].NodeType())
	// 3. Services and domains
	assert.Equal(t, graphrag.NodeTypeService, nodes[2].NodeType())
	assert.Equal(t, graphrag.NodeTypeDomain, nodes[3].NodeType())
	assert.Equal(t, graphrag.NodeTypeSubdomain, nodes[4].NodeType())
	assert.Equal(t, graphrag.NodeTypeTechnology, nodes[5].NodeType())
	assert.Equal(t, graphrag.NodeTypeCertificate, nodes[6].NodeType())
	assert.Equal(t, graphrag.NodeTypeCloudAsset, nodes[7].NodeType())
	// 4. Web/API resources
	assert.Equal(t, graphrag.NodeTypeApi, nodes[8].NodeType())
	assert.Equal(t, graphrag.NodeTypeEndpoint, nodes[9].NodeType())
	// 5. Custom nodes
	assert.Equal(t, "k8s:pod", nodes[10].NodeType())
}

// TestDiscoveryResult_IsEmpty tests DiscoveryResult.IsEmpty().
func TestDiscoveryResult_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		result *DiscoveryResult
		want   bool
	}{
		{
			name:   "completely empty",
			result: &DiscoveryResult{},
			want:   true,
		},
		{
			name: "has hosts",
			result: &DiscoveryResult{
				Hosts: []*Host{{IP: "192.168.1.1"}},
			},
			want: false,
		},
		{
			name: "has custom nodes",
			result: &DiscoveryResult{
				Custom: []GraphNode{
					NewCustomEntity("k8s", "pod").WithIDProps(map[string]any{"name": "test"}),
				},
			},
			want: false,
		},
		{
			name: "has ports only",
			result: &DiscoveryResult{
				Ports: []*Port{{HostID: "192.168.1.1", Number: 80, Protocol: "tcp"}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.IsEmpty())
		})
	}
}

// TestDiscoveryResult_NodeCount tests DiscoveryResult.NodeCount().
func TestDiscoveryResult_NodeCount(t *testing.T) {
	tests := []struct {
		name   string
		result *DiscoveryResult
		want   int
	}{
		{
			name:   "empty result",
			result: &DiscoveryResult{},
			want:   0,
		},
		{
			name: "single host",
			result: &DiscoveryResult{
				Hosts: []*Host{{IP: "192.168.1.1"}},
			},
			want: 1,
		},
		{
			name: "multiple types",
			result: &DiscoveryResult{
				Hosts:   []*Host{{IP: "192.168.1.1"}, {IP: "192.168.1.2"}},
				Ports:   []*Port{{HostID: "192.168.1.1", Number: 80, Protocol: "tcp"}},
				Domains: []*Domain{{Name: "example.com"}},
				Custom: []GraphNode{
					NewCustomEntity("k8s", "pod").WithIDProps(map[string]any{"name": "pod-1"}),
					NewCustomEntity("k8s", "service").WithIDProps(map[string]any{"name": "svc-1"}),
				},
			},
			want: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.result.NodeCount())
		})
	}
}

// TestNewDiscoveryResult tests NewDiscoveryResult() initializes all slices.
func TestNewDiscoveryResult(t *testing.T) {
	result := NewDiscoveryResult()

	assert.NotNil(t, result.Hosts)
	assert.NotNil(t, result.Ports)
	assert.NotNil(t, result.Services)
	assert.NotNil(t, result.Endpoints)
	assert.NotNil(t, result.Domains)
	assert.NotNil(t, result.Subdomains)
	assert.NotNil(t, result.Technologies)
	assert.NotNil(t, result.Certificates)
	assert.NotNil(t, result.CloudAssets)
	assert.NotNil(t, result.APIs)
	assert.NotNil(t, result.Custom)

	assert.True(t, result.IsEmpty())
	assert.Equal(t, 0, result.NodeCount())
}

// TestDiscoveryResult_AllNodes_PreservesOrder tests that AllNodes preserves insertion order for custom nodes.
func TestDiscoveryResult_AllNodes_PreservesOrder(t *testing.T) {
	result := &DiscoveryResult{
		Custom: []GraphNode{
			NewCustomEntity("k8s", "pod").WithIDProps(map[string]any{"name": "pod-1"}),
			NewCustomEntity("k8s", "service").WithIDProps(map[string]any{"name": "svc-1"}),
			NewCustomEntity("k8s", "deployment").WithIDProps(map[string]any{"name": "deploy-1"}),
		},
	}

	nodes := result.AllNodes()
	require.Len(t, nodes, 3)

	// Verify custom nodes are in insertion order
	assert.Equal(t, "k8s:pod", nodes[0].NodeType())
	assert.Equal(t, "k8s:service", nodes[1].NodeType())
	assert.Equal(t, "k8s:deployment", nodes[2].NodeType())
}

// TestDiscoveryResult_AllNodes_EmptyResult tests AllNodes with completely empty result.
func TestDiscoveryResult_AllNodes_EmptyResult(t *testing.T) {
	result := &DiscoveryResult{}
	nodes := result.AllNodes()
	// In Go, an uninitialized slice is nil, which is valid and has length 0
	assert.Len(t, nodes, 0, "AllNodes should return empty/nil slice for empty result")
	assert.True(t, result.IsEmpty())
	assert.Equal(t, 0, result.NodeCount())
}

// TestCustomEntity_ImmutableProperties tests that IdentifyingProperties and Properties return copies.
func TestCustomEntity_ImmutableProperties(t *testing.T) {
	entity := NewCustomEntity("test", "entity").
		WithIDProps(map[string]any{"id": "123"}).
		WithAllProps(map[string]any{"id": "123", "name": "test"})

	// Get properties
	idProps := entity.IdentifyingProperties()
	allProps := entity.Properties()

	// Modify returned maps
	idProps["id"] = "modified"
	allProps["name"] = "modified"

	// Verify original entity is unchanged
	assert.Equal(t, "123", entity.IDProps["id"])
	assert.Equal(t, "test", entity.AllProps["name"])

	// Get properties again - should be unchanged
	idProps2 := entity.IdentifyingProperties()
	allProps2 := entity.Properties()

	assert.Equal(t, "123", idProps2["id"])
	assert.Equal(t, "test", allProps2["name"])
}

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/graphrag"
)

// TestPortBelongsTo tests the Port BelongsTo pattern
func TestPortBelongsTo(t *testing.T) {
	t.Run("NewPort creates port with required properties", func(t *testing.T) {
		port := NewPort(443, "tcp")
		assert.Equal(t, 443, port.Number)
		assert.Equal(t, "tcp", port.Protocol)
		assert.Empty(t, port.HostID) // Not set until BelongsTo is called
	})

	t.Run("BelongsTo sets parent and HostID", func(t *testing.T) {
		host := &Host{IP: "192.168.1.1"}
		port := NewPort(443, "tcp").BelongsTo(host)

		// Check HostID is set for backward compatibility
		assert.Equal(t, "192.168.1.1", port.HostID)

		// Check ParentRef uses internal parent
		parentRef := port.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeHost, parentRef.NodeType)
		assert.Equal(t, "192.168.1.1", parentRef.Properties[graphrag.PropIP])
	})

	t.Run("BelongsTo returns port for method chaining", func(t *testing.T) {
		host := &Host{IP: "192.168.1.1"}
		port := NewPort(443, "tcp").BelongsTo(host)
		port.State = "open" // Should be chainable

		assert.Equal(t, "open", port.State)
		assert.Equal(t, "192.168.1.1", port.HostID)
	})

	t.Run("BelongsTo panics on nil host", func(t *testing.T) {
		port := NewPort(443, "tcp")
		assert.Panics(t, func() {
			port.BelongsTo(nil)
		})
	})

	t.Run("BelongsTo panics on empty host IP", func(t *testing.T) {
		host := &Host{IP: ""}
		port := NewPort(443, "tcp")
		assert.Panics(t, func() {
			port.BelongsTo(host)
		})
	})

	t.Run("ParentRef falls back to HostID when parent not set", func(t *testing.T) {
		// Legacy pattern - setting HostID directly
		port := &Port{
			HostID:   "10.0.0.1",
			Number:   80,
			Protocol: "tcp",
		}

		parentRef := port.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeHost, parentRef.NodeType)
		assert.Equal(t, "10.0.0.1", parentRef.Properties[graphrag.PropIP])
	})

	t.Run("ParentRef returns nil when neither parent nor HostID set", func(t *testing.T) {
		port := NewPort(443, "tcp")
		assert.Nil(t, port.ParentRef())
	})

	t.Run("BelongsTo takes precedence over HostID", func(t *testing.T) {
		port := &Port{
			HostID:   "10.0.0.1", // Legacy value
			Number:   80,
			Protocol: "tcp",
		}

		host := &Host{IP: "192.168.1.1"}
		port.BelongsTo(host)

		// HostID should be updated
		assert.Equal(t, "192.168.1.1", port.HostID)

		// ParentRef should use new parent
		parentRef := port.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, "192.168.1.1", parentRef.Properties[graphrag.PropIP])
	})
}

// TestServiceBelongsTo tests the Service BelongsTo pattern
func TestServiceBelongsTo(t *testing.T) {
	t.Run("NewService creates service with required properties", func(t *testing.T) {
		service := NewService("http")
		assert.Equal(t, "http", service.Name)
		assert.Empty(t, service.PortID)
	})

	t.Run("BelongsTo sets parent and PortID", func(t *testing.T) {
		port := &Port{
			HostID:   "192.168.1.1",
			Number:   80,
			Protocol: "tcp",
		}
		service := NewService("http").BelongsTo(port)

		// Check PortID is set for backward compatibility
		assert.Equal(t, "192.168.1.1:80:tcp", service.PortID)

		// Check ParentRef uses internal parent
		parentRef := service.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypePort, parentRef.NodeType)
		assert.Equal(t, "192.168.1.1", parentRef.Properties[graphrag.PropHostID])
		assert.Equal(t, 80, parentRef.Properties[graphrag.PropNumber])
		assert.Equal(t, "tcp", parentRef.Properties[graphrag.PropProtocol])
	})

	t.Run("BelongsTo works with builder chain", func(t *testing.T) {
		host := &Host{IP: "192.168.1.1"}
		port := NewPort(443, "tcp").BelongsTo(host)
		service := NewService("https").BelongsTo(port)

		assert.Equal(t, "https", service.Name)
		assert.Equal(t, "192.168.1.1:443:tcp", service.PortID)

		parentRef := service.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypePort, parentRef.NodeType)
	})

	t.Run("BelongsTo panics on nil port", func(t *testing.T) {
		service := NewService("http")
		assert.Panics(t, func() {
			service.BelongsTo(nil)
		})
	})

	t.Run("BelongsTo panics on invalid port", func(t *testing.T) {
		service := NewService("http")

		// Missing Number
		assert.Panics(t, func() {
			service.BelongsTo(&Port{HostID: "192.168.1.1", Protocol: "tcp"})
		})

		// Missing Protocol
		assert.Panics(t, func() {
			service.BelongsTo(&Port{HostID: "192.168.1.1", Number: 80})
		})

		// Missing HostID
		assert.Panics(t, func() {
			service.BelongsTo(&Port{Number: 80, Protocol: "tcp"})
		})
	})

	t.Run("ParentRef falls back to PortID parsing", func(t *testing.T) {
		// Legacy pattern - setting PortID directly
		service := &Service{
			PortID: "10.0.0.1:443:tcp",
			Name:   "https",
		}

		parentRef := service.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypePort, parentRef.NodeType)
		assert.Equal(t, "10.0.0.1", parentRef.Properties[graphrag.PropHostID])
		assert.Equal(t, 443, parentRef.Properties[graphrag.PropNumber])
		assert.Equal(t, "tcp", parentRef.Properties[graphrag.PropProtocol])
	})

	t.Run("ParentRef returns nil on invalid PortID", func(t *testing.T) {
		service := &Service{
			PortID: "invalid",
			Name:   "http",
		}
		assert.Nil(t, service.ParentRef())
	})
}

// TestEndpointBelongsTo tests the Endpoint BelongsTo pattern
func TestEndpointBelongsTo(t *testing.T) {
	t.Run("NewEndpoint creates endpoint with required properties", func(t *testing.T) {
		endpoint := NewEndpoint("/api/users", "GET")
		assert.Equal(t, "/api/users", endpoint.URL)
		assert.Equal(t, "GET", endpoint.Method)
		assert.Empty(t, endpoint.ServiceID)
	})

	t.Run("BelongsTo sets parent and ServiceID", func(t *testing.T) {
		service := &Service{
			PortID: "192.168.1.1:443:tcp",
			Name:   "https",
		}
		endpoint := NewEndpoint("/api/users", "GET").BelongsTo(service)

		// Check ServiceID is set for backward compatibility
		assert.Equal(t, "192.168.1.1:443:tcp:https", endpoint.ServiceID)

		// Check ParentRef uses internal parent
		parentRef := endpoint.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeService, parentRef.NodeType)
		assert.Equal(t, "192.168.1.1:443:tcp", parentRef.Properties[graphrag.PropPortID])
		assert.Equal(t, "https", parentRef.Properties[graphrag.PropName])
	})

	t.Run("BelongsTo works with full builder chain", func(t *testing.T) {
		host := &Host{IP: "192.168.1.1"}
		port := NewPort(443, "tcp").BelongsTo(host)
		service := NewService("https").BelongsTo(port)
		endpoint := NewEndpoint("/api/users", "GET").BelongsTo(service)

		assert.Equal(t, "/api/users", endpoint.URL)
		assert.Equal(t, "GET", endpoint.Method)
		assert.Equal(t, "192.168.1.1:443:tcp:https", endpoint.ServiceID)

		parentRef := endpoint.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeService, parentRef.NodeType)
	})

	t.Run("BelongsTo panics on nil service", func(t *testing.T) {
		endpoint := NewEndpoint("/api", "GET")
		assert.Panics(t, func() {
			endpoint.BelongsTo(nil)
		})
	})

	t.Run("BelongsTo panics on invalid service", func(t *testing.T) {
		endpoint := NewEndpoint("/api", "GET")

		// Missing Name
		assert.Panics(t, func() {
			endpoint.BelongsTo(&Service{PortID: "192.168.1.1:443:tcp"})
		})

		// Missing PortID
		assert.Panics(t, func() {
			endpoint.BelongsTo(&Service{Name: "https"})
		})
	})

	t.Run("ParentRef falls back to ServiceID", func(t *testing.T) {
		// Legacy pattern - setting ServiceID directly
		endpoint := &Endpoint{
			ServiceID: "10.0.0.1:443:tcp:https",
			URL:       "/api",
			Method:    "GET",
		}

		parentRef := endpoint.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeService, parentRef.NodeType)
		assert.Equal(t, "10.0.0.1:443:tcp:https", parentRef.Properties["service_id"])
	})
}

// TestSubdomainBelongsTo tests the Subdomain BelongsTo pattern
func TestSubdomainBelongsTo(t *testing.T) {
	t.Run("NewSubdomain creates subdomain with required properties", func(t *testing.T) {
		subdomain := NewSubdomain("api.example.com")
		assert.Equal(t, "api.example.com", subdomain.Name)
		assert.Empty(t, subdomain.ParentDomain)
	})

	t.Run("BelongsTo sets parent and ParentDomain", func(t *testing.T) {
		domain := &Domain{Name: "example.com"}
		subdomain := NewSubdomain("api.example.com").BelongsTo(domain)

		// Check ParentDomain is set for backward compatibility
		assert.Equal(t, "example.com", subdomain.ParentDomain)

		// Check ParentRef uses internal parent
		parentRef := subdomain.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeDomain, parentRef.NodeType)
		assert.Equal(t, "example.com", parentRef.Properties["name"])
	})

	t.Run("BelongsTo returns subdomain for method chaining", func(t *testing.T) {
		domain := &Domain{Name: "example.com"}
		subdomain := NewSubdomain("api.example.com").BelongsTo(domain)
		subdomain.RecordType = "A"
		subdomain.RecordValue = "192.168.1.1"

		assert.Equal(t, "A", subdomain.RecordType)
		assert.Equal(t, "192.168.1.1", subdomain.RecordValue)
	})

	t.Run("BelongsTo panics on nil domain", func(t *testing.T) {
		subdomain := NewSubdomain("api.example.com")
		assert.Panics(t, func() {
			subdomain.BelongsTo(nil)
		})
	})

	t.Run("BelongsTo panics on empty domain name", func(t *testing.T) {
		domain := &Domain{Name: ""}
		subdomain := NewSubdomain("api.example.com")
		assert.Panics(t, func() {
			subdomain.BelongsTo(domain)
		})
	})

	t.Run("ParentRef falls back to ParentDomain", func(t *testing.T) {
		// Legacy pattern - setting ParentDomain directly
		subdomain := &Subdomain{
			ParentDomain: "example.com",
			Name:         "api.example.com",
		}

		parentRef := subdomain.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeDomain, parentRef.NodeType)
		assert.Equal(t, "example.com", parentRef.Properties["name"])
	})

	t.Run("ParentRef returns nil when neither parent nor ParentDomain set", func(t *testing.T) {
		subdomain := NewSubdomain("api.example.com")
		assert.Nil(t, subdomain.ParentRef())
	})
}

// TestFullHierarchyWithBelongsTo tests building a complete hierarchy using BelongsTo
func TestFullHierarchyWithBelongsTo(t *testing.T) {
	t.Run("Host -> Port -> Service -> Endpoint chain", func(t *testing.T) {
		// Build the hierarchy
		host := &Host{IP: "192.168.1.100", Hostname: "web-server"}
		port := NewPort(443, "tcp").BelongsTo(host)
		port.State = "open"

		service := NewService("https").BelongsTo(port)
		service.Version = "nginx 1.18.0"

		endpoint := NewEndpoint("/api/v1/users", "GET").BelongsTo(service)
		endpoint.StatusCode = 200

		// Verify host (root node)
		assert.Equal(t, "192.168.1.100", host.IP)
		assert.Nil(t, host.ParentRef())

		// Verify port
		assert.Equal(t, 443, port.Number)
		assert.Equal(t, "192.168.1.100", port.HostID)
		portParent := port.ParentRef()
		require.NotNil(t, portParent)
		assert.Equal(t, graphrag.NodeTypeHost, portParent.NodeType)

		// Verify service
		assert.Equal(t, "https", service.Name)
		assert.Equal(t, "192.168.1.100:443:tcp", service.PortID)
		serviceParent := service.ParentRef()
		require.NotNil(t, serviceParent)
		assert.Equal(t, graphrag.NodeTypePort, serviceParent.NodeType)

		// Verify endpoint
		assert.Equal(t, "/api/v1/users", endpoint.URL)
		assert.Equal(t, "192.168.1.100:443:tcp:https", endpoint.ServiceID)
		endpointParent := endpoint.ParentRef()
		require.NotNil(t, endpointParent)
		assert.Equal(t, graphrag.NodeTypeService, endpointParent.NodeType)
	})

	t.Run("Domain -> Subdomain chain", func(t *testing.T) {
		// Build the hierarchy
		domain := &Domain{Name: "example.com"}
		subdomain := NewSubdomain("api.example.com").BelongsTo(domain)
		subdomain.RecordType = "A"
		subdomain.RecordValue = "203.0.113.1"

		// Verify domain (root node)
		assert.Equal(t, "example.com", domain.Name)
		assert.Nil(t, domain.ParentRef())

		// Verify subdomain
		assert.Equal(t, "api.example.com", subdomain.Name)
		assert.Equal(t, "example.com", subdomain.ParentDomain)
		assert.Equal(t, "A", subdomain.RecordType)
		subdomainParent := subdomain.ParentRef()
		require.NotNil(t, subdomainParent)
		assert.Equal(t, graphrag.NodeTypeDomain, subdomainParent.NodeType)
		assert.Equal(t, "example.com", subdomainParent.Properties["name"])
	})
}

// TestBackwardCompatibility ensures legacy patterns still work
func TestBackwardCompatibility(t *testing.T) {
	t.Run("Port with direct HostID still works", func(t *testing.T) {
		port := &Port{
			HostID:   "10.0.0.1",
			Number:   22,
			Protocol: "tcp",
			State:    "open",
		}

		assert.Equal(t, graphrag.NodeTypePort, port.NodeType())

		idProps := port.IdentifyingProperties()
		assert.Equal(t, "10.0.0.1", idProps[graphrag.PropHostID])
		assert.Equal(t, 22, idProps[graphrag.PropNumber])
		assert.Equal(t, "tcp", idProps[graphrag.PropProtocol])

		parentRef := port.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeHost, parentRef.NodeType)
		assert.Equal(t, "10.0.0.1", parentRef.Properties[graphrag.PropIP])
	})

	t.Run("Service with direct PortID still works", func(t *testing.T) {
		service := &Service{
			PortID:  "10.0.0.1:443:tcp",
			Name:    "https",
			Version: "Apache 2.4",
		}

		assert.Equal(t, graphrag.NodeTypeService, service.NodeType())

		parentRef := service.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypePort, parentRef.NodeType)
	})

	t.Run("Subdomain with direct ParentDomain still works", func(t *testing.T) {
		subdomain := &Subdomain{
			ParentDomain: "example.com",
			Name:         "www.example.com",
			RecordType:   "CNAME",
		}

		assert.Equal(t, graphrag.NodeTypeSubdomain, subdomain.NodeType())

		parentRef := subdomain.ParentRef()
		require.NotNil(t, parentRef)
		assert.Equal(t, graphrag.NodeTypeDomain, parentRef.NodeType)
		assert.Equal(t, "example.com", parentRef.Properties["name"])
	})
}

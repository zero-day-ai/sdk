package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHost(t *testing.T) {
	host := NewHost().SetIp("192.168.1.1")

	assert.Equal(t, "host", host.NodeType())
	assert.Equal(t, "192.168.1.1", host.Ip())
	assert.Nil(t, host.ParentRef()) // Host is a root node
}

func TestHostProperties(t *testing.T) {
	host := NewHost().
		SetIp("192.168.1.1").
		SetHostname("server.local").
		SetState("up").
		SetOs("Linux")

	props := host.Properties()
	assert.Equal(t, "192.168.1.1", props["ip"])
	assert.Equal(t, "server.local", props["hostname"])
	assert.Equal(t, "up", props["state"])
	assert.Equal(t, "Linux", props["os"])
}

func TestHostIdentifyingProperties(t *testing.T) {
	host := NewHost().
		SetIp("192.168.1.1").
		SetHostname("server.local")

	idProps := host.IdentifyingProperties()
	assert.Equal(t, "192.168.1.1", idProps["ip"])
	// hostname should NOT be in identifying properties (only ip is)
	_, hasHostname := idProps["hostname"]
	assert.False(t, hasHostname)
}

func TestNewPort(t *testing.T) {
	port := NewPort(443, "tcp")

	assert.Equal(t, "port", port.NodeType())
	assert.Equal(t, int32(443), port.Number())
	assert.Equal(t, "tcp", port.Protocol())
}

func TestPortBelongsToHost(t *testing.T) {
	host := NewHost().SetIp("192.168.1.1")
	port := NewPort(443, "tcp").BelongsTo(host)

	parentRef := port.ParentRef()
	require.NotNil(t, parentRef)
	assert.Equal(t, "host", parentRef.NodeType)
	assert.Equal(t, "HAS_PORT", parentRef.Relationship)
	assert.Equal(t, "192.168.1.1", parentRef.Properties["ip"])
}

func TestNewService(t *testing.T) {
	service := NewService("https")

	assert.Equal(t, "service", service.NodeType())
	assert.Equal(t, "https", service.Name())
}

func TestServiceBelongsToPort(t *testing.T) {
	host := NewHost().SetIp("192.168.1.1")
	port := NewPort(443, "tcp").BelongsTo(host)
	service := NewService("https").BelongsTo(port)

	parentRef := service.ParentRef()
	require.NotNil(t, parentRef)
	assert.Equal(t, "port", parentRef.NodeType)
	assert.Equal(t, "RUNS_SERVICE", parentRef.Relationship)
}

func TestNewDomain(t *testing.T) {
	domain := NewDomain("example.com")

	assert.Equal(t, "domain", domain.NodeType())
	assert.Equal(t, "example.com", domain.Name())
	assert.Nil(t, domain.ParentRef()) // Domain is a root node
}

func TestNewSubdomain(t *testing.T) {
	subdomain := NewSubdomain("www")

	assert.Equal(t, "subdomain", subdomain.NodeType())
	assert.Equal(t, "www", subdomain.Name())
}

func TestSubdomainBelongsToDomain(t *testing.T) {
	domain := NewDomain("example.com")
	subdomain := NewSubdomain("www").BelongsTo(domain)

	parentRef := subdomain.ParentRef()
	require.NotNil(t, parentRef)
	assert.Equal(t, "domain", parentRef.NodeType)
	assert.Equal(t, "HAS_SUBDOMAIN", parentRef.Relationship)
	assert.Equal(t, "example.com", parentRef.Properties["name"])
}

func TestNewFinding(t *testing.T) {
	finding := NewFinding("SQL Injection", "high")

	assert.Equal(t, "finding", finding.NodeType())
	assert.Equal(t, "SQL Injection", finding.Title())
	assert.Equal(t, "high", finding.Severity())
}

func TestFindingSetters(t *testing.T) {
	finding := NewFinding("SQL Injection", "high").
		SetDescription("Found SQL injection in login form").
		SetConfidence(0.95).
		SetCategory("injection").
		SetRemediation("Use parameterized queries")

	assert.Equal(t, "Found SQL injection in login form", finding.Description())
	assert.Equal(t, float64(0.95), finding.Confidence())
	assert.Equal(t, "injection", finding.Category())
	assert.Equal(t, "Use parameterized queries", finding.Remediation())
}

func TestNewTechnology(t *testing.T) {
	tech := NewTechnology("nginx")

	assert.Equal(t, "technology", tech.NodeType())
	assert.Equal(t, "nginx", tech.Name())
}

func TestTechnologySetters(t *testing.T) {
	tech := NewTechnology("nginx").
		SetVersion("1.18.0").
		SetCategory("web-server").
		SetConfidence(90) // 90% confidence as int32

	assert.Equal(t, "1.18.0", tech.Version())
	assert.Equal(t, "web-server", tech.Category())
	assert.Equal(t, int32(90), tech.Confidence())
}

func TestHostValidation(t *testing.T) {
	// Valid host
	host := NewHost().SetIp("192.168.1.1")
	err := host.Validate()
	assert.NoError(t, err)

	// Invalid host - empty IP
	invalidHost := NewHost()
	err = invalidHost.Validate()
	assert.Error(t, err)
}

func TestPortValidation(t *testing.T) {
	// Valid port (with parent)
	host := NewHost().SetIp("192.168.1.1")
	port := NewPort(443, "tcp").BelongsTo(host)
	err := port.Validate()
	assert.NoError(t, err)

	// Invalid port without parent
	portNoParent := NewPort(443, "tcp")
	err = portNoParent.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires a parent")
}

func TestToProto(t *testing.T) {
	host := NewHost().SetIp("192.168.1.1").SetHostname("server.local")
	host.SetID("test-id-123")

	proto := host.ToProto()
	require.NotNil(t, proto)
	assert.Equal(t, "test-id-123", proto.Id)
	assert.Equal(t, "host", proto.Type)
	assert.NotEmpty(t, proto.Properties)
}

func TestIDGetterSetter(t *testing.T) {
	host := NewHost().SetIp("192.168.1.1")

	// Initially empty
	assert.Empty(t, host.ID())

	// Set ID
	host.SetID("node-123")
	assert.Equal(t, "node-123", host.ID())
}

func TestFluentSettersReturnSelf(t *testing.T) {
	// Verify fluent setters return *Host for chaining
	host := NewHost().
		SetIp("192.168.1.1").
		SetHostname("server.local").
		SetState("up").
		SetOs("Linux").
		SetOsVersion("5.4")

	// If we got here without compile errors, chaining works
	assert.Equal(t, "192.168.1.1", host.Ip())
	assert.Equal(t, "server.local", host.Hostname())
}

func TestAnyToValue(t *testing.T) {
	tests := []struct {
		name  string
		input any
	}{
		{"string", "test"},
		{"int", 42},
		{"int32", int32(42)},
		{"int64", int64(42)},
		{"float32", float32(3.14)},
		{"float64", float64(3.14)},
		{"bool", true},
		{"bytes", []byte("data")},
		{"nil", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := anyToValue(tt.input)
			if tt.input == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestPropsToValueMap(t *testing.T) {
	props := map[string]any{
		"ip":       "192.168.1.1",
		"port":     443,
		"active":   true,
		"weight":   0.95,
		"data":     []byte("test"),
	}

	result := propsToValueMap(props)
	assert.Len(t, result, 5)
	assert.NotNil(t, result["ip"])
	assert.NotNil(t, result["port"])
	assert.NotNil(t, result["active"])
	assert.NotNil(t, result["weight"])
	assert.NotNil(t, result["data"])
}

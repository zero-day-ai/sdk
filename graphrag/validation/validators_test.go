// Package validation provides tests for generated validators.
package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/api/gen/graphragpb"
	"google.golang.org/protobuf/proto"
)

func TestIsCoreType(t *testing.T) {
	// Core types should return true
	coreTypes := []string{
		"domain", "subdomain", "host", "port", "service", "endpoint",
		"technology", "certificate", "finding", "evidence",
	}

	for _, ct := range coreTypes {
		assert.True(t, IsCoreType(ct), "type %s should be a core type", ct)
	}

	// Custom types should return false
	customTypes := []string{
		"custom_type", "my_node", "unknown", "", "HOST", "Domain",
		"mission", "mission_run", "agent_run", // These are not in graphragpb
	}

	for _, ct := range customTypes {
		assert.False(t, IsCoreType(ct), "type %s should NOT be a core type", ct)
	}
}

func TestGetParentRequirement(t *testing.T) {
	tests := []struct {
		nodeType       string
		expectFound    bool
		expectParent   string
		expectRelation string
		expectRequired bool
	}{
		{"port", true, "host", "HAS_PORT", true},
		{"service", true, "port", "RUNS_SERVICE", true},
		{"subdomain", true, "domain", "HAS_SUBDOMAIN", true},
		{"endpoint", true, "service", "HAS_ENDPOINT", true},
		{"evidence", true, "finding", "HAS_EVIDENCE", true},
		// Types without parent requirements
		{"host", false, "", "", false},
		{"domain", false, "", "", false},
		{"finding", false, "", "", false},
		{"technology", false, "", "", false},
		{"certificate", false, "", "", false},
		// Custom type
		{"custom", false, "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.nodeType, func(t *testing.T) {
			req, found := GetParentRequirement(tt.nodeType)
			assert.Equal(t, tt.expectFound, found)
			if found {
				assert.Equal(t, tt.expectParent, req.ParentType)
				assert.Equal(t, tt.expectRelation, req.Relationship)
				assert.Equal(t, tt.expectRequired, req.Required)
			}
		})
	}
}

func TestValidateNode(t *testing.T) {
	tests := []struct {
		name       string
		nodeType   string
		properties map[string]any
		hasParent  bool
		wantErr    bool
	}{
		{
			name:       "port with parent passes",
			nodeType:   "port",
			properties: map[string]any{"number": 443},
			hasParent:  true,
			wantErr:    false,
		},
		{
			name:       "port without parent fails",
			nodeType:   "port",
			properties: map[string]any{"number": 443},
			hasParent:  false,
			wantErr:    true,
		},
		{
			name:       "host without parent passes",
			nodeType:   "host",
			properties: map[string]any{"ip": "192.168.1.1"},
			hasParent:  false,
			wantErr:    false,
		},
		{
			name:       "custom type always passes",
			nodeType:   "custom_type",
			properties: map[string]any{},
			hasParent:  false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNode(tt.nodeType, tt.properties, tt.hasParent)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  *graphragpb.Domain
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid domain",
			domain:  &graphragpb.Domain{Name: "example.com"},
			wantErr: false,
		},
		{
			name:    "empty name fails",
			domain:  &graphragpb.Domain{Name: ""},
			wantErr: true,
			errMsg:  "domain name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomain(tt.domain)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateHost(t *testing.T) {
	tests := []struct {
		name    string
		host    *graphragpb.Host
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid host with IP",
			host:    &graphragpb.Host{Ip: "192.168.1.1"},
			wantErr: false,
		},
		{
			name:    "valid host with hostname",
			host:    &graphragpb.Host{Hostname: proto.String("server.example.com")},
			wantErr: false,
		},
		{
			name:    "valid host with both IP and hostname",
			host:    &graphragpb.Host{Ip: "192.168.1.1", Hostname: proto.String("server.example.com")},
			wantErr: false,
		},
		{
			name:    "missing ip and hostname fails",
			host:    &graphragpb.Host{},
			wantErr: true,
			errMsg:  "host requires either ip or hostname",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHost(tt.host)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    *graphragpb.Port
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid port 443",
			port:    &graphragpb.Port{Number: 443, Protocol: "tcp"},
			wantErr: false,
		},
		{
			name:    "valid port 1",
			port:    &graphragpb.Port{Number: 1, Protocol: "tcp"},
			wantErr: false,
		},
		{
			name:    "valid port 65535",
			port:    &graphragpb.Port{Number: 65535, Protocol: "udp"},
			wantErr: false,
		},
		{
			name:    "invalid port 0",
			port:    &graphragpb.Port{Number: 0, Protocol: "tcp"},
			wantErr: true,
			errMsg:  "port number must be between 1 and 65535",
		},
		{
			name:    "invalid port -1",
			port:    &graphragpb.Port{Number: -1, Protocol: "tcp"},
			wantErr: true,
			errMsg:  "port number must be between 1 and 65535",
		},
		{
			name:    "invalid port 65536",
			port:    &graphragpb.Port{Number: 65536, Protocol: "tcp"},
			wantErr: true,
			errMsg:  "port number must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint *graphragpb.Endpoint
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid endpoint",
			endpoint: &graphragpb.Endpoint{Url: "https://example.com/api"},
			wantErr:  false,
		},
		{
			name:     "empty URL fails",
			endpoint: &graphragpb.Endpoint{Url: ""},
			wantErr:  true,
			errMsg:   "endpoint URL cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpoint(tt.endpoint)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateFinding(t *testing.T) {
	tests := []struct {
		name    string
		finding *graphragpb.Finding
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid finding",
			finding: &graphragpb.Finding{Title: "SQL Injection", Severity: "high"},
			wantErr: false,
		},
		{
			name:    "empty title fails",
			finding: &graphragpb.Finding{Title: "", Severity: "high"},
			wantErr: true,
			errMsg:  "finding title cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFinding(tt.finding)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSubdomain(t *testing.T) {
	tests := []struct {
		name      string
		subdomain *graphragpb.Subdomain
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid subdomain",
			subdomain: &graphragpb.Subdomain{Name: "www", ParentDomain: "example.com"},
			wantErr:   false,
		},
		{
			name:      "empty name fails",
			subdomain: &graphragpb.Subdomain{Name: "", ParentDomain: "example.com"},
			wantErr:   true,
			errMsg:    "subdomain name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSubdomain(tt.subdomain)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidatorsWithNoRules tests validators that have no specific rules
func TestValidatorsWithNoRules(t *testing.T) {
	// These validators have no rules and should always pass
	t.Run("Service", func(t *testing.T) {
		err := ValidateService(&graphragpb.Service{Name: "https"})
		assert.NoError(t, err)
	})

	t.Run("Technology", func(t *testing.T) {
		err := ValidateTechnology(&graphragpb.Technology{Name: "nginx"})
		assert.NoError(t, err)
	})

	t.Run("Certificate", func(t *testing.T) {
		err := ValidateCertificate(&graphragpb.Certificate{FingerprintSha256: proto.String("abc123")})
		assert.NoError(t, err)
	})

	t.Run("Evidence", func(t *testing.T) {
		err := ValidateEvidence(&graphragpb.Evidence{FindingId: "test-finding"})
		assert.NoError(t, err)
	})
}

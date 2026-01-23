// Package validation provides tests for generated validators.
package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/api/gen/taxonomypb"
	"google.golang.org/protobuf/proto"
)

func TestIsCoreType(t *testing.T) {
	// Core types should return true
	coreTypes := []string{
		"mission", "mission_run", "agent_run", "tool_execution", "llm_call",
		"domain", "subdomain", "host", "port", "service", "endpoint",
		"technology", "certificate", "finding", "evidence", "technique",
	}

	for _, ct := range coreTypes {
		assert.True(t, IsCoreType(ct), "type %s should be a core type", ct)
	}

	// Custom types should return false
	customTypes := []string{
		"custom_type", "my_node", "unknown", "", "HOST", "Mission",
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
		{"mission_run", true, "mission", "HAS_RUN", true},
		{"agent_run", true, "mission_run", "CONTAINS_AGENT_RUN", true},
		{"evidence", true, "finding", "HAS_EVIDENCE", true},
		// Types without parent requirements
		{"host", false, "", "", false},
		{"domain", false, "", "", false},
		{"mission", false, "", "", false},
		{"finding", false, "", "", false},
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
		name      string
		nodeType  string
		props     map[string]any
		hasParent bool
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "custom type passes without validation",
			nodeType:  "custom_node",
			props:     map[string]any{},
			hasParent: false,
			wantErr:   false,
		},
		{
			name:      "port requires parent",
			nodeType:  "port",
			props:     map[string]any{"number": 443},
			hasParent: false,
			wantErr:   true,
			errMsg:    "requires a parent",
		},
		{
			name:      "port with parent passes",
			nodeType:  "port",
			props:     map[string]any{"number": 443},
			hasParent: true,
			wantErr:   false,
		},
		{
			name:      "service requires parent",
			nodeType:  "service",
			props:     map[string]any{"name": "https"},
			hasParent: false,
			wantErr:   true,
			errMsg:    "requires a parent",
		},
		{
			name:      "host is root - no parent needed",
			nodeType:  "host",
			props:     map[string]any{"ip": "192.168.1.1"},
			hasParent: false,
			wantErr:   false,
		},
		{
			name:      "domain is root - no parent needed",
			nodeType:  "domain",
			props:     map[string]any{"name": "example.com"},
			hasParent: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNode(tt.nodeType, tt.props, tt.hasParent)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMission(t *testing.T) {
	tests := []struct {
		name    string
		mission *taxonomypb.Mission
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid mission",
			mission: &taxonomypb.Mission{
				Name:   "Test Mission",
				Target: "https://example.com",
			},
			wantErr: false,
		},
		{
			name: "empty name fails",
			mission: &taxonomypb.Mission{
				Name:   "",
				Target: "https://example.com",
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "empty target fails",
			mission: &taxonomypb.Mission{
				Name:   "Test Mission",
				Target: "",
			},
			wantErr: true,
			errMsg:  "target cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMission(tt.mission)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDomain(t *testing.T) {
	tests := []struct {
		name    string
		domain  *taxonomypb.Domain
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid domain",
			domain:  &taxonomypb.Domain{Name: "example.com"},
			wantErr: false,
		},
		{
			name:    "empty name fails",
			domain:  &taxonomypb.Domain{Name: ""},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDomain(tt.domain)
			if tt.wantErr {
				assert.Error(t, err)
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
		host    *taxonomypb.Host
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid host with IP",
			host:    &taxonomypb.Host{Ip: proto.String("192.168.1.1")},
			wantErr: false,
		},
		{
			name:    "valid host with hostname",
			host:    &taxonomypb.Host{Hostname: proto.String("server.local")},
			wantErr: false,
		},
		{
			name:    "valid host with both",
			host:    &taxonomypb.Host{Ip: proto.String("192.168.1.1"), Hostname: proto.String("server.local")},
			wantErr: false,
		},
		{
			name:    "invalid host without ip or hostname",
			host:    &taxonomypb.Host{},
			wantErr: true,
			errMsg:  "requires either ip or hostname",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHost(tt.host)
			if tt.wantErr {
				assert.Error(t, err)
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
		port    *taxonomypb.Port
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid port 443",
			port:    &taxonomypb.Port{Number: 443, Protocol: "tcp"},
			wantErr: false,
		},
		{
			name:    "valid port 1",
			port:    &taxonomypb.Port{Number: 1, Protocol: "tcp"},
			wantErr: false,
		},
		{
			name:    "valid port 65535",
			port:    &taxonomypb.Port{Number: 65535, Protocol: "tcp"},
			wantErr: false,
		},
		{
			name:    "invalid port 0",
			port:    &taxonomypb.Port{Number: 0, Protocol: "tcp"},
			wantErr: true,
			errMsg:  "port number must be between 1 and 65535",
		},
		{
			name:    "invalid port -1",
			port:    &taxonomypb.Port{Number: -1, Protocol: "tcp"},
			wantErr: true,
			errMsg:  "port number must be between 1 and 65535",
		},
		{
			name:    "invalid port 65536",
			port:    &taxonomypb.Port{Number: 65536, Protocol: "tcp"},
			wantErr: true,
			errMsg:  "port number must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if tt.wantErr {
				assert.Error(t, err)
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
		endpoint *taxonomypb.Endpoint
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid endpoint",
			endpoint: &taxonomypb.Endpoint{Url: "https://example.com/api"},
			wantErr:  false,
		},
		{
			name:     "empty URL fails",
			endpoint: &taxonomypb.Endpoint{Url: ""},
			wantErr:  true,
			errMsg:   "URL cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEndpoint(tt.endpoint)
			if tt.wantErr {
				assert.Error(t, err)
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
		finding *taxonomypb.Finding
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid finding",
			finding: &taxonomypb.Finding{
				Title:    "SQL Injection",
				Severity: "high",
			},
			wantErr: false,
		},
		{
			name: "valid finding with confidence",
			finding: &taxonomypb.Finding{
				Title:      "XSS",
				Severity:   "medium",
				Confidence: proto.Float64(0.95),
			},
			wantErr: false,
		},
		{
			name: "valid finding with CVSS",
			finding: &taxonomypb.Finding{
				Title:     "RCE",
				Severity:  "critical",
				CvssScore: proto.Float64(9.8),
			},
			wantErr: false,
		},
		{
			name: "empty title fails",
			finding: &taxonomypb.Finding{
				Title:    "",
				Severity: "high",
			},
			wantErr: true,
			errMsg:  "title cannot be empty",
		},
		{
			name: "confidence above 1.0 fails",
			finding: &taxonomypb.Finding{
				Title:      "Test",
				Severity:   "low",
				Confidence: proto.Float64(1.5),
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "negative confidence fails",
			finding: &taxonomypb.Finding{
				Title:      "Test",
				Severity:   "low",
				Confidence: proto.Float64(-0.1),
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "CVSS above 10.0 fails",
			finding: &taxonomypb.Finding{
				Title:     "Test",
				Severity:  "low",
				CvssScore: proto.Float64(10.5),
			},
			wantErr: true,
			errMsg:  "CVSS score must be between 0.0 and 10.0",
		},
		{
			name: "negative CVSS fails",
			finding: &taxonomypb.Finding{
				Title:     "Test",
				Severity:  "low",
				CvssScore: proto.Float64(-1.0),
			},
			wantErr: true,
			errMsg:  "CVSS score must be between 0.0 and 10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFinding(tt.finding)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTypesWithoutRules(t *testing.T) {
	// These types have no specific validation rules (just parent requirements)
	// but we should ensure their validators don't panic

	t.Run("MissionRun", func(t *testing.T) {
		err := ValidateMissionRun(&taxonomypb.MissionRun{})
		assert.NoError(t, err)
	})

	t.Run("AgentRun", func(t *testing.T) {
		err := ValidateAgentRun(&taxonomypb.AgentRun{})
		assert.NoError(t, err)
	})

	t.Run("ToolExecution", func(t *testing.T) {
		err := ValidateToolExecution(&taxonomypb.ToolExecution{})
		assert.NoError(t, err)
	})

	t.Run("LlmCall", func(t *testing.T) {
		err := ValidateLlmCall(&taxonomypb.LlmCall{})
		assert.NoError(t, err)
	})

	t.Run("Subdomain", func(t *testing.T) {
		err := ValidateSubdomain(&taxonomypb.Subdomain{})
		assert.NoError(t, err)
	})

	t.Run("Service", func(t *testing.T) {
		err := ValidateService(&taxonomypb.Service{})
		assert.NoError(t, err)
	})

	t.Run("Technology", func(t *testing.T) {
		err := ValidateTechnology(&taxonomypb.Technology{})
		assert.NoError(t, err)
	})

	t.Run("Certificate", func(t *testing.T) {
		err := ValidateCertificate(&taxonomypb.Certificate{})
		assert.NoError(t, err)
	})

	t.Run("Evidence", func(t *testing.T) {
		err := ValidateEvidence(&taxonomypb.Evidence{})
		assert.NoError(t, err)
	})

	t.Run("Technique", func(t *testing.T) {
		err := ValidateTechnique(&taxonomypb.Technique{})
		assert.NoError(t, err)
	})
}

func TestParentRequirement_AllChildTypes(t *testing.T) {
	// Verify all child types have parent requirements defined
	childTypes := []string{
		"mission_run", "agent_run", "tool_execution", "llm_call",
		"subdomain", "port", "service", "endpoint", "evidence",
	}

	for _, ct := range childTypes {
		t.Run(ct, func(t *testing.T) {
			req, found := GetParentRequirement(ct)
			require.True(t, found, "child type %s should have a parent requirement", ct)
			assert.NotEmpty(t, req.ParentType, "parent type should not be empty")
			assert.NotEmpty(t, req.Relationship, "relationship should not be empty")
		})
	}
}

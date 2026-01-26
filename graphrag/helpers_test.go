// Code generated tests for helpers_generated.go
package graphrag

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/api/gen/taxonomypb"
)

// ==================== ROOT TYPE TESTS ====================

func TestNewMission(t *testing.T) {
	mission := NewMission("test-mission", "https://target.com")

	require.NotNil(t, mission)
	assert.NotEmpty(t, mission.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(mission.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, "test-mission", mission.Name)
	assert.Equal(t, "https://target.com", mission.Target)
}

func TestNewDomain(t *testing.T) {
	domain := NewDomain("example.com")

	require.NotNil(t, domain)
	assert.NotEmpty(t, domain.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(domain.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, "example.com", domain.Name)
}

func TestNewHost(t *testing.T) {
	host := NewHost()

	require.NotNil(t, host)
	assert.NotEmpty(t, host.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(host.Id)
	assert.NoError(t, err, "Id should be a valid UUID")
}

func TestNewTechnology(t *testing.T) {
	tech := NewTechnology("nginx")

	require.NotNil(t, tech)
	assert.NotEmpty(t, tech.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(tech.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, "nginx", tech.Name)
}

func TestNewCertificate(t *testing.T) {
	cert := NewCertificate()

	require.NotNil(t, cert)
	assert.NotEmpty(t, cert.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(cert.Id)
	assert.NoError(t, err, "Id should be a valid UUID")
}

func TestNewFinding(t *testing.T) {
	finding := NewFinding("SQL Injection", "critical")

	require.NotNil(t, finding)
	assert.NotEmpty(t, finding.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(finding.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, "SQL Injection", finding.Title)
	assert.Equal(t, "critical", finding.Severity)
}

func TestNewTechnique(t *testing.T) {
	technique := NewTechnique("T1190", "Exploit Public-Facing Application")

	require.NotNil(t, technique)
	assert.NotEmpty(t, technique.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(technique.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, "T1190", technique.TechniqueId)
	assert.Equal(t, "Exploit Public-Facing Application", technique.Name)
}

// ==================== CHILD TYPE TESTS ====================

func TestNewMissionRun(t *testing.T) {
	mission := NewMission("test-mission", "https://target.com")
	run := NewMissionRun(mission, 1)

	require.NotNil(t, run)
	assert.NotEmpty(t, run.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(run.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, mission.Id, run.ParentMissionId)
	assert.Equal(t, int32(1), run.RunNumber)
}

func TestNewMissionRun_PanicsOnEmptyParentId(t *testing.T) {
	mission := &taxonomypb.Mission{Name: "test-mission"} // No Id set

	assert.PanicsWithValue(t,
		"parent Mission must have Id set - use NewMission() or set Id manually",
		func() {
			NewMissionRun(mission, 1)
		},
	)
}

func TestNewAgentRun(t *testing.T) {
	mission := NewMission("test-mission", "https://target.com")
	missionRun := NewMissionRun(mission, 1)
	agentRun := NewAgentRun(missionRun, "network-recon")

	require.NotNil(t, agentRun)
	assert.NotEmpty(t, agentRun.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(agentRun.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, missionRun.Id, agentRun.ParentMissionRunId)
	assert.Equal(t, "network-recon", agentRun.AgentName)
}

func TestNewAgentRun_PanicsOnEmptyParentId(t *testing.T) {
	missionRun := &taxonomypb.MissionRun{RunNumber: 1} // No Id set

	assert.PanicsWithValue(t,
		"parent MissionRun must have Id set - use NewMissionRun() or set Id manually",
		func() {
			NewAgentRun(missionRun, "network-recon")
		},
	)
}

func TestNewToolExecution(t *testing.T) {
	mission := NewMission("test-mission", "https://target.com")
	missionRun := NewMissionRun(mission, 1)
	agentRun := NewAgentRun(missionRun, "network-recon")
	toolExec := NewToolExecution(agentRun, "nmap")

	require.NotNil(t, toolExec)
	assert.NotEmpty(t, toolExec.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(toolExec.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, agentRun.Id, toolExec.ParentAgentRunId)
	assert.Equal(t, "nmap", toolExec.ToolName)
}

func TestNewToolExecution_PanicsOnEmptyParentId(t *testing.T) {
	agentRun := &taxonomypb.AgentRun{AgentName: "network-recon"} // No Id set

	assert.PanicsWithValue(t,
		"parent AgentRun must have Id set - use NewAgentRun() or set Id manually",
		func() {
			NewToolExecution(agentRun, "nmap")
		},
	)
}

func TestNewLlmCall(t *testing.T) {
	mission := NewMission("test-mission", "https://target.com")
	missionRun := NewMissionRun(mission, 1)
	agentRun := NewAgentRun(missionRun, "network-recon")
	llmCall := NewLlmCall(agentRun, "claude-3-opus")

	require.NotNil(t, llmCall)
	assert.NotEmpty(t, llmCall.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(llmCall.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, agentRun.Id, llmCall.ParentAgentRunId)
	assert.Equal(t, "claude-3-opus", llmCall.Model)
}

func TestNewLlmCall_PanicsOnEmptyParentId(t *testing.T) {
	agentRun := &taxonomypb.AgentRun{AgentName: "network-recon"} // No Id set

	assert.PanicsWithValue(t,
		"parent AgentRun must have Id set - use NewAgentRun() or set Id manually",
		func() {
			NewLlmCall(agentRun, "claude-3-opus")
		},
	)
}

func TestNewSubdomain(t *testing.T) {
	domain := NewDomain("example.com")
	subdomain := NewSubdomain(domain, "www.example.com")

	require.NotNil(t, subdomain)
	assert.NotEmpty(t, subdomain.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(subdomain.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, domain.Id, subdomain.ParentDomainId)
	assert.Equal(t, "www.example.com", subdomain.Name)
}

func TestNewSubdomain_PanicsOnEmptyParentId(t *testing.T) {
	domain := &taxonomypb.Domain{Name: "example.com"} // No Id set

	assert.PanicsWithValue(t,
		"parent Domain must have Id set - use NewDomain() or set Id manually",
		func() {
			NewSubdomain(domain, "www.example.com")
		},
	)
}

func TestNewPort(t *testing.T) {
	host := NewHost()
	port := NewPort(host, 80, "tcp")

	require.NotNil(t, port)
	assert.NotEmpty(t, port.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(port.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, host.Id, port.ParentHostId)
	assert.Equal(t, int32(80), port.Number)
	assert.Equal(t, "tcp", port.Protocol)
}

func TestNewPort_PanicsOnEmptyParentId(t *testing.T) {
	host := &taxonomypb.Host{} // No Id set

	assert.PanicsWithValue(t,
		"parent Host must have Id set - use NewHost() or set Id manually",
		func() {
			NewPort(host, 80, "tcp")
		},
	)
}

func TestNewService(t *testing.T) {
	host := NewHost()
	port := NewPort(host, 80, "tcp")
	service := NewService(port, "http")

	require.NotNil(t, service)
	assert.NotEmpty(t, service.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(service.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, port.Id, service.ParentPortId)
	assert.Equal(t, "http", service.Name)
}

func TestNewService_PanicsOnEmptyParentId(t *testing.T) {
	port := &taxonomypb.Port{Number: 80, Protocol: "tcp"} // No Id set

	assert.PanicsWithValue(t,
		"parent Port must have Id set - use NewPort() or set Id manually",
		func() {
			NewService(port, "http")
		},
	)
}

func TestNewEndpoint(t *testing.T) {
	host := NewHost()
	port := NewPort(host, 443, "tcp")
	service := NewService(port, "https")
	endpoint := NewEndpoint(service, "https://api.example.com/v1")

	require.NotNil(t, endpoint)
	assert.NotEmpty(t, endpoint.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(endpoint.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, service.Id, endpoint.ParentServiceId)
	assert.Equal(t, "https://api.example.com/v1", endpoint.Url)
}

func TestNewEndpoint_PanicsOnEmptyParentId(t *testing.T) {
	service := &taxonomypb.Service{Name: "https"} // No Id set

	assert.PanicsWithValue(t,
		"parent Service must have Id set - use NewService() or set Id manually",
		func() {
			NewEndpoint(service, "https://api.example.com/v1")
		},
	)
}

func TestNewEvidence(t *testing.T) {
	finding := NewFinding("SQL Injection", "critical")
	evidence := NewEvidence(finding, "request")

	require.NotNil(t, evidence)
	assert.NotEmpty(t, evidence.Id)

	// Verify it's a valid UUID
	_, err := uuid.Parse(evidence.Id)
	assert.NoError(t, err, "Id should be a valid UUID")

	assert.Equal(t, finding.Id, evidence.ParentFindingId)
	assert.Equal(t, "request", evidence.Type)
}

func TestNewEvidence_PanicsOnEmptyParentId(t *testing.T) {
	finding := &taxonomypb.Finding{Title: "SQL Injection"} // No Id set

	assert.PanicsWithValue(t,
		"parent Finding must have Id set - use NewFinding() or set Id manually",
		func() {
			NewEvidence(finding, "request")
		},
	)
}

// ==================== INTEGRATION TESTS ====================

func TestHelpers_FullHierarchy(t *testing.T) {
	// Create a complete mission hierarchy
	mission := NewMission("pentest-2024", "https://target.com")
	missionRun := NewMissionRun(mission, 1)
	agentRun := NewAgentRun(missionRun, "network-recon")
	toolExec := NewToolExecution(agentRun, "nmap")
	llmCall := NewLlmCall(agentRun, "claude-3-opus")

	// Verify all IDs are unique
	ids := []string{
		mission.Id,
		missionRun.Id,
		agentRun.Id,
		toolExec.Id,
		llmCall.Id,
	}

	uniqueIds := make(map[string]bool)
	for _, id := range ids {
		assert.NotEmpty(t, id)
		assert.False(t, uniqueIds[id], "Duplicate ID detected: %s", id)
		uniqueIds[id] = true

		// Verify each is a valid UUID
		_, err := uuid.Parse(id)
		assert.NoError(t, err, "Invalid UUID: %s", id)
	}

	// Verify parent relationships
	assert.Equal(t, mission.Id, missionRun.ParentMissionId)
	assert.Equal(t, missionRun.Id, agentRun.ParentMissionRunId)
	assert.Equal(t, agentRun.Id, toolExec.ParentAgentRunId)
	assert.Equal(t, agentRun.Id, llmCall.ParentAgentRunId)
}

func TestHelpers_NetworkHierarchy(t *testing.T) {
	// Create a complete network hierarchy
	host := NewHost()
	port := NewPort(host, 443, "tcp")
	service := NewService(port, "https")
	endpoint := NewEndpoint(service, "https://api.example.com")

	// Verify all IDs are unique
	ids := []string{
		host.Id,
		port.Id,
		service.Id,
		endpoint.Id,
	}

	uniqueIds := make(map[string]bool)
	for _, id := range ids {
		assert.NotEmpty(t, id)
		assert.False(t, uniqueIds[id], "Duplicate ID detected: %s", id)
		uniqueIds[id] = true

		// Verify each is a valid UUID
		_, err := uuid.Parse(id)
		assert.NoError(t, err, "Invalid UUID: %s", id)
	}

	// Verify parent relationships
	assert.Equal(t, host.Id, port.ParentHostId)
	assert.Equal(t, port.Id, service.ParentPortId)
	assert.Equal(t, service.Id, endpoint.ParentServiceId)
}

func TestHelpers_FindingHierarchy(t *testing.T) {
	// Create finding with evidence
	finding := NewFinding("XSS Vulnerability", "high")
	evidence1 := NewEvidence(finding, "request")
	evidence2 := NewEvidence(finding, "response")

	// Verify all IDs are unique
	ids := []string{
		finding.Id,
		evidence1.Id,
		evidence2.Id,
	}

	uniqueIds := make(map[string]bool)
	for _, id := range ids {
		assert.NotEmpty(t, id)
		assert.False(t, uniqueIds[id], "Duplicate ID detected: %s", id)
		uniqueIds[id] = true

		// Verify each is a valid UUID
		_, err := uuid.Parse(id)
		assert.NoError(t, err, "Invalid UUID: %s", id)
	}

	// Verify parent relationships
	assert.Equal(t, finding.Id, evidence1.ParentFindingId)
	assert.Equal(t, finding.Id, evidence2.ParentFindingId)
}

func TestHelpers_DomainHierarchy(t *testing.T) {
	// Create domain with subdomains
	domain := NewDomain("example.com")
	sub1 := NewSubdomain(domain, "www.example.com")
	sub2 := NewSubdomain(domain, "api.example.com")

	// Verify all IDs are unique
	ids := []string{
		domain.Id,
		sub1.Id,
		sub2.Id,
	}

	uniqueIds := make(map[string]bool)
	for _, id := range ids {
		assert.NotEmpty(t, id)
		assert.False(t, uniqueIds[id], "Duplicate ID detected: %s", id)
		uniqueIds[id] = true

		// Verify each is a valid UUID
		_, err := uuid.Parse(id)
		assert.NoError(t, err, "Invalid UUID: %s", id)
	}

	// Verify parent relationships
	assert.Equal(t, domain.Id, sub1.ParentDomainId)
	assert.Equal(t, domain.Id, sub2.ParentDomainId)
}

// ==================== EDGE CASE TESTS ====================

func TestHelpers_EmptyStringParameters(t *testing.T) {
	// Test that helpers accept empty strings (validation is separate)
	mission := NewMission("", "")
	assert.NotEmpty(t, mission.Id)
	assert.Equal(t, "", mission.Name)
	assert.Equal(t, "", mission.Target)

	domain := NewDomain("")
	assert.NotEmpty(t, domain.Id)
	assert.Equal(t, "", domain.Name)
}

func TestHelpers_UniqueIdsOnMultipleCalls(t *testing.T) {
	// Verify that multiple calls generate different UUIDs
	mission1 := NewMission("test1", "https://target1.com")
	mission2 := NewMission("test2", "https://target2.com")

	assert.NotEqual(t, mission1.Id, mission2.Id,
		"Multiple calls should generate different UUIDs")
}

func TestHelpers_ParentIdImmutability(t *testing.T) {
	// Verify that child keeps reference to parent even if parent Id changes later
	host := NewHost()
	originalHostId := host.Id

	port := NewPort(host, 80, "tcp")
	assert.Equal(t, originalHostId, port.ParentHostId)

	// Change parent's Id (shouldn't affect child's reference)
	host.Id = uuid.New().String()
	assert.Equal(t, originalHostId, port.ParentHostId,
		"Child should keep original parent Id")
}

// ==================== BENCHMARK TESTS ====================

func BenchmarkNewMission(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewMission("benchmark-mission", "https://target.com")
	}
}

func BenchmarkNewMissionRun(b *testing.B) {
	mission := NewMission("benchmark-mission", "https://target.com")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewMissionRun(mission, 1)
	}
}

func BenchmarkNewFullHierarchy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mission := NewMission("benchmark-mission", "https://target.com")
		missionRun := NewMissionRun(mission, 1)
		agentRun := NewAgentRun(missionRun, "network-recon")
		_ = NewToolExecution(agentRun, "nmap")
		_ = NewLlmCall(agentRun, "claude-3-opus")
	}
}

func BenchmarkNewNetworkHierarchy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		host := NewHost()
		port := NewPort(host, 443, "tcp")
		service := NewService(port, "https")
		_ = NewEndpoint(service, "https://api.example.com")
	}
}

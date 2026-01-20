package domain

import "github.com/zero-day-ai/sdk/graphrag"

// AgentRun represents a single execution run of an agent within a mission.
// AgentRuns are the primary execution context for tracking discoveries and tool executions.
//
// Hierarchy: AgentRun is a root node (no parent in the domain hierarchy)
// Note: AgentRuns are typically linked to missions via PART_OF_MISSION relationships,
// but this is handled separately from the parent hierarchy.
//
// Identifying Properties: id, mission_id, agent_name, run_number
// Parent: None (root node in domain hierarchy)
//
// Example:
//
//	agentRun := &AgentRun{
//	    ID:        "run-123",
//	    MissionID: "mission-456",
//	    AgentName: "network-recon",
//	    RunNumber: 1,
//	    StartTime: "2024-01-20T10:00:00Z",
//	    Status:    "running",
//	}
type AgentRun struct {
	// ID is the unique identifier for this agent run.
	// This is an identifying property.
	ID string

	// MissionID is the mission this agent run belongs to.
	// This is an identifying property.
	MissionID string

	// AgentName is the name of the agent being run.
	// This is an identifying property.
	AgentName string

	// RunNumber is the sequence number of this run within the mission.
	// This is an identifying property.
	RunNumber int

	// StartTime is the agent run start timestamp (optional).
	StartTime string

	// EndTime is the agent run end timestamp (optional).
	EndTime string

	// Status is the agent run status (e.g., "running", "completed", "failed") (optional).
	Status string

	// Error is the error message if the run failed (optional).
	Error string

	// Duration is the run duration in seconds (optional).
	Duration float64
}

// NodeType returns the canonical node type for agent runs.
func (a *AgentRun) NodeType() string {
	return graphrag.NodeTypeAgentRun
}

// IdentifyingProperties returns the properties that uniquely identify this agent run.
// An agent run is identified by its ID, mission ID, agent name, and run number.
func (a *AgentRun) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id":          a.ID,
		"mission_id":  a.MissionID,
		"agent_name":  a.AgentName,
		"run_number":  a.RunNumber,
	}
}

// Properties returns all properties to set on the agent run node.
func (a *AgentRun) Properties() map[string]any {
	props := map[string]any{
		"id":          a.ID,
		"mission_id":  a.MissionID,
		"agent_name":  a.AgentName,
		"run_number":  a.RunNumber,
	}

	// Add optional properties if present
	if a.StartTime != "" {
		props["start_time"] = a.StartTime
	}
	if a.EndTime != "" {
		props["end_time"] = a.EndTime
	}
	if a.Status != "" {
		props["status"] = a.Status
	}
	if a.Error != "" {
		props["error"] = a.Error
	}
	if a.Duration != 0 {
		props["duration"] = a.Duration
	}

	return props
}

// ParentRef returns nil because AgentRun is a root node in the domain hierarchy.
// Note: AgentRuns may have PART_OF_MISSION relationships to missions, but this is
// handled separately from the parent hierarchy in the graph loader.
func (a *AgentRun) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because AgentRun has no parent.
func (a *AgentRun) RelationshipType() string {
	return ""
}

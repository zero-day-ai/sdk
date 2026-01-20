package domain

import "github.com/zero-day-ai/sdk/graphrag"

// ToolExecution represents the execution of a security tool (nmap, nuclei, sqlmap, etc.).
// ToolExecutions are linked to agent runs and track individual tool invocations.
//
// Hierarchy: AgentRun -> ToolExecution
//
// Identifying Properties: id, agent_run_id, tool_name, sequence
// Parent: AgentRun (via PART_OF relationship)
//
// Example:
//
//	toolExecution := &ToolExecution{
//	    ID:          "exec-789",
//	    AgentRunID:  "run-123",
//	    ToolName:    "nmap",
//	    Sequence:    1,
//	    StartTime:   "2024-01-20T10:05:00Z",
//	    Duration:    5.2,
//	}
type ToolExecution struct {
	// ID is the unique identifier for this tool execution.
	// This is an identifying property.
	ID string

	// AgentRunID is the ID of the agent run that executed this tool.
	// This is an identifying property.
	AgentRunID string

	// ToolName is the name of the tool that was executed.
	// This is an identifying property.
	ToolName string

	// Sequence is the sequence number of this execution within the agent run.
	// This is an identifying property.
	Sequence int

	// StartTime is the tool execution start timestamp (optional).
	StartTime string

	// EndTime is the tool execution end timestamp (optional).
	EndTime string

	// Duration is the execution duration in seconds (optional).
	Duration float64

	// Status is the execution status (e.g., "success", "failed", "timeout") (optional).
	Status string

	// Error is the error message if the execution failed (optional).
	Error string

	// ExitCode is the tool exit code (optional).
	ExitCode int

	// Command is the full command that was executed (optional).
	Command string
}

// NodeType returns the canonical node type for tool executions.
func (t *ToolExecution) NodeType() string {
	return graphrag.NodeTypeToolExecution
}

// IdentifyingProperties returns the properties that uniquely identify this tool execution.
// A tool execution is identified by its ID, agent run ID, tool name, and sequence number.
func (t *ToolExecution) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id":           t.ID,
		"agent_run_id": t.AgentRunID,
		"tool_name":    t.ToolName,
		"sequence":     t.Sequence,
	}
}

// Properties returns all properties to set on the tool execution node.
func (t *ToolExecution) Properties() map[string]any {
	props := map[string]any{
		"id":           t.ID,
		"agent_run_id": t.AgentRunID,
		"tool_name":    t.ToolName,
		"sequence":     t.Sequence,
	}

	// Add optional properties if present
	if t.StartTime != "" {
		props["start_time"] = t.StartTime
	}
	if t.EndTime != "" {
		props["end_time"] = t.EndTime
	}
	if t.Duration != 0 {
		props["duration"] = t.Duration
	}
	if t.Status != "" {
		props["status"] = t.Status
	}
	if t.Error != "" {
		props["error"] = t.Error
	}
	if t.ExitCode != 0 {
		props["exit_code"] = t.ExitCode
	}
	if t.Command != "" {
		props["command"] = t.Command
	}

	return props
}

// ParentRef returns a reference to the parent AgentRun node.
func (t *ToolExecution) ParentRef() *NodeRef {
	if t.AgentRunID == "" {
		return nil
	}

	// AgentRun is identified by agent_run_id
	// We need to construct a reference with the minimal identifying properties
	return &NodeRef{
		NodeType: graphrag.NodeTypeAgentRun,
		Properties: map[string]any{
			"id": t.AgentRunID,
		},
	}
}

// RelationshipType returns the relationship type to the parent agent run.
func (t *ToolExecution) RelationshipType() string {
	return graphrag.RelTypePartOf
}

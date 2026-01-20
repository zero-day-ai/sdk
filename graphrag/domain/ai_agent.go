package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// AIAgent represents an AI agent in the knowledge graph.
// An AIAgent is identified by its name.
// AIAgent is a root-level node with no parent relationships.
//
// Example:
//
//	agent := &AIAgent{
//	    Name:        "security-analyzer",
//	    Type:        "autonomous",
//	    Version:     "1.2.0",
//	    Description: "Analyzes code for security vulnerabilities",
//	}
//
// Identifying Properties:
//   - name (required): Agent name
//
// Relationships:
//   - None (root node)
//   - Children: AgentConfig, AgentMemory, AgentTask nodes
type AIAgent struct {
	// Name is the agent name.
	// This is an identifying property and is required.
	Name string

	// Type is the agent type.
	// Optional. Examples: "autonomous", "reactive", "collaborative"
	Type string

	// Version is the agent version.
	// Optional. Example: "1.2.0"
	Version string

	// Description describes the agent's purpose.
	// Optional.
	Description string
}

// NodeType returns the canonical node type for AIAgent nodes.
// Implements GraphNode interface.
func (a *AIAgent) NodeType() string {
	return graphrag.NodeTypeAIAgent
}

// IdentifyingProperties returns the properties that uniquely identify this agent.
// For AIAgent nodes, only name is identifying.
// Implements GraphNode interface.
func (a *AIAgent) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: a.Name,
	}
}

// Properties returns all properties to set on the AIAgent node.
// Implements GraphNode interface.
func (a *AIAgent) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: a.Name,
	}

	if a.Type != "" {
		props["type"] = a.Type
	}
	if a.Version != "" {
		props["version"] = a.Version
	}
	if a.Description != "" {
		props[graphrag.PropDescription] = a.Description
	}

	return props
}

// ParentRef returns nil because AIAgent is a root node with no parent.
// Implements GraphNode interface.
func (a *AIAgent) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because AIAgent is a root node.
// Implements GraphNode interface.
func (a *AIAgent) RelationshipType() string {
	return ""
}

// AgentConfig represents the configuration for an AI agent.
// An AgentConfig is identified by its agent ID.
// AgentConfig nodes are children of AIAgent nodes.
//
// Example:
//
//	config := &AgentConfig{
//	    AgentID:     "security-analyzer",
//	    Temperature: 0.7,
//	    MaxTokens:   2000,
//	    Model:       "claude-3-opus",
//	}
//
// Identifying Properties:
//   - agent_id (required): Parent agent name
//
// Relationships:
//   - Parent: AIAgent node (via HAS_CONFIG relationship)
type AgentConfig struct {
	// AgentID is the parent agent name.
	// This is an identifying property and is required.
	AgentID string

	// Temperature is the LLM temperature setting.
	// Optional.
	Temperature float64

	// MaxTokens is the maximum output tokens.
	// Optional.
	MaxTokens int

	// Model is the LLM model identifier.
	// Optional.
	Model string
}

// NodeType returns the canonical node type for AgentConfig nodes.
// Implements GraphNode interface.
func (c *AgentConfig) NodeType() string {
	return graphrag.NodeTypeAgentConfig
}

// IdentifyingProperties returns the properties that uniquely identify this config.
// For AgentConfig nodes, only agent_id is identifying.
// Implements GraphNode interface.
func (c *AgentConfig) IdentifyingProperties() map[string]any {
	return map[string]any{
		"agent_id": c.AgentID,
	}
}

// Properties returns all properties to set on the AgentConfig node.
// Implements GraphNode interface.
func (c *AgentConfig) Properties() map[string]any {
	props := map[string]any{
		"agent_id": c.AgentID,
	}

	if c.Temperature > 0 {
		props["temperature"] = c.Temperature
	}
	if c.MaxTokens > 0 {
		props["max_tokens"] = c.MaxTokens
	}
	if c.Model != "" {
		props[graphrag.PropModel] = c.Model
	}

	return props
}

// ParentRef returns a reference to the parent AIAgent node.
// Implements GraphNode interface.
func (c *AgentConfig) ParentRef() *NodeRef {
	if c.AgentID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeAIAgent,
		Properties: map[string]any{
			graphrag.PropName: c.AgentID,
		},
	}
}

// RelationshipType returns the relationship type to the parent AIAgent node.
// Implements GraphNode interface.
func (c *AgentConfig) RelationshipType() string {
	return "HAS_CONFIG"
}

// AgentMemory represents memory storage for an AI agent.
// An AgentMemory is identified by its agent ID and memory type.
// AgentMemory nodes are children of AIAgent nodes.
//
// Example:
//
//	memory := &AgentMemory{
//	    AgentID:    "security-analyzer",
//	    MemoryType: "working",
//	    Size:       1024,
//	    Backend:    "redis",
//	}
//
// Identifying Properties:
//   - agent_id (required): Parent agent name
//   - memory_type (required): Memory type (working, mission, long_term)
//
// Relationships:
//   - Parent: AIAgent node (via HAS_MEMORY relationship)
type AgentMemory struct {
	// AgentID is the parent agent name.
	// This is an identifying property and is required.
	AgentID string

	// MemoryType is the type of memory.
	// This is an identifying property and is required.
	// Examples: "working", "mission", "long_term"
	MemoryType string

	// Size is the memory size or entry count.
	// Optional.
	Size int

	// Backend is the storage backend.
	// Optional. Examples: "redis", "postgres", "vector_store"
	Backend string
}

// NodeType returns the canonical node type for AgentMemory nodes.
// Implements GraphNode interface.
func (m *AgentMemory) NodeType() string {
	return graphrag.NodeTypeAgentMemory
}

// IdentifyingProperties returns the properties that uniquely identify this memory.
// For AgentMemory nodes, agent_id and memory_type are both identifying.
// Implements GraphNode interface.
func (m *AgentMemory) IdentifyingProperties() map[string]any {
	return map[string]any{
		"agent_id":    m.AgentID,
		"memory_type": m.MemoryType,
	}
}

// Properties returns all properties to set on the AgentMemory node.
// Implements GraphNode interface.
func (m *AgentMemory) Properties() map[string]any {
	props := map[string]any{
		"agent_id":    m.AgentID,
		"memory_type": m.MemoryType,
	}

	if m.Size > 0 {
		props["size"] = m.Size
	}
	if m.Backend != "" {
		props["backend"] = m.Backend
	}

	return props
}

// ParentRef returns a reference to the parent AIAgent node.
// Implements GraphNode interface.
func (m *AgentMemory) ParentRef() *NodeRef {
	if m.AgentID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeAIAgent,
		Properties: map[string]any{
			graphrag.PropName: m.AgentID,
		},
	}
}

// RelationshipType returns the relationship type to the parent AIAgent node.
// Implements GraphNode interface.
func (m *AgentMemory) RelationshipType() string {
	return "HAS_MEMORY"
}

// AgentTool represents a tool available to an AI agent.
// An AgentTool is identified by its agent ID and tool name.
// AgentTool nodes are children of AIAgent nodes.
//
// Example:
//
//	tool := &AgentTool{
//	    AgentID:     "security-analyzer",
//	    Name:        "nmap",
//	    Description: "Network port scanner",
//	    Version:     "7.94",
//	}
//
// Identifying Properties:
//   - agent_id (required): Parent agent name
//   - name (required): Tool name
//
// Relationships:
//   - Parent: AIAgent node (via HAS_TOOL relationship)
type AgentTool struct {
	// AgentID is the parent agent name.
	// This is an identifying property and is required.
	AgentID string

	// Name is the tool name.
	// This is an identifying property and is required.
	Name string

	// Description describes the tool's purpose.
	// Optional.
	Description string

	// Version is the tool version.
	// Optional.
	Version string
}

// NodeType returns the canonical node type for AgentTool nodes.
// Implements GraphNode interface.
func (t *AgentTool) NodeType() string {
	return graphrag.NodeTypeAgentTool
}

// IdentifyingProperties returns the properties that uniquely identify this agent tool.
// For AgentTool nodes, agent_id and name are both identifying.
// Implements GraphNode interface.
func (t *AgentTool) IdentifyingProperties() map[string]any {
	return map[string]any{
		"agent_id":        t.AgentID,
		graphrag.PropName: t.Name,
	}
}

// Properties returns all properties to set on the AgentTool node.
// Implements GraphNode interface.
func (t *AgentTool) Properties() map[string]any {
	props := map[string]any{
		"agent_id":        t.AgentID,
		graphrag.PropName: t.Name,
	}

	if t.Description != "" {
		props[graphrag.PropDescription] = t.Description
	}
	if t.Version != "" {
		props["version"] = t.Version
	}

	return props
}

// ParentRef returns a reference to the parent AIAgent node.
// Implements GraphNode interface.
func (t *AgentTool) ParentRef() *NodeRef {
	if t.AgentID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeAIAgent,
		Properties: map[string]any{
			graphrag.PropName: t.AgentID,
		},
	}
}

// RelationshipType returns the relationship type to the parent AIAgent node.
// Implements GraphNode interface.
func (t *AgentTool) RelationshipType() string {
	return "HAS_TOOL"
}

// Chain represents a chain of agents or LLM calls.
// A Chain is identified by its name.
// Chain is a root-level node with no parent relationships.
//
// Example:
//
//	chain := &Chain{
//	    Name:        "security-pipeline",
//	    Description: "Multi-stage security analysis pipeline",
//	    Steps:       5,
//	}
//
// Identifying Properties:
//   - name (required): Chain name
//
// Relationships:
//   - None (root node)
type Chain struct {
	// Name is the chain name.
	// This is an identifying property and is required.
	Name string

	// Description describes the chain's purpose.
	// Optional.
	Description string

	// Steps is the number of steps in the chain.
	// Optional.
	Steps int
}

// NodeType returns the canonical node type for Chain nodes.
// Implements GraphNode interface.
func (c *Chain) NodeType() string {
	return graphrag.NodeTypeChain
}

// IdentifyingProperties returns the properties that uniquely identify this chain.
// For Chain nodes, only name is identifying.
// Implements GraphNode interface.
func (c *Chain) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
	}
}

// Properties returns all properties to set on the Chain node.
// Implements GraphNode interface.
func (c *Chain) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
	}

	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	if c.Steps > 0 {
		props["steps"] = c.Steps
	}

	return props
}

// ParentRef returns nil because Chain is a root node with no parent.
// Implements GraphNode interface.
func (c *Chain) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Chain is a root node.
// Implements GraphNode interface.
func (c *Chain) RelationshipType() string {
	return ""
}

// Workflow represents an agent workflow or orchestration pattern.
// A Workflow is identified by its name.
// Workflow is a root-level node with no parent relationships.
//
// Example:
//
//	workflow := &Workflow{
//	    Name:        "vulnerability-scan",
//	    Type:        "dag",
//	    Description: "Automated vulnerability scanning workflow",
//	}
//
// Identifying Properties:
//   - name (required): Workflow name
//
// Relationships:
//   - None (root node)
type Workflow struct {
	// Name is the workflow name.
	// This is an identifying property and is required.
	Name string

	// Type is the workflow type.
	// Optional. Examples: "dag", "sequential", "parallel"
	Type string

	// Description describes the workflow's purpose.
	// Optional.
	Description string
}

// NodeType returns the canonical node type for Workflow nodes.
// Implements GraphNode interface.
func (w *Workflow) NodeType() string {
	return graphrag.NodeTypeWorkflow
}

// IdentifyingProperties returns the properties that uniquely identify this workflow.
// For Workflow nodes, only name is identifying.
// Implements GraphNode interface.
func (w *Workflow) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: w.Name,
	}
}

// Properties returns all properties to set on the Workflow node.
// Implements GraphNode interface.
func (w *Workflow) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: w.Name,
	}

	if w.Type != "" {
		props["type"] = w.Type
	}
	if w.Description != "" {
		props[graphrag.PropDescription] = w.Description
	}

	return props
}

// ParentRef returns nil because Workflow is a root node with no parent.
// Implements GraphNode interface.
func (w *Workflow) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Workflow is a root node.
// Implements GraphNode interface.
func (w *Workflow) RelationshipType() string {
	return ""
}

// Crew represents a group of collaborating agents.
// A Crew is identified by its name.
// Crew is a root-level node with no parent relationships.
//
// Example:
//
//	crew := &Crew{
//	    Name:        "security-team",
//	    Description: "Collaborative security testing crew",
//	    Size:        3,
//	}
//
// Identifying Properties:
//   - name (required): Crew name
//
// Relationships:
//   - None (root node)
type Crew struct {
	// Name is the crew name.
	// This is an identifying property and is required.
	Name string

	// Description describes the crew's purpose.
	// Optional.
	Description string

	// Size is the number of agents in the crew.
	// Optional.
	Size int
}

// NodeType returns the canonical node type for Crew nodes.
// Implements GraphNode interface.
func (c *Crew) NodeType() string {
	return graphrag.NodeTypeCrew
}

// IdentifyingProperties returns the properties that uniquely identify this crew.
// For Crew nodes, only name is identifying.
// Implements GraphNode interface.
func (c *Crew) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: c.Name,
	}
}

// Properties returns all properties to set on the Crew node.
// Implements GraphNode interface.
func (c *Crew) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: c.Name,
	}

	if c.Description != "" {
		props[graphrag.PropDescription] = c.Description
	}
	if c.Size > 0 {
		props["size"] = c.Size
	}

	return props
}

// ParentRef returns nil because Crew is a root node with no parent.
// Implements GraphNode interface.
func (c *Crew) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because Crew is a root node.
// Implements GraphNode interface.
func (c *Crew) RelationshipType() string {
	return ""
}

// AgentTask represents a task assigned to an agent.
// An AgentTask is identified by its unique ID.
// AgentTask is a root-level node with no parent relationships.
//
// Example:
//
//	task := &AgentTask{
//	    ID:          "task-123",
//	    Name:        "scan-network",
//	    Status:      "completed",
//	    Priority:    "high",
//	}
//
// Identifying Properties:
//   - id (required): Task identifier
//
// Relationships:
//   - None (root node)
type AgentTask struct {
	// ID is the task identifier.
	// This is an identifying property and is required.
	ID string

	// Name is the task name.
	// Optional.
	Name string

	// Status is the task status.
	// Optional. Examples: "pending", "running", "completed", "failed"
	Status string

	// Priority is the task priority.
	// Optional. Examples: "low", "medium", "high", "critical"
	Priority string
}

// NodeType returns the canonical node type for AgentTask nodes.
// Implements GraphNode interface.
func (t *AgentTask) NodeType() string {
	return graphrag.NodeTypeAgentTask
}

// IdentifyingProperties returns the properties that uniquely identify this task.
// For AgentTask nodes, only id is identifying.
// Implements GraphNode interface.
func (t *AgentTask) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": t.ID,
	}
}

// Properties returns all properties to set on the AgentTask node.
// Implements GraphNode interface.
func (t *AgentTask) Properties() map[string]any {
	props := map[string]any{
		"id": t.ID,
	}

	if t.Name != "" {
		props[graphrag.PropName] = t.Name
	}
	if t.Status != "" {
		props["status"] = t.Status
	}
	if t.Priority != "" {
		props["priority"] = t.Priority
	}

	return props
}

// ParentRef returns nil because AgentTask is a root node with no parent.
// Implements GraphNode interface.
func (t *AgentTask) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because AgentTask is a root node.
// Implements GraphNode interface.
func (t *AgentTask) RelationshipType() string {
	return ""
}

// AgentRole represents a role or persona for an agent.
// An AgentRole is identified by its name.
// AgentRole is a root-level node with no parent relationships.
//
// Example:
//
//	role := &AgentRole{
//	    Name:        "penetration-tester",
//	    Description: "Simulates adversarial attacks",
//	}
//
// Identifying Properties:
//   - name (required): Role name
//
// Relationships:
//   - None (root node)
type AgentRole struct {
	// Name is the role name.
	// This is an identifying property and is required.
	Name string

	// Description describes the role's purpose.
	// Optional.
	Description string
}

// NodeType returns the canonical node type for AgentRole nodes.
// Implements GraphNode interface.
func (r *AgentRole) NodeType() string {
	return graphrag.NodeTypeAgentRole
}

// IdentifyingProperties returns the properties that uniquely identify this role.
// For AgentRole nodes, only name is identifying.
// Implements GraphNode interface.
func (r *AgentRole) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropName: r.Name,
	}
}

// Properties returns all properties to set on the AgentRole node.
// Implements GraphNode interface.
func (r *AgentRole) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropName: r.Name,
	}

	if r.Description != "" {
		props[graphrag.PropDescription] = r.Description
	}

	return props
}

// ParentRef returns nil because AgentRole is a root node with no parent.
// Implements GraphNode interface.
func (r *AgentRole) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because AgentRole is a root node.
// Implements GraphNode interface.
func (r *AgentRole) RelationshipType() string {
	return ""
}

// ToolCall represents a tool invocation by an agent or LLM.
// A ToolCall is identified by its unique ID.
// ToolCall is a root-level node with no parent relationships.
//
// Example:
//
//	toolCall := &ToolCall{
//	    ID:       "tc-789",
//	    ToolName: "nmap",
//	    Status:   "completed",
//	    Duration: 1500,
//	}
//
// Identifying Properties:
//   - id (required): Tool call identifier
//
// Relationships:
//   - None (root node)
type ToolCall struct {
	// ID is the tool call identifier.
	// This is an identifying property and is required.
	ID string

	// ToolName is the name of the tool being called.
	// Optional.
	ToolName string

	// Status is the call status.
	// Optional. Examples: "pending", "running", "completed", "failed"
	Status string

	// Duration is the execution duration in milliseconds.
	// Optional.
	Duration int
}

// NodeType returns the canonical node type for ToolCall nodes.
// Implements GraphNode interface.
func (t *ToolCall) NodeType() string {
	return graphrag.NodeTypeToolCall
}

// IdentifyingProperties returns the properties that uniquely identify this tool call.
// For ToolCall nodes, only id is identifying.
// Implements GraphNode interface.
func (t *ToolCall) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": t.ID,
	}
}

// Properties returns all properties to set on the ToolCall node.
// Implements GraphNode interface.
func (t *ToolCall) Properties() map[string]any {
	props := map[string]any{
		"id": t.ID,
	}

	if t.ToolName != "" {
		props["tool_name"] = t.ToolName
	}
	if t.Status != "" {
		props["status"] = t.Status
	}
	if t.Duration > 0 {
		props["duration"] = t.Duration
	}

	return props
}

// ParentRef returns nil because ToolCall is a root node with no parent.
// Implements GraphNode interface.
func (t *ToolCall) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because ToolCall is a root node.
// Implements GraphNode interface.
func (t *ToolCall) RelationshipType() string {
	return ""
}

// ReasoningStep represents a step in an agent's reasoning process.
// A ReasoningStep is identified by its agent run ID and step number.
// ReasoningStep nodes are children of AgentRun nodes.
//
// Example:
//
//	step := &ReasoningStep{
//	    AgentRunID: "run-456",
//	    StepNumber: 3,
//	    Type:       "analysis",
//	    Content:    "Analyzing the target for vulnerabilities...",
//	}
//
// Identifying Properties:
//   - agent_run_id (required): Parent agent run identifier
//   - step_number (required): Step sequence number
//
// Relationships:
//   - Parent: AgentRun node (via HAS_REASONING_STEP relationship)
type ReasoningStep struct {
	// AgentRunID is the parent agent run identifier.
	// This is an identifying property and is required.
	AgentRunID string

	// StepNumber is the step sequence number.
	// This is an identifying property and is required.
	StepNumber int

	// Type is the reasoning step type.
	// Optional. Examples: "analysis", "planning", "execution", "reflection"
	Type string

	// Content is the reasoning content.
	// Optional.
	Content string
}

// NodeType returns the canonical node type for ReasoningStep nodes.
// Implements GraphNode interface.
func (r *ReasoningStep) NodeType() string {
	return graphrag.NodeTypeReasoningStep
}

// IdentifyingProperties returns the properties that uniquely identify this reasoning step.
// For ReasoningStep nodes, agent_run_id and step_number are both identifying.
// Implements GraphNode interface.
func (r *ReasoningStep) IdentifyingProperties() map[string]any {
	return map[string]any{
		graphrag.PropAgentRunID: r.AgentRunID,
		"step_number":           r.StepNumber,
	}
}

// Properties returns all properties to set on the ReasoningStep node.
// Implements GraphNode interface.
func (r *ReasoningStep) Properties() map[string]any {
	props := map[string]any{
		graphrag.PropAgentRunID: r.AgentRunID,
		"step_number":           r.StepNumber,
	}

	if r.Type != "" {
		props["type"] = r.Type
	}
	if r.Content != "" {
		props["content"] = r.Content
	}

	return props
}

// ParentRef returns a reference to the parent AgentRun node.
// Implements GraphNode interface.
func (r *ReasoningStep) ParentRef() *NodeRef {
	if r.AgentRunID == "" {
		return nil
	}
	return &NodeRef{
		NodeType: graphrag.NodeTypeAgentRun,
		Properties: map[string]any{
			"id": r.AgentRunID,
		},
	}
}

// RelationshipType returns the relationship type to the parent AgentRun node.
// Implements GraphNode interface.
func (r *ReasoningStep) RelationshipType() string {
	// Using empty string for now as RelTypeHasReasoningStep doesn't exist
	// This will be added in Phase 1 taxonomy generation
	return ""
}

// MemoryEntry represents an entry in agent memory.
// A MemoryEntry is identified by its unique ID.
// MemoryEntry is a root-level node with no parent relationships.
//
// Example:
//
//	entry := &MemoryEntry{
//	    ID:        "mem-123",
//	    Type:      "observation",
//	    Content:   "Target is running outdated software",
//	    Timestamp: 1704110400,
//	}
//
// Identifying Properties:
//   - id (required): Memory entry identifier
//
// Relationships:
//   - None (root node)
type MemoryEntry struct {
	// ID is the memory entry identifier.
	// This is an identifying property and is required.
	ID string

	// Type is the memory entry type.
	// Optional. Examples: "observation", "fact", "reflection"
	Type string

	// Content is the memory content.
	// Optional.
	Content string

	// Timestamp is the Unix timestamp when the entry was created.
	// Optional.
	Timestamp int64
}

// NodeType returns the canonical node type for MemoryEntry nodes.
// Implements GraphNode interface.
func (m *MemoryEntry) NodeType() string {
	return graphrag.NodeTypeMemoryEntry
}

// IdentifyingProperties returns the properties that uniquely identify this memory entry.
// For MemoryEntry nodes, only id is identifying.
// Implements GraphNode interface.
func (m *MemoryEntry) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": m.ID,
	}
}

// Properties returns all properties to set on the MemoryEntry node.
// Implements GraphNode interface.
func (m *MemoryEntry) Properties() map[string]any {
	props := map[string]any{
		"id": m.ID,
	}

	if m.Type != "" {
		props["type"] = m.Type
	}
	if m.Content != "" {
		props["content"] = m.Content
	}
	if m.Timestamp > 0 {
		props[graphrag.PropTimestamp] = m.Timestamp
	}

	return props
}

// ParentRef returns nil because MemoryEntry is a root node with no parent.
// Implements GraphNode interface.
func (m *MemoryEntry) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because MemoryEntry is a root node.
// Implements GraphNode interface.
func (m *MemoryEntry) RelationshipType() string {
	return ""
}

// AgentLoop represents an agent loop or iteration cycle.
// An AgentLoop is identified by its loop ID.
// AgentLoop is a root-level node with no parent relationships.
//
// Example:
//
//	loop := &AgentLoop{
//	    LoopID:     "loop-456",
//	    Iterations: 5,
//	    Status:     "completed",
//	}
//
// Identifying Properties:
//   - loop_id (required): Loop identifier
//
// Relationships:
//   - None (root node)
type AgentLoop struct {
	// LoopID is the loop identifier.
	// This is an identifying property and is required.
	LoopID string

	// Iterations is the number of iterations completed.
	// Optional.
	Iterations int

	// Status is the loop status.
	// Optional. Examples: "running", "completed", "failed", "max_iterations"
	Status string
}

// NodeType returns the canonical node type for AgentLoop nodes.
// Implements GraphNode interface.
func (l *AgentLoop) NodeType() string {
	return graphrag.NodeTypeAgentLoop
}

// IdentifyingProperties returns the properties that uniquely identify this agent loop.
// For AgentLoop nodes, only loop_id is identifying.
// Implements GraphNode interface.
func (l *AgentLoop) IdentifyingProperties() map[string]any {
	return map[string]any{
		"loop_id": l.LoopID,
	}
}

// Properties returns all properties to set on the AgentLoop node.
// Implements GraphNode interface.
func (l *AgentLoop) Properties() map[string]any {
	props := map[string]any{
		"loop_id": l.LoopID,
	}

	if l.Iterations > 0 {
		props["iterations"] = l.Iterations
	}
	if l.Status != "" {
		props["status"] = l.Status
	}

	return props
}

// ParentRef returns nil because AgentLoop is a root node with no parent.
// Implements GraphNode interface.
func (l *AgentLoop) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because AgentLoop is a root node.
// Implements GraphNode interface.
func (l *AgentLoop) RelationshipType() string {
	return ""
}

// PlanningStep represents a step in an agent's planning phase.
// A PlanningStep is identified by its step ID.
// PlanningStep is a root-level node with no parent relationships.
//
// Example:
//
//	step := &PlanningStep{
//	    StepID:      "step-789",
//	    Name:        "reconnaissance",
//	    Order:       1,
//	    Status:      "completed",
//	}
//
// Identifying Properties:
//   - step_id (required): Planning step identifier
//
// Relationships:
//   - None (root node)
type PlanningStep struct {
	// StepID is the planning step identifier.
	// This is an identifying property and is required.
	StepID string

	// Name is the step name.
	// Optional.
	Name string

	// Order is the step order in the plan.
	// Optional.
	Order int

	// Status is the step status.
	// Optional. Examples: "pending", "in_progress", "completed", "skipped"
	Status string
}

// NodeType returns the canonical node type for PlanningStep nodes.
// Implements GraphNode interface.
func (p *PlanningStep) NodeType() string {
	return graphrag.NodeTypePlanningStep
}

// IdentifyingProperties returns the properties that uniquely identify this planning step.
// For PlanningStep nodes, only step_id is identifying.
// Implements GraphNode interface.
func (p *PlanningStep) IdentifyingProperties() map[string]any {
	return map[string]any{
		"step_id": p.StepID,
	}
}

// Properties returns all properties to set on the PlanningStep node.
// Implements GraphNode interface.
func (p *PlanningStep) Properties() map[string]any {
	props := map[string]any{
		"step_id": p.StepID,
	}

	if p.Name != "" {
		props[graphrag.PropName] = p.Name
	}
	if p.Order > 0 {
		props["order"] = p.Order
	}
	if p.Status != "" {
		props["status"] = p.Status
	}

	return props
}

// ParentRef returns nil because PlanningStep is a root node with no parent.
// Implements GraphNode interface.
func (p *PlanningStep) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because PlanningStep is a root node.
// Implements GraphNode interface.
func (p *PlanningStep) RelationshipType() string {
	return ""
}

// AgentArtifact represents an artifact produced by an agent.
// An AgentArtifact is identified by its unique ID.
// AgentArtifact is a root-level node with no parent relationships.
//
// Example:
//
//	artifact := &AgentArtifact{
//	    ID:       "artifact-123",
//	    Type:     "report",
//	    Name:     "security-findings.pdf",
//	    Size:     1024000,
//	}
//
// Identifying Properties:
//   - id (required): Artifact identifier
//
// Relationships:
//   - None (root node)
type AgentArtifact struct {
	// ID is the artifact identifier.
	// This is an identifying property and is required.
	ID string

	// Type is the artifact type.
	// Optional. Examples: "report", "log", "screenshot", "data"
	Type string

	// Name is the artifact name.
	// Optional.
	Name string

	// Size is the artifact size in bytes.
	// Optional.
	Size int64
}

// NodeType returns the canonical node type for AgentArtifact nodes.
// Implements GraphNode interface.
func (a *AgentArtifact) NodeType() string {
	return graphrag.NodeTypeAgentArtifact
}

// IdentifyingProperties returns the properties that uniquely identify this artifact.
// For AgentArtifact nodes, only id is identifying.
// Implements GraphNode interface.
func (a *AgentArtifact) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": a.ID,
	}
}

// Properties returns all properties to set on the AgentArtifact node.
// Implements GraphNode interface.
func (a *AgentArtifact) Properties() map[string]any {
	props := map[string]any{
		"id": a.ID,
	}

	if a.Type != "" {
		props["type"] = a.Type
	}
	if a.Name != "" {
		props[graphrag.PropName] = a.Name
	}
	if a.Size > 0 {
		props["size"] = a.Size
	}

	return props
}

// ParentRef returns nil because AgentArtifact is a root node with no parent.
// Implements GraphNode interface.
func (a *AgentArtifact) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string because AgentArtifact is a root node.
// Implements GraphNode interface.
func (a *AgentArtifact) RelationshipType() string {
	return ""
}

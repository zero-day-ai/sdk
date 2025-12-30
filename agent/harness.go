package agent

import (
	"context"
	"log/slog"

	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/trace"
)

// Harness provides the runtime environment for agent execution.
// It provides access to LLMs, tools, plugins, findings, memory, and observability.
type Harness interface {
	// LLM Access Methods
	//
	// These methods provide access to LLM completions through named slots.
	// Slots are configured based on the agent's LLMSlots() requirements.

	// Complete performs a single LLM completion request.
	// The slot parameter identifies which LLM to use (e.g., "primary", "vision").
	// Options can be provided to customize temperature, max tokens, etc.
	Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error)

	// CompleteWithTools performs a completion with tool calling enabled.
	// The LLM can request to invoke tools and will receive tool results in subsequent turns.
	CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error)

	// Stream performs a streaming completion request.
	// Returns a channel that yields incremental chunks as they arrive.
	// The channel will be closed when the stream completes or an error occurs.
	Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error)

	// Tool Access Methods
	//
	// These methods provide access to external tools (e.g., HTTP client, shell, browser).

	// CallTool invokes a tool by name with the given input parameters.
	// Returns the tool's output as a map or an error if the tool fails.
	CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error)

	// ListTools returns descriptors for all available tools.
	// This can be used to discover available functionality.
	ListTools(ctx context.Context) ([]ToolDescriptor, error)

	// Plugin Access Methods
	//
	// These methods provide access to plugins (modular extensions to the framework).

	// QueryPlugin sends a query to a plugin and returns the result.
	// The method parameter identifies the plugin operation to invoke.
	// The params provide input data for the operation.
	QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error)

	// ListPlugins returns descriptors for all available plugins.
	ListPlugins(ctx context.Context) ([]PluginDescriptor, error)

	// Agent Delegation Methods
	//
	// These methods allow agents to delegate tasks to other agents.

	// DelegateToAgent assigns a task to another agent for execution.
	// This enables hierarchical agent architectures and specialization.
	DelegateToAgent(ctx context.Context, name string, task Task) (Result, error)

	// ListAgents returns descriptors for all available agents.
	ListAgents(ctx context.Context) ([]Descriptor, error)

	// Finding Management Methods
	//
	// These methods manage security findings discovered during testing.

	// SubmitFinding records a new security finding.
	// The finding will be validated, stored, and included in reports.
	SubmitFinding(ctx context.Context, f Finding) error

	// GetFindings retrieves findings matching the given filter criteria.
	GetFindings(ctx context.Context, filter FindingFilter) ([]Finding, error)

	// Memory Access
	//
	// Provides access to the agent's memory store for persistence.

	// Memory returns the memory store for this agent.
	// The agent can use this to store and retrieve state across task executions.
	Memory() MemoryStore

	// Context Access
	//
	// These methods provide access to mission and target context.

	// Mission returns the current mission context.
	// This includes mission parameters, constraints, and metadata.
	Mission() types.MissionContext

	// Target returns information about the target being tested.
	// This includes target URL, type, authentication, and metadata.
	Target() types.TargetInfo

	// Observability
	//
	// These methods provide access to logging, tracing, and metrics.

	// Tracer returns an OpenTelemetry tracer for distributed tracing.
	// Agents should create spans for major operations to enable observability.
	Tracer() trace.Tracer

	// Logger returns a structured logger for the agent.
	// All log output should use this logger for consistent formatting.
	Logger() *slog.Logger

	// TokenUsage returns the token usage tracker for this execution.
	// This tracks token consumption across all LLM slots.
	TokenUsage() llm.TokenTracker

	// GraphRAG Query Methods
	//
	// These methods provide access to the GraphRAG knowledge graph for semantic search,
	// pattern discovery, and relationship traversal.

	// QueryGraphRAG performs a semantic or hybrid query against the knowledge graph.
	// Returns nodes matching the query criteria, ranked by combined relevance score.
	QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error)

	// FindSimilarAttacks searches for attack patterns semantically similar to the given content.
	// Returns up to topK attack patterns ordered by similarity score.
	FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error)

	// FindSimilarFindings searches for findings semantically similar to the referenced finding.
	// Returns up to topK findings ordered by similarity score.
	FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error)

	// GetAttackChains discovers multi-step attack paths starting from a technique.
	// Returns attack chains up to maxDepth hops from the starting technique.
	GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error)

	// GetRelatedFindings retrieves findings connected via SIMILAR_TO or RELATED_TO relationships.
	// Returns all directly related findings for the given finding ID.
	GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error)

	// GraphRAG Storage Methods
	//
	// These methods enable agents to store arbitrary data in the knowledge graph
	// with custom node types, properties, and relationships.

	// StoreGraphNode stores an arbitrary node in the knowledge graph.
	// The node will be enriched with mission context and timestamps.
	// If Content is provided, embeddings will be automatically generated.
	// Returns the assigned node ID.
	StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error)

	// CreateGraphRelationship creates a relationship between two existing nodes.
	// Both nodes must exist; returns an error if either node is not found.
	CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error

	// StoreGraphBatch stores multiple nodes and relationships atomically.
	// Nodes are processed before relationships to ensure all targets exist.
	// Returns all assigned node IDs in order.
	StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error)

	// TraverseGraph walks the graph from a starting node following relationships.
	// Returns visited nodes with their paths and distances from the start.
	TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error)

	// GraphRAGHealth returns the health status of the GraphRAG subsystem.
	// Use this to check availability before performing GraphRAG operations.
	GraphRAGHealth(ctx context.Context) types.HealthStatus
}

// Descriptor provides metadata about an agent.
// This is used for agent discovery and selection.
type Descriptor struct {
	// Name is the unique identifier for the agent.
	Name string

	// Version is the semantic version of the agent.
	Version string

	// Description explains what the agent does.
	Description string

	// Capabilities lists the security testing capabilities the agent provides.
	Capabilities []Capability

	// TargetTypes lists the types of targets the agent can test.
	TargetTypes []types.TargetType

	// TechniqueTypes lists the attack techniques the agent employs.
	TechniqueTypes []types.TechniqueType
}

// ToolDescriptor provides metadata about a tool.
// This is a placeholder type until the tool package is implemented.
type ToolDescriptor struct {
	// Name is the unique identifier for the tool.
	Name string

	// Description explains what the tool does.
	Description string

	// Schema describes the tool's input and output structure.
	Schema map[string]any
}

// PluginDescriptor provides metadata about a plugin.
// This is a placeholder type until the plugin package is implemented.
type PluginDescriptor struct {
	// Name is the unique identifier for the plugin.
	Name string

	// Description explains what the plugin does.
	Description string

	// Version is the semantic version of the plugin.
	Version string

	// Methods lists the available plugin methods.
	Methods []string
}

// Finding represents a security vulnerability or issue discovered during testing.
// This is a placeholder type until the finding package is fully implemented.
type Finding interface {
	// ID returns the unique identifier for this finding.
	ID() string

	// Severity returns the severity level of this finding.
	Severity() string

	// Category returns the category of this finding.
	Category() string
}

// FindingFilter specifies criteria for filtering findings.
// This is a placeholder type until the finding package is fully implemented.
type FindingFilter struct {
	// MissionID filters by mission ID.
	MissionID string

	// AgentName filters by the agent that discovered the finding.
	AgentName string

	// MinSeverity filters by minimum severity level.
	MinSeverity string

	// Category filters by finding category.
	Category string

	// Limit limits the number of results returned.
	Limit int
}

// MemoryStore provides persistent storage for agent state.
// This is a placeholder type until the memory package is implemented.
type MemoryStore interface {
	// Get retrieves a value by key.
	Get(ctx context.Context, key string) (any, error)

	// Set stores a value with the given key.
	Set(ctx context.Context, key string, value any) error

	// Delete removes a value by key.
	Delete(ctx context.Context, key string) error

	// List returns all keys matching the given prefix.
	List(ctx context.Context, prefix string) ([]string, error)
}

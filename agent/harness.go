package agent

import (
	"context"
	"log/slog"
	"time"

	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/memory"
	"github.com/zero-day-ai/sdk/mission"
	"github.com/zero-day-ai/sdk/planning"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/tool"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/trace"
)

// ToolCall represents a single tool invocation request for parallel execution
type ToolCall struct {
	Name  string         // Tool name to invoke
	Input map[string]any // Tool input parameters
}

// ToolResult represents the result of a tool invocation
type ToolResult struct {
	Name   string         // Tool name that was invoked
	Output map[string]any // Tool output (nil if error)
	Error  error          // Error if tool failed (nil if success)
}

// MissionManager provides mission lifecycle operations for agents.
// This interface enables agents to autonomously create, run, monitor, and manage missions,
// supporting hierarchical agent architectures and autonomous operation patterns.
type MissionManager interface {
	// CreateMission creates a new mission from a workflow definition.
	// The workflow parameter should be a Gibson workflow.Workflow instance.
	// Returns mission metadata including the assigned mission ID.
	//
	// If the creating agent is part of a mission, the parent mission ID
	// will be automatically tracked for lineage.
	//
	// Example:
	//   workflow := BuildReconWorkflow()
	//   opts := &mission.CreateMissionOpts{
	//       Name: "Subdomain Enumeration",
	//       Constraints: &mission.MissionConstraints{
	//           MaxDuration: 30 * time.Minute,
	//           MaxTokens:   100000,
	//       },
	//   }
	//   info, err := harness.CreateMission(ctx, workflow, targetID, opts)
	CreateMission(ctx context.Context, workflow any, targetID string, opts *mission.CreateMissionOpts) (*mission.MissionInfo, error)

	// RunMission queues a mission for execution.
	// This method is non-blocking by default and returns immediately after queuing.
	// Use WaitForMission to block until the mission completes.
	//
	// Returns an error if:
	//   - The mission does not exist
	//   - The mission is already running
	//   - The mission is in a terminal state (completed, failed, cancelled)
	//
	// Example:
	//   err := harness.RunMission(ctx, missionID, nil)
	//   if err != nil {
	//       return fmt.Errorf("failed to start mission: %w", err)
	//   }
	RunMission(ctx context.Context, missionID string, opts *mission.RunMissionOpts) error

	// GetMissionStatus returns the current state of a mission.
	// Returns detailed status information including progress, findings count,
	// token usage, and error messages if applicable.
	//
	// Returns an error if the mission does not exist.
	//
	// Example:
	//   status, err := harness.GetMissionStatus(ctx, missionID)
	//   if err != nil {
	//       return err
	//   }
	//   log.Printf("Mission %s: %s (%.1f%% complete)", missionID, status.Status, status.Progress*100)
	GetMissionStatus(ctx context.Context, missionID string) (*mission.MissionStatusInfo, error)

	// WaitForMission blocks until a mission completes or the timeout expires.
	// Returns the final mission result including findings and output.
	//
	// The timeout parameter specifies how long to wait. Use 0 for no timeout.
	// Returns context.DeadlineExceeded if the timeout is reached before completion.
	//
	// Example:
	//   result, err := harness.WaitForMission(ctx, missionID, 10*time.Minute)
	//   if err != nil {
	//       return fmt.Errorf("mission wait failed: %w", err)
	//   }
	//   log.Printf("Mission completed with %d findings", len(result.Findings))
	WaitForMission(ctx context.Context, missionID string, timeout time.Duration) (*mission.MissionResult, error)

	// ListMissions returns missions matching the provided filter criteria.
	// Returns an empty slice if no missions match the filter.
	//
	// The filter supports:
	//   - Status filtering (pending, running, completed, etc.)
	//   - Target ID filtering
	//   - Parent mission ID filtering (for finding child missions)
	//   - Time range filtering
	//   - Tag filtering
	//   - Pagination (limit/offset)
	//
	// Example:
	//   filter := &mission.MissionFilter{
	//       Status:   &statusRunning,
	//       TargetID: &currentTargetID,
	//       Limit:    10,
	//   }
	//   missions, err := harness.ListMissions(ctx, filter)
	ListMissions(ctx context.Context, filter *mission.MissionFilter) ([]*mission.MissionInfo, error)

	// CancelMission requests cancellation of a running mission.
	// The mission will be gracefully interrupted and its status will transition to "cancelled".
	//
	// This operation is idempotent - calling it on an already cancelled or
	// completed mission returns success.
	//
	// Example:
	//   err := harness.CancelMission(ctx, missionID)
	//   if err != nil {
	//       log.Printf("Failed to cancel mission: %v", err)
	//   }
	CancelMission(ctx context.Context, missionID string) error

	// GetMissionResults returns the final results of a completed mission.
	// Results include findings, output data, and execution metrics.
	//
	// Returns an error if:
	//   - The mission does not exist
	//   - The mission has not completed yet (use WaitForMission to wait)
	//
	// Example:
	//   result, err := harness.GetMissionResults(ctx, missionID)
	//   if err != nil {
	//       return err
	//   }
	//   for _, finding := range result.Findings {
	//       log.Printf("Found %s: %s", finding.Severity, finding.Title)
	//   }
	GetMissionResults(ctx context.Context, missionID string) (*mission.MissionResult, error)
}

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

	// CompleteStructured performs a completion with provider-native structured output.
	// The response schema is derived from the provided struct type.
	// For Anthropic: uses tool_use pattern (schema becomes a tool definition)
	// For OpenAI: uses response_format with json_schema
	// The prompt should be natural language - no JSON instructions needed.
	// Returns a pointer to the populated struct or an error.
	// The schema parameter should be an instance of the struct type (e.g., MyStruct{}).
	CompleteStructured(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error)

	// CompleteStructuredAny is an alias for CompleteStructured for compatibility.
	CompleteStructuredAny(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error)

	// Tool Access Methods
	//
	// These methods provide access to external tools (e.g., HTTP client, shell, browser).

	// CallTool invokes a tool by name with the given input parameters.
	// Returns the tool's output as a map or an error if the tool fails.
	CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error)

	// ListTools returns descriptors for all available tools.
	// This can be used to discover available functionality.
	ListTools(ctx context.Context) ([]tool.Descriptor, error)

	// CallToolsParallel executes multiple tool calls concurrently.
	// Results are returned in the same order as the input calls.
	// Individual tool failures are captured in ToolResult.Error and do not
	// abort other calls. The context timeout applies to the entire batch.
	// maxConcurrency of 0 uses the default (10).
	CallToolsParallel(ctx context.Context, calls []ToolCall, maxConcurrency int) ([]ToolResult, error)

	// Plugin Access Methods
	//
	// These methods provide access to plugins (modular extensions to the framework).

	// QueryPlugin sends a query to a plugin and returns the result.
	// The method parameter identifies the plugin operation to invoke.
	// The params provide input data for the operation.
	QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error)

	// ListPlugins returns descriptors for all available plugins.
	ListPlugins(ctx context.Context) ([]plugin.Descriptor, error)

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
	SubmitFinding(ctx context.Context, f *finding.Finding) error

	// GetFindings retrieves findings matching the given filter criteria.
	GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)

	// Memory Access
	//
	// Provides access to the agent's memory store for persistence.

	// Memory returns the memory store for this agent.
	// The agent can use this to store and retrieve state across task executions.
	// The store provides access to three memory tiers: Working, Mission, and LongTerm.
	Memory() memory.Store

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
	// This method uses auto-routing to select semantic or structured query based on the Query fields.
	QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error)

	// QuerySemantic performs a semantic query using vector embeddings.
	// Use this when you want to search by meaning/content similarity.
	// The query MUST have Text or Embedding set.
	// Forces semantic search even if NodeTypes are specified.
	QuerySemantic(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error)

	// QueryStructured performs a structured query without semantic search.
	// Use this when you want to filter by node types, properties, or mission scope.
	// The query should have NodeTypes or other structural filters set.
	// Forces structured query even if Text/Embedding are present.
	QueryStructured(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error)

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
	//
	// DEPRECATED: Use StoreSemantic() or StoreStructured() for explicit intent.
	StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error)

	// StoreSemantic stores a node WITH semantic embeddings for semantic search.
	// Use this when the node contains text content that should be semantically searchable.
	// The Content field is required and will be embedded automatically.
	// Returns the assigned node ID.
	StoreSemantic(ctx context.Context, node graphrag.GraphNode) (string, error)

	// StoreStructured stores a node WITHOUT semantic embeddings.
	// Use this for pure metadata/structured data that doesn't need semantic search.
	// The Content field is optional and won't be embedded even if provided.
	// Returns the assigned node ID.
	StoreStructured(ctx context.Context, node graphrag.GraphNode) (string, error)

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

	// Planning Context Methods
	//
	// These methods provide access to planning context and allow agents to
	// report feedback to the planning system.

	// PlanContext returns the planning context for the current execution.
	// Returns nil if no planning context is available (non-planned execution).
	// Agents can use this to access mission goals, step budgets, and position
	// in the overall plan.
	PlanContext() planning.PlanningContext

	// ReportStepHints allows agents to provide feedback to the planning system.
	// Agents can report confidence levels, suggest next steps, recommend replanning,
	// and share key findings that should influence future planning decisions.
	// This method is a no-op if planning is not enabled.
	ReportStepHints(ctx context.Context, hints *planning.StepHints) error

	// Mission Execution Context Methods
	//
	// These methods provide access to extended mission context including
	// run history, resume status, and cross-run queries.

	// MissionExecutionContext returns the full execution context for the current run
	// including run number, resume status, and previous run info.
	// This provides more detail than Mission() for agents that need run awareness.
	MissionExecutionContext() types.MissionExecutionContext

	// GetMissionRunHistory returns all runs for this mission name.
	// Returns runs in chronological order (oldest first).
	// Returns empty slice if this is the first run.
	GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error)

	// GetPreviousRunFindings returns findings from the immediate prior run.
	// Returns empty slice if no prior run exists.
	// Use this to avoid re-discovering known vulnerabilities.
	GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)

	// GetAllRunFindings returns findings from all runs of this mission.
	// Useful for comprehensive analysis across the mission's history.
	GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error)

	// QueryGraphRAGScoped executes a GraphRAG query with explicit scope.
	// This is a convenience method that sets scope before calling QueryGraphRAG.
	// Scope can be: ScopeCurrentRun, ScopeSameMission, or ScopeAll.
	QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error)

	// Credential Access Methods
	//
	// These methods provide secure access to stored credentials.
	// Agents, plugins, and tools should ALWAYS use the credential store
	// for secrets rather than accepting raw API keys as parameters.

	// GetCredential retrieves a credential by name from the credential store.
	// The credential is decrypted and returned with its secret value.
	// Returns an error if the credential does not exist.
	//
	// Example:
	//   cred, err := harness.GetCredential(ctx, "hackerone-api")
	//   if err != nil {
	//       return fmt.Errorf("credential not found: %w", err)
	//   }
	//   apiKey := cred.Secret
	GetCredential(ctx context.Context, name string) (*types.Credential, error)

	// Mission Management Methods
	//
	// These methods provide mission lifecycle management for autonomous operation.
	// Agents can create, run, monitor, and manage missions programmatically,
	// enabling hierarchical agent architectures and autonomous campaigns.
	MissionManager
}

// StreamingHarness extends Harness with real-time event emission capabilities.
// It provides methods for emitting events during agent execution for live streaming
// to clients and receiving steering messages for interactive control.
//
// This interface is implemented by the serve package's streaming harness implementation.
// Agents that want streaming support should use the StreamingExecuteFunc type with
// SetStreamingExecuteFunc when building their configuration.
type StreamingHarness interface {
	// Embed the base Harness interface to inherit all standard capabilities
	Harness

	// EmitOutput emits a text output chunk to the client.
	// Use isReasoning=true for internal reasoning/thinking output,
	// or isReasoning=false for final user-facing output.
	EmitOutput(content string, isReasoning bool) error

	// EmitToolCall emits an event indicating a tool invocation is starting.
	// The callID should be a unique identifier for correlating with the result.
	EmitToolCall(toolName string, input map[string]any, callID string) error

	// EmitToolResult emits an event with the result of a tool invocation.
	// The callID should match the ID used in the corresponding EmitToolCall.
	EmitToolResult(output map[string]any, err error, callID string) error

	// EmitFinding emits an event when a security finding is discovered.
	// This allows clients to receive findings in real-time as they're found.
	EmitFinding(finding *finding.Finding) error

	// EmitStatus emits a status change event.
	// Use this to indicate progress through different phases of execution.
	EmitStatus(status string, message string) error

	// EmitError emits an error event without terminating execution.
	// Use this for non-fatal errors that the agent recovers from.
	EmitError(err error, context string) error

	// Steering returns a read-only channel for receiving steering messages.
	// Agents can select on this channel to respond to user guidance during execution.
	Steering() <-chan SteeringMessage

	// Mode returns the current execution mode (autonomous, semi-autonomous, manual).
	// Agents should adjust their behavior based on the current mode.
	Mode() ExecutionMode
}

// SteeringMessage represents a message from the client to steer agent behavior.
// This is a placeholder type that will be properly defined when implementing steering.
type SteeringMessage struct {
	// Content is the steering message content from the user.
	Content string

	// Priority indicates if this is a high-priority steering message.
	Priority bool
}

// ExecutionMode represents the agent's execution mode.
type ExecutionMode int

const (
	// ExecutionModeAutonomous means the agent operates independently.
	ExecutionModeAutonomous ExecutionMode = iota

	// ExecutionModeSemiAutonomous means the agent pauses for approval on critical actions.
	ExecutionModeSemiAutonomous

	// ExecutionModeManual means the agent waits for explicit user direction.
	ExecutionModeManual
)

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
	Capabilities []string

	// TargetTypes lists the types of targets the agent can test.
	TargetTypes []string

	// TechniqueTypes lists the attack techniques the agent employs.
	TechniqueTypes []string
}

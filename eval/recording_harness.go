// Package eval provides evaluation capabilities for the Gibson SDK.
// This file implements RecordingHarness, a transparent wrapper around agent.Harness
// that records all operations for trajectory-based evaluation.
package eval

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/graphragpb"
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/memory"
	"github.com/zero-day-ai/sdk/planning"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/tool"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// RecordingHarness wraps an agent.Harness and records all operations as trajectory steps.
// It implements the full agent.Harness interface by delegating to an inner harness
// while capturing inputs, outputs, timing, and errors for evaluation.
type RecordingHarness struct {
	inner      agent.Harness
	trajectory Trajectory
	mu         sync.Mutex
}

// NewRecordingHarness creates a new recording harness that wraps the given inner harness.
// All method calls will be delegated to the inner harness while recording trajectory steps.
func NewRecordingHarness(inner agent.Harness) *RecordingHarness {
	return &RecordingHarness{
		inner: inner,
		trajectory: Trajectory{
			Steps:     make([]TrajectoryStep, 0),
			StartTime: time.Now(),
		},
	}
}

// recordStep adds a trajectory step to the recording in a thread-safe manner.
func (r *RecordingHarness) recordStep(step TrajectoryStep) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trajectory.Steps = append(r.trajectory.Steps, step)
}

// Trajectory returns the recorded trajectory of operations.
// This returns a copy to prevent external modification.
func (r *RecordingHarness) Trajectory() Trajectory {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Update end time
	r.trajectory.EndTime = time.Now()

	// Return a copy with copied steps slice
	trajCopy := r.trajectory
	trajCopy.Steps = make([]TrajectoryStep, len(r.trajectory.Steps))
	copy(trajCopy.Steps, r.trajectory.Steps)

	return trajCopy
}

// Reset clears the recorded trajectory and starts a new recording session.
func (r *RecordingHarness) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.trajectory = Trajectory{
		Steps:     make([]TrajectoryStep, 0),
		StartTime: time.Now(),
	}
}

// Complete performs a single LLM completion request and records it.
func (r *RecordingHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	startTime := time.Now()

	// Delegate to inner harness
	resp, err := r.inner.Complete(ctx, slot, messages, opts...)

	// Record the step
	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "llm",
		Name:      slot,
		Input:     messages,
		Output:    resp,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return resp, err
}

// CompleteWithTools performs a completion with tool calling enabled and records it.
func (r *RecordingHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	startTime := time.Now()

	// Delegate to inner harness
	resp, err := r.inner.CompleteWithTools(ctx, slot, messages, tools)

	// Record the step with tools in input
	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type: "llm",
		Name: slot,
		Input: map[string]any{
			"messages": messages,
			"tools":    tools,
		},
		Output:    resp,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return resp, err
}

// Stream performs a streaming completion request and records it.
func (r *RecordingHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	startTime := time.Now()

	// Delegate to inner harness
	ch, err := r.inner.Stream(ctx, slot, messages)

	// Record the step (note: output will be incomplete since streaming is async)
	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "llm",
		Name:      slot,
		Input:     messages,
		Output:    "streaming",
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return ch, err
}

// CallToolProto invokes a tool with proto messages and records the invocation.
func (r *RecordingHarness) CallToolProto(ctx context.Context, name string, request proto.Message, response proto.Message) error {
	startTime := time.Now()

	// Delegate to inner harness
	err := r.inner.CallToolProto(ctx, name, request, response)

	// Record the step
	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "tool",
		Name:      name,
		Input:     request,
		Output:    response,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return err
}

// ListTools returns descriptors for all available tools.
func (r *RecordingHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	// No recording for list operations
	return r.inner.ListTools(ctx)
}

// QueryPlugin sends a query to a plugin and records it.
func (r *RecordingHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	startTime := time.Now()

	// Delegate to inner harness
	result, err := r.inner.QueryPlugin(ctx, name, method, params)

	// Record the step
	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type: "plugin",
		Name: name,
		Input: map[string]any{
			"method": method,
			"params": params,
		},
		Output:    result,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return result, err
}

// ListPlugins returns descriptors for all available plugins.
func (r *RecordingHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	// No recording for list operations
	return r.inner.ListPlugins(ctx)
}

// DelegateToAgent assigns a task to another agent and records the delegation.
func (r *RecordingHarness) DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error) {
	startTime := time.Now()

	// Delegate to inner harness
	result, err := r.inner.DelegateToAgent(ctx, name, task)

	// Record the step
	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "delegate",
		Name:      name,
		Input:     task,
		Output:    result,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return result, err
}

// ListAgents returns descriptors for all available agents.
func (r *RecordingHarness) ListAgents(ctx context.Context) ([]agent.Descriptor, error) {
	// No recording for list operations
	return r.inner.ListAgents(ctx)
}

// SubmitFinding records a new security finding and records the submission.
func (r *RecordingHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	startTime := time.Now()

	// Delegate to inner harness
	err := r.inner.SubmitFinding(ctx, f)

	// Record the step
	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "finding",
		Name:      "submit",
		Input:     f,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return err
}

// GetFindings retrieves findings matching the given filter criteria.
func (r *RecordingHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	// No recording for read operations
	return r.inner.GetFindings(ctx, filter)
}

// Memory returns the memory store for this agent.
// Memory operations are recorded when methods on the returned store are called.
func (r *RecordingHarness) Memory() memory.Store {
	// Wrap the memory store to record operations
	return &recordingMemoryStore{
		inner:    r.inner.Memory(),
		recorder: r,
	}
}

// recordingMemoryStore wraps a MemoryStore to record operations.
type recordingMemoryStore struct {
	inner    memory.Store
	recorder *RecordingHarness
}

// Working returns the working memory tier wrapped in a recording layer.
func (m *recordingMemoryStore) Working() memory.WorkingMemory {
	return &recordingWorkingMemory{
		inner:    m.inner.Working(),
		recorder: m.recorder,
	}
}

// Mission returns the mission memory tier wrapped in a recording layer.
func (m *recordingMemoryStore) Mission() memory.MissionMemory {
	return &recordingMissionMemory{
		inner:    m.inner.Mission(),
		recorder: m.recorder,
	}
}

// LongTerm returns the long-term memory tier wrapped in a recording layer.
func (m *recordingMemoryStore) LongTerm() memory.LongTermMemory {
	return &recordingLongTermMemory{
		inner:    m.inner.LongTerm(),
		recorder: m.recorder,
	}
}

// ============================================================================
// Working Memory Recording
// ============================================================================

type recordingWorkingMemory struct {
	inner    memory.WorkingMemory
	recorder *RecordingHarness
}

// Get retrieves a value by key and records the operation.
func (m *recordingWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	startTime := time.Now()

	value, err := m.inner.Get(ctx, key)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "memory.working",
		Name:      "get",
		Input:     key,
		Output:    value,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return value, err
}

// Set stores a value and records the operation.
func (m *recordingWorkingMemory) Set(ctx context.Context, key string, value any) error {
	startTime := time.Now()

	err := m.inner.Set(ctx, key, value)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type: "memory.working",
		Name: "set",
		Input: map[string]any{
			"key":   key,
			"value": value,
		},
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return err
}

// Delete removes a value and records the operation.
func (m *recordingWorkingMemory) Delete(ctx context.Context, key string) error {
	startTime := time.Now()

	err := m.inner.Delete(ctx, key)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "memory.working",
		Name:      "delete",
		Input:     key,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return err
}

// Clear removes all values and records the operation.
func (m *recordingWorkingMemory) Clear(ctx context.Context) error {
	startTime := time.Now()

	err := m.inner.Clear(ctx)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "memory.working",
		Name:      "clear",
		Input:     nil,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return err
}

// Keys returns all keys and records the operation.
func (m *recordingWorkingMemory) Keys(ctx context.Context) ([]string, error) {
	startTime := time.Now()

	keys, err := m.inner.Keys(ctx)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "memory.working",
		Name:      "keys",
		Input:     nil,
		Output:    keys,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return keys, err
}

// ============================================================================
// Mission Memory Recording (Stub - delegates without detailed recording)
// ============================================================================

type recordingMissionMemory struct {
	inner    memory.MissionMemory
	recorder *RecordingHarness
}

func (m *recordingMissionMemory) Get(ctx context.Context, key string) (*memory.Item, error) {
	startTime := time.Now()
	item, err := m.inner.Get(ctx, key)
	duration := time.Since(startTime)

	step := TrajectoryStep{
		Type:      "memory.mission",
		Name:      "get",
		Input:     key,
		Output:    item,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return item, err
}

func (m *recordingMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	startTime := time.Now()
	err := m.inner.Set(ctx, key, value, metadata)
	duration := time.Since(startTime)

	step := TrajectoryStep{
		Type: "memory.mission",
		Name: "set",
		Input: map[string]any{
			"key":      key,
			"value":    value,
			"metadata": metadata,
		},
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return err
}

func (m *recordingMissionMemory) Delete(ctx context.Context, key string) error {
	return m.inner.Delete(ctx, key)
}

func (m *recordingMissionMemory) Search(ctx context.Context, query string, limit int) ([]memory.Result, error) {
	startTime := time.Now()
	results, err := m.inner.Search(ctx, query, limit)
	duration := time.Since(startTime)

	step := TrajectoryStep{
		Type: "memory.mission",
		Name: "search",
		Input: map[string]any{
			"query": query,
			"limit": limit,
		},
		Output:    results,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return results, err
}

func (m *recordingMissionMemory) History(ctx context.Context, limit int) ([]memory.Item, error) {
	return m.inner.History(ctx, limit)
}

func (m *recordingMissionMemory) GetPreviousRunValue(ctx context.Context, key string) (any, error) {
	return m.inner.GetPreviousRunValue(ctx, key)
}

func (m *recordingMissionMemory) GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error) {
	return m.inner.GetValueHistory(ctx, key)
}

func (m *recordingMissionMemory) ContinuityMode() memory.MemoryContinuityMode {
	return m.inner.ContinuityMode()
}

// ============================================================================
// Long-Term Memory Recording (Stub - delegates without detailed recording)
// ============================================================================

type recordingLongTermMemory struct {
	inner    memory.LongTermMemory
	recorder *RecordingHarness
}

func (m *recordingLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	startTime := time.Now()
	id, err := m.inner.Store(ctx, content, metadata)
	duration := time.Since(startTime)

	step := TrajectoryStep{
		Type: "memory.longterm",
		Name: "store",
		Input: map[string]any{
			"content":  content,
			"metadata": metadata,
		},
		Output:    id,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return id, err
}

func (m *recordingLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]memory.Result, error) {
	startTime := time.Now()
	results, err := m.inner.Search(ctx, query, topK, filters)
	duration := time.Since(startTime)

	step := TrajectoryStep{
		Type: "memory.longterm",
		Name: "search",
		Input: map[string]any{
			"query":   query,
			"topK":    topK,
			"filters": filters,
		},
		Output:    results,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	m.recorder.recordStep(step)

	return results, err
}

func (m *recordingLongTermMemory) Delete(ctx context.Context, id string) error {
	return m.inner.Delete(ctx, id)
}

// Mission returns the current mission context.
func (r *RecordingHarness) Mission() types.MissionContext {
	// No recording for context access
	return r.inner.Mission()
}

// Target returns information about the target being tested.
func (r *RecordingHarness) Target() types.TargetInfo {
	// No recording for context access
	return r.inner.Target()
}

// Tracer returns an OpenTelemetry tracer for distributed tracing.
func (r *RecordingHarness) Tracer() trace.Tracer {
	// No recording for observability access
	return r.inner.Tracer()
}

// Logger returns a structured logger for the agent.
func (r *RecordingHarness) Logger() *slog.Logger {
	// No recording for observability access
	return r.inner.Logger()
}

// TokenUsage returns the token usage tracker for this execution.
func (r *RecordingHarness) TokenUsage() llm.TokenTracker {
	// No recording for observability access
	return r.inner.TokenUsage()
}

// QueryNodes performs a query against the knowledge graph using proto messages and records it.
func (r *RecordingHarness) QueryNodes(ctx context.Context, query *graphragpb.GraphQuery) ([]*graphragpb.QueryResult, error) {
	startTime := time.Now()

	results, err := r.inner.QueryNodes(ctx, query)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "graphrag",
		Name:      "query_nodes",
		Input:     query,
		Output:    results,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return results, err
}

// FindSimilarAttacks searches for attack patterns and records the operation.
func (r *RecordingHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	startTime := time.Now()

	patterns, err := r.inner.FindSimilarAttacks(ctx, content, topK)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type: "graphrag",
		Name: "find_similar_attacks",
		Input: map[string]any{
			"content": content,
			"topK":    topK,
		},
		Output:    patterns,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return patterns, err
}

// FindSimilarFindings searches for similar findings and records the operation.
func (r *RecordingHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	startTime := time.Now()

	findings, err := r.inner.FindSimilarFindings(ctx, findingID, topK)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type: "graphrag",
		Name: "find_similar_findings",
		Input: map[string]any{
			"findingID": findingID,
			"topK":      topK,
		},
		Output:    findings,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return findings, err
}

// GetAttackChains discovers multi-step attack paths and records the operation.
func (r *RecordingHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	startTime := time.Now()

	chains, err := r.inner.GetAttackChains(ctx, techniqueID, maxDepth)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type: "graphrag",
		Name: "get_attack_chains",
		Input: map[string]any{
			"techniqueID": techniqueID,
			"maxDepth":    maxDepth,
		},
		Output:    chains,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return chains, err
}

// GetRelatedFindings retrieves connected findings and records the operation.
func (r *RecordingHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	startTime := time.Now()

	findings, err := r.inner.GetRelatedFindings(ctx, findingID)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type: "graphrag",
		Name: "get_related_findings",
		Input: map[string]any{
			"findingID": findingID,
		},
		Output:    findings,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return findings, err
}

// StoreNode stores a graph node using proto messages and records the operation.
func (r *RecordingHarness) StoreNode(ctx context.Context, node *graphragpb.GraphNode) (string, error) {
	startTime := time.Now()

	nodeID, err := r.inner.StoreNode(ctx, node)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "graphrag",
		Name:      "store_node",
		Input:     node,
		Output:    nodeID,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return nodeID, err
}

// GraphRAGHealth returns the health status of the GraphRAG subsystem.
func (r *RecordingHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	// No recording for health checks
	return r.inner.GraphRAGHealth(ctx)
}

// ============================================================================
// Planning Operations
// ============================================================================

// PlanContext returns the planning context for the current execution.
func (r *RecordingHarness) PlanContext() planning.PlanningContext {
	// No recording for context access
	return r.inner.PlanContext()
}

// ReportStepHints allows agents to provide feedback to the planning system and records it.
func (r *RecordingHarness) ReportStepHints(ctx context.Context, hints *planning.StepHints) error {
	startTime := time.Now()

	err := r.inner.ReportStepHints(ctx, hints)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "planning",
		Name:      "report_step_hints",
		Input:     hints,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return err
}

// ============================================================================
// Mission Execution Context Operations
// ============================================================================

// MissionExecutionContext returns the full execution context for the current run.
func (r *RecordingHarness) MissionExecutionContext() types.MissionExecutionContext {
	// No recording for context access
	return r.inner.MissionExecutionContext()
}

// GetMissionRunHistory returns all runs for this mission name.
func (r *RecordingHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
	startTime := time.Now()

	history, err := r.inner.GetMissionRunHistory(ctx)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "mission",
		Name:      "get_run_history",
		Output:    history,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return history, err
}

// GetPreviousRunFindings returns findings from the immediate prior run.
func (r *RecordingHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	startTime := time.Now()

	findings, err := r.inner.GetPreviousRunFindings(ctx, filter)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "mission",
		Name:      "get_previous_run_findings",
		Input:     filter,
		Output:    findings,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return findings, err
}

// GetAllRunFindings returns findings from all runs of this mission.
func (r *RecordingHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	startTime := time.Now()

	findings, err := r.inner.GetAllRunFindings(ctx, filter)

	duration := time.Since(startTime)
	step := TrajectoryStep{
		Type:      "mission",
		Name:      "get_all_run_findings",
		Input:     filter,
		Output:    findings,
		StartTime: startTime,
		Duration:  duration,
	}
	if err != nil {
		step.Error = err.Error()
	}
	r.recordStep(step)

	return findings, err
}

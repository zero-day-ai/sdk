package serve

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/graphragpb"
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
	"go.opentelemetry.io/otel/trace/noop"
	protolib "google.golang.org/protobuf/proto"
)

// LocalHarness provides a minimal harness implementation for standalone agent execution.
// It implements the full agent.Harness interface but only provides in-memory memory storage.
// All LLM, tool, plugin, and finding operations return "not available" errors.
//
// This is used when agents run without an orchestrator connection, allowing them to
// execute basic operations without requiring full framework infrastructure.
type LocalHarness struct {
	// Memory storage
	memory memory.Store

	// Observability
	logger *slog.Logger
	tracer trace.Tracer

	// Context (minimal defaults)
	mission      types.MissionContext
	target       types.TargetInfo
	tokenTracker llm.TokenTracker
}

// newLocalHarness creates a new local harness with in-memory storage.
func newLocalHarness() *LocalHarness {
	return &LocalHarness{
		memory:       newInMemoryStore(),
		logger:       slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})),
		tracer:       noop.NewTracerProvider().Tracer("local-harness"),
		mission:      types.MissionContext{},
		target:       types.TargetInfo{},
		tokenTracker: &localTokenTracker{usage: make(map[string]llm.TokenUsage)},
	}
}

// ============================================================================
// Core Harness Methods
// ============================================================================

// Logger returns the structured logger for the agent.
func (h *LocalHarness) Logger() *slog.Logger {
	return h.logger
}

// Tracer returns the OpenTelemetry tracer for distributed tracing.
func (h *LocalHarness) Tracer() trace.Tracer {
	return h.tracer
}

// TokenUsage returns the token usage tracker for this execution.
func (h *LocalHarness) TokenUsage() llm.TokenTracker {
	return h.tokenTracker
}

// Mission returns the current mission context.
func (h *LocalHarness) Mission() types.MissionContext {
	return h.mission
}

// Target returns information about the target being tested.
func (h *LocalHarness) Target() types.TargetInfo {
	return h.target
}

// Memory returns the memory store for this agent.
func (h *LocalHarness) Memory() memory.Store {
	return h.memory
}

// ============================================================================
// LLM Operations (Not Available)
// ============================================================================

// Complete returns an error indicating LLM operations are not available.
func (h *LocalHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	h.logger.Warn("LLM Complete not available in standalone mode", "slot", slot)
	return nil, fmt.Errorf("LLM operations not available in standalone mode (no orchestrator connected)")
}

// CompleteWithTools returns an error indicating LLM operations are not available.
func (h *LocalHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	h.logger.Warn("LLM CompleteWithTools not available in standalone mode", "slot", slot)
	return nil, fmt.Errorf("LLM operations not available in standalone mode (no orchestrator connected)")
}

// Stream returns an error indicating LLM operations are not available.
func (h *LocalHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	h.logger.Warn("LLM Stream not available in standalone mode", "slot", slot)
	return nil, fmt.Errorf("LLM operations not available in standalone mode (no orchestrator connected)")
}

// CompleteStructured returns an error indicating LLM operations are not available.
func (h *LocalHarness) CompleteStructured(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	h.logger.Warn("LLM CompleteStructured not available in standalone mode", "slot", slot)
	return nil, fmt.Errorf("LLM operations not available in standalone mode (no orchestrator connected)")
}

// CompleteStructuredAny is an alias for CompleteStructured for compatibility.
func (h *LocalHarness) CompleteStructuredAny(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return h.CompleteStructured(ctx, slot, messages, schema)
}

// ============================================================================
// Tool Operations (Not Available)
// ============================================================================

// CallToolProto returns an error indicating proto tool operations are not available.
func (h *LocalHarness) CallToolProto(ctx context.Context, name string, request protolib.Message, response protolib.Message) error {
	h.logger.Warn("CallToolProto not available in standalone mode", "tool", name)
	return fmt.Errorf("proto tool operations not available in standalone mode (no orchestrator connected)")
}

// ListTools returns an empty list with a warning.
func (h *LocalHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	h.logger.Warn("ListTools not available in standalone mode")
	return []tool.Descriptor{}, nil
}

// ============================================================================
// Plugin Operations (Not Available)
// ============================================================================

// QueryPlugin returns an error indicating plugin operations are not available.
func (h *LocalHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	h.logger.Warn("QueryPlugin not available in standalone mode", "plugin", name, "method", method)
	return nil, fmt.Errorf("plugin operations not available in standalone mode (no orchestrator connected)")
}

// ListPlugins returns an empty list with a warning.
func (h *LocalHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	h.logger.Warn("ListPlugins not available in standalone mode")
	return []plugin.Descriptor{}, nil
}

// ============================================================================
// Agent Delegation Operations (Not Available)
// ============================================================================

// DelegateToAgent returns an error indicating delegation is not available.
func (h *LocalHarness) DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error) {
	h.logger.Warn("DelegateToAgent not available in standalone mode", "agent", name)
	return agent.Result{}, fmt.Errorf("agent delegation not available in standalone mode (no orchestrator connected)")
}

// ListAgents returns an empty list with a warning.
func (h *LocalHarness) ListAgents(ctx context.Context) ([]agent.Descriptor, error) {
	h.logger.Warn("ListAgents not available in standalone mode")
	return []agent.Descriptor{}, nil
}

// ============================================================================
// Finding Operations (Not Available)
// ============================================================================

// SubmitFinding logs the finding but cannot persist it.
func (h *LocalHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	h.logger.Warn("SubmitFinding not available in standalone mode - finding will not be persisted",
		"finding_id", f.ID,
		"severity", f.Severity,
		"category", f.Category)
	return fmt.Errorf("finding operations not available in standalone mode (no orchestrator connected)")
}

// GetFindings returns an empty list with a warning.
func (h *LocalHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	h.logger.Warn("GetFindings not available in standalone mode")
	return []*finding.Finding{}, nil
}

// ============================================================================
// GraphRAG Query Operations (Not Available)
// ============================================================================

// QueryGraphRAG returns an error indicating GraphRAG is not available.
// QueryNodes returns an error indicating proto GraphRAG is not available.
func (h *LocalHarness) QueryNodes(ctx context.Context, query *graphragpb.GraphQuery) ([]*graphragpb.QueryResult, error) {
	h.logger.Warn("QueryNodes not available in standalone mode")
	return nil, fmt.Errorf("proto GraphRAG not available in standalone mode (no orchestrator connected)")
}

func (h *LocalHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	h.logger.Warn("QueryGraphRAG not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// QuerySemantic returns an error indicating GraphRAG is not available.
func (h *LocalHarness) QuerySemantic(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	h.logger.Warn("QuerySemantic not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// QueryStructured returns an error indicating GraphRAG is not available.
func (h *LocalHarness) QueryStructured(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	h.logger.Warn("QueryStructured not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// FindSimilarAttacks returns an error indicating GraphRAG is not available.
func (h *LocalHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	h.logger.Warn("FindSimilarAttacks not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// FindSimilarFindings returns an error indicating GraphRAG is not available.
func (h *LocalHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	h.logger.Warn("FindSimilarFindings not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// GetAttackChains returns an error indicating GraphRAG is not available.
func (h *LocalHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	h.logger.Warn("GetAttackChains not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// GetRelatedFindings returns an error indicating GraphRAG is not available.
func (h *LocalHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	h.logger.Warn("GetRelatedFindings not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// ============================================================================
// GraphRAG Storage Operations (Not Available)
// ============================================================================

// StoreNode returns an error indicating proto GraphRAG is not available.
func (h *LocalHarness) StoreNode(ctx context.Context, node *graphragpb.GraphNode) (string, error) {
	h.logger.Warn("StoreNode not available in standalone mode")
	return "", fmt.Errorf("proto GraphRAG not available in standalone mode (no orchestrator connected)")
}

// StoreGraphNode returns an error indicating GraphRAG is not available.
func (h *LocalHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	h.logger.Warn("StoreGraphNode not available in standalone mode")
	return "", fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// StoreSemantic returns an error indicating GraphRAG is not available.
func (h *LocalHarness) StoreSemantic(ctx context.Context, node graphrag.GraphNode) (string, error) {
	h.logger.Warn("StoreSemantic not available in standalone mode")
	return "", fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// StoreStructured returns an error indicating GraphRAG is not available.
func (h *LocalHarness) StoreStructured(ctx context.Context, node graphrag.GraphNode) (string, error) {
	h.logger.Warn("StoreStructured not available in standalone mode")
	return "", fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// CreateGraphRelationship returns an error indicating GraphRAG is not available.
func (h *LocalHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	h.logger.Warn("CreateGraphRelationship not available in standalone mode")
	return fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// StoreGraphBatch returns an error indicating GraphRAG is not available.
func (h *LocalHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	h.logger.Warn("StoreGraphBatch not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// TraverseGraph returns an error indicating GraphRAG is not available.
func (h *LocalHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	h.logger.Warn("TraverseGraph not available in standalone mode")
	return nil, fmt.Errorf("GraphRAG not available in standalone mode (no orchestrator connected)")
}

// GraphRAGHealth returns unhealthy status indicating GraphRAG is not available.
func (h *LocalHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	return types.NewUnhealthyStatus("GraphRAG not available in standalone mode", nil)
}

// ============================================================================
// Planning Operations (Not Available)
// ============================================================================

// PlanContext returns nil indicating no planning context is available.
func (h *LocalHarness) PlanContext() planning.PlanningContext {
	return nil
}

// ReportStepHints is a no-op in standalone mode.
func (h *LocalHarness) ReportStepHints(ctx context.Context, hints *planning.StepHints) error {
	h.logger.Warn("ReportStepHints not available in standalone mode")
	return nil // No-op is acceptable per interface documentation
}

// ============================================================================
// Mission Execution Context Operations (Not Available)
// ============================================================================

// MissionExecutionContext returns an empty execution context.
func (h *LocalHarness) MissionExecutionContext() types.MissionExecutionContext {
	return types.MissionExecutionContext{}
}

// GetMissionRunHistory returns an empty slice in standalone mode.
func (h *LocalHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
	h.logger.Warn("GetMissionRunHistory not available in standalone mode")
	return []types.MissionRunSummary{}, nil
}

// GetPreviousRunFindings returns an empty slice in standalone mode.
func (h *LocalHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	h.logger.Warn("GetPreviousRunFindings not available in standalone mode")
	return []*finding.Finding{}, nil
}

// GetAllRunFindings returns an empty slice in standalone mode.
func (h *LocalHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	h.logger.Warn("GetAllRunFindings not available in standalone mode")
	return []*finding.Finding{}, nil
}

// ============================================================================
// In-Memory Storage Implementation
// ============================================================================

// inMemoryStore provides a simple in-memory implementation of memory.Store.
type inMemoryStore struct {
	mu      sync.RWMutex
	data    map[string]any
	working *inMemoryWorkingMemory
}

// newInMemoryStore creates a new in-memory store.
func newInMemoryStore() *inMemoryStore {
	store := &inMemoryStore{
		data: make(map[string]any),
	}
	store.working = &inMemoryWorkingMemory{store: store}
	return store
}

// Working returns the working memory tier (ephemeral, in-memory).
func (s *inMemoryStore) Working() memory.WorkingMemory {
	return s.working
}

// Mission returns the mission memory tier (persistent per-mission).
// For local harness, this is a stub that returns ErrNotImplemented.
func (s *inMemoryStore) Mission() memory.MissionMemory {
	return &stubMissionMemory{}
}

// LongTerm returns the long-term memory tier (vector-based).
// For local harness, this is a stub that returns ErrNotImplemented.
func (s *inMemoryStore) LongTerm() memory.LongTermMemory {
	return &stubLongTermMemory{}
}

// ============================================================================
// Working Memory Implementation
// ============================================================================

type inMemoryWorkingMemory struct {
	store *inMemoryStore
}

// Get retrieves a value by key.
func (w *inMemoryWorkingMemory) Get(ctx context.Context, key string) (any, error) {
	w.store.mu.RLock()
	defer w.store.mu.RUnlock()

	val, ok := w.store.data[key]
	if !ok {
		return nil, memory.ErrNotFound
	}
	return val, nil
}

// Set stores a value with the given key.
func (w *inMemoryWorkingMemory) Set(ctx context.Context, key string, value any) error {
	w.store.mu.Lock()
	defer w.store.mu.Unlock()

	w.store.data[key] = value
	return nil
}

// Delete removes a value by key.
func (w *inMemoryWorkingMemory) Delete(ctx context.Context, key string) error {
	w.store.mu.Lock()
	defer w.store.mu.Unlock()

	if _, ok := w.store.data[key]; !ok {
		return memory.ErrNotFound
	}
	delete(w.store.data, key)
	return nil
}

// Clear removes all values from working memory.
func (w *inMemoryWorkingMemory) Clear(ctx context.Context) error {
	w.store.mu.Lock()
	defer w.store.mu.Unlock()

	w.store.data = make(map[string]any)
	return nil
}

// Keys returns all keys currently in working memory.
func (w *inMemoryWorkingMemory) Keys(ctx context.Context) ([]string, error) {
	w.store.mu.RLock()
	defer w.store.mu.RUnlock()

	keys := make([]string, 0, len(w.store.data))
	for k := range w.store.data {
		keys = append(keys, k)
	}
	return keys, nil
}

// ============================================================================
// Stub Mission Memory Implementation
// ============================================================================

type stubMissionMemory struct{}

func (s *stubMissionMemory) Get(ctx context.Context, key string) (*memory.Item, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) Set(ctx context.Context, key string, value any, metadata map[string]any) error {
	return memory.ErrNotImplemented
}

func (s *stubMissionMemory) Delete(ctx context.Context, key string) error {
	return memory.ErrNotImplemented
}

func (s *stubMissionMemory) Search(ctx context.Context, query string, limit int) ([]memory.Result, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) History(ctx context.Context, limit int) ([]memory.Item, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) GetPreviousRunValue(ctx context.Context, key string) (any, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) GetValueHistory(ctx context.Context, key string) ([]memory.HistoricalValue, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubMissionMemory) ContinuityMode() memory.MemoryContinuityMode {
	return memory.MemoryIsolated
}

// ============================================================================
// Stub Long-Term Memory Implementation
// ============================================================================

type stubLongTermMemory struct{}

func (s *stubLongTermMemory) Store(ctx context.Context, content string, metadata map[string]any) (string, error) {
	return "", memory.ErrNotImplemented
}

func (s *stubLongTermMemory) Search(ctx context.Context, query string, topK int, filters map[string]any) ([]memory.Result, error) {
	return nil, memory.ErrNotImplemented
}

func (s *stubLongTermMemory) Delete(ctx context.Context, id string) error {
	return memory.ErrNotImplemented
}

// ============================================================================
// Local Token Tracker Implementation
// ============================================================================

// localTokenTracker provides a simple in-memory token usage tracker.
type localTokenTracker struct {
	mu    sync.RWMutex
	usage map[string]llm.TokenUsage
}

// Add records token usage for a slot.
func (t *localTokenTracker) Add(slot string, usage llm.TokenUsage) {
	t.mu.Lock()
	defer t.mu.Unlock()

	existing := t.usage[slot]
	existing.InputTokens += usage.InputTokens
	existing.OutputTokens += usage.OutputTokens
	existing.TotalTokens += usage.TotalTokens
	t.usage[slot] = existing
}

// BySlot returns the token usage for a specific slot.
func (t *localTokenTracker) BySlot(slot string) llm.TokenUsage {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.usage[slot]
}

// Total returns total token usage across all slots.
func (t *localTokenTracker) Total() llm.TokenUsage {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var total llm.TokenUsage
	for _, usage := range t.usage {
		total.InputTokens += usage.InputTokens
		total.OutputTokens += usage.OutputTokens
		total.TotalTokens += usage.TotalTokens
	}
	return total
}

// Reset clears all token usage data.
func (t *localTokenTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.usage = make(map[string]llm.TokenUsage)
}

// Slots returns a list of all tracked slot names.
func (t *localTokenTracker) Slots() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	slots := make([]string, 0, len(t.usage))
	for slot := range t.usage {
		slots = append(slots, slot)
	}
	return slots
}

// ============================================================================
// MissionManager Methods
// ============================================================================

// CreateMission creates a new mission from a workflow definition.
// This is a stub implementation that will be implemented in a future release.
func (h *LocalHarness) CreateMission(ctx context.Context, workflow any, targetID string, opts *mission.CreateMissionOpts) (*mission.MissionInfo, error) {
	return nil, fmt.Errorf("mission management not yet implemented in local harness")
}

// RunMission queues a mission for execution.
// This is a stub implementation that will be implemented in a future release.
func (h *LocalHarness) RunMission(ctx context.Context, missionID string, opts *mission.RunMissionOpts) error {
	return fmt.Errorf("mission management not yet implemented in local harness")
}

// GetMissionStatus returns the current state of a mission.
// This is a stub implementation that will be implemented in a future release.
func (h *LocalHarness) GetMissionStatus(ctx context.Context, missionID string) (*mission.MissionStatusInfo, error) {
	return nil, fmt.Errorf("mission management not yet implemented in local harness")
}

// WaitForMission blocks until a mission completes or the timeout expires.
// This is a stub implementation that will be implemented in a future release.
func (h *LocalHarness) WaitForMission(ctx context.Context, missionID string, timeout time.Duration) (*mission.MissionResult, error) {
	return nil, fmt.Errorf("mission management not yet implemented in local harness")
}

// ListMissions returns missions matching the provided filter criteria.
// This is a stub implementation that will be implemented in a future release.
func (h *LocalHarness) ListMissions(ctx context.Context, filter *mission.MissionFilter) ([]*mission.MissionInfo, error) {
	return nil, fmt.Errorf("mission management not yet implemented in local harness")
}

// CancelMission requests cancellation of a running mission.
// This is a stub implementation that will be implemented in a future release.
func (h *LocalHarness) CancelMission(ctx context.Context, missionID string) error {
	return fmt.Errorf("mission management not yet implemented in local harness")
}

// GetMissionResults returns the final results of a completed mission.
// This is a stub implementation that will be implemented in a future release.
func (h *LocalHarness) GetMissionResults(ctx context.Context, missionID string) (*mission.MissionResult, error) {
	return nil, fmt.Errorf("mission management not yet implemented in local harness")
}

// ============================================================================
// Credential Operations (Not Available)
// ============================================================================

// GetCredential returns an error indicating credentials are not available.
// In standalone mode, there is no credential store available.
func (h *LocalHarness) GetCredential(ctx context.Context, name string) (*types.Credential, error) {
	h.logger.Warn("GetCredential not available in standalone mode", "name", name)
	return nil, fmt.Errorf("credential store not available in standalone mode (no orchestrator connected)")
}

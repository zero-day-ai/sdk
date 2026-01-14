package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/finding"
	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/memory"
	"github.com/zero-day-ai/sdk/mission"
	"github.com/zero-day-ai/sdk/planning"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/tool"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// CallbackHarness implements agent.Harness by forwarding all operations
// to the orchestrator via gRPC callbacks. This allows agents running in
// standalone mode to access the full harness functionality.
type CallbackHarness struct {
	// Core dependencies
	client       *CallbackClient
	memory       memory.Store
	tokenTracker llm.TokenTracker

	// Context
	logger         *slog.Logger
	tracer         trace.Tracer
	mission        types.MissionContext
	target         types.TargetInfo
	planContext    planning.PlanningContext
	missionExecCtx types.MissionExecutionContext

	// Caching for list operations
	cacheMu      sync.RWMutex
	toolsCache   []tool.Descriptor
	pluginsCache []plugin.Descriptor
	agentsCache  []agent.Descriptor
}

// NewCallbackHarness creates a new callback-based harness.
func NewCallbackHarness(
	client *CallbackClient,
	logger *slog.Logger,
	tracer trace.Tracer,
	mission types.MissionContext,
	target types.TargetInfo,
) *CallbackHarness {
	return &CallbackHarness{
		client:       client,
		memory:       NewCallbackMemoryStore(client, tracer),
		tokenTracker: NewCallbackTokenTracker(),
		logger:       logger,
		tracer:       tracer,
		mission:      mission,
		target:       target,
		planContext:  nil, // Set via SetPlanContext if planning is enabled
	}
}

// SetPlanContext sets the planning context for this harness.
// This should be called by the orchestrator when executing a planned mission.
func (h *CallbackHarness) SetPlanContext(ctx planning.PlanningContext) {
	h.planContext = ctx
}

// SetMissionExecutionContext sets the mission execution context for this harness.
// This should be called by the orchestrator when executing a mission with run history.
func (h *CallbackHarness) SetMissionExecutionContext(ctx types.MissionExecutionContext) {
	h.missionExecCtx = ctx
}

// ============================================================================
// Core Harness Methods
// ============================================================================

// Logger returns the structured logger for the agent.
func (h *CallbackHarness) Logger() *slog.Logger {
	return h.logger
}

// Tracer returns the OpenTelemetry tracer for distributed tracing.
func (h *CallbackHarness) Tracer() trace.Tracer {
	return h.tracer
}

// TokenUsage returns the token usage tracker for this execution.
func (h *CallbackHarness) TokenUsage() llm.TokenTracker {
	return h.tokenTracker
}

// Mission returns the current mission context.
func (h *CallbackHarness) Mission() types.MissionContext {
	return h.mission
}

// Target returns information about the target being tested.
func (h *CallbackHarness) Target() types.TargetInfo {
	return h.target
}

// Memory returns the memory store for this agent.
func (h *CallbackHarness) Memory() memory.Store {
	return h.memory
}

// ============================================================================
// LLM Operations
// ============================================================================

// Complete performs a single LLM completion request via the orchestrator.
func (h *CallbackHarness) Complete(ctx context.Context, slot string, messages []llm.Message, opts ...llm.CompletionOption) (*llm.CompletionResponse, error) {
	// Start span for LLM completion
	ctx, span := h.tracer.Start(ctx, "gen_ai.chat",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gen_ai.system", "anthropic"),
			attribute.String("gen_ai.request.model", slot),
			attribute.Int("gen_ai.request.message_count", len(messages)),
		),
	)
	defer span.End()

	// Add prompt attribute for observability
	span.SetAttributes(attribute.String("gen_ai.prompt", formatMessagesForPrompt(messages)))

	// Build completion request with options
	req := llm.NewCompletionRequest(messages, opts...)

	// Convert to proto request
	protoReq := &proto.LLMCompleteRequest{
		Slot:     slot,
		Messages: h.messagesToProto(messages),
	}

	// Apply options
	if req.Temperature != nil {
		temp := *req.Temperature
		protoReq.Temperature = &temp
		span.SetAttributes(attribute.Float64("gen_ai.request.temperature", float64(temp)))
	}
	if req.MaxTokens != nil {
		maxTokens := int32(*req.MaxTokens)
		protoReq.MaxTokens = &maxTokens
		span.SetAttributes(attribute.Int("gen_ai.request.max_tokens", int(*req.MaxTokens)))
	}
	if req.TopP != nil {
		topP := *req.TopP
		protoReq.TopP = &topP
		span.SetAttributes(attribute.Float64("gen_ai.request.top_p", float64(topP)))
	}
	protoReq.Stop = req.Stop

	// Call orchestrator
	resp, err := h.client.LLMComplete(ctx, protoReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("LLM complete callback failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("LLM complete error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	// Convert response
	result := &llm.CompletionResponse{
		Content:      resp.Content,
		ToolCalls:    h.toolCallsFromProto(resp.ToolCalls),
		FinishReason: resp.FinishReason,
		Usage: llm.TokenUsage{
			InputTokens:  int(resp.Usage.InputTokens),
			OutputTokens: int(resp.Usage.OutputTokens),
			TotalTokens:  int(resp.Usage.TotalTokens),
		},
	}

	// Record token usage and response in span
	span.SetAttributes(
		attribute.Int("gen_ai.usage.input_tokens", result.Usage.InputTokens),
		attribute.Int("gen_ai.usage.output_tokens", result.Usage.OutputTokens),
		attribute.String("gen_ai.response.finish_reason", result.FinishReason),
		attribute.String("gen_ai.completion", result.Content),
		attribute.String("gen_ai.response.model", slot),
	)

	// Track token usage
	h.tokenTracker.Add(slot, result.Usage)

	return result, nil
}

// CompleteWithTools performs a completion with tool calling enabled.
func (h *CallbackHarness) CompleteWithTools(ctx context.Context, slot string, messages []llm.Message, tools []llm.ToolDef) (*llm.CompletionResponse, error) {
	// Start span for LLM completion with tools
	ctx, span := h.tracer.Start(ctx, "gen_ai.chat",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gen_ai.system", "anthropic"),
			attribute.String("gen_ai.request.model", slot),
			attribute.Int("gen_ai.request.message_count", len(messages)),
			attribute.Int("gen_ai.request.tool_count", len(tools)),
		),
	)
	defer span.End()

	// Add prompt attribute for observability
	span.SetAttributes(attribute.String("gen_ai.prompt", formatMessagesForPrompt(messages)))

	protoReq := &proto.LLMCompleteWithToolsRequest{
		Slot:     slot,
		Messages: h.messagesToProto(messages),
		Tools:    h.toolDefsToProto(tools),
	}

	resp, err := h.client.LLMCompleteWithTools(ctx, protoReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("LLM complete with tools callback failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("LLM complete with tools error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	result := &llm.CompletionResponse{
		Content:      resp.Content,
		ToolCalls:    h.toolCallsFromProto(resp.ToolCalls),
		FinishReason: resp.FinishReason,
		Usage: llm.TokenUsage{
			InputTokens:  int(resp.Usage.InputTokens),
			OutputTokens: int(resp.Usage.OutputTokens),
			TotalTokens:  int(resp.Usage.TotalTokens),
		},
	}

	// Record token usage and response in span
	span.SetAttributes(
		attribute.Int("gen_ai.usage.input_tokens", result.Usage.InputTokens),
		attribute.Int("gen_ai.usage.output_tokens", result.Usage.OutputTokens),
		attribute.String("gen_ai.response.finish_reason", result.FinishReason),
		attribute.Int("gen_ai.response.tool_call_count", len(result.ToolCalls)),
		attribute.String("gen_ai.completion", result.Content),
		attribute.String("gen_ai.response.model", slot),
	)

	// Track token usage
	h.tokenTracker.Add(slot, result.Usage)

	return result, nil
}

// CompleteStructured performs a completion with provider-native structured output.
// This forwards the request to the orchestrator which handles schema conversion
// and provider-specific structured output mechanisms.
//
// The schemaType parameter should be a Go struct (or pointer to struct) that
// defines the expected response structure. The method generates a JSON schema
// from the type and sends it to the daemon for LLM completion.
func (h *CallbackHarness) CompleteStructured(ctx context.Context, slot string, messages []llm.Message, schemaType any) (any, error) {
	// Generate JSON schema from the Go type
	// This converts the struct definition to a proper JSON schema that the LLM can use
	jsonSchema := schema.FromType(schemaType)

	// Serialize the schema to JSON for transmission
	schemaJSON, err := json.Marshal(jsonSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	protoReq := &proto.LLMCompleteStructuredRequest{
		Slot:       slot,
		Messages:   h.messagesToProto(messages),
		SchemaJson: string(schemaJSON),
	}

	resp, err := h.client.LLMCompleteStructured(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("LLM complete structured callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("LLM complete structured error: %s", resp.Error.Message)
	}

	// Deserialize the result
	var result any
	if err := json.Unmarshal([]byte(resp.ResultJson), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal structured result: %w", err)
	}

	// Track token usage if available
	if resp.Usage != nil {
		usage := llm.TokenUsage{
			InputTokens:  int(resp.Usage.InputTokens),
			OutputTokens: int(resp.Usage.OutputTokens),
			TotalTokens:  int(resp.Usage.TotalTokens),
		}
		h.tokenTracker.Add(slot, usage)
	}

	return result, nil
}

// CompleteStructuredAny is an alias for CompleteStructured for compatibility.
func (h *CallbackHarness) CompleteStructuredAny(ctx context.Context, slot string, messages []llm.Message, schema any) (any, error) {
	return h.CompleteStructured(ctx, slot, messages, schema)
}

// Stream performs a streaming completion request.
func (h *CallbackHarness) Stream(ctx context.Context, slot string, messages []llm.Message) (<-chan llm.StreamChunk, error) {
	// Start span for streaming LLM completion
	ctx, span := h.tracer.Start(ctx, "gen_ai.chat.stream",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gen_ai.system", "anthropic"),
			attribute.String("gen_ai.request.model", slot),
			attribute.Int("gen_ai.request.message_count", len(messages)),
		),
	)

	protoReq := &proto.LLMStreamRequest{
		Slot:     slot,
		Messages: h.messagesToProto(messages),
	}

	stream, err := h.client.LLMStream(ctx, protoReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		span.End()
		return nil, fmt.Errorf("LLM stream callback failed: %w", err)
	}

	// Create output channel
	chunkChan := make(chan llm.StreamChunk, 10)

	// Spawn goroutine to receive stream chunks
	go func() {
		defer close(chunkChan)
		defer span.End()

		for {
			protoChunk, err := stream.Recv()
			if err != nil {
				// Stream ended (could be EOF or error)
				if err.Error() != "EOF" {
					span.RecordError(err)
					span.SetStatus(codes.Error, err.Error())
				}
				return
			}

			if protoChunk.Error != nil {
				h.logger.Error("stream chunk error", "error", protoChunk.Error.Message)
				err := fmt.Errorf("stream chunk error: %s", protoChunk.Error.Message)
				span.RecordError(err)
				span.SetStatus(codes.Error, protoChunk.Error.Message)
				return
			}

			chunk := llm.StreamChunk{
				Delta:        protoChunk.Delta,
				ToolCalls:    h.toolCallsFromProto(protoChunk.ToolCalls),
				FinishReason: protoChunk.FinishReason,
			}

			if protoChunk.Usage != nil {
				usage := llm.TokenUsage{
					InputTokens:  int(protoChunk.Usage.InputTokens),
					OutputTokens: int(protoChunk.Usage.OutputTokens),
					TotalTokens:  int(protoChunk.Usage.TotalTokens),
				}
				chunk.Usage = &usage

				// Track token usage on final chunk
				if chunk.FinishReason != "" {
					h.tokenTracker.Add(slot, usage)
					// Record final token usage in span
					span.SetAttributes(
						attribute.Int("gen_ai.usage.input_tokens", usage.InputTokens),
						attribute.Int("gen_ai.usage.output_tokens", usage.OutputTokens),
						attribute.String("gen_ai.response.finish_reason", chunk.FinishReason),
					)
				}
			}

			select {
			case chunkChan <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunkChan, nil
}

// ============================================================================
// Tool Operations
// ============================================================================

// CallTool invokes a tool by name with the given input parameters.
func (h *CallbackHarness) CallTool(ctx context.Context, name string, input map[string]any) (map[string]any, error) {
	// Start span for tool call
	ctx, span := h.tracer.Start(ctx, "gen_ai.tool",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.tool.name", name),
		),
	)
	defer span.End()

	// Serialize input
	inputJSON, err := json.Marshal(input)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to marshal tool input: %w", err)
	}

	protoReq := &proto.CallToolRequest{
		Name:      name,
		InputJson: string(inputJSON),
	}

	resp, err := h.client.CallTool(ctx, protoReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("call tool callback failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("call tool error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	// Deserialize output
	var output map[string]any
	if err := json.Unmarshal([]byte(resp.OutputJson), &output); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to unmarshal tool output: %w", err)
	}

	return output, nil
}

// CallToolsParallel executes multiple tool calls concurrently.
func (h *CallbackHarness) CallToolsParallel(ctx context.Context, calls []agent.ToolCall, maxConcurrency int) ([]agent.ToolResult, error) {
	if len(calls) == 0 {
		return []agent.ToolResult{}, nil
	}

	// Default concurrency
	if maxConcurrency <= 0 {
		maxConcurrency = 10
	}

	// Create results slice (same length as calls, preserves order)
	results := make([]agent.ToolResult, len(calls))

	// Semaphore for concurrency limiting
	sem := make(chan struct{}, maxConcurrency)

	// WaitGroup for completion tracking
	var wg sync.WaitGroup

	// Execute calls in parallel
	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c agent.ToolCall) {
			defer wg.Done()

			// Acquire semaphore slot
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = agent.ToolResult{
					Name:  c.Name,
					Error: ctx.Err(),
				}
				return
			}

			// Execute tool call using existing CallTool
			output, err := h.CallTool(ctx, c.Name, c.Input)
			results[idx] = agent.ToolResult{
				Name:   c.Name,
				Output: output,
				Error:  err,
			}
		}(i, call)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	return results, nil
}

// ListTools returns descriptors for all available tools.
// Results are cached per task execution.
func (h *CallbackHarness) ListTools(ctx context.Context) ([]tool.Descriptor, error) {
	// Check cache first
	h.cacheMu.RLock()
	if h.toolsCache != nil {
		defer h.cacheMu.RUnlock()
		return h.toolsCache, nil
	}
	h.cacheMu.RUnlock()

	// Fetch from orchestrator
	protoReq := &proto.ListToolsRequest{}
	resp, err := h.client.ListTools(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("list tools callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list tools error: %s", resp.Error.Message)
	}

	// Convert to tool.Descriptor with full schema support including taxonomy
	tools := make([]tool.Descriptor, len(resp.Tools))
	for i, protoTool := range resp.Tools {
		tools[i] = tool.Descriptor{
			Name:        protoTool.Name,
			Description: protoTool.Description,
			Version:     "unknown", // Proto doesn't include version yet
		}

		// Prefer structured schemas (with taxonomy) over JSON strings
		if protoTool.InputSchema != nil {
			tools[i].InputSchema = protoToSchema(protoTool.InputSchema)
		}
		if protoTool.OutputSchema != nil {
			tools[i].OutputSchema = protoToSchema(protoTool.OutputSchema)
		}
	}

	// Cache results
	h.cacheMu.Lock()
	h.toolsCache = tools
	h.cacheMu.Unlock()

	return tools, nil
}

// ============================================================================
// Plugin Operations
// ============================================================================

// QueryPlugin sends a query to a plugin and returns the result.
func (h *CallbackHarness) QueryPlugin(ctx context.Context, name string, method string, params map[string]any) (any, error) {
	// Serialize params
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin params: %w", err)
	}

	protoReq := &proto.QueryPluginRequest{
		Name:       name,
		Method:     method,
		ParamsJson: string(paramsJSON),
	}

	resp, err := h.client.QueryPlugin(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("query plugin callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("query plugin error: %s", resp.Error.Message)
	}

	// Deserialize result
	var result any
	if err := json.Unmarshal([]byte(resp.ResultJson), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin result: %w", err)
	}

	return result, nil
}

// ListPlugins returns descriptors for all available plugins.
// Results are cached per task execution.
func (h *CallbackHarness) ListPlugins(ctx context.Context) ([]plugin.Descriptor, error) {
	// Check cache first
	h.cacheMu.RLock()
	if h.pluginsCache != nil {
		defer h.cacheMu.RUnlock()
		return h.pluginsCache, nil
	}
	h.cacheMu.RUnlock()

	// Fetch from orchestrator
	protoReq := &proto.ListPluginsRequest{}
	resp, err := h.client.ListPlugins(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("list plugins callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list plugins error: %s", resp.Error.Message)
	}

	// Convert to plugin.Descriptor
	// Note: This conversion is simplified - full MethodDescriptors require proto update
	plugins := make([]plugin.Descriptor, len(resp.Plugins))
	for i, protoPlugin := range resp.Plugins {
		// Convert method names to MethodDescriptors (minimal for now)
		methods := make([]plugin.MethodDescriptor, len(protoPlugin.Methods))
		for j, methodName := range protoPlugin.Methods {
			methods[j] = plugin.MethodDescriptor{
				Name: methodName,
				// TODO: Add Description, InputSchema, OutputSchema once proto is updated
			}
		}

		plugins[i] = plugin.Descriptor{
			Name:        protoPlugin.Name,
			Version:     protoPlugin.Version,
			Description: protoPlugin.Description,
			Methods:     methods,
		}
	}

	// Cache results
	h.cacheMu.Lock()
	h.pluginsCache = plugins
	h.cacheMu.Unlock()

	return plugins, nil
}

// ============================================================================
// Agent Delegation Operations
// ============================================================================

// DelegateToAgent assigns a task to another agent for execution.
func (h *CallbackHarness) DelegateToAgent(ctx context.Context, name string, task agent.Task) (agent.Result, error) {
	// Serialize task
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return agent.Result{}, fmt.Errorf("failed to marshal task: %w", err)
	}

	protoReq := &proto.DelegateToAgentRequest{
		Name:     name,
		TaskJson: string(taskJSON),
	}

	resp, err := h.client.DelegateToAgent(ctx, protoReq)
	if err != nil {
		return agent.Result{}, fmt.Errorf("delegate to agent callback failed: %w", err)
	}

	if resp.Error != nil {
		return agent.Result{}, fmt.Errorf("delegate to agent error: %s", resp.Error.Message)
	}

	// Deserialize result
	var result agent.Result
	if err := json.Unmarshal([]byte(resp.ResultJson), &result); err != nil {
		return agent.Result{}, fmt.Errorf("failed to unmarshal agent result: %w", err)
	}

	return result, nil
}

// ListAgents returns descriptors for all available agents.
// Results are cached per task execution.
func (h *CallbackHarness) ListAgents(ctx context.Context) ([]agent.Descriptor, error) {
	// Check cache first
	h.cacheMu.RLock()
	if h.agentsCache != nil {
		defer h.cacheMu.RUnlock()
		return h.agentsCache, nil
	}
	h.cacheMu.RUnlock()

	// Fetch from orchestrator
	protoReq := &proto.ListAgentsRequest{}
	resp, err := h.client.ListAgents(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("list agents callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("list agents error: %s", resp.Error.Message)
	}

	// Convert to agent.Descriptor
	agents := make([]agent.Descriptor, len(resp.Agents))
	for i, protoAgent := range resp.Agents {
		// Proto returns strings, which match the new agent interface
		agents[i] = agent.Descriptor{
			Name:           protoAgent.Name,
			Version:        protoAgent.Version,
			Description:    protoAgent.Description,
			Capabilities:   protoAgent.Capabilities,
			TargetTypes:    protoAgent.TargetTypes,
			TechniqueTypes: protoAgent.TechniqueTypes,
		}
	}

	// Cache results
	h.cacheMu.Lock()
	h.agentsCache = agents
	h.cacheMu.Unlock()

	return agents, nil
}

// ============================================================================
// Finding Operations
// ============================================================================

// SubmitFinding records a new security finding.
func (h *CallbackHarness) SubmitFinding(ctx context.Context, f *finding.Finding) error {
	// Start span for finding submission
	ctx, span := h.tracer.Start(ctx, "gibson.finding.submit",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.finding.title", f.Title),
			attribute.String("gibson.finding.severity", string(f.Severity)),
		),
	)
	defer span.End()

	// Serialize finding
	findingJSON, err := json.Marshal(f)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("failed to marshal finding: %w", err)
	}

	protoReq := &proto.SubmitFindingRequest{
		FindingJson: string(findingJSON),
	}

	resp, err := h.client.SubmitFinding(ctx, protoReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("submit finding callback failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("submit finding error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return err
	}

	return nil
}

// GetFindings retrieves findings matching the given filter criteria.
func (h *CallbackHarness) GetFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	// Serialize filter
	filterJSON, err := json.Marshal(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal filter: %w", err)
	}

	protoReq := &proto.GetFindingsRequest{
		FilterJson: string(filterJSON),
	}

	resp, err := h.client.GetFindings(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("get findings callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("get findings error: %s", resp.Error.Message)
	}

	// Deserialize findings
	findings := make([]*finding.Finding, 0)
	for _, findingJSON := range resp.FindingsJson {
		var finding *finding.Finding
		if err := json.Unmarshal([]byte(findingJSON), &finding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal finding: %w", err)
		}
		findings = append(findings, finding)
	}

	return findings, nil
}

// ============================================================================
// GraphRAG Query Operations
// ============================================================================

// QueryGraphRAG performs a semantic or hybrid query against the knowledge graph.
func (h *CallbackHarness) QueryGraphRAG(ctx context.Context, query graphrag.Query) ([]graphrag.Result, error) {
	// Start span for GraphRAG query
	ctx, span := h.tracer.Start(ctx, "gibson.graphrag.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("gibson.graphrag.query_text", query.Text),
			attribute.Int("gibson.graphrag.top_k", query.TopK),
		),
	)
	defer span.End()

	// Serialize query
	queryJSON, err := json.Marshal(query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	protoReq := &proto.GraphRAGQueryRequest{
		QueryJson: string(queryJSON),
	}

	resp, err := h.client.GraphRAGQuery(ctx, protoReq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("GraphRAG query callback failed: %w", err)
	}

	if resp.Error != nil {
		err := fmt.Errorf("GraphRAG query error: %s", resp.Error.Message)
		span.RecordError(err)
		span.SetStatus(codes.Error, resp.Error.Message)
		return nil, err
	}

	// Convert results
	results := make([]graphrag.Result, len(resp.Results))
	for i, protoResult := range resp.Results {
		results[i] = graphrag.Result{
			Node:        h.graphNodeFromProto(protoResult.Node),
			Score:       protoResult.Score,
			VectorScore: protoResult.VectorScore,
			GraphScore:  protoResult.GraphScore,
			Path:        protoResult.Path,
			Distance:    int(protoResult.Distance),
		}
	}

	// Record result count in span
	span.SetAttributes(
		attribute.Int("gibson.graphrag.result_count", len(results)),
	)

	return results, nil
}

// FindSimilarAttacks searches for attack patterns semantically similar to the given content.
func (h *CallbackHarness) FindSimilarAttacks(ctx context.Context, content string, topK int) ([]graphrag.AttackPattern, error) {
	protoReq := &proto.FindSimilarAttacksRequest{
		Content: content,
		TopK:    int32(topK),
	}

	resp, err := h.client.FindSimilarAttacks(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("find similar attacks callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("find similar attacks error: %s", resp.Error.Message)
	}

	// Convert attack patterns
	attacks := make([]graphrag.AttackPattern, len(resp.Attacks))
	for i, protoAttack := range resp.Attacks {
		attacks[i] = graphrag.AttackPattern{
			TechniqueID: protoAttack.TechniqueId,
			Name:        protoAttack.Name,
			Description: protoAttack.Description,
			Tactics:     protoAttack.Tactics,
			Platforms:   protoAttack.Platforms,
			Similarity:  protoAttack.Similarity,
		}
	}

	return attacks, nil
}

// FindSimilarFindings searches for findings semantically similar to the referenced finding.
func (h *CallbackHarness) FindSimilarFindings(ctx context.Context, findingID string, topK int) ([]graphrag.FindingNode, error) {
	protoReq := &proto.FindSimilarFindingsRequest{
		FindingId: findingID,
		TopK:      int32(topK),
	}

	resp, err := h.client.FindSimilarFindings(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("find similar findings callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("find similar findings error: %s", resp.Error.Message)
	}

	// Convert finding nodes
	findings := make([]graphrag.FindingNode, len(resp.Findings))
	for i, protoFinding := range resp.Findings {
		findings[i] = graphrag.FindingNode{
			ID:          protoFinding.Id,
			Title:       protoFinding.Title,
			Description: protoFinding.Description,
			Severity:    protoFinding.Severity,
			Category:    protoFinding.Category,
			Confidence:  protoFinding.Confidence,
			Similarity:  protoFinding.Similarity,
		}
	}

	return findings, nil
}

// GetAttackChains discovers multi-step attack paths starting from a technique.
func (h *CallbackHarness) GetAttackChains(ctx context.Context, techniqueID string, maxDepth int) ([]graphrag.AttackChain, error) {
	protoReq := &proto.GetAttackChainsRequest{
		TechniqueId: techniqueID,
		MaxDepth:    int32(maxDepth),
	}

	resp, err := h.client.GetAttackChains(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("get attack chains callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("get attack chains error: %s", resp.Error.Message)
	}

	// Convert attack chains
	chains := make([]graphrag.AttackChain, len(resp.Chains))
	for i, protoChain := range resp.Chains {
		steps := make([]graphrag.AttackStep, len(protoChain.Steps))
		for j, protoStep := range protoChain.Steps {
			steps[j] = graphrag.AttackStep{
				Order:       int(protoStep.Order),
				TechniqueID: protoStep.TechniqueId,
				NodeID:      protoStep.NodeId,
				Description: protoStep.Description,
				Confidence:  protoStep.Confidence,
			}
		}

		chains[i] = graphrag.AttackChain{
			ID:       protoChain.Id,
			Name:     protoChain.Name,
			Severity: protoChain.Severity,
			Steps:    steps,
		}
	}

	return chains, nil
}

// GetRelatedFindings retrieves findings connected via SIMILAR_TO or RELATED_TO relationships.
func (h *CallbackHarness) GetRelatedFindings(ctx context.Context, findingID string) ([]graphrag.FindingNode, error) {
	protoReq := &proto.GetRelatedFindingsRequest{
		FindingId: findingID,
	}

	resp, err := h.client.GetRelatedFindings(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("get related findings callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("get related findings error: %s", resp.Error.Message)
	}

	// Convert finding nodes
	findings := make([]graphrag.FindingNode, len(resp.Findings))
	for i, protoFinding := range resp.Findings {
		findings[i] = graphrag.FindingNode{
			ID:          protoFinding.Id,
			Title:       protoFinding.Title,
			Description: protoFinding.Description,
			Severity:    protoFinding.Severity,
			Category:    protoFinding.Category,
			Confidence:  protoFinding.Confidence,
			Similarity:  protoFinding.Similarity,
		}
	}

	return findings, nil
}

// ============================================================================
// GraphRAG Storage Operations
// ============================================================================

// StoreGraphNode stores an arbitrary node in the knowledge graph.
func (h *CallbackHarness) StoreGraphNode(ctx context.Context, node graphrag.GraphNode) (string, error) {
	protoReq := &proto.StoreGraphNodeRequest{
		Node: h.graphNodeToProto(node),
	}

	resp, err := h.client.StoreGraphNode(ctx, protoReq)
	if err != nil {
		return "", fmt.Errorf("store graph node callback failed: %w", err)
	}

	if resp.Error != nil {
		return "", fmt.Errorf("store graph node error: %s", resp.Error.Message)
	}

	return resp.NodeId, nil
}

// CreateGraphRelationship creates a relationship between two existing nodes.
func (h *CallbackHarness) CreateGraphRelationship(ctx context.Context, rel graphrag.Relationship) error {
	protoReq := &proto.CreateGraphRelationshipRequest{
		Relationship: h.relationshipToProto(rel),
	}

	resp, err := h.client.CreateGraphRelationship(ctx, protoReq)
	if err != nil {
		return fmt.Errorf("create graph relationship callback failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("create graph relationship error: %s", resp.Error.Message)
	}

	return nil
}

// StoreGraphBatch stores multiple nodes and relationships atomically.
func (h *CallbackHarness) StoreGraphBatch(ctx context.Context, batch graphrag.Batch) ([]string, error) {
	// Convert nodes
	protoNodes := make([]*proto.GraphNode, len(batch.Nodes))
	for i, node := range batch.Nodes {
		protoNodes[i] = h.graphNodeToProto(node)
	}

	// Convert relationships
	protoRels := make([]*proto.Relationship, len(batch.Relationships))
	for i, rel := range batch.Relationships {
		protoRels[i] = h.relationshipToProto(rel)
	}

	protoReq := &proto.StoreGraphBatchRequest{
		Nodes:         protoNodes,
		Relationships: protoRels,
	}

	resp, err := h.client.StoreGraphBatch(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("store graph batch callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("store graph batch error: %s", resp.Error.Message)
	}

	return resp.NodeIds, nil
}

// TraverseGraph walks the graph from a starting node following relationships.
func (h *CallbackHarness) TraverseGraph(ctx context.Context, startNodeID string, opts graphrag.TraversalOptions) ([]graphrag.TraversalResult, error) {
	protoReq := &proto.TraverseGraphRequest{
		StartNodeId: startNodeID,
		Options: &proto.TraversalOptions{
			MaxDepth:          int32(opts.MaxDepth),
			RelationshipTypes: opts.RelationshipTypes,
			NodeTypes:         opts.NodeTypes,
			Direction:         opts.Direction,
		},
	}

	resp, err := h.client.TraverseGraph(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("traverse graph callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("traverse graph error: %s", resp.Error.Message)
	}

	// Convert results
	results := make([]graphrag.TraversalResult, len(resp.Results))
	for i, protoResult := range resp.Results {
		results[i] = graphrag.TraversalResult{
			Node:     h.graphNodeFromProto(protoResult.Node),
			Path:     protoResult.Path,
			Distance: int(protoResult.Distance),
		}
	}

	return results, nil
}

// GraphRAGHealth returns the health status of the GraphRAG subsystem.
func (h *CallbackHarness) GraphRAGHealth(ctx context.Context) types.HealthStatus {
	protoReq := &proto.GraphRAGHealthRequest{}

	resp, err := h.client.GraphRAGHealth(ctx, protoReq)
	if err != nil {
		return types.NewUnhealthyStatus(fmt.Sprintf("GraphRAG health check failed: %v", err), nil)
	}

	return types.HealthStatus{
		Status:  resp.Status.State,
		Message: resp.Status.Message,
	}
}

// ============================================================================
// Planning Operations
// ============================================================================

// PlanContext returns the planning context for the current execution.
// Returns nil if no planning context is available (non-planned execution).
func (h *CallbackHarness) PlanContext() planning.PlanningContext {
	return h.planContext
}

// ReportStepHints allows agents to provide feedback to the planning system.
// This forwards the hints to the orchestrator via gRPC callback.
func (h *CallbackHarness) ReportStepHints(ctx context.Context, hints *planning.StepHints) error {
	if hints == nil {
		return nil // Nothing to report
	}

	// Convert to proto message
	protoReq := &proto.ReportStepHintsRequest{
		Hints: &proto.StepHints{
			Confidence:    hints.Confidence(),
			SuggestedNext: hints.SuggestedNext(),
			ReplanReason:  hints.ReplanReason(),
			KeyFindings:   hints.KeyFindings(),
		},
	}

	resp, err := h.client.ReportStepHints(ctx, protoReq)
	if err != nil {
		return fmt.Errorf("report step hints callback failed: %w", err)
	}

	if resp.Error != nil {
		return fmt.Errorf("report step hints error: %s", resp.Error.Message)
	}

	return nil
}

// ============================================================================
// Mission Execution Context Operations
// ============================================================================

// MissionExecutionContext returns the full execution context for the current run
// including run number, resume status, and previous run info.
func (h *CallbackHarness) MissionExecutionContext() types.MissionExecutionContext {
	return h.missionExecCtx
}

// GetMissionRunHistory returns all runs for this mission name.
// Returns runs in chronological order (oldest first).
// Returns empty slice if this is the first run.
//
// Note: This method requires orchestrator callback support. Currently returns
// empty slice until proto and callback client are updated.
func (h *CallbackHarness) GetMissionRunHistory(ctx context.Context) ([]types.MissionRunSummary, error) {
	// TODO: Implement callback to orchestrator once proto is defined
	// For now, return empty slice - orchestrator support pending
	h.logger.Debug("GetMissionRunHistory called - orchestrator callback not yet implemented")
	return []types.MissionRunSummary{}, nil
}

// GetPreviousRunFindings returns findings from the immediate prior run.
// Returns empty slice if no prior run exists.
// Use this to avoid re-discovering known vulnerabilities.
//
// Note: This method requires orchestrator callback support. Currently returns
// empty slice until proto and callback client are updated.
func (h *CallbackHarness) GetPreviousRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	// TODO: Implement callback to orchestrator once proto is defined
	// For now, return empty slice - orchestrator support pending
	h.logger.Debug("GetPreviousRunFindings called - orchestrator callback not yet implemented")
	return []*finding.Finding{}, nil
}

// GetAllRunFindings returns findings from all runs of this mission.
// Useful for comprehensive analysis across the mission's history.
//
// Note: This method requires orchestrator callback support. Currently returns
// empty slice until proto and callback client are updated.
func (h *CallbackHarness) GetAllRunFindings(ctx context.Context, filter finding.Filter) ([]*finding.Finding, error) {
	// TODO: Implement callback to orchestrator once proto is defined
	// For now, return empty slice - orchestrator support pending
	h.logger.Debug("GetAllRunFindings called - orchestrator callback not yet implemented")
	return []*finding.Finding{}, nil
}

// QueryGraphRAGScoped executes a GraphRAG query with explicit scope.
// This modifies the query to include scope and delegates to QueryGraphRAG.
// Scope can be: ScopeCurrentRun, ScopeSameMission, or ScopeAll.
func (h *CallbackHarness) QueryGraphRAGScoped(ctx context.Context, query graphrag.Query, scope graphrag.MissionScope) ([]graphrag.Result, error) {
	// Apply scope to query
	query.MissionScope = scope

	// Add mission context for scoped queries
	if scope == graphrag.ScopeCurrentRun || scope == graphrag.ScopeSameMission {
		query.MissionID = h.mission.ID
		query.MissionName = h.mission.Name
	}

	return h.QueryGraphRAG(ctx, query)
}

// ============================================================================
// Helper Methods for Proto Conversions
// ============================================================================

func (h *CallbackHarness) messagesToProto(messages []llm.Message) []*proto.LLMMessage {
	protoMessages := make([]*proto.LLMMessage, len(messages))
	for i, msg := range messages {
		protoMessages[i] = &proto.LLMMessage{
			Role:        string(msg.Role),
			Content:     msg.Content,
			ToolCalls:   h.toolCallsToProto(msg.ToolCalls),
			ToolResults: h.toolResultsToProto(msg.ToolResults),
			Name:        msg.Name,
		}
	}
	return protoMessages
}

func (h *CallbackHarness) toolCallsToProto(calls []llm.ToolCall) []*proto.ToolCall {
	protoCalls := make([]*proto.ToolCall, len(calls))
	for i, call := range calls {
		protoCalls[i] = &proto.ToolCall{
			Id:        call.ID,
			Name:      call.Name,
			Arguments: call.Arguments,
		}
	}
	return protoCalls
}

func (h *CallbackHarness) toolCallsFromProto(calls []*proto.ToolCall) []llm.ToolCall {
	toolCalls := make([]llm.ToolCall, len(calls))
	for i, call := range calls {
		toolCalls[i] = llm.ToolCall{
			ID:        call.Id,
			Name:      call.Name,
			Arguments: call.Arguments,
		}
	}
	return toolCalls
}

func (h *CallbackHarness) toolResultsToProto(results []llm.ToolResult) []*proto.ToolResult {
	protoResults := make([]*proto.ToolResult, len(results))
	for i, result := range results {
		protoResults[i] = &proto.ToolResult{
			ToolCallId: result.ToolCallID,
			Content:    result.Content,
			IsError:    result.IsError,
		}
	}
	return protoResults
}

func (h *CallbackHarness) toolDefsToProto(tools []llm.ToolDef) []*proto.ToolDef {
	protoTools := make([]*proto.ToolDef, len(tools))
	for i, tool := range tools {
		paramsJSON, err := json.Marshal(tool.Parameters)
		if err != nil {
			h.logger.Error("failed to marshal tool parameters", "error", err, "tool", tool.Name)
			paramsJSON = []byte("{}")
		}
		protoTools[i] = &proto.ToolDef{
			Name:           tool.Name,
			Description:    tool.Description,
			ParametersJson: string(paramsJSON),
		}
	}
	return protoTools
}

func (h *CallbackHarness) graphNodeToProto(node graphrag.GraphNode) *proto.GraphNode {
	propsJSON, err := json.Marshal(node.Properties)
	if err != nil {
		h.logger.Error("failed to marshal graph node properties", "error", err, "node_id", node.ID)
		propsJSON = []byte("{}")
	}
	return &proto.GraphNode{
		Id:             node.ID,
		Type:           node.Type,
		PropertiesJson: string(propsJSON),
		Content:        node.Content,
		MissionId:      node.MissionID,
		AgentName:      node.AgentName,
		CreatedAt:      node.CreatedAt.Unix(),
		UpdatedAt:      node.UpdatedAt.Unix(),
	}
}

func (h *CallbackHarness) graphNodeFromProto(protoNode *proto.GraphNode) graphrag.GraphNode {
	var props map[string]any
	if protoNode.PropertiesJson != "" {
		if err := json.Unmarshal([]byte(protoNode.PropertiesJson), &props); err != nil {
			h.logger.Error("failed to unmarshal graph node properties", "error", err, "node_id", protoNode.Id)
			// Continue with empty props
		}
	}

	return graphrag.GraphNode{
		ID:         protoNode.Id,
		Type:       protoNode.Type,
		Properties: props,
		Content:    protoNode.Content,
		MissionID:  protoNode.MissionId,
		AgentName:  protoNode.AgentName,
	}
}

func (h *CallbackHarness) relationshipToProto(rel graphrag.Relationship) *proto.Relationship {
	propsJSON, err := json.Marshal(rel.Properties)
	if err != nil {
		h.logger.Error("failed to marshal relationship properties", "error", err, "from", rel.FromID, "to", rel.ToID)
		propsJSON = []byte("{}")
	}
	return &proto.Relationship{
		FromId:         rel.FromID,
		ToId:           rel.ToID,
		Type:           rel.Type,
		PropertiesJson: string(propsJSON),
		Bidirectional:  rel.Bidirectional,
	}
}

// formatMessagesForPrompt formats LLM messages into a readable prompt string
// for observability in traces.
func formatMessagesForPrompt(messages []llm.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var result string
	for i, msg := range messages {
		if i > 0 {
			result += "\n---\n"
		}
		result += fmt.Sprintf("[%s]: %s", msg.Role, msg.Content)
	}
	return result
}

// ============================================================================
// MissionManager Methods
// ============================================================================

// CreateMission creates a new mission from a workflow definition.
// This is a stub implementation that will be implemented in a future release.
func (h *CallbackHarness) CreateMission(ctx context.Context, workflow any, targetID string, opts *mission.CreateMissionOpts) (*mission.MissionInfo, error) {
	return nil, fmt.Errorf("mission management not yet implemented in callback harness")
}

// RunMission queues a mission for execution.
// This is a stub implementation that will be implemented in a future release.
func (h *CallbackHarness) RunMission(ctx context.Context, missionID string, opts *mission.RunMissionOpts) error {
	return fmt.Errorf("mission management not yet implemented in callback harness")
}

// GetMissionStatus returns the current state of a mission.
// This is a stub implementation that will be implemented in a future release.
func (h *CallbackHarness) GetMissionStatus(ctx context.Context, missionID string) (*mission.MissionStatusInfo, error) {
	return nil, fmt.Errorf("mission management not yet implemented in callback harness")
}

// WaitForMission blocks until a mission completes or the timeout expires.
// This is a stub implementation that will be implemented in a future release.
func (h *CallbackHarness) WaitForMission(ctx context.Context, missionID string, timeout time.Duration) (*mission.MissionResult, error) {
	return nil, fmt.Errorf("mission management not yet implemented in callback harness")
}

// ListMissions returns missions matching the provided filter criteria.
// This is a stub implementation that will be implemented in a future release.
func (h *CallbackHarness) ListMissions(ctx context.Context, filter *mission.MissionFilter) ([]*mission.MissionInfo, error) {
	return nil, fmt.Errorf("mission management not yet implemented in callback harness")
}

// CancelMission requests cancellation of a running mission.
// This is a stub implementation that will be implemented in a future release.
func (h *CallbackHarness) CancelMission(ctx context.Context, missionID string) error {
	return fmt.Errorf("mission management not yet implemented in callback harness")
}

// GetMissionResults returns the final results of a completed mission.
// This is a stub implementation that will be implemented in a future release.
func (h *CallbackHarness) GetMissionResults(ctx context.Context, missionID string) (*mission.MissionResult, error) {
	return nil, fmt.Errorf("mission management not yet implemented in callback harness")
}

// ============================================================================
// Credential Operations
// ============================================================================

// GetCredential retrieves a credential by name from the credential store.
// The credential is decrypted and returned with its secret value.
// Returns an error if the credential does not exist.
func (h *CallbackHarness) GetCredential(ctx context.Context, name string) (*types.Credential, error) {
	protoReq := &proto.GetCredentialRequest{
		Name: name,
	}

	resp, err := h.client.GetCredential(ctx, protoReq)
	if err != nil {
		return nil, fmt.Errorf("get credential callback failed: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("get credential error: %s", resp.Error.Message)
	}

	if resp.Credential == nil {
		return nil, fmt.Errorf("credential %q not found", name)
	}

	// Convert proto credential type to SDK type
	var credType types.CredentialType
	switch resp.Credential.Type {
	case proto.CredentialType_CREDENTIAL_TYPE_API_KEY:
		credType = types.CredentialTypeAPIKey
	case proto.CredentialType_CREDENTIAL_TYPE_BEARER:
		credType = types.CredentialTypeBearer
	case proto.CredentialType_CREDENTIAL_TYPE_BASIC:
		credType = types.CredentialTypeBasic
	case proto.CredentialType_CREDENTIAL_TYPE_OAUTH:
		credType = types.CredentialTypeOAuth
	case proto.CredentialType_CREDENTIAL_TYPE_CUSTOM:
		credType = types.CredentialTypeCustom
	default:
		credType = types.CredentialTypeAPIKey // Default
	}

	// Parse metadata if present
	var metadata map[string]any
	if resp.Credential.MetadataJson != "" {
		if err := json.Unmarshal([]byte(resp.Credential.MetadataJson), &metadata); err != nil {
			h.logger.Warn("failed to parse credential metadata", "error", err, "name", name)
		}
	}

	return &types.Credential{
		Name:     resp.Credential.Name,
		Type:     credType,
		Secret:   resp.Credential.Secret,
		Username: resp.Credential.Username,
		Metadata: metadata,
	}, nil
}

// ============================================================================
// Proto to Schema Conversion (for taxonomy support)
// ============================================================================

// protoToSchema converts harness callback proto JSONSchemaNode to SDK schema.JSON.
// This reconstructs the full SDK schema with taxonomy from the proto representation.
func protoToSchema(node *proto.JSONSchemaNode) schema.JSON {
	if node == nil {
		return schema.JSON{}
	}

	s := schema.JSON{
		Type:        node.Type,
		Description: node.Description,
		Required:    node.Required,
		Format:      node.Format,
	}

	if node.Pattern != nil {
		s.Pattern = *node.Pattern
	}

	// Convert properties recursively
	if len(node.Properties) > 0 {
		s.Properties = make(map[string]schema.JSON)
		for name, prop := range node.Properties {
			s.Properties[name] = protoToSchema(prop)
		}
	}

	// Convert items recursively
	if node.Items != nil {
		items := protoToSchema(node.Items)
		s.Items = &items
	}

	// Convert enum values
	if len(node.EnumValues) > 0 {
		for _, v := range node.EnumValues {
			s.Enum = append(s.Enum, v)
		}
	}

	// Convert numeric constraints
	if node.Minimum != nil {
		s.Minimum = node.Minimum
	}
	if node.Maximum != nil {
		s.Maximum = node.Maximum
	}
	if node.MinLength != nil {
		minLen := int(*node.MinLength)
		s.MinLength = &minLen
	}
	if node.MaxLength != nil {
		maxLen := int(*node.MaxLength)
		s.MaxLength = &maxLen
	}

	// Convert default value (JSON decode)
	if node.DefaultValue != nil && *node.DefaultValue != "" {
		var def any
		if err := json.Unmarshal([]byte(*node.DefaultValue), &def); err == nil {
			s.Default = def
		}
	}

	// Convert taxonomy (THE KEY FEATURE for knowledge graph extraction)
	if node.Taxonomy != nil {
		s.Taxonomy = protoToTaxonomy(node.Taxonomy)
	}

	return s
}

// protoToTaxonomy converts harness callback proto TaxonomyMapping to SDK TaxonomyMapping.
func protoToTaxonomy(p *proto.TaxonomyMapping) *schema.TaxonomyMapping {
	if p == nil {
		return nil
	}

	t := &schema.TaxonomyMapping{
		NodeType:   p.NodeType,
		IDTemplate: p.IdTemplate,
	}

	// Convert property mappings
	for _, prop := range p.Properties {
		t.Properties = append(t.Properties, schema.PropertyMapping{
			Source: prop.Source,
			Target: prop.Target,
		})
	}

	// Convert relationship mappings
	for _, rel := range p.Relationships {
		t.Relationships = append(t.Relationships, schema.RelationshipMapping{
			Type:       rel.Type,
			ToTemplate: rel.TargetTemplate,
		})
	}

	return t
}

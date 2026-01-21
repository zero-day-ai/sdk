package serve

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// CallbackClient manages the gRPC connection to the orchestrator's HarnessCallbackService.
// It provides thread-safe access to all harness operations via RPC callbacks.
type CallbackClient struct {
	// Connection management
	conn   *grpc.ClientConn
	client proto.HarnessCallbackServiceClient
	mu     sync.RWMutex

	// Configuration
	endpoint string
	tlsConf  *tls.Config
	token    string

	// Context tracking
	taskID          string
	agentName       string
	missionID       string
	traceID         string
	spanID          string
	missionRunID    string // Unique ID for this mission execution
	agentRunID      string // Unique ID for this agent execution
	runNumber       int32  // Sequential run number (1, 2, 3...)
	toolExecutionID string // ID for tool execution provenance

	// Connection lifecycle
	connected bool
	closed    bool
}

// NewCallbackClient creates a new callback client with the given endpoint.
// The client is not connected until Connect() is called.
func NewCallbackClient(endpoint string, opts ...CallbackClientOption) (*CallbackClient, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint cannot be empty")
	}

	client := &CallbackClient{
		endpoint: endpoint,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// CallbackClientOption is a functional option for configuring CallbackClient.
type CallbackClientOption func(*CallbackClient)

// WithCallbackTLS configures TLS for the callback client connection.
func WithCallbackTLS(conf *tls.Config) CallbackClientOption {
	return func(c *CallbackClient) {
		c.tlsConf = conf
	}
}

// WithCallbackToken sets the authentication token for callback requests.
func WithCallbackToken(token string) CallbackClientOption {
	return func(c *CallbackClient) {
		c.token = token
	}
}

// Connect establishes the gRPC connection to the orchestrator.
// This must be called before any RPC methods can be invoked.
func (c *CallbackClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return fmt.Errorf("client is closed")
	}

	// Check if already connected AND connection is actually healthy
	if c.connected && c.conn != nil {
		state := c.conn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			return nil // Already connected and healthy
		}
		// Connection exists but is unhealthy - close and reconnect
		c.conn.Close()
		c.connected = false
	}

	// Build dial options
	var dialOpts []grpc.DialOption

	// Configure transport credentials
	if c.tlsConf != nil {
		creds := credentials.NewTLS(c.tlsConf)
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add keepalive configuration
	dialOpts = append(dialOpts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             5 * time.Second,
		PermitWithoutStream: true,
	}))

	// Create context with timeout for connection establishment
	connCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Establish connection
	conn, err := grpc.DialContext(connCtx, c.endpoint, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	c.conn = conn
	c.client = proto.NewHarnessCallbackServiceClient(conn)
	c.connected = true

	// Wait for connection to be ready (with timeout)
	// This ensures the connection is actually established, not just dialed
	readyCtx, readyCancel := context.WithTimeout(ctx, 5*time.Second)
	defer readyCancel()
	c.conn.WaitForStateChange(readyCtx, connectivity.Idle)
	state := c.conn.GetState()
	if state == connectivity.TransientFailure || state == connectivity.Shutdown {
		return fmt.Errorf("connection failed to establish: state=%s", state)
	}

	return nil
}

// SetTaskContext updates the task context for subsequent RPC calls.
// This should be called at the start of each task execution.
// Deprecated: Use SetFullContext instead which includes all context fields.
func (c *CallbackClient) SetTaskContext(taskID, agentName, missionID, traceID, spanID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.taskID = taskID
	c.agentName = agentName
	c.missionID = missionID
	c.traceID = traceID
	c.spanID = spanID
}

// TaskContextParams contains all the context parameters for RPC calls.
type TaskContextParams struct {
	TaskID          string
	AgentName       string
	MissionID       string
	TraceID         string
	SpanID          string
	MissionRunID    string // Unique ID for this mission execution
	AgentRunID      string // Unique ID for this agent execution
	RunNumber       int32  // Sequential run number (1, 2, 3...)
	ToolExecutionID string // ID for tool execution provenance
}

// SetFullContext updates the complete task context for subsequent RPC calls.
// This should be called at the start of each task execution with all available context.
func (c *CallbackClient) SetFullContext(params TaskContextParams) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.taskID = params.TaskID
	c.agentName = params.AgentName
	c.missionID = params.MissionID
	c.traceID = params.TraceID
	c.spanID = params.SpanID
	c.missionRunID = params.MissionRunID
	c.agentRunID = params.AgentRunID
	c.runNumber = params.RunNumber
	c.toolExecutionID = params.ToolExecutionID
}

// contextInfo builds the ContextInfo proto message with current task context.
func (c *CallbackClient) contextInfo() *proto.ContextInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &proto.ContextInfo{
		TaskId:          c.taskID,
		AgentName:       c.agentName,
		MissionId:       c.missionID,
		TraceId:         c.traceID,
		SpanId:          c.spanID,
		MissionRunId:    c.missionRunID,
		AgentRunId:      c.agentRunID,
		RunNumber:       c.runNumber,
		ToolExecutionId: c.toolExecutionID,
	}
}

// contextWithMetadata creates a context with authentication metadata if a token is set.
func (c *CallbackClient) contextWithMetadata(ctx context.Context) context.Context {
	if c.token == "" {
		return ctx
	}

	md := metadata.New(map[string]string{
		"authorization": "Bearer " + c.token,
	})
	return metadata.NewOutgoingContext(ctx, md)
}

// Close closes the gRPC connection and cleans up resources.
// The client cannot be reused after Close() is called.
func (c *CallbackClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	c.connected = false

	if c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

// IsConnected returns true if the client is connected to the orchestrator.
// This checks both the internal state and the actual gRPC connection state.
func (c *CallbackClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.connected || c.closed || c.conn == nil {
		return false
	}

	// Check actual gRPC connection state
	// Accept Ready, Idle, and Connecting as valid states (connection may be establishing)
	state := c.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle || state == connectivity.Connecting
}

// ============================================================================
// LLM Operations
// ============================================================================

// LLMComplete performs an LLM completion request via the orchestrator.
func (c *CallbackClient) LLMComplete(ctx context.Context, req *proto.LLMCompleteRequest) (*proto.LLMCompleteResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LLMComplete: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.LLMComplete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLMComplete: %w", err)
	}
	return resp, nil
}

// LLMCompleteWithTools performs an LLM completion with tool calling enabled.
func (c *CallbackClient) LLMCompleteWithTools(ctx context.Context, req *proto.LLMCompleteWithToolsRequest) (*proto.LLMCompleteResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LLMCompleteWithTools: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.LLMCompleteWithTools(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLMCompleteWithTools: %w", err)
	}
	return resp, nil
}

// LLMCompleteStructured performs an LLM completion with structured output via the orchestrator.
func (c *CallbackClient) LLMCompleteStructured(ctx context.Context, req *proto.LLMCompleteStructuredRequest) (*proto.LLMCompleteStructuredResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LLMCompleteStructured: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.LLMCompleteStructured(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLMCompleteStructured: %w", err)
	}
	return resp, nil
}

// LLMStream performs a streaming LLM completion request.
func (c *CallbackClient) LLMStream(ctx context.Context, req *proto.LLMStreamRequest) (proto.HarnessCallbackService_LLMStreamClient, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LLMStream: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.LLMStream(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLMStream: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Tool Operations
// ============================================================================

// CallTool invokes a tool via the orchestrator.
func (c *CallbackClient) CallTool(ctx context.Context, req *proto.CallToolRequest) (*proto.CallToolResponse, error) {
	// Try to reconnect if not connected
	if !c.IsConnected() {
		if err := c.Connect(ctx); err != nil {
			return nil, fmt.Errorf("CallTool: client not connected and reconnect failed: %w", err)
		}
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.CallTool(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CallTool: %w", err)
	}
	return resp, nil
}

// ListTools retrieves the list of available tools.
func (c *CallbackClient) ListTools(ctx context.Context, req *proto.ListToolsRequest) (*proto.ListToolsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("ListTools: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.ListTools(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListTools: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Plugin Operations
// ============================================================================

// QueryPlugin sends a query to a plugin via the orchestrator.
func (c *CallbackClient) QueryPlugin(ctx context.Context, req *proto.QueryPluginRequest) (*proto.QueryPluginResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("QueryPlugin: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.QueryPlugin(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("QueryPlugin: %w", err)
	}
	return resp, nil
}

// ListPlugins retrieves the list of available plugins.
func (c *CallbackClient) ListPlugins(ctx context.Context, req *proto.ListPluginsRequest) (*proto.ListPluginsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("ListPlugins: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.ListPlugins(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListPlugins: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Agent Operations
// ============================================================================

// DelegateToAgent delegates a task to another agent.
func (c *CallbackClient) DelegateToAgent(ctx context.Context, req *proto.DelegateToAgentRequest) (*proto.DelegateToAgentResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("DelegateToAgent: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.DelegateToAgent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("DelegateToAgent: %w", err)
	}
	return resp, nil
}

// ListAgents retrieves the list of available agents.
func (c *CallbackClient) ListAgents(ctx context.Context, req *proto.ListAgentsRequest) (*proto.ListAgentsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("ListAgents: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.ListAgents(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListAgents: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Finding Operations
// ============================================================================

// SubmitFinding submits a security finding to the orchestrator.
func (c *CallbackClient) SubmitFinding(ctx context.Context, req *proto.SubmitFindingRequest) (*proto.SubmitFindingResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("SubmitFinding: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.SubmitFinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("SubmitFinding: %w", err)
	}
	return resp, nil
}

// GetFindings retrieves findings matching the filter criteria.
func (c *CallbackClient) GetFindings(ctx context.Context, req *proto.GetFindingsRequest) (*proto.GetFindingsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GetFindings: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GetFindings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetFindings: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Memory Operations
// ============================================================================

// MemoryGet retrieves a value from memory.
func (c *CallbackClient) MemoryGet(ctx context.Context, req *proto.MemoryGetRequest) (*proto.MemoryGetResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MemoryGet: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MemoryGet(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MemoryGet: %w", err)
	}
	return resp, nil
}

// MemorySet stores a value in memory.
func (c *CallbackClient) MemorySet(ctx context.Context, req *proto.MemorySetRequest) (*proto.MemorySetResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MemorySet: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MemorySet(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MemorySet: %w", err)
	}
	return resp, nil
}

// MemoryDelete removes a value from memory.
func (c *CallbackClient) MemoryDelete(ctx context.Context, req *proto.MemoryDeleteRequest) (*proto.MemoryDeleteResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MemoryDelete: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MemoryDelete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MemoryDelete: %w", err)
	}
	return resp, nil
}

// MemoryList lists all keys matching a prefix.
func (c *CallbackClient) MemoryList(ctx context.Context, req *proto.MemoryListRequest) (*proto.MemoryListResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MemoryList: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MemoryList(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MemoryList: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Mission Memory Operations
// ============================================================================

// MissionMemorySearch performs a full-text search on mission memory.
func (c *CallbackClient) MissionMemorySearch(ctx context.Context, req *proto.MissionMemorySearchRequest) (*proto.MissionMemorySearchResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MissionMemorySearch: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MissionMemorySearch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MissionMemorySearch: %w", err)
	}
	return resp, nil
}

// MissionMemoryHistory retrieves recent mission memory entries.
func (c *CallbackClient) MissionMemoryHistory(ctx context.Context, req *proto.MissionMemoryHistoryRequest) (*proto.MissionMemoryHistoryResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MissionMemoryHistory: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MissionMemoryHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MissionMemoryHistory: %w", err)
	}
	return resp, nil
}

// MissionMemoryGetPreviousRunValue retrieves a value from a previous mission run.
func (c *CallbackClient) MissionMemoryGetPreviousRunValue(ctx context.Context, req *proto.MissionMemoryGetPreviousRunValueRequest) (*proto.MissionMemoryGetPreviousRunValueResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MissionMemoryGetPreviousRunValue: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MissionMemoryGetPreviousRunValue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MissionMemoryGetPreviousRunValue: %w", err)
	}
	return resp, nil
}

// MissionMemoryGetValueHistory retrieves the history of values for a key across runs.
func (c *CallbackClient) MissionMemoryGetValueHistory(ctx context.Context, req *proto.MissionMemoryGetValueHistoryRequest) (*proto.MissionMemoryGetValueHistoryResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MissionMemoryGetValueHistory: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MissionMemoryGetValueHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MissionMemoryGetValueHistory: %w", err)
	}
	return resp, nil
}

// MissionMemoryContinuityMode retrieves the current mission memory continuity mode.
func (c *CallbackClient) MissionMemoryContinuityMode(ctx context.Context, req *proto.MissionMemoryContinuityModeRequest) (*proto.MissionMemoryContinuityModeResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("MissionMemoryContinuityMode: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.MissionMemoryContinuityMode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("MissionMemoryContinuityMode: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Long-Term Memory Operations
// ============================================================================

// LongTermMemoryStore stores content in long-term vector memory.
func (c *CallbackClient) LongTermMemoryStore(ctx context.Context, req *proto.LongTermMemoryStoreRequest) (*proto.LongTermMemoryStoreResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LongTermMemoryStore: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.LongTermMemoryStore(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LongTermMemoryStore: %w", err)
	}
	return resp, nil
}

// LongTermMemorySearch performs a semantic search on long-term memory.
func (c *CallbackClient) LongTermMemorySearch(ctx context.Context, req *proto.LongTermMemorySearchRequest) (*proto.LongTermMemorySearchResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LongTermMemorySearch: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.LongTermMemorySearch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LongTermMemorySearch: %w", err)
	}
	return resp, nil
}

// LongTermMemoryDelete removes an entry from long-term memory.
func (c *CallbackClient) LongTermMemoryDelete(ctx context.Context, req *proto.LongTermMemoryDeleteRequest) (*proto.LongTermMemoryDeleteResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("LongTermMemoryDelete: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.LongTermMemoryDelete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LongTermMemoryDelete: %w", err)
	}
	return resp, nil
}

// ============================================================================
// GraphRAG Query Operations
// ============================================================================

// GraphRAGQuery performs a GraphRAG query.
func (c *CallbackClient) GraphRAGQuery(ctx context.Context, req *proto.GraphRAGQueryRequest) (*proto.GraphRAGQueryResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GraphRAGQuery: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GraphRAGQuery(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GraphRAGQuery: %w", err)
	}
	return resp, nil
}

// FindSimilarAttacks searches for similar attack patterns.
func (c *CallbackClient) FindSimilarAttacks(ctx context.Context, req *proto.FindSimilarAttacksRequest) (*proto.FindSimilarAttacksResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("FindSimilarAttacks: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.FindSimilarAttacks(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("FindSimilarAttacks: %w", err)
	}
	return resp, nil
}

// FindSimilarFindings searches for similar findings.
func (c *CallbackClient) FindSimilarFindings(ctx context.Context, req *proto.FindSimilarFindingsRequest) (*proto.FindSimilarFindingsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("FindSimilarFindings: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.FindSimilarFindings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("FindSimilarFindings: %w", err)
	}
	return resp, nil
}

// GetAttackChains discovers attack chains starting from a technique.
func (c *CallbackClient) GetAttackChains(ctx context.Context, req *proto.GetAttackChainsRequest) (*proto.GetAttackChainsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GetAttackChains: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GetAttackChains(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetAttackChains: %w", err)
	}
	return resp, nil
}

// GetRelatedFindings retrieves related findings.
func (c *CallbackClient) GetRelatedFindings(ctx context.Context, req *proto.GetRelatedFindingsRequest) (*proto.GetRelatedFindingsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GetRelatedFindings: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GetRelatedFindings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetRelatedFindings: %w", err)
	}
	return resp, nil
}

// ============================================================================
// GraphRAG Storage Operations
// ============================================================================

// StoreGraphNode stores a node in the knowledge graph.
func (c *CallbackClient) StoreGraphNode(ctx context.Context, req *proto.StoreGraphNodeRequest) (*proto.StoreGraphNodeResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("StoreGraphNode: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.StoreGraphNode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("StoreGraphNode: %w", err)
	}
	return resp, nil
}

// CreateGraphRelationship creates a relationship between nodes.
func (c *CallbackClient) CreateGraphRelationship(ctx context.Context, req *proto.CreateGraphRelationshipRequest) (*proto.CreateGraphRelationshipResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("CreateGraphRelationship: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.CreateGraphRelationship(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreateGraphRelationship: %w", err)
	}
	return resp, nil
}

// StoreGraphBatch stores multiple nodes and relationships atomically.
func (c *CallbackClient) StoreGraphBatch(ctx context.Context, req *proto.StoreGraphBatchRequest) (*proto.StoreGraphBatchResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("StoreGraphBatch: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.StoreGraphBatch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("StoreGraphBatch: %w", err)
	}
	return resp, nil
}

// TraverseGraph walks the graph from a starting node.
func (c *CallbackClient) TraverseGraph(ctx context.Context, req *proto.TraverseGraphRequest) (*proto.TraverseGraphResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("TraverseGraph: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.TraverseGraph(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TraverseGraph: %w", err)
	}
	return resp, nil
}

// GraphRAGHealth checks the health of the GraphRAG subsystem.
func (c *CallbackClient) GraphRAGHealth(ctx context.Context, req *proto.GraphRAGHealthRequest) (*proto.GraphRAGHealthResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GraphRAGHealth: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GraphRAGHealth(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GraphRAGHealth: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Planning Operations
// ============================================================================

// GetPlanContext retrieves the planning context from the orchestrator.
func (c *CallbackClient) GetPlanContext(ctx context.Context, req *proto.GetPlanContextRequest) (*proto.GetPlanContextResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GetPlanContext: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GetPlanContext(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetPlanContext: %w", err)
	}
	return resp, nil
}

// ReportStepHints reports step hints to the orchestrator.
func (c *CallbackClient) ReportStepHints(ctx context.Context, req *proto.ReportStepHintsRequest) (*proto.ReportStepHintsResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("ReportStepHints: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.ReportStepHints(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ReportStepHints: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Tracing Operations
// ============================================================================

// RecordSpans sends a batch of spans to the orchestrator for distributed tracing.
func (c *CallbackClient) RecordSpans(ctx context.Context, req *proto.RecordSpansRequest) (*proto.RecordSpansResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("RecordSpans: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.RecordSpans(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("RecordSpans: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Credential Operations
// ============================================================================

// GetCredential retrieves a credential by name from the orchestrator's credential store.
func (c *CallbackClient) GetCredential(ctx context.Context, req *proto.GetCredentialRequest) (*proto.GetCredentialResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GetCredential: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GetCredential(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetCredential: %w", err)
	}
	return resp, nil
}

// ============================================================================
// Taxonomy Operations
// ============================================================================

// GetTaxonomySchema retrieves the full taxonomy schema from the orchestrator.
func (c *CallbackClient) GetTaxonomySchema(ctx context.Context, req *proto.GetTaxonomySchemaRequest) (*proto.GetTaxonomySchemaResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GetTaxonomySchema: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GetTaxonomySchema(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetTaxonomySchema: %w", err)
	}
	return resp, nil
}

// GenerateNodeID generates a deterministic node ID using taxonomy templates.
func (c *CallbackClient) GenerateNodeID(ctx context.Context, req *proto.GenerateNodeIDRequest) (*proto.GenerateNodeIDResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("GenerateNodeID: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.GenerateNodeID(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GenerateNodeID: %w", err)
	}
	return resp, nil
}

// ValidateFinding validates a finding against the taxonomy schema.
func (c *CallbackClient) ValidateFinding(ctx context.Context, req *proto.ValidateFindingRequest) (*proto.ValidationResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("ValidateFinding: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.ValidateFinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ValidateFinding: %w", err)
	}
	return resp, nil
}

// ValidateGraphNode validates a graph node against the taxonomy schema.
func (c *CallbackClient) ValidateGraphNode(ctx context.Context, req *proto.ValidateGraphNodeRequest) (*proto.ValidationResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("ValidateGraphNode: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.ValidateGraphNode(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ValidateGraphNode: %w", err)
	}
	return resp, nil
}

// ValidateRelationship validates a relationship against the taxonomy schema.
func (c *CallbackClient) ValidateRelationship(ctx context.Context, req *proto.ValidateRelationshipRequest) (*proto.ValidationResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("ValidateRelationship: client not connected")
	}

	req.Context = c.contextInfo()
	ctx = c.contextWithMetadata(ctx)
	resp, err := c.client.ValidateRelationship(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ValidateRelationship: %w", err)
	}
	return resp, nil
}

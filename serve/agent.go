package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/types"
	"go.opentelemetry.io/otel/sdk/trace"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Agent starts a gRPC server for an agent.
// It creates a server, registers the agent service, and serves requests
// until a shutdown signal is received or an error occurs.
//
// Example:
//
//	agent := &MyAgent{}
//	err := serve.Agent(agent,
//	    serve.WithPort(50051),
//	    serve.WithGracefulShutdown(30*time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Agent(a agent.Agent, opts ...Option) error {
	// Build configuration
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Create server
	srv, err := NewServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Create and register agent service
	agentSvc := &agentServiceServer{
		agent: a,
	}
	proto.RegisterAgentServiceServer(srv.GRPCServer(), agentSvc)

	// Set health status to serving
	srv.HealthServer().SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	slog.Info("agent server started", "component", "agent", "name", a.Name(), "version", a.Version(), "port", srv.Port())

	// Register with registry if configured
	var serviceInfo interface{}
	if cfg.Registry != nil {
		// Build endpoint based on LocalMode, AdvertiseAddr, or TCP
		endpoint := ""
		if cfg.LocalMode != "" {
			endpoint = fmt.Sprintf("unix://%s", cfg.LocalMode)
		} else if cfg.AdvertiseAddr != "" {
			// Use advertise address - append port if not present
			if strings.Contains(cfg.AdvertiseAddr, ":") {
				endpoint = cfg.AdvertiseAddr
			} else {
				endpoint = fmt.Sprintf("%s:%d", cfg.AdvertiseAddr, srv.Port())
			}
		} else {
			endpoint = fmt.Sprintf("localhost:%d", srv.Port())
		}

		// Extract agent metadata
		capabilities := make([]string, len(a.Capabilities()))
		for i, cap := range a.Capabilities() {
			capabilities[i] = cap.String()
		}

		targetTypes := make([]string, len(a.TargetTypes()))
		for i, tt := range a.TargetTypes() {
			targetTypes[i] = tt.String()
		}

		techniqueTypes := make([]string, len(a.TechniqueTypes()))
		for i, tt := range a.TechniqueTypes() {
			techniqueTypes[i] = tt.String()
		}

		// Create ServiceInfo struct (using map to avoid circular dependency)
		serviceInfo = map[string]interface{}{
			"kind":        "agent",
			"name":        a.Name(),
			"version":     a.Version(),
			"instance_id": uuid.New().String(),
			"endpoint":    endpoint,
			"metadata": map[string]string{
				"description":     a.Description(),
				"capabilities":    strings.Join(capabilities, ","),
				"target_types":    strings.Join(targetTypes, ","),
				"technique_types": strings.Join(techniqueTypes, ","),
			},
			"started_at": time.Now(),
		}

		// Register with the registry
		ctx := context.Background()
		if err := cfg.Registry.Register(ctx, serviceInfo); err != nil {
			slog.Warn("failed to register with registry", "error", err, "endpoint", endpoint, "component", "agent", "name", a.Name())
		} else {
			slog.Info("registered with registry", "endpoint", endpoint, "component", "agent", "name", a.Name())
			// Deregister on shutdown
			defer func() {
				ctx := context.Background()
				if err := cfg.Registry.Deregister(ctx, serviceInfo); err != nil {
					slog.Warn("failed to deregister from registry", "error", err, "endpoint", endpoint, "component", "agent", "name", a.Name())
				}
			}()
		}
	}

	// Start serving
	return srv.Serve(context.Background())
}

// agentServiceServer implements the gRPC AgentService for an SDK agent.
// It bridges the gRPC protocol to the agent.Agent interface.
type agentServiceServer struct {
	proto.UnimplementedAgentServiceServer
	agent agent.Agent
}

// GetDescriptor returns the agent's descriptor including name, version,
// capabilities, target types, and technique types.
func (s *agentServiceServer) GetDescriptor(ctx context.Context, req *proto.AgentGetDescriptorRequest) (*proto.AgentDescriptor, error) {
	capabilities := make([]string, len(s.agent.Capabilities()))
	for i, cap := range s.agent.Capabilities() {
		capabilities[i] = cap.String()
	}

	targetTypes := make([]string, len(s.agent.TargetTypes()))
	for i, tt := range s.agent.TargetTypes() {
		targetTypes[i] = tt.String()
	}

	techniqueTypes := make([]string, len(s.agent.TechniqueTypes()))
	for i, tt := range s.agent.TechniqueTypes() {
		techniqueTypes[i] = tt.String()
	}

	return &proto.AgentDescriptor{
		Name:           s.agent.Name(),
		Version:        s.agent.Version(),
		Description:    s.agent.Description(),
		Capabilities:   capabilities,
		TargetTypes:    targetTypes,
		TechniqueTypes: techniqueTypes,
	}, nil
}

// GetSlotSchema returns the LLM slot definitions required by the agent.
// Each slot defines requirements and constraints for LLM provisioning.
func (s *agentServiceServer) GetSlotSchema(ctx context.Context, req *proto.AgentGetSlotSchemaRequest) (*proto.AgentGetSlotSchemaResponse, error) {
	slots := s.agent.LLMSlots()
	protoSlots := make([]*proto.AgentSlotDefinition, len(slots))

	for i, slot := range slots {
		// Note: The current SlotDefinition doesn't have DefaultConfig or separate Constraints fields.
		// We map the available fields to the proto structure.
		var constraints *proto.AgentSlotConstraints
		if len(slot.RequiredFeatures) > 0 || slot.MinContextWindow > 0 {
			constraints = &proto.AgentSlotConstraints{
				MinContextWindow: int32(slot.MinContextWindow),
				RequiredFeatures: slot.RequiredFeatures,
			}
		}

		// Use PreferredModels to suggest a default config if available
		var defaultConfig *proto.AgentSlotConfig
		if len(slot.PreferredModels) > 0 {
			defaultConfig = &proto.AgentSlotConfig{
				Model: slot.PreferredModels[0], // Use first preferred model as default
			}
		}

		protoSlots[i] = &proto.AgentSlotDefinition{
			Name:          slot.Name,
			Description:   slot.Description,
			Required:      slot.Required,
			DefaultConfig: defaultConfig,
			Constraints:   constraints,
		}
	}

	return &proto.AgentGetSlotSchemaResponse{
		Slots: protoSlots,
	}, nil
}

// Execute runs the agent with the provided task.
// The task is serialized as JSON in the request and the result is
// serialized as JSON in the response.
func (s *agentServiceServer) Execute(ctx context.Context, req *proto.AgentExecuteRequest) (*proto.AgentExecuteResponse, error) {
	// Parse task from JSON
	var task agent.Task
	if err := json.Unmarshal([]byte(req.TaskJson), &task); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid task JSON: %v", err)
	}

	// Apply timeout if specified
	if req.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	// Create harness if callback endpoint is provided
	var harness agent.Harness
	var tracerProvider *trace.TracerProvider
	if req.CallbackEndpoint != "" {
		callbackHarness, tp, err := s.createCallbackHarness(ctx, req, task)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create callback harness: %v", err)
		}
		harness = callbackHarness
		tracerProvider = tp
		// Clean up callback client and tracer provider when done
		defer func() {
			if ch, ok := harness.(*CallbackHarness); ok && ch.client != nil {
				ch.client.Close()
			}
			if tracerProvider != nil {
				if err := tracerProvider.Shutdown(context.Background()); err != nil {
					slog.Warn("failed to shutdown tracer provider", "error", err)
				}
			}
		}()
	}

	// Execute the agent with the harness (may be nil if no callback endpoint)
	result, err := s.agent.Execute(ctx, harness, task)

	// Build response
	resp := &proto.AgentExecuteResponse{}

	// Serialize result as JSON
	if err == nil {
		resultJSON, jsonErr := json.Marshal(result)
		if jsonErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal result: %v", jsonErr)
		}
		resp.ResultJson = string(resultJSON)
	} else {
		// Map error to proto error
		resp.Error = &proto.Error{
			Code:      "EXECUTION_ERROR",
			Message:   err.Error(),
			Retryable: false,
		}
	}

	return resp, nil
}

// createCallbackHarness creates a CallbackHarness connected to the orchestrator.
func (s *agentServiceServer) createCallbackHarness(ctx context.Context, req *proto.AgentExecuteRequest, task agent.Task) (*CallbackHarness, *trace.TracerProvider, error) {
	// Create callback client options
	var clientOpts []CallbackClientOption
	if req.CallbackToken != "" {
		clientOpts = append(clientOpts, WithCallbackToken(req.CallbackToken))
	}

	// Create and connect callback client
	client, err := NewCallbackClient(req.CallbackEndpoint, clientOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create callback client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	// Parse mission context if provided
	var mission types.MissionContext
	if req.MissionJson != "" {
		if err := json.Unmarshal([]byte(req.MissionJson), &mission); err != nil {
			client.Close()
			return nil, nil, fmt.Errorf("failed to parse mission JSON: %w", err)
		}
	}

	// Set task context for callback requests
	// Pass mission ID explicitly so the callback service can use it directly
	// for mission-based harness lookup (keyed by missionID:agentName)
	client.SetTaskContext(task.ID, s.agent.Name(), mission.ID, req.TraceId, req.ParentSpanId)

	// Parse target info if provided
	var target types.TargetInfo
	if req.TargetJson != "" {
		if err := json.Unmarshal([]byte(req.TargetJson), &target); err != nil {
			client.Close()
			return nil, nil, fmt.Errorf("failed to parse target JSON: %w", err)
		}
	}

	// Create logger for this agent execution
	logger := slog.Default().With(
		"agent", s.agent.Name(),
		"task_id", task.ID,
	)

	// Create tracer based on whether trace context is present
	var tracer otelTrace.Tracer
	var tracerProvider *trace.TracerProvider
	if req.TraceId != "" {
		// Create real tracer with proxy exporter
		tracerProvider = NewProxyTracerProvider(client, req.TraceId, req.ParentSpanId, logger)
		tracer = tracerProvider.Tracer("gibson-agent")
		logger.Debug("created real tracer for distributed tracing",
			"trace_id", req.TraceId,
			"parent_span_id", req.ParentSpanId,
		)
	} else {
		// Use no-op tracer when no trace context is provided
		tracer = noop.NewTracerProvider().Tracer("gibson-agent")
		logger.Debug("created no-op tracer (no trace context)")
	}

	// Create the callback harness
	harness := NewCallbackHarness(client, logger, tracer, mission, target)

	return harness, tracerProvider, nil
}

// Health returns the current health status of the agent.
func (s *agentServiceServer) Health(ctx context.Context, req *proto.AgentHealthRequest) (*proto.HealthStatus, error) {
	health := s.agent.Health(ctx)

	return &proto.HealthStatus{
		State:     health.Status,
		Message:   health.Message,
		CheckedAt: time.Now().UnixMilli(),
	}, nil
}

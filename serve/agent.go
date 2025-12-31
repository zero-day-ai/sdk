package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/agent"
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

	fmt.Printf("Agent %s v%s listening on :%d\n", a.Name(), a.Version(), srv.Port())

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

	// Execute the agent
	// Note: This is a simplified implementation. In a real implementation,
	// the harness would be provided by the framework. For now, we pass nil
	// and agents should handle this gracefully.
	result, err := s.agent.Execute(ctx, nil, task)

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

// Health returns the current health status of the agent.
func (s *agentServiceServer) Health(ctx context.Context, req *proto.AgentHealthRequest) (*proto.HealthStatus, error) {
	health := s.agent.Health(ctx)

	return &proto.HealthStatus{
		State:     health.Status,
		Message:   health.Message,
		CheckedAt: time.Now().UnixMilli(),
	}, nil
}

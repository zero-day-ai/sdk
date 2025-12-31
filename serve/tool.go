package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/tool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Tool starts a gRPC server for a tool.
// It creates a server, registers the tool service, and serves requests
// until a shutdown signal is received or an error occurs.
//
// Example:
//
//	tool := &MyTool{}
//	err := serve.Tool(tool,
//	    serve.WithPort(50052),
//	    serve.WithGracefulShutdown(30*time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Tool(t tool.Tool, opts ...Option) error {
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

	// Create and register tool service
	toolSvc := &toolServiceServer{
		tool: t,
	}
	proto.RegisterToolServiceServer(srv.GRPCServer(), toolSvc)

	// Set health status to serving
	srv.HealthServer().SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	fmt.Printf("Tool %s v%s listening on :%d\n", t.Name(), t.Version(), srv.Port())

	// Start serving
	return srv.Serve(context.Background())
}

// toolServiceServer implements the gRPC ToolService for an SDK tool.
// It bridges the gRPC protocol to the tool.Tool interface.
type toolServiceServer struct {
	proto.UnimplementedToolServiceServer
	tool tool.Tool
}

// GetDescriptor returns the tool's descriptor including name, version,
// description, tags, and input/output schemas.
func (s *toolServiceServer) GetDescriptor(ctx context.Context, req *proto.ToolGetDescriptorRequest) (*proto.ToolDescriptor, error) {
	// Serialize input schema to JSON
	inputSchemaJSON, err := json.Marshal(s.tool.InputSchema())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal input schema: %v", err)
	}

	// Serialize output schema to JSON
	outputSchemaJSON, err := json.Marshal(s.tool.OutputSchema())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal output schema: %v", err)
	}

	return &proto.ToolDescriptor{
		Name:        s.tool.Name(),
		Description: s.tool.Description(),
		Version:     s.tool.Version(),
		Tags:        s.tool.Tags(),
		InputSchema: &proto.JSONSchema{
			Json: string(inputSchemaJSON),
		},
		OutputSchema: &proto.JSONSchema{
			Json: string(outputSchemaJSON),
		},
	}, nil
}

// Execute runs the tool with the provided input.
// The input is serialized as JSON in the request and the output is
// serialized as JSON in the response.
func (s *toolServiceServer) Execute(ctx context.Context, req *proto.ToolExecuteRequest) (*proto.ToolExecuteResponse, error) {
	// Parse input from JSON
	var input map[string]any
	if err := json.Unmarshal([]byte(req.InputJson), &input); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input JSON: %v", err)
	}

	// Apply timeout if specified
	if req.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	// Execute the tool
	output, err := s.tool.Execute(ctx, input)

	// Build response
	resp := &proto.ToolExecuteResponse{}

	// Serialize output as JSON
	if err == nil {
		outputJSON, jsonErr := json.Marshal(output)
		if jsonErr != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal output: %v", jsonErr)
		}
		resp.OutputJson = string(outputJSON)
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

// Health returns the current health status of the tool.
func (s *toolServiceServer) Health(ctx context.Context, req *proto.ToolHealthRequest) (*proto.HealthStatus, error) {
	health := s.tool.Health(ctx)

	return &proto.HealthStatus{
		State:     health.Status,
		Message:   health.Message,
		CheckedAt: time.Now().UnixMilli(),
	}, nil
}

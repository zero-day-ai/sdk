package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Plugin is the interface that SDK plugins must implement.
// This is a placeholder until the plugin package is created.
// TODO: Import from github.com/zero-day-ai/sdk/plugin once that package exists.
type Plugin interface {
	// Initialize initializes the plugin with the provided configuration.
	Initialize(ctx context.Context, config map[string]any) error

	// Shutdown gracefully shuts down the plugin.
	Shutdown(ctx context.Context) error

	// ListMethods returns the available plugin methods.
	ListMethods(ctx context.Context) ([]MethodDescriptor, error)

	// Query executes a plugin method with the provided parameters.
	Query(ctx context.Context, method string, params map[string]any) (any, error)

	// Health returns the current health status.
	Health(ctx context.Context) HealthStatus
}

// MethodDescriptor describes a plugin method.
// This is a placeholder until the plugin package is created.
type MethodDescriptor struct {
	Name         string
	Description  string
	InputSchema  map[string]any
	OutputSchema map[string]any
}

// HealthStatus represents the health state of a component.
// This is a placeholder until we can import from types.
type HealthStatus struct {
	Status  string
	Message string
	Details map[string]any
}

// PluginFunc starts a gRPC server for a plugin.
// It creates a server, registers the plugin service, and serves requests
// until a shutdown signal is received or an error occurs.
//
// Example:
//
//	plugin := &MyPlugin{}
//	err := serve.Plugin(plugin,
//	    serve.WithPort(50053),
//	    serve.WithGracefulShutdown(30*time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
func PluginFunc(p Plugin, opts ...Option) error {
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

	// Create and register plugin service
	pluginSvc := &pluginServiceServer{
		plugin: p,
	}
	proto.RegisterPluginServiceServer(srv.GRPCServer(), pluginSvc)

	// Set health status to serving
	srv.HealthServer().SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	fmt.Printf("Plugin listening on :%d\n", srv.Port())

	// Start serving
	return srv.Serve(context.Background())
}

// pluginServiceServer implements the gRPC PluginService for an SDK plugin.
// It bridges the gRPC protocol to the Plugin interface.
type pluginServiceServer struct {
	proto.UnimplementedPluginServiceServer
	plugin Plugin
}

// Initialize initializes the plugin with the provided configuration.
func (s *pluginServiceServer) Initialize(ctx context.Context, req *proto.PluginInitializeRequest) (*proto.PluginInitializeResponse, error) {
	// Parse config from JSON
	var config map[string]any
	if req.ConfigJson != "" {
		if err := json.Unmarshal([]byte(req.ConfigJson), &config); err != nil {
			return &proto.PluginInitializeResponse{
				Error: &proto.Error{
					Code:      "INVALID_CONFIG",
					Message:   fmt.Sprintf("invalid config JSON: %v", err),
					Retryable: false,
				},
			}, nil
		}
	}

	// Initialize the plugin
	err := s.plugin.Initialize(ctx, config)

	// Build response
	resp := &proto.PluginInitializeResponse{}
	if err != nil {
		resp.Error = &proto.Error{
			Code:      "INITIALIZATION_ERROR",
			Message:   err.Error(),
			Retryable: false,
		}
	}

	return resp, nil
}

// Shutdown gracefully shuts down the plugin.
func (s *pluginServiceServer) Shutdown(ctx context.Context, req *proto.PluginShutdownRequest) (*proto.PluginShutdownResponse, error) {
	err := s.plugin.Shutdown(ctx)

	// Build response
	resp := &proto.PluginShutdownResponse{}
	if err != nil {
		resp.Error = &proto.Error{
			Code:      "SHUTDOWN_ERROR",
			Message:   err.Error(),
			Retryable: false,
		}
	}

	return resp, nil
}

// ListMethods returns the available plugin methods.
func (s *pluginServiceServer) ListMethods(ctx context.Context, req *proto.PluginListMethodsRequest) (*proto.PluginListMethodsResponse, error) {
	methods, err := s.plugin.ListMethods(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list methods: %v", err)
	}

	// Convert to proto methods
	protoMethods := make([]*proto.PluginMethodDescriptor, len(methods))
	for i, method := range methods {
		// Serialize schemas to JSON
		inputSchemaJSON, err := json.Marshal(method.InputSchema)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal input schema: %v", err)
		}

		outputSchemaJSON, err := json.Marshal(method.OutputSchema)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to marshal output schema: %v", err)
		}

		protoMethods[i] = &proto.PluginMethodDescriptor{
			Name:        method.Name,
			Description: method.Description,
			InputSchema: &proto.JSONSchema{
				Json: string(inputSchemaJSON),
			},
			OutputSchema: &proto.JSONSchema{
				Json: string(outputSchemaJSON),
			},
		}
	}

	return &proto.PluginListMethodsResponse{
		Methods: protoMethods,
	}, nil
}

// Query executes a plugin method with the provided parameters.
func (s *pluginServiceServer) Query(ctx context.Context, req *proto.PluginQueryRequest) (*proto.PluginQueryResponse, error) {
	// Parse params from JSON
	var params map[string]any
	if req.ParamsJson != "" {
		if err := json.Unmarshal([]byte(req.ParamsJson), &params); err != nil {
			return &proto.PluginQueryResponse{
				Error: &proto.Error{
					Code:      "INVALID_PARAMS",
					Message:   fmt.Sprintf("invalid params JSON: %v", err),
					Retryable: false,
				},
			}, nil
		}
	}

	// Apply timeout if specified
	if req.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	// Execute the query
	result, err := s.plugin.Query(ctx, req.Method, params)

	// Build response
	resp := &proto.PluginQueryResponse{}

	// Serialize result as JSON
	if err == nil {
		resultJSON, jsonErr := json.Marshal(result)
		if jsonErr != nil {
			return &proto.PluginQueryResponse{
				Error: &proto.Error{
					Code:      "SERIALIZATION_ERROR",
					Message:   fmt.Sprintf("failed to marshal result: %v", jsonErr),
					Retryable: false,
				},
			}, nil
		}
		resp.ResultJson = string(resultJSON)
	} else {
		// Map error to proto error
		resp.Error = &proto.Error{
			Code:      "QUERY_ERROR",
			Message:   err.Error(),
			Retryable: false,
		}
	}

	return resp, nil
}

// Health returns the current health status of the plugin.
func (s *pluginServiceServer) Health(ctx context.Context, req *proto.PluginHealthRequest) (*proto.HealthStatus, error) {
	health := s.plugin.Health(ctx)

	return &proto.HealthStatus{
		State:     health.Status,
		Message:   health.Message,
		CheckedAt: time.Now().UnixMilli(),
	}, nil
}

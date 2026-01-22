package serve

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/tool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// SubprocessModeEnvVar is the environment variable that indicates subprocess execution mode.
// When set to "subprocess", the tool will run in subprocess mode instead of starting a gRPC server.
const SubprocessModeEnvVar = "GIBSON_TOOL_MODE"

// SubprocessModeValue is the value of GIBSON_TOOL_MODE that triggers subprocess mode.
const SubprocessModeValue = "subprocess"

// SchemaFlag is the command-line flag used to request schema output.
const SchemaFlag = "--schema"

// Tool serves a tool implementation.
//
// The execution mode is automatically detected based on environment and command-line flags:
//
// 1. Schema Mode: If --schema flag is passed, outputs tool schema to stdout and exits.
// 2. Subprocess Mode: If GIBSON_TOOL_MODE=subprocess, runs in subprocess mode (stdin/stdout JSON).
// 3. gRPC Server Mode: Otherwise, starts a gRPC server for traditional RPC communication.
//
// For subprocess mode, input is read as JSON from stdin and output is written as JSON to stdout.
// Errors are written to stderr and result in a non-zero exit code.
//
// Example (gRPC mode):
//
//	tool := &MyTool{}
//	err := serve.Tool(tool,
//	    serve.WithPort(50052),
//	    serve.WithGracefulShutdown(30*time.Second),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Example (subprocess mode - automatic):
//
//	// When invoked with GIBSON_TOOL_MODE=subprocess:
//	// echo '{"target": "localhost"}' | GIBSON_TOOL_MODE=subprocess ./mytool
//	// Or with --schema flag:
//	// ./mytool --schema
func Tool(t tool.Tool, opts ...Option) error {
	// Check for --schema flag first (regardless of mode)
	if hasSchemaFlag() {
		return OutputSchema(t)
	}

	// Check for subprocess mode via environment variable
	if isSubprocessMode() {
		return RunSubprocess(t)
	}

	// Default: gRPC server mode
	return serveToolGRPC(t, opts...)
}

// hasSchemaFlag checks if --schema was passed as a command-line argument.
func hasSchemaFlag() bool {
	for _, arg := range os.Args[1:] {
		if arg == SchemaFlag {
			return true
		}
	}
	return false
}

// isSubprocessMode checks if GIBSON_TOOL_MODE=subprocess is set.
func isSubprocessMode() bool {
	return os.Getenv(SubprocessModeEnvVar) == SubprocessModeValue
}

// serveToolGRPC starts the tool as a gRPC server.
// This is the traditional mode where the tool listens on a port and handles RPC requests.
func serveToolGRPC(t tool.Tool, opts ...Option) error {
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

	slog.Info("tool server started", "component", "tool", "name", t.Name(), "version", t.Version(), "port", srv.Port())

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

		// Extract tool metadata
		metadata := map[string]string{
			"description": t.Description(),
		}

		// Add tags if available
		if len(t.Tags()) > 0 {
			metadata["tags"] = strings.Join(t.Tags(), ",")
		}

		// Add proto message types
		metadata["input_message_type"] = t.InputMessageType()
		metadata["output_message_type"] = t.OutputMessageType()

		// Create ServiceInfo struct (using map to avoid circular dependency)
		serviceInfo = map[string]interface{}{
			"kind":        "tool",
			"name":        t.Name(),
			"version":     t.Version(),
			"instance_id": uuid.New().String(),
			"endpoint":    endpoint,
			"metadata":    metadata,
			"started_at":  time.Now(),
		}

		// Register with the registry
		ctx := context.Background()
		if err := cfg.Registry.Register(ctx, serviceInfo); err != nil {
			slog.Warn("failed to register with registry", "error", err, "endpoint", endpoint, "component", "tool", "name", t.Name())
		} else {
			slog.Info("registered with registry", "endpoint", endpoint, "component", "tool", "name", t.Name())
			// Deregister on shutdown
			defer func() {
				ctx := context.Background()
				if err := cfg.Registry.Deregister(ctx, serviceInfo); err != nil {
					slog.Warn("failed to deregister from registry", "error", err, "endpoint", endpoint, "component", "tool", "name", t.Name())
				}
			}()
		}
	}

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
// description, and tags. Input/output schemas are left empty as tools now use proto messages.
func (s *toolServiceServer) GetDescriptor(ctx context.Context, req *proto.ToolGetDescriptorRequest) (*proto.ToolDescriptor, error) {
	return &proto.ToolDescriptor{
		Name:        s.tool.Name(),
		Description: s.tool.Description(),
		Version:     s.tool.Version(),
		Tags:        s.tool.Tags(),
		// InputSchema and OutputSchema are deprecated - tools use proto messages now
		// Clients should use InputMessageType() and OutputMessageType() instead
		InputSchema:  &proto.JSONSchema{Json: "{}"},
		OutputSchema: &proto.JSONSchema{Json: "{}"},
	}, nil
}

// Execute runs the tool with the provided input.
// The input is serialized as JSON in the request and the output is
// serialized as JSON in the response.
func (s *toolServiceServer) Execute(ctx context.Context, req *proto.ToolExecuteRequest) (*proto.ToolExecuteResponse, error) {
	// Apply timeout if specified
	if req.TimeoutMs > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
		defer cancel()
	}

	// Get the tool's input message type
	inputTypeName := s.tool.InputMessageType()
	if inputTypeName == "" {
		return nil, status.Errorf(codes.Unimplemented, "tool does not specify InputMessageType")
	}

	// Find the proto message type in the global registry
	messageType, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(inputTypeName))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find message type %q: %v", inputTypeName, err)
	}

	// Create a new instance of the proto message
	protoReq := messageType.New().Interface()

	// Unmarshal JSON input into the proto message
	if err := protojson.Unmarshal([]byte(req.InputJson), protoReq); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid input JSON for type %s: %v", inputTypeName, err)
	}

	// Execute the tool using ExecuteProto
	protoResp, err := s.tool.ExecuteProto(ctx, protoReq)

	// Build response
	resp := &proto.ToolExecuteResponse{}

	// Handle execution result
	if err == nil {
		// Marshal proto response to JSON
		outputJSON, jsonErr := protojson.Marshal(protoResp)
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

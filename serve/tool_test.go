package serve

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/tool"
	"github.com/zero-day-ai/sdk/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// mockTool is a mock implementation of tool.Tool for testing.
type mockTool struct {
	name        string
	version     string
	description string
	tags        []string
	inputSchema schema.JSON
	health      types.HealthStatus
	executeFunc func(ctx context.Context, input map[string]any) (map[string]any, error)
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Version() string     { return m.version }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) Tags() []string      { return m.tags }

func (m *mockTool) InputSchema() schema.JSON {
	if m.inputSchema.Type != "" {
		return m.inputSchema
	}
	return schema.Object(map[string]schema.JSON{
		"message": schema.String(),
	})
}

func (m *mockTool) OutputSchema() schema.JSON {
	return schema.Object(map[string]schema.JSON{
		"result": schema.String(),
	})
}

func (m *mockTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return map[string]any{"result": "success"}, nil
}

func (m *mockTool) Health(ctx context.Context) types.HealthStatus {
	if m.health.Status == "" {
		return types.NewHealthyStatus("Tool is healthy")
	}
	return m.health
}

// setupToolTestServer creates an in-memory gRPC server for testing using bufconn.
func setupToolTestServer(t *testing.T, tool tool.Tool) (*grpc.ClientConn, func()) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	toolSvc := &toolServiceServer{tool: tool}
	proto.RegisterToolServiceServer(srv, toolSvc)

	// Start server
	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	// Create client connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		conn.Close()
		srv.Stop()
		lis.Close()
	}

	return conn, cleanup
}

func TestToolServiceServer_GetDescriptor(t *testing.T) {
	mockT := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "Test tool for unit testing",
		tags:        []string{"test", "mock"},
	}

	conn, cleanup := setupToolTestServer(t, mockT)
	defer cleanup()

	client := proto.NewToolServiceClient(conn)
	resp, err := client.GetDescriptor(context.Background(), &proto.ToolGetDescriptorRequest{})

	require.NoError(t, err)
	assert.Equal(t, "test-tool", resp.Name)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.Equal(t, "Test tool for unit testing", resp.Description)
	assert.Equal(t, []string{"test", "mock"}, resp.Tags)
	assert.NotNil(t, resp.InputSchema)
	assert.NotEmpty(t, resp.InputSchema.Json)
	assert.NotNil(t, resp.OutputSchema)
	assert.NotEmpty(t, resp.OutputSchema.Json)

	// Verify schema is valid JSON
	var inputSchema map[string]any
	err = json.Unmarshal([]byte(resp.InputSchema.Json), &inputSchema)
	require.NoError(t, err)
	assert.Equal(t, "object", inputSchema["type"])

	var outputSchema map[string]any
	err = json.Unmarshal([]byte(resp.OutputSchema.Json), &outputSchema)
	require.NoError(t, err)
	assert.Equal(t, "object", outputSchema["type"])
}

func TestToolServiceServer_Execute(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		executeFunc func(ctx context.Context, input map[string]any) (map[string]any, error)
		wantErr     bool
		checkResult func(t *testing.T, resp *proto.ToolExecuteResponse)
	}{
		{
			name: "successful execution",
			input: map[string]any{
				"message": "hello",
			},
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return map[string]any{
					"result":  "processed",
					"message": input["message"],
				}, nil
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *proto.ToolExecuteResponse) {
				assert.Nil(t, resp.Error)
				assert.NotEmpty(t, resp.OutputJson)

				var output map[string]any
				err := json.Unmarshal([]byte(resp.OutputJson), &output)
				require.NoError(t, err)
				assert.Equal(t, "processed", output["result"])
				assert.Equal(t, "hello", output["message"])
			},
		},
		{
			name: "execution with error",
			input: map[string]any{
				"message": "fail",
			},
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, assert.AnError
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *proto.ToolExecuteResponse) {
				assert.NotNil(t, resp.Error)
				assert.Equal(t, "EXECUTION_ERROR", resp.Error.Code)
			},
		},
		{
			name: "empty input",
			input: map[string]any{},
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return map[string]any{"result": "ok"}, nil
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *proto.ToolExecuteResponse) {
				assert.Nil(t, resp.Error)
				assert.NotEmpty(t, resp.OutputJson)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTool{
				name:        "test-tool",
				version:     "1.0.0",
				executeFunc: tt.executeFunc,
			}

			conn, cleanup := setupToolTestServer(t, mockT)
			defer cleanup()

			client := proto.NewToolServiceClient(conn)

			inputJSON, err := json.Marshal(tt.input)
			require.NoError(t, err)

			resp, err := client.Execute(context.Background(), &proto.ToolExecuteRequest{
				InputJson: string(inputJSON),
			})

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, resp)
			}
		})
	}
}

func TestToolServiceServer_Execute_InvalidJSON(t *testing.T) {
	mockT := &mockTool{
		name:    "test-tool",
		version: "1.0.0",
	}

	conn, cleanup := setupToolTestServer(t, mockT)
	defer cleanup()

	client := proto.NewToolServiceClient(conn)

	_, err := client.Execute(context.Background(), &proto.ToolExecuteRequest{
		InputJson: "invalid json",
	})

	assert.Error(t, err)
}

func TestToolServiceServer_Execute_WithTimeout(t *testing.T) {
	mockT := &mockTool{
		name:    "test-tool",
		version: "1.0.0",
		executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
			// Check that context has timeout
			_, hasDeadline := ctx.Deadline()
			return map[string]any{"has_deadline": hasDeadline}, nil
		},
	}

	conn, cleanup := setupToolTestServer(t, mockT)
	defer cleanup()

	client := proto.NewToolServiceClient(conn)

	inputJSON, err := json.Marshal(map[string]any{"test": "value"})
	require.NoError(t, err)

	resp, err := client.Execute(context.Background(), &proto.ToolExecuteRequest{
		InputJson:  string(inputJSON),
		TimeoutMs: 5000, // 5 second timeout
	})

	require.NoError(t, err)
	assert.Nil(t, resp.Error)

	var output map[string]any
	err = json.Unmarshal([]byte(resp.OutputJson), &output)
	require.NoError(t, err)
	assert.True(t, output["has_deadline"].(bool))
}

func TestToolServiceServer_Health(t *testing.T) {
	tests := []struct {
		name         string
		health       types.HealthStatus
		expectStatus string
	}{
		{
			name:         "healthy tool",
			health:       types.NewHealthyStatus("Tool is operational"),
			expectStatus: types.StatusHealthy,
		},
		{
			name:         "degraded tool",
			health:       types.NewDegradedStatus("Performance degraded", nil),
			expectStatus: types.StatusDegraded,
		},
		{
			name:         "unhealthy tool",
			health:       types.NewUnhealthyStatus("Tool unavailable", nil),
			expectStatus: types.StatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTool{
				name:    "test-tool",
				version: "1.0.0",
				health:  tt.health,
			}

			conn, cleanup := setupToolTestServer(t, mockT)
			defer cleanup()

			client := proto.NewToolServiceClient(conn)
			resp, err := client.Health(context.Background(), &proto.ToolHealthRequest{})

			require.NoError(t, err)
			assert.Equal(t, tt.expectStatus, resp.State)
			assert.NotEmpty(t, resp.Message)
			assert.Greater(t, resp.CheckedAt, int64(0))
		})
	}
}

// Tests for subprocess mode detection

func TestHasSchemaFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "with --schema flag",
			args:     []string{"tool", "--schema"},
			expected: true,
		},
		{
			name:     "with --schema flag among other args",
			args:     []string{"tool", "--verbose", "--schema", "--debug"},
			expected: true,
		},
		{
			name:     "without --schema flag",
			args:     []string{"tool"},
			expected: false,
		},
		{
			name:     "with similar but different flag",
			args:     []string{"tool", "--schemas", "--schema-file=foo"},
			expected: false,
		},
		{
			name:     "empty args",
			args:     []string{"tool"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			originalArgs := os.Args
			defer func() { os.Args = originalArgs }()

			os.Args = tt.args
			result := hasSchemaFlag()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSubprocessMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "GIBSON_TOOL_MODE=subprocess",
			envValue: "subprocess",
			expected: true,
		},
		{
			name:     "GIBSON_TOOL_MODE=grpc",
			envValue: "grpc",
			expected: false,
		},
		{
			name:     "GIBSON_TOOL_MODE not set",
			envValue: "",
			expected: false,
		},
		{
			name:     "GIBSON_TOOL_MODE=SUBPROCESS (case sensitive)",
			envValue: "SUBPROCESS",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env
			originalValue := os.Getenv(SubprocessModeEnvVar)
			defer func() {
				if originalValue == "" {
					os.Unsetenv(SubprocessModeEnvVar)
				} else {
					os.Setenv(SubprocessModeEnvVar, originalValue)
				}
			}()

			if tt.envValue == "" {
				os.Unsetenv(SubprocessModeEnvVar)
			} else {
				os.Setenv(SubprocessModeEnvVar, tt.envValue)
			}

			result := isSubprocessMode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToolConstants(t *testing.T) {
	// Verify constants match expected values
	assert.Equal(t, "GIBSON_TOOL_MODE", SubprocessModeEnvVar)
	assert.Equal(t, "subprocess", SubprocessModeValue)
	assert.Equal(t, "--schema", SchemaFlag)
}

package serve

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/enum"
	"github.com/zero-day-ai/sdk/tool"
	"github.com/zero-day-ai/sdk/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	protolib "google.golang.org/protobuf/proto"
)

// mockTool is a mock implementation of tool.Tool for testing.
type mockTool struct {
	name             string
	version          string
	description      string
	tags             []string
	health           types.HealthStatus
	executeProtoFunc func(ctx context.Context, input protolib.Message) (protolib.Message, error)
}

func (m *mockTool) Name() string        { return m.name }
func (m *mockTool) Version() string     { return m.version }
func (m *mockTool) Description() string { return m.description }
func (m *mockTool) Tags() []string      { return m.tags }

func (m *mockTool) InputMessageType() string {
	return "gibson.common.TypedMap"
}

func (m *mockTool) OutputMessageType() string {
	return "gibson.common.TypedMap"
}

func (m *mockTool) ExecuteProto(ctx context.Context, input protolib.Message) (protolib.Message, error) {
	if m.executeProtoFunc != nil {
		return m.executeProtoFunc(ctx, input)
	}
	// Default implementation: echo back the input with a status field added
	inputMap, ok := input.(*proto.TypedMap)
	if !ok {
		return nil, errors.New("invalid input type")
	}
	// Create output map with input entries plus a result field
	outputMap := &proto.TypedMap{
		Entries: make(map[string]*proto.TypedValue),
	}
	// Copy input entries
	for k, v := range inputMap.Entries {
		outputMap.Entries[k] = v
	}
	// Add result field
	outputMap.Entries["result"] = &proto.TypedValue{
		Kind: &proto.TypedValue_StringValue{StringValue: "success"},
	}
	return outputMap, nil
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

	// Verify schemas are present but deprecated (empty objects)
	assert.NotNil(t, resp.InputSchema)
	assert.Equal(t, "{}", resp.InputSchema.Json, "InputSchema should be empty JSON object (deprecated)")
	assert.NotNil(t, resp.OutputSchema)
	assert.Equal(t, "{}", resp.OutputSchema.Json, "OutputSchema should be empty JSON object (deprecated)")
}

func TestToolServiceServer_Execute(t *testing.T) {
	tests := []struct {
		name             string
		input            map[string]any
		executeProtoFunc func(ctx context.Context, input protolib.Message) (protolib.Message, error)
		wantErr          bool
		checkResult      func(t *testing.T, resp *proto.ToolExecuteResponse)
	}{
		{
			name: "successful execution",
			input: map[string]any{
				"entries": map[string]any{
					"message": map[string]any{
						"stringValue": "hello",
					},
				},
			},
			executeProtoFunc: func(ctx context.Context, input protolib.Message) (protolib.Message, error) {
				// Verify input is a TypedMap
				inputMap, ok := input.(*proto.TypedMap)
				require.True(t, ok, "input should be TypedMap")

				// Verify message field exists
				messageVal, exists := inputMap.Entries["message"]
				require.True(t, exists, "message field should exist")
				assert.Equal(t, "hello", messageVal.GetStringValue())

				// Create response
				return &proto.TypedMap{
					Entries: map[string]*proto.TypedValue{
						"result": {
							Kind: &proto.TypedValue_StringValue{StringValue: "processed"},
						},
						"message": messageVal,
					},
				}, nil
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *proto.ToolExecuteResponse) {
				assert.Nil(t, resp.Error)
				assert.NotEmpty(t, resp.OutputJson)

				var output map[string]any
				err := json.Unmarshal([]byte(resp.OutputJson), &output)
				require.NoError(t, err)

				// Output is in TypedMap format
				entries, ok := output["entries"].(map[string]any)
				require.True(t, ok, "output should have entries field")

				resultField, ok := entries["result"].(map[string]any)
				require.True(t, ok, "entries should have result field")
				assert.Equal(t, "processed", resultField["stringValue"])

				messageField, ok := entries["message"].(map[string]any)
				require.True(t, ok, "entries should have message field")
				assert.Equal(t, "hello", messageField["stringValue"])
			},
		},
		{
			name: "execution with error",
			input: map[string]any{
				"entries": map[string]any{
					"message": map[string]any{
						"stringValue": "fail",
					},
				},
			},
			executeProtoFunc: func(ctx context.Context, input protolib.Message) (protolib.Message, error) {
				// Verify ExecuteProto is called with correct type
				_, ok := input.(*proto.TypedMap)
				require.True(t, ok, "input should be TypedMap")
				return nil, assert.AnError
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *proto.ToolExecuteResponse) {
				assert.NotNil(t, resp.Error)
				assert.Equal(t, "EXECUTION_ERROR", resp.Error.Code)
			},
		},
		{
			name:  "empty input",
			input: map[string]any{},
			executeProtoFunc: func(ctx context.Context, input protolib.Message) (protolib.Message, error) {
				// Verify ExecuteProto is called
				inputMap, ok := input.(*proto.TypedMap)
				require.True(t, ok, "input should be TypedMap")
				assert.Empty(t, inputMap.Entries)

				return &proto.TypedMap{
					Entries: map[string]*proto.TypedValue{
						"result": {
							Kind: &proto.TypedValue_StringValue{StringValue: "ok"},
						},
					},
				}, nil
			},
			wantErr: false,
			checkResult: func(t *testing.T, resp *proto.ToolExecuteResponse) {
				assert.Nil(t, resp.Error)
				assert.NotEmpty(t, resp.OutputJson)

				var output map[string]any
				err := json.Unmarshal([]byte(resp.OutputJson), &output)
				require.NoError(t, err)

				// Output is in TypedMap format
				entries, ok := output["entries"].(map[string]any)
				require.True(t, ok, "output should have entries field")

				resultField, ok := entries["result"].(map[string]any)
				require.True(t, ok, "entries should have result field")
				assert.Equal(t, "ok", resultField["stringValue"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockT := &mockTool{
				name:             "test-tool",
				version:          "1.0.0",
				executeProtoFunc: tt.executeProtoFunc,
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
		executeProtoFunc: func(ctx context.Context, input protolib.Message) (protolib.Message, error) {
			// Verify ExecuteProto is called with TypedMap
			_, ok := input.(*proto.TypedMap)
			require.True(t, ok, "input should be TypedMap")

			// Check that context has timeout
			_, hasDeadline := ctx.Deadline()

			return &proto.TypedMap{
				Entries: map[string]*proto.TypedValue{
					"has_deadline": {
						Kind: &proto.TypedValue_BoolValue{BoolValue: hasDeadline},
					},
				},
			}, nil
		},
	}

	conn, cleanup := setupToolTestServer(t, mockT)
	defer cleanup()

	client := proto.NewToolServiceClient(conn)

	inputJSON, err := json.Marshal(map[string]any{
		"entries": map[string]any{
			"test": map[string]any{
				"stringValue": "value",
			},
		},
	})
	require.NoError(t, err)

	resp, err := client.Execute(context.Background(), &proto.ToolExecuteRequest{
		InputJson: string(inputJSON),
		TimeoutMs: 5000, // 5 second timeout
	})

	require.NoError(t, err)
	assert.Nil(t, resp.Error)

	var output map[string]any
	err = json.Unmarshal([]byte(resp.OutputJson), &output)
	require.NoError(t, err)

	// Output is in TypedMap format
	entries, ok := output["entries"].(map[string]any)
	require.True(t, ok, "output should have entries field")

	hasDeadlineField, ok := entries["has_deadline"].(map[string]any)
	require.True(t, ok, "entries should have has_deadline field")
	assert.True(t, hasDeadlineField["boolValue"].(bool))
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

// TestToolExecute_WithEnumNormalization verifies that tools with registered enum mappings
// receive normalized values according to the enum registry.
func TestToolExecute_WithEnumNormalization(t *testing.T) {
	// Clear registry before and after test for isolation
	enum.Clear()
	defer enum.Clear()

	// Register enum mappings for a test tool
	// Use camelCase field names to match proto conventions
	enum.Register("test-tool-enum", "scanType", map[string]string{
		"syn":     "SYN_SCAN",
		"connect": "CONNECT_SCAN",
		"udp":     "UDP_SCAN",
	})
	enum.Register("test-tool-enum", "verbosity", map[string]string{
		"low":    "VERBOSITY_LOW",
		"medium": "VERBOSITY_MEDIUM",
		"high":   "VERBOSITY_HIGH",
	})

	// Create a mock tool that captures the received input
	var receivedInput *proto.TypedMap
	mockT := &mockTool{
		name:    "test-tool-enum",
		version: "1.0.0",
		executeProtoFunc: func(ctx context.Context, input protolib.Message) (protolib.Message, error) {
			// Capture input to verify normalization
			inputMap, ok := input.(*proto.TypedMap)
			require.True(t, ok, "input should be TypedMap")
			receivedInput = inputMap

			// Return success response
			return &proto.TypedMap{
				Entries: map[string]*proto.TypedValue{
					"result": {
						Kind: &proto.TypedValue_StringValue{StringValue: "normalized"},
					},
				},
			}, nil
		},
	}

	conn, cleanup := setupToolTestServer(t, mockT)
	defer cleanup()

	client := proto.NewToolServiceClient(conn)

	// Send request with shorthand enum values
	inputJSON, err := json.Marshal(map[string]any{
		"entries": map[string]any{
			"scanType": map[string]any{
				"stringValue": "syn", // Should be normalized to "SYN_SCAN"
			},
			"verbosity": map[string]any{
				"stringValue": "high", // Should be normalized to "VERBOSITY_HIGH"
			},
			"target": map[string]any{
				"stringValue": "localhost",
			},
		},
	})
	require.NoError(t, err)

	resp, err := client.Execute(context.Background(), &proto.ToolExecuteRequest{
		InputJson: string(inputJSON),
	})

	require.NoError(t, err)
	assert.Nil(t, resp.Error)

	// Verify that the tool received normalized enum values
	require.NotNil(t, receivedInput, "tool should have received input")
	assert.NotNil(t, receivedInput.Entries["scanType"])
	assert.Equal(t, "SYN_SCAN", receivedInput.Entries["scanType"].GetStringValue(),
		"scanType should be normalized to SYN_SCAN")
	assert.NotNil(t, receivedInput.Entries["verbosity"])
	assert.Equal(t, "VERBOSITY_HIGH", receivedInput.Entries["verbosity"].GetStringValue(),
		"verbosity should be normalized to VERBOSITY_HIGH")

	// Verify non-enum fields pass through unchanged
	assert.NotNil(t, receivedInput.Entries["target"])
	assert.Equal(t, "localhost", receivedInput.Entries["target"].GetStringValue(),
		"non-enum field should pass through unchanged")
}

// TestToolExecute_WithoutEnumMappings verifies that tools without enum mappings
// receive input values unchanged (pass-through behavior).
func TestToolExecute_WithoutEnumMappings(t *testing.T) {
	// Clear registry before and after test for isolation
	enum.Clear()
	defer enum.Clear()

	// No enum mappings registered for this tool

	// Create a mock tool that captures the received input
	var receivedInput *proto.TypedMap
	mockT := &mockTool{
		name:    "test-tool-no-enum",
		version: "1.0.0",
		executeProtoFunc: func(ctx context.Context, input protolib.Message) (protolib.Message, error) {
			// Capture input to verify it passes through unchanged
			inputMap, ok := input.(*proto.TypedMap)
			require.True(t, ok, "input should be TypedMap")
			receivedInput = inputMap

			// Return success response
			return &proto.TypedMap{
				Entries: map[string]*proto.TypedValue{
					"result": {
						Kind: &proto.TypedValue_StringValue{StringValue: "passthrough"},
					},
				},
			}, nil
		},
	}

	conn, cleanup := setupToolTestServer(t, mockT)
	defer cleanup()

	client := proto.NewToolServiceClient(conn)

	// Send request with shorthand values
	inputJSON, err := json.Marshal(map[string]any{
		"entries": map[string]any{
			"scanType": map[string]any{
				"stringValue": "syn", // Should pass through unchanged
			},
			"verbosity": map[string]any{
				"stringValue": "high", // Should pass through unchanged
			},
			"target": map[string]any{
				"stringValue": "localhost",
			},
		},
	})
	require.NoError(t, err)

	resp, err := client.Execute(context.Background(), &proto.ToolExecuteRequest{
		InputJson: string(inputJSON),
	})

	require.NoError(t, err)
	assert.Nil(t, resp.Error)

	// Verify that the tool received original values unchanged
	require.NotNil(t, receivedInput, "tool should have received input")
	assert.NotNil(t, receivedInput.Entries["scanType"])
	assert.Equal(t, "syn", receivedInput.Entries["scanType"].GetStringValue(),
		"scanType should pass through unchanged")
	assert.NotNil(t, receivedInput.Entries["verbosity"])
	assert.Equal(t, "high", receivedInput.Entries["verbosity"].GetStringValue(),
		"verbosity should pass through unchanged")
	assert.NotNil(t, receivedInput.Entries["target"])
	assert.Equal(t, "localhost", receivedInput.Entries["target"].GetStringValue(),
		"target should pass through unchanged")
}

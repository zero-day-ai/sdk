package serve

import (
	"context"
	"encoding/json"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/gibson/api/gen/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// mockPlugin is a mock implementation of Plugin for testing.
type mockPlugin struct {
	initFunc    func(ctx context.Context, config map[string]any) error
	shutdownErr error
	methods     []MethodDescriptor
	queryFunc   func(ctx context.Context, method string, params map[string]any) (any, error)
	health      HealthStatus
}

func (m *mockPlugin) Initialize(ctx context.Context, config map[string]any) error {
	if m.initFunc != nil {
		return m.initFunc(ctx, config)
	}
	return nil
}

func (m *mockPlugin) Shutdown(ctx context.Context) error {
	return m.shutdownErr
}

func (m *mockPlugin) ListMethods(ctx context.Context) ([]MethodDescriptor, error) {
	if m.methods != nil {
		return m.methods, nil
	}
	return []MethodDescriptor{
		{
			Name:        "test_method",
			Description: "Test method",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"param": map[string]any{"type": "string"},
				},
			},
			OutputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"result": map[string]any{"type": "string"},
				},
			},
		},
	}, nil
}

func (m *mockPlugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, method, params)
	}
	return map[string]any{"status": "ok"}, nil
}

func (m *mockPlugin) Health(ctx context.Context) HealthStatus {
	if m.health.Status == "" {
		return HealthStatus{
			Status:  "healthy",
			Message: "Plugin is healthy",
		}
	}
	return m.health
}

// setupPluginTestServer creates an in-memory gRPC server for testing using bufconn.
func setupPluginTestServer(t *testing.T, p Plugin) (*grpc.ClientConn, func()) {
	const bufSize = 1024 * 1024
	lis := bufconn.Listen(bufSize)

	srv := grpc.NewServer()
	pluginSvc := &pluginServiceServer{plugin: p}
	proto.RegisterPluginServiceServer(srv, pluginSvc)

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

func TestPluginServiceServer_Initialize(t *testing.T) {
	tests := []struct {
		name       string
		configJSON string
		initFunc   func(ctx context.Context, config map[string]any) error
		wantError  bool
	}{
		{
			name:       "successful initialization with config",
			configJSON: `{"setting": "value"}`,
			initFunc: func(ctx context.Context, config map[string]any) error {
				assert.Equal(t, "value", config["setting"])
				return nil
			},
			wantError: false,
		},
		{
			name:       "successful initialization without config",
			configJSON: "",
			initFunc:   func(ctx context.Context, config map[string]any) error { return nil },
			wantError:  false,
		},
		{
			name:       "initialization error",
			configJSON: `{"setting": "value"}`,
			initFunc:   func(ctx context.Context, config map[string]any) error { return assert.AnError },
			wantError:  true,
		},
		{
			name:       "invalid config JSON",
			configJSON: "invalid json",
			initFunc:   func(ctx context.Context, config map[string]any) error { return nil },
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockP := &mockPlugin{
				initFunc: tt.initFunc,
			}

			conn, cleanup := setupPluginTestServer(t, mockP)
			defer cleanup()

			client := proto.NewPluginServiceClient(conn)
			resp, err := client.Initialize(context.Background(), &proto.PluginInitializeRequest{
				ConfigJson: tt.configJSON,
			})

			require.NoError(t, err) // gRPC call itself should succeed

			if tt.wantError {
				assert.NotNil(t, resp.Error)
			} else {
				assert.Nil(t, resp.Error)
			}
		})
	}
}

func TestPluginServiceServer_Shutdown(t *testing.T) {
	tests := []struct {
		name        string
		shutdownErr error
		wantError   bool
	}{
		{
			name:        "successful shutdown",
			shutdownErr: nil,
			wantError:   false,
		},
		{
			name:        "shutdown error",
			shutdownErr: assert.AnError,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockP := &mockPlugin{
				shutdownErr: tt.shutdownErr,
			}

			conn, cleanup := setupPluginTestServer(t, mockP)
			defer cleanup()

			client := proto.NewPluginServiceClient(conn)
			resp, err := client.Shutdown(context.Background(), &proto.PluginShutdownRequest{})

			require.NoError(t, err) // gRPC call itself should succeed

			if tt.wantError {
				assert.NotNil(t, resp.Error)
			} else {
				assert.Nil(t, resp.Error)
			}
		})
	}
}

func TestPluginServiceServer_ListMethods(t *testing.T) {
	mockP := &mockPlugin{
		methods: []MethodDescriptor{
			{
				Name:        "method1",
				Description: "First method",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"input": map[string]any{"type": "string"},
					},
				},
				OutputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"output": map[string]any{"type": "string"},
					},
				},
			},
			{
				Name:        "method2",
				Description: "Second method",
				InputSchema: map[string]any{
					"type": "object",
				},
				OutputSchema: map[string]any{
					"type": "object",
				},
			},
		},
	}

	conn, cleanup := setupPluginTestServer(t, mockP)
	defer cleanup()

	client := proto.NewPluginServiceClient(conn)
	resp, err := client.ListMethods(context.Background(), &proto.PluginListMethodsRequest{})

	require.NoError(t, err)
	require.Len(t, resp.Methods, 2)

	// Check first method
	method1 := resp.Methods[0]
	assert.Equal(t, "method1", method1.Name)
	assert.Equal(t, "First method", method1.Description)
	assert.NotNil(t, method1.InputSchema)
	assert.NotEmpty(t, method1.InputSchema.Json)
	assert.NotNil(t, method1.OutputSchema)
	assert.NotEmpty(t, method1.OutputSchema.Json)

	// Verify schema is valid JSON
	var inputSchema map[string]any
	err = json.Unmarshal([]byte(method1.InputSchema.Json), &inputSchema)
	require.NoError(t, err)
	assert.Equal(t, "object", inputSchema["type"])

	// Check second method
	method2 := resp.Methods[1]
	assert.Equal(t, "method2", method2.Name)
	assert.Equal(t, "Second method", method2.Description)
}

func TestPluginServiceServer_Query(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		paramsJSON  string
		timeoutMs   int64
		queryFunc   func(ctx context.Context, method string, params map[string]any) (any, error)
		wantError   bool
		checkResult func(t *testing.T, resp *proto.PluginQueryResponse)
	}{
		{
			name:       "successful query",
			method:     "test_method",
			paramsJSON: `{"param": "value"}`,
			queryFunc: func(ctx context.Context, method string, params map[string]any) (any, error) {
				assert.Equal(t, "test_method", method)
				assert.Equal(t, "value", params["param"])
				return map[string]any{"result": "success"}, nil
			},
			wantError: false,
			checkResult: func(t *testing.T, resp *proto.PluginQueryResponse) {
				assert.Nil(t, resp.Error)
				assert.NotEmpty(t, resp.ResultJson)

				var result map[string]any
				err := json.Unmarshal([]byte(resp.ResultJson), &result)
				require.NoError(t, err)
				assert.Equal(t, "success", result["result"])
			},
		},
		{
			name:       "query with error",
			method:     "test_method",
			paramsJSON: `{"param": "value"}`,
			queryFunc: func(ctx context.Context, method string, params map[string]any) (any, error) {
				return nil, assert.AnError
			},
			wantError: false,
			checkResult: func(t *testing.T, resp *proto.PluginQueryResponse) {
				assert.NotNil(t, resp.Error)
				assert.Equal(t, "QUERY_ERROR", resp.Error.Code)
			},
		},
		{
			name:       "query without params",
			method:     "test_method",
			paramsJSON: "",
			queryFunc: func(ctx context.Context, method string, params map[string]any) (any, error) {
				return map[string]any{"result": "ok"}, nil
			},
			wantError: false,
			checkResult: func(t *testing.T, resp *proto.PluginQueryResponse) {
				assert.Nil(t, resp.Error)
			},
		},
		{
			name:       "invalid params JSON",
			method:     "test_method",
			paramsJSON: "invalid json",
			queryFunc: func(ctx context.Context, method string, params map[string]any) (any, error) {
				return nil, nil
			},
			wantError: false,
			checkResult: func(t *testing.T, resp *proto.PluginQueryResponse) {
				assert.NotNil(t, resp.Error)
				assert.Equal(t, "INVALID_PARAMS", resp.Error.Code)
			},
		},
		{
			name:       "query with timeout",
			method:     "test_method",
			paramsJSON: `{"param": "value"}`,
			timeoutMs:  5000,
			queryFunc: func(ctx context.Context, method string, params map[string]any) (any, error) {
				_, hasDeadline := ctx.Deadline()
				return map[string]any{"has_deadline": hasDeadline}, nil
			},
			wantError: false,
			checkResult: func(t *testing.T, resp *proto.PluginQueryResponse) {
				assert.Nil(t, resp.Error)

				var result map[string]any
				err := json.Unmarshal([]byte(resp.ResultJson), &result)
				require.NoError(t, err)
				assert.True(t, result["has_deadline"].(bool))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockP := &mockPlugin{
				queryFunc: tt.queryFunc,
			}

			conn, cleanup := setupPluginTestServer(t, mockP)
			defer cleanup()

			client := proto.NewPluginServiceClient(conn)
			resp, err := client.Query(context.Background(), &proto.PluginQueryRequest{
				Method:     tt.method,
				ParamsJson: tt.paramsJSON,
				TimeoutMs:  tt.timeoutMs,
			})

			if tt.wantError {
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

func TestPluginServiceServer_Health(t *testing.T) {
	tests := []struct {
		name         string
		health       HealthStatus
		expectStatus string
	}{
		{
			name: "healthy plugin",
			health: HealthStatus{
				Status:  "healthy",
				Message: "Plugin is operational",
			},
			expectStatus: "healthy",
		},
		{
			name: "degraded plugin",
			health: HealthStatus{
				Status:  "degraded",
				Message: "Performance issues",
			},
			expectStatus: "degraded",
		},
		{
			name: "unhealthy plugin",
			health: HealthStatus{
				Status:  "unhealthy",
				Message: "Plugin unavailable",
			},
			expectStatus: "unhealthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockP := &mockPlugin{
				health: tt.health,
			}

			conn, cleanup := setupPluginTestServer(t, mockP)
			defer cleanup()

			client := proto.NewPluginServiceClient(conn)
			resp, err := client.Health(context.Background(), &proto.PluginHealthRequest{})

			require.NoError(t, err)
			assert.Equal(t, tt.expectStatus, resp.State)
			assert.NotEmpty(t, resp.Message)
			assert.Greater(t, resp.CheckedAt, int64(0))
		})
	}
}

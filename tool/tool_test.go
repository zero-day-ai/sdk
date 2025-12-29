package tool

import (
	"context"
	"errors"
	"testing"

	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

// mockTool is a test implementation of the Tool interface.
type mockTool struct {
	name         string
	version      string
	description  string
	tags         []string
	inputSchema  schema.JSON
	outputSchema schema.JSON
	executeFunc  func(ctx context.Context, input map[string]any) (map[string]any, error)
	healthFunc   func(ctx context.Context) types.HealthStatus
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Version() string {
	return m.version
}

func (m *mockTool) Description() string {
	return m.description
}

func (m *mockTool) Tags() []string {
	return m.tags
}

func (m *mockTool) InputSchema() schema.JSON {
	return m.inputSchema
}

func (m *mockTool) OutputSchema() schema.JSON {
	return m.outputSchema
}

func (m *mockTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return nil, errors.New("execute not implemented")
}

func (m *mockTool) Health(ctx context.Context) types.HealthStatus {
	if m.healthFunc != nil {
		return m.healthFunc(ctx)
	}
	return types.NewHealthyStatus("mock tool healthy")
}

func TestMockTool_Name(t *testing.T) {
	tool := &mockTool{
		name: "test-tool",
	}

	if got := tool.Name(); got != "test-tool" {
		t.Errorf("Name() = %v, want %v", got, "test-tool")
	}
}

func TestMockTool_Version(t *testing.T) {
	tool := &mockTool{
		version: "1.0.0",
	}

	if got := tool.Version(); got != "1.0.0" {
		t.Errorf("Version() = %v, want %v", got, "1.0.0")
	}
}

func TestMockTool_Description(t *testing.T) {
	tool := &mockTool{
		description: "A test tool",
	}

	if got := tool.Description(); got != "A test tool" {
		t.Errorf("Description() = %v, want %v", got, "A test tool")
	}
}

func TestMockTool_Tags(t *testing.T) {
	want := []string{"test", "mock"}
	tool := &mockTool{
		tags: want,
	}

	got := tool.Tags()
	if len(got) != len(want) {
		t.Fatalf("Tags() length = %v, want %v", len(got), len(want))
	}

	for i, tag := range got {
		if tag != want[i] {
			t.Errorf("Tags()[%d] = %v, want %v", i, tag, want[i])
		}
	}
}

func TestMockTool_InputSchema(t *testing.T) {
	inputSchema := schema.Object(map[string]schema.JSON{
		"message": schema.String(),
	}, "message")

	tool := &mockTool{
		inputSchema: inputSchema,
	}

	got := tool.InputSchema()
	if got.Type != "object" {
		t.Errorf("InputSchema().Type = %v, want %v", got.Type, "object")
	}
}

func TestMockTool_OutputSchema(t *testing.T) {
	outputSchema := schema.Object(map[string]schema.JSON{
		"result": schema.String(),
	}, "result")

	tool := &mockTool{
		outputSchema: outputSchema,
	}

	got := tool.OutputSchema()
	if got.Type != "object" {
		t.Errorf("OutputSchema().Type = %v, want %v", got.Type, "object")
	}
}

func TestMockTool_Execute(t *testing.T) {
	tests := []struct {
		name        string
		executeFunc func(ctx context.Context, input map[string]any) (map[string]any, error)
		input       map[string]any
		wantOutput  map[string]any
		wantErr     bool
	}{
		{
			name: "successful execution",
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return map[string]any{"result": "success"}, nil
			},
			input:      map[string]any{"message": "hello"},
			wantOutput: map[string]any{"result": "success"},
			wantErr:    false,
		},
		{
			name: "execution with error",
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, errors.New("execution failed")
			},
			input:      map[string]any{"message": "hello"},
			wantOutput: nil,
			wantErr:    true,
		},
		{
			name:        "no execute function",
			executeFunc: nil,
			input:       map[string]any{"message": "hello"},
			wantOutput:  nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &mockTool{
				executeFunc: tt.executeFunc,
			}

			got, err := tool.Execute(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got == nil && tt.wantOutput != nil {
					t.Errorf("Execute() = nil, want %v", tt.wantOutput)
					return
				}
				if got != nil && tt.wantOutput != nil {
					if got["result"] != tt.wantOutput["result"] {
						t.Errorf("Execute() = %v, want %v", got, tt.wantOutput)
					}
				}
			}
		})
	}
}

func TestMockTool_Execute_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tool := &mockTool{
		executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return map[string]any{"result": "success"}, nil
			}
		},
	}

	_, err := tool.Execute(ctx, map[string]any{"message": "hello"})
	if err != context.Canceled {
		t.Errorf("Execute() with canceled context error = %v, want %v", err, context.Canceled)
	}
}

func TestMockTool_Health(t *testing.T) {
	tests := []struct {
		name       string
		healthFunc func(ctx context.Context) types.HealthStatus
		wantStatus string
	}{
		{
			name: "healthy status",
			healthFunc: func(ctx context.Context) types.HealthStatus {
				return types.NewHealthyStatus("all systems operational")
			},
			wantStatus: types.StatusHealthy,
		},
		{
			name: "degraded status",
			healthFunc: func(ctx context.Context) types.HealthStatus {
				return types.NewDegradedStatus("some issues", nil)
			},
			wantStatus: types.StatusDegraded,
		},
		{
			name: "unhealthy status",
			healthFunc: func(ctx context.Context) types.HealthStatus {
				return types.NewUnhealthyStatus("critical failure", nil)
			},
			wantStatus: types.StatusUnhealthy,
		},
		{
			name:       "default healthy",
			healthFunc: nil,
			wantStatus: types.StatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &mockTool{
				healthFunc: tt.healthFunc,
			}

			status := tool.Health(context.Background())
			if status.Status != tt.wantStatus {
				t.Errorf("Health() status = %v, want %v", status.Status, tt.wantStatus)
			}
		})
	}
}

func TestMockTool_InterfaceCompliance(t *testing.T) {
	var _ Tool = (*mockTool)(nil)
}

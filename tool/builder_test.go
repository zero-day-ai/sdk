package tool

import (
	"context"
	"errors"
	"testing"

	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

func TestNewConfig(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("NewConfig() returned nil")
	}

	if cfg.version != "1.0.0" {
		t.Errorf("NewConfig() default version = %v, want %v", cfg.version, "1.0.0")
	}

	if cfg.tags == nil {
		t.Error("NewConfig() tags should not be nil")
	}

	if len(cfg.tags) != 0 {
		t.Errorf("NewConfig() tags length = %v, want 0", len(cfg.tags))
	}
}

func TestConfig_Setters(t *testing.T) {
	cfg := NewConfig()

	// Test SetName
	cfg.SetName("test-tool")
	if cfg.name != "test-tool" {
		t.Errorf("SetName() name = %v, want %v", cfg.name, "test-tool")
	}

	// Test SetVersion
	cfg.SetVersion("2.0.0")
	if cfg.version != "2.0.0" {
		t.Errorf("SetVersion() version = %v, want %v", cfg.version, "2.0.0")
	}

	// Test SetDescription
	cfg.SetDescription("A test tool")
	if cfg.description != "A test tool" {
		t.Errorf("SetDescription() description = %v, want %v", cfg.description, "A test tool")
	}

	// Test SetTags
	tags := []string{"test", "mock"}
	cfg.SetTags(tags)
	if len(cfg.tags) != len(tags) {
		t.Errorf("SetTags() tags length = %v, want %v", len(cfg.tags), len(tags))
	}

	// Test SetInputSchema
	inputSchema := schema.Object(map[string]schema.JSON{
		"message": schema.String(),
	})
	cfg.SetInputSchema(inputSchema)
	if cfg.inputSchema.Type != "object" {
		t.Errorf("SetInputSchema() type = %v, want %v", cfg.inputSchema.Type, "object")
	}

	// Test SetOutputSchema
	outputSchema := schema.Object(map[string]schema.JSON{
		"result": schema.String(),
	})
	cfg.SetOutputSchema(outputSchema)
	if cfg.outputSchema.Type != "object" {
		t.Errorf("SetOutputSchema() type = %v, want %v", cfg.outputSchema.Type, "object")
	}

	// Test SetExecuteFunc
	executeFunc := func(ctx context.Context, input map[string]any) (map[string]any, error) {
		return map[string]any{"result": "success"}, nil
	}
	cfg.SetExecuteFunc(executeFunc)
	if cfg.executeFunc == nil {
		t.Error("SetExecuteFunc() executeFunc should not be nil")
	}
}

func TestConfig_MethodChaining(t *testing.T) {
	cfg := NewConfig().
		SetName("chained-tool").
		SetVersion("1.2.3").
		SetDescription("Chained configuration").
		SetTags([]string{"test"})

	if cfg.name != "chained-tool" {
		t.Errorf("method chaining name = %v, want %v", cfg.name, "chained-tool")
	}

	if cfg.version != "1.2.3" {
		t.Errorf("method chaining version = %v, want %v", cfg.version, "1.2.3")
	}

	if cfg.description != "Chained configuration" {
		t.Errorf("method chaining description = %v, want %v", cfg.description, "Chained configuration")
	}

	if len(cfg.tags) != 1 || cfg.tags[0] != "test" {
		t.Errorf("method chaining tags = %v, want [test]", cfg.tags)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "missing name",
			config: NewConfig().SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, nil
			}),
			wantErr: true,
			errMsg:  "tool name is required",
		},
		{
			name:    "missing execute function",
			config:  NewConfig().SetName("test-tool"),
			wantErr: true,
			errMsg:  "execute function is required",
		},
		{
			name: "valid config",
			config: NewConfig().
				SetName("valid-tool").
				SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
					return map[string]any{"result": "ok"}, nil
				}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := New(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("New() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error = %v", err)
				return
			}

			if tool == nil {
				t.Error("New() returned nil tool")
			}
		})
	}
}

func TestSdkTool_Name(t *testing.T) {
	cfg := NewConfig().
		SetName("test-tool").
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := tool.Name(); got != "test-tool" {
		t.Errorf("Name() = %v, want %v", got, "test-tool")
	}
}

func TestSdkTool_Version(t *testing.T) {
	cfg := NewConfig().
		SetName("test-tool").
		SetVersion("2.5.1").
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := tool.Version(); got != "2.5.1" {
		t.Errorf("Version() = %v, want %v", got, "2.5.1")
	}
}

func TestSdkTool_Description(t *testing.T) {
	cfg := NewConfig().
		SetName("test-tool").
		SetDescription("Test description").
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := tool.Description(); got != "Test description" {
		t.Errorf("Description() = %v, want %v", got, "Test description")
	}
}

func TestSdkTool_Tags(t *testing.T) {
	tags := []string{"test", "mock", "utility"}
	cfg := NewConfig().
		SetName("test-tool").
		SetTags(tags).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	got := tool.Tags()
	if len(got) != len(tags) {
		t.Fatalf("Tags() length = %v, want %v", len(got), len(tags))
	}

	for i, tag := range got {
		if tag != tags[i] {
			t.Errorf("Tags()[%d] = %v, want %v", i, tag, tags[i])
		}
	}
}

func TestSdkTool_InputSchema(t *testing.T) {
	inputSchema := schema.Object(map[string]schema.JSON{
		"name":  schema.String(),
		"age":   schema.Int(),
		"email": schema.String(),
	}, "name", "email")

	cfg := NewConfig().
		SetName("test-tool").
		SetInputSchema(inputSchema).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	got := tool.InputSchema()
	if got.Type != "object" {
		t.Errorf("InputSchema().Type = %v, want object", got.Type)
	}

	if len(got.Required) != 2 {
		t.Errorf("InputSchema().Required length = %v, want 2", len(got.Required))
	}
}

func TestSdkTool_OutputSchema(t *testing.T) {
	outputSchema := schema.Object(map[string]schema.JSON{
		"status":  schema.String(),
		"message": schema.String(),
	}, "status")

	cfg := NewConfig().
		SetName("test-tool").
		SetOutputSchema(outputSchema).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	got := tool.OutputSchema()
	if got.Type != "object" {
		t.Errorf("OutputSchema().Type = %v, want object", got.Type)
	}

	if len(got.Required) != 1 {
		t.Errorf("OutputSchema().Required length = %v, want 1", len(got.Required))
	}
}

func TestSdkTool_Execute(t *testing.T) {
	tests := []struct {
		name           string
		inputSchema    schema.JSON
		outputSchema   schema.JSON
		executeFunc    ExecuteFunc
		input          map[string]any
		wantOutput     map[string]any
		wantErr        bool
		checkOutputVal bool
	}{
		{
			name: "successful execution",
			inputSchema: schema.Object(map[string]schema.JSON{
				"value": schema.Int(),
			}, "value"),
			outputSchema: schema.Object(map[string]schema.JSON{
				"doubled": schema.Int(),
			}, "doubled"),
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				value := input["value"].(int)
				return map[string]any{"doubled": value * 2}, nil
			},
			input:          map[string]any{"value": 5},
			wantOutput:     map[string]any{"doubled": 10},
			wantErr:        false,
			checkOutputVal: true,
		},
		{
			name: "invalid input",
			inputSchema: schema.Object(map[string]schema.JSON{
				"value": schema.Int(),
			}, "value"),
			outputSchema: schema.Object(map[string]schema.JSON{
				"result": schema.String(),
			}),
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return map[string]any{"result": "ok"}, nil
			},
			input:   map[string]any{"wrong": "field"},
			wantErr: true,
		},
		{
			name: "execution error",
			inputSchema: schema.Object(map[string]schema.JSON{
				"value": schema.Int(),
			}),
			outputSchema: schema.Object(map[string]schema.JSON{
				"result": schema.String(),
			}),
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, errors.New("execution failed")
			},
			input:   map[string]any{"value": 5},
			wantErr: true,
		},
		{
			name: "invalid output",
			inputSchema: schema.Object(map[string]schema.JSON{
				"value": schema.Int(),
			}),
			outputSchema: schema.Object(map[string]schema.JSON{
				"result": schema.String(),
			}, "result"),
			executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return map[string]any{"wrong": "field"}, nil
			},
			input:   map[string]any{"value": 5},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig().
				SetName("test-tool").
				SetInputSchema(tt.inputSchema).
				SetOutputSchema(tt.outputSchema).
				SetExecuteFunc(tt.executeFunc)

			tool, err := New(cfg)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}

			got, err := tool.Execute(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Execute() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("Execute() unexpected error = %v", err)
				return
			}

			if tt.checkOutputVal {
				if got == nil {
					t.Error("Execute() returned nil output")
					return
				}
				if got["doubled"] != tt.wantOutput["doubled"] {
					t.Errorf("Execute() output = %v, want %v", got, tt.wantOutput)
				}
			}
		})
	}
}

func TestSdkTool_Execute_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := NewConfig().
		SetName("test-tool").
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return map[string]any{"result": "ok"}, nil
			}
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	cancel() // Cancel context before execution

	_, err = tool.Execute(ctx, map[string]any{})
	if err != context.Canceled {
		t.Errorf("Execute() with canceled context error = %v, want %v", err, context.Canceled)
	}
}

func TestSdkTool_Health(t *testing.T) {
	cfg := NewConfig().
		SetName("test-tool").
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	status := tool.Health(context.Background())

	if status.Status != types.StatusHealthy {
		t.Errorf("Health() status = %v, want %v", status.Status, types.StatusHealthy)
	}

	if status.Message != "tool is operational" {
		t.Errorf("Health() message = %v, want %v", status.Message, "tool is operational")
	}
}

func TestSdkTool_InterfaceCompliance(t *testing.T) {
	var _ Tool = (*sdkTool)(nil)
}

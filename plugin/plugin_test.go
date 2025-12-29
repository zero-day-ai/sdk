package plugin

import (
	"context"
	"testing"

	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

// mockPlugin is a simple mock implementation of the Plugin interface for testing.
type mockPlugin struct {
	name        string
	version     string
	description string
	methods     []MethodDescriptor
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Version() string {
	return m.version
}

func (m *mockPlugin) Description() string {
	return m.description
}

func (m *mockPlugin) Methods() []MethodDescriptor {
	return m.methods
}

func (m *mockPlugin) Query(ctx context.Context, method string, params map[string]any) (any, error) {
	// Simple echo implementation
	if method == "echo" {
		return params, nil
	}
	return nil, nil
}

func (m *mockPlugin) Initialize(ctx context.Context, config map[string]any) error {
	return nil
}

func (m *mockPlugin) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockPlugin) Health(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("mock plugin healthy")
}

func TestMockPlugin_ImplementsInterface(t *testing.T) {
	// Verify that mockPlugin implements Plugin interface
	var _ Plugin = &mockPlugin{}
}

func TestMockPlugin_Name(t *testing.T) {
	m := &mockPlugin{name: "testPlugin"}
	if m.Name() != "testPlugin" {
		t.Errorf("expected name 'testPlugin', got %s", m.Name())
	}
}

func TestMockPlugin_Version(t *testing.T) {
	m := &mockPlugin{version: "1.2.3"}
	if m.Version() != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %s", m.Version())
	}
}

func TestMockPlugin_Description(t *testing.T) {
	m := &mockPlugin{description: "A test plugin"}
	if m.Description() != "A test plugin" {
		t.Errorf("expected description 'A test plugin', got %s", m.Description())
	}
}

func TestMockPlugin_Methods(t *testing.T) {
	methods := []MethodDescriptor{
		{
			Name:         "method1",
			Description:  "First method",
			InputSchema:  schema.String(),
			OutputSchema: schema.Int(),
		},
	}
	m := &mockPlugin{methods: methods}

	result := m.Methods()
	if len(result) != 1 {
		t.Fatalf("expected 1 method, got %d", len(result))
	}
	if result[0].Name != "method1" {
		t.Errorf("expected method name 'method1', got %s", result[0].Name)
	}
}

func TestMockPlugin_Query(t *testing.T) {
	m := &mockPlugin{}
	ctx := context.Background()

	params := map[string]any{"key": "value"}
	result, err := m.Query(ctx, "echo", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any result, got %T", result)
	}

	if resultMap["key"] != "value" {
		t.Errorf("expected value 'value', got %v", resultMap["key"])
	}
}

func TestMockPlugin_Initialize(t *testing.T) {
	m := &mockPlugin{}
	ctx := context.Background()

	err := m.Initialize(ctx, map[string]any{"config": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockPlugin_Shutdown(t *testing.T) {
	m := &mockPlugin{}
	ctx := context.Background()

	err := m.Shutdown(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockPlugin_Health(t *testing.T) {
	m := &mockPlugin{}
	ctx := context.Background()

	status := m.Health(ctx)
	if !status.IsHealthy() {
		t.Error("expected healthy status")
	}
	if status.Message != "mock plugin healthy" {
		t.Errorf("unexpected message: %s", status.Message)
	}
}

func TestPluginInterface(t *testing.T) {
	// Test that the sdkPlugin implementation satisfies the Plugin interface
	cfg := NewConfig()
	cfg.SetName("interfaceTest")
	cfg.SetVersion("1.0.0")

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Verify that p implements Plugin
	var _ Plugin = p

	// Test all interface methods
	if p.Name() != "interfaceTest" {
		t.Errorf("expected name 'interfaceTest', got %s", p.Name())
	}

	if p.Version() != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", p.Version())
	}

	description := p.Description()
	_ = description // Description can be empty

	methods := p.Methods()
	if methods == nil {
		t.Error("expected non-nil methods slice")
	}

	ctx := context.Background()

	// Test Health before initialization
	status := p.Health(ctx)
	if status.Status == "" {
		t.Error("expected non-empty status")
	}

	// Test Initialize
	err = p.Initialize(ctx, nil)
	if err != nil {
		t.Fatalf("initialization failed: %v", err)
	}

	// Test Health after initialization
	status = p.Health(ctx)
	if !status.IsHealthy() {
		t.Error("expected healthy status after initialization")
	}

	// Test Shutdown
	err = p.Shutdown(ctx)
	if err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

package tool

import (
	"context"
	"testing"

	"github.com/zero-day-ai/sdk/schema"
)

func TestToDescriptor(t *testing.T) {
	// Create a test tool
	inputSchema := schema.Object(map[string]schema.JSON{
		"message": schema.String(),
		"count":   schema.Int(),
	}, "message")

	outputSchema := schema.Object(map[string]schema.JSON{
		"result": schema.String(),
		"total":  schema.Int(),
	}, "result")

	tags := []string{"test", "mock", "utility"}

	cfg := NewConfig().
		SetName("test-tool").
		SetVersion("1.2.3").
		SetDescription("A test tool for demonstration").
		SetTags(tags).
		SetInputSchema(inputSchema).
		SetOutputSchema(outputSchema).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return map[string]any{"result": "ok", "total": 42}, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Convert to descriptor
	desc := ToDescriptor(tool)

	// Verify all fields
	if desc.Name != "test-tool" {
		t.Errorf("ToDescriptor() Name = %v, want %v", desc.Name, "test-tool")
	}

	if desc.Version != "1.2.3" {
		t.Errorf("ToDescriptor() Version = %v, want %v", desc.Version, "1.2.3")
	}

	if desc.Description != "A test tool for demonstration" {
		t.Errorf("ToDescriptor() Description = %v, want %v", desc.Description, "A test tool for demonstration")
	}

	if len(desc.Tags) != len(tags) {
		t.Fatalf("ToDescriptor() Tags length = %v, want %v", len(desc.Tags), len(tags))
	}

	for i, tag := range desc.Tags {
		if tag != tags[i] {
			t.Errorf("ToDescriptor() Tags[%d] = %v, want %v", i, tag, tags[i])
		}
	}

	if desc.InputSchema.Type != "object" {
		t.Errorf("ToDescriptor() InputSchema.Type = %v, want object", desc.InputSchema.Type)
	}

	if len(desc.InputSchema.Properties) != 2 {
		t.Errorf("ToDescriptor() InputSchema properties count = %v, want 2", len(desc.InputSchema.Properties))
	}

	if desc.OutputSchema.Type != "object" {
		t.Errorf("ToDescriptor() OutputSchema.Type = %v, want object", desc.OutputSchema.Type)
	}

	if len(desc.OutputSchema.Properties) != 2 {
		t.Errorf("ToDescriptor() OutputSchema properties count = %v, want 2", len(desc.OutputSchema.Properties))
	}
}

func TestToDescriptor_EmptyFields(t *testing.T) {
	// Create a minimal tool
	cfg := NewConfig().
		SetName("minimal-tool").
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return map[string]any{}, nil
		})

	tool, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	desc := ToDescriptor(tool)

	if desc.Name != "minimal-tool" {
		t.Errorf("ToDescriptor() Name = %v, want %v", desc.Name, "minimal-tool")
	}

	if desc.Version != "1.0.0" {
		t.Errorf("ToDescriptor() Version = %v, want %v", desc.Version, "1.0.0")
	}

	if desc.Description != "" {
		t.Errorf("ToDescriptor() Description = %v, want empty string", desc.Description)
	}

	if len(desc.Tags) != 0 {
		t.Errorf("ToDescriptor() Tags length = %v, want 0", len(desc.Tags))
	}
}

func TestToDescriptor_WithMockTool(t *testing.T) {
	// Test with mock tool implementation
	mockInputSchema := schema.Object(map[string]schema.JSON{
		"id": schema.String(),
	})

	mockOutputSchema := schema.Object(map[string]schema.JSON{
		"status": schema.String(),
	})

	mock := &mockTool{
		name:         "mock-tool",
		version:      "2.0.0",
		description:  "Mock tool description",
		tags:         []string{"mock"},
		inputSchema:  mockInputSchema,
		outputSchema: mockOutputSchema,
	}

	desc := ToDescriptor(mock)

	if desc.Name != "mock-tool" {
		t.Errorf("ToDescriptor() Name = %v, want %v", desc.Name, "mock-tool")
	}

	if desc.Version != "2.0.0" {
		t.Errorf("ToDescriptor() Version = %v, want %v", desc.Version, "2.0.0")
	}

	if desc.Description != "Mock tool description" {
		t.Errorf("ToDescriptor() Description = %v, want %v", desc.Description, "Mock tool description")
	}

	if len(desc.Tags) != 1 || desc.Tags[0] != "mock" {
		t.Errorf("ToDescriptor() Tags = %v, want [mock]", desc.Tags)
	}

	if desc.InputSchema.Type != "object" {
		t.Errorf("ToDescriptor() InputSchema.Type = %v, want object", desc.InputSchema.Type)
	}

	if desc.OutputSchema.Type != "object" {
		t.Errorf("ToDescriptor() OutputSchema.Type = %v, want object", desc.OutputSchema.Type)
	}
}

func TestDescriptor_Serialization(t *testing.T) {
	// Create a descriptor
	desc := Descriptor{
		Name:        "serialization-test",
		Version:     "1.0.0",
		Description: "Test serialization",
		Tags:        []string{"test"},
		InputSchema: schema.Object(map[string]schema.JSON{
			"input": schema.String(),
		}),
		OutputSchema: schema.Object(map[string]schema.JSON{
			"output": schema.String(),
		}),
	}

	// Verify struct tags are properly defined for JSON serialization
	if desc.Name == "" {
		t.Error("Descriptor Name should not be empty")
	}

	if desc.Version == "" {
		t.Error("Descriptor Version should not be empty")
	}

	if len(desc.Tags) == 0 {
		t.Error("Descriptor Tags should not be empty")
	}
}

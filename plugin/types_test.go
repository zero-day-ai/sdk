package plugin

import (
	"context"
	"testing"

	"github.com/zero-day-ai/sdk/schema"
)

func TestMethodDescriptor(t *testing.T) {
	desc := MethodDescriptor{
		Name:         "testMethod",
		Description:  "A test method",
		InputSchema:  schema.String(),
		OutputSchema: schema.Int(),
	}

	if desc.Name != "testMethod" {
		t.Errorf("expected name 'testMethod', got %s", desc.Name)
	}
	if desc.Description != "A test method" {
		t.Errorf("expected description 'A test method', got %s", desc.Description)
	}
	if desc.InputSchema.Type != "string" {
		t.Errorf("expected input schema type 'string', got %s", desc.InputSchema.Type)
	}
	if desc.OutputSchema.Type != "integer" {
		t.Errorf("expected output schema type 'integer', got %s", desc.OutputSchema.Type)
	}
}

func TestDescriptor(t *testing.T) {
	methods := []MethodDescriptor{
		{
			Name:         "method1",
			Description:  "First method",
			InputSchema:  schema.String(),
			OutputSchema: schema.Int(),
		},
		{
			Name:         "method2",
			Description:  "Second method",
			InputSchema:  schema.Bool(),
			OutputSchema: schema.String(),
		},
	}

	desc := Descriptor{
		Name:        "testPlugin",
		Version:     "1.0.0",
		Description: "A test plugin",
		Methods:     methods,
	}

	if desc.Name != "testPlugin" {
		t.Errorf("expected name 'testPlugin', got %s", desc.Name)
	}
	if desc.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", desc.Version)
	}
	if desc.Description != "A test plugin" {
		t.Errorf("expected description 'A test plugin', got %s", desc.Description)
	}
	if len(desc.Methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(desc.Methods))
	}
}

func TestToDescriptor(t *testing.T) {
	// Create a mock plugin
	cfg := NewConfig()
	cfg.SetName("mockPlugin")
	cfg.SetVersion("2.0.0")
	cfg.SetDescription("A mock plugin for testing")
	cfg.AddMethodWithDesc(
		"testMethod",
		"A test method",
		func(ctx context.Context, params map[string]any) (any, error) {
			return "result", nil
		},
		schema.Object(map[string]schema.JSON{
			"input": schema.String(),
		}, "input"),
		schema.String(),
	)

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create plugin: %v", err)
	}

	// Convert to descriptor
	desc := ToDescriptor(p)

	if desc.Name != "mockPlugin" {
		t.Errorf("expected name 'mockPlugin', got %s", desc.Name)
	}
	if desc.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %s", desc.Version)
	}
	if desc.Description != "A mock plugin for testing" {
		t.Errorf("expected description 'A mock plugin for testing', got %s", desc.Description)
	}
	if len(desc.Methods) != 1 {
		t.Fatalf("expected 1 method, got %d", len(desc.Methods))
	}
	if desc.Methods[0].Name != "testMethod" {
		t.Errorf("expected method name 'testMethod', got %s", desc.Methods[0].Name)
	}
	if desc.Methods[0].Description != "A test method" {
		t.Errorf("expected method description 'A test method', got %s", desc.Methods[0].Description)
	}
}

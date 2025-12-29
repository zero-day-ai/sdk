package sdk

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/schema"
)

func TestNewFramework(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		fw, err := NewFramework()
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		if fw == nil {
			t.Fatal("expected framework to be non-nil")
		}

		// Verify registries are accessible
		if fw.Agents() == nil {
			t.Error("expected agents registry to be non-nil")
		}
		if fw.Tools() == nil {
			t.Error("expected tools registry to be non-nil")
		}
		if fw.Plugins() == nil {
			t.Error("expected plugins registry to be non-nil")
		}
	})

	t.Run("with custom logger", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		fw, err := NewFramework(WithLogger(logger))
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		if fw == nil {
			t.Fatal("expected framework to be non-nil")
		}
	})

	t.Run("with config path", func(t *testing.T) {
		fw, err := NewFramework(WithConfig("/path/to/config.yaml"))
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		if fw == nil {
			t.Fatal("expected framework to be non-nil")
		}
	})

	t.Run("with multiple options", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		fw, err := NewFramework(
			WithLogger(logger),
			WithConfig("/config.yaml"),
			WithTracer(nil),
		)
		if err != nil {
			t.Fatalf("failed to create framework: %v", err)
		}

		if fw == nil {
			t.Fatal("expected framework to be non-nil")
		}
	})
}

func TestNewAgent(t *testing.T) {
	t.Run("valid agent", func(t *testing.T) {
		a, err := NewAgent(
			WithName("test-agent"),
			WithVersion("1.0.0"),
			WithDescription("A test agent"),
			WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
				return agent.NewSuccessResult("done"), nil
			}),
		)

		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		if a.Name() != "test-agent" {
			t.Errorf("expected name 'test-agent', got %s", a.Name())
		}
		if a.Version() != "1.0.0" {
			t.Errorf("expected version '1.0.0', got %s", a.Version())
		}
		if a.Description() != "A test agent" {
			t.Errorf("expected description 'A test agent', got %s", a.Description())
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := NewAgent(
			WithName("incomplete-agent"),
			// Missing version, description, and execute func
		)

		if err == nil {
			t.Error("expected error for incomplete agent configuration")
		}
	})

	t.Run("with all options", func(t *testing.T) {
		a, err := NewAgent(
			WithName("full-agent"),
			WithVersion("2.0.0"),
			WithDescription("A fully configured agent"),
			WithCapabilities(agent.CapabilityPromptInjection),
			WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
				return agent.NewSuccessResult("done"), nil
			}),
			WithInitFunc(func(ctx context.Context, config map[string]any) error {
				return nil
			}),
			WithShutdownFunc(func(ctx context.Context) error {
				return nil
			}),
		)

		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		if a == nil {
			t.Fatal("expected agent to be non-nil")
		}
	})
}

func TestNewTool(t *testing.T) {
	t.Run("valid tool", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("test-tool"),
			WithToolDescription("A test tool"),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return map[string]any{"result": "success"}, nil
			}),
		)

		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		if tl.Name() != "test-tool" {
			t.Errorf("expected name 'test-tool', got %s", tl.Name())
		}
		if tl.Description() != "A test tool" {
			t.Errorf("expected description 'A test tool', got %s", tl.Description())
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		_, err := NewTool(
			WithToolName("incomplete-tool"),
			// Missing execute handler
		)

		if err == nil {
			t.Error("expected error for incomplete tool configuration")
		}
	})

	t.Run("with schemas", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("schema-tool"),
			WithInputSchema(schema.Object(map[string]schema.JSON{
				"url": schema.String(),
			})),
			WithOutputSchema(schema.Object(map[string]schema.JSON{
				"status": schema.Number(),
			})),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return map[string]any{"status": 200}, nil
			}),
		)

		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		// Verify schemas were set by calling them - they shouldn't panic
		_ = tl.InputSchema()
		_ = tl.OutputSchema()
	})

	t.Run("with tags", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("tagged-tool"),
			WithToolTags("http", "network", "testing"),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, nil
			}),
		)

		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		tags := tl.Tags()
		if len(tags) != 3 {
			t.Errorf("expected 3 tags, got %d", len(tags))
		}
	})
}

func TestNewPlugin(t *testing.T) {
	t.Run("valid plugin", func(t *testing.T) {
		p, err := NewPlugin(
			WithPluginName("test-plugin"),
			WithPluginVersion("1.0.0"),
			WithPluginDescription("A test plugin"),
		)

		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		if p.Name() != "test-plugin" {
			t.Errorf("expected name 'test-plugin', got %s", p.Name())
		}
		if p.Version() != "1.0.0" {
			t.Errorf("expected version '1.0.0', got %s", p.Version())
		}
	})

	t.Run("missing name", func(t *testing.T) {
		_, err := NewPlugin(
			WithPluginVersion("1.0.0"),
		)

		if err == nil {
			t.Error("expected error for plugin without name")
		}
	})

	t.Run("default version", func(t *testing.T) {
		p, err := NewPlugin(
			WithPluginName("default-version-plugin"),
		)

		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		if p.Version() != "1.0.0" {
			t.Errorf("expected default version '1.0.0', got %s", p.Version())
		}
	})
}

func TestServeAgent(t *testing.T) {
	t.Run("not yet implemented", func(t *testing.T) {
		a, err := NewAgent(
			WithName("test-agent"),
			WithVersion("1.0.0"),
			WithDescription("Test"),
			WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
				return agent.NewSuccessResult(nil), nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		err = ServeAgent(a, WithPort(8080))
		if err == nil {
			t.Error("expected error for unimplemented serve function")
		}
	})
}

func TestServeTool(t *testing.T) {
	t.Run("not yet implemented", func(t *testing.T) {
		tl, err := NewTool(
			WithToolName("test-tool"),
			WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
				return nil, nil
			}),
		)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		err = ServeTool(tl, WithPort(8081))
		if err == nil {
			t.Error("expected error for unimplemented serve function")
		}
	})
}

func TestServePlugin(t *testing.T) {
	t.Run("not yet implemented", func(t *testing.T) {
		p, err := NewPlugin(
			WithPluginName("test-plugin"),
		)
		if err != nil {
			t.Fatalf("failed to create plugin: %v", err)
		}

		err = ServePlugin(p, WithPort(8082))
		if err == nil {
			t.Error("expected error for unimplemented serve function")
		}
	})
}

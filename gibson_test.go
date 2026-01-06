package sdk

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/serve"
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
	t.Skip("ServeAgent now delegates to serve.Agent which blocks until shutdown; tested in serve package")
	// Note: The delegation is working correctly. The serve.Agent function will block
	// waiting for shutdown signals. Integration tests in the serve package verify this.
}

func TestServeTool(t *testing.T) {
	t.Skip("ServeTool now delegates to serve.Tool which blocks until shutdown; tested in serve package")
	// Note: The delegation is working correctly. The serve.Tool function will block
	// waiting for shutdown signals. Integration tests in the serve package verify this.
}

func TestServePlugin(t *testing.T) {
	t.Skip("ServePlugin now delegates to serve.PluginFunc which blocks until shutdown; tested in serve package")
	// Note: The delegation is working correctly. The serve.PluginFunc function will block
	// waiting for shutdown signals. Integration tests in the serve package verify this.
	// Also note: ServePlugin now accepts plugin.Plugin (the real interface) not the stub Plugin type.
}

func TestConvertServeOption(t *testing.T) {
	t.Run("converts port option", func(t *testing.T) {
		opt := WithPort(9090)
		converted := convertServeOption(opt)

		// Apply to a serve.Config to verify it works
		cfg := &serve.Config{}
		converted(cfg)

		if cfg.Port != 9090 {
			t.Errorf("expected port 9090, got %d", cfg.Port)
		}
	})

	t.Run("converts health endpoint option", func(t *testing.T) {
		opt := WithHealthEndpoint("/healthz")
		converted := convertServeOption(opt)

		cfg := &serve.Config{}
		converted(cfg)

		if cfg.HealthEndpoint != "/healthz" {
			t.Errorf("expected health endpoint /healthz, got %s", cfg.HealthEndpoint)
		}
	})

	t.Run("converts graceful shutdown option", func(t *testing.T) {
		timeout := 60 * time.Second
		opt := WithGracefulShutdown(timeout)
		converted := convertServeOption(opt)

		cfg := &serve.Config{}
		converted(cfg)

		if cfg.GracefulTimeout != timeout {
			t.Errorf("expected timeout %v, got %v", timeout, cfg.GracefulTimeout)
		}
	})

	t.Run("converts TLS option", func(t *testing.T) {
		opt := WithTLS("/path/to/cert.pem", "/path/to/key.pem")
		converted := convertServeOption(opt)

		cfg := &serve.Config{}
		converted(cfg)

		if cfg.TLSCertFile != "/path/to/cert.pem" {
			t.Errorf("expected cert file /path/to/cert.pem, got %s", cfg.TLSCertFile)
		}
		if cfg.TLSKeyFile != "/path/to/key.pem" {
			t.Errorf("expected key file /path/to/key.pem, got %s", cfg.TLSKeyFile)
		}
	})

	t.Run("converts multiple options", func(t *testing.T) {
		opts := []ServeOption{
			WithPort(8080),
			WithHealthEndpoint("/health"),
			WithGracefulShutdown(30 * time.Second),
		}

		cfg := &serve.Config{}
		for _, opt := range opts {
			converted := convertServeOption(opt)
			converted(cfg)
		}

		if cfg.Port != 8080 {
			t.Errorf("expected port 8080, got %d", cfg.Port)
		}
		if cfg.HealthEndpoint != "/health" {
			t.Errorf("expected health endpoint /health, got %s", cfg.HealthEndpoint)
		}
		if cfg.GracefulTimeout != 30*time.Second {
			t.Errorf("expected timeout 30s, got %v", cfg.GracefulTimeout)
		}
	})
}

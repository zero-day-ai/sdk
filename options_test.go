package sdk

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/tool"
)

func TestFrameworkOptions(t *testing.T) {
	t.Run("WithConfig", func(t *testing.T) {
		cfg := &frameworkConfig{}
		opt := WithConfig("/path/to/config.yaml")
		opt(cfg)

		if cfg.configPath != "/path/to/config.yaml" {
			t.Errorf("expected config path '/path/to/config.yaml', got %s", cfg.configPath)
		}
	})

	t.Run("WithLogger", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		cfg := &frameworkConfig{}
		opt := WithLogger(logger)
		opt(cfg)

		if cfg.logger != logger {
			t.Error("expected logger to be set")
		}
	})

	t.Run("WithTracer", func(t *testing.T) {
		// We can't easily create a real tracer in tests, so we'll just verify
		// the option sets the field to nil (which is valid)
		cfg := &frameworkConfig{}
		opt := WithTracer(nil)
		opt(cfg)

		if cfg.tracer != nil {
			t.Error("expected tracer to be nil")
		}
	})
}

func TestAgentOptions(t *testing.T) {
	t.Run("WithName", func(t *testing.T) {
		cfg := agent.NewConfig()
		opt := WithName("test-agent")
		opt(cfg)

		// Build agent to verify name was set
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Test agent")
		cfg.SetExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		})

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		if a.Name() != "test-agent" {
			t.Errorf("expected name 'test-agent', got %s", a.Name())
		}
	})

	t.Run("WithVersion", func(t *testing.T) {
		cfg := agent.NewConfig()
		opt := WithVersion("2.0.0")
		opt(cfg)

		cfg.SetName("test")
		cfg.SetDescription("Test")
		cfg.SetExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		})

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		if a.Version() != "2.0.0" {
			t.Errorf("expected version '2.0.0', got %s", a.Version())
		}
	})

	t.Run("WithDescription", func(t *testing.T) {
		cfg := agent.NewConfig()
		opt := WithDescription("Test description")
		opt(cfg)

		cfg.SetName("test")
		cfg.SetVersion("1.0.0")
		cfg.SetExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		})

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		if a.Description() != "Test description" {
			t.Errorf("expected description 'Test description', got %s", a.Description())
		}
	})

	t.Run("WithCapabilities", func(t *testing.T) {
		cfg := agent.NewConfig()
		caps := []string{"prompt_injection", "jailbreak"}
		opt := WithCapabilities(caps...)
		opt(cfg)

		cfg.SetName("test")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Test")
		cfg.SetExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		})

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		gotCaps := a.Capabilities()
		if len(gotCaps) != 2 {
			t.Errorf("expected 2 capabilities, got %d", len(gotCaps))
		}
	})

	t.Run("WithTargetTypes", func(t *testing.T) {
		cfg := agent.NewConfig()
		targetTypes := []string{"llm_chat", "llm_api"}
		opt := WithTargetTypes(targetTypes...)
		opt(cfg)

		cfg.SetName("test")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Test")
		cfg.SetExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		})

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		gotTypes := a.TargetTypes()
		if len(gotTypes) != 2 {
			t.Errorf("expected 2 target types, got %d", len(gotTypes))
		}
	})

	t.Run("WithTechniqueTypes", func(t *testing.T) {
		cfg := agent.NewConfig()
		techTypes := []string{"prompt_injection"}
		opt := WithTechniqueTypes(techTypes...)
		opt(cfg)

		cfg.SetName("test")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Test")
		cfg.SetExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		})

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		gotTypes := a.TechniqueTypes()
		if len(gotTypes) != 1 {
			t.Errorf("expected 1 technique type, got %d", len(gotTypes))
		}
	})

	t.Run("WithLLMSlot", func(t *testing.T) {
		cfg := agent.NewConfig()
		requirements := llm.SlotRequirements{
			MinContextWindow: 8000,
			RequiredFeatures: []string{"function_calling"},
		}
		opt := WithLLMSlot("primary", requirements)
		opt(cfg)

		cfg.SetName("test")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Test")
		cfg.SetExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		})

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		slots := a.LLMSlots()
		if len(slots) != 1 {
			t.Errorf("expected 1 LLM slot, got %d", len(slots))
		}
		if slots[0].Name != "primary" {
			t.Errorf("expected slot name 'primary', got %s", slots[0].Name)
		}
	})

	t.Run("WithExecuteFunc", func(t *testing.T) {
		called := false
		executeFunc := func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			called = true
			return agent.NewSuccessResult(nil), nil
		}

		cfg := agent.NewConfig()
		opt := WithExecuteFunc(executeFunc)
		opt(cfg)

		cfg.SetName("test")
		cfg.SetVersion("1.0.0")
		cfg.SetDescription("Test")

		a, err := agent.New(cfg)
		if err != nil {
			t.Fatalf("failed to create agent: %v", err)
		}

		// Execute should call our function
		_, _ = a.Execute(context.Background(), nil, agent.Task{})
		if !called {
			t.Error("expected execute function to be called")
		}
	})
}

func TestToolOptions(t *testing.T) {
	t.Run("WithToolName", func(t *testing.T) {
		cfg := tool.NewConfig()
		opt := WithToolName("test-tool")
		opt(cfg)

		cfg.SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

		tl, err := tool.New(cfg)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		if tl.Name() != "test-tool" {
			t.Errorf("expected name 'test-tool', got %s", tl.Name())
		}
	})

	t.Run("WithToolVersion", func(t *testing.T) {
		cfg := tool.NewConfig()
		opt := WithToolVersion("2.0.0")
		opt(cfg)

		cfg.SetName("test")
		cfg.SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

		tl, err := tool.New(cfg)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		if tl.Version() != "2.0.0" {
			t.Errorf("expected version '2.0.0', got %s", tl.Version())
		}
	})

	t.Run("WithToolDescription", func(t *testing.T) {
		cfg := tool.NewConfig()
		opt := WithToolDescription("Test tool description")
		opt(cfg)

		cfg.SetName("test")
		cfg.SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

		tl, err := tool.New(cfg)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		if tl.Description() != "Test tool description" {
			t.Errorf("expected description 'Test tool description', got %s", tl.Description())
		}
	})

	t.Run("WithToolTags", func(t *testing.T) {
		cfg := tool.NewConfig()
		opt := WithToolTags("http", "network")
		opt(cfg)

		cfg.SetName("test")
		cfg.SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

		tl, err := tool.New(cfg)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		tags := tl.Tags()
		if len(tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(tags))
		}
	})

	t.Run("WithInputSchema", func(t *testing.T) {
		cfg := tool.NewConfig()
		inputSchema := schema.Object(map[string]schema.JSON{
			"url": schema.String(),
		})
		opt := WithInputSchema(inputSchema)
		opt(cfg)

		cfg.SetName("test")
		cfg.SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

		tl, err := tool.New(cfg)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		// Verify schema was set by calling InputSchema() - it shouldn't panic
		_ = tl.InputSchema()
	})

	t.Run("WithOutputSchema", func(t *testing.T) {
		cfg := tool.NewConfig()
		outputSchema := schema.Object(map[string]schema.JSON{
			"status": schema.Number(),
		})
		opt := WithOutputSchema(outputSchema)
		opt(cfg)

		cfg.SetName("test")
		cfg.SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return nil, nil
		})

		tl, err := tool.New(cfg)
		if err != nil {
			t.Fatalf("failed to create tool: %v", err)
		}

		// Verify schema was set by calling OutputSchema() - it shouldn't panic
		_ = tl.OutputSchema()
	})
}

func TestPluginOptions(t *testing.T) {
	t.Run("WithPluginName", func(t *testing.T) {
		cfg := &pluginConfig{}
		opt := WithPluginName("test-plugin")
		opt(cfg)

		if cfg.name != "test-plugin" {
			t.Errorf("expected name 'test-plugin', got %s", cfg.name)
		}
	})

	t.Run("WithPluginVersion", func(t *testing.T) {
		cfg := &pluginConfig{}
		opt := WithPluginVersion("3.0.0")
		opt(cfg)

		if cfg.version != "3.0.0" {
			t.Errorf("expected version '3.0.0', got %s", cfg.version)
		}
	})

	t.Run("WithPluginDescription", func(t *testing.T) {
		cfg := &pluginConfig{}
		opt := WithPluginDescription("Test plugin")
		opt(cfg)

		if cfg.description != "Test plugin" {
			t.Errorf("expected description 'Test plugin', got %s", cfg.description)
		}
	})
}

func TestServeOptions(t *testing.T) {
	t.Run("WithPort", func(t *testing.T) {
		cfg := &serveConfig{}
		opt := WithPort(9090)
		opt(cfg)

		if cfg.port != 9090 {
			t.Errorf("expected port 9090, got %d", cfg.port)
		}
	})

	t.Run("WithHealthEndpoint", func(t *testing.T) {
		cfg := &serveConfig{}
		opt := WithHealthEndpoint("/healthz")
		opt(cfg)

		if cfg.healthEndpoint != "/healthz" {
			t.Errorf("expected health endpoint '/healthz', got %s", cfg.healthEndpoint)
		}
	})

	t.Run("WithGracefulShutdown", func(t *testing.T) {
		cfg := &serveConfig{}
		timeout := 45 * time.Second
		opt := WithGracefulShutdown(timeout)
		opt(cfg)

		if cfg.gracefulTimeout != timeout {
			t.Errorf("expected timeout %v, got %v", timeout, cfg.gracefulTimeout)
		}
	})

	t.Run("WithTLS", func(t *testing.T) {
		cfg := &serveConfig{}
		opt := WithTLS("/path/to/cert.pem", "/path/to/key.pem")
		opt(cfg)

		if cfg.tlsCertFile != "/path/to/cert.pem" {
			t.Errorf("expected cert file '/path/to/cert.pem', got %s", cfg.tlsCertFile)
		}
		if cfg.tlsKeyFile != "/path/to/key.pem" {
			t.Errorf("expected key file '/path/to/key.pem', got %s", cfg.tlsKeyFile)
		}
	})
}

func TestMissionOptions(t *testing.T) {
	t.Run("WithMissionName", func(t *testing.T) {
		cfg := &missionConfig{}
		opt := WithMissionName("Test Mission")
		opt(cfg)

		if cfg.name != "Test Mission" {
			t.Errorf("expected name 'Test Mission', got %s", cfg.name)
		}
	})

	t.Run("WithMissionDescription", func(t *testing.T) {
		cfg := &missionConfig{}
		opt := WithMissionDescription("Testing the target")
		opt(cfg)

		if cfg.description != "Testing the target" {
			t.Errorf("expected description 'Testing the target', got %s", cfg.description)
		}
	})

	t.Run("WithMissionTarget", func(t *testing.T) {
		cfg := &missionConfig{}
		opt := WithMissionTarget("target-123")
		opt(cfg)

		if cfg.targetID != "target-123" {
			t.Errorf("expected target ID 'target-123', got %s", cfg.targetID)
		}
	})

	t.Run("WithMissionAgents", func(t *testing.T) {
		cfg := &missionConfig{}
		opt := WithMissionAgents("agent-1", "agent-2")
		opt(cfg)

		if len(cfg.agentNames) != 2 {
			t.Errorf("expected 2 agents, got %d", len(cfg.agentNames))
		}
	})

	t.Run("WithMissionTimeout", func(t *testing.T) {
		cfg := &missionConfig{}
		timeout := 1 * time.Hour
		opt := WithMissionTimeout(timeout)
		opt(cfg)

		if cfg.timeout != timeout {
			t.Errorf("expected timeout %v, got %v", timeout, cfg.timeout)
		}
	})

	t.Run("WithMissionMetadata", func(t *testing.T) {
		cfg := &missionConfig{}
		opt := WithMissionMetadata("key", "value")
		opt(cfg)

		if cfg.metadata["key"] != "value" {
			t.Errorf("expected metadata['key'] = 'value', got %v", cfg.metadata["key"])
		}
	})
}

func TestListOptions(t *testing.T) {
	t.Run("WithLimit", func(t *testing.T) {
		cfg := &listConfig{}
		opt := WithLimit(50)
		opt(cfg)

		if cfg.limit != 50 {
			t.Errorf("expected limit 50, got %d", cfg.limit)
		}
	})

	t.Run("WithOffset", func(t *testing.T) {
		cfg := &listConfig{}
		opt := WithOffset(100)
		opt(cfg)

		if cfg.offset != 100 {
			t.Errorf("expected offset 100, got %d", cfg.offset)
		}
	})

	t.Run("WithFilter", func(t *testing.T) {
		cfg := &listConfig{}
		opt := WithFilter("status", "active")
		opt(cfg)

		if cfg.filter["status"] != "active" {
			t.Errorf("expected filter['status'] = 'active', got %v", cfg.filter["status"])
		}
	})
}

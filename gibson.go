package sdk

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/tool"
)

// NewFramework creates a new Gibson framework instance.
// The framework provides the main SDK interface for mission management,
// registry access, and finding operations.
//
// Example:
//
//	framework, err := sdk.NewFramework(
//	    sdk.WithLogger(logger),
//	    sdk.WithConfig("/path/to/config.yaml"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer framework.Shutdown(context.Background())
func NewFramework(opts ...FrameworkOption) (Framework, error) {
	cfg := &frameworkConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Create default logger if not provided
	if cfg.logger == nil {
		cfg.logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	// Create the framework instance
	f := &defaultFramework{
		logger:         cfg.logger,
		tracer:         cfg.tracer,
		configPath:     cfg.configPath,
		agents:         newAgentRegistry(cfg.logger),
		tools:          newToolRegistry(cfg.logger),
		plugins:        newPluginRegistry(cfg.logger),
		missions:       make(map[string]*Mission),
		findingsStore:  make(map[string][]findingRecord),
	}

	return f, nil
}

// NewAgent creates a new agent with the provided options.
// The agent must have at minimum a name, version, description, and execute function.
//
// Example:
//
//	agent, err := sdk.NewAgent(
//	    sdk.WithName("prompt-injector"),
//	    sdk.WithVersion("1.0.0"),
//	    sdk.WithDescription("Tests for prompt injection vulnerabilities"),
//	    sdk.WithCapabilities(agent.CapabilityPromptInjection),
//	    sdk.WithExecuteFunc(func(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
//	        // Agent implementation
//	        return agent.NewSuccessResult("completed"), nil
//	    }),
//	)
func NewAgent(opts ...AgentOption) (agent.Agent, error) {
	cfg := agent.NewConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return agent.New(cfg)
}

// NewTool creates a new tool with the provided options.
// The tool must have at minimum a name and execute handler.
//
// Example:
//
//	tool, err := sdk.NewTool(
//	    sdk.WithToolName("http-request"),
//	    sdk.WithToolDescription("Makes HTTP requests"),
//	    sdk.WithToolTags("http", "network"),
//	    sdk.WithInputSchema(schema.Object(map[string]schema.JSON{
//	        "url": schema.String(),
//	    })),
//	    sdk.WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
//	        // Tool implementation
//	        return map[string]any{"status": 200}, nil
//	    }),
//	)
func NewTool(opts ...ToolOption) (tool.Tool, error) {
	cfg := tool.NewConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return tool.New(cfg)
}

// NewPlugin creates a new plugin with the provided options.
// Note: Plugin infrastructure is not yet fully implemented in the SDK.
// This function is a placeholder for future plugin support.
//
// Example (future):
//
//	plugin, err := sdk.NewPlugin(
//	    sdk.WithPluginName("custom-llm"),
//	    sdk.WithPluginVersion("1.0.0"),
//	    sdk.WithMethod("complete", handler, inputSchema, outputSchema),
//	)
func NewPlugin(opts ...PluginOption) (Plugin, error) {
	cfg := &pluginConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// Validate required fields
	if cfg.name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	if cfg.version == "" {
		cfg.version = "1.0.0"
	}

	// Create stub plugin
	return &stubPlugin{
		name:        cfg.name,
		version:     cfg.version,
		description: cfg.description,
	}, nil
}

// ServeAgent starts a gRPC server for the agent.
// The server exposes the agent's functionality over the network,
// allowing the Gibson framework to communicate with it.
//
// Note: The serve infrastructure is not yet implemented.
// This function is a placeholder that returns an error.
//
// Example (future):
//
//	err := sdk.ServeAgent(myAgent,
//	    sdk.WithPort(8080),
//	    sdk.WithHealthEndpoint("/health"),
//	    sdk.WithGracefulShutdown(30*time.Second),
//	)
func ServeAgent(a agent.Agent, opts ...ServeOption) error {
	cfg := &serveConfig{
		port:           8080,
		healthEndpoint: "/health",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return fmt.Errorf("agent serving is not yet implemented")
}

// ServeTool starts a gRPC server for the tool.
// The server exposes the tool's functionality over the network,
// allowing agents to invoke it remotely.
//
// Note: The serve infrastructure is not yet implemented.
// This function is a placeholder that returns an error.
//
// Example (future):
//
//	err := sdk.ServeTool(myTool,
//	    sdk.WithPort(8081),
//	    sdk.WithHealthEndpoint("/health"),
//	)
func ServeTool(t tool.Tool, opts ...ServeOption) error {
	cfg := &serveConfig{
		port:           8081,
		healthEndpoint: "/health",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return fmt.Errorf("tool serving is not yet implemented")
}

// ServePlugin starts a gRPC server for the plugin.
// The server exposes the plugin's methods over the network.
//
// Note: The serve infrastructure is not yet implemented.
// This function is a placeholder that returns an error.
//
// Example (future):
//
//	err := sdk.ServePlugin(myPlugin,
//	    sdk.WithPort(8082),
//	)
func ServePlugin(p Plugin, opts ...ServeOption) error {
	cfg := &serveConfig{
		port:           8082,
		healthEndpoint: "/health",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return fmt.Errorf("plugin serving is not yet implemented")
}

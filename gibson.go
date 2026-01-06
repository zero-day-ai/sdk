package sdk

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/serve"
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
// Plugins extend the SDK functionality by providing named methods that can be invoked
// with parameters and return results. Plugins support initialization, shutdown, and health checks.
//
// Example:
//
//	plugin, err := sdk.NewPlugin(
//	    sdk.WithPluginName("custom-llm"),
//	    sdk.WithPluginVersion("1.0.0"),
//	    sdk.WithPluginDescription("Custom LLM provider"),
//	    sdk.WithMethod("complete", handler, inputSchema, outputSchema),
//	)
func NewPlugin(opts ...PluginOption) (plugin.Plugin, error) {
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

	// Create plugin using the builder
	pluginCfg := plugin.NewConfig()
	pluginCfg.SetName(cfg.name)
	pluginCfg.SetVersion(cfg.version)
	pluginCfg.SetDescription(cfg.description)

	// Add methods to the plugin
	for methodName, methodDef := range cfg.methods {
		// Convert the handler interface to the proper MethodHandler type
		handler, ok := methodDef.handler.(plugin.MethodHandler)
		if !ok {
			return nil, fmt.Errorf("invalid handler type for method %s", methodName)
		}
		pluginCfg.AddMethod(methodName, handler, methodDef.inputSchema, methodDef.outputSchema)
	}

	// Set init and shutdown functions if provided
	if cfg.initFunc != nil {
		pluginCfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
			return cfg.initFunc()
		})
	}
	if cfg.shutdownFunc != nil {
		pluginCfg.SetShutdownFunc(func(ctx context.Context) error {
			return cfg.shutdownFunc()
		})
	}

	return plugin.New(pluginCfg)
}

// convertServeOption converts a public ServeOption to an internal serve.Option.
// This bridges the public SDK API with the internal serve package implementation.
func convertServeOption(opt ServeOption) serve.Option {
	return func(c *serve.Config) {
		// Create a temporary serveConfig to capture the option's values
		tempCfg := &serveConfig{}
		opt(tempCfg)

		// Map serveConfig fields to serve.Config fields
		if tempCfg.port != 0 {
			c.Port = tempCfg.port
		}
		if tempCfg.healthEndpoint != "" {
			c.HealthEndpoint = tempCfg.healthEndpoint
		}
		if tempCfg.gracefulTimeout != 0 {
			c.GracefulTimeout = tempCfg.gracefulTimeout
		}
		if tempCfg.tlsCertFile != "" {
			c.TLSCertFile = tempCfg.tlsCertFile
		}
		if tempCfg.tlsKeyFile != "" {
			c.TLSKeyFile = tempCfg.tlsKeyFile
		}
	}
}

// ServeAgent starts a gRPC server for the agent.
// The server exposes the agent's functionality over the network,
// allowing the Gibson framework to communicate with it.
//
// The server handles agent execution requests via gRPC and provides
// health check endpoints. It supports graceful shutdown and optional TLS.
//
// Example:
//
//	err := sdk.ServeAgent(myAgent,
//	    sdk.WithPort(8080),
//	    sdk.WithHealthEndpoint("/health"),
//	    sdk.WithGracefulShutdown(30*time.Second),
//	)
func ServeAgent(a agent.Agent, opts ...ServeOption) error {
	// Convert public ServeOptions to internal serve.Options
	serveOpts := make([]serve.Option, len(opts))
	for i, opt := range opts {
		serveOpts[i] = convertServeOption(opt)
	}

	// Delegate to the serve package implementation
	return serve.Agent(a, serveOpts...)
}

// ServeTool starts a gRPC server for the tool.
// The server exposes the tool's functionality over the network,
// allowing agents to invoke it remotely.
//
// The server handles tool execution requests via gRPC and provides
// health check endpoints. It supports graceful shutdown and optional TLS.
//
// Example:
//
//	err := sdk.ServeTool(myTool,
//	    sdk.WithPort(8081),
//	    sdk.WithHealthEndpoint("/health"),
//	)
func ServeTool(t tool.Tool, opts ...ServeOption) error {
	// Convert public ServeOptions to internal serve.Options
	serveOpts := make([]serve.Option, len(opts))
	for i, opt := range opts {
		serveOpts[i] = convertServeOption(opt)
	}

	// Delegate to the serve package implementation
	return serve.Tool(t, serveOpts...)
}

// ServePlugin starts a gRPC server for the plugin.
// The server exposes the plugin's methods over the network.
//
// The server handles plugin query requests via gRPC and provides
// health check endpoints. It supports graceful shutdown and optional TLS.
//
// Example:
//
//	err := sdk.ServePlugin(myPlugin,
//	    sdk.WithPort(8082),
//	)
func ServePlugin(p plugin.Plugin, opts ...ServeOption) error {
	// Convert public ServeOptions to internal serve.Options
	serveOpts := make([]serve.Option, len(opts))
	for i, opt := range opts {
		serveOpts[i] = convertServeOption(opt)
	}

	// Delegate to the serve package implementation
	return serve.PluginFunc(p, serveOpts...)
}

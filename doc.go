// Package sdk provides the official Software Development Kit for the Gibson Framework.
//
// The Gibson SDK enables developers to build, deploy, and manage AI agents, tools,
// and plugins within the Gibson Framework ecosystem. It provides a comprehensive set
// of APIs for interacting with the Gibson runtime, creating custom agents, integrating
// tools, and extending framework functionality through plugins.
//
// # Core Concepts
//
// The SDK is organized around several key concepts:
//
//   - Agents: AI-powered entities that perform tasks and interact with users
//   - Tools: Reusable capabilities that agents can invoke to accomplish tasks
//   - Plugins: Extensions that add new functionality to the framework
//   - Slots: Requirements that define LLM capabilities needed by agents
//   - Runtime: The execution environment that manages agent lifecycle and tool execution
//
// # Architecture
//
// The SDK follows a layered architecture:
//
//   - Client Layer: High-level APIs for common operations
//   - Protocol Layer: gRPC-based communication with Gibson runtime
//   - Plugin Layer: Plugin development and integration APIs
//   - Observability Layer: OpenTelemetry-based monitoring and tracing
//
// # Getting Started
//
// To use the SDK, first create a client instance:
//
//	import "github.com/zero-day-ai/sdk"
//
//	client, err := sdk.NewClient(sdk.Config{
//		RuntimeAddr: "localhost:50051",
//		TLS:         true,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
// # Agent Development
//
// Create custom agents by implementing the Agent interface:
//
//	type MyAgent struct {
//		sdk.BaseAgent
//	}
//
//	func (a *MyAgent) Execute(ctx context.Context, input string) (string, error) {
//		// Agent logic here
//		return "result", nil
//	}
//
// # Tool Development
//
// Create custom tools by implementing the Tool interface:
//
//	type MyTool struct {
//		sdk.BaseTool
//	}
//
//	func (t *MyTool) Execute(ctx context.Context, params map[string]any) (any, error) {
//		// Tool logic here
//		return result, nil
//	}
//
// # Plugin Development
//
// Create plugins to extend framework functionality:
//
//	type MyPlugin struct {
//		sdk.BasePlugin
//	}
//
//	func (p *MyPlugin) Initialize(ctx context.Context) error {
//		// Plugin initialization
//		return nil
//	}
//
// # Error Handling
//
// The SDK uses sentinel errors and structured error types for robust error handling:
//
//	if err != nil {
//		if errors.Is(err, sdk.ErrAgentNotFound) {
//			// Handle agent not found
//		}
//		// Handle other errors
//	}
//
// # Observability
//
// The SDK integrates OpenTelemetry for distributed tracing and metrics:
//
//	import "go.opentelemetry.io/otel"
//
//	tracer := otel.Tracer("my-agent")
//	ctx, span := tracer.Start(ctx, "agent-execution")
//	defer span.End()
//
// # Thread Safety
//
// All SDK client methods are safe for concurrent use. Agent and tool implementations
// should ensure thread safety when managing shared state.
//
// # Best Practices
//
//   - Always use context for cancellation and timeouts
//   - Implement proper error handling and error wrapping
//   - Use structured logging for debugging and monitoring
//   - Implement graceful shutdown for long-running operations
//   - Validate input parameters before processing
//   - Use dependency injection for testability
//
// # Examples
//
// See the examples directory for complete working examples of:
//
//   - Creating and deploying custom agents
//   - Building reusable tools
//   - Developing framework plugins
//   - Integrating with existing systems
//   - Testing agents and tools
//
// # Support
//
// For more information, visit:
//
//	Documentation: https://docs.gibson.ai
//	GitHub: https://github.com/zero-day-ai/gibson
//	Community: https://community.gibson.ai
package sdk

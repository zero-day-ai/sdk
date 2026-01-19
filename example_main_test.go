package sdk_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/zero-day-ai/sdk"
	"github.com/zero-day-ai/sdk/agent"
	"github.com/zero-day-ai/sdk/schema"
)

// Helper to create framework without logging
func newQuietFramework() (sdk.Framework, error) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return sdk.NewFramework(sdk.WithLogger(logger))
}

// ExampleNewFramework demonstrates creating and using the Gibson SDK framework.
func ExampleNewFramework() {
	// Create a new framework instance
	framework, err := newQuietFramework()
	if err != nil {
		log.Fatal(err)
	}

	// Start the framework
	ctx := context.Background()
	if err := framework.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer framework.Shutdown(ctx)

	// Access registries
	agents := framework.Agents()
	tools := framework.Tools()

	fmt.Printf("Framework started with %d agents and %d tools\n",
		len(agents.List()), len(tools.List()))

	// Output: Framework started with 0 agents and 0 tools
}

// ExampleNewAgent demonstrates creating a custom agent.
func ExampleNewAgent() {
	// Create an agent that tests for prompt injection vulnerabilities
	agent, err := sdk.NewAgent(
		sdk.WithName("prompt-injector"),
		sdk.WithVersion("1.0.0"),
		sdk.WithDescription("Tests for prompt injection vulnerabilities"),
		sdk.WithCapabilities("prompt_injection"),
		sdk.WithTargetTypes("llm_chat", "llm_api"),
		sdk.WithTechniqueTypes("prompt_injection"),
		sdk.WithExecuteFunc(func(ctx context.Context, harness agent.Harness, task agent.Task) (agent.Result, error) {
			// Agent implementation
			return agent.NewSuccessResult(map[string]any{
				"tests_run":       10,
				"vulnerabilities": 2,
			}), nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Agent: %s v%s - %s\n", agent.Name(), agent.Version(), agent.Description())

	// Output: Agent: prompt-injector v1.0.0 - Tests for prompt injection vulnerabilities
}

// ExampleNewTool demonstrates creating a custom tool.
func ExampleNewTool() {
	// Create a tool for making HTTP requests
	tool, err := sdk.NewTool(
		sdk.WithToolName("http-request"),
		sdk.WithToolDescription("Makes HTTP requests to test endpoints"),
		sdk.WithToolVersion("1.0.0"),
		sdk.WithToolTags("http", "network", "testing"),
		sdk.WithInputSchema(schema.Object(map[string]schema.JSON{
			"url":    schema.String(),
			"method": schema.String(),
		}, "url")),
		sdk.WithOutputSchema(schema.Object(map[string]schema.JSON{
			"status": schema.Number(),
			"body":   schema.String(),
		})),
		sdk.WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			// Tool implementation
			return map[string]any{
				"status": 200,
				"body":   "OK",
			}, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tool: %s - %s\n", tool.Name(), tool.Description())

	// Output: Tool: http-request - Makes HTTP requests to test endpoints
}

// ExampleFramework_CreateMission demonstrates creating and managing missions.
func ExampleFramework_CreateMission() {
	framework, err := newQuietFramework()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := framework.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer framework.Shutdown(ctx)

	// Create a mission
	mission, err := framework.CreateMission(ctx,
		sdk.WithMissionName("Test ChatGPT Security"),
		sdk.WithMissionDescription("Security testing of ChatGPT API"),
		sdk.WithMissionTarget("chatgpt-api"),
		sdk.WithMissionAgents("prompt-injector", "jailbreak-tester"),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Mission: %s (%s)\n", mission.Name, mission.Status)

	// Start the mission
	if err := framework.StartMission(ctx, mission.ID); err != nil {
		log.Fatal(err)
	}

	// Get mission status
	updated, err := framework.GetMission(ctx, mission.ID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Mission status: %s\n", updated.Status)

	// Output:
	// Mission: Test ChatGPT Security (pending)
	// Mission status: running
}

// ExampleFramework_Agents demonstrates using the agent registry.
func ExampleFramework_Agents() {
	framework, err := newQuietFramework()
	if err != nil {
		log.Fatal(err)
	}

	registry := framework.Agents()

	// Create and register an agent
	agent, err := sdk.NewAgent(
		sdk.WithName("test-agent"),
		sdk.WithVersion("1.0.0"),
		sdk.WithDescription("Test agent"),
		sdk.WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			return agent.NewSuccessResult(nil), nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := registry.Register(agent); err != nil {
		log.Fatal(err)
	}

	// List registered agents
	agents := registry.List()
	fmt.Printf("Registered agents: %d\n", len(agents))

	// Get a specific agent
	retrieved, err := registry.Get("test-agent")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Agent: %s v%s\n", retrieved.Name(), retrieved.Version())

	// Output:
	// Registered agents: 1
	// Agent: test-agent v1.0.0
}

// This example is not meant to be run, just to show example usage in documentation
func Example() {
	// Initialize the SDK framework
	framework, err := newQuietFramework()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if err := framework.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer framework.Shutdown(ctx)

	// Create an agent
	myAgent, err := sdk.NewAgent(
		sdk.WithName("my-agent"),
		sdk.WithVersion("1.0.0"),
		sdk.WithDescription("My custom security testing agent"),
		sdk.WithExecuteFunc(func(ctx context.Context, h agent.Harness, task agent.Task) (agent.Result, error) {
			// Execute task
			return agent.NewSuccessResult("completed"), nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register the agent
	if err := framework.Agents().Register(myAgent); err != nil {
		log.Fatal(err)
	}

	// Create a tool
	myTool, err := sdk.NewTool(
		sdk.WithToolName("my-tool"),
		sdk.WithExecuteHandler(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return map[string]any{"result": "success"}, nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register the tool
	if err := framework.Tools().Register(myTool); err != nil {
		log.Fatal(err)
	}

	// Create and run a mission
	mission, err := framework.CreateMission(ctx,
		sdk.WithMissionName("Security Assessment"),
		sdk.WithMissionAgents("my-agent"),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := framework.StartMission(ctx, mission.ID); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Mission started successfully")
	// Output: Mission started successfully
}

func init() {
	// Suppress logging output in examples
	log.SetOutput(os.Stderr)
}

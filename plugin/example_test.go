package plugin_test

import (
	"context"
	"fmt"
	"log"

	"github.com/zero-day-ai/sdk/plugin"
	"github.com/zero-day-ai/sdk/schema"
)

// Example demonstrates creating and using a simple calculator plugin.
func Example() {
	// Create a new plugin configuration
	cfg := plugin.NewConfig()
	cfg.SetName("calculator")
	cfg.SetVersion("1.0.0")
	cfg.SetDescription("A simple calculator plugin")

	// Add an addition method
	cfg.AddMethodWithDesc(
		"add",
		"Adds two numbers together",
		func(ctx context.Context, params map[string]any) (any, error) {
			a := params["a"].(float64)
			b := params["b"].(float64)
			return map[string]any{"result": a + b}, nil
		},
		schema.Object(map[string]schema.JSON{
			"a": schema.Number(),
			"b": schema.Number(),
		}, "a", "b"),
		schema.Object(map[string]schema.JSON{
			"result": schema.Number(),
		}, "result"),
	)

	// Add a multiplication method
	cfg.AddMethodWithDesc(
		"multiply",
		"Multiplies two numbers",
		func(ctx context.Context, params map[string]any) (any, error) {
			a := params["a"].(float64)
			b := params["b"].(float64)
			return map[string]any{"result": a * b}, nil
		},
		schema.Object(map[string]schema.JSON{
			"a": schema.Number(),
			"b": schema.Number(),
		}, "a", "b"),
		schema.Object(map[string]schema.JSON{
			"result": schema.Number(),
		}, "result"),
	)

	// Build the plugin
	p, err := plugin.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Initialize the plugin
	err = p.Initialize(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Use the plugin
	result, err := p.Query(ctx, "add", map[string]any{"a": 5.0, "b": 3.0})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("5 + 3 = %.0f\n", result.(map[string]any)["result"])

	result, err = p.Query(ctx, "multiply", map[string]any{"a": 4.0, "b": 7.0})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("4 * 7 = %.0f\n", result.(map[string]any)["result"])

	// Shutdown the plugin
	err = p.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Output:
	// 5 + 3 = 8
	// 4 * 7 = 28
}

// Example_withInitialization demonstrates plugin initialization with configuration.
func Example_withInitialization() {
	cfg := plugin.NewConfig()
	cfg.SetName("greeter")
	cfg.SetVersion("1.0.0")

	// Store configuration in a variable accessible to handlers
	var prefix string

	cfg.SetInitFunc(func(ctx context.Context, config map[string]any) error {
		if p, ok := config["prefix"].(string); ok {
			prefix = p
		}
		return nil
	})

	cfg.AddMethodWithDesc(
		"greet",
		"Greets a person with the configured prefix",
		func(ctx context.Context, params map[string]any) (any, error) {
			name := params["name"].(string)
			return map[string]any{"greeting": prefix + name}, nil
		},
		schema.Object(map[string]schema.JSON{
			"name": schema.String(),
		}, "name"),
		schema.Object(map[string]schema.JSON{
			"greeting": schema.String(),
		}, "greeting"),
	)

	p, err := plugin.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Initialize with configuration
	err = p.Initialize(ctx, map[string]any{"prefix": "Hello, "})
	if err != nil {
		log.Fatal(err)
	}

	result, err := p.Query(ctx, "greet", map[string]any{"name": "World"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.(map[string]any)["greeting"])

	// Output:
	// Hello, World
}

// Example_healthCheck demonstrates plugin health checking.
func Example_healthCheck() {
	cfg := plugin.NewConfig()
	cfg.SetName("healthyPlugin")
	cfg.SetVersion("1.0.0")

	p, err := plugin.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Check health before initialization
	status := p.Health(ctx)
	fmt.Printf("Before init: %s\n", status.Status)

	// Initialize
	err = p.Initialize(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Check health after initialization
	status = p.Health(ctx)
	fmt.Printf("After init: %s\n", status.Status)

	// Shutdown
	err = p.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Check health after shutdown
	status = p.Health(ctx)
	fmt.Printf("After shutdown: %s\n", status.Status)

	// Output:
	// Before init: unhealthy
	// After init: healthy
	// After shutdown: unhealthy
}

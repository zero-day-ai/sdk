package tool_test

import (
	"context"
	"fmt"
	"log"

	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/tool"
)

// Example demonstrates creating and using a simple calculator tool.
func Example() {
	// Create a calculator tool configuration
	cfg := tool.NewConfig().
		SetName("calculator").
		SetVersion("1.0.0").
		SetDescription("Performs basic arithmetic operations").
		SetTags([]string{"math", "utility"}).
		SetInputSchema(schema.Object(map[string]schema.JSON{
			"operation": schema.StringWithDesc("The operation to perform: add, subtract, multiply, divide"),
			"a":         schema.Number(),
			"b":         schema.Number(),
		}, "operation", "a", "b")).
		SetOutputSchema(schema.Object(map[string]schema.JSON{
			"result": schema.Number(),
		}, "result")).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			op := input["operation"].(string)
			a := input["a"].(float64)
			b := input["b"].(float64)

			var result float64
			switch op {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			case "multiply":
				result = a * b
			case "divide":
				if b == 0 {
					return nil, fmt.Errorf("division by zero")
				}
				result = a / b
			default:
				return nil, fmt.Errorf("unknown operation: %s", op)
			}

			return map[string]any{"result": result}, nil
		})

	// Create the tool
	calculator, err := tool.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Execute the tool
	result, err := calculator.Execute(context.Background(), map[string]any{
		"operation": "add",
		"a":         5.0,
		"b":         3.0,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %.0f\n", result["result"])

	// Output:
	// Result: 8
}

// ExampleToDescriptor demonstrates converting a tool to its descriptor.
func ExampleToDescriptor() {
	// Create a simple tool
	cfg := tool.NewConfig().
		SetName("greeter").
		SetVersion("1.0.0").
		SetDescription("Greets users by name").
		SetTags([]string{"greeting", "example"}).
		SetInputSchema(schema.Object(map[string]schema.JSON{
			"name": schema.String(),
		}, "name")).
		SetOutputSchema(schema.Object(map[string]schema.JSON{
			"message": schema.String(),
		}, "message")).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			name := input["name"].(string)
			return map[string]any{"message": fmt.Sprintf("Hello, %s!", name)}, nil
		})

	greeter, err := tool.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Convert to descriptor
	desc := tool.ToDescriptor(greeter)

	fmt.Printf("Tool: %s v%s\n", desc.Name, desc.Version)
	fmt.Printf("Description: %s\n", desc.Description)
	fmt.Printf("Tags: %v\n", desc.Tags)

	// Output:
	// Tool: greeter v1.0.0
	// Description: Greets users by name
	// Tags: [greeting example]
}

// ExampleTool_Health demonstrates checking tool health.
func ExampleTool_Health() {
	cfg := tool.NewConfig().
		SetName("health-check-example").
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return map[string]any{}, nil
		})

	t, err := tool.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	status := t.Health(context.Background())
	if status.IsHealthy() {
		fmt.Println("Tool is operational")
	}

	// Output:
	// Tool is operational
}

// ExampleNew demonstrates creating a tool with validation.
func ExampleNew() {
	// Create a tool with input and output validation
	cfg := tool.NewConfig().
		SetName("string-reverser").
		SetVersion("1.0.0").
		SetDescription("Reverses a string").
		SetInputSchema(schema.Object(map[string]schema.JSON{
			"text": schema.String(),
		}, "text")).
		SetOutputSchema(schema.Object(map[string]schema.JSON{
			"reversed": schema.String(),
		}, "reversed")).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			text := input["text"].(string)
			runes := []rune(text)
			for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
				runes[i], runes[j] = runes[j], runes[i]
			}
			return map[string]any{"reversed": string(runes)}, nil
		})

	reverser, err := tool.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	result, err := reverser.Execute(context.Background(), map[string]any{
		"text": "hello",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(result["reversed"])

	// Output:
	// olleh
}

// ExampleConfig_SetTags demonstrates setting tool tags for categorization.
func ExampleConfig_SetTags() {
	cfg := tool.NewConfig().
		SetName("tagged-tool").
		SetTags([]string{"data", "transformation", "utility"}).
		SetExecuteFunc(func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return map[string]any{}, nil
		})

	t, err := tool.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Tags: %v\n", t.Tags())

	// Output:
	// Tags: [data transformation utility]
}

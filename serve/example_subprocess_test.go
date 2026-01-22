package serve_test

import (
	"context"
	"fmt"
	"os"

	"github.com/zero-day-ai/sdk/api/gen/proto"
	"github.com/zero-day-ai/sdk/serve"
	"github.com/zero-day-ai/sdk/types"
	protolib "google.golang.org/protobuf/proto"
)

// ExampleTool demonstrates a simple tool implementation using proto messages
type ExampleTool struct{}

func (t *ExampleTool) Name() string {
	return "example-tool"
}

func (t *ExampleTool) Version() string {
	return "1.0.0"
}

func (t *ExampleTool) Description() string {
	return "An example tool that processes messages"
}

func (t *ExampleTool) Tags() []string {
	return []string{"example", "demo"}
}

func (t *ExampleTool) InputMessageType() string {
	return "gibson.common.TypedMap"
}

func (t *ExampleTool) OutputMessageType() string {
	return "gibson.common.TypedMap"
}

func (t *ExampleTool) ExecuteProto(ctx context.Context, input protolib.Message) (protolib.Message, error) {
	// Cast input to TypedMap
	inputMap, ok := input.(*proto.TypedMap)
	if !ok {
		return nil, fmt.Errorf("expected TypedMap input, got %T", input)
	}

	// Get message field
	messageVal, exists := inputMap.Entries["message"]
	if !exists {
		return nil, fmt.Errorf("message field is required")
	}

	msg := messageVal.GetStringValue()

	// Create output with result
	return &proto.TypedMap{
		Entries: map[string]*proto.TypedValue{
			"result": {
				Kind: &proto.TypedValue_StringValue{
					StringValue: fmt.Sprintf("Processed: %s", msg),
				},
			},
		},
	}, nil
}

func (t *ExampleTool) Health(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("Tool is operational")
}

// ExampleRunSubprocess demonstrates how to use RunSubprocess
// This would typically be the main function of a subprocess-based tool
func ExampleRunSubprocess() {
	tool := &ExampleTool{}

	// In a real subprocess, this would read from stdin and write to stdout
	// For demonstration purposes, we show the pattern:
	if err := serve.RunSubprocess(tool); err != nil {
		fmt.Fprintf(os.Stderr, "Tool execution failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// ExampleOutputSchema demonstrates how to use OutputSchema
// This would typically be called when the tool binary is invoked with --schema
func ExampleOutputSchema() {
	tool := &ExampleTool{}

	if err := serve.OutputSchema(tool); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to output schema: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// Example_subprocessMain demonstrates a typical main function for a subprocess-based tool
func Example_subprocessMain() {
	tool := &ExampleTool{}

	// Check for --schema flag
	if len(os.Args) > 1 && os.Args[1] == "--schema" {
		if err := serve.OutputSchema(tool); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Run in subprocess mode
	if err := serve.RunSubprocess(tool); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

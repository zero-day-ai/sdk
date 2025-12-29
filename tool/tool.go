package tool

import (
	"context"

	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

// Tool is the interface for SDK tools.
// Tools are executable components that perform specific operations with defined inputs and outputs.
type Tool interface {
	// Name returns the unique identifier for this tool.
	Name() string

	// Version returns the semantic version of this tool.
	Version() string

	// Description returns a human-readable description of what this tool does.
	Description() string

	// Tags returns a list of tags for categorizing and discovering this tool.
	Tags() []string

	// InputSchema returns the JSON schema defining the expected input structure.
	InputSchema() schema.JSON

	// OutputSchema returns the JSON schema defining the output structure.
	OutputSchema() schema.JSON

	// Execute runs the tool with the provided input and returns the output.
	// The input must conform to the InputSchema, and output will conform to OutputSchema.
	// Context is used for cancellation, deadlines, and request-scoped values.
	Execute(ctx context.Context, input map[string]any) (map[string]any, error)

	// Health checks the operational status of the tool.
	// This can be used to verify dependencies, resources, and readiness.
	Health(ctx context.Context) types.HealthStatus
}

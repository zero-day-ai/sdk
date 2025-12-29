package tool

import "github.com/zero-day-ai/sdk/schema"

// Descriptor describes a tool's metadata.
// It provides a snapshot of a tool's configuration without the execution logic.
type Descriptor struct {
	// Name is the unique identifier for the tool.
	Name string `json:"name"`

	// Version is the semantic version of the tool.
	Version string `json:"version"`

	// Description is a human-readable description of what the tool does.
	Description string `json:"description"`

	// Tags are labels for categorizing and discovering the tool.
	Tags []string `json:"tags"`

	// InputSchema defines the expected input structure.
	InputSchema schema.JSON `json:"input_schema"`

	// OutputSchema defines the expected output structure.
	OutputSchema schema.JSON `json:"output_schema"`
}

// ToDescriptor converts a Tool to its Descriptor.
// This extracts the metadata from a Tool without including the execution logic.
func ToDescriptor(t Tool) Descriptor {
	return Descriptor{
		Name:         t.Name(),
		Version:      t.Version(),
		Description:  t.Description(),
		Tags:         t.Tags(),
		InputSchema:  t.InputSchema(),
		OutputSchema: t.OutputSchema(),
	}
}

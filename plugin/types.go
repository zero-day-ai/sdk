package plugin

import "github.com/zero-day-ai/sdk/schema"

// MethodDescriptor describes a plugin method.
// It defines the method's interface, including its name, purpose, and data schemas
// for both input parameters and output results.
type MethodDescriptor struct {
	// Name is the unique identifier for the method within the plugin.
	Name string

	// Description provides a human-readable explanation of what the method does.
	Description string

	// InputSchema defines the JSON schema for the method's input parameters.
	// This is used for validation and documentation.
	InputSchema schema.JSON

	// OutputSchema defines the JSON schema for the method's return value.
	// This is used for validation and documentation.
	OutputSchema schema.JSON
}

// Descriptor describes a plugin's metadata.
// It provides comprehensive information about the plugin including its identity,
// purpose, and available methods.
type Descriptor struct {
	// Name is the unique identifier for the plugin.
	Name string

	// Version is the semantic version of the plugin.
	Version string

	// Description provides a human-readable explanation of the plugin's purpose.
	Description string

	// Methods lists all available methods that the plugin provides.
	Methods []MethodDescriptor
}

// ToDescriptor converts a Plugin to its Descriptor.
// This extracts the plugin's metadata without requiring access to its implementation.
func ToDescriptor(p Plugin) Descriptor {
	return Descriptor{
		Name:        p.Name(),
		Version:     p.Version(),
		Description: p.Description(),
		Methods:     p.Methods(),
	}
}

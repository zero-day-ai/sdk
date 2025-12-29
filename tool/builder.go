package tool

import (
	"context"
	"errors"

	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

// ExecuteFunc is a function that implements the tool's execution logic.
type ExecuteFunc func(ctx context.Context, input map[string]any) (map[string]any, error)

// Config holds the configuration for building a Tool.
type Config struct {
	name         string
	version      string
	description  string
	tags         []string
	inputSchema  schema.JSON
	outputSchema schema.JSON
	executeFunc  ExecuteFunc
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{
		version:      "1.0.0",
		tags:         []string{},
		inputSchema:  schema.Object(map[string]schema.JSON{}, nil...),
		outputSchema: schema.Object(map[string]schema.JSON{}, nil...),
	}
}

// SetName sets the tool name.
func (c *Config) SetName(name string) *Config {
	c.name = name
	return c
}

// SetVersion sets the tool version.
func (c *Config) SetVersion(version string) *Config {
	c.version = version
	return c
}

// SetDescription sets the tool description.
func (c *Config) SetDescription(desc string) *Config {
	c.description = desc
	return c
}

// SetTags sets the tool tags.
func (c *Config) SetTags(tags []string) *Config {
	c.tags = tags
	return c
}

// SetInputSchema sets the input schema.
func (c *Config) SetInputSchema(s schema.JSON) *Config {
	c.inputSchema = s
	return c
}

// SetOutputSchema sets the output schema.
func (c *Config) SetOutputSchema(s schema.JSON) *Config {
	c.outputSchema = s
	return c
}

// SetExecuteFunc sets the execution function.
func (c *Config) SetExecuteFunc(fn ExecuteFunc) *Config {
	c.executeFunc = fn
	return c
}

// sdkTool is the internal implementation of the Tool interface.
type sdkTool struct {
	name         string
	version      string
	description  string
	tags         []string
	inputSchema  schema.JSON
	outputSchema schema.JSON
	executeFunc  ExecuteFunc
}

// New creates a new Tool from the provided Config.
// Returns an error if required fields (name, executeFunc) are missing.
func New(cfg *Config) (Tool, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	if cfg.name == "" {
		return nil, errors.New("tool name is required")
	}

	if cfg.executeFunc == nil {
		return nil, errors.New("execute function is required")
	}

	return &sdkTool{
		name:         cfg.name,
		version:      cfg.version,
		description:  cfg.description,
		tags:         cfg.tags,
		inputSchema:  cfg.inputSchema,
		outputSchema: cfg.outputSchema,
		executeFunc:  cfg.executeFunc,
	}, nil
}

// Name returns the tool name.
func (t *sdkTool) Name() string {
	return t.name
}

// Version returns the tool version.
func (t *sdkTool) Version() string {
	return t.version
}

// Description returns the tool description.
func (t *sdkTool) Description() string {
	return t.description
}

// Tags returns the tool tags.
func (t *sdkTool) Tags() []string {
	return t.tags
}

// InputSchema returns the input schema.
func (t *sdkTool) InputSchema() schema.JSON {
	return t.inputSchema
}

// OutputSchema returns the output schema.
func (t *sdkTool) OutputSchema() schema.JSON {
	return t.outputSchema
}

// Execute runs the tool's execution function.
func (t *sdkTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	// Validate input against schema
	if err := t.inputSchema.Validate(input); err != nil {
		return nil, err
	}

	// Execute the tool
	output, err := t.executeFunc(ctx, input)
	if err != nil {
		return nil, err
	}

	// Validate output against schema
	if err := t.outputSchema.Validate(output); err != nil {
		return nil, err
	}

	return output, nil
}

// Health returns the health status of the tool.
// By default, tools are always healthy unless they implement custom health checks.
func (t *sdkTool) Health(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("tool is operational")
}

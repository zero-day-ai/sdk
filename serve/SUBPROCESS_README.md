# Subprocess Mode for SDK Tools

This document describes the subprocess mode implementation for SDK tools, which allows tools to be executed as standalone processes that communicate via stdin/stdout.

## Overview

The subprocess mode provides two main functions:

1. **RunSubprocess(t tool.Tool) error** - Executes a tool by reading JSON input from stdin and writing JSON output to stdout
2. **OutputSchema(t tool.Tool) error** - Outputs the tool's schema as JSON to stdout (used with `--schema` flag)

## Usage

### Basic Tool Implementation

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/zero-day-ai/sdk/schema"
    "github.com/zero-day-ai/sdk/serve"
    "github.com/zero-day-ai/sdk/types"
)

type MyTool struct{}

func (t *MyTool) Name() string {
    return "my-tool"
}

func (t *MyTool) Version() string {
    return "1.0.0"
}

func (t *MyTool) Description() string {
    return "My example tool"
}

func (t *MyTool) Tags() []string {
    return []string{"example"}
}

func (t *MyTool) InputSchema() schema.JSON {
    return schema.Object(map[string]schema.JSON{
        "message": schema.StringWithDesc("Input message"),
    }, "message")
}

func (t *MyTool) OutputSchema() schema.JSON {
    return schema.Object(map[string]schema.JSON{
        "result": schema.StringWithDesc("Output result"),
    }, "result")
}

func (t *MyTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
    msg := input["message"].(string)
    return map[string]any{
        "result": fmt.Sprintf("Processed: %s", msg),
    }, nil
}

func (t *MyTool) Health(ctx context.Context) types.HealthStatus {
    return types.NewHealthyStatus("Tool is operational")
}

func main() {
    tool := &MyTool{}

    // Handle --schema flag
    if len(os.Args) > 1 && os.Args[1] == "--schema" {
        if err := serve.OutputSchema(tool); err != nil {
            fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
            os.Exit(1)
        }
        os.Exit(0)
    }

    // Execute tool
    if err := serve.RunSubprocess(tool); err != nil {
        os.Exit(1)
    }
    os.Exit(0)
}
```

### Building and Running

```bash
# Build the tool
go build -o my-tool main.go

# Get the tool schema
./my-tool --schema

# Execute the tool
echo '{"message": "hello"}' | ./my-tool
```

## Schema Output Format

When invoked with `--schema`, the tool outputs a JSON object with the following structure:

```json
{
  "name": "my-tool",
  "version": "1.0.0",
  "description": "My example tool",
  "tags": ["example"],
  "input_schema": {
    "type": "object",
    "properties": {
      "message": {
        "type": "string",
        "description": "Input message"
      }
    },
    "required": ["message"]
  },
  "output_schema": {
    "type": "object",
    "properties": {
      "result": {
        "type": "string",
        "description": "Output result"
      }
    },
    "required": ["result"]
  }
}
```

## Error Handling

The subprocess mode follows these error handling conventions:

1. **stdin read failure** - Error written to stderr, exit code 1
2. **JSON parse failure** - Error written to stderr, exit code 1
3. **Tool execution failure** - Error written to stderr, exit code 1
4. **Output marshal failure** - Error written to stderr, exit code 1
5. **Success** - JSON output written to stdout, exit code 0

All error messages are prefixed with "ERROR: " when written to stderr.

## Integration with Tool Executor Service

This subprocess mode is designed to work with the tool-executor-service, which can:

1. Discover tools by running them with `--schema`
2. Execute tools by spawning the process and sending JSON via stdin
3. Collect results by reading JSON from stdout
4. Handle errors by monitoring stderr and exit codes

## Testing

The implementation includes comprehensive tests:

- `TestRunSubprocess_Success` - Verifies successful execution
- `TestRunSubprocess_InvalidJSON` - Verifies error handling for invalid input
- `TestOutputSchema_Success` - Verifies schema output format

Run tests with:

```bash
go test -v ./serve -run "TestRunSubprocess|TestOutputSchema"
```

## Files

- `subprocess.go` - Core implementation
- `subprocess_test.go` - Unit tests
- `subprocess_integration_test.go` - Integration tests (requires build environment)
- `example_subprocess_test.go` - Example usage patterns

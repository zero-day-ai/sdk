# LLM Package

The `llm` package provides types and interfaces for working with Large Language Models in the Gibson framework.

## Features

- **Message Types**: Structured message types for conversations (system, user, assistant, tool)
- **Completion Requests**: Flexible request/response types with functional options
- **Streaming Support**: Handle streaming LLM responses with automatic accumulation
- **Tool Calling**: Define and execute tools/functions with LLMs
- **Slot Definitions**: Specify LLM requirements and capabilities
- **Token Tracking**: Thread-safe token usage tracking across multiple slots

## Installation

```bash
go get github.com/zero-day-ai/sdk/llm
```

## Quick Start

### Basic Completion Request

```go
package main

import "github.com/zero-day-ai/sdk/llm"

func main() {
    messages := []llm.Message{
        {Role: llm.RoleSystem, Content: "You are a helpful assistant."},
        {Role: llm.RoleUser, Content: "Hello!"},
    }

    req := llm.NewCompletionRequest(messages,
        llm.WithTemperature(0.7),
        llm.WithMaxTokens(1000),
    )
}
```

### Streaming Responses

```go
acc := llm.NewStreamAccumulator()

for chunk := range streamChannel {
    acc.Add(chunk)

    if chunk.HasContent() {
        fmt.Print(chunk.Delta)
    }

    if chunk.IsFinal() {
        response := acc.ToResponse()
        fmt.Printf("\n\nUsage: %d tokens\n", response.Usage.TotalTokens)
        break
    }
}
```

### Tool Calling

```go
// Define a tool
weatherTool := llm.ToolDef{
    Name:        "get_weather",
    Description: "Get current weather for a location",
    Parameters: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "location": map[string]any{
                "type":        "string",
                "description": "City name",
            },
        },
        "required": []string{"location"},
    },
}

// Use in a completion request
req := llm.NewCompletionRequest(messages,
    llm.WithTools(weatherTool),
)

// Handle tool calls in response
if response.HasToolCalls() {
    for _, call := range response.ToolCalls {
        // Parse arguments
        var args struct {
            Location string `json:"location"`
        }
        if err := call.ParseArguments(&args); err != nil {
            // Handle error
        }

        // Execute tool and create result
        result := llm.NewToolResult(call.ID, `{"temp": 72, "condition": "sunny"}`)
    }
}
```

### Token Tracking

```go
tracker := llm.NewTokenTracker()

// Track usage from multiple completions
tracker.Add("primary", response1.Usage)
tracker.Add("primary", response2.Usage)
tracker.Add("vision", response3.Usage)

// Get total usage
total := tracker.Total()
fmt.Printf("Total tokens: %d (input: %d, output: %d)\n",
    total.TotalTokens, total.InputTokens, total.OutputTokens)

// Get usage by slot
primaryUsage := tracker.BySlot("primary")
fmt.Printf("Primary slot tokens: %d\n", primaryUsage.TotalTokens)

// Get all tracked slots
slots := tracker.Slots()
fmt.Printf("Tracked slots: %v\n", slots)
```

### Slot Definitions

```go
// Define LLM requirements
primarySlot := llm.SlotDefinition{
    Name:             "primary",
    Description:      "Main conversational LLM",
    Required:         true,
    MinContextWindow: 32000,
    RequiredFeatures: []string{"function_calling", "streaming"},
    PreferredModels:  []string{"gpt-4-turbo", "claude-3-opus"},
}

// Validate the definition
if err := primarySlot.Validate(); err != nil {
    // Handle error
}

// Check if model satisfies requirements
requirements := primarySlot.ToRequirements()
features := []string{"function_calling", "streaming", "vision"}
contextWindow := 128000

if requirements.Satisfies(features, contextWindow) {
    fmt.Println("Model meets requirements!")
}
```

## Package Structure

- `message.go` - Message types and roles
- `completion.go` - Completion request/response types
- `stream.go` - Streaming response handling
- `tools.go` - Tool definition and execution types
- `slot.go` - Slot definitions and requirements
- `tracker.go` - Token usage tracking

## Testing

The package includes comprehensive tests with 99.4% coverage:

```bash
go test ./llm/...
go test -cover ./llm/...
```

## Thread Safety

The `DefaultTokenTracker` implementation is thread-safe and can be safely used from multiple goroutines. All read and write operations are protected by mutexes.

## Design Principles

1. **Immutability**: Value types are preferred over pointer types where possible
2. **Validation**: Types include validation methods to ensure correctness
3. **Flexibility**: Functional options pattern for extensible configuration
4. **Type Safety**: Strong typing with custom types for roles, choices, etc.
5. **Simplicity**: Clear, idiomatic Go code following best practices

## License

Part of the Gibson Framework - see main repository for license details.

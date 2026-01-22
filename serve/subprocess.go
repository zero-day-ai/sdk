// Package serve provides functions for serving SDK tools and agents.
//
// This package includes two main modes of operation:
//
//  1. gRPC Server Mode: Use Tool() or Agent() to start a gRPC server that handles
//     requests via the SDK protocol. This is the recommended mode for production deployments.
//
//  2. Subprocess Mode: Use RunSubprocess() and OutputSchema() for tools that will be
//     executed as subprocesses. This mode is useful for simple execution models where
//     the parent process wants to spawn tool processes and communicate via stdin/stdout.
//
// Subprocess Mode Usage:
//
// Tools using subprocess mode should implement a main function like:
//
//	func main() {
//	    tool := &MyTool{}
//
//	    // Handle --schema flag
//	    if len(os.Args) > 1 && os.Args[1] == "--schema" {
//	        if err := serve.OutputSchema(tool); err != nil {
//	            os.Exit(1)
//	        }
//	        os.Exit(0)
//	    }
//
//	    // Execute tool
//	    if err := serve.RunSubprocess(tool); err != nil {
//	        os.Exit(1)
//	    }
//	    os.Exit(0)
//	}
//
// The parent process can then:
// - Get the tool schema: `./tool --schema`
// - Execute the tool: `echo '{"input": "data"}' | ./tool`
package serve

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/zero-day-ai/sdk/tool"
)

// RunSubprocess executes a tool in subprocess mode.
// It reads JSON input from stdin, executes the tool, and writes JSON output to stdout.
// Errors are written to stderr and the process exits with a non-zero code.
//
// This mode is designed for tools that are executed as subprocesses by a parent process,
// allowing for simple JSON-based IPC through standard streams.
//
// Example:
//
//	tool := &MyTool{}
//	if err := serve.RunSubprocess(tool); err != nil {
//	    // Error is already written to stderr
//	    os.Exit(1)
//	}
func RunSubprocess(t tool.Tool) error {
	// NOTE: Subprocess mode with Execute has been removed.
	// Tools must now use gRPC mode with ExecuteProto.
	// See serve.Tool() for the recommended approach.
	writeError("subprocess mode with Execute is no longer supported - use gRPC mode with serve.Tool()")
	return fmt.Errorf("subprocess mode is deprecated")
}

// OutputSchema outputs the tool's schema as JSON to stdout.
// This is called when the tool is invoked with the --schema flag.
//
// The schema output includes:
//   - name: Tool name
//   - version: Tool version
//   - description: Tool description
//   - tags: Tool tags
//   - input_schema: JSON schema for input
//   - output_schema: JSON schema for output
//
// Example:
//
//	tool := &MyTool{}
//	if err := serve.OutputSchema(tool); err != nil {
//	    // Error is already written to stderr
//	    os.Exit(1)
//	}
func OutputSchema(t tool.Tool) error {
	// Build schema output
	schema := map[string]any{
		"name":                t.Name(),
		"version":             t.Version(),
		"description":         t.Description(),
		"tags":                t.Tags(),
		"input_message_type":  t.InputMessageType(),
		"output_message_type": t.OutputMessageType(),
	}

	// Marshal schema to JSON
	schemaBytes, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		writeError("failed to marshal schema: %v", err)
		return err
	}

	// Write schema to stdout
	if _, err := os.Stdout.Write(schemaBytes); err != nil {
		writeError("failed to write schema: %v", err)
		return err
	}

	// Write newline for clean output
	if _, err := os.Stdout.Write([]byte("\n")); err != nil {
		writeError("failed to write newline: %v", err)
		return err
	}

	return nil
}

// writeError writes a formatted error message to stderr.
func writeError(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", msg)
}

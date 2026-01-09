package serve

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/schema"
)

func TestRunSubprocess_Success(t *testing.T) {
	// Create a mock tool
	mt := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "A test tool",
		tags:        []string{"test"},
		inputSchema: schema.Object(map[string]schema.JSON{
			"message": schema.String(),
		}, "message"),
		executeFunc: func(ctx context.Context, input map[string]any) (map[string]any, error) {
			return map[string]any{
				"result": "processed: " + input["message"].(string),
			}, nil
		},
	}

	// Create input
	input := map[string]any{
		"message": "hello",
	}
	inputBytes, err := json.Marshal(input)
	require.NoError(t, err)

	// Replace stdin with our input
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write input to pipe
	go func() {
		w.Write(inputBytes)
		w.Close()
	}()

	// Capture stdout
	oldStdout := os.Stdout
	rOut, wOut, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = wOut
	defer func() { os.Stdout = oldStdout }()

	// Run subprocess
	errChan := make(chan error, 1)
	go func() {
		errChan <- RunSubprocess(mt)
		wOut.Close()
	}()

	// Read output
	var outputBuf bytes.Buffer
	outputBuf.ReadFrom(rOut)

	// Wait for completion
	err = <-errChan
	require.NoError(t, err)

	// Parse output
	var output map[string]any
	err = json.Unmarshal(outputBuf.Bytes(), &output)
	require.NoError(t, err)

	// Verify output
	assert.Equal(t, "processed: hello", output["result"])
}

func TestRunSubprocess_InvalidJSON(t *testing.T) {
	// Create a mock tool
	mt := &mockTool{
		name:    "test-tool",
		version: "1.0.0",
	}

	// Replace stdin with invalid JSON
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Write invalid JSON to pipe
	go func() {
		w.Write([]byte("invalid json"))
		w.Close()
	}()

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = wErr
	defer func() { os.Stderr = oldStderr }()

	// Run subprocess
	errChan := make(chan error, 1)
	go func() {
		errChan <- RunSubprocess(mt)
		wErr.Close()
	}()

	// Read stderr
	var stderrBuf bytes.Buffer
	stderrBuf.ReadFrom(rErr)

	// Wait for completion
	err = <-errChan
	require.Error(t, err)

	// Verify error message was written to stderr
	assert.Contains(t, stderrBuf.String(), "failed to parse input JSON")
}

func TestOutputSchema_Success(t *testing.T) {
	// Create a mock tool
	mt := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "A test tool",
		tags:        []string{"test", "example"},
		inputSchema: schema.Object(map[string]schema.JSON{
			"message": schema.StringWithDesc("Input message"),
		}, "message"),
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	// Run OutputSchema
	errChan := make(chan error, 1)
	go func() {
		errChan <- OutputSchema(mt)
		w.Close()
	}()

	// Read output
	var outputBuf bytes.Buffer
	outputBuf.ReadFrom(r)

	// Wait for completion
	err = <-errChan
	require.NoError(t, err)

	// Parse output
	var schemaOutput map[string]any
	err = json.Unmarshal(outputBuf.Bytes(), &schemaOutput)
	require.NoError(t, err)

	// Verify schema fields
	assert.Equal(t, "test-tool", schemaOutput["name"])
	assert.Equal(t, "1.0.0", schemaOutput["version"])
	assert.Equal(t, "A test tool", schemaOutput["description"])
	assert.Equal(t, []any{"test", "example"}, schemaOutput["tags"])
	assert.NotNil(t, schemaOutput["input_schema"])
	assert.NotNil(t, schemaOutput["output_schema"])

	// Verify input schema structure
	inputSchema := schemaOutput["input_schema"].(map[string]any)
	assert.Equal(t, "object", inputSchema["type"])
	assert.NotNil(t, inputSchema["properties"])
	assert.Equal(t, []any{"message"}, inputSchema["required"])

	// Verify output schema structure
	outputSchema := schemaOutput["output_schema"].(map[string]any)
	assert.Equal(t, "object", outputSchema["type"])

	// Check properties exist and contain "result"
	properties, ok := outputSchema["properties"].(map[string]any)
	assert.True(t, ok, "properties should be a map")
	assert.Contains(t, properties, "result")
}

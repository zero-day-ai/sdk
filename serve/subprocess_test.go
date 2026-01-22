package serve

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSubprocess_Deprecated(t *testing.T) {
	// Create a mock tool
	mt := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "A test tool",
		tags:        []string{"test"},
	}

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = wErr
	defer func() { os.Stderr = oldStderr }()

	// Run subprocess (should fail with deprecation message)
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

	// Verify deprecation message
	assert.Contains(t, stderrBuf.String(), "subprocess mode with Execute is no longer supported")
	assert.Contains(t, stderrBuf.String(), "use gRPC mode with serve.Tool()")
}

func TestRunSubprocess_AlsoDeprecated(t *testing.T) {
	// Create a mock tool
	mt := &mockTool{
		name:    "test-tool",
		version: "1.0.0",
	}

	// Capture stderr
	oldStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = wErr
	defer func() { os.Stderr = oldStderr }()

	// Run subprocess (should fail with deprecation message)
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

	// Verify deprecation message was written to stderr
	assert.Contains(t, stderrBuf.String(), "subprocess mode with Execute is no longer supported")
}

func TestOutputSchema_Success(t *testing.T) {
	// Create a mock tool
	mt := &mockTool{
		name:        "test-tool",
		version:     "1.0.0",
		description: "A test tool",
		tags:        []string{"test", "example"},
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

	// Verify proto message types are present (schemas are deprecated)
	assert.Equal(t, "gibson.common.TypedMap", schemaOutput["input_message_type"])
	assert.Equal(t, "gibson.common.TypedMap", schemaOutput["output_message_type"])

	// Old schema fields should not be present
	_, hasInputSchema := schemaOutput["input_schema"]
	_, hasOutputSchema := schemaOutput["output_schema"]
	assert.False(t, hasInputSchema, "input_schema should not be present in output")
	assert.False(t, hasOutputSchema, "output_schema should not be present in output")
}

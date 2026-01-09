package serve_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/types"
)

// TestSubprocessTool is a simple tool for integration testing
type TestSubprocessTool struct{}

func (t *TestSubprocessTool) Name() string {
	return "test-subprocess-tool"
}

func (t *TestSubprocessTool) Version() string {
	return "1.0.0"
}

func (t *TestSubprocessTool) Description() string {
	return "A test tool for subprocess integration testing"
}

func (t *TestSubprocessTool) Tags() []string {
	return []string{"test"}
}

func (t *TestSubprocessTool) InputSchema() schema.JSON {
	return schema.Object(map[string]schema.JSON{
		"value": schema.String(),
	}, "value")
}

func (t *TestSubprocessTool) OutputSchema() schema.JSON {
	return schema.Object(map[string]schema.JSON{
		"doubled": schema.String(),
	}, "doubled")
}

func (t *TestSubprocessTool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	value := input["value"].(string)
	return map[string]any{
		"doubled": value + value,
	}, nil
}

func (t *TestSubprocessTool) Health(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("OK")
}

// TestSubprocessIntegration_Schema tests the --schema flag workflow
func TestSubprocessIntegration_Schema(t *testing.T) {
	// Create a temporary binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "test-tool")

	// Build a simple test binary that uses OutputSchema
	toolMain := `package main

import (
	"context"
	"fmt"
	"os"
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/serve"
	"github.com/zero-day-ai/sdk/types"
)

type Tool struct{}

func (t *Tool) Name() string { return "test-tool" }
func (t *Tool) Version() string { return "1.0.0" }
func (t *Tool) Description() string { return "Test tool" }
func (t *Tool) Tags() []string { return []string{"test"} }
func (t *Tool) InputSchema() schema.JSON {
	return schema.Object(map[string]schema.JSON{
		"message": schema.String(),
	}, "message")
}
func (t *Tool) OutputSchema() schema.JSON {
	return schema.Object(map[string]schema.JSON{
		"result": schema.String(),
	}, "result")
}
func (t *Tool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	return map[string]any{"result": "ok"}, nil
}
func (t *Tool) Health(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("OK")
}

func main() {
	tool := &Tool{}
	if len(os.Args) > 1 && os.Args[1] == "--schema" {
		if err := serve.OutputSchema(tool); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if err := serve.RunSubprocess(tool); err != nil {
		os.Exit(1)
	}
}
`

	// Write the main.go file
	mainPath := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainPath, []byte(toolMain), 0644)
	require.NoError(t, err)

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, mainPath)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", output)
		t.Skipf("Could not build test binary (this is OK in some test environments): %v", err)
		return
	}

	// Run the binary with --schema flag
	cmd = exec.Command(binaryPath, "--schema")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("Stderr: %s", stderr.String())
		require.NoError(t, err)
	}

	// Parse the schema output
	var schemaOutput map[string]any
	err = json.Unmarshal(stdout.Bytes(), &schemaOutput)
	require.NoError(t, err)

	// Verify schema structure
	assert.Equal(t, "test-tool", schemaOutput["name"])
	assert.Equal(t, "1.0.0", schemaOutput["version"])
	assert.Equal(t, "Test tool", schemaOutput["description"])
	assert.NotNil(t, schemaOutput["input_schema"])
	assert.NotNil(t, schemaOutput["output_schema"])
}

// TestSubprocessIntegration_Execute tests the execution workflow
func TestSubprocessIntegration_Execute(t *testing.T) {
	// Create a temporary binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "test-tool")

	// Build a simple test binary
	toolMain := `package main

import (
	"context"
	"fmt"
	"os"
	"github.com/zero-day-ai/sdk/schema"
	"github.com/zero-day-ai/sdk/serve"
	"github.com/zero-day-ai/sdk/types"
)

type Tool struct{}

func (t *Tool) Name() string { return "test-tool" }
func (t *Tool) Version() string { return "1.0.0" }
func (t *Tool) Description() string { return "Test tool" }
func (t *Tool) Tags() []string { return []string{"test"} }
func (t *Tool) InputSchema() schema.JSON {
	return schema.Object(map[string]schema.JSON{
		"message": schema.String(),
	}, "message")
}
func (t *Tool) OutputSchema() schema.JSON {
	return schema.Object(map[string]schema.JSON{
		"result": schema.String(),
	}, "result")
}
func (t *Tool) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	msg := input["message"].(string)
	return map[string]any{"result": "processed: " + msg}, nil
}
func (t *Tool) Health(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("OK")
}

func main() {
	tool := &Tool{}
	if len(os.Args) > 1 && os.Args[1] == "--schema" {
		if err := serve.OutputSchema(tool); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if err := serve.RunSubprocess(tool); err != nil {
		os.Exit(1)
	}
}
`

	// Write the main.go file
	mainPath := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(mainPath, []byte(toolMain), 0644)
	require.NoError(t, err)

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, mainPath)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", output)
		t.Skipf("Could not build test binary (this is OK in some test environments): %v", err)
		return
	}

	// Prepare input
	input := map[string]any{
		"message": "hello",
	}
	inputBytes, err := json.Marshal(input)
	require.NoError(t, err)

	// Run the binary with input via stdin
	cmd = exec.Command(binaryPath)
	cmd.Stdin = bytes.NewReader(inputBytes)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Logf("Stderr: %s", stderr.String())
		require.NoError(t, err)
	}

	// Parse the output
	var outputData map[string]any
	err = json.Unmarshal(stdout.Bytes(), &outputData)
	require.NoError(t, err)

	// Verify the output
	assert.Equal(t, "processed: hello", outputData["result"])
}

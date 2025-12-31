// Package sdk provides integration tests verifying the exec, input, toolerr, and health
// packages work together correctly for tool development.
package sdk_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zero-day-ai/sdk/exec"
	"github.com/zero-day-ai/sdk/health"
	"github.com/zero-day-ai/sdk/input"
	"github.com/zero-day-ai/sdk/toolerr"
	"github.com/zero-day-ai/sdk/types"
)

// contains is a helper for checking if s contains substr
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TestIntegration_SampleTool demonstrates a minimal tool implementation using all four
// SDK utility packages. This validates that the packages work together correctly
// and can be used to build real tools.
func TestIntegration_SampleTool(t *testing.T) {
	// Create a sample tool that:
	// 1. Uses input helpers to extract parameters
	// 2. Uses health checks to verify dependencies
	// 3. Uses exec to run a command
	// 4. Uses toolerr for structured error handling

	tool := &sampleTool{
		name:        "echo-tool",
		description: "A sample tool that echoes input using the echo command",
	}

	// Test health check
	t.Run("Health", func(t *testing.T) {
		status := tool.Health(context.Background())
		if status.Status != types.StatusHealthy {
			t.Errorf("Expected healthy status, got %s: %s", status.Status, status.Message)
		}
	})

	// Test successful execution
	t.Run("Execute_Success", func(t *testing.T) {
		inputs := map[string]any{
			"message": "Hello, SDK!",
			"timeout": 5,
		}

		result, err := tool.Execute(context.Background(), inputs)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		output, ok := result["output"].(string)
		if !ok {
			t.Fatal("Expected output to be a string")
		}
		if output != "Hello, SDK!\n" {
			t.Errorf("Expected 'Hello, SDK!\\n', got %q", output)
		}
	})

	// Test with default values
	t.Run("Execute_Defaults", func(t *testing.T) {
		inputs := map[string]any{} // Empty inputs, should use defaults

		result, err := tool.Execute(context.Background(), inputs)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		output, ok := result["output"].(string)
		if !ok {
			t.Fatal("Expected output to be a string")
		}
		if output != "default message\n" {
			t.Errorf("Expected 'default message\\n', got %q", output)
		}
	})

	// Test timeout handling
	t.Run("Execute_Timeout", func(t *testing.T) {
		inputs := map[string]any{
			"message": "will timeout",
			"timeout": "50ms", // Very short timeout
			"sleep":   true,   // Make the command sleep
		}

		_, err := tool.Execute(context.Background(), inputs)
		if err == nil {
			t.Fatal("Expected timeout error, got nil")
		}

		// Verify it's a structured tool error
		var toolErr *toolerr.Error
		if !errors.As(err, &toolErr) {
			t.Fatalf("Expected toolerr.Error, got %T: %v", err, err)
		}

		if toolErr.Code != toolerr.ErrCodeTimeout {
			t.Errorf("Expected error code %s, got %s", toolerr.ErrCodeTimeout, toolErr.Code)
		}
	})

	// Test error handling for missing binary
	t.Run("Execute_MissingBinary", func(t *testing.T) {
		inputs := map[string]any{
			"command": "nonexistent-binary-xyz",
		}

		_, err := tool.Execute(context.Background(), inputs)
		if err == nil {
			t.Fatal("Expected error for missing binary, got nil")
		}

		var toolErr *toolerr.Error
		if !errors.As(err, &toolErr) {
			t.Fatalf("Expected toolerr.Error, got %T: %v", err, err)
		}

		if toolErr.Code != toolerr.ErrCodeBinaryNotFound {
			t.Errorf("Expected error code %s, got %s", toolerr.ErrCodeBinaryNotFound, toolErr.Code)
		}
	})
}

// TestIntegration_HealthCombine verifies health check combination works correctly
// with checks from different sources.
func TestIntegration_HealthCombine(t *testing.T) {
	checks := []types.HealthStatus{
		health.BinaryCheck("echo"),           // Should be healthy
		health.FileCheck("/tmp"),             // Should be healthy
		health.BinaryCheck("nonexistent123"), // Should be unhealthy
	}

	combined := health.Combine(checks...)

	if combined.Status != types.StatusUnhealthy {
		t.Errorf("Expected unhealthy (one check failed), got %s", combined.Status)
	}
}

// TestIntegration_InputTypes verifies input helpers work with realistic tool inputs
func TestIntegration_InputTypes(t *testing.T) {
	// Simulate inputs that would come from JSON-decoded tool invocation
	inputs := map[string]any{
		"target":      "192.168.1.1",
		"ports":       []interface{}{"80", "443", "8080"}, // JSON decodes arrays as []interface{}
		"timeout":     30,                                 // JSON decodes numbers as float64 or int
		"verbose":     true,
		"concurrency": float64(10), // JSON always decodes numbers as float64
		"duration":    "5m",
	}

	target := input.GetString(inputs, "target", "")
	if target != "192.168.1.1" {
		t.Errorf("Expected target '192.168.1.1', got %q", target)
	}

	ports := input.GetStringSlice(inputs, "ports")
	if len(ports) != 3 || ports[0] != "80" {
		t.Errorf("Expected 3 ports starting with '80', got %v", ports)
	}

	timeout := input.GetTimeout(inputs, "timeout", time.Minute)
	if timeout != 30*time.Second {
		t.Errorf("Expected 30s timeout, got %v", timeout)
	}

	verbose := input.GetBool(inputs, "verbose", false)
	if !verbose {
		t.Error("Expected verbose to be true")
	}

	concurrency := input.GetInt(inputs, "concurrency", 1)
	if concurrency != 10 {
		t.Errorf("Expected concurrency 10, got %d", concurrency)
	}

	duration := input.GetTimeout(inputs, "duration", time.Minute)
	if duration != 5*time.Minute {
		t.Errorf("Expected 5m duration, got %v", duration)
	}
}

// TestIntegration_ErrorChaining verifies error wrapping and unwrapping works
func TestIntegration_ErrorChaining(t *testing.T) {
	// Create a chain of errors like a real tool would
	originalErr := errors.New("connection refused")

	toolErr := toolerr.New("port-scanner", "connect", toolerr.ErrCodeNetworkError, "failed to connect to target").
		WithCause(originalErr).
		WithDetails(map[string]any{
			"host": "192.168.1.1",
			"port": 8080,
		})

	// Verify error message format
	errStr := toolErr.Error()
	expected := "port-scanner [connect/NETWORK_ERROR]: failed to connect to target: connection refused"
	if errStr != expected {
		t.Errorf("Expected error string:\n%s\nGot:\n%s", expected, errStr)
	}

	// Verify unwrap works
	if !errors.Is(toolErr, originalErr) {
		t.Error("errors.Is should find the original error")
	}

	// Verify errors.As works
	var te *toolerr.Error
	if !errors.As(toolErr, &te) {
		t.Error("errors.As should work with toolerr.Error")
	}
	if te.Code != toolerr.ErrCodeNetworkError {
		t.Errorf("Expected code %s, got %s", toolerr.ErrCodeNetworkError, te.Code)
	}
}

// sampleTool is a minimal tool implementation demonstrating all SDK utility packages
type sampleTool struct {
	name        string
	description string
}

func (t *sampleTool) Name() string        { return t.name }
func (t *sampleTool) Description() string { return t.description }

func (t *sampleTool) Health(ctx context.Context) types.HealthStatus {
	// Use health package to check dependencies
	echoCheck := health.BinaryCheck("echo")
	sleepCheck := health.BinaryCheck("sleep")

	return health.Combine(echoCheck, sleepCheck)
}

func (t *sampleTool) Execute(ctx context.Context, inputs map[string]any) (map[string]any, error) {
	// Use input package to extract parameters with defaults
	message := input.GetString(inputs, "message", "default message")
	timeout := input.GetTimeout(inputs, "timeout", 30*time.Second)
	shouldSleep := input.GetBool(inputs, "sleep", false)
	command := input.GetString(inputs, "command", "echo")

	// Check if binary exists
	if !exec.BinaryExists(command) {
		return nil, toolerr.New(t.name, "execute", toolerr.ErrCodeBinaryNotFound,
			"required binary not found").
			WithDetails(map[string]any{"binary": command})
	}

	// Build command config
	var cfg exec.Config
	if shouldSleep {
		// Sleep command to test timeout
		cfg = exec.Config{
			Command: "sleep",
			Args:    []string{"10"},
			Timeout: timeout,
		}
	} else {
		cfg = exec.Config{
			Command: command,
			Args:    []string{message},
			Timeout: timeout,
		}
	}

	// Execute command
	result, err := exec.Run(ctx, cfg)
	if err != nil {
		// Check if it's a timeout by examining the error message
		// The exec package returns "command timed out after X" for timeouts
		errStr := err.Error()
		if contains(errStr, "timed out") || contains(errStr, "cancelled") {
			return nil, toolerr.New(t.name, "execute", toolerr.ErrCodeTimeout,
				"command execution timed out").
				WithCause(err).
				WithDetails(map[string]any{
					"timeout": timeout.String(),
					"command": cfg.Command,
				})
		}

		return nil, toolerr.New(t.name, "execute", toolerr.ErrCodeExecutionFailed,
			"command execution failed").
			WithCause(err)
	}

	// Check for non-zero exit code
	if result.ExitCode != 0 {
		return nil, toolerr.New(t.name, "execute", toolerr.ErrCodeExecutionFailed,
			"command exited with non-zero status").
			WithDetails(map[string]any{
				"exitCode": result.ExitCode,
				"stderr":   string(result.Stderr),
			})
	}

	return map[string]any{
		"output":   string(result.Stdout),
		"duration": result.Duration.String(),
	}, nil
}

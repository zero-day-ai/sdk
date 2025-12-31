// Package exec provides command execution utilities with timeout support.
// It wraps os/exec with a simple, context-aware API for executing shell commands
// and capturing their output.
package exec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// Config holds the configuration for command execution.
type Config struct {
	// Command is the name or path of the command to execute (required)
	Command string

	// Args are the command-line arguments (optional)
	Args []string

	// WorkDir is the working directory for the command (optional)
	WorkDir string

	// Env specifies the environment variables in "KEY=value" format (optional)
	// If nil, the command inherits the parent process environment
	Env []string

	// Timeout specifies the maximum execution duration (optional)
	// If zero, no timeout is enforced (uses parent context)
	Timeout time.Duration

	// StdinData is the data to send to the command's stdin (optional)
	StdinData []byte
}

// Result holds the result of command execution.
type Result struct {
	// Stdout contains the captured stdout
	Stdout []byte

	// Stderr contains the captured stderr
	Stderr []byte

	// ExitCode is the process exit code
	// 0 indicates success, non-zero indicates an error
	ExitCode int

	// Duration is the actual execution time
	Duration time.Duration
}

// Run executes a command with the given configuration.
// It returns a Result containing stdout, stderr, exit code, and duration.
//
// The function respects context cancellation and the configured timeout.
// If the command times out or the context is cancelled, the process is killed.
//
// A non-zero exit code is not treated as an error - the Result is returned
// with the exit code populated. This allows the caller to decide how to
// handle non-zero exits. Only actual execution failures (binary not found,
// permission denied, etc.) return an error.
//
// Example:
//
//	ctx := context.Background()
//	cfg := Config{
//		Command: "echo",
//		Args:    []string{"hello", "world"},
//		Timeout: 5 * time.Second,
//	}
//	result, err := Run(ctx, cfg)
//	if err != nil {
//		// Execution failed (binary not found, etc.)
//		return err
//	}
//	if result.ExitCode != 0 {
//		// Command ran but failed
//		return fmt.Errorf("command failed: %s", result.Stderr)
//	}
//	fmt.Println(string(result.Stdout))
func Run(ctx context.Context, cfg Config) (*Result, error) {
	if cfg.Command == "" {
		return nil, errors.New("command is required")
	}

	// Create context with timeout if specified
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	// Create command
	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)

	// Set working directory if specified
	if cfg.WorkDir != "" {
		cmd.Dir = cfg.WorkDir
	}

	// Set environment if specified
	if cfg.Env != nil {
		cmd.Env = cfg.Env
	}

	// Set up stdout and stderr capture
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set up stdin if provided
	if len(cfg.StdinData) > 0 {
		cmd.Stdin = bytes.NewReader(cfg.StdinData)
	}

	// Record start time
	start := time.Now()

	// Execute command
	err := cmd.Run()
	duration := time.Since(start)

	// Build result
	result := &Result{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: 0,
		Duration: duration,
	}

	// Extract exit code if available
	if err != nil {
		// Check for context errors first (timeout/cancellation)
		if ctx.Err() == context.DeadlineExceeded {
			return result, fmt.Errorf("command timed out after %v", cfg.Timeout)
		}

		if ctx.Err() == context.Canceled {
			return result, fmt.Errorf("command cancelled")
		}

		// Check for normal exit with non-zero code
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// Command ran but exited with non-zero code
			result.ExitCode = exitErr.ExitCode()
			return result, nil
		}

		// Other execution error (binary not found, permission denied, etc.)
		return result, fmt.Errorf("command execution failed: %w", err)
	}

	return result, nil
}

// BinaryExists checks if a binary exists in the system PATH.
// It returns true if the binary is found and executable, false otherwise.
//
// Example:
//
//	if !BinaryExists("docker") {
//		return errors.New("docker is not installed")
//	}
func BinaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// BinaryPath returns the full path to a binary in the system PATH.
// It returns an error if the binary is not found.
//
// Example:
//
//	path, err := BinaryPath("docker")
//	if err != nil {
//		return fmt.Errorf("docker not found: %w", err)
//	}
//	fmt.Printf("Docker is at: %s\n", path)
func BinaryPath(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("binary %q not found in PATH: %w", name, err)
	}
	return path, nil
}

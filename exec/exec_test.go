package exec

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRun_Success(t *testing.T) {
	tests := []struct {
		name           string
		cfg            Config
		expectedStdout string
		expectedStderr string
		expectedCode   int
	}{
		{
			name: "simple echo",
			cfg: Config{
				Command: "echo",
				Args:    []string{"hello", "world"},
			},
			expectedStdout: "hello world\n",
			expectedCode:   0,
		},
		{
			name: "echo without args",
			cfg: Config{
				Command: "echo",
			},
			expectedStdout: "\n",
			expectedCode:   0,
		},
		{
			name: "multiple args",
			cfg: Config{
				Command: "echo",
				Args:    []string{"-n", "no", "newline"},
			},
			expectedStdout: "no newline",
			expectedCode:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := Run(ctx, tt.cfg)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if result.ExitCode != tt.expectedCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedCode, result.ExitCode)
			}

			stdout := string(result.Stdout)
			if stdout != tt.expectedStdout {
				t.Errorf("expected stdout %q, got %q", tt.expectedStdout, stdout)
			}

			if result.Duration <= 0 {
				t.Error("expected positive duration")
			}
		})
	}
}

func TestRun_NonZeroExit(t *testing.T) {
	// Use sh with exit command to guarantee non-zero exit
	cfg := Config{
		Command: "sh",
		Args:    []string{"-c", "echo error message >&2; exit 42"},
	}

	ctx := context.Background()
	result, err := Run(ctx, cfg)

	// Non-zero exit should NOT return an error
	if err != nil {
		t.Fatalf("unexpected error for non-zero exit: %v", err)
	}

	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}

	stderr := string(result.Stderr)
	if !strings.Contains(stderr, "error message") {
		t.Errorf("expected stderr to contain 'error message', got %q", stderr)
	}
}

func TestRun_Timeout(t *testing.T) {
	// Use sleep command with timeout
	cfg := Config{
		Command: "sleep",
		Args:    []string{"10"}, // Sleep for 10 seconds
		Timeout: 100 * time.Millisecond,
	}

	ctx := context.Background()
	start := time.Now()
	result, err := Run(ctx, cfg)
	duration := time.Since(start)

	// Timeout should return an error
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("expected timeout error message, got: %v", err)
	}

	// Should complete quickly (within timeout + overhead)
	if duration > 2*time.Second {
		t.Errorf("timeout took too long: %v", duration)
	}

	// Result should still be returned with partial data
	if result == nil {
		t.Error("expected result even on timeout")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Start a long-running command
	cfg := Config{
		Command: "sleep",
		Args:    []string{"10"},
	}

	// Cancel context after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	result, err := Run(ctx, cfg)
	duration := time.Since(start)

	// Cancellation should return an error
	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}

	if !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("expected cancelled error message, got: %v", err)
	}

	// Should complete quickly
	if duration > 2*time.Second {
		t.Errorf("cancellation took too long: %v", duration)
	}

	// Result should still be returned
	if result == nil {
		t.Error("expected result even on cancellation")
	}
}

func TestRun_WithWorkDir(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Use pwd to verify working directory
	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "cd"
	} else {
		cmd = "pwd"
	}

	cfg := Config{
		Command: cmd,
		WorkDir: tmpDir,
	}

	ctx := context.Background()
	result, err := Run(ctx, cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("command failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	stdout := strings.TrimSpace(string(result.Stdout))
	if !strings.Contains(stdout, tmpDir) {
		t.Errorf("expected working dir %q in output, got %q", tmpDir, stdout)
	}
}

func TestRun_WithEnv(t *testing.T) {
	cfg := Config{
		Command: "sh",
		Args:    []string{"-c", "echo $TEST_VAR"},
		Env:     []string{"TEST_VAR=test_value"},
	}

	ctx := context.Background()
	result, err := Run(ctx, cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("command failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	stdout := strings.TrimSpace(string(result.Stdout))
	if stdout != "test_value" {
		t.Errorf("expected 'test_value', got %q", stdout)
	}
}

func TestRun_WithStdin(t *testing.T) {
	cfg := Config{
		Command:   "cat",
		StdinData: []byte("hello from stdin"),
	}

	ctx := context.Background()
	result, err := Run(ctx, cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 0 {
		t.Fatalf("command failed with exit code %d: %s", result.ExitCode, result.Stderr)
	}

	stdout := string(result.Stdout)
	if stdout != "hello from stdin" {
		t.Errorf("expected 'hello from stdin', got %q", stdout)
	}
}

func TestRun_BinaryNotFound(t *testing.T) {
	cfg := Config{
		Command: "this-binary-does-not-exist-12345",
	}

	ctx := context.Background()
	result, err := Run(ctx, cfg)

	// Binary not found should return an error
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}

	if !strings.Contains(err.Error(), "execution failed") {
		t.Errorf("expected 'execution failed' in error, got: %v", err)
	}

	// Result should still be returned
	if result == nil {
		t.Error("expected result even on error")
	}
}

func TestRun_EmptyCommand(t *testing.T) {
	cfg := Config{
		Command: "",
	}

	ctx := context.Background()
	result, err := Run(ctx, cfg)

	// Empty command should return an error
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}

	if !strings.Contains(err.Error(), "command is required") {
		t.Errorf("expected 'command is required' in error, got: %v", err)
	}

	if result != nil {
		t.Error("expected nil result for empty command")
	}
}

func TestRun_Duration(t *testing.T) {
	// Use a command that takes some time
	cfg := Config{
		Command: "sleep",
		Args:    []string{"0.1"}, // Sleep for 100ms
	}

	ctx := context.Background()
	result, err := Run(ctx, cfg)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Duration should be at least 100ms
	if result.Duration < 100*time.Millisecond {
		t.Errorf("expected duration >= 100ms, got %v", result.Duration)
	}

	// Duration should be reasonable (not more than 1 second)
	if result.Duration > 1*time.Second {
		t.Errorf("expected duration < 1s, got %v", result.Duration)
	}
}

func TestBinaryExists(t *testing.T) {
	tests := []struct {
		name     string
		binary   string
		expected bool
	}{
		{
			name:     "echo exists",
			binary:   "echo",
			expected: true,
		},
		{
			name:     "sh exists",
			binary:   "sh",
			expected: runtime.GOOS != "windows", // sh exists on Unix-like systems
		},
		{
			name:     "nonexistent binary",
			binary:   "this-binary-does-not-exist-12345",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := BinaryExists(tt.binary)
			if exists != tt.expected {
				t.Errorf("BinaryExists(%q) = %v, expected %v", tt.binary, exists, tt.expected)
			}
		})
	}
}

func TestBinaryPath(t *testing.T) {
	tests := []struct {
		name        string
		binary      string
		shouldExist bool
	}{
		{
			name:        "echo path",
			binary:      "echo",
			shouldExist: true,
		},
		{
			name:        "sh path",
			binary:      "sh",
			shouldExist: runtime.GOOS != "windows",
		},
		{
			name:        "nonexistent binary",
			binary:      "this-binary-does-not-exist-12345",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := BinaryPath(tt.binary)

			if tt.shouldExist {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if path == "" {
					t.Error("expected non-empty path")
				}
				// Verify the path exists
				if _, err := os.Stat(path); err != nil {
					t.Errorf("path %q does not exist: %v", path, err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected error for nonexistent binary, got path: %s", path)
				}
				if !strings.Contains(err.Error(), "not found") {
					t.Errorf("expected 'not found' in error, got: %v", err)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkRun_SimpleEcho(b *testing.B) {
	cfg := Config{
		Command: "echo",
		Args:    []string{"hello"},
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Run(ctx, cfg)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkBinaryExists(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BinaryExists("echo")
	}
}

func BenchmarkBinaryPath(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BinaryPath("echo")
	}
}

// Example tests (shown in godoc)
func ExampleRun() {
	ctx := context.Background()
	cfg := Config{
		Command: "echo",
		Args:    []string{"hello", "world"},
		Timeout: 5 * time.Second,
	}

	result, err := Run(ctx, cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Exit code: %d\n", result.ExitCode)
	fmt.Printf("Output: %s", result.Stdout)
	// Output:
	// Exit code: 0
	// Output: hello world
}

func ExampleBinaryExists() {
	if BinaryExists("echo") {
		fmt.Println("echo is available")
	} else {
		fmt.Println("echo is not available")
	}
	// Output: echo is available
}

func ExampleBinaryPath() {
	path, err := BinaryPath("echo")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("echo is at: %s\n", path)
}

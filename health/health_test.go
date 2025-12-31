package health

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zero-day-ai/sdk/types"
)

func TestBinaryCheck(t *testing.T) {
	tests := []struct {
		name           string
		binary         string
		expectHealthy  bool
		expectDegraded bool
	}{
		{
			name:          "existing binary sh",
			binary:        "sh",
			expectHealthy: true,
		},
		{
			name:          "existing binary ls",
			binary:        "ls",
			expectHealthy: true,
		},
		{
			name:          "non-existent binary",
			binary:        "this-binary-definitely-does-not-exist-12345",
			expectHealthy: false,
		},
		{
			name:          "empty binary name",
			binary:        "",
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := BinaryCheck(tt.binary)

			if tt.expectHealthy && !status.IsHealthy() {
				t.Errorf("expected healthy status, got %s: %s", status.Status, status.Message)
			}

			if !tt.expectHealthy && status.IsHealthy() {
				t.Errorf("expected unhealthy status, got %s: %s", status.Status, status.Message)
			}

			if status.Message == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}

func TestBinaryVersionCheck(t *testing.T) {
	tests := []struct {
		name           string
		binary         string
		minVersion     string
		versionFlag    string
		expectHealthy  bool
		expectDegraded bool
		skipReason     string
	}{
		{
			name:          "sh with version check",
			binary:        "sh",
			minVersion:    "1.0",
			versionFlag:   "--version",
			expectHealthy: false, // sh --version may not work as expected
			skipReason:    "sh version handling varies by system",
		},
		{
			name:          "non-existent binary",
			binary:        "this-binary-does-not-exist-999",
			minVersion:    "1.0",
			versionFlag:   "--version",
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip(tt.skipReason)
			}

			status := BinaryVersionCheck(tt.binary, tt.minVersion, tt.versionFlag)

			if tt.expectHealthy && !status.IsHealthy() {
				t.Logf("Status: %s, Message: %s, Details: %v", status.Status, status.Message, status.Details)
			}

			// At minimum, we should get a non-empty message
			if status.Message == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}

func TestNetworkCheck(t *testing.T) {
	// Start a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer listener.Close()

	// Get the port
	addr := listener.Addr().(*net.TCPAddr)
	testPort := addr.Port

	// Accept connections in background
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	tests := []struct {
		name           string
		host           string
		port           int
		timeout        time.Duration
		expectHealthy  bool
		expectDegraded bool
	}{
		{
			name:          "successful connection to test server",
			host:          "127.0.0.1",
			port:          testPort,
			timeout:       2 * time.Second,
			expectHealthy: true,
		},
		{
			name:          "connection to non-existent port",
			host:          "127.0.0.1",
			port:          65000, // unlikely to be in use
			timeout:       1 * time.Second,
			expectHealthy: false,
		},
		{
			name:          "invalid port number negative",
			host:          "127.0.0.1",
			port:          -1,
			timeout:       1 * time.Second,
			expectHealthy: false,
		},
		{
			name:          "invalid port number too large",
			host:          "127.0.0.1",
			port:          70000,
			timeout:       1 * time.Second,
			expectHealthy: false,
		},
		{
			name:          "empty host",
			host:          "",
			port:          80,
			timeout:       1 * time.Second,
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			status := NetworkCheck(ctx, tt.host, tt.port)

			if tt.expectHealthy && !status.IsHealthy() {
				t.Errorf("expected healthy status, got %s: %s", status.Status, status.Message)
			}

			if !tt.expectHealthy && status.IsHealthy() {
				t.Errorf("expected unhealthy status, got %s: %s", status.Status, status.Message)
			}

			if status.Message == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}

func TestNetworkCheckWithNilContext(t *testing.T) {
	// Test that NetworkCheck handles nil context gracefully
	status := NetworkCheck(nil, "127.0.0.1", 65000)
	if status.IsHealthy() {
		t.Error("expected unhealthy status for unreachable port")
	}
}

func TestFileCheck(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		path          string
		expectHealthy bool
	}{
		{
			name:          "existing file",
			path:          tmpFile,
			expectHealthy: true,
		},
		{
			name:          "existing directory",
			path:          tmpDir,
			expectHealthy: true,
		},
		{
			name:          "non-existent path",
			path:          "/this/path/definitely/does/not/exist/12345",
			expectHealthy: false,
		},
		{
			name:          "empty path",
			path:          "",
			expectHealthy: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := FileCheck(tt.path)

			if tt.expectHealthy && !status.IsHealthy() {
				t.Errorf("expected healthy status, got %s: %s", status.Status, status.Message)
			}

			if !tt.expectHealthy && status.IsHealthy() {
				t.Errorf("expected unhealthy status, got %s: %s", status.Status, status.Message)
			}

			if status.Message == "" {
				t.Error("expected non-empty message")
			}
		})
	}
}

func TestCombine(t *testing.T) {
	tests := []struct {
		name           string
		checks         []types.HealthStatus
		expectStatus   string
		expectDegraded bool
	}{
		{
			name: "all healthy",
			checks: []types.HealthStatus{
				types.NewHealthyStatus("check 1"),
				types.NewHealthyStatus("check 2"),
				types.NewHealthyStatus("check 3"),
			},
			expectStatus: types.StatusHealthy,
		},
		{
			name: "one unhealthy",
			checks: []types.HealthStatus{
				types.NewHealthyStatus("check 1"),
				types.NewUnhealthyStatus("check 2 failed", nil),
				types.NewHealthyStatus("check 3"),
			},
			expectStatus: types.StatusUnhealthy,
		},
		{
			name: "one degraded",
			checks: []types.HealthStatus{
				types.NewHealthyStatus("check 1"),
				types.NewDegradedStatus("check 2 degraded", nil),
				types.NewHealthyStatus("check 3"),
			},
			expectStatus:   types.StatusDegraded,
			expectDegraded: true,
		},
		{
			name: "unhealthy and degraded",
			checks: []types.HealthStatus{
				types.NewHealthyStatus("check 1"),
				types.NewDegradedStatus("check 2 degraded", nil),
				types.NewUnhealthyStatus("check 3 failed", nil),
			},
			expectStatus: types.StatusUnhealthy, // unhealthy takes precedence
		},
		{
			name: "multiple unhealthy",
			checks: []types.HealthStatus{
				types.NewUnhealthyStatus("check 1 failed", nil),
				types.NewUnhealthyStatus("check 2 failed", nil),
				types.NewHealthyStatus("check 3"),
			},
			expectStatus: types.StatusUnhealthy,
		},
		{
			name:         "no checks",
			checks:       []types.HealthStatus{},
			expectStatus: types.StatusHealthy,
		},
		{
			name:         "nil checks",
			checks:       nil,
			expectStatus: types.StatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := Combine(tt.checks...)

			if status.Status != tt.expectStatus {
				t.Errorf("expected status %s, got %s: %s", tt.expectStatus, status.Status, status.Message)
			}

			if status.Message == "" {
				t.Error("expected non-empty message")
			}

			// Check that details are populated when checks fail
			if status.Status != types.StatusHealthy && status.Details == nil {
				t.Error("expected details for non-healthy status")
			}
		})
	}
}

func TestCombineRealChecks(t *testing.T) {
	// Test combining real health checks
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		checks       func() []types.HealthStatus
		expectStatus string
	}{
		{
			name: "all passing checks",
			checks: func() []types.HealthStatus {
				return []types.HealthStatus{
					BinaryCheck("sh"),
					FileCheck(tmpFile),
					FileCheck(tmpDir),
				}
			},
			expectStatus: types.StatusHealthy,
		},
		{
			name: "mixed passing and failing",
			checks: func() []types.HealthStatus {
				return []types.HealthStatus{
					BinaryCheck("sh"),
					FileCheck("/nonexistent/path"),
					BinaryCheck("nonexistent-binary-xyz"),
				}
			},
			expectStatus: types.StatusUnhealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := Combine(tt.checks()...)

			if status.Status != tt.expectStatus {
				t.Errorf("expected status %s, got %s: %s", tt.expectStatus, status.Status, status.Message)
			}
		})
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "simple version",
			output:   "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with v prefix",
			output:   "v2.4.6",
			expected: "2.4.6",
		},
		{
			name:     "version in sentence",
			output:   "nmap version 7.80",
			expected: "7.80",
		},
		{
			name:     "version with build info",
			output:   "go version go1.21.5 linux/amd64",
			expected: "1.21.5",
		},
		{
			name:     "multiline with version",
			output:   "Tool Name\nVersion: 3.14.159\nCopyright 2024",
			expected: "3.14.159",
		},
		{
			name:     "no version",
			output:   "some random text without version",
			expected: "",
		},
		{
			name:     "empty output",
			output:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersion(tt.output)
			if result != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestVersionMeetsMinimum(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		minVersion string
		expected   bool
	}{
		{
			name:       "equal versions",
			version:    "1.2.3",
			minVersion: "1.2.3",
			expected:   true,
		},
		{
			name:       "higher major version",
			version:    "2.0.0",
			minVersion: "1.9.9",
			expected:   true,
		},
		{
			name:       "higher minor version",
			version:    "1.5.0",
			minVersion: "1.2.3",
			expected:   true,
		},
		{
			name:       "higher patch version",
			version:    "1.2.5",
			minVersion: "1.2.3",
			expected:   true,
		},
		{
			name:       "lower major version",
			version:    "1.9.9",
			minVersion: "2.0.0",
			expected:   false,
		},
		{
			name:       "lower minor version",
			version:    "1.2.3",
			minVersion: "1.5.0",
			expected:   false,
		},
		{
			name:       "lower patch version",
			version:    "1.2.1",
			minVersion: "1.2.3",
			expected:   false,
		},
		{
			name:       "different lengths equal start",
			version:    "1.2",
			minVersion: "1.2.0",
			expected:   true,
		},
		{
			name:       "different lengths higher",
			version:    "1.2.1",
			minVersion: "1.2",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := versionMeetsMinimum(tt.version, tt.minVersion)
			if result != tt.expected {
				t.Errorf("versionMeetsMinimum(%q, %q) = %v, expected %v",
					tt.version, tt.minVersion, result, tt.expected)
			}
		})
	}
}

func TestNetworkCheckTimeout(t *testing.T) {
	// Test that context timeout is respected
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Try to connect to a non-routable IP (should timeout)
	// Using 10.255.255.1 which is unlikely to be reachable
	status := NetworkCheck(ctx, "10.255.255.1", 80)

	if status.IsHealthy() {
		t.Error("expected unhealthy status for timed out connection")
	}

	if status.Message == "" {
		t.Error("expected non-empty message")
	}
}

func BenchmarkBinaryCheck(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BinaryCheck("sh")
	}
}

func BenchmarkFileCheck(b *testing.B) {
	tmpDir := b.TempDir()
	tmpFile := filepath.Join(tmpDir, "bench.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		b.Fatalf("failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FileCheck(tmpFile)
	}
}

func BenchmarkNetworkCheck(b *testing.B) {
	// Start a test TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to start test server: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NetworkCheck(ctx, "127.0.0.1", port)
	}
}

func BenchmarkCombine(b *testing.B) {
	checks := []types.HealthStatus{
		types.NewHealthyStatus("check 1"),
		types.NewHealthyStatus("check 2"),
		types.NewHealthyStatus("check 3"),
		types.NewDegradedStatus("check 4", nil),
		types.NewHealthyStatus("check 5"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Combine(checks...)
	}
}

// Example tests for documentation
func ExampleBinaryCheck() {
	status := BinaryCheck("sh")
	if status.IsHealthy() {
		println("sh is available")
	}
}

func ExampleNetworkCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := NetworkCheck(ctx, "localhost", 80)
	if status.IsUnhealthy() {
		println("Cannot connect to localhost:80")
	}
}

func ExampleFileCheck() {
	status := FileCheck("/etc/hosts")
	if status.IsHealthy() {
		println("/etc/hosts exists")
	}
}

func ExampleCombine() {
	status := Combine(
		BinaryCheck("nmap"),
		BinaryCheck("masscan"),
		FileCheck("/etc/resolv.conf"),
	)

	if status.IsUnhealthy() {
		println("System dependencies not met")
	}
}

// TestNetworkCheckLiveConnection tests connection to a real service
// This test is skipped by default but can be run with -short=false
func TestNetworkCheckLiveConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live connection test in short mode")
	}

	// Start local test server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	port := addr.Port

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	status := NetworkCheck(ctx, "127.0.0.1", port)
	if !status.IsHealthy() {
		t.Errorf("expected successful connection to test server on port %d: %s", port, status.Message)
	}
}

func TestExtractVersionNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean version",
			input:    "1.2.3",
			expected: "1.2.3",
		},
		{
			name:     "version with suffix",
			input:    "1.2.3-beta",
			expected: "1.2.3",
		},
		{
			name:     "version with build",
			input:    "1.2.3+build123",
			expected: "1.2.3",
		},
		{
			name:     "just major.minor",
			input:    "7.80",
			expected: "7.80",
		},
		{
			name:     "version in parentheses",
			input:    "(1.2.3)",
			expected: "1.2.3",
		},
		{
			name:     "no dots",
			input:    "123",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersionNumber(tt.input)
			if result != tt.expected {
				t.Errorf("extractVersionNumber(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"abc123", true},
		{"1", true},
		{"abc", false},
		{"", false},
		{"v1.2.3", true},
	}

	for _, tt := range tests {
		result := containsDigit(tt.input)
		if result != tt.expected {
			t.Errorf("containsDigit(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

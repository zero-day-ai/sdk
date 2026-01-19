// Package health provides reusable health check functions for Gibson tools.
// It offers standardized ways to verify dependencies, connectivity, and system state.
package health

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/zero-day-ai/sdk/types"
)

// BinaryCheck verifies that a binary exists and is executable in the system PATH.
// It returns a healthy status if the binary is found, unhealthy otherwise.
//
// Example:
//
//	status := health.BinaryCheck("nmap")
//	if status.IsUnhealthy() {
//	    log.Fatal("nmap is required but not installed")
//	}
func BinaryCheck(name string) types.HealthStatus {
	if name == "" {
		return types.NewUnhealthyStatus("binary name cannot be empty", nil)
	}

	path, err := exec.LookPath(name)
	if err != nil {
		return types.NewUnhealthyStatus(
			fmt.Sprintf("binary '%s' not found in PATH", name),
			map[string]any{
				"binary": name,
				"error":  err.Error(),
			},
		)
	}

	return types.NewHealthyStatus(
		fmt.Sprintf("binary '%s' found at %s", name, path),
	)
}

// BinaryVersionCheck verifies that a binary exists and meets a minimum version requirement.
// It executes the binary with the specified version flag (e.g., "--version") and parses the output.
// The version comparison is basic string-based and expects semver-like format (e.g., "1.2.3").
//
// Parameters:
//   - name: The binary name to check
//   - minVersion: The minimum required version (e.g., "2.0.0")
//   - versionFlag: The flag to get version info (e.g., "--version" or "-v")
//
// Example:
//
//	status := health.BinaryVersionCheck("nmap", "7.80", "--version")
//	if status.IsUnhealthy() {
//	    log.Fatal("nmap version 7.80 or higher is required")
//	}
func BinaryVersionCheck(name, minVersion, versionFlag string) types.HealthStatus {
	// First check if binary exists
	binaryStatus := BinaryCheck(name)
	if binaryStatus.IsUnhealthy() {
		return binaryStatus
	}

	if versionFlag == "" {
		versionFlag = "--version"
	}

	// Execute binary with version flag
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, versionFlag)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return types.NewUnhealthyStatus(
			fmt.Sprintf("failed to get version for '%s'", name),
			map[string]any{
				"binary": name,
				"error":  err.Error(),
				"output": string(output),
			},
		)
	}

	outputStr := string(output)
	version := parseVersion(outputStr)
	if version == "" {
		return types.NewDegradedStatus(
			fmt.Sprintf("could not parse version from '%s' output", name),
			map[string]any{
				"binary": name,
				"output": outputStr,
			},
		)
	}

	// Compare versions (basic semver comparison)
	if !versionMeetsMinimum(version, minVersion) {
		return types.NewUnhealthyStatus(
			fmt.Sprintf("binary '%s' version %s does not meet minimum requirement %s", name, version, minVersion),
			map[string]any{
				"binary":      name,
				"version":     version,
				"min_version": minVersion,
			},
		)
	}

	return types.NewHealthyStatus(
		fmt.Sprintf("binary '%s' version %s meets requirement %s", name, version, minVersion),
	)
}

// NetworkCheck verifies TCP connectivity to a host and port.
// It uses the provided context for timeout and cancellation control.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	status := health.NetworkCheck(ctx, "example.com", 443)
//	if status.IsUnhealthy() {
//	    log.Println("Cannot reach example.com:443")
//	}
func NetworkCheck(ctx context.Context, host string, port int) types.HealthStatus {
	if host == "" {
		return types.NewUnhealthyStatus("host cannot be empty", nil)
	}

	if port <= 0 || port > 65535 {
		return types.NewUnhealthyStatus(
			fmt.Sprintf("invalid port number: %d", port),
			map[string]any{"port": port},
		)
	}

	// Use context with timeout if not already set
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))
	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return types.NewUnhealthyStatus(
			fmt.Sprintf("failed to connect to %s", address),
			map[string]any{
				"host":  host,
				"port":  port,
				"error": err.Error(),
			},
		)
	}

	// Close connection immediately
	conn.Close()

	return types.NewHealthyStatus(
		fmt.Sprintf("successfully connected to %s", address),
	)
}

// FileCheck verifies that a file or directory exists at the specified path.
// It returns healthy if the path exists, unhealthy otherwise.
//
// Example:
//
//	status := health.FileCheck("/etc/hosts")
//	if status.IsUnhealthy() {
//	    log.Fatal("/etc/hosts does not exist")
//	}
func FileCheck(path string) types.HealthStatus {
	if path == "" {
		return types.NewUnhealthyStatus("path cannot be empty", nil)
	}

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return types.NewUnhealthyStatus(
				fmt.Sprintf("path '%s' does not exist", path),
				map[string]any{
					"path": path,
				},
			)
		}

		return types.NewUnhealthyStatus(
			fmt.Sprintf("failed to stat path '%s'", path),
			map[string]any{
				"path":  path,
				"error": err.Error(),
			},
		)
	}

	fileType := "file"
	if info.IsDir() {
		fileType = "directory"
	}

	return types.NewHealthyStatus(
		fmt.Sprintf("%s '%s' exists", fileType, path),
	)
}

// Combine aggregates multiple health checks into a single status.
// The result follows this priority:
//   - If any check is unhealthy, the result is unhealthy
//   - If any check is degraded (and none unhealthy), the result is degraded
//   - If all checks are healthy, the result is healthy
//
// Example:
//
//	status := health.Combine(
//	    health.BinaryCheck("nmap"),
//	    health.BinaryCheck("masscan"),
//	    health.FileCheck("/etc/resolv.conf"),
//	)
//	if status.IsUnhealthy() {
//	    log.Fatal("System dependencies not met")
//	}
func Combine(checks ...types.HealthStatus) types.HealthStatus {
	if len(checks) == 0 {
		return types.NewHealthyStatus("no checks provided")
	}

	var unhealthyChecks []string
	var degradedChecks []string
	var healthyCount int

	for _, check := range checks {
		switch check.Status {
		case types.StatusUnhealthy:
			msg := check.Message
			if msg == "" {
				msg = "unnamed check"
			}
			unhealthyChecks = append(unhealthyChecks, msg)
		case types.StatusDegraded:
			msg := check.Message
			if msg == "" {
				msg = "unnamed check"
			}
			degradedChecks = append(degradedChecks, msg)
		case types.StatusHealthy:
			healthyCount++
		}
	}

	// Return unhealthy if any check is unhealthy
	if len(unhealthyChecks) > 0 {
		return types.NewUnhealthyStatus(
			fmt.Sprintf("%d check(s) failed", len(unhealthyChecks)),
			map[string]any{
				"total":         len(checks),
				"unhealthy":     len(unhealthyChecks),
				"degraded":      len(degradedChecks),
				"healthy":       healthyCount,
				"failed_checks": unhealthyChecks,
			},
		)
	}

	// Return degraded if any check is degraded
	if len(degradedChecks) > 0 {
		return types.NewDegradedStatus(
			fmt.Sprintf("%d check(s) degraded", len(degradedChecks)),
			map[string]any{
				"total":           len(checks),
				"degraded":        len(degradedChecks),
				"healthy":         healthyCount,
				"degraded_checks": degradedChecks,
			},
		)
	}

	// All checks are healthy
	return types.NewHealthyStatus(
		fmt.Sprintf("all %d check(s) passed", len(checks)),
	)
}

// parseVersion extracts a version string from command output.
// It looks for common version patterns like "1.2.3" or "v1.2.3".
func parseVersion(output string) string {
	// Common version patterns
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for version patterns (e.g., "1.2.3", "v1.2.3", "version 1.2.3")
		fields := strings.Fields(line)
		for _, field := range fields {
			// Remove common prefixes
			field = strings.TrimPrefix(field, "v")
			field = strings.TrimPrefix(field, "V")

			// Check if it looks like a version (contains digits and dots)
			if strings.Contains(field, ".") && containsDigit(field) {
				// Extract version-like substring
				if version := extractVersionNumber(field); version != "" {
					return version
				}
			}
		}
	}

	return ""
}

// containsDigit checks if a string contains at least one digit.
func containsDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// extractVersionNumber extracts a semantic version number from a string.
// It handles formats like "1.2.3", "1.2.3-beta", "1.2.3+build", etc.
func extractVersionNumber(s string) string {
	var version strings.Builder
	dotCount := 0

	for i, c := range s {
		if c >= '0' && c <= '9' {
			version.WriteRune(c)
		} else if c == '.' && dotCount < 2 && i > 0 && version.Len() > 0 {
			version.WriteRune(c)
			dotCount++
		} else if version.Len() > 0 {
			// Stop at first non-version character after we've started
			break
		}
	}

	result := version.String()
	// Ensure version has at least one dot
	if strings.Contains(result, ".") && len(result) > 2 {
		return result
	}
	return ""
}

// versionMeetsMinimum performs basic semantic version comparison.
// Returns true if version >= minVersion.
func versionMeetsMinimum(version, minVersion string) bool {
	vParts := strings.Split(version, ".")
	minParts := strings.Split(minVersion, ".")

	// Compare each part
	maxLen := len(vParts)
	if len(minParts) > maxLen {
		maxLen = len(minParts)
	}

	for i := 0; i < maxLen; i++ {
		vPart := 0
		minPart := 0

		if i < len(vParts) {
			vPart, _ = strconv.Atoi(strings.TrimSpace(vParts[i]))
		}
		if i < len(minParts) {
			minPart, _ = strconv.Atoi(strings.TrimSpace(minParts[i]))
		}

		if vPart > minPart {
			return true
		} else if vPart < minPart {
			return false
		}
		// Continue if equal
	}

	return true // Equal versions
}

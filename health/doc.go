// Package health provides reusable health check functions for Gibson tools.
//
// This package offers standardized ways to verify dependencies, connectivity,
// and system state. It is designed to help tools implement consistent health
// checking patterns.
//
// # Health Check Functions
//
// The package provides five main health check functions:
//
//   - BinaryCheck: Verify a binary exists in PATH
//   - BinaryVersionCheck: Verify a binary meets minimum version requirements
//   - NetworkCheck: Verify TCP connectivity to a host:port
//   - FileCheck: Verify a file or directory exists
//   - Combine: Aggregate multiple health checks into a single status
//
// # Usage Example
//
//	import (
//	    "context"
//	    "time"
//	    "github.com/zero-day-ai/sdk/health"
//	)
//
//	// Check individual dependencies
//	nmapStatus := health.BinaryCheck("nmap")
//	if nmapStatus.IsUnhealthy() {
//	    log.Fatal("nmap is required but not installed")
//	}
//
//	// Check network connectivity
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	apiStatus := health.NetworkCheck(ctx, "api.example.com", 443)
//
//	// Combine multiple checks
//	overall := health.Combine(
//	    health.BinaryCheck("nmap"),
//	    health.BinaryCheck("masscan"),
//	    health.FileCheck("/etc/resolv.conf"),
//	    apiStatus,
//	)
//
//	if overall.IsUnhealthy() {
//	    log.Printf("Health check failed: %s", overall.Message)
//	    log.Printf("Details: %+v", overall.Details)
//	}
//
// # Health Status Priority
//
// When combining health checks with Combine(), the result follows this priority:
//
//   - Unhealthy: If any check is unhealthy, the combined result is unhealthy
//   - Degraded: If any check is degraded (and none unhealthy), the result is degraded
//   - Healthy: If all checks are healthy, the result is healthy
//
// # Context and Timeouts
//
// NetworkCheck accepts a context for timeout and cancellation control.
// If nil is passed, a default 5-second timeout is used.
//
// BinaryVersionCheck has a built-in 5-second timeout when executing
// binaries to check their version.
//
// # Version Comparison
//
// BinaryVersionCheck performs basic semantic version comparison.
// It supports common version formats like:
//
//   - "1.2.3"
//   - "v2.4.6"
//   - "nmap version 7.80"
//   - "go version go1.21.5 linux/amd64"
//
// Version comparison is done numerically on each segment (major.minor.patch).
package health

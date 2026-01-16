package toolerr_test

import (
	"fmt"

	"github.com/zero-day-ai/sdk/toolerr"
)

// Example_defaultHints demonstrates how default recovery hints are automatically
// registered and enriched when creating errors.
func Example_defaultHints() {
	// Create an error without manually adding hints
	err := toolerr.New("nmap", "scan", toolerr.ErrCodeBinaryNotFound, "nmap binary not found in PATH")

	// Enrich the error with default hints from the registry
	enriched := toolerr.EnrichError(err)

	// Display the enriched error
	fmt.Printf("Tool: %s\n", enriched.Tool)
	fmt.Printf("Error Code: %s\n", enriched.Code)
	fmt.Printf("Error Class: %s\n", enriched.Class)
	fmt.Printf("Number of Hints: %d\n", len(enriched.Hints))

	// Show the first recovery hint
	if len(enriched.Hints) > 0 {
		hint := enriched.Hints[0]
		fmt.Printf("\nRecovery Option:\n")
		fmt.Printf("  Strategy: %s\n", hint.Strategy)
		fmt.Printf("  Alternative: %s\n", hint.Alternative)
		fmt.Printf("  Confidence: %.1f\n", hint.Confidence)
		fmt.Printf("  Priority: %d\n", hint.Priority)
		fmt.Printf("  Reason: %s\n", hint.Reason)
	}

	// Output:
	// Tool: nmap
	// Error Code: BINARY_NOT_FOUND
	// Error Class: infrastructure
	// Number of Hints: 1
	//
	// Recovery Option:
	//   Strategy: use_alternative_tool
	//   Alternative: masscan
	//   Confidence: 0.8
	//   Priority: 1
	//   Reason: masscan provides similar port scanning capabilities with faster performance
}

// Example_nmapTimeout demonstrates recovery hints for nmap timeout errors.
func Example_nmapTimeout() {
	// Create a timeout error for nmap
	err := toolerr.New("nmap", "scan", toolerr.ErrCodeTimeout, "scan operation timed out")
	enriched := toolerr.EnrichError(err)

	fmt.Printf("Error: %s\n", enriched.Message)
	fmt.Printf("Class: %s\n", enriched.Class)
	fmt.Printf("\nRecovery Options:\n")

	for i, hint := range enriched.Hints {
		fmt.Printf("%d. [%s] %s (confidence: %.2f)\n",
			i+1, hint.Strategy, hint.Reason, hint.Confidence)
	}

	// Output:
	// Error: scan operation timed out
	// Class: transient
	//
	// Recovery Options:
	// 1. [modify_params] slower timing template (T2) and TCP connect scan reduce timeout risk on congested networks (confidence: 0.70)
	// 2. [modify_params] scanning only common ports significantly reduces scan time and timeout likelihood (confidence: 0.65)
	// 3. [use_alternative_tool] masscan is faster and less likely to timeout on large target ranges (confidence: 0.60)
}

// Example_masscanNotFound demonstrates how masscan errors suggest nmap as alternative.
func Example_masscanNotFound() {
	err := toolerr.New("masscan", "scan", toolerr.ErrCodeBinaryNotFound, "masscan not available")
	enriched := toolerr.EnrichError(err)

	if len(enriched.Hints) > 0 {
		hint := enriched.Hints[0]
		fmt.Printf("Try using %s instead: %s\n", hint.Alternative, hint.Reason)
	}

	// Output:
	// Try using nmap instead: nmap provides similar port scanning functionality with more features
}

// Example_nucleiDependencyMissing demonstrates nuclei template dependency errors.
func Example_nucleiDependencyMissing() {
	err := toolerr.New("nuclei", "scan", toolerr.ErrCodeDependencyMissing, "nuclei templates not found")
	enriched := toolerr.EnrichError(err)

	fmt.Printf("Error Class: %s\n", enriched.Class)
	if len(enriched.Hints) > 0 {
		hint := enriched.Hints[0]
		fmt.Printf("Suggested Action: %s\n", hint.Strategy)
		fmt.Printf("Reason: %s\n", hint.Reason)
	}

	// Output:
	// Error Class: infrastructure
	// Suggested Action: skip
	// Reason: nuclei templates may need to be downloaded or updated separately
}

// Example_subfinderToAmass demonstrates subdomain tool alternatives.
func Example_subfinderToAmass() {
	err := toolerr.New("subfinder", "enumerate", toolerr.ErrCodeBinaryNotFound, "subfinder not installed")
	enriched := toolerr.EnrichError(err)

	if len(enriched.Hints) > 0 {
		hint := enriched.Hints[0]
		fmt.Printf("Alternative: %s\n", hint.Alternative)
		fmt.Printf("Confidence: %.2f\n", hint.Confidence)
		fmt.Printf("Reason: %s\n", hint.Reason)
	}

	// Output:
	// Alternative: amass
	// Confidence: 0.85
	// Reason: amass provides comprehensive subdomain enumeration with additional features
}

// Example_genericNetworkError demonstrates that generic hints are tool-specific.
// Note: Generic hints are registered for "*" tool, not as fallbacks for unknown tools.
func Example_genericNetworkError() {
	// Create an error using the generic "*" tool identifier
	err := toolerr.New("*", "execute", toolerr.ErrCodeNetworkError, "connection refused")

	// Enrich will use the registered hints for "*" tool
	enriched := toolerr.EnrichError(err)

	fmt.Printf("Error Class: %s\n", enriched.Class)
	if len(enriched.Hints) > 0 {
		hint := enriched.Hints[0]
		fmt.Printf("Strategy: %s\n", hint.Strategy)
		fmt.Printf("Reason: %s\n", hint.Reason)
	}

	// Output:
	// Error Class: transient
	// Strategy: retry_with_backoff
	// Reason: network issues are often temporary and resolve within seconds
}

// Example_multipleHintsWithPriority demonstrates how hints are ordered by priority.
func Example_multipleHintsWithPriority() {
	err := toolerr.New("amass", "enumerate", toolerr.ErrCodeTimeout, "enumeration timed out")
	enriched := toolerr.EnrichError(err)

	fmt.Printf("Recovery hints (ordered by priority):\n")
	for _, hint := range enriched.Hints {
		fmt.Printf("Priority %d: %s\n", hint.Priority, hint.Reason)
	}

	// Output:
	// Recovery hints (ordered by priority):
	// Priority 1: passive enumeration mode is faster and less likely to timeout
	// Priority 2: limiting DNS recursion depth significantly reduces scan time
	// Priority 3: subfinder is generally faster for basic subdomain enumeration
}

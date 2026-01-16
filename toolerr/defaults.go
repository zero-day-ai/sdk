package toolerr

// This file registers default recovery hints for common Gibson tools.
// The init() function is called automatically when the package is imported,
// ensuring that all tools have sensible fallback options registered.

func init() {
	registerNmapHints()
	registerMasscanHints()
	registerNucleiHints()
	registerHttpxHints()
	registerSubfinderHints()
	registerAmassHints()
	registerGenericHints()
}

// registerNmapHints registers recovery hints for the nmap tool
func registerNmapHints() {
	// Binary not found - suggest alternative port scanners
	Register("nmap", ErrCodeBinaryNotFound,
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "masscan",
			Reason:      "masscan provides similar port scanning capabilities with faster performance",
			Confidence:  0.8,
			Priority:    1,
		},
	)

	// Timeout - suggest parameter adjustments
	Register("nmap", ErrCodeTimeout,
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"timing": 2, "scan_type": "connect"},
			Reason:      "slower timing template (T2) and TCP connect scan reduce timeout risk on congested networks",
			Confidence:  0.7,
			Priority:    1,
		},
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"ports": "22,80,443,8080"},
			Reason:      "scanning only common ports significantly reduces scan time and timeout likelihood",
			Confidence:  0.65,
			Priority:    2,
		},
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "masscan",
			Reason:      "masscan is faster and less likely to timeout on large target ranges",
			Confidence:  0.6,
			Priority:    3,
		},
	)

	// Permission denied - suggest non-privileged scan types
	Register("nmap", ErrCodePermissionDenied,
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"scan_type": "connect"},
			Reason:      "TCP connect scan (-sT) does not require root/administrator privileges",
			Confidence:  0.9,
			Priority:    1,
		},
	)

	// Network error - suggest retry with backoff
	Register("nmap", ErrCodeNetworkError,
		RecoveryHint{
			Strategy:    StrategyRetryWithBackoff,
			Reason:      "network errors are often transient and resolve after a brief delay",
			Confidence:  0.7,
			Priority:    1,
		},
	)
}

// registerMasscanHints registers recovery hints for the masscan tool
func registerMasscanHints() {
	// Binary not found - suggest nmap as alternative
	Register("masscan", ErrCodeBinaryNotFound,
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "nmap",
			Reason:      "nmap provides similar port scanning functionality with more features",
			Confidence:  0.8,
			Priority:    1,
		},
	)

	// Timeout - suggest rate limiting and alternative
	Register("masscan", ErrCodeTimeout,
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"rate": 100},
			Reason:      "reducing scan rate to 100 packets/sec prevents network congestion and timeouts",
			Confidence:  0.75,
			Priority:    1,
		},
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "nmap",
			Reason:      "nmap has better timeout handling and adaptive timing for challenging networks",
			Confidence:  0.6,
			Priority:    2,
		},
	)

	// Permission denied - masscan requires root
	Register("masscan", ErrCodePermissionDenied,
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "nmap",
			Reason:      "nmap can operate without privileges using TCP connect scan mode",
			Confidence:  0.85,
			Priority:    1,
		},
	)

	// Network error - suggest retry with backoff
	Register("masscan", ErrCodeNetworkError,
		RecoveryHint{
			Strategy:    StrategyRetryWithBackoff,
			Reason:      "network errors are often transient and resolve after a brief delay",
			Confidence:  0.7,
			Priority:    1,
		},
	)
}

// registerNucleiHints registers recovery hints for the nuclei tool
func registerNucleiHints() {
	// Binary not found - suggest alternatives
	Register("nuclei", ErrCodeBinaryNotFound,
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "nmap",
			Reason:      "nmap scripts (NSE) can detect some vulnerabilities, though less comprehensive than nuclei",
			Confidence:  0.5,
			Priority:    1,
		},
	)

	// Timeout - suggest rate limiting
	Register("nuclei", ErrCodeTimeout,
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"rate_limit": 50},
			Reason:      "reducing rate limit to 50 requests/second prevents overwhelming target and reduces timeouts",
			Confidence:  0.75,
			Priority:    1,
		},
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"severity": []string{"critical", "high"}},
			Reason:      "limiting to high-severity templates reduces scan time and timeout risk",
			Confidence:  0.65,
			Priority:    2,
		},
	)

	// Network error - suggest retry with backoff
	Register("nuclei", ErrCodeNetworkError,
		RecoveryHint{
			Strategy:    StrategyRetryWithBackoff,
			Reason:      "network errors are often transient and resolve after a brief delay",
			Confidence:  0.7,
			Priority:    1,
		},
	)

	// Dependency missing - nuclei templates may need updating
	Register("nuclei", ErrCodeDependencyMissing,
		RecoveryHint{
			Strategy:    StrategySkip,
			Reason:      "nuclei templates may need to be downloaded or updated separately",
			Confidence:  0.6,
			Priority:    1,
		},
	)
}

// registerHttpxHints registers recovery hints for the httpx tool
func registerHttpxHints() {
	// Binary not found - suggest alternatives
	Register("httpx", ErrCodeBinaryNotFound,
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "nmap",
			Reason:      "nmap can probe HTTP services though with less detail than httpx",
			Confidence:  0.6,
			Priority:    1,
		},
	)

	// Timeout - suggest parameter adjustments
	Register("httpx", ErrCodeTimeout,
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"timeout": "10s"},
			Reason:      "increasing timeout to 10 seconds allows slow-responding servers to reply",
			Confidence:  0.7,
			Priority:    1,
		},
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"follow_redirects": false},
			Reason:      "disabling redirect following reduces complexity and timeout risk",
			Confidence:  0.6,
			Priority:    2,
		},
	)

	// Network error - suggest retry with backoff
	Register("httpx", ErrCodeNetworkError,
		RecoveryHint{
			Strategy:    StrategyRetryWithBackoff,
			Reason:      "network errors are often transient and resolve after a brief delay",
			Confidence:  0.7,
			Priority:    1,
		},
	)
}

// registerSubfinderHints registers recovery hints for the subfinder tool
func registerSubfinderHints() {
	// Binary not found - suggest amass as alternative
	Register("subfinder", ErrCodeBinaryNotFound,
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "amass",
			Reason:      "amass provides comprehensive subdomain enumeration with additional features",
			Confidence:  0.85,
			Priority:    1,
		},
	)

	// Timeout - suggest limiting sources
	Register("subfinder", ErrCodeTimeout,
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"all": false},
			Reason:      "disabling all sources and using only fast sources reduces timeout risk",
			Confidence:  0.7,
			Priority:    1,
		},
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "amass",
			Reason:      "amass passive mode may complete faster on timeout-prone targets",
			Confidence:  0.6,
			Priority:    2,
		},
	)

	// Network error - suggest retry with backoff
	Register("subfinder", ErrCodeNetworkError,
		RecoveryHint{
			Strategy:    StrategyRetryWithBackoff,
			Reason:      "network errors are often transient and resolve after a brief delay",
			Confidence:  0.7,
			Priority:    1,
		},
	)
}

// registerAmassHints registers recovery hints for the amass tool
func registerAmassHints() {
	// Binary not found - suggest subfinder as alternative
	Register("amass", ErrCodeBinaryNotFound,
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "subfinder",
			Reason:      "subfinder provides fast passive subdomain enumeration",
			Confidence:  0.8,
			Priority:    1,
		},
	)

	// Timeout - suggest passive mode
	Register("amass", ErrCodeTimeout,
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"mode": "passive"},
			Reason:      "passive enumeration mode is faster and less likely to timeout",
			Confidence:  0.75,
			Priority:    1,
		},
		RecoveryHint{
			Strategy:    StrategyModifyParams,
			Params:      map[string]any{"max_depth": 1},
			Reason:      "limiting DNS recursion depth significantly reduces scan time",
			Confidence:  0.7,
			Priority:    2,
		},
		RecoveryHint{
			Strategy:    StrategyUseAlternative,
			Alternative: "subfinder",
			Reason:      "subfinder is generally faster for basic subdomain enumeration",
			Confidence:  0.65,
			Priority:    3,
		},
	)

	// Network error - suggest retry with backoff
	Register("amass", ErrCodeNetworkError,
		RecoveryHint{
			Strategy:    StrategyRetryWithBackoff,
			Reason:      "network errors are often transient and resolve after a brief delay",
			Confidence:  0.7,
			Priority:    1,
		},
	)
}

// registerGenericHints registers recovery hints that apply to all tools
func registerGenericHints() {
	// Note: These are fallback hints that apply when tool-specific hints aren't available.
	// The registry lookup is tool-specific, so we register these for common scenarios.

	// Generic timeout handling
	Register("*", ErrCodeTimeout,
		RecoveryHint{
			Strategy:    StrategyRetry,
			Reason:      "timeouts may be transient; a single retry often succeeds",
			Confidence:  0.6,
			Priority:    1,
		},
	)

	// Generic network error handling
	Register("*", ErrCodeNetworkError,
		RecoveryHint{
			Strategy:    StrategyRetryWithBackoff,
			Reason:      "network issues are often temporary and resolve within seconds",
			Confidence:  0.7,
			Priority:    1,
		},
	)

	// Generic execution failure
	Register("*", ErrCodeExecutionFailed,
		RecoveryHint{
			Strategy:    StrategyRetry,
			Reason:      "execution failures may be transient resource issues",
			Confidence:  0.5,
			Priority:    1,
		},
	)
}

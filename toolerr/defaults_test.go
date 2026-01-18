package toolerr

import (
	"testing"
)

// TestDefaultsRegistered verifies that default recovery hints are registered at init time
func TestDefaultsRegistered(t *testing.T) {
	tests := []struct {
		name      string
		tool      string
		errorCode string
		wantHints bool
	}{
		// nmap hints
		{
			name:      "nmap binary not found",
			tool:      "nmap",
			errorCode: ErrCodeBinaryNotFound,
			wantHints: true,
		},
		{
			name:      "nmap timeout",
			tool:      "nmap",
			errorCode: ErrCodeTimeout,
			wantHints: true,
		},
		{
			name:      "nmap permission denied",
			tool:      "nmap",
			errorCode: ErrCodePermissionDenied,
			wantHints: true,
		},
		{
			name:      "nmap network error",
			tool:      "nmap",
			errorCode: ErrCodeNetworkError,
			wantHints: true,
		},
		// masscan hints
		{
			name:      "masscan binary not found",
			tool:      "masscan",
			errorCode: ErrCodeBinaryNotFound,
			wantHints: true,
		},
		{
			name:      "masscan timeout",
			tool:      "masscan",
			errorCode: ErrCodeTimeout,
			wantHints: true,
		},
		{
			name:      "masscan permission denied",
			tool:      "masscan",
			errorCode: ErrCodePermissionDenied,
			wantHints: true,
		},
		// nuclei hints
		{
			name:      "nuclei binary not found",
			tool:      "nuclei",
			errorCode: ErrCodeBinaryNotFound,
			wantHints: true,
		},
		{
			name:      "nuclei timeout",
			tool:      "nuclei",
			errorCode: ErrCodeTimeout,
			wantHints: true,
		},
		{
			name:      "nuclei dependency missing",
			tool:      "nuclei",
			errorCode: ErrCodeDependencyMissing,
			wantHints: true,
		},
		// httpx hints
		{
			name:      "httpx binary not found",
			tool:      "httpx",
			errorCode: ErrCodeBinaryNotFound,
			wantHints: true,
		},
		{
			name:      "httpx timeout",
			tool:      "httpx",
			errorCode: ErrCodeTimeout,
			wantHints: true,
		},
		// subfinder hints
		{
			name:      "subfinder binary not found",
			tool:      "subfinder",
			errorCode: ErrCodeBinaryNotFound,
			wantHints: true,
		},
		{
			name:      "subfinder timeout",
			tool:      "subfinder",
			errorCode: ErrCodeTimeout,
			wantHints: true,
		},
		// amass hints
		{
			name:      "amass binary not found",
			tool:      "amass",
			errorCode: ErrCodeBinaryNotFound,
			wantHints: true,
		},
		{
			name:      "amass timeout",
			tool:      "amass",
			errorCode: ErrCodeTimeout,
			wantHints: true,
		},
		// generic hints
		{
			name:      "generic timeout",
			tool:      "*",
			errorCode: ErrCodeTimeout,
			wantHints: true,
		},
		{
			name:      "generic network error",
			tool:      "*",
			errorCode: ErrCodeNetworkError,
			wantHints: true,
		},
		{
			name:      "generic execution failed",
			tool:      "*",
			errorCode: ErrCodeExecutionFailed,
			wantHints: true,
		},
		// not registered cases
		{
			name:      "unknown tool",
			tool:      "unknown",
			errorCode: ErrCodeBinaryNotFound,
			wantHints: false,
		},
		{
			name:      "nmap parse error not registered",
			tool:      "nmap",
			errorCode: ErrCodeParseError,
			wantHints: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hints := GetHints(tt.tool, tt.errorCode)
			hasHints := len(hints) > 0

			if hasHints != tt.wantHints {
				t.Errorf("GetHints(%q, %q) returned hints=%v, want hints=%v",
					tt.tool, tt.errorCode, hasHints, tt.wantHints)
			}
		})
	}
}

// TestNmapAlternatives verifies specific nmap recovery hints
func TestNmapAlternatives(t *testing.T) {
	hints := GetHints("nmap", ErrCodeBinaryNotFound)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint for nmap binary not found, got %d", len(hints))
	}

	hint := hints[0]
	if hint.Strategy != StrategyUseAlternative {
		t.Errorf("expected strategy %q, got %q", StrategyUseAlternative, hint.Strategy)
	}
	if hint.Alternative != "masscan" {
		t.Errorf("expected alternative %q, got %q", "masscan", hint.Alternative)
	}
	if hint.Confidence < 0.5 || hint.Confidence > 1.0 {
		t.Errorf("expected confidence in range [0.5, 1.0], got %f", hint.Confidence)
	}
	if hint.Priority != 1 {
		t.Errorf("expected priority 1, got %d", hint.Priority)
	}
	if hint.Reason == "" {
		t.Error("expected non-empty reason")
	}
}

// TestMasscanAlternatives verifies specific masscan recovery hints
func TestMasscanAlternatives(t *testing.T) {
	hints := GetHints("masscan", ErrCodeBinaryNotFound)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint for masscan binary not found, got %d", len(hints))
	}

	hint := hints[0]
	if hint.Strategy != StrategyUseAlternative {
		t.Errorf("expected strategy %q, got %q", StrategyUseAlternative, hint.Strategy)
	}
	if hint.Alternative != "nmap" {
		t.Errorf("expected alternative %q, got %q", "nmap", hint.Alternative)
	}
}

// TestNucleiTimeoutHints verifies nuclei timeout recovery hints
func TestNucleiTimeoutHints(t *testing.T) {
	hints := GetHints("nuclei", ErrCodeTimeout)
	if len(hints) != 2 {
		t.Fatalf("expected 2 hints for nuclei timeout, got %d", len(hints))
	}

	// First hint should be rate limiting
	if hints[0].Strategy != StrategyModifyParams {
		t.Errorf("expected first hint strategy %q, got %q", StrategyModifyParams, hints[0].Strategy)
	}
	if hints[0].Params == nil {
		t.Error("expected params in first hint")
	}

	// Second hint should be severity filtering
	if hints[1].Strategy != StrategyModifyParams {
		t.Errorf("expected second hint strategy %q, got %q", StrategyModifyParams, hints[1].Strategy)
	}
}

// TestSubfinderAmassAlternatives verifies subfinder suggests amass
func TestSubfinderAmassAlternatives(t *testing.T) {
	hints := GetHints("subfinder", ErrCodeBinaryNotFound)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint for subfinder binary not found, got %d", len(hints))
	}

	hint := hints[0]
	if hint.Alternative != "amass" {
		t.Errorf("expected alternative %q, got %q", "amass", hint.Alternative)
	}
	if hint.Confidence < 0.8 {
		t.Errorf("expected high confidence >= 0.8, got %f", hint.Confidence)
	}
}

// TestAmassSubfinderAlternatives verifies amass suggests subfinder
func TestAmassSubfinderAlternatives(t *testing.T) {
	hints := GetHints("amass", ErrCodeBinaryNotFound)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint for amass binary not found, got %d", len(hints))
	}

	hint := hints[0]
	if hint.Alternative != "subfinder" {
		t.Errorf("expected alternative %q, got %q", "subfinder", hint.Alternative)
	}
}

// TestNmapPermissionDeniedHint verifies nmap permission denied hints
func TestNmapPermissionDeniedHint(t *testing.T) {
	hints := GetHints("nmap", ErrCodePermissionDenied)
	if len(hints) != 1 {
		t.Fatalf("expected 1 hint for nmap permission denied, got %d", len(hints))
	}

	hint := hints[0]
	if hint.Strategy != StrategyModifyParams {
		t.Errorf("expected strategy %q, got %q", StrategyModifyParams, hint.Strategy)
	}
	if hint.Confidence < 0.9 {
		t.Errorf("expected very high confidence >= 0.9, got %f", hint.Confidence)
	}

	// Should suggest connect scan
	params := hint.Params
	if params == nil {
		t.Fatal("expected params to be set")
	}
	if scanType, ok := params["scan_type"].(string); !ok || scanType != "connect" {
		t.Errorf("expected scan_type=connect in params, got %v", params)
	}
}

// TestConfidenceScores verifies all confidence scores are in valid range
func TestConfidenceScores(t *testing.T) {
	tools := []string{"nmap", "masscan", "nuclei", "httpx", "subfinder", "amass", "*"}
	errorCodes := []string{
		ErrCodeBinaryNotFound,
		ErrCodeTimeout,
		ErrCodePermissionDenied,
		ErrCodeNetworkError,
		ErrCodeDependencyMissing,
		ErrCodeExecutionFailed,
	}

	for _, tool := range tools {
		for _, code := range errorCodes {
			hints := GetHints(tool, code)
			for i, hint := range hints {
				if hint.Confidence < 0.0 || hint.Confidence > 1.0 {
					t.Errorf("%s/%s hint %d: confidence %f out of range [0.0, 1.0]",
						tool, code, i, hint.Confidence)
				}
				// Check that confidence is in realistic range (0.5-0.9 as per spec)
				if hint.Confidence < 0.5 || hint.Confidence > 0.9 {
					t.Errorf("%s/%s hint %d: confidence %f outside realistic range [0.5, 0.9]",
						tool, code, i, hint.Confidence)
				}
			}
		}
	}
}

// TestPriorityOrdering verifies hints have valid priority values
func TestPriorityOrdering(t *testing.T) {
	hints := GetHints("nmap", ErrCodeTimeout)
	if len(hints) < 2 {
		t.Skip("test requires multiple hints")
	}

	// Priorities should be sequential starting from 1
	for i, hint := range hints {
		expectedPriority := i + 1
		if hint.Priority != expectedPriority {
			t.Errorf("hint %d: expected priority %d, got %d", i, expectedPriority, hint.Priority)
		}
	}
}

// TestEnrichErrorWithDefaults verifies EnrichError uses default hints
func TestEnrichErrorWithDefaults(t *testing.T) {
	// Create an error without class or hints
	err := New("nmap", "scan", ErrCodeBinaryNotFound, "nmap not found")

	// Enrich it
	enriched := EnrichError(err)

	// Should have class set
	if enriched.Class == "" {
		t.Error("expected class to be set after enrichment")
	}
	if enriched.Class != ErrorClassInfrastructure {
		t.Errorf("expected class %q, got %q", ErrorClassInfrastructure, enriched.Class)
	}

	// Should have hints attached
	if len(enriched.Hints) == 0 {
		t.Error("expected hints to be attached after enrichment")
	}

	// Should have masscan alternative
	found := false
	for _, hint := range enriched.Hints {
		if hint.Alternative == "masscan" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find masscan alternative hint")
	}
}

// TestGenericHintsExist verifies generic fallback hints are registered
func TestGenericHintsExist(t *testing.T) {
	// Generic hints use "*" as the tool name
	hints := GetHints("*", ErrCodeTimeout)
	if len(hints) == 0 {
		t.Error("expected generic timeout hints to be registered")
	}

	hints = GetHints("*", ErrCodeNetworkError)
	if len(hints) == 0 {
		t.Error("expected generic network error hints to be registered")
	}

	hints = GetHints("*", ErrCodeExecutionFailed)
	if len(hints) == 0 {
		t.Error("expected generic execution failed hints to be registered")
	}
}

// TestAllHintsHaveReasons verifies every hint has a meaningful reason
func TestAllHintsHaveReasons(t *testing.T) {
	tools := []string{"nmap", "masscan", "nuclei", "httpx", "subfinder", "amass", "*"}
	errorCodes := []string{
		ErrCodeBinaryNotFound,
		ErrCodeTimeout,
		ErrCodePermissionDenied,
		ErrCodeNetworkError,
		ErrCodeDependencyMissing,
		ErrCodeExecutionFailed,
	}

	for _, tool := range tools {
		for _, code := range errorCodes {
			hints := GetHints(tool, code)
			for i, hint := range hints {
				if hint.Reason == "" {
					t.Errorf("%s/%s hint %d: missing reason", tool, code, i)
				}
				if len(hint.Reason) < 10 {
					t.Errorf("%s/%s hint %d: reason too short (%d chars): %q",
						tool, code, i, len(hint.Reason), hint.Reason)
				}
			}
		}
	}
}

// TestAlternativesExist verifies alternative tools actually exist in tools
func TestAlternativesExist(t *testing.T) {
	// Known tools that exist in tools
	knownTools := map[string]bool{
		"nmap":      true,
		"masscan":   true,
		"nuclei":    true,
		"httpx":     true,
		"subfinder": true,
		"amass":     true,
	}

	tools := []string{"nmap", "masscan", "nuclei", "httpx", "subfinder", "amass"}
	errorCodes := []string{
		ErrCodeBinaryNotFound,
		ErrCodeTimeout,
		ErrCodePermissionDenied,
		ErrCodeNetworkError,
	}

	for _, tool := range tools {
		for _, code := range errorCodes {
			hints := GetHints(tool, code)
			for _, hint := range hints {
				if hint.Alternative != "" && hint.Strategy == StrategyUseAlternative {
					// Verify alternative is a known tool
					if !knownTools[hint.Alternative] {
						t.Errorf("%s/%s suggests unknown alternative %q",
							tool, code, hint.Alternative)
					}
				}
			}
		}
	}
}

package result

import (
	"testing"
)

func TestResultQuality(t *testing.T) {
	tests := []struct {
		name     string
		quality  ResultQuality
		expected string
	}{
		{"Full quality", QualityFull, "full"},
		{"Partial quality", QualityPartial, "partial"},
		{"Empty quality", QualityEmpty, "empty"},
		{"Suspect quality", QualitySuspect, "suspect"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.quality) != tt.expected {
				t.Errorf("Quality = %v, want %v", tt.quality, tt.expected)
			}
		})
	}
}

func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator() returned nil")
	}
	if len(v.rules) == 0 {
		t.Error("NewValidator() should have default rules")
	}
	// Should have at least checkEmpty and checkAnomalies
	if len(v.rules) < 2 {
		t.Errorf("Expected at least 2 default rules, got %d", len(v.rules))
	}
}

func TestValidator_WithRules(t *testing.T) {
	v := NewValidator()
	initialRuleCount := len(v.rules)

	customRule := func(output map[string]any) (ResultQuality, float64, []string) {
		return QualityFull, 1.0, nil
	}

	v = v.WithRules(customRule)
	if len(v.rules) != initialRuleCount+1 {
		t.Errorf("Expected %d rules after adding custom rule, got %d", initialRuleCount+1, len(v.rules))
	}
}

func TestValidator_Validate_FullQuality(t *testing.T) {
	v := NewValidator()

	// Normal nmap-style output with hosts and ports
	output := map[string]any{
		"hosts": []any{
			map[string]any{
				"ip":    "192.168.1.1",
				"state": "up",
				"ports": []any{
					map[string]any{"port": 22, "state": "open"},
					map[string]any{"port": 80, "state": "open"},
				},
			},
		},
		"total_hosts":  1,
		"hosts_up":     1,
		"scan_time_ms": 1500,
	}

	result := v.Validate(output)

	if result.Quality != QualityFull {
		t.Errorf("Expected QualityFull, got %v", result.Quality)
	}
	if result.Confidence != 1.0 {
		t.Errorf("Expected confidence 1.0, got %v", result.Confidence)
	}
	if len(result.Warnings) > 0 {
		t.Errorf("Expected no warnings, got %v", result.Warnings)
	}
	if len(result.Suggestions) > 0 {
		t.Errorf("Expected no suggestions for full quality, got %v", result.Suggestions)
	}
	if result.Output == nil {
		t.Error("Expected output to be preserved")
	}
}

func TestValidator_Validate_EmptyHosts(t *testing.T) {
	v := NewValidator()

	output := map[string]any{
		"hosts":        []any{},
		"total_hosts":  0,
		"scan_time_ms": 1000,
	}

	result := v.Validate(output)

	if result.Quality != QualityEmpty {
		t.Errorf("Expected QualityEmpty, got %v", result.Quality)
	}
	if result.Confidence >= 1.0 {
		t.Errorf("Expected confidence < 1.0, got %v", result.Confidence)
	}
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for empty hosts")
	}
	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions for empty quality")
	}
}

func TestValidator_Validate_EmptyFindings(t *testing.T) {
	v := NewValidator()

	// Nuclei-style output with no findings
	output := map[string]any{
		"target":         "https://example.com",
		"findings":       []any{},
		"total_findings": 0,
		"scan_time_ms":   2000,
	}

	result := v.Validate(output)

	if result.Quality != QualityEmpty {
		t.Errorf("Expected QualityEmpty, got %v", result.Quality)
	}
	if result.Confidence >= 1.0 {
		t.Errorf("Expected confidence < 1.0, got %v", result.Confidence)
	}
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for empty findings")
	}
}

func TestValidator_Validate_EmptyResults(t *testing.T) {
	v := NewValidator()

	// Generic tool output with empty results
	output := map[string]any{
		"results":      []any{},
		"scan_time_ms": 500,
	}

	result := v.Validate(output)

	if result.Quality != QualityEmpty {
		t.Errorf("Expected QualityEmpty, got %v", result.Quality)
	}
	if result.Confidence >= 1.0 {
		t.Errorf("Expected confidence < 1.0, got %v", result.Confidence)
	}
}

func TestValidator_Validate_PartialQuality(t *testing.T) {
	v := NewValidator()

	// Hosts discovered but no ports found (incomplete scan)
	output := map[string]any{
		"hosts": []any{
			map[string]any{
				"ip":    "192.168.1.1",
				"state": "up",
				"ports": []any{},
			},
			map[string]any{
				"ip":    "192.168.1.2",
				"state": "up",
				"ports": []any{},
			},
		},
		"total_hosts":  2,
		"hosts_up":     2,
		"scan_time_ms": 1000,
	}

	result := v.Validate(output)

	if result.Quality != QualityPartial {
		t.Errorf("Expected QualityPartial, got %v", result.Quality)
	}
	if result.Confidence >= 1.0 {
		t.Errorf("Expected confidence < 1.0, got %v", result.Confidence)
	}
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for partial results")
	}
}

func TestValidator_Validate_SuspectFastScan(t *testing.T) {
	v := NewValidator()

	// Suspiciously fast scan time (< 100ms)
	output := map[string]any{
		"hosts": []any{
			map[string]any{
				"ip":    "192.168.1.1",
				"ports": []any{map[string]any{"port": 22}},
			},
		},
		"scan_time_ms": 50, // Anomalously fast
	}

	result := v.Validate(output)

	if result.Quality != QualitySuspect {
		t.Errorf("Expected QualitySuspect, got %v", result.Quality)
	}
	if result.Confidence >= 0.5 {
		t.Errorf("Expected low confidence, got %v", result.Confidence)
	}
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for suspect results")
	}
	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions for suspect quality")
	}
}

func TestValidator_Validate_SuspectInvalidPortCount(t *testing.T) {
	v := NewValidator()

	// Invalid port count (> 65535)
	output := map[string]any{
		"total_ports":  70000, // Invalid
		"scan_time_ms": 1000,
	}

	result := v.Validate(output)

	if result.Quality != QualitySuspect {
		t.Errorf("Expected QualitySuspect, got %v", result.Quality)
	}
	if result.Confidence >= 0.5 {
		t.Errorf("Expected low confidence, got %v", result.Confidence)
	}
}

func TestValidator_Validate_SuspectZeroScanRate(t *testing.T) {
	v := NewValidator()

	output := map[string]any{
		"hosts": []any{
			map[string]any{"ip": "192.168.1.1"},
		},
		"scan_rate":    0, // Suspicious
		"scan_time_ms": 1000,
	}

	result := v.Validate(output)

	if result.Quality != QualitySuspect {
		t.Errorf("Expected QualitySuspect, got %v", result.Quality)
	}
}

func TestValidator_Validate_PartialHostsUpZero(t *testing.T) {
	v := NewValidator()

	// Hosts discovered but none marked as "up"
	output := map[string]any{
		"hosts": []any{
			map[string]any{"ip": "192.168.1.1"},
		},
		"total_hosts":  1,
		"hosts_up":     0, // Suspicious
		"scan_time_ms": 1000,
	}

	result := v.Validate(output)

	if result.Quality != QualityPartial {
		t.Errorf("Expected QualityPartial, got %v", result.Quality)
	}
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for hosts_up being zero")
	}
}

func TestValidator_Validate_MultipleIssues(t *testing.T) {
	v := NewValidator()

	// Both fast scan AND empty results
	output := map[string]any{
		"hosts":        []any{},
		"scan_time_ms": 20,
	}

	result := v.Validate(output)

	// Should downgrade to the worst quality (Empty or Suspect)
	if result.Quality != QualityEmpty && result.Quality != QualitySuspect {
		t.Errorf("Expected QualityEmpty or QualitySuspect, got %v", result.Quality)
	}
	// Should have multiple warnings
	if len(result.Warnings) < 2 {
		t.Errorf("Expected at least 2 warnings, got %v", result.Warnings)
	}
	// Confidence should be low
	if result.Confidence >= 0.5 {
		t.Errorf("Expected low confidence, got %v", result.Confidence)
	}
}

func TestValidator_CustomRules(t *testing.T) {
	v := NewValidator()

	// Add a custom rule that checks for a specific field
	customRule := func(output map[string]any) (ResultQuality, float64, []string) {
		if _, ok := output["custom_field"]; !ok {
			return QualitySuspect, 0.5, []string{"Missing custom_field"}
		}
		return QualityFull, 1.0, nil
	}

	v = v.WithRules(customRule)

	// Test without custom field
	output1 := map[string]any{
		"hosts":        []any{map[string]any{"ip": "192.168.1.1"}},
		"scan_time_ms": 1000,
	}

	result1 := v.Validate(output1)
	if result1.Quality != QualitySuspect {
		t.Errorf("Expected QualitySuspect with missing custom_field, got %v", result1.Quality)
	}

	// Test with custom field
	output2 := map[string]any{
		"hosts":        []any{map[string]any{"ip": "192.168.1.1"}},
		"scan_time_ms": 1000,
		"custom_field": "present",
	}

	result2 := v.Validate(output2)
	if result2.Quality != QualityFull {
		t.Errorf("Expected QualityFull with custom_field, got %v", result2.Quality)
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"nil value", nil, true},
		{"empty slice", []any{}, true},
		{"empty array", [0]int{}, true},
		{"empty map", map[string]any{}, true},
		{"empty string", "", true},
		{"non-empty slice", []any{1}, false},
		{"non-empty map", map[string]any{"key": "value"}, false},
		{"non-empty string", "hello", false},
		{"zero int", 0, false},
		{"non-zero int", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmpty(tt.value)
			if result != tt.expected {
				t.Errorf("isEmpty(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestGetNumericValue(t *testing.T) {
	tests := []struct {
		name          string
		output        map[string]any
		key           string
		expectedValue float64
		expectedOk    bool
	}{
		{"int value", map[string]any{"key": 42}, "key", 42.0, true},
		{"int64 value", map[string]any{"key": int64(1000)}, "key", 1000.0, true},
		{"float64 value", map[string]any{"key": 3.14}, "key", 3.14, true},
		{"float32 value", map[string]any{"key": float32(2.71)}, "key", float64(float32(2.71)), true},
		{"missing key", map[string]any{}, "key", 0, false},
		{"string value", map[string]any{"key": "not a number"}, "key", 0, false},
		{"nil value", map[string]any{"key": nil}, "key", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, ok := getNumericValue(tt.output, tt.key)
			if ok != tt.expectedOk {
				t.Errorf("getNumericValue() ok = %v, want %v", ok, tt.expectedOk)
			}
			if ok && value != tt.expectedValue {
				t.Errorf("getNumericValue() value = %v, want %v", value, tt.expectedValue)
			}
		})
	}
}

func TestSuggestionsForQuality(t *testing.T) {
	tests := []struct {
		name             string
		quality          ResultQuality
		expectSuggestions bool
	}{
		{"Full quality - no suggestions", QualityFull, false},
		{"Empty quality - has suggestions", QualityEmpty, true},
		{"Partial quality - has suggestions", QualityPartial, true},
		{"Suspect quality - has suggestions", QualitySuspect, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := suggestionsForQuality(tt.quality)
			hasSuggestions := len(suggestions) > 0
			if hasSuggestions != tt.expectSuggestions {
				t.Errorf("suggestionsForQuality(%v) has suggestions = %v, want %v",
					tt.quality, hasSuggestions, tt.expectSuggestions)
			}
		})
	}
}

func TestShouldDowngradeQuality(t *testing.T) {
	tests := []struct {
		name     string
		current  ResultQuality
		candidate ResultQuality
		expected bool
	}{
		{"Full to Partial", QualityFull, QualityPartial, true},
		{"Full to Empty", QualityFull, QualityEmpty, true},
		{"Full to Suspect", QualityFull, QualitySuspect, true},
		{"Partial to Empty", QualityPartial, QualityEmpty, true},
		{"Partial to Suspect", QualityPartial, QualitySuspect, true},
		{"Empty to Suspect", QualityEmpty, QualitySuspect, true},
		{"Partial to Full", QualityPartial, QualityFull, false},
		{"Empty to Full", QualityEmpty, QualityFull, false},
		{"Suspect to Full", QualitySuspect, QualityFull, false},
		{"Full to Full", QualityFull, QualityFull, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldDowngradeQuality(tt.current, tt.candidate)
			if result != tt.expected {
				t.Errorf("shouldDowngradeQuality(%v, %v) = %v, want %v",
					tt.current, tt.candidate, result, tt.expected)
			}
		})
	}
}

// Benchmarks
func BenchmarkValidator_Validate(b *testing.B) {
	v := NewValidator()
	output := map[string]any{
		"hosts": []any{
			map[string]any{
				"ip":    "192.168.1.1",
				"ports": []any{map[string]any{"port": 22}},
			},
		},
		"scan_time_ms": 1500,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(output)
	}
}

func BenchmarkValidator_ValidateComplex(b *testing.B) {
	v := NewValidator()
	output := map[string]any{
		"hosts": []any{
			map[string]any{
				"ip":    "192.168.1.1",
				"state": "up",
				"ports": []any{
					map[string]any{"port": 22, "state": "open"},
					map[string]any{"port": 80, "state": "open"},
					map[string]any{"port": 443, "state": "open"},
				},
			},
			map[string]any{
				"ip":    "192.168.1.2",
				"state": "up",
				"ports": []any{
					map[string]any{"port": 22, "state": "open"},
				},
			},
		},
		"total_hosts":  2,
		"hosts_up":     2,
		"scan_time_ms": 1500,
		"scan_rate":    100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(output)
	}
}

// Test concurrent validation (race detector)
func TestValidator_ConcurrentValidation(t *testing.T) {
	v := NewValidator()
	output := map[string]any{
		"hosts": []any{
			map[string]any{"ip": "192.168.1.1"},
		},
		"scan_time_ms": 1000,
	}

	// Run validation concurrently to test for race conditions
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			v.Validate(output)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

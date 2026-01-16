package result

import (
	"fmt"
	"reflect"
)

// ResultQuality indicates the quality/completeness of results
type ResultQuality string

const (
	// QualityFull represents complete, meaningful results
	QualityFull ResultQuality = "full"
	// QualityPartial represents some results but incomplete
	QualityPartial ResultQuality = "partial"
	// QualityEmpty represents ran successfully but no findings
	QualityEmpty ResultQuality = "empty"
	// QualitySuspect represents results present but anomalous
	QualitySuspect ResultQuality = "suspect"
)

// ValidatedResult wraps tool output with quality assessment
type ValidatedResult struct {
	Output      map[string]any `json:"output"`
	Quality     ResultQuality  `json:"quality"`
	Confidence  float64        `json:"confidence"`    // 0.0-1.0
	Warnings    []string       `json:"warnings,omitempty"`
	Suggestions []string       `json:"suggestions,omitempty"`
}

// ValidationRule defines a function that validates tool output
// Returns: quality level, confidence score (0.0-1.0), and warnings
type ValidationRule func(output map[string]any) (ResultQuality, float64, []string)

// Validator validates tool results using configurable rules
type Validator struct {
	rules []ValidationRule
}

// NewValidator creates a validator with default rules
func NewValidator() *Validator {
	return &Validator{
		rules: []ValidationRule{
			checkEmpty,
			checkAnomalies,
		},
	}
}

// WithRules creates a validator with custom rules
func (v *Validator) WithRules(rules ...ValidationRule) *Validator {
	v.rules = append(v.rules, rules...)
	return v
}

// Validate assesses the quality of tool output
func (v *Validator) Validate(output map[string]any) *ValidatedResult {
	result := &ValidatedResult{
		Output:     output,
		Quality:    QualityFull,
		Confidence: 1.0,
	}

	for _, rule := range v.rules {
		quality, confidence, warnings := rule(output)

		// Downgrade quality if rule indicates issues
		// Quality ordering: Full > Partial > Empty/Suspect
		if shouldDowngradeQuality(result.Quality, quality) {
			result.Quality = quality
		}

		// Use lowest confidence score
		if confidence < result.Confidence {
			result.Confidence = confidence
		}

		// Accumulate warnings
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add suggestions based on quality
	result.Suggestions = suggestionsForQuality(result.Quality)

	return result
}

// shouldDowngradeQuality determines if quality should be downgraded
func shouldDowngradeQuality(current, candidate ResultQuality) bool {
	// Quality hierarchy: Full > Partial > Empty/Suspect
	qualityScore := map[ResultQuality]int{
		QualityFull:    4,
		QualityPartial: 3,
		QualityEmpty:   2,
		QualitySuspect: 1,
	}
	return qualityScore[candidate] < qualityScore[current]
}

// checkEmpty validates that output contains meaningful data
func checkEmpty(output map[string]any) (ResultQuality, float64, []string) {
	var warnings []string

	// Check for empty hosts array (common in network scanners)
	if hosts, ok := output["hosts"]; ok {
		if isEmpty(hosts) {
			warnings = append(warnings, "No hosts discovered - verify target is reachable")
			return QualityEmpty, 0.5, warnings
		}
	}

	// Check for empty findings array (common in vulnerability scanners)
	if findings, ok := output["findings"]; ok {
		if isEmpty(findings) {
			warnings = append(warnings, "No findings - target may be hardened or scan incomplete")
			return QualityEmpty, 0.6, warnings
		}
	}

	// Check for empty results array (generic case)
	if results, ok := output["results"]; ok {
		if isEmpty(results) {
			warnings = append(warnings, "No results returned - scan may have failed silently")
			return QualityEmpty, 0.5, warnings
		}
	}

	// Check for empty ports in hosts (partial results)
	if hosts, ok := output["hosts"].([]any); ok && len(hosts) > 0 {
		emptyPortCount := 0
		for _, h := range hosts {
			if host, ok := h.(map[string]any); ok {
				if ports, ok := host["ports"]; ok && isEmpty(ports) {
					emptyPortCount++
				}
			}
		}
		if emptyPortCount > 0 && emptyPortCount == len(hosts) {
			warnings = append(warnings, "Hosts discovered but no ports found - scan may be incomplete")
			return QualityPartial, 0.7, warnings
		}
	}

	return QualityFull, 1.0, nil
}

// checkAnomalies validates that results are not anomalous
func checkAnomalies(output map[string]any) (ResultQuality, float64, []string) {
	var warnings []string

	// Check for suspiciously fast scan times (less than 100ms)
	if scanTime, ok := getNumericValue(output, "scan_time_ms"); ok {
		if scanTime < 100 {
			warnings = append(warnings, fmt.Sprintf(
				"Scan completed unusually fast (%dms) - results may be incomplete",
				int(scanTime),
			))
			return QualitySuspect, 0.4, warnings
		}
	}

	// Check for unreasonably high port counts (possible parsing error)
	if totalPorts, ok := getNumericValue(output, "total_ports"); ok {
		if totalPorts > 65535 {
			warnings = append(warnings, fmt.Sprintf(
				"Port count (%d) exceeds valid range - possible parsing error",
				int(totalPorts),
			))
			return QualitySuspect, 0.3, warnings
		}
	}

	// Check for zero scan rate (indicates possible timeout or premature termination)
	if scanRate, ok := getNumericValue(output, "scan_rate"); ok {
		if scanRate == 0 {
			warnings = append(warnings, "Scan rate is zero - scan may have been interrupted")
			return QualitySuspect, 0.5, warnings
		}
	}

	// Check for hosts_up being zero when hosts exist
	if hostsUp, hasHostsUp := getNumericValue(output, "hosts_up"); hasHostsUp {
		if totalHosts, hasTotalHosts := getNumericValue(output, "total_hosts"); hasTotalHosts {
			if totalHosts > 0 && hostsUp == 0 {
				warnings = append(warnings, "No hosts marked as up despite hosts being discovered")
				return QualityPartial, 0.6, warnings
			}
		}
	}

	return QualityFull, 1.0, warnings
}

// suggestionsForQuality returns actionable suggestions based on quality
func suggestionsForQuality(quality ResultQuality) []string {
	switch quality {
	case QualityEmpty:
		return []string{
			"Verify target host is reachable and accessible",
			"Check if firewall or security rules are blocking scans",
			"Try scanning with different parameters or timing",
			"Confirm the target IP/domain is correct",
		}
	case QualityPartial:
		return []string{
			"Consider increasing scan timeout for more complete results",
			"Try running the scan again with broader parameters",
			"Check network conditions and bandwidth availability",
		}
	case QualitySuspect:
		return []string{
			"Review scan parameters and output carefully",
			"Re-run the scan to verify results",
			"Check tool logs for errors or warnings",
			"Consider using alternative tools to validate findings",
		}
	case QualityFull:
		return []string{}
	default:
		return []string{}
	}
}

// isEmpty checks if a value is empty (nil, empty array, or zero-length)
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		if val.IsNil() {
			return true
		}
		return isEmpty(val.Elem().Interface())
	default:
		return false
	}
}

// getNumericValue extracts a numeric value from output map, supporting int, int64, float64
func getNumericValue(output map[string]any, key string) (float64, bool) {
	v, ok := output[key]
	if !ok {
		return 0, false
	}

	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case float32:
		return float64(val), true
	default:
		return 0, false
	}
}

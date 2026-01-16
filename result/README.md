# Result Validator Package

The `result` package provides validation capabilities for tool outputs beyond simple success/failure. It catches false positives like empty results, anomalously fast scans, and other quality issues.

## Overview

Tool execution can succeed (exit code 0) but still produce problematic results:
- Empty host lists from network scans
- Zero findings from vulnerability scans
- Suspiciously fast completion times indicating incomplete execution
- Invalid data (e.g., port counts exceeding 65535)

The Result Validator assesses output quality and provides actionable suggestions for improvement.

## Usage

### Basic Validation

```go
import "github.com/zero-day-ai/sdk/result"

// Create validator with default rules
validator := result.NewValidator()

// Tool output from nmap, masscan, nuclei, etc.
output := map[string]any{
    "hosts": []any{
        map[string]any{
            "ip": "192.168.1.1",
            "ports": []any{
                map[string]any{"port": 22, "state": "open"},
            },
        },
    },
    "scan_time_ms": 1500,
}

// Validate the output
validated := validator.Validate(output)

fmt.Printf("Quality: %s\n", validated.Quality)
fmt.Printf("Confidence: %.2f\n", validated.Confidence)
for _, warning := range validated.Warnings {
    fmt.Printf("Warning: %s\n", warning)
}
```

### Custom Validation Rules

```go
// Add custom validation rules
customRule := func(output map[string]any) (result.ResultQuality, float64, []string) {
    if severity, ok := output["severity"]; ok && severity == "critical" {
        if count, ok := output["count"].(int); ok && count == 0 {
            return result.QualitySuspect, 0.5,
                []string{"Expected critical findings but got zero"}
        }
    }
    return result.QualityFull, 1.0, nil
}

validator := result.NewValidator().WithRules(customRule)
validated := validator.Validate(output)
```

## Quality Levels

### QualityFull
Complete, meaningful results with no anomalies detected.
- **Confidence**: 1.0
- **Suggestions**: None
- **Action**: Proceed with results

### QualityPartial
Some results but incomplete or with minor issues.
- **Confidence**: 0.6-0.7
- **Suggestions**: Increase timeout, broaden parameters
- **Action**: Consider re-running with adjusted parameters

### QualityEmpty
Execution succeeded but no meaningful data returned.
- **Confidence**: 0.5-0.6
- **Suggestions**: Verify target reachability, check firewall rules
- **Action**: Investigate why no results were found

### QualitySuspect
Results present but anomalous or potentially invalid.
- **Confidence**: 0.3-0.5
- **Suggestions**: Review carefully, re-run scan, use alternative tools
- **Action**: Validate results before trusting them

## Built-in Validation Rules

### Empty Check
Detects empty or missing data in common output fields:
- `hosts[]` - Network scan results
- `findings[]` - Vulnerability scan results
- `results[]` - Generic tool results
- `ports[]` within hosts - Incomplete port scans

### Anomaly Check
Detects suspicious patterns:
- **Fast scans**: `scan_time_ms < 100` suggests incomplete execution
- **Invalid ports**: `total_ports > 65535` indicates parsing errors
- **Zero rate**: `scan_rate == 0` suggests interruption
- **Inconsistent state**: Hosts exist but `hosts_up == 0`

## Common Output Patterns

The validator understands common output formats from:

### Network Scanners (nmap, masscan)
```go
{
    "hosts": []any{...},
    "total_hosts": 10,
    "hosts_up": 8,
    "scan_time_ms": 1500
}
```

### Vulnerability Scanners (nuclei)
```go
{
    "target": "https://example.com",
    "findings": []any{...},
    "total_findings": 5,
    "scan_time_ms": 2000
}
```

### Generic Tools
```go
{
    "results": []any{...},
    "scan_time_ms": 1000
}
```

## Performance

The validator is designed to be fast and non-intrusive:
- **Validation time**: ~400ns per output
- **Memory**: 1 allocation per validation (~80 bytes)
- **Concurrency**: Thread-safe, no shared state

Benchmark results:
```
BenchmarkValidator_Validate-8          	 3025726	       401.8 ns/op	      80 B/op	       1 allocs/op
BenchmarkValidator_ValidateComplex-8   	 2590450	       440.2 ns/op	      80 B/op	       1 allocs/op
```

## Integration Example

```go
func executeTool(ctx context.Context, tool tool.Tool, input map[string]any) (*result.ValidatedResult, error) {
    // Execute the tool
    output, err := tool.Execute(ctx, input)
    if err != nil {
        return nil, err
    }

    // Validate output quality
    validator := result.NewValidator()
    validated := validator.Validate(output)

    // Log warnings if present
    if len(validated.Warnings) > 0 {
        for _, warning := range validated.Warnings {
            log.Warn().Str("warning", warning).Msg("Output quality issue")
        }
    }

    // In debug mode, fail on suspect results
    if os.Getenv("GIBSON_RUN_MODE") == "debug" {
        if validated.Quality == result.QualitySuspect {
            return validated, fmt.Errorf("suspect results detected: %v", validated.Warnings)
        }
    }

    return validated, nil
}
```

## Design Principles

1. **Optional**: Validation is opt-in, tools don't need to use it
2. **Non-breaking**: Works with any `map[string]any` output format
3. **Fast**: Sub-microsecond validation with minimal allocations
4. **Extensible**: Custom rules can be added per-tool or per-domain
5. **Actionable**: Provides suggestions for improving results

## API Reference

### Types

```go
type ResultQuality string
const (
    QualityFull    ResultQuality = "full"
    QualityPartial ResultQuality = "partial"
    QualityEmpty   ResultQuality = "empty"
    QualitySuspect ResultQuality = "suspect"
)

type ValidatedResult struct {
    Output      map[string]any  // Original output
    Quality     ResultQuality   // Assessed quality level
    Confidence  float64         // 0.0-1.0
    Warnings    []string        // Issues detected
    Suggestions []string        // Actionable recommendations
}

type ValidationRule func(output map[string]any) (ResultQuality, float64, []string)
```

### Functions

```go
// NewValidator creates a validator with default rules
func NewValidator() *Validator

// WithRules adds custom validation rules
func (v *Validator) WithRules(rules ...ValidationRule) *Validator

// Validate assesses output quality
func (v *Validator) Validate(output map[string]any) *ValidatedResult
```

## Testing

Run tests with race detector:
```bash
cd sdk/result
go test -v -race
```

Run benchmarks:
```bash
go test -bench=. -benchmem
```

## See Also

- [toolerr Package](../toolerr/README.md) - Enhanced error types for recovery
- [Design Document](/.spec-workflow/specs/semantic-error-recovery/design.md) - Full architecture

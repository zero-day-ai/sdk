// Package input provides type-safe helpers for extracting values from map[string]any.
//
// This package is designed to simplify working with JSON unmarshaled data or configuration
// maps where types may vary (e.g., numbers as float64, int, or string). All functions
// gracefully handle type mismatches by returning sensible defaults rather than erroring.
//
// # Key Features
//
//   - Type-safe extraction with automatic coercion
//   - Nil-safe operations (handles nil maps and values)
//   - No panics or errors - always returns defaults on mismatch
//   - Zero allocations in hot paths for optimal performance
//   - Comprehensive handling of JSON unmarshaling quirks
//
// # Usage
//
// Extract values from a configuration map:
//
//	config := map[string]any{
//	    "host":    "example.com",
//	    "port":    8080,
//	    "timeout": "30s",
//	    "enabled": true,
//	    "tags":    []string{"web", "api"},
//	}
//
//	host := input.GetString(config, "host", "localhost")
//	port := input.GetInt(config, "port", 80)
//	timeout := input.GetTimeout(config, "timeout", 10*time.Second)
//	enabled := input.GetBool(config, "enabled", false)
//	tags := input.GetStringSlice(config, "tags")
//
// # Type Coercion
//
// The package handles common type coercion scenarios:
//
//   - GetInt: Handles int, int64, float64, and numeric strings
//   - GetFloat64: Handles float64, float32, int, int64, and numeric strings
//   - GetStringSlice: Handles []string, []interface{}, and single strings
//   - GetTimeout: Handles time.Duration, int (as seconds), and duration strings like "5m"
//
// # Design Philosophy
//
// This package follows the principle of "be liberal in what you accept" to handle
// real-world scenarios where data comes from JSON APIs, configuration files, or
// user input. Instead of strict type checking that would require error handling
// everywhere, it provides sensible defaults and automatic conversion.
//
// This makes tool development simpler and more robust, as tools don't need to
// worry about whether a number came in as int, int64, or float64 from JSON
// unmarshaling.
package input

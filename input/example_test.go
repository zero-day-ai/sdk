package input_test

import (
	"fmt"
	"time"

	"github.com/zero-day-ai/sdk/input"
)

// Example demonstrates basic usage of the input package helpers.
func Example() {
	// Simulate JSON unmarshaled into map[string]any
	config := map[string]any{
		"host":     "example.com",
		"port":     8080,  // int from JSON
		"timeout":  "30s", // string duration
		"retries":  3.0,   // float64 from JSON
		"enabled":  true,
		"tags":     []string{"web", "api"},
		"settings": map[string]any{"debug": true},
	}

	// Extract values with type-safe helpers
	host := input.GetString(config, "host", "localhost")
	port := input.GetInt(config, "port", 80)
	timeout := input.GetTimeout(config, "timeout", 10*time.Second)
	retries := input.GetInt(config, "retries", 1)
	enabled := input.GetBool(config, "enabled", false)
	tags := input.GetStringSlice(config, "tags")
	settings := input.GetMap(config, "settings")

	fmt.Printf("Host: %s\n", host)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Timeout: %v\n", timeout)
	fmt.Printf("Retries: %d\n", retries)
	fmt.Printf("Enabled: %t\n", enabled)
	fmt.Printf("Tags: %v\n", tags)
	fmt.Printf("Settings: %v\n", settings)

	// Output:
	// Host: example.com
	// Port: 8080
	// Timeout: 30s
	// Retries: 3
	// Enabled: true
	// Tags: [web api]
	// Settings: map[debug:true]
}

// ExampleGetTimeout demonstrates different timeout formats.
func ExampleGetTimeout() {
	// Different ways to specify timeout
	configs := []map[string]any{
		{"timeout": 30},               // int seconds
		{"timeout": "5m"},             // string duration
		{"timeout": 45 * time.Second}, // time.Duration
		{"timeout": "1h30m"},          // complex duration
		{"timeout": "not-a-duration"}, // invalid - uses default
		{},                            // missing - uses default
	}

	for _, config := range configs {
		timeout := input.GetTimeout(config, "timeout", 10*time.Second)
		fmt.Printf("%v -> %v\n", config["timeout"], timeout)
	}

	// Output:
	// 30 -> 30s
	// 5m -> 5m0s
	// 45s -> 45s
	// 1h30m -> 1h30m0s
	// not-a-duration -> 10s
	// <nil> -> 10s
}

// ExampleGetStringSlice demonstrates different slice input formats.
func ExampleGetStringSlice() {
	// Different ways to provide string slices
	configs := []map[string]any{
		{"items": []string{"a", "b", "c"}},         // direct []string
		{"items": []interface{}{"x", "y", "z"}},    // []interface{} from JSON
		{"items": []interface{}{"str", 123, true}}, // mixed types - converted to strings
		{"items": "single"},                        // single string - wrapped in slice
		{"items": []interface{}{"a", nil, "b"}},    // nil elements filtered out
	}

	for _, config := range configs {
		items := input.GetStringSlice(config, "items")
		fmt.Printf("%v\n", items)
	}

	// Output:
	// [a b c]
	// [x y z]
	// [str 123 true]
	// [single]
	// [a b]
}

// ExampleGetInt demonstrates type coercion for numeric values.
func ExampleGetInt() {
	// Different numeric formats that can be coerced to int
	config := map[string]any{
		"int_value":    42,
		"int64_value":  int64(100),
		"float_value":  123.5,
		"string_value": "456",
		"invalid":      "not-a-number",
	}

	fmt.Printf("int: %d\n", input.GetInt(config, "int_value", 0))
	fmt.Printf("int64: %d\n", input.GetInt(config, "int64_value", 0))
	fmt.Printf("float64: %d\n", input.GetInt(config, "float_value", 0))
	fmt.Printf("string: %d\n", input.GetInt(config, "string_value", 0))
	fmt.Printf("invalid: %d\n", input.GetInt(config, "invalid", 99))
	fmt.Printf("missing: %d\n", input.GetInt(config, "missing", 77))

	// Output:
	// int: 42
	// int64: 100
	// float64: 123
	// string: 456
	// invalid: 99
	// missing: 77
}

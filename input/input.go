// Package input provides type-safe helpers for extracting values from map[string]any.
//
// These functions are designed to handle JSON unmarshaling scenarios where types may
// vary (e.g., numbers as float64, int, or string). All functions return sensible
// defaults on type mismatch and handle nil maps gracefully.
package input

import (
	"fmt"
	"strconv"
	"time"
)

// GetString extracts a string value from the map with a default fallback.
// Returns defaultVal if the key doesn't exist, the value is nil, or not a string.
func GetString(m map[string]any, key string, defaultVal string) string {
	if m == nil {
		return defaultVal
	}

	val, ok := m[key]
	if !ok || val == nil {
		return defaultVal
	}

	str, ok := val.(string)
	if !ok {
		return defaultVal
	}

	return str
}

// GetInt extracts an int value from the map with type coercion and default fallback.
// Handles int, int64, float64, and string types.
// Returns defaultVal if the key doesn't exist, the value is nil, or cannot be converted.
func GetInt(m map[string]any, key string, defaultVal int) int {
	if m == nil {
		return defaultVal
	}

	val, ok := m[key]
	if !ok || val == nil {
		return defaultVal
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		// Try to parse string as integer
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
		return defaultVal
	default:
		return defaultVal
	}
}

// GetBool extracts a bool value from the map with a default fallback.
// Returns defaultVal if the key doesn't exist, the value is nil, or not a bool.
func GetBool(m map[string]any, key string, defaultVal bool) bool {
	if m == nil {
		return defaultVal
	}

	val, ok := m[key]
	if !ok || val == nil {
		return defaultVal
	}

	b, ok := val.(bool)
	if !ok {
		return defaultVal
	}

	return b
}

// GetFloat64 extracts a float64 value from the map with type coercion and default fallback.
// Handles float64, int, int64, and string types.
// Returns defaultVal if the key doesn't exist, the value is nil, or cannot be converted.
func GetFloat64(m map[string]any, key string, defaultVal float64) float64 {
	if m == nil {
		return defaultVal
	}

	val, ok := m[key]
	if !ok || val == nil {
		return defaultVal
	}

	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		// Try to parse string as float
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
		return defaultVal
	default:
		return defaultVal
	}
}

// GetStringSlice extracts a []string value from the map.
// Handles []string, []interface{} (converting each element to string), and single string values.
// Returns nil if the key doesn't exist, the value is nil, or cannot be converted.
func GetStringSlice(m map[string]any, key string) []string {
	if m == nil {
		return nil
	}

	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}

	// Handle []string directly
	if slice, ok := val.([]string); ok {
		return slice
	}

	// Handle []interface{} by converting each element
	if slice, ok := val.([]interface{}); ok {
		result := make([]string, 0, len(slice))
		for _, item := range slice {
			if item == nil {
				continue
			}
			// Convert each element to string
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	}

	// Handle single string by wrapping in slice
	if str, ok := val.(string); ok {
		return []string{str}
	}

	return nil
}

// GetMap extracts a nested map[string]any from the map.
// Returns nil if the key doesn't exist, the value is nil, or not a map.
func GetMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}

	val, ok := m[key]
	if !ok || val == nil {
		return nil
	}

	nested, ok := val.(map[string]any)
	if !ok {
		return nil
	}

	return nested
}

// GetTimeout extracts a duration value from the map with type coercion and default fallback.
// Handles int/int64 (interpreted as seconds), string (parsed as duration like "5m", "30s"),
// and time.Duration types.
// Returns defaultVal if the key doesn't exist, the value is nil, or cannot be converted.
func GetTimeout(m map[string]any, key string, defaultVal time.Duration) time.Duration {
	if m == nil {
		return defaultVal
	}

	val, ok := m[key]
	if !ok || val == nil {
		return defaultVal
	}

	switch v := val.(type) {
	case time.Duration:
		return v
	case int:
		// Interpret as seconds
		return time.Duration(v) * time.Second
	case int64:
		// Interpret as seconds
		return time.Duration(v) * time.Second
	case float64:
		// Interpret as seconds
		return time.Duration(v) * time.Second
	case string:
		// Try to parse as duration string
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
		// Try to parse as integer seconds
		if seconds, err := strconv.Atoi(v); err == nil {
			return time.Duration(seconds) * time.Second
		}
		return defaultVal
	default:
		return defaultVal
	}
}

// DefaultTimeout returns the default execution timeout (5 minutes).
// This is commonly used as a fallback when no timeout is specified.
func DefaultTimeout() time.Duration {
	return 5 * time.Minute
}

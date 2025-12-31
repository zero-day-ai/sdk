package input

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		defVal   string
		expected string
	}{
		{
			name:     "existing string value",
			m:        map[string]any{"key": "value"},
			key:      "key",
			defVal:   "default",
			expected: "value",
		},
		{
			name:     "missing key returns default",
			m:        map[string]any{"other": "value"},
			key:      "key",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "nil value returns default",
			m:        map[string]any{"key": nil},
			key:      "key",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "wrong type returns default",
			m:        map[string]any{"key": 123},
			key:      "key",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "nil map returns default",
			m:        nil,
			key:      "key",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "empty string value",
			m:        map[string]any{"key": ""},
			key:      "key",
			defVal:   "default",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetString(tt.m, tt.key, tt.defVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		defVal   int
		expected int
	}{
		{
			name:     "int value",
			m:        map[string]any{"key": 42},
			key:      "key",
			defVal:   0,
			expected: 42,
		},
		{
			name:     "int64 value",
			m:        map[string]any{"key": int64(100)},
			key:      "key",
			defVal:   0,
			expected: 100,
		},
		{
			name:     "float64 value",
			m:        map[string]any{"key": 123.5},
			key:      "key",
			defVal:   0,
			expected: 123,
		},
		{
			name:     "string value as number",
			m:        map[string]any{"key": "456"},
			key:      "key",
			defVal:   0,
			expected: 456,
		},
		{
			name:     "string value not a number",
			m:        map[string]any{"key": "not a number"},
			key:      "key",
			defVal:   99,
			expected: 99,
		},
		{
			name:     "missing key returns default",
			m:        map[string]any{"other": 42},
			key:      "key",
			defVal:   77,
			expected: 77,
		},
		{
			name:     "nil value returns default",
			m:        map[string]any{"key": nil},
			key:      "key",
			defVal:   88,
			expected: 88,
		},
		{
			name:     "wrong type returns default",
			m:        map[string]any{"key": true},
			key:      "key",
			defVal:   66,
			expected: 66,
		},
		{
			name:     "nil map returns default",
			m:        nil,
			key:      "key",
			defVal:   55,
			expected: 55,
		},
		{
			name:     "negative int",
			m:        map[string]any{"key": -42},
			key:      "key",
			defVal:   0,
			expected: -42,
		},
		{
			name:     "negative float64",
			m:        map[string]any{"key": -123.9},
			key:      "key",
			defVal:   0,
			expected: -123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInt(tt.m, tt.key, tt.defVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		defVal   bool
		expected bool
	}{
		{
			name:     "true value",
			m:        map[string]any{"key": true},
			key:      "key",
			defVal:   false,
			expected: true,
		},
		{
			name:     "false value",
			m:        map[string]any{"key": false},
			key:      "key",
			defVal:   true,
			expected: false,
		},
		{
			name:     "missing key returns default",
			m:        map[string]any{"other": true},
			key:      "key",
			defVal:   true,
			expected: true,
		},
		{
			name:     "nil value returns default",
			m:        map[string]any{"key": nil},
			key:      "key",
			defVal:   true,
			expected: true,
		},
		{
			name:     "wrong type returns default",
			m:        map[string]any{"key": "true"},
			key:      "key",
			defVal:   false,
			expected: false,
		},
		{
			name:     "nil map returns default",
			m:        nil,
			key:      "key",
			defVal:   true,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetBool(tt.m, tt.key, tt.defVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		defVal   float64
		expected float64
	}{
		{
			name:     "float64 value",
			m:        map[string]any{"key": 3.14},
			key:      "key",
			defVal:   0.0,
			expected: 3.14,
		},
		{
			name:     "float32 value",
			m:        map[string]any{"key": float32(2.5)},
			key:      "key",
			defVal:   0.0,
			expected: 2.5,
		},
		{
			name:     "int value",
			m:        map[string]any{"key": 42},
			key:      "key",
			defVal:   0.0,
			expected: 42.0,
		},
		{
			name:     "int64 value",
			m:        map[string]any{"key": int64(100)},
			key:      "key",
			defVal:   0.0,
			expected: 100.0,
		},
		{
			name:     "string value as number",
			m:        map[string]any{"key": "3.14159"},
			key:      "key",
			defVal:   0.0,
			expected: 3.14159,
		},
		{
			name:     "string value not a number",
			m:        map[string]any{"key": "not a number"},
			key:      "key",
			defVal:   99.9,
			expected: 99.9,
		},
		{
			name:     "missing key returns default",
			m:        map[string]any{"other": 3.14},
			key:      "key",
			defVal:   1.23,
			expected: 1.23,
		},
		{
			name:     "nil value returns default",
			m:        map[string]any{"key": nil},
			key:      "key",
			defVal:   4.56,
			expected: 4.56,
		},
		{
			name:     "wrong type returns default",
			m:        map[string]any{"key": true},
			key:      "key",
			defVal:   7.89,
			expected: 7.89,
		},
		{
			name:     "nil map returns default",
			m:        nil,
			key:      "key",
			defVal:   8.88,
			expected: 8.88,
		},
		{
			name:     "negative float64",
			m:        map[string]any{"key": -3.14},
			key:      "key",
			defVal:   0.0,
			expected: -3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFloat64(tt.m, tt.key, tt.defVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected []string
	}{
		{
			name:     "[]string value",
			m:        map[string]any{"key": []string{"a", "b", "c"}},
			key:      "key",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "[]interface{} value with strings",
			m:        map[string]any{"key": []interface{}{"x", "y", "z"}},
			key:      "key",
			expected: []string{"x", "y", "z"},
		},
		{
			name:     "[]interface{} value with mixed types",
			m:        map[string]any{"key": []interface{}{"string", 123, true}},
			key:      "key",
			expected: []string{"string", "123", "true"},
		},
		{
			name:     "[]interface{} with nil elements",
			m:        map[string]any{"key": []interface{}{"a", nil, "b"}},
			key:      "key",
			expected: []string{"a", "b"},
		},
		{
			name:     "single string value",
			m:        map[string]any{"key": "single"},
			key:      "key",
			expected: []string{"single"},
		},
		{
			name:     "missing key returns nil",
			m:        map[string]any{"other": []string{"a"}},
			key:      "key",
			expected: nil,
		},
		{
			name:     "nil value returns nil",
			m:        map[string]any{"key": nil},
			key:      "key",
			expected: nil,
		},
		{
			name:     "wrong type returns nil",
			m:        map[string]any{"key": 123},
			key:      "key",
			expected: nil,
		},
		{
			name:     "nil map returns nil",
			m:        nil,
			key:      "key",
			expected: nil,
		},
		{
			name:     "empty []string",
			m:        map[string]any{"key": []string{}},
			key:      "key",
			expected: []string{},
		},
		{
			name:     "empty []interface{}",
			m:        map[string]any{"key": []interface{}{}},
			key:      "key",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringSlice(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		expected map[string]any
	}{
		{
			name:     "nested map",
			m:        map[string]any{"key": map[string]any{"nested": "value"}},
			key:      "key",
			expected: map[string]any{"nested": "value"},
		},
		{
			name:     "missing key returns nil",
			m:        map[string]any{"other": map[string]any{"x": "y"}},
			key:      "key",
			expected: nil,
		},
		{
			name:     "nil value returns nil",
			m:        map[string]any{"key": nil},
			key:      "key",
			expected: nil,
		},
		{
			name:     "wrong type returns nil",
			m:        map[string]any{"key": "not a map"},
			key:      "key",
			expected: nil,
		},
		{
			name:     "nil map returns nil",
			m:        nil,
			key:      "key",
			expected: nil,
		},
		{
			name:     "empty nested map",
			m:        map[string]any{"key": map[string]any{}},
			key:      "key",
			expected: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMap(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]any
		key      string
		defVal   time.Duration
		expected time.Duration
	}{
		{
			name:     "time.Duration value",
			m:        map[string]any{"key": 5 * time.Second},
			key:      "key",
			defVal:   0,
			expected: 5 * time.Second,
		},
		{
			name:     "int value as seconds",
			m:        map[string]any{"key": 30},
			key:      "key",
			defVal:   0,
			expected: 30 * time.Second,
		},
		{
			name:     "int64 value as seconds",
			m:        map[string]any{"key": int64(60)},
			key:      "key",
			defVal:   0,
			expected: 60 * time.Second,
		},
		{
			name:     "float64 value as seconds",
			m:        map[string]any{"key": 45.5},
			key:      "key",
			defVal:   0,
			expected: 45 * time.Second,
		},
		{
			name:     "string duration format",
			m:        map[string]any{"key": "5m"},
			key:      "key",
			defVal:   0,
			expected: 5 * time.Minute,
		},
		{
			name:     "string duration with multiple units",
			m:        map[string]any{"key": "1h30m"},
			key:      "key",
			defVal:   0,
			expected: 90 * time.Minute,
		},
		{
			name:     "string numeric seconds",
			m:        map[string]any{"key": "120"},
			key:      "key",
			defVal:   0,
			expected: 120 * time.Second,
		},
		{
			name:     "string invalid format returns default",
			m:        map[string]any{"key": "invalid"},
			key:      "key",
			defVal:   10 * time.Second,
			expected: 10 * time.Second,
		},
		{
			name:     "missing key returns default",
			m:        map[string]any{"other": "5m"},
			key:      "key",
			defVal:   15 * time.Second,
			expected: 15 * time.Second,
		},
		{
			name:     "nil value returns default",
			m:        map[string]any{"key": nil},
			key:      "key",
			defVal:   20 * time.Second,
			expected: 20 * time.Second,
		},
		{
			name:     "wrong type returns default",
			m:        map[string]any{"key": true},
			key:      "key",
			defVal:   25 * time.Second,
			expected: 25 * time.Second,
		},
		{
			name:     "nil map returns default",
			m:        nil,
			key:      "key",
			defVal:   30 * time.Second,
			expected: 30 * time.Second,
		},
		{
			name:     "zero int",
			m:        map[string]any{"key": 0},
			key:      "key",
			defVal:   10 * time.Second,
			expected: 0,
		},
		{
			name:     "zero duration",
			m:        map[string]any{"key": time.Duration(0)},
			key:      "key",
			defVal:   10 * time.Second,
			expected: 0,
		},
		{
			name:     "string zero seconds",
			m:        map[string]any{"key": "0"},
			key:      "key",
			defVal:   10 * time.Second,
			expected: 0,
		},
		{
			name:     "string milliseconds",
			m:        map[string]any{"key": "500ms"},
			key:      "key",
			defVal:   0,
			expected: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTimeout(tt.m, tt.key, tt.defVal)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests to ensure no allocations in hot paths
func BenchmarkGetString(b *testing.B) {
	m := map[string]any{"key": "value"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetString(m, "key", "default")
	}
}

func BenchmarkGetInt(b *testing.B) {
	m := map[string]any{"key": 42}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetInt(m, "key", 0)
	}
}

func BenchmarkGetIntCoercion(b *testing.B) {
	m := map[string]any{"key": 42.5}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetInt(m, "key", 0)
	}
}

func BenchmarkGetStringSlice(b *testing.B) {
	m := map[string]any{"key": []string{"a", "b", "c"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetStringSlice(m, "key")
	}
}

func BenchmarkGetTimeout(b *testing.B) {
	m := map[string]any{"key": "5m"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetTimeout(m, "key", 0)
	}
}

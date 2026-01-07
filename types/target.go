package types

import "github.com/zero-day-ai/sdk/input"

// TargetType represents the category of target system being tested.
// Deprecated: Use string type for Target.Type field instead.
type TargetType string

// Target type constants define the supported categories of AI systems.
// Deprecated: These constants are kept for backward compatibility but new code
// should use string literals directly.
const (
	// TargetTypeLLMChat represents a conversational LLM interface (e.g., ChatGPT, Claude).
	TargetTypeLLMChat TargetType = "llm_chat"

	// TargetTypeLLMAPI represents a programmatic LLM API endpoint.
	TargetTypeLLMAPI TargetType = "llm_api"

	// TargetTypeRAG represents a Retrieval-Augmented Generation system.
	TargetTypeRAG TargetType = "rag"

	// TargetTypeAgent represents an autonomous AI agent system.
	TargetTypeAgent TargetType = "agent"

	// TargetTypeCopilot represents an AI coding assistant or copilot.
	TargetTypeCopilot TargetType = "copilot"
)

// String returns the string representation of the target type.
func (t TargetType) String() string {
	return string(t)
}

// IsValid returns true if the target type is a recognized value.
// Deprecated: Type validation is now handled via target schemas.
func (t TargetType) IsValid() bool {
	switch t {
	case TargetTypeLLMChat, TargetTypeLLMAPI, TargetTypeRAG, TargetTypeAgent, TargetTypeCopilot:
		return true
	default:
		return false
	}
}

// TargetInfo contains detailed information about a target system.
// It provides all necessary context for agents to interact with and test the target.
type TargetInfo struct {
	// ID is a unique identifier for the target.
	ID string `json:"id"`

	// Name is a human-readable name for the target.
	Name string `json:"name"`

	// Type categorizes the target system (e.g., "http_api", "kubernetes", "smart_contract").
	// Changed from TargetType enum to string for extensibility.
	Type string `json:"type"`

	// Provider identifies the vendor or service (e.g., "openai", "anthropic", "custom").
	Provider string `json:"provider,omitempty"`

	// Connection contains type-specific connection parameters.
	// For example, http_api targets use {"url": "...", "headers": {...}},
	// kubernetes targets use {"cluster": "...", "namespace": "..."},
	// smart_contract targets use {"chain": "...", "address": "..."}.
	Connection map[string]any `json:"connection,omitempty"`

	// DeprecatedURL is the legacy URL field.
	// Deprecated: Use Connection["url"] instead. Kept for backward compatibility.
	// Will be removed in a future version.
	DeprecatedURL string `json:"url,omitempty"`

	// DeprecatedHeaders are the legacy headers.
	// Deprecated: Use Connection["headers"] instead. Kept for backward compatibility.
	// Will be removed in a future version.
	DeprecatedHeaders map[string]string `json:"headers,omitempty"`

	// Metadata stores additional target-specific information and context.
	// This can include model versions, capabilities, rate limits, etc.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Validate checks if the TargetInfo has all required fields.
// Supports both legacy URL-based targets and new Connection-based targets.
func (t *TargetInfo) Validate() error {
	if t.ID == "" {
		return &ValidationError{Field: "ID", Message: "target ID is required"}
	}

	if t.Name == "" {
		return &ValidationError{Field: "Name", Message: "target name is required"}
	}

	if t.Type == "" {
		return &ValidationError{Field: "Type", Message: "target type is required"}
	}

	// Check either new Connection map or deprecated URL field
	hasConnection := len(t.Connection) > 0
	hasURL := t.DeprecatedURL != ""

	if !hasConnection && !hasURL {
		return &ValidationError{Field: "URL/Connection", Message: "target must have either URL or Connection parameters"}
	}

	return nil
}

// URL returns the URL for backward compatibility.
// If Connection["url"] exists, returns it as a string.
// Otherwise, returns the deprecated DeprecatedURL field.
// This method provides a consistent way to retrieve the URL regardless of
// whether the target uses the old URL field or new Connection map.
func (t *TargetInfo) URL() string {
	if t.Connection != nil {
		if url := input.GetString(t.Connection, "url", ""); url != "" {
			return url
		}
	}
	return t.DeprecatedURL
}

// GetConnection retrieves a connection parameter value by key.
// Returns the value and true if the key exists, nil and false otherwise.
func (t *TargetInfo) GetConnection(key string) (any, bool) {
	if t.Connection == nil {
		return nil, false
	}
	val, ok := t.Connection[key]
	return val, ok
}

// GetConnectionString retrieves a string connection parameter by key.
// Returns the string value or empty string if the key doesn't exist or is not a string.
// This is a convenience wrapper around GetConnection for common string parameters.
func (t *TargetInfo) GetConnectionString(key string) string {
	return input.GetString(t.Connection, key, "")
}

// GetHeader retrieves a header value by key, returns empty string if not found.
// First checks Connection["headers"], then falls back to deprecated DeprecatedHeaders field.
func (t *TargetInfo) GetHeader(key string) string {
	// Check Connection["headers"] first
	if t.Connection != nil {
		if headers := input.GetMap(t.Connection, "headers"); headers != nil {
			if val, ok := headers[key].(string); ok {
				return val
			}
		}
	}
	// Fall back to deprecated field
	if t.DeprecatedHeaders == nil {
		return ""
	}
	return t.DeprecatedHeaders[key]
}

// SetHeader sets a header value in Connection["headers"].
// Also updates deprecated DeprecatedHeaders field for backward compatibility.
func (t *TargetInfo) SetHeader(key, value string) {
	// Ensure Connection map exists
	if t.Connection == nil {
		t.Connection = make(map[string]any)
	}
	// Ensure headers map exists in Connection
	headers, _ := t.Connection["headers"].(map[string]any)
	if headers == nil {
		headers = make(map[string]any)
		t.Connection["headers"] = headers
	}
	headers[key] = value

	// Also update deprecated field for backward compatibility
	if t.DeprecatedHeaders == nil {
		t.DeprecatedHeaders = make(map[string]string)
	}
	t.DeprecatedHeaders[key] = value
}

// GetMetadata retrieves a metadata value by key.
func (t *TargetInfo) GetMetadata(key string) (any, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	val, ok := t.Metadata[key]
	return val, ok
}

// SetMetadata sets a metadata value.
func (t *TargetInfo) SetMetadata(key string, value any) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]any)
	}
	t.Metadata[key] = value
}

// ValidationError represents a validation error for target info.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

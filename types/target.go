package types

// TargetType represents the category of target system being tested.
type TargetType string

// Target type constants define the supported categories of AI systems.
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

	// URL is the endpoint or interface URL for the target.
	URL string `json:"url"`

	// Type categorizes the target system.
	Type TargetType `json:"type"`

	// Provider identifies the vendor or service (e.g., "openai", "anthropic", "custom").
	Provider string `json:"provider,omitempty"`

	// Headers contains HTTP headers required for authentication or configuration.
	Headers map[string]string `json:"headers,omitempty"`

	// Metadata stores additional target-specific information and context.
	// This can include model versions, capabilities, rate limits, etc.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Validate checks if the TargetInfo has all required fields.
func (t *TargetInfo) Validate() error {
	if t.ID == "" {
		return &ValidationError{Field: "ID", Message: "target ID is required"}
	}

	if t.Name == "" {
		return &ValidationError{Field: "Name", Message: "target name is required"}
	}

	if t.URL == "" {
		return &ValidationError{Field: "URL", Message: "target URL is required"}
	}

	if !t.Type.IsValid() {
		return &ValidationError{Field: "Type", Message: "invalid target type"}
	}

	return nil
}

// GetHeader retrieves a header value by key, returns empty string if not found.
func (t *TargetInfo) GetHeader(key string) string {
	if t.Headers == nil {
		return ""
	}
	return t.Headers[key]
}

// SetHeader sets a header value.
func (t *TargetInfo) SetHeader(key, value string) {
	if t.Headers == nil {
		t.Headers = make(map[string]string)
	}
	t.Headers[key] = value
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

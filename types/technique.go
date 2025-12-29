package types

// TechniqueType represents a category of attack or testing technique.
type TechniqueType string

// Technique type constants define common AI security testing techniques.
const (
	// TechniquePromptInjection represents attempts to inject malicious instructions into prompts.
	// Example: Appending "Ignore previous instructions and..." to manipulate model behavior.
	TechniquePromptInjection TechniqueType = "prompt_injection"

	// TechniqueJailbreak represents attempts to bypass safety guardrails and restrictions.
	// Example: Role-playing scenarios, hypothetical questions, or encoded instructions.
	TechniqueJailbreak TechniqueType = "jailbreak"

	// TechniqueDataExtraction represents attempts to extract training data or sensitive information.
	// Example: Probing for memorized content, PII, or proprietary data.
	TechniqueDataExtraction TechniqueType = "data_extraction"

	// TechniqueModelManipulation represents attempts to alter model behavior or outputs.
	// Example: Bias exploitation, output formatting attacks, or context window manipulation.
	TechniqueModelManipulation TechniqueType = "model_manipulation"

	// TechniqueDOS represents denial-of-service attacks against AI systems.
	// Example: Resource exhaustion through expensive queries or infinite loops.
	TechniqueDOS TechniqueType = "dos"
)

// String returns the string representation of the technique type.
func (t TechniqueType) String() string {
	return string(t)
}

// IsValid returns true if the technique type is a recognized value.
func (t TechniqueType) IsValid() bool {
	switch t {
	case TechniquePromptInjection, TechniqueJailbreak, TechniqueDataExtraction,
		TechniqueModelManipulation, TechniqueDOS:
		return true
	default:
		return false
	}
}

// Description returns a human-readable description of the technique type.
func (t TechniqueType) Description() string {
	switch t {
	case TechniquePromptInjection:
		return "Prompt injection attacks that attempt to inject malicious instructions"
	case TechniqueJailbreak:
		return "Jailbreak attempts to bypass safety guardrails and restrictions"
	case TechniqueDataExtraction:
		return "Data extraction techniques to retrieve sensitive or training data"
	case TechniqueModelManipulation:
		return "Model manipulation to alter behavior or outputs"
	case TechniqueDOS:
		return "Denial-of-service attacks to exhaust resources or cause failures"
	default:
		return "Unknown technique type"
	}
}

// Severity returns the default severity level for this technique type.
// This can be overridden based on specific findings.
func (t TechniqueType) Severity() string {
	switch t {
	case TechniquePromptInjection:
		return "high"
	case TechniqueJailbreak:
		return "high"
	case TechniqueDataExtraction:
		return "critical"
	case TechniqueModelManipulation:
		return "medium"
	case TechniqueDOS:
		return "high"
	default:
		return "unknown"
	}
}

// TechniqueInfo provides detailed information about a specific technique.
type TechniqueInfo struct {
	// Type is the category of technique.
	Type TechniqueType `json:"type"`

	// Name is a human-readable name for the specific technique variant.
	Name string `json:"name"`

	// Description provides details about what the technique does and how it works.
	Description string `json:"description,omitempty"`

	// Tags categorize the technique for filtering and organization.
	Tags []string `json:"tags,omitempty"`

	// Metadata stores additional technique-specific information.
	// This can include success rates, known bypasses, references, etc.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Validate checks if the TechniqueInfo has all required fields.
func (t *TechniqueInfo) Validate() error {
	if !t.Type.IsValid() {
		return &ValidationError{Field: "Type", Message: "invalid technique type"}
	}

	if t.Name == "" {
		return &ValidationError{Field: "Name", Message: "technique name is required"}
	}

	return nil
}

// HasTag returns true if the technique has the specified tag.
func (t *TechniqueInfo) HasTag(tag string) bool {
	for _, t := range t.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag to the technique if it doesn't already exist.
func (t *TechniqueInfo) AddTag(tag string) {
	if !t.HasTag(tag) {
		t.Tags = append(t.Tags, tag)
	}
}

// GetMetadata retrieves a metadata value by key.
func (t *TechniqueInfo) GetMetadata(key string) (any, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	val, ok := t.Metadata[key]
	return val, ok
}

// SetMetadata sets a metadata value.
func (t *TechniqueInfo) SetMetadata(key string, value any) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]any)
	}
	t.Metadata[key] = value
}

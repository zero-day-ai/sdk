package finding

import "fmt"

// Category represents the type of security finding.
type Category string

const (
	// CategoryJailbreak indicates attempts to bypass LLM safety controls.
	// Examples: Prompt manipulation to bypass content filters, role-playing attacks
	CategoryJailbreak Category = "jailbreak"

	// CategoryPromptInjection indicates malicious prompt injection attacks.
	// Examples: System prompt manipulation, indirect prompt injection
	CategoryPromptInjection Category = "prompt_injection"

	// CategoryDataExtraction indicates unauthorized data access or exfiltration.
	// Examples: Training data extraction, PII leakage, model inversion
	CategoryDataExtraction Category = "data_extraction"

	// CategoryPrivilegeEscalation indicates unauthorized privilege elevation.
	// Examples: Role hijacking, permission bypass, access control violations
	CategoryPrivilegeEscalation Category = "privilege_escalation"

	// CategoryDOS indicates denial of service or resource exhaustion attacks.
	// Examples: Token flooding, infinite loops, resource exhaustion
	CategoryDOS Category = "dos"

	// CategoryModelManipulation indicates attacks that modify model behavior.
	// Examples: Poisoning attacks, backdoor injection, model reprogramming
	CategoryModelManipulation Category = "model_manipulation"

	// CategoryInformationDisclosure indicates unintended information exposure.
	// Examples: System information leaks, configuration disclosure, metadata exposure
	CategoryInformationDisclosure Category = "information_disclosure"
)

// IsValid returns true if the category is valid.
func (c Category) IsValid() bool {
	switch c {
	case CategoryJailbreak,
		CategoryPromptInjection,
		CategoryDataExtraction,
		CategoryPrivilegeEscalation,
		CategoryDOS,
		CategoryModelManipulation,
		CategoryInformationDisclosure:
		return true
	default:
		return false
	}
}

// String returns the string representation of the category.
func (c Category) String() string {
	return string(c)
}

// DisplayName returns a human-readable display name for the category.
func (c Category) DisplayName() string {
	switch c {
	case CategoryJailbreak:
		return "Jailbreak"
	case CategoryPromptInjection:
		return "Prompt Injection"
	case CategoryDataExtraction:
		return "Data Extraction"
	case CategoryPrivilegeEscalation:
		return "Privilege Escalation"
	case CategoryDOS:
		return "Denial of Service"
	case CategoryModelManipulation:
		return "Model Manipulation"
	case CategoryInformationDisclosure:
		return "Information Disclosure"
	default:
		return string(c)
	}
}

// Description returns a brief description of the category.
func (c Category) Description() string {
	switch c {
	case CategoryJailbreak:
		return "Attempts to bypass LLM safety controls and content filters"
	case CategoryPromptInjection:
		return "Malicious prompt injection to manipulate model behavior"
	case CategoryDataExtraction:
		return "Unauthorized access or exfiltration of sensitive data"
	case CategoryPrivilegeEscalation:
		return "Unauthorized elevation of privileges or permissions"
	case CategoryDOS:
		return "Denial of service or resource exhaustion attacks"
	case CategoryModelManipulation:
		return "Attacks that modify or reprogram model behavior"
	case CategoryInformationDisclosure:
		return "Unintended exposure of system or sensitive information"
	default:
		return ""
	}
}

// ParseCategory parses a string into a Category value.
// Returns an error if the string is not a valid category.
func ParseCategory(s string) (Category, error) {
	category := Category(s)
	if !category.IsValid() {
		return "", fmt.Errorf("invalid category: %s", s)
	}
	return category, nil
}

// AllCategories returns all valid categories.
func AllCategories() []Category {
	return []Category{
		CategoryJailbreak,
		CategoryPromptInjection,
		CategoryDataExtraction,
		CategoryPrivilegeEscalation,
		CategoryDOS,
		CategoryModelManipulation,
		CategoryInformationDisclosure,
	}
}

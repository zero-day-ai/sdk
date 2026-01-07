package agent

import (
	"context"

	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/types"
)

// Agent is the interface that all SDK agents must implement.
// Agents are autonomous components that execute security testing tasks
// using LLMs, tools, and plugins provided by the harness.
type Agent interface {
	// Name returns the unique identifier for this agent.
	// This should be a short, kebab-case name (e.g., "prompt-injector").
	Name() string

	// Version returns the semantic version of this agent.
	// Format: "major.minor.patch" (e.g., "1.0.0").
	Version() string

	// Description returns a human-readable description of what this agent does.
	// This should explain the agent's purpose and capabilities.
	Description() string

	// Capabilities returns the security testing capabilities this agent provides.
	// These indicate what types of vulnerabilities the agent can discover.
	Capabilities() []Capability

	// TargetSchemas returns the target schemas this agent supports.
	// This defines the connection parameter requirements for each target type.
	// Returns nil or empty slice to accept any target type (opt-out validation).
	TargetSchemas() []types.TargetSchema

	// Deprecated: Use TargetSchemas() instead.
	// TargetTypes returns the types of AI systems this agent can test.
	// This helps the framework match agents to appropriate targets.
	TargetTypes() []types.TargetType

	// TechniqueTypes returns the attack techniques this agent employs.
	// This categorizes the agent's testing methodology.
	TechniqueTypes() []types.TechniqueType

	// LLMSlots returns the LLM slot definitions required by this agent.
	// The framework will provision LLMs that meet these requirements.
	LLMSlots() []llm.SlotDefinition

	// Execute performs a task using the provided harness.
	// The harness provides access to LLMs, tools, plugins, and other resources.
	// The context can be used for cancellation and timeout control.
	Execute(ctx context.Context, harness Harness, task Task) (Result, error)

	// Initialize prepares the agent for execution.
	// This is called once when the agent is loaded, before any tasks are executed.
	// The config map contains agent-specific configuration parameters.
	Initialize(ctx context.Context, config map[string]any) error

	// Shutdown gracefully stops the agent and releases resources.
	// This is called when the agent is being unloaded or the system is shutting down.
	Shutdown(ctx context.Context) error

	// Health returns the current health status of the agent.
	// This is used for monitoring and diagnostics.
	Health(ctx context.Context) types.HealthStatus
}

// Capability represents a security testing capability that an agent provides.
// Capabilities describe what types of vulnerabilities the agent can discover.
type Capability string

const (
	// CapabilityPromptInjection indicates the agent can test for prompt injection vulnerabilities.
	// This includes direct injection, indirect injection, and cross-context attacks.
	CapabilityPromptInjection Capability = "prompt_injection"

	// CapabilityJailbreak indicates the agent can test for jailbreak vulnerabilities.
	// This includes attempts to bypass safety guardrails and content filters.
	CapabilityJailbreak Capability = "jailbreak"

	// CapabilityDataExtraction indicates the agent can test for data extraction vulnerabilities.
	// This includes extracting training data, PII, or sensitive information.
	CapabilityDataExtraction Capability = "data_extraction"

	// CapabilityModelManipulation indicates the agent can test for model manipulation vulnerabilities.
	// This includes bias exploitation, output formatting attacks, and behavior modification.
	CapabilityModelManipulation Capability = "model_manipulation"

	// CapabilityDOS indicates the agent can test for denial-of-service vulnerabilities.
	// This includes resource exhaustion, infinite loops, and performance degradation.
	CapabilityDOS Capability = "dos"
)

// String returns the string representation of the capability.
func (c Capability) String() string {
	return string(c)
}

// IsValid checks if the capability is a recognized value.
func (c Capability) IsValid() bool {
	switch c {
	case CapabilityPromptInjection, CapabilityJailbreak, CapabilityDataExtraction,
		CapabilityModelManipulation, CapabilityDOS:
		return true
	default:
		return false
	}
}

// Description returns a human-readable description of the capability.
func (c Capability) Description() string {
	switch c {
	case CapabilityPromptInjection:
		return "Tests for prompt injection vulnerabilities and instruction manipulation"
	case CapabilityJailbreak:
		return "Tests for jailbreak attempts to bypass safety guardrails"
	case CapabilityDataExtraction:
		return "Tests for data extraction and information disclosure vulnerabilities"
	case CapabilityModelManipulation:
		return "Tests for model manipulation and behavior modification"
	case CapabilityDOS:
		return "Tests for denial-of-service and resource exhaustion vulnerabilities"
	default:
		return "Unknown capability"
	}
}

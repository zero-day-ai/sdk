package agent

import (
	"context"
	"testing"

	"github.com/zero-day-ai/sdk/llm"
	"github.com/zero-day-ai/sdk/types"
)

// mockAgent is a simple implementation of the Agent interface for testing.
type mockAgent struct {
	name           string
	version        string
	description    string
	capabilities   []Capability
	targetTypes    []types.TargetType
	techniqueTypes []types.TechniqueType
	llmSlots       []llm.SlotDefinition
	initCalled     bool
	shutdownCalled bool
}

func (m *mockAgent) Name() string {
	return m.name
}

func (m *mockAgent) Version() string {
	return m.version
}

func (m *mockAgent) Description() string {
	return m.description
}

func (m *mockAgent) Capabilities() []Capability {
	return m.capabilities
}

func (m *mockAgent) TargetTypes() []types.TargetType {
	return m.targetTypes
}

func (m *mockAgent) TechniqueTypes() []types.TechniqueType {
	return m.techniqueTypes
}

func (m *mockAgent) LLMSlots() []llm.SlotDefinition {
	return m.llmSlots
}

func (m *mockAgent) Execute(ctx context.Context, harness Harness, task Task) (Result, error) {
	return NewSuccessResult("mock execution result"), nil
}

func (m *mockAgent) Initialize(ctx context.Context, config map[string]any) error {
	m.initCalled = true
	return nil
}

func (m *mockAgent) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func (m *mockAgent) Health(ctx context.Context) types.HealthStatus {
	return types.NewHealthyStatus("mock agent is healthy")
}

func TestMockAgent(t *testing.T) {
	agent := &mockAgent{
		name:         "test-agent",
		version:      "1.0.0",
		description:  "A test agent",
		capabilities: []Capability{CapabilityPromptInjection},
		targetTypes:  []types.TargetType{types.TargetTypeLLMChat},
		techniqueTypes: []types.TechniqueType{types.TechniquePromptInjection},
		llmSlots: []llm.SlotDefinition{
			{
				Name:             "primary",
				Description:      "Primary LLM slot",
				Required:         true,
				MinContextWindow: 8000,
			},
		},
	}

	ctx := context.Background()

	// Test metadata methods
	if agent.Name() != "test-agent" {
		t.Errorf("Name() = %s, want test-agent", agent.Name())
	}
	if agent.Version() != "1.0.0" {
		t.Errorf("Version() = %s, want 1.0.0", agent.Version())
	}
	if agent.Description() != "A test agent" {
		t.Errorf("Description() = %s, want 'A test agent'", agent.Description())
	}

	// Test capabilities
	caps := agent.Capabilities()
	if len(caps) != 1 || caps[0] != CapabilityPromptInjection {
		t.Errorf("Capabilities() = %v, want [%s]", caps, CapabilityPromptInjection)
	}

	// Test target types
	targets := agent.TargetTypes()
	if len(targets) != 1 || targets[0] != types.TargetTypeLLMChat {
		t.Errorf("TargetTypes() = %v, want [%s]", targets, types.TargetTypeLLMChat)
	}

	// Test technique types
	techniques := agent.TechniqueTypes()
	if len(techniques) != 1 || techniques[0] != types.TechniquePromptInjection {
		t.Errorf("TechniqueTypes() = %v, want [%s]", techniques, types.TechniquePromptInjection)
	}

	// Test LLM slots
	slots := agent.LLMSlots()
	if len(slots) != 1 {
		t.Fatalf("LLMSlots() returned %d slots, want 1", len(slots))
	}
	if slots[0].Name != "primary" {
		t.Errorf("LLMSlots()[0].Name = %s, want primary", slots[0].Name)
	}

	// Test lifecycle methods
	err := agent.Initialize(ctx, map[string]any{"key": "value"})
	if err != nil {
		t.Errorf("Initialize() error = %v, want nil", err)
	}
	if !agent.initCalled {
		t.Error("Initialize() did not set initCalled flag")
	}

	err = agent.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}
	if !agent.shutdownCalled {
		t.Error("Shutdown() did not set shutdownCalled flag")
	}

	// Test health check
	health := agent.Health(ctx)
	if !health.IsHealthy() {
		t.Errorf("Health() = %v, want healthy status", health)
	}

	// Test execution
	task := NewTask("test-task", "test goal")
	result, err := agent.Execute(ctx, nil, *task)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	if result.Status != StatusSuccess {
		t.Errorf("Execute() result.Status = %v, want %v", result.Status, StatusSuccess)
	}
}

func TestCapability(t *testing.T) {
	tests := []struct {
		name        string
		capability  Capability
		wantValid   bool
		wantString  string
		wantDescLen int
	}{
		{
			name:        "prompt injection",
			capability:  CapabilityPromptInjection,
			wantValid:   true,
			wantString:  "prompt_injection",
			wantDescLen: 50,
		},
		{
			name:        "jailbreak",
			capability:  CapabilityJailbreak,
			wantValid:   true,
			wantString:  "jailbreak",
			wantDescLen: 50,
		},
		{
			name:        "data extraction",
			capability:  CapabilityDataExtraction,
			wantValid:   true,
			wantString:  "data_extraction",
			wantDescLen: 50,
		},
		{
			name:        "model manipulation",
			capability:  CapabilityModelManipulation,
			wantValid:   true,
			wantString:  "model_manipulation",
			wantDescLen: 40,
		},
		{
			name:        "dos",
			capability:  CapabilityDOS,
			wantValid:   true,
			wantString:  "dos",
			wantDescLen: 50,
		},
		{
			name:        "invalid",
			capability:  Capability("invalid"),
			wantValid:   false,
			wantString:  "invalid",
			wantDescLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.capability.IsValid(); got != tt.wantValid {
				t.Errorf("IsValid() = %v, want %v", got, tt.wantValid)
			}
			if got := tt.capability.String(); got != tt.wantString {
				t.Errorf("String() = %v, want %v", got, tt.wantString)
			}
			desc := tt.capability.Description()
			if len(desc) < tt.wantDescLen {
				t.Errorf("Description() length = %d, want at least %d", len(desc), tt.wantDescLen)
			}
		})
	}
}

func TestCapability_AllValid(t *testing.T) {
	// Ensure all defined capabilities are valid
	capabilities := []Capability{
		CapabilityPromptInjection,
		CapabilityJailbreak,
		CapabilityDataExtraction,
		CapabilityModelManipulation,
		CapabilityDOS,
	}

	for _, cap := range capabilities {
		t.Run(cap.String(), func(t *testing.T) {
			if !cap.IsValid() {
				t.Errorf("capability %s should be valid", cap)
			}
			if cap.String() == "" {
				t.Error("String() should not be empty")
			}
			if cap.Description() == "" {
				t.Error("Description() should not be empty")
			}
		})
	}
}

func TestCapability_Description_Coverage(t *testing.T) {
	// Verify that all valid capabilities have meaningful descriptions
	capabilities := []Capability{
		CapabilityPromptInjection,
		CapabilityJailbreak,
		CapabilityDataExtraction,
		CapabilityModelManipulation,
		CapabilityDOS,
	}

	for _, cap := range capabilities {
		desc := cap.Description()
		if desc == "" || desc == "Unknown capability" {
			t.Errorf("capability %s has no description", cap)
		}
	}

	// Test unknown capability
	unknown := Capability("unknown")
	if unknown.Description() != "Unknown capability" {
		t.Errorf("unknown capability should return 'Unknown capability'")
	}
}

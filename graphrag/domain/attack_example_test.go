package domain_test

import (
	"fmt"

	"github.com/zero-day-ai/sdk/graphrag"
	"github.com/zero-day-ai/sdk/graphrag/domain"
)

// ExampleTacticRegistry demonstrates how to use the TacticRegistry.
func ExampleTacticRegistry() {
	registry := domain.NewTacticRegistry()

	// Get all tactics
	tactics := registry.AllTactics()
	fmt.Printf("Total tactics: %d\n", len(tactics))

	// Get a specific tactic
	recon := registry.GetTactic(graphrag.TacticReconnaissance)
	if recon != nil {
		fmt.Printf("Tactic: %s - %s\n", recon.ID, recon.Name)
	}

	// Use tactic as a GraphNode
	var node domain.GraphNode = recon
	fmt.Printf("Node type: %s\n", node.NodeType())

	// Output:
	// Total tactics: 14
	// Tactic: GIB-TA01 - Reconnaissance
	// Node type: tactic
}

// ExampleTechniqueRegistry demonstrates how to use the TechniqueRegistry.
func ExampleTechniqueRegistry() {
	registry := domain.NewTechniqueRegistry()

	// Get all techniques
	techniques := registry.AllTechniques()
	fmt.Printf("Total techniques: %d\n", len(techniques))

	// Get a specific technique
	promptInjection := registry.GetTechnique(graphrag.TechniquePromptInjection)
	if promptInjection != nil {
		fmt.Printf("Technique: %s - %s\n", promptInjection.ID, promptInjection.Name)
		fmt.Printf("Severity: %s\n", promptInjection.Severity)
		fmt.Printf("Tactics: %d\n", len(promptInjection.TacticIDs))
	}

	// Output:
	// Total techniques: 20
	// Technique: GIB-T1001 - Prompt Injection
	// Severity: high
	// Tactics: 2
}

// ExampleTechniqueRegistry_TechniquesByTactic demonstrates finding techniques by tactic.
func ExampleTechniqueRegistry_TechniquesByTactic() {
	registry := domain.NewTechniqueRegistry()

	// Find all techniques that help achieve Initial Access
	techniques := registry.TechniquesByTactic(graphrag.TacticInitialAccess)
	fmt.Printf("Initial Access techniques: %d\n", len(techniques))

	for _, tech := range techniques {
		fmt.Printf("- %s: %s\n", tech.ID, tech.Name)
	}

	// Output will vary based on implementation, but should include:
	// Initial Access techniques: 3
	// - GIB-T1001: Prompt Injection
	// - GIB-T1018: Payload Splitting
	// - GIB-T1019: Indirect Prompt Injection
}

// ExampleTactic_asGraphNode demonstrates using a Tactic as a GraphNode.
func ExampleTactic_asGraphNode() {
	// Create a tactic instance
	tactic := &domain.Tactic{
		ID:          graphrag.TacticExecution,
		Name:        "Execution",
		Description: "Running malicious prompts or code",
		Phase:       4,
	}

	// Use as GraphNode interface
	var node domain.GraphNode = tactic

	// GraphNode interface methods
	fmt.Printf("Type: %s\n", node.NodeType())
	fmt.Printf("ID: %v\n", node.IdentifyingProperties()["id"])
	fmt.Printf("Has parent: %v\n", node.ParentRef() != nil)

	// Access properties
	props := node.Properties()
	fmt.Printf("Name: %s\n", props["name"])
	fmt.Printf("Phase: %d\n", props["phase"])

	// Output:
	// Type: tactic
	// ID: GIB-TA04
	// Has parent: false
	// Name: Execution
	// Phase: 4
}

// ExampleTechnique_asGraphNode demonstrates using a Technique as a GraphNode.
func ExampleTechnique_asGraphNode() {
	// Create a technique instance
	technique := &domain.Technique{
		ID:          graphrag.TechniqueJailbreak,
		Name:        "Jailbreak",
		Description: "Bypassing model safety restrictions",
		TacticIDs:   []string{graphrag.TacticDefenseEvasion},
		Severity:    "high",
		Detection:   "Track roleplay scenarios and hypothetical framing",
		Mitigation:  "Multi-layer filtering and semantic analysis",
	}

	// Use as GraphNode interface
	var node domain.GraphNode = technique

	// GraphNode interface methods
	fmt.Printf("Type: %s\n", node.NodeType())
	fmt.Printf("ID: %v\n", node.IdentifyingProperties()["id"])
	fmt.Printf("Has parent: %v\n", node.ParentRef() != nil)

	// Access properties
	props := node.Properties()
	fmt.Printf("Name: %s\n", props["name"])
	fmt.Printf("Severity: %s\n", props["severity"])

	// Output:
	// Type: technique
	// ID: GIB-T1002
	// Has parent: false
	// Name: Jailbreak
	// Severity: high
}

// Example_attackTaxonomyIntegration demonstrates working with both tactics and techniques.
func Example_attackTaxonomyIntegration() {
	tacticReg := domain.NewTacticRegistry()
	techniqueReg := domain.NewTechniqueRegistry()

	// Get a tactic
	defenseEvasion := tacticReg.GetTactic(graphrag.TacticDefenseEvasion)
	fmt.Printf("Tactic: %s (Phase %d)\n", defenseEvasion.Name, defenseEvasion.Phase)

	// Find techniques for this tactic
	techniques := techniqueReg.TechniquesByTactic(graphrag.TacticDefenseEvasion)
	fmt.Printf("Techniques: %d\n", len(techniques))

	// List some techniques
	for i, tech := range techniques {
		if i >= 3 {
			break // Just show first 3
		}
		fmt.Printf("- %s: %s (severity: %s)\n", tech.ID, tech.Name, tech.Severity)
	}

	// Output will include techniques like:
	// Tactic: Defense Evasion (Phase 7)
	// Techniques: 9
	// - GIB-T1002: Jailbreak (severity: high)
	// - GIB-T1006: RAG Poisoning (severity: high)
	// - GIB-T1007: Citation Injection (severity: medium)
}

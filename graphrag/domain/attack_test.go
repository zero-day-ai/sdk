package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zero-day-ai/sdk/graphrag"
)

// TestTactic_GraphNodeInterface tests that Tactic implements GraphNode correctly.
func TestTactic_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name     string
		tactic   *Tactic
		wantType string
		wantID   map[string]any
	}{
		{
			name: "reconnaissance tactic",
			tactic: &Tactic{
				ID:          graphrag.TacticReconnaissance,
				Name:        "Reconnaissance",
				Description: "Information gathering",
				Phase:       1,
			},
			wantType: graphrag.NodeTypeTactic,
			wantID: map[string]any{
				"id": graphrag.TacticReconnaissance,
			},
		},
		{
			name: "execution tactic",
			tactic: &Tactic{
				ID:          graphrag.TacticExecution,
				Name:        "Execution",
				Description: "Running malicious code",
				Phase:       4,
			},
			wantType: graphrag.NodeTypeTactic,
			wantID: map[string]any{
				"id": graphrag.TacticExecution,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test NodeType
			assert.Equal(t, tt.wantType, tt.tactic.NodeType())

			// Test IdentifyingProperties
			assert.Equal(t, tt.wantID, tt.tactic.IdentifyingProperties())

			// Test Properties (should include all fields)
			props := tt.tactic.Properties()
			assert.Equal(t, tt.tactic.ID, props["id"])
			assert.Equal(t, tt.tactic.Name, props["name"])
			assert.Equal(t, tt.tactic.Description, props["description"])
			assert.Equal(t, tt.tactic.Phase, props["phase"])

			// Test ParentRef (tactics are root nodes)
			assert.Nil(t, tt.tactic.ParentRef())

			// Test RelationshipType (tactics have no parent)
			assert.Equal(t, "", tt.tactic.RelationshipType())
		})
	}
}

// TestTechnique_GraphNodeInterface tests that Technique implements GraphNode correctly.
func TestTechnique_GraphNodeInterface(t *testing.T) {
	tests := []struct {
		name       string
		technique  *Technique
		wantType   string
		wantID     map[string]any
		checkProps func(*testing.T, map[string]any)
	}{
		{
			name: "prompt injection technique",
			technique: &Technique{
				ID:          graphrag.TechniquePromptInjection,
				Name:        "Prompt Injection",
				Description: "Direct injection of malicious prompts",
				TacticIDs:   []string{graphrag.TacticInitialAccess, graphrag.TacticExecution},
				Detection:   "Monitor for unusual prompt patterns",
				Mitigation:  "Input validation, prompt sandboxing",
				Severity:    "high",
			},
			wantType: graphrag.NodeTypeTechnique,
			wantID: map[string]any{
				"id": graphrag.TechniquePromptInjection,
			},
			checkProps: func(t *testing.T, props map[string]any) {
				assert.Equal(t, graphrag.TechniquePromptInjection, props["id"])
				assert.Equal(t, "Prompt Injection", props["name"])
				assert.Equal(t, "high", props["severity"])
				assert.Contains(t, props, "detection")
				assert.Contains(t, props, "mitigation")
			},
		},
		{
			name: "jailbreak technique without optional fields",
			technique: &Technique{
				ID:          graphrag.TechniqueJailbreak,
				Name:        "Jailbreak",
				Description: "Bypassing model safety restrictions",
				TacticIDs:   []string{graphrag.TacticDefenseEvasion},
				Severity:    "high",
			},
			wantType: graphrag.NodeTypeTechnique,
			wantID: map[string]any{
				"id": graphrag.TechniqueJailbreak,
			},
			checkProps: func(t *testing.T, props map[string]any) {
				assert.Equal(t, graphrag.TechniqueJailbreak, props["id"])
				assert.Equal(t, "Jailbreak", props["name"])
				// Optional fields should not be present if empty
				assert.NotContains(t, props, "detection")
				assert.NotContains(t, props, "mitigation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test NodeType
			assert.Equal(t, tt.wantType, tt.technique.NodeType())

			// Test IdentifyingProperties
			assert.Equal(t, tt.wantID, tt.technique.IdentifyingProperties())

			// Test Properties with custom check function
			props := tt.technique.Properties()
			tt.checkProps(t, props)

			// Test ParentRef (techniques don't use parent relationships)
			assert.Nil(t, tt.technique.ParentRef())

			// Test RelationshipType (techniques have no parent)
			assert.Equal(t, "", tt.technique.RelationshipType())
		})
	}
}

// TestTacticRegistry tests the TacticRegistry functionality.
func TestTacticRegistry(t *testing.T) {
	registry := NewTacticRegistry()
	require.NotNil(t, registry)

	t.Run("AllTactics returns all 14 tactics", func(t *testing.T) {
		tactics := registry.AllTactics()
		assert.Len(t, tactics, 14)

		// Verify tactics are in phase order
		for i, tactic := range tactics {
			assert.Equal(t, i+1, tactic.Phase, "Tactic %s should be in phase %d", tactic.ID, i+1)
		}
	})

	t.Run("GetTactic finds existing tactics", func(t *testing.T) {
		tactic := registry.GetTactic(graphrag.TacticReconnaissance)
		require.NotNil(t, tactic)
		assert.Equal(t, graphrag.TacticReconnaissance, tactic.ID)
		assert.Equal(t, "Reconnaissance", tactic.Name)
		assert.Equal(t, 1, tactic.Phase)

		tactic = registry.GetTactic(graphrag.TacticExecution)
		require.NotNil(t, tactic)
		assert.Equal(t, graphrag.TacticExecution, tactic.ID)
		assert.Equal(t, "Execution", tactic.Name)
		assert.Equal(t, 4, tactic.Phase)

		tactic = registry.GetTactic(graphrag.TacticImpact)
		require.NotNil(t, tactic)
		assert.Equal(t, graphrag.TacticImpact, tactic.ID)
		assert.Equal(t, "Impact", tactic.Name)
		assert.Equal(t, 13, tactic.Phase)
	})

	t.Run("GetTactic returns nil for non-existent tactic", func(t *testing.T) {
		tactic := registry.GetTactic("GIB-TA99")
		assert.Nil(t, tactic)
	})

	t.Run("All tactics have valid structure", func(t *testing.T) {
		for _, tactic := range registry.AllTactics() {
			assert.NotEmpty(t, tactic.ID, "Tactic should have an ID")
			assert.NotEmpty(t, tactic.Name, "Tactic should have a name")
			assert.NotEmpty(t, tactic.Description, "Tactic should have a description")
			assert.Greater(t, tactic.Phase, 0, "Tactic phase should be positive")
			assert.LessOrEqual(t, tactic.Phase, 14, "Tactic phase should be <= 14")
		}
	})
}

// TestTechniqueRegistry tests the TechniqueRegistry functionality.
func TestTechniqueRegistry(t *testing.T) {
	registry := NewTechniqueRegistry()
	require.NotNil(t, registry)

	t.Run("AllTechniques returns all 20 techniques", func(t *testing.T) {
		techniques := registry.AllTechniques()
		assert.Len(t, techniques, 20)
	})

	t.Run("GetTechnique finds existing techniques", func(t *testing.T) {
		technique := registry.GetTechnique(graphrag.TechniquePromptInjection)
		require.NotNil(t, technique)
		assert.Equal(t, graphrag.TechniquePromptInjection, technique.ID)
		assert.Equal(t, "Prompt Injection", technique.Name)
		assert.Equal(t, "high", technique.Severity)

		technique = registry.GetTechnique(graphrag.TechniqueJailbreak)
		require.NotNil(t, technique)
		assert.Equal(t, graphrag.TechniqueJailbreak, technique.ID)
		assert.Equal(t, "Jailbreak", technique.Name)

		technique = registry.GetTechnique(graphrag.TechniqueMultiModalInjection)
		require.NotNil(t, technique)
		assert.Equal(t, graphrag.TechniqueMultiModalInjection, technique.ID)
		assert.Equal(t, "Multi-Modal Injection", technique.Name)
	})

	t.Run("GetTechnique returns nil for non-existent technique", func(t *testing.T) {
		technique := registry.GetTechnique("GIB-T9999")
		assert.Nil(t, technique)
	})

	t.Run("All techniques have valid structure", func(t *testing.T) {
		for _, technique := range registry.AllTechniques() {
			assert.NotEmpty(t, technique.ID, "Technique should have an ID")
			assert.NotEmpty(t, technique.Name, "Technique should have a name")
			assert.NotEmpty(t, technique.Description, "Technique should have a description")
			assert.NotEmpty(t, technique.Severity, "Technique should have a severity")
			assert.NotEmpty(t, technique.TacticIDs, "Technique should map to at least one tactic")

			// Verify severity is valid
			validSeverities := map[string]bool{
				"critical": true,
				"high":     true,
				"medium":   true,
				"low":      true,
				"info":     true,
			}
			assert.True(t, validSeverities[technique.Severity],
				"Technique %s has invalid severity: %s", technique.ID, technique.Severity)
		}
	})

	t.Run("TechniquesByTactic returns correct techniques", func(t *testing.T) {
		// Test Initial Access tactic
		techniques := registry.TechniquesByTactic(graphrag.TacticInitialAccess)
		assert.NotEmpty(t, techniques)

		// Verify Prompt Injection is included (uses Initial Access)
		found := false
		for _, tech := range techniques {
			if tech.ID == graphrag.TechniquePromptInjection {
				found = true
				break
			}
		}
		assert.True(t, found, "Prompt Injection should be in Initial Access techniques")

		// Test Defense Evasion tactic
		techniques = registry.TechniquesByTactic(graphrag.TacticDefenseEvasion)
		assert.NotEmpty(t, techniques)

		// Verify Jailbreak is included (uses Defense Evasion)
		found = false
		for _, tech := range techniques {
			if tech.ID == graphrag.TechniqueJailbreak {
				found = true
				break
			}
		}
		assert.True(t, found, "Jailbreak should be in Defense Evasion techniques")
	})

	t.Run("TechniquesByTactic returns empty for non-existent tactic", func(t *testing.T) {
		techniques := registry.TechniquesByTactic("GIB-TA99")
		assert.Empty(t, techniques)
	})
}

// TestAttackTaxonomyIntegration tests the integration between tactics and techniques.
func TestAttackTaxonomyIntegration(t *testing.T) {
	tacticRegistry := NewTacticRegistry()
	techniqueRegistry := NewTechniqueRegistry()

	t.Run("All technique tactic IDs reference valid tactics", func(t *testing.T) {
		validTacticIDs := make(map[string]bool)
		for _, tactic := range tacticRegistry.AllTactics() {
			validTacticIDs[tactic.ID] = true
		}

		for _, technique := range techniqueRegistry.AllTechniques() {
			for _, tacticID := range technique.TacticIDs {
				assert.True(t, validTacticIDs[tacticID],
					"Technique %s references invalid tactic ID: %s", technique.ID, tacticID)
			}
		}
	})

	t.Run("Tactics with techniques have valid mappings", func(t *testing.T) {
		// Count how many tactics have at least one technique
		tacticsWithTechniques := 0
		for _, tactic := range tacticRegistry.AllTactics() {
			techniques := techniqueRegistry.TechniquesByTactic(tactic.ID)
			if len(techniques) > 0 {
				tacticsWithTechniques++
			}
		}
		// At least some tactics should have techniques mapped
		assert.Greater(t, tacticsWithTechniques, 0,
			"At least one tactic should have techniques mapped")
	})

	t.Run("Verify specific tactic-technique mappings", func(t *testing.T) {
		// Prompt Injection should map to Initial Access and Execution
		promptInjection := techniqueRegistry.GetTechnique(graphrag.TechniquePromptInjection)
		require.NotNil(t, promptInjection)
		assert.Contains(t, promptInjection.TacticIDs, graphrag.TacticInitialAccess)
		assert.Contains(t, promptInjection.TacticIDs, graphrag.TacticExecution)

		// Jailbreak should map to Defense Evasion
		jailbreak := techniqueRegistry.GetTechnique(graphrag.TechniqueJailbreak)
		require.NotNil(t, jailbreak)
		assert.Contains(t, jailbreak.TacticIDs, graphrag.TacticDefenseEvasion)

		// System Prompt Extraction should map to Discovery and Collection
		sysPromptExtract := techniqueRegistry.GetTechnique(graphrag.TechniqueSystemPromptExtraction)
		require.NotNil(t, sysPromptExtract)
		assert.Contains(t, sysPromptExtract.TacticIDs, graphrag.TacticDiscovery)
		assert.Contains(t, sysPromptExtract.TacticIDs, graphrag.TacticCollection)
	})
}

// TestAttackTypes_AsGraphNodes tests that attack types can be used as GraphNode interface.
func TestAttackTypes_AsGraphNodes(t *testing.T) {
	t.Run("Tactic as GraphNode", func(t *testing.T) {
		tactic := &Tactic{
			ID:          graphrag.TacticReconnaissance,
			Name:        "Reconnaissance",
			Description: "Info gathering",
			Phase:       1,
		}

		// Should be usable as GraphNode interface
		var node GraphNode = tactic
		assert.Equal(t, graphrag.NodeTypeTactic, node.NodeType())
		assert.NotNil(t, node.IdentifyingProperties())
		assert.NotNil(t, node.Properties())
		assert.Nil(t, node.ParentRef())
		assert.Equal(t, "", node.RelationshipType())
	})

	t.Run("Technique as GraphNode", func(t *testing.T) {
		technique := &Technique{
			ID:          graphrag.TechniquePromptInjection,
			Name:        "Prompt Injection",
			Description: "Malicious prompt injection",
			TacticIDs:   []string{graphrag.TacticInitialAccess},
			Severity:    "high",
		}

		// Should be usable as GraphNode interface
		var node GraphNode = technique
		assert.Equal(t, graphrag.NodeTypeTechnique, node.NodeType())
		assert.NotNil(t, node.IdentifyingProperties())
		assert.NotNil(t, node.Properties())
		assert.Nil(t, node.ParentRef())
		assert.Equal(t, "", node.RelationshipType())
	})
}

package domain

import (
	"github.com/zero-day-ai/sdk/graphrag"
)

// Attack Taxonomy Types (Gibson-native)
// These types represent the Gibson attack taxonomy (GIB-TA and GIB-T series)

// Tactic represents a Gibson-native attack tactic (GIB-TA series).
// Tactics represent high-level adversary goals in the attack lifecycle.
// Examples: Reconnaissance, Initial Access, Execution, etc.
//
// Tactics are root nodes (no parent) in the knowledge graph.
type Tactic struct {
	// ID is the Gibson tactic identifier (e.g., "GIB-TA01")
	ID string

	// Name is the human-readable tactic name (e.g., "Reconnaissance")
	Name string

	// Description explains the tactic's purpose and goals
	Description string

	// Phase is the ordering phase in the attack lifecycle (1-14)
	Phase int
}

// NodeType returns the canonical node type for tactics.
func (t *Tactic) NodeType() string {
	return graphrag.NodeTypeTactic
}

// IdentifyingProperties returns the properties that uniquely identify this tactic.
func (t *Tactic) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": t.ID,
	}
}

// Properties returns all properties to set on the tactic node.
func (t *Tactic) Properties() map[string]any {
	return map[string]any{
		"id":          t.ID,
		"name":        t.Name,
		"description": t.Description,
		"phase":       t.Phase,
	}
}

// ParentRef returns nil since tactics are root nodes.
func (t *Tactic) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string since tactics have no parent.
func (t *Tactic) RelationshipType() string {
	return ""
}

// Technique represents a Gibson-native attack technique (GIB-T series).
// Techniques are specific methods to achieve tactical goals.
// Examples: Prompt Injection, Jailbreak, System Prompt Extraction, etc.
//
// Techniques can reference multiple tactics they help achieve.
type Technique struct {
	// ID is the Gibson technique identifier (e.g., "GIB-T1001")
	ID string

	// Name is the human-readable technique name (e.g., "Prompt Injection")
	Name string

	// Description explains the technique and how it works
	Description string

	// TacticIDs lists the tactics this technique helps achieve
	TacticIDs []string

	// Detection describes how to detect this technique in use
	Detection string

	// Mitigation describes how to mitigate or prevent this technique
	Mitigation string

	// Severity indicates the typical severity of findings using this technique
	// Values: "critical", "high", "medium", "low", "info"
	Severity string
}

// NodeType returns the canonical node type for techniques.
func (t *Technique) NodeType() string {
	return graphrag.NodeTypeTechnique
}

// IdentifyingProperties returns the properties that uniquely identify this technique.
func (t *Technique) IdentifyingProperties() map[string]any {
	return map[string]any{
		"id": t.ID,
	}
}

// Properties returns all properties to set on the technique node.
func (t *Technique) Properties() map[string]any {
	props := map[string]any{
		"id":          t.ID,
		"name":        t.Name,
		"description": t.Description,
		"tactic_ids":  t.TacticIDs,
		"severity":    t.Severity,
	}

	// Only add optional fields if they're set
	if t.Detection != "" {
		props["detection"] = t.Detection
	}
	if t.Mitigation != "" {
		props["mitigation"] = t.Mitigation
	}

	return props
}

// ParentRef returns nil since techniques don't use parent relationships.
// Techniques link to tactics via USES_TECHNIQUE relationships, not parent-child.
func (t *Technique) ParentRef() *NodeRef {
	return nil
}

// RelationshipType returns empty string since techniques have no parent.
func (t *Technique) RelationshipType() string {
	return ""
}

// TacticRegistry provides access to predefined Gibson tactics.
type TacticRegistry struct{}

// NewTacticRegistry creates a new TacticRegistry instance.
func NewTacticRegistry() *TacticRegistry {
	return &TacticRegistry{}
}

// GetTactic returns a specific tactic by ID, or nil if not found.
func (r *TacticRegistry) GetTactic(id string) *Tactic {
	tactics := r.AllTactics()
	for _, tactic := range tactics {
		if tactic.ID == id {
			return tactic
		}
	}
	return nil
}

// AllTactics returns all predefined Gibson tactics in phase order.
func (r *TacticRegistry) AllTactics() []*Tactic {
	return []*Tactic{
		{
			ID:          graphrag.TacticReconnaissance,
			Name:        "Reconnaissance",
			Description: "Information gathering about target systems, AI models, and infrastructure to plan attacks",
			Phase:       1,
		},
		{
			ID:          graphrag.TacticResourceDevelopment,
			Name:        "Resource Development",
			Description: "Building attack infrastructure, crafting payloads, and preparing attack resources",
			Phase:       2,
		},
		{
			ID:          graphrag.TacticInitialAccess,
			Name:        "Initial Access",
			Description: "Gaining initial foothold in target system through prompts, API calls, or injection",
			Phase:       3,
		},
		{
			ID:          graphrag.TacticExecution,
			Name:        "Execution",
			Description: "Running malicious prompts, commands, or code on target AI systems",
			Phase:       4,
		},
		{
			ID:          graphrag.TacticPersistence,
			Name:        "Persistence",
			Description: "Maintaining access over time through memory poisoning, backdoors, or state manipulation",
			Phase:       5,
		},
		{
			ID:          graphrag.TacticPrivilegeEscalation,
			Name:        "Privilege Escalation",
			Description: "Gaining higher-level permissions, roles, or capabilities in AI systems",
			Phase:       6,
		},
		{
			ID:          graphrag.TacticDefenseEvasion,
			Name:        "Defense Evasion",
			Description: "Avoiding detection by guardrails, content filters, and security controls",
			Phase:       7,
		},
		{
			ID:          graphrag.TacticCredentialAccess,
			Name:        "Credential Access",
			Description: "Stealing API keys, tokens, credentials, or authentication secrets",
			Phase:       8,
		},
		{
			ID:          graphrag.TacticDiscovery,
			Name:        "Discovery",
			Description: "Exploring the target environment to understand capabilities, data, and vulnerabilities",
			Phase:       9,
		},
		{
			ID:          graphrag.TacticLateralMovement,
			Name:        "Lateral Movement",
			Description: "Moving through connected systems, agents, or AI infrastructure components",
			Phase:       10,
		},
		{
			ID:          graphrag.TacticCollection,
			Name:        "Collection",
			Description: "Gathering target data of interest from models, memory, or knowledge bases",
			Phase:       11,
		},
		{
			ID:          graphrag.TacticExfiltration,
			Name:        "Exfiltration",
			Description: "Stealing data, training data, or sensitive information from AI systems",
			Phase:       12,
		},
		{
			ID:          graphrag.TacticImpact,
			Name:        "Impact",
			Description: "Disrupting or destroying target systems, models, or data integrity",
			Phase:       13,
		},
		{
			ID:          graphrag.TacticAIManipulation,
			Name:        "AI Manipulation",
			Description: "AI/LLM-specific attack techniques for model manipulation and behavior control",
			Phase:       14,
		},
	}
}

// TechniqueRegistry provides access to predefined Gibson techniques.
type TechniqueRegistry struct{}

// NewTechniqueRegistry creates a new TechniqueRegistry instance.
func NewTechniqueRegistry() *TechniqueRegistry {
	return &TechniqueRegistry{}
}

// GetTechnique returns a specific technique by ID, or nil if not found.
func (r *TechniqueRegistry) GetTechnique(id string) *Technique {
	techniques := r.AllTechniques()
	for _, technique := range techniques {
		if technique.ID == id {
			return technique
		}
	}
	return nil
}

// AllTechniques returns all predefined Gibson techniques.
func (r *TechniqueRegistry) AllTechniques() []*Technique {
	return []*Technique{
		{
			ID:          graphrag.TechniquePromptInjection,
			Name:        "Prompt Injection",
			Description: "Direct injection of malicious prompts to override intended behavior",
			TacticIDs:   []string{graphrag.TacticInitialAccess, graphrag.TacticExecution},
			Detection:   "Monitor for unusual prompt patterns, system command keywords, or role-switching attempts",
			Mitigation:  "Input validation, prompt sandboxing, instruction hierarchy enforcement",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueJailbreak,
			Name:        "Jailbreak",
			Description: "Bypassing model safety restrictions and content filters through roleplay or encoding",
			TacticIDs:   []string{graphrag.TacticDefenseEvasion},
			Detection:   "Track roleplay scenarios, hypothetical framing, and character-based prompts",
			Mitigation:  "Multi-layer filtering, semantic analysis, context-aware guardrails",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueSystemPromptExtraction,
			Name:        "System Prompt Extraction",
			Description: "Extracting hidden system prompts through recursive queries or meta-prompting",
			TacticIDs:   []string{graphrag.TacticDiscovery, graphrag.TacticCollection},
			Detection:   "Monitor for self-referential queries, instruction extraction attempts",
			Mitigation:  "Separate system context, avoid echoing instructions, use sealed prompts",
			Severity:    "medium",
		},
		{
			ID:          graphrag.TechniqueTrainingDataExtraction,
			Name:        "Training Data Extraction",
			Description: "Extracting memorized training data through repeated or targeted queries",
			TacticIDs:   []string{graphrag.TacticCollection, graphrag.TacticExfiltration},
			Detection:   "Monitor for verbatim text repetition, memorization probing patterns",
			Mitigation:  "Data deduplication in training, output filtering, rate limiting",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueModelInversion,
			Name:        "Model Inversion",
			Description: "Inferring training data or sensitive information from model behavior",
			TacticIDs:   []string{graphrag.TacticCollection, graphrag.TacticExfiltration},
			Detection:   "Detect statistical probing patterns and inference attacks",
			Mitigation:  "Differential privacy in training, query result sanitization",
			Severity:    "medium",
		},
		{
			ID:          graphrag.TechniqueRAGPoisoning,
			Name:        "RAG Poisoning",
			Description: "Poisoning RAG knowledge base with malicious or misleading data",
			TacticIDs:   []string{graphrag.TacticPersistence, graphrag.TacticDefenseEvasion},
			Detection:   "Content provenance tracking, anomaly detection in retrieval results",
			Mitigation:  "Source verification, content integrity checks, trusted data sources",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueCitationInjection,
			Name:        "Citation Injection",
			Description: "Injecting false citations into RAG responses to spread misinformation",
			TacticIDs:   []string{graphrag.TacticPersistence, graphrag.TacticDefenseEvasion},
			Detection:   "Validate citation sources and document relationships",
			Mitigation:  "Citation verification, source authentication, metadata validation",
			Severity:    "medium",
		},
		{
			ID:          graphrag.TechniqueToolAbuse,
			Name:        "Tool Abuse",
			Description: "Abusing agent tools for malicious purposes beyond intended scope",
			TacticIDs:   []string{graphrag.TacticExecution, graphrag.TacticPrivilegeEscalation},
			Detection:   "Monitor tool usage patterns, parameter anomalies, privilege violations",
			Mitigation:  "Tool sandboxing, permission boundaries, usage auditing",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueAgentHijacking,
			Name:        "Agent Hijacking",
			Description: "Taking control of autonomous agents through goal manipulation or prompt override",
			TacticIDs:   []string{graphrag.TacticExecution, graphrag.TacticDefenseEvasion},
			Detection:   "Track goal changes, behavior deviations, unauthorized task execution",
			Mitigation:  "Goal validation, behavior monitoring, state integrity checks",
			Severity:    "critical",
		},
		{
			ID:          graphrag.TechniqueMCPToolInjection,
			Name:        "MCP Tool Injection",
			Description: "Injecting malicious MCP tools or manipulating tool definitions",
			TacticIDs:   []string{graphrag.TacticPersistence, graphrag.TacticExecution},
			Detection:   "Tool signature verification, schema validation, provenance tracking",
			Mitigation:  "Tool whitelisting, signature requirements, sandbox execution",
			Severity:    "critical",
		},
		{
			ID:          graphrag.TechniqueMemoryPoisoning,
			Name:        "Memory Poisoning",
			Description: "Poisoning agent memory stores with false information or malicious context",
			TacticIDs:   []string{graphrag.TacticPersistence, graphrag.TacticDefenseEvasion},
			Detection:   "Memory integrity checks, anomaly detection in stored context",
			Mitigation:  "Memory validation, context verification, trusted memory sources",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueGuardrailBypass,
			Name:        "Guardrail Bypass",
			Description: "Bypassing safety guardrails through encoding, language switching, or obfuscation",
			TacticIDs:   []string{graphrag.TacticDefenseEvasion},
			Detection:   "Multi-language analysis, encoding detection, semantic intent analysis",
			Mitigation:  "Multi-layer guardrails, semantic analysis, universal filtering",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueModelDoS,
			Name:        "Model DoS",
			Description: "Denial of service attacks on models through resource exhaustion or token overflow",
			TacticIDs:   []string{graphrag.TacticImpact},
			Detection:   "Rate limiting, resource monitoring, token consumption tracking",
			Mitigation:  "Input length limits, timeout enforcement, resource quotas",
			Severity:    "medium",
		},
		{
			ID:          graphrag.TechniqueEncodingObfuscation,
			Name:        "Encoding Obfuscation",
			Description: "Using encoding (Base64, hex, etc.) to evade content filters",
			TacticIDs:   []string{graphrag.TacticDefenseEvasion},
			Detection:   "Encoding detection, decoded content analysis",
			Mitigation:  "Decode inputs before filtering, multi-stage analysis",
			Severity:    "medium",
		},
		{
			ID:          graphrag.TechniqueLanguageSwitching,
			Name:        "Language Switching",
			Description: "Switching languages to evade English-only detection systems",
			TacticIDs:   []string{graphrag.TacticDefenseEvasion},
			Detection:   "Multi-language content analysis, translation-based detection",
			Mitigation:  "Universal language filtering, translation before analysis",
			Severity:    "medium",
		},
		{
			ID:          graphrag.TechniqueTokenSmuggling,
			Name:        "Token Smuggling",
			Description: "Token-level attack techniques using special tokens or tokenization quirks",
			TacticIDs:   []string{graphrag.TacticDefenseEvasion},
			Detection:   "Token-level analysis, special token monitoring",
			Mitigation:  "Token normalization, special token filtering",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniqueInstructionHierarchy,
			Name:        "Instruction Hierarchy",
			Description: "Exploiting instruction priority conflicts between system, user, and retrieved content",
			TacticIDs:   []string{graphrag.TacticExecution, graphrag.TacticPrivilegeEscalation},
			Detection:   "Instruction conflict detection, priority violation monitoring",
			Mitigation:  "Clear instruction hierarchy, priority enforcement, conflict resolution",
			Severity:    "high",
		},
		{
			ID:          graphrag.TechniquePayloadSplitting,
			Name:        "Payload Splitting",
			Description: "Splitting payloads across multiple turns to evade single-turn detection",
			TacticIDs:   []string{graphrag.TacticInitialAccess, graphrag.TacticExecution},
			Detection:   "Cross-turn context analysis, conversation state tracking",
			Mitigation:  "Stateful filtering, conversation-level analysis",
			Severity:    "medium",
		},
		{
			ID:          graphrag.TechniqueIndirectPromptInjection,
			Name:        "Indirect Prompt Injection",
			Description: "Injection via external data sources (web pages, documents, emails)",
			TacticIDs:   []string{graphrag.TacticInitialAccess, graphrag.TacticExecution},
			Detection:   "External content sandboxing, provenance tracking",
			Mitigation:  "Content sanitization, source trust boundaries, context isolation",
			Severity:    "critical",
		},
		{
			ID:          graphrag.TechniqueMultiModalInjection,
			Name:        "Multi-Modal Injection",
			Description: "Injection via images, audio, or video with embedded adversarial content",
			TacticIDs:   []string{graphrag.TacticInitialAccess, graphrag.TacticExecution},
			Detection:   "Multi-modal content analysis, adversarial input detection",
			Mitigation:  "Multi-modal filtering, content validation, modality-specific guardrails",
			Severity:    "high",
		},
	}
}

// TechniquesByTactic returns all techniques that help achieve a specific tactic.
func (r *TechniqueRegistry) TechniquesByTactic(tacticID string) []*Technique {
	var results []*Technique
	for _, technique := range r.AllTechniques() {
		for _, tid := range technique.TacticIDs {
			if tid == tacticID {
				results = append(results, technique)
				break
			}
		}
	}
	return results
}

package graphrag

import (
	"log"
	"strings"
	"sync"
)

// TaxonomyReader provides read-only access to the taxonomy from agent code.
// This interface is intentionally minimal to prevent agents from modifying the taxonomy.
type TaxonomyReader interface {
	// Version returns the taxonomy version string.
	Version() string

	// IsCanonicalNodeType checks if a node type is in the canonical taxonomy.
	IsCanonicalNodeType(typeName string) bool

	// IsCanonicalRelationType checks if a relationship type is in the canonical taxonomy.
	IsCanonicalRelationType(typeName string) bool

	// ValidateNodeType checks if a node type string is valid. Returns true if valid.
	// Logs a warning if the type is not canonical but doesn't fail.
	ValidateNodeType(typeName string) bool

	// ValidateRelationType checks if a relationship type string is valid. Returns true if valid.
	// Logs a warning if the type is not canonical but doesn't fail.
	ValidateRelationType(typeName string) bool
}

// TaxonomyIntrospector extends TaxonomyReader with methods to list all types.
// This interface enables dynamic/LLM-driven agents to query the taxonomy at runtime.
type TaxonomyIntrospector interface {
	TaxonomyReader

	// NodeTypes returns all canonical node type names.
	NodeTypes() []string

	// RelationshipTypes returns all canonical relationship type names.
	RelationshipTypes() []string

	// TechniqueIDs returns all technique IDs, optionally filtered by source ("mitre", "arcanum", or "" for all).
	TechniqueIDs(source string) []string

	// NodeTypeInfo returns detailed info about a node type (description, properties, etc).
	NodeTypeInfo(typeName string) *NodeTypeInfo

	// RelationshipTypeInfo returns detailed info about a relationship type.
	RelationshipTypeInfo(typeName string) *RelationshipTypeInfo

	// TechniqueInfo returns detailed info about a technique.
	TechniqueInfo(techniqueID string) *TechniqueInfo
}

// NodeTypeInfo contains metadata about a node type for introspection.
type NodeTypeInfo struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Category    string         `json:"category"`
	Description string         `json:"description"`
	Properties  []PropertyInfo `json:"properties"`
}

// RelationshipTypeInfo contains metadata about a relationship type for introspection.
type RelationshipTypeInfo struct {
	Type          string   `json:"type"`
	Name          string   `json:"name"`
	Category      string   `json:"category"`
	Description   string   `json:"description"`
	FromTypes     []string `json:"from_types"`
	ToTypes       []string `json:"to_types"`
	Bidirectional bool     `json:"bidirectional"`
}

// TechniqueInfo contains metadata about an attack technique for introspection.
type TechniqueInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Taxonomy    string `json:"taxonomy"` // "mitre" or "arcanum"
	Description string `json:"description"`
	Tactic      string `json:"tactic,omitempty"`
	URL         string `json:"url,omitempty"`
}

// PropertyInfo describes a node or relationship property.
type PropertyInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

var (
	// Global taxonomy instance
	globalTaxonomy TaxonomyReader
	taxonomyMu     sync.RWMutex
)

// SetTaxonomy sets the global taxonomy instance.
// This should only be called by the Gibson harness during initialization.
func SetTaxonomy(taxonomy TaxonomyReader) {
	taxonomyMu.Lock()
	defer taxonomyMu.Unlock()
	globalTaxonomy = taxonomy
}

// Taxonomy returns the global taxonomy reader.
// Returns nil if taxonomy has not been initialized.
func Taxonomy() TaxonomyReader {
	taxonomyMu.RLock()
	defer taxonomyMu.RUnlock()
	return globalTaxonomy
}

// NewNodeWithValidation creates a new GraphNode with taxonomy validation.
// If the node type is not in the canonical taxonomy, a warning is logged but the node is still created.
// This allows agents to use custom types while encouraging use of canonical types.
func NewNodeWithValidation(nodeType string) *GraphNode {
	// Get taxonomy if available
	tax := Taxonomy()
	if tax != nil {
		// Validate node type
		if !tax.IsCanonicalNodeType(nodeType) {
			log.Printf("WARNING: Node type '%s' is not in the canonical taxonomy. Consider using a canonical type from taxonomy_generated.go", nodeType)
		}
	}

	// Create node using standard constructor
	return NewGraphNode(nodeType)
}

// ValidateAndWarnNodeType validates a node type and logs a warning if not canonical.
// This is a convenience function for agents that want to validate before creating nodes.
func ValidateAndWarnNodeType(nodeType string) bool {
	tax := Taxonomy()
	if tax == nil {
		// No taxonomy loaded, allow any type
		return true
	}

	if !tax.IsCanonicalNodeType(nodeType) {
		log.Printf("WARNING: Node type '%s' is not canonical. Available types: see taxonomy_generated.go", nodeType)
		return false
	}

	return true
}

// ValidateAndWarnRelationType validates a relationship type and logs a warning if not canonical.
// This is a convenience function for agents that want to validate before creating relationships.
func ValidateAndWarnRelationType(relType string) bool {
	tax := Taxonomy()
	if tax == nil {
		// No taxonomy loaded, allow any type
		return true
	}

	if !tax.IsCanonicalRelationType(relType) {
		log.Printf("WARNING: Relationship type '%s' is not canonical. Available types: see taxonomy_generated.go", relType)
		return false
	}

	return true
}

// NewRelationshipWithValidation creates a new relationship with taxonomy validation.
// If the relationship type is not in the canonical taxonomy, a warning is logged but the relationship is still created.
func NewRelationshipWithValidation(fromID, toID, relType string) *Relationship {
	// Get taxonomy if available
	tax := Taxonomy()
	if tax != nil {
		// Validate relationship type
		if !tax.IsCanonicalRelationType(relType) {
			log.Printf("WARNING: Relationship type '%s' is not in the canonical taxonomy. Consider using a canonical type from taxonomy_generated.go", relType)
		}
	}

	// Create relationship using standard constructor
	return NewRelationship(fromID, toID, relType)
}

// TaxonomyIntrospect returns the taxonomy as a TaxonomyIntrospector if it supports introspection.
// Returns nil if taxonomy is not loaded or doesn't support introspection.
func TaxonomyIntrospect() TaxonomyIntrospector {
	taxonomyMu.RLock()
	defer taxonomyMu.RUnlock()

	if globalTaxonomy == nil {
		return nil
	}

	// Check if it implements TaxonomyIntrospector
	if intro, ok := globalTaxonomy.(TaxonomyIntrospector); ok {
		return intro
	}

	return nil
}

// TaxonomyPrompt generates a prompt-friendly description of the taxonomy for LLM consumption.
// This enables dynamic agents to query the taxonomy and let the LLM structure data correctly.
func TaxonomyPrompt() string {
	intro := TaxonomyIntrospect()
	if intro == nil {
		// Fall back to generated constants if introspection not available
		return taxonomyPromptFromConstants()
	}

	return taxonomyPromptFromIntrospector(intro)
}

// taxonomyPromptFromConstants builds a taxonomy prompt from the generated constants.
func taxonomyPromptFromConstants() string {
	return `# GraphRAG Taxonomy

## Node Types (Assets)
- domain: Root domain entity (e.g., example.com)
- subdomain: Subdomain under a root domain
- host: IP address or hostname
- port: Network port on a host
- service: Service running on a port
- endpoint: Web endpoint or URL
- api: Web API with multiple endpoints
- technology: Software/framework detected
- cloud_asset: Cloud infrastructure resource
- certificate: TLS/SSL certificate

## Node Types (Findings)
- finding: Security vulnerability or issue
- evidence: Supporting evidence for a finding
- mitigation: Remediation measure

## Node Types (Execution)
- mission: Top-level security assessment
- agent_run: Single agent execution within a mission
- tool_execution: Tool invocation
- llm_call: LLM API call

## Node Types (Attack)
- technique: Attack technique (MITRE/Arcanum)
- tactic: High-level adversary goal

## Relationship Types (Asset Hierarchy)
- HAS_SUBDOMAIN: Domain → Subdomain
- RESOLVES_TO: Domain/Subdomain → Host
- HAS_PORT: Host → Port
- RUNS_SERVICE: Port → Service
- HAS_ENDPOINT: Service → Endpoint
- USES_TECHNOLOGY: Service/Endpoint → Technology
- SERVES_CERTIFICATE: Host → Certificate
- HOSTS: CloudAsset → Host/Service

## Relationship Types (Finding Links)
- AFFECTS: Finding → Asset
- HAS_EVIDENCE: Finding → Evidence
- USES_TECHNIQUE: Finding → Technique
- EXPLOITS: Finding → Technology
- MITIGATES: Mitigation → Finding
- LEADS_TO: Finding → Finding (attack chains)
- SIMILAR_TO: Finding ↔ Finding (bidirectional)

## Relationship Types (Execution Context)
- PART_OF: AgentRun → Mission
- EXECUTED_BY: ToolExecution → AgentRun
- DISCOVERED: AgentRun → Asset
- PRODUCED: ToolExecution → Finding
- MADE_CALL: AgentRun → LlmCall

## Common Properties
- name: Entity name
- title: Finding title
- description: Detailed description
- severity: low/medium/high/critical
- confidence: 0-100
- url: URL or endpoint
- ip: IP address
- port: Port number
- protocol: tcp/udp
- method: HTTP method
- timestamp: Unix timestamp
`
}

// taxonomyPromptFromIntrospector builds a detailed taxonomy prompt from introspection.
func taxonomyPromptFromIntrospector(intro TaxonomyIntrospector) string {
	var sb strings.Builder

	sb.WriteString("# GraphRAG Taxonomy v")
	sb.WriteString(intro.Version())
	sb.WriteString("\n\n")

	// Node types
	sb.WriteString("## Node Types\n\n")
	for _, nt := range intro.NodeTypes() {
		info := intro.NodeTypeInfo(nt)
		if info != nil {
			sb.WriteString("### ")
			sb.WriteString(info.Type)
			sb.WriteString(" (")
			sb.WriteString(info.Category)
			sb.WriteString(")\n")
			sb.WriteString(info.Description)
			sb.WriteString("\n")

			if len(info.Properties) > 0 {
				sb.WriteString("Properties:\n")
				for _, p := range info.Properties {
					sb.WriteString("- ")
					sb.WriteString(p.Name)
					sb.WriteString(" (")
					sb.WriteString(p.Type)
					if p.Required {
						sb.WriteString(", required")
					}
					sb.WriteString(")")
					if p.Description != "" {
						sb.WriteString(": ")
						sb.WriteString(p.Description)
					}
					sb.WriteString("\n")
				}
			}
			sb.WriteString("\n")
		}
	}

	// Relationship types
	sb.WriteString("## Relationship Types\n\n")
	for _, rt := range intro.RelationshipTypes() {
		info := intro.RelationshipTypeInfo(rt)
		if info != nil {
			sb.WriteString("### ")
			sb.WriteString(info.Type)
			sb.WriteString("\n")
			sb.WriteString(info.Description)
			sb.WriteString("\n")
			sb.WriteString("From: ")
			sb.WriteString(strings.Join(info.FromTypes, ", "))
			sb.WriteString(" → To: ")
			sb.WriteString(strings.Join(info.ToTypes, ", "))
			if info.Bidirectional {
				sb.WriteString(" (bidirectional)")
			}
			sb.WriteString("\n\n")
		}
	}

	// Techniques
	sb.WriteString("## Attack Techniques\n\n")
	sb.WriteString("### MITRE ATT&CK\n")
	for _, tid := range intro.TechniqueIDs("mitre") {
		info := intro.TechniqueInfo(tid)
		if info != nil {
			sb.WriteString("- ")
			sb.WriteString(info.ID)
			sb.WriteString(": ")
			sb.WriteString(info.Name)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n### Arcanum Prompt Injection\n")
	for _, tid := range intro.TechniqueIDs("arcanum") {
		info := intro.TechniqueInfo(tid)
		if info != nil {
			sb.WriteString("- ")
			sb.WriteString(info.ID)
			sb.WriteString(": ")
			sb.WriteString(info.Name)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// TaxonomyJSON returns the taxonomy as a JSON-serializable structure for LLM tool use.
// This can be included in a tool response or system prompt for structured output.
func TaxonomyJSON() map[string]any {
	intro := TaxonomyIntrospect()

	result := map[string]any{
		"version": TaxonomyVersion,
	}

	// Node types
	nodeTypes := make([]map[string]any, 0)
	if intro != nil {
		for _, nt := range intro.NodeTypes() {
			info := intro.NodeTypeInfo(nt)
			if info != nil {
				nodeTypes = append(nodeTypes, map[string]any{
					"type":        info.Type,
					"name":        info.Name,
					"category":    info.Category,
					"description": info.Description,
				})
			}
		}
	} else {
		// Fall back to constants
		for _, nt := range []string{
			NodeTypeDomain, NodeTypeSubdomain, NodeTypeHost, NodeTypePort,
			NodeTypeService, NodeTypeEndpoint, NodeTypeApi, NodeTypeTechnology,
			NodeTypeCloudAsset, NodeTypeCertificate, NodeTypeFinding,
			NodeTypeEvidence, NodeTypeMitigation, NodeTypeMission,
			NodeTypeAgentRun, NodeTypeToolExecution, NodeTypeLlmCall,
			NodeTypeTechnique, NodeTypeTactic,
		} {
			nodeTypes = append(nodeTypes, map[string]any{"type": nt})
		}
	}
	result["node_types"] = nodeTypes

	// Relationship types
	relTypes := make([]map[string]any, 0)
	if intro != nil {
		for _, rt := range intro.RelationshipTypes() {
			info := intro.RelationshipTypeInfo(rt)
			if info != nil {
				relTypes = append(relTypes, map[string]any{
					"type":          info.Type,
					"name":          info.Name,
					"description":   info.Description,
					"from_types":    info.FromTypes,
					"to_types":      info.ToTypes,
					"bidirectional": info.Bidirectional,
				})
			}
		}
	} else {
		// Fall back to constants
		for _, rt := range []string{
			RelTypeHasSubdomain, RelTypeResolvesTo, RelTypeHasPort, RelTypeRunsService,
			RelTypeHasEndpoint, RelTypeUsesTechnology, RelTypeServesCertificate, RelTypeHosts,
			RelTypeAffects, RelTypeHasEvidence, RelTypeUsesTechnique, RelTypeExploits,
			RelTypeMitigates, RelTypeLeadsTo, RelTypeSimilarTo, RelTypePartOf,
			RelTypeExecutedBy, RelTypeDiscovered, RelTypeProduced, RelTypeMadeCall,
		} {
			relTypes = append(relTypes, map[string]any{"type": rt})
		}
	}
	result["relationship_types"] = relTypes

	// Techniques
	techniques := make([]map[string]any, 0)
	if intro != nil {
		for _, tid := range intro.TechniqueIDs("") {
			info := intro.TechniqueInfo(tid)
			if info != nil {
				techniques = append(techniques, map[string]any{
					"id":       info.ID,
					"name":     info.Name,
					"taxonomy": info.Taxonomy,
				})
			}
		}
	}
	result["techniques"] = techniques

	return result
}

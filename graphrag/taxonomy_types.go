package graphrag

import (
	"fmt"
	"sync"
)

// TaxonomyIntrospector provides runtime access to taxonomy metadata.
// This interface allows agents to query the taxonomy schema dynamically.
type TaxonomyIntrospector interface {
	// Version returns the taxonomy version.
	Version() string

	// NodeTypes returns all registered node type names.
	NodeTypes() []string

	// NodeTypeInfo returns metadata for a specific node type.
	// Returns nil if the node type is not found.
	NodeTypeInfo(nodeType string) *NodeTypeInfo

	// RelationshipTypes returns all registered relationship type names.
	RelationshipTypes() []string

	// RelationshipTypeInfo returns metadata for a specific relationship type.
	// Returns nil if the relationship type is not found.
	RelationshipTypeInfo(relType string) *RelationshipTypeInfo

	// TechniqueIDs returns all technique IDs, optionally filtered by taxonomy.
	// Pass empty string to get all techniques.
	TechniqueIDs(taxonomy string) []string

	// TechniqueInfo returns metadata for a specific technique.
	// Returns nil if the technique is not found.
	TechniqueInfo(techniqueID string) *TechniqueInfo
}

// NodeTypeInfo contains metadata about a node type.
type NodeTypeInfo struct {
	Type        string // e.g., "host", "port"
	Name        string // Human-readable name
	Category    string // e.g., "asset", "finding", "execution"
	Description string
	Properties  []PropertyInfo
}

// PropertyInfo contains metadata about a property.
type PropertyInfo struct {
	Name        string
	Type        string // e.g., "string", "int32", "float64"
	Required    bool
	Description string
	Format      string
	Enum        []string
}

// RelationshipTypeInfo contains metadata about a relationship type.
type RelationshipTypeInfo struct {
	Type          string // e.g., "HAS_PORT", "RUNS_SERVICE"
	Name          string // Human-readable name
	Category      string
	Description   string
	FromTypes     []string
	ToTypes       []string
	Bidirectional bool
}

// TechniqueInfo contains metadata about an attack technique.
type TechniqueInfo struct {
	ID          string
	Name        string
	Description string
	Taxonomy    string // e.g., "gibson", "mitre"
	Tactic      string
	URL         string
}

// Global taxonomy instance
var (
	globalTaxonomyMu sync.RWMutex
	globalTaxonomy   TaxonomyIntrospector
)

// SetTaxonomy sets the global taxonomy instance.
func SetTaxonomy(t TaxonomyIntrospector) {
	globalTaxonomyMu.Lock()
	defer globalTaxonomyMu.Unlock()
	globalTaxonomy = t
}

// GetTaxonomy returns the global taxonomy instance.
func GetTaxonomy() TaxonomyIntrospector {
	globalTaxonomyMu.RLock()
	defer globalTaxonomyMu.RUnlock()
	return globalTaxonomy
}

// TaxonomyRegistry manages taxonomy extensions from agents and plugins.
// This allows runtime extension of the core taxonomy with custom node types
// and relationships.
type TaxonomyRegistry interface {
	// RegisterExtension adds custom taxonomy definitions from an agent or plugin.
	RegisterExtension(name string, ext TaxonomyExtension) error

	// UnregisterExtension removes taxonomy definitions from an agent or plugin.
	UnregisterExtension(name string) error

	// GetExtension returns the taxonomy extension for a registered name.
	GetExtension(name string) (TaxonomyExtension, bool)

	// AllExtensions returns all registered taxonomy extensions.
	AllExtensions() map[string]TaxonomyExtension
}

// TaxonomyExtension contains custom taxonomy definitions contributed by an agent or plugin.
type TaxonomyExtension struct {
	NodeTypes     []NodeTypeDefinition
	Relationships []RelationshipDefinition
}

// NodeTypeDefinition defines a custom node type.
type NodeTypeDefinition struct {
	Name        string
	Category    string
	Description string
	Properties  []PropertyInfo
}

// RelationshipDefinition defines a custom relationship type.
type RelationshipDefinition struct {
	Name        string
	Category    string
	Description string
	FromTypes   []string
	ToTypes     []string
}

// ==================== CONCRETE IMPLEMENTATIONS ====================

// SimpleTaxonomy is a concrete implementation of TaxonomyIntrospector
// using the generated constants from constants_generated.go.
type SimpleTaxonomy struct {
	version        string
	nodeTypeInfo   map[string]*NodeTypeInfo
	relTypeInfo    map[string]*RelationshipTypeInfo
	techniqueInfo  map[string]*TechniqueInfo
	techniquesByTx map[string][]string // taxonomy -> technique IDs
}

// NewSimpleTaxonomy creates a new SimpleTaxonomy using the generated taxonomy data.
func NewSimpleTaxonomy() *SimpleTaxonomy {
	t := &SimpleTaxonomy{
		version:        "3.0.0", // From constants_generated.go header
		nodeTypeInfo:   make(map[string]*NodeTypeInfo),
		relTypeInfo:    make(map[string]*RelationshipTypeInfo),
		techniqueInfo:  make(map[string]*TechniqueInfo),
		techniquesByTx: make(map[string][]string),
	}
	t.initFromGenerated()
	return t
}

// initFromGenerated initializes the taxonomy from generated constants.
func (t *SimpleTaxonomy) initFromGenerated() {
	// Initialize node types from AllNodeTypes
	categoryMap := map[string]string{
		"mission":        "execution",
		"mission_run":    "execution",
		"agent_run":      "execution",
		"tool_execution": "execution",
		"llm_call":       "execution",
		"domain":         "asset",
		"subdomain":      "asset",
		"host":           "asset",
		"port":           "asset",
		"service":        "asset",
		"endpoint":       "asset",
		"technology":     "asset",
		"certificate":    "asset",
		"finding":        "finding",
		"evidence":       "finding",
		"technique":      "attack",
	}

	for _, nt := range AllNodeTypes {
		t.nodeTypeInfo[nt] = &NodeTypeInfo{
			Type:     nt,
			Name:     nt,
			Category: categoryMap[nt],
		}
	}

	// Initialize relationship types from AllRelationshipTypes
	for _, rt := range AllRelationshipTypes {
		t.relTypeInfo[rt] = &RelationshipTypeInfo{
			Type: rt,
			Name: rt,
		}
	}
}

// Version returns the taxonomy version.
func (t *SimpleTaxonomy) Version() string {
	return t.version
}

// NodeTypes returns all registered node type names.
func (t *SimpleTaxonomy) NodeTypes() []string {
	return AllNodeTypes
}

// NodeTypeInfo returns metadata for a specific node type.
func (t *SimpleTaxonomy) NodeTypeInfo(nodeType string) *NodeTypeInfo {
	return t.nodeTypeInfo[nodeType]
}

// RelationshipTypes returns all registered relationship type names.
func (t *SimpleTaxonomy) RelationshipTypes() []string {
	return AllRelationshipTypes
}

// RelationshipTypeInfo returns metadata for a specific relationship type.
func (t *SimpleTaxonomy) RelationshipTypeInfo(relType string) *RelationshipTypeInfo {
	return t.relTypeInfo[relType]
}

// TechniqueIDs returns all technique IDs, optionally filtered by taxonomy.
func (t *SimpleTaxonomy) TechniqueIDs(taxonomy string) []string {
	if taxonomy == "" {
		// Return all techniques
		var ids []string
		for id := range t.techniqueInfo {
			ids = append(ids, id)
		}
		return ids
	}
	return t.techniquesByTx[taxonomy]
}

// TechniqueInfo returns metadata for a specific technique.
func (t *SimpleTaxonomy) TechniqueInfo(techniqueID string) *TechniqueInfo {
	return t.techniqueInfo[techniqueID]
}

// ==================== DEFAULT TAXONOMY REGISTRY ====================

// DefaultTaxonomyRegistry is a concrete implementation of TaxonomyRegistry.
type DefaultTaxonomyRegistry struct {
	mu         sync.RWMutex
	core       TaxonomyIntrospector
	extensions map[string]TaxonomyExtension
}

// NewTaxonomyRegistry creates a new DefaultTaxonomyRegistry with the given core taxonomy.
func NewTaxonomyRegistry(core TaxonomyIntrospector) *DefaultTaxonomyRegistry {
	return &DefaultTaxonomyRegistry{
		core:       core,
		extensions: make(map[string]TaxonomyExtension),
	}
}

// RegisterExtension adds custom taxonomy definitions from an agent or plugin.
func (r *DefaultTaxonomyRegistry) RegisterExtension(name string, ext TaxonomyExtension) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.extensions[name]; exists {
		return fmt.Errorf("taxonomy extension already registered: %s", name)
	}
	r.extensions[name] = ext
	return nil
}

// UnregisterExtension removes taxonomy definitions from an agent or plugin.
func (r *DefaultTaxonomyRegistry) UnregisterExtension(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.extensions[name]; !exists {
		return fmt.Errorf("taxonomy extension not registered: %s", name)
	}
	delete(r.extensions, name)
	return nil
}

// GetExtension returns the taxonomy extension for a registered name.
func (r *DefaultTaxonomyRegistry) GetExtension(name string) (TaxonomyExtension, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ext, ok := r.extensions[name]
	return ext, ok
}

// AllExtensions returns all registered taxonomy extensions.
func (r *DefaultTaxonomyRegistry) AllExtensions() map[string]TaxonomyExtension {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]TaxonomyExtension, len(r.extensions))
	for k, v := range r.extensions {
		result[k] = v
	}
	return result
}

// Core returns the core taxonomy.
func (r *DefaultTaxonomyRegistry) Core() TaxonomyIntrospector {
	return r.core
}

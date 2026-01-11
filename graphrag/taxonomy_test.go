package graphrag

import (
	"bytes"
	"log"
	"testing"
)

// mockTaxonomyReader implements TaxonomyReader for testing
type mockTaxonomyReader struct {
	version           string
	canonicalNodes    map[string]bool
	canonicalRels     map[string]bool
}

func newMockTaxonomyReader() *mockTaxonomyReader {
	return &mockTaxonomyReader{
		version: "test-1.0.0",
		canonicalNodes: map[string]bool{
			"domain":    true,
			"subdomain": true,
			"host":      true,
			"finding":   true,
		},
		canonicalRels: map[string]bool{
			"HAS_SUBDOMAIN":  true,
			"RESOLVES_TO":    true,
			"AFFECTS":        true,
		},
	}
}

func (m *mockTaxonomyReader) Version() string {
	return m.version
}

func (m *mockTaxonomyReader) IsCanonicalNodeType(typeName string) bool {
	return m.canonicalNodes[typeName]
}

func (m *mockTaxonomyReader) IsCanonicalRelationType(typeName string) bool {
	return m.canonicalRels[typeName]
}

func (m *mockTaxonomyReader) ValidateNodeType(typeName string) bool {
	if !m.IsCanonicalNodeType(typeName) {
		log.Printf("WARNING: Node type '%s' is not canonical", typeName)
	}
	return true
}

func (m *mockTaxonomyReader) ValidateRelationType(typeName string) bool {
	if !m.IsCanonicalRelationType(typeName) {
		log.Printf("WARNING: Relationship type '%s' is not canonical", typeName)
	}
	return true
}

func TestSetTaxonomy(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Create mock taxonomy
	mock := newMockTaxonomyReader()

	// Set taxonomy
	SetTaxonomy(mock)

	// Verify it was set
	tax := Taxonomy()
	if tax == nil {
		t.Fatal("Taxonomy() returned nil after SetTaxonomy")
	}

	if tax.Version() != "test-1.0.0" {
		t.Errorf("Taxonomy().Version() = %v, want test-1.0.0", tax.Version())
	}
}

func TestTaxonomy_NilWhenNotSet(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set to nil
	SetTaxonomy(nil)

	// Verify nil is returned
	if tax := Taxonomy(); tax != nil {
		t.Errorf("Taxonomy() = %v, want nil", tax)
	}
}

func TestNewNodeWithValidation_CanonicalType(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set mock taxonomy
	mock := newMockTaxonomyReader()
	SetTaxonomy(mock)

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	// Create node with canonical type - should NOT warn
	node := NewNodeWithValidation("domain")

	if node == nil {
		t.Fatal("NewNodeWithValidation() returned nil")
	}

	if node.Type != "domain" {
		t.Errorf("NewNodeWithValidation() type = %v, want domain", node.Type)
	}

	// Should not have logged a warning
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("NewNodeWithValidation() logged warning for canonical type: %v", logOutput)
	}
}

func TestNewNodeWithValidation_NonCanonicalType(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set mock taxonomy
	mock := newMockTaxonomyReader()
	SetTaxonomy(mock)

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	// Create node with non-canonical type - should warn
	node := NewNodeWithValidation("custom_type")

	if node == nil {
		t.Fatal("NewNodeWithValidation() returned nil")
	}

	if node.Type != "custom_type" {
		t.Errorf("NewNodeWithValidation() type = %v, want custom_type", node.Type)
	}

	// Should have logged a warning
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("NewNodeWithValidation() did not log warning for non-canonical type")
	}
	if !bytes.Contains(buf.Bytes(), []byte("WARNING")) {
		t.Errorf("NewNodeWithValidation() log should contain WARNING, got: %v", logOutput)
	}
}

func TestNewNodeWithValidation_NoTaxonomy(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set to nil
	SetTaxonomy(nil)

	// Should still create node without error
	node := NewNodeWithValidation("any_type")

	if node == nil {
		t.Fatal("NewNodeWithValidation() returned nil when taxonomy is nil")
	}

	if node.Type != "any_type" {
		t.Errorf("NewNodeWithValidation() type = %v, want any_type", node.Type)
	}
}

func TestValidateAndWarnNodeType(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set mock taxonomy
	mock := newMockTaxonomyReader()
	SetTaxonomy(mock)

	tests := []struct {
		name     string
		nodeType string
		want     bool
	}{
		{
			name:     "canonical type",
			nodeType: "domain",
			want:     true,
		},
		{
			name:     "non-canonical type",
			nodeType: "custom_type",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			got := ValidateAndWarnNodeType(tt.nodeType)

			if got != tt.want {
				t.Errorf("ValidateAndWarnNodeType() = %v, want %v", got, tt.want)
			}

			// Check if warning was logged for non-canonical types
			if !tt.want {
				logOutput := buf.String()
				if logOutput == "" {
					t.Error("ValidateAndWarnNodeType() did not log warning for non-canonical type")
				}
			}
		})
	}
}

func TestValidateAndWarnNodeType_NoTaxonomy(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set to nil
	SetTaxonomy(nil)

	// Should return true when no taxonomy is loaded
	if got := ValidateAndWarnNodeType("any_type"); !got {
		t.Errorf("ValidateAndWarnNodeType() = %v, want true when no taxonomy", got)
	}
}

func TestValidateAndWarnRelationType(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set mock taxonomy
	mock := newMockTaxonomyReader()
	SetTaxonomy(mock)

	tests := []struct {
		name    string
		relType string
		want    bool
	}{
		{
			name:    "canonical type",
			relType: "HAS_SUBDOMAIN",
			want:    true,
		},
		{
			name:    "non-canonical type",
			relType: "CUSTOM_REL",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			got := ValidateAndWarnRelationType(tt.relType)

			if got != tt.want {
				t.Errorf("ValidateAndWarnRelationType() = %v, want %v", got, tt.want)
			}

			// Check if warning was logged for non-canonical types
			if !tt.want {
				logOutput := buf.String()
				if logOutput == "" {
					t.Error("ValidateAndWarnRelationType() did not log warning for non-canonical type")
				}
			}
		})
	}
}

func TestValidateAndWarnRelationType_NoTaxonomy(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set to nil
	SetTaxonomy(nil)

	// Should return true when no taxonomy is loaded
	if got := ValidateAndWarnRelationType("any_type"); !got {
		t.Errorf("ValidateAndWarnRelationType() = %v, want true when no taxonomy", got)
	}
}

func TestNewRelationshipWithValidation_CanonicalType(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set mock taxonomy
	mock := newMockTaxonomyReader()
	SetTaxonomy(mock)

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	// Create relationship with canonical type - should NOT warn
	rel := NewRelationshipWithValidation("from-id", "to-id", "HAS_SUBDOMAIN")

	if rel == nil {
		t.Fatal("NewRelationshipWithValidation() returned nil")
	}

	if rel.Type != "HAS_SUBDOMAIN" {
		t.Errorf("NewRelationshipWithValidation() type = %v, want HAS_SUBDOMAIN", rel.Type)
	}

	// Should not have logged a warning
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("NewRelationshipWithValidation() logged warning for canonical type: %v", logOutput)
	}
}

func TestNewRelationshipWithValidation_NonCanonicalType(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set mock taxonomy
	mock := newMockTaxonomyReader()
	SetTaxonomy(mock)

	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	// Create relationship with non-canonical type - should warn
	rel := NewRelationshipWithValidation("from-id", "to-id", "CUSTOM_REL")

	if rel == nil {
		t.Fatal("NewRelationshipWithValidation() returned nil")
	}

	if rel.Type != "CUSTOM_REL" {
		t.Errorf("NewRelationshipWithValidation() type = %v, want CUSTOM_REL", rel.Type)
	}

	// Should have logged a warning
	logOutput := buf.String()
	if logOutput == "" {
		t.Error("NewRelationshipWithValidation() did not log warning for non-canonical type")
	}
	if !bytes.Contains(buf.Bytes(), []byte("WARNING")) {
		t.Errorf("NewRelationshipWithValidation() log should contain WARNING, got: %v", logOutput)
	}
}

func TestNewRelationshipWithValidation_NoTaxonomy(t *testing.T) {
	// Save original taxonomy and restore after test
	originalTax := Taxonomy()
	defer SetTaxonomy(originalTax)

	// Set to nil
	SetTaxonomy(nil)

	// Should still create relationship without error
	rel := NewRelationshipWithValidation("from-id", "to-id", "any_type")

	if rel == nil {
		t.Fatal("NewRelationshipWithValidation() returned nil when taxonomy is nil")
	}

	if rel.Type != "any_type" {
		t.Errorf("NewRelationshipWithValidation() type = %v, want any_type", rel.Type)
	}
}

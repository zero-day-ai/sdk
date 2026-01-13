package schema

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestTaxonomyMappingJSONRoundTrip tests JSON serialization and deserialization of TaxonomyMapping
func TestTaxonomyMappingJSONRoundTrip(t *testing.T) {
	original := TaxonomyMapping{
		NodeType:   "Asset",
		IDTemplate: "asset:{{.hostname}}",
		Properties: []PropertyMapping{
			{
				Source: "hostname",
				Target: "name",
			},
			{
				Source:  "ip_address",
				Target:  "ip",
				Default: "0.0.0.0",
			},
			{
				Source:    "domain",
				Target:    "domain",
				Transform: "lowercase",
			},
		},
		Relationships: []RelationshipMapping{
			{
				Type:         "HAS_VULNERABILITY",
				FromTemplate: "asset:{{.hostname}}",
				ToTemplate:   "vuln:{{.cve_id}}",
			},
			{
				Type:         "AFFECTS",
				FromTemplate: "vuln:{{.cve_id}}",
				ToTemplate:   "asset:{{.hostname}}",
				Condition:    "{{.severity}} == 'critical'",
				Properties: []PropertyMapping{
					{
						Source: "discovered_at",
						Target: "timestamp",
					},
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal TaxonomyMapping: %v", err)
	}

	// Unmarshal back
	var result TaxonomyMapping
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal TaxonomyMapping: %v", err)
	}

	// Verify fields
	if result.NodeType != original.NodeType {
		t.Errorf("NodeType mismatch: got %q, want %q", result.NodeType, original.NodeType)
	}
	if result.IDTemplate != original.IDTemplate {
		t.Errorf("IDTemplate mismatch: got %q, want %q", result.IDTemplate, original.IDTemplate)
	}
	if len(result.Properties) != len(original.Properties) {
		t.Errorf("Properties length mismatch: got %d, want %d", len(result.Properties), len(original.Properties))
	}
	if len(result.Relationships) != len(original.Relationships) {
		t.Errorf("Relationships length mismatch: got %d, want %d", len(result.Relationships), len(original.Relationships))
	}

	// Deep comparison
	if !reflect.DeepEqual(result, original) {
		t.Errorf("TaxonomyMapping not equal after round-trip\ngot:  %+v\nwant: %+v", result, original)
	}
}

// TestPropertyMappingJSON tests JSON serialization of PropertyMapping
func TestPropertyMappingJSON(t *testing.T) {
	tests := []struct {
		name     string
		mapping  PropertyMapping
		expected string
	}{
		{
			name: "simple mapping",
			mapping: PropertyMapping{
				Source: "hostname",
				Target: "name",
			},
			expected: `{"source":"hostname","target":"name"}`,
		},
		{
			name: "mapping with string default",
			mapping: PropertyMapping{
				Source:  "ip_address",
				Target:  "ip",
				Default: "0.0.0.0",
			},
			expected: `{"source":"ip_address","target":"ip","default":"0.0.0.0"}`,
		},
		{
			name: "mapping with numeric default",
			mapping: PropertyMapping{
				Source:  "port",
				Target:  "port",
				Default: float64(443), // JSON unmarshals numbers as float64
			},
			expected: `{"source":"port","target":"port","default":443}`,
		},
		{
			name: "mapping with transform",
			mapping: PropertyMapping{
				Source:    "domain",
				Target:    "domain",
				Transform: "lowercase",
			},
			expected: `{"source":"domain","target":"domain","transform":"lowercase"}`,
		},
		{
			name: "mapping with default and transform",
			mapping: PropertyMapping{
				Source:    "status",
				Target:    "status",
				Default:   "unknown",
				Transform: "uppercase",
			},
			expected: `{"source":"status","target":"status","default":"unknown","transform":"uppercase"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mapping)
			if err != nil {
				t.Fatalf("failed to marshal PropertyMapping: %v", err)
			}

			if string(data) != tt.expected {
				t.Errorf("JSON mismatch\ngot:  %s\nwant: %s", string(data), tt.expected)
			}

			// Round-trip test
			var result PropertyMapping
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal PropertyMapping: %v", err)
			}

			if !reflect.DeepEqual(result, tt.mapping) {
				t.Errorf("PropertyMapping not equal after round-trip\ngot:  %+v\nwant: %+v", result, tt.mapping)
			}
		})
	}
}

// TestRelationshipMappingJSON tests JSON serialization of RelationshipMapping
func TestRelationshipMappingJSON(t *testing.T) {
	tests := []struct {
		name     string
		mapping  RelationshipMapping
		wantJSON string
	}{
		{
			name: "simple relationship",
			mapping: RelationshipMapping{
				Type:         "HAS_VULNERABILITY",
				FromTemplate: "asset:{{.hostname}}",
				ToTemplate:   "vuln:{{.cve_id}}",
			},
			wantJSON: `{"type":"HAS_VULNERABILITY","from_template":"asset:{{.hostname}}","to_template":"vuln:{{.cve_id}}"}`,
		},
		{
			name: "relationship with condition",
			mapping: RelationshipMapping{
				Type:         "AFFECTS",
				FromTemplate: "vuln:{{.cve_id}}",
				ToTemplate:   "asset:{{.hostname}}",
				Condition:    "{{.severity}} == 'critical'",
			},
			wantJSON: `{"type":"AFFECTS","from_template":"vuln:{{.cve_id}}","to_template":"asset:{{.hostname}}","condition":"{{.severity}} == 'critical'"}`,
		},
		{
			name: "relationship with properties",
			mapping: RelationshipMapping{
				Type:         "CONNECTED_TO",
				FromTemplate: "server:{{.source}}",
				ToTemplate:   "server:{{.target}}",
				Properties: []PropertyMapping{
					{
						Source: "bandwidth",
						Target: "bandwidth",
					},
					{
						Source:  "latency",
						Target:  "latency",
						Default: float64(0), // JSON unmarshals numbers as float64
					},
				},
			},
			wantJSON: `{"type":"CONNECTED_TO","from_template":"server:{{.source}}","to_template":"server:{{.target}}","properties":[{"source":"bandwidth","target":"bandwidth"},{"source":"latency","target":"latency","default":0}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mapping)
			if err != nil {
				t.Fatalf("failed to marshal RelationshipMapping: %v", err)
			}

			if string(data) != tt.wantJSON {
				t.Errorf("JSON mismatch\ngot:  %s\nwant: %s", string(data), tt.wantJSON)
			}

			// Round-trip test
			var result RelationshipMapping
			if err := json.Unmarshal(data, &result); err != nil {
				t.Fatalf("failed to unmarshal RelationshipMapping: %v", err)
			}

			if !reflect.DeepEqual(result, tt.mapping) {
				t.Errorf("RelationshipMapping not equal after round-trip\ngot:  %+v\nwant: %+v", result, tt.mapping)
			}
		})
	}
}

// TestPropMapHelpers tests the property mapping helper functions
func TestPropMapHelpers(t *testing.T) {
	t.Run("PropMap", func(t *testing.T) {
		mapping := PropMap("hostname", "name")

		if mapping.Source != "hostname" {
			t.Errorf("Source mismatch: got %q, want %q", mapping.Source, "hostname")
		}
		if mapping.Target != "name" {
			t.Errorf("Target mismatch: got %q, want %q", mapping.Target, "name")
		}
		if mapping.Default != nil {
			t.Errorf("Default should be nil, got %v", mapping.Default)
		}
		if mapping.Transform != "" {
			t.Errorf("Transform should be empty, got %q", mapping.Transform)
		}
	})

	t.Run("PropMapWithDefault", func(t *testing.T) {
		mapping := PropMapWithDefault("ip_address", "ip", "0.0.0.0")

		if mapping.Source != "ip_address" {
			t.Errorf("Source mismatch: got %q, want %q", mapping.Source, "ip_address")
		}
		if mapping.Target != "ip" {
			t.Errorf("Target mismatch: got %q, want %q", mapping.Target, "ip")
		}
		if mapping.Default != "0.0.0.0" {
			t.Errorf("Default mismatch: got %v, want %q", mapping.Default, "0.0.0.0")
		}
		if mapping.Transform != "" {
			t.Errorf("Transform should be empty, got %q", mapping.Transform)
		}
	})

	t.Run("PropMapWithDefault numeric", func(t *testing.T) {
		mapping := PropMapWithDefault("port", "port", 443)

		if mapping.Default != 443 {
			t.Errorf("Default mismatch: got %v, want %d", mapping.Default, 443)
		}
	})

	t.Run("PropMapWithTransform", func(t *testing.T) {
		mapping := PropMapWithTransform("domain", "domain", "lowercase")

		if mapping.Source != "domain" {
			t.Errorf("Source mismatch: got %q, want %q", mapping.Source, "domain")
		}
		if mapping.Target != "domain" {
			t.Errorf("Target mismatch: got %q, want %q", mapping.Target, "domain")
		}
		if mapping.Transform != "lowercase" {
			t.Errorf("Transform mismatch: got %q, want %q", mapping.Transform, "lowercase")
		}
		if mapping.Default != nil {
			t.Errorf("Default should be nil, got %v", mapping.Default)
		}
	})
}

// TestRelHelpers tests the relationship mapping helper functions
func TestRelHelpers(t *testing.T) {
	t.Run("Rel", func(t *testing.T) {
		rel := Rel("HAS_VULNERABILITY", "asset:{{.hostname}}", "vuln:{{.cve_id}}")

		if rel.Type != "HAS_VULNERABILITY" {
			t.Errorf("Type mismatch: got %q, want %q", rel.Type, "HAS_VULNERABILITY")
		}
		if rel.FromTemplate != "asset:{{.hostname}}" {
			t.Errorf("FromTemplate mismatch: got %q, want %q", rel.FromTemplate, "asset:{{.hostname}}")
		}
		if rel.ToTemplate != "vuln:{{.cve_id}}" {
			t.Errorf("ToTemplate mismatch: got %q, want %q", rel.ToTemplate, "vuln:{{.cve_id}}")
		}
		if rel.Condition != "" {
			t.Errorf("Condition should be empty, got %q", rel.Condition)
		}
		if rel.Properties != nil {
			t.Errorf("Properties should be nil, got %v", rel.Properties)
		}
	})

	t.Run("RelWithCondition", func(t *testing.T) {
		rel := RelWithCondition(
			"AFFECTS",
			"vuln:{{.cve_id}}",
			"asset:{{.hostname}}",
			"{{.severity}} == 'critical'",
		)

		if rel.Type != "AFFECTS" {
			t.Errorf("Type mismatch: got %q, want %q", rel.Type, "AFFECTS")
		}
		if rel.FromTemplate != "vuln:{{.cve_id}}" {
			t.Errorf("FromTemplate mismatch: got %q, want %q", rel.FromTemplate, "vuln:{{.cve_id}}")
		}
		if rel.ToTemplate != "asset:{{.hostname}}" {
			t.Errorf("ToTemplate mismatch: got %q, want %q", rel.ToTemplate, "asset:{{.hostname}}")
		}
		if rel.Condition != "{{.severity}} == 'critical'" {
			t.Errorf("Condition mismatch: got %q, want %q", rel.Condition, "{{.severity}} == 'critical'")
		}
		if rel.Properties != nil {
			t.Errorf("Properties should be nil, got %v", rel.Properties)
		}
	})

	t.Run("RelWithProps", func(t *testing.T) {
		props := []PropertyMapping{
			PropMap("discovered_at", "timestamp"),
			PropMapWithDefault("severity", "severity", "unknown"),
		}
		rel := RelWithProps(
			"CONNECTED_TO",
			"server:{{.source}}",
			"server:{{.target}}",
			props...,
		)

		if rel.Type != "CONNECTED_TO" {
			t.Errorf("Type mismatch: got %q, want %q", rel.Type, "CONNECTED_TO")
		}
		if rel.FromTemplate != "server:{{.source}}" {
			t.Errorf("FromTemplate mismatch: got %q, want %q", rel.FromTemplate, "server:{{.source}}")
		}
		if rel.ToTemplate != "server:{{.target}}" {
			t.Errorf("ToTemplate mismatch: got %q, want %q", rel.ToTemplate, "server:{{.target}}")
		}
		if rel.Condition != "" {
			t.Errorf("Condition should be empty, got %q", rel.Condition)
		}
		if len(rel.Properties) != 2 {
			t.Errorf("Properties length mismatch: got %d, want %d", len(rel.Properties), 2)
		}

		// Verify properties
		if !reflect.DeepEqual(rel.Properties, props) {
			t.Errorf("Properties mismatch\ngot:  %+v\nwant: %+v", rel.Properties, props)
		}
	})

	t.Run("RelWithProps empty", func(t *testing.T) {
		rel := RelWithProps("LINKS_TO", "a:{{.id}}", "b:{{.id}}")

		if len(rel.Properties) != 0 {
			t.Errorf("Properties should be empty, got %v", rel.Properties)
		}
	})
}

// TestWithTaxonomyImmutable tests that WithTaxonomy doesn't modify the original schema
func TestWithTaxonomyImmutable(t *testing.T) {
	original := String()
	if original.Taxonomy != nil {
		t.Fatal("original schema should not have taxonomy set")
	}

	taxonomy := TaxonomyMapping{
		NodeType:   "Asset",
		IDTemplate: "asset:{{.hostname}}",
		Properties: []PropertyMapping{
			PropMap("hostname", "name"),
		},
	}

	// Call WithTaxonomy
	modified := original.WithTaxonomy(taxonomy)

	// Verify original is unchanged
	if original.Taxonomy != nil {
		t.Error("original schema was modified (taxonomy should still be nil)")
	}

	// Verify modified has taxonomy
	if modified.Taxonomy == nil {
		t.Fatal("modified schema should have taxonomy set")
	}

	// Verify taxonomy values
	if modified.Taxonomy.NodeType != "Asset" {
		t.Errorf("NodeType mismatch: got %q, want %q", modified.Taxonomy.NodeType, "Asset")
	}
	if modified.Taxonomy.IDTemplate != "asset:{{.hostname}}" {
		t.Errorf("IDTemplate mismatch: got %q, want %q", modified.Taxonomy.IDTemplate, "asset:{{.hostname}}")
	}

	// Verify other schema fields are copied
	if modified.Type != original.Type {
		t.Errorf("Type mismatch: got %q, want %q", modified.Type, original.Type)
	}
}

// TestWithTaxonomyShallowCopy tests that WithTaxonomy creates a shallow copy
// Note: This is expected behavior - slices are shared between copies
func TestWithTaxonomyShallowCopy(t *testing.T) {
	taxonomy := TaxonomyMapping{
		NodeType:   "Asset",
		IDTemplate: "asset:{{.hostname}}",
		Properties: []PropertyMapping{
			PropMap("hostname", "name"),
		},
	}

	schema := String().WithTaxonomy(taxonomy)

	// Modify the original taxonomy's top-level fields
	taxonomy.NodeType = "Modified"

	// Verify schema's taxonomy top-level fields are unchanged (struct was copied)
	if schema.Taxonomy.NodeType != "Asset" {
		t.Errorf("NodeType should be 'Asset', got %q (was affected by modification)", schema.Taxonomy.NodeType)
	}

	// Note: Slice fields are shared (shallow copy), so modifications to slice elements
	// would affect both. This is expected Go behavior for struct copies with slice fields.
}

// TestWithTaxonomyJSONSerialization tests that taxonomy is correctly serialized
func TestWithTaxonomyJSONSerialization(t *testing.T) {
	taxonomy := TaxonomyMapping{
		NodeType:   "Asset",
		IDTemplate: "asset:{{.hostname}}",
		Properties: []PropertyMapping{
			PropMap("hostname", "name"),
			PropMapWithDefault("ip", "ip", "0.0.0.0"),
		},
		Relationships: []RelationshipMapping{
			Rel("HAS_VULNERABILITY", "asset:{{.hostname}}", "vuln:{{.cve_id}}"),
		},
	}

	schema := Object(map[string]JSON{
		"hostname": String(),
		"ip":       String(),
		"port":     Int(),
	}, "hostname").WithTaxonomy(taxonomy)

	// Marshal to JSON
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("failed to marshal schema with taxonomy: %v", err)
	}

	// Unmarshal back
	var result JSON
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal schema with taxonomy: %v", err)
	}

	// Verify taxonomy was preserved
	if result.Taxonomy == nil {
		t.Fatal("taxonomy was lost during JSON round-trip")
	}

	if result.Taxonomy.NodeType != taxonomy.NodeType {
		t.Errorf("NodeType mismatch: got %q, want %q", result.Taxonomy.NodeType, taxonomy.NodeType)
	}
	if result.Taxonomy.IDTemplate != taxonomy.IDTemplate {
		t.Errorf("IDTemplate mismatch: got %q, want %q", result.Taxonomy.IDTemplate, taxonomy.IDTemplate)
	}
	if len(result.Taxonomy.Properties) != len(taxonomy.Properties) {
		t.Errorf("Properties length mismatch: got %d, want %d", len(result.Taxonomy.Properties), len(taxonomy.Properties))
	}
	if len(result.Taxonomy.Relationships) != len(taxonomy.Relationships) {
		t.Errorf("Relationships length mismatch: got %d, want %d", len(result.Taxonomy.Relationships), len(taxonomy.Relationships))
	}

	// Deep comparison of taxonomy
	if !reflect.DeepEqual(*result.Taxonomy, taxonomy) {
		t.Errorf("Taxonomy not equal after round-trip\ngot:  %+v\nwant: %+v", *result.Taxonomy, taxonomy)
	}
}

// TestWithTaxonomyMultipleCalls tests chaining WithTaxonomy calls
func TestWithTaxonomyMultipleCalls(t *testing.T) {
	schema := String()

	taxonomy1 := TaxonomyMapping{
		NodeType:   "Asset",
		IDTemplate: "asset:{{.id}}",
	}

	taxonomy2 := TaxonomyMapping{
		NodeType:   "Vulnerability",
		IDTemplate: "vuln:{{.cve_id}}",
	}

	// First call
	schema1 := schema.WithTaxonomy(taxonomy1)
	if schema1.Taxonomy.NodeType != "Asset" {
		t.Errorf("First taxonomy NodeType mismatch: got %q, want %q", schema1.Taxonomy.NodeType, "Asset")
	}

	// Second call (overwrite)
	schema2 := schema1.WithTaxonomy(taxonomy2)
	if schema2.Taxonomy.NodeType != "Vulnerability" {
		t.Errorf("Second taxonomy NodeType mismatch: got %q, want %q", schema2.Taxonomy.NodeType, "Vulnerability")
	}

	// Verify first schema unchanged
	if schema1.Taxonomy.NodeType != "Asset" {
		t.Errorf("First schema was modified: got %q, want %q", schema1.Taxonomy.NodeType, "Asset")
	}

	// Verify original unchanged
	if schema.Taxonomy != nil {
		t.Error("Original schema was modified")
	}
}

// TestTaxonomyMappingEmptyFields tests handling of empty/nil fields
func TestTaxonomyMappingEmptyFields(t *testing.T) {
	// Minimal taxonomy
	taxonomy := TaxonomyMapping{
		NodeType:   "Asset",
		IDTemplate: "asset:{{.id}}",
	}

	data, err := json.Marshal(taxonomy)
	if err != nil {
		t.Fatalf("failed to marshal minimal taxonomy: %v", err)
	}

	var result TaxonomyMapping
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal minimal taxonomy: %v", err)
	}

	if result.Properties != nil {
		t.Errorf("Properties should be nil, got %v", result.Properties)
	}
	if result.Relationships != nil {
		t.Errorf("Relationships should be nil, got %v", result.Relationships)
	}
}

// TestComplexTaxonomyMapping tests a complex real-world taxonomy mapping
func TestComplexTaxonomyMapping(t *testing.T) {
	taxonomy := TaxonomyMapping{
		NodeType:   "Asset",
		IDTemplate: "asset:{{.hostname}}:{{.ip}}",
		Properties: []PropertyMapping{
			PropMap("hostname", "name"),
			PropMapWithDefault("ip", "ip_address", "0.0.0.0"),
			PropMapWithTransform("domain", "domain", "lowercase"),
			{
				Source:    "os",
				Target:    "operating_system",
				Default:   "unknown",
				Transform: "uppercase",
			},
		},
		Relationships: []RelationshipMapping{
			Rel("HAS_VULNERABILITY", "asset:{{.hostname}}", "vuln:{{.cve_id}}"),
			RelWithCondition(
				"CRITICAL_VULN",
				"vuln:{{.cve_id}}",
				"asset:{{.hostname}}",
				"{{.cvss_score}} >= 9.0",
			),
			RelWithProps(
				"CONNECTS_TO",
				"asset:{{.source_hostname}}",
				"asset:{{.target_hostname}}",
				PropMap("port", "dst_port"),
				PropMapWithDefault("protocol", "protocol", "tcp"),
				PropMap("timestamp", "last_seen"),
			),
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(taxonomy)
	if err != nil {
		t.Fatalf("failed to marshal complex taxonomy: %v", err)
	}

	// Unmarshal back
	var result TaxonomyMapping
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal complex taxonomy: %v", err)
	}

	// Deep comparison
	if !reflect.DeepEqual(result, taxonomy) {
		t.Errorf("Complex taxonomy not equal after round-trip\ngot:  %+v\nwant: %+v", result, taxonomy)
	}

	// Verify specific fields
	if len(result.Properties) != 4 {
		t.Errorf("Properties length: got %d, want %d", len(result.Properties), 4)
	}
	if len(result.Relationships) != 3 {
		t.Errorf("Relationships length: got %d, want %d", len(result.Relationships), 3)
	}

	// Verify relationship with properties
	connRel := result.Relationships[2]
	if len(connRel.Properties) != 3 {
		t.Errorf("Relationship properties length: got %d, want %d", len(connRel.Properties), 3)
	}
}

package taxonomy

import "testing"

func TestParentRelationshipsExist(t *testing.T) {
	tests := []struct {
		childType        string
		expectedParent   string
		expectedRefField string
		expectedRel      string
	}{
		{"port", "host", "host_id", "HAS_PORT"},
		{"service", "port", "port_id", "RUNS_SERVICE"},
		{"endpoint", "service", "service_id", "HAS_ENDPOINT"},
		{"subdomain", "domain", "domain_id", "HAS_SUBDOMAIN"},
		{"evidence", "finding", "finding_id", "HAS_EVIDENCE"},
		{"mission_run", "mission", "mission_id", "HAS_RUN"},
		{"agent_run", "mission_run", "mission_run_id", "CONTAINS_AGENT_RUN"},
		{"tool_execution", "agent_run", "agent_run_id", "EXECUTED_TOOL"},
		{"llm_call", "agent_run", "agent_run_id", "MADE_CALL"},
	}

	for _, tt := range tests {
		t.Run(tt.childType, func(t *testing.T) {
			rel := GetParentRelationship(tt.childType)
			if rel == nil {
				t.Fatalf("GetParentRelationship(%q) returned nil", tt.childType)
			}
			if rel.ParentType != tt.expectedParent {
				t.Errorf("ParentType = %q, want %q", rel.ParentType, tt.expectedParent)
			}
			if rel.RefField != tt.expectedRefField {
				t.Errorf("RefField = %q, want %q", rel.RefField, tt.expectedRefField)
			}
			if rel.Relationship != tt.expectedRel {
				t.Errorf("Relationship = %q, want %q", rel.Relationship, tt.expectedRel)
			}
			if rel.ParentField != "id" {
				t.Errorf("ParentField = %q, want \"id\"", rel.ParentField)
			}
		})
	}
}

func TestRootNodeTypes(t *testing.T) {
	rootTypes := []string{"host", "domain", "technology", "certificate", "finding", "mission", "technique"}
	for _, nodeType := range rootTypes {
		t.Run(nodeType, func(t *testing.T) {
			if !IsRootNodeType(nodeType) {
				t.Errorf("IsRootNodeType(%q) = false, want true", nodeType)
			}
			if rel := GetParentRelationship(nodeType); rel != nil {
				t.Errorf("GetParentRelationship(%q) = %+v, want nil", nodeType, rel)
			}
		})
	}
}

func TestNonRootNodeTypes(t *testing.T) {
	nonRootTypes := []string{"port", "service", "endpoint", "subdomain", "evidence", "mission_run", "agent_run", "tool_execution", "llm_call"}
	for _, nodeType := range nonRootTypes {
		t.Run(nodeType, func(t *testing.T) {
			if IsRootNodeType(nodeType) {
				t.Errorf("IsRootNodeType(%q) = true, want false", nodeType)
			}
			if rel := GetParentRelationship(nodeType); rel == nil {
				t.Errorf("GetParentRelationship(%q) = nil, want non-nil", nodeType)
			}
		})
	}
}

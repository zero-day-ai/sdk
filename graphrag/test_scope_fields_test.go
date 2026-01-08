package graphrag

import (
	"testing"
)

// TestQuery_ScopeFields verifies the new mission scope fields work correctly
func TestQuery_ScopeFields(t *testing.T) {
	// Test basic query with scope
	q := NewQuery("test query").
		WithMissionScope(ScopeSameMission).
		WithMissionName("recon-webapp").
		WithRunNumber(3).
		WithIncludeRunMetadata(true)

	if q.MissionScope != ScopeSameMission {
		t.Errorf("expected MissionScope to be %s, got %s", ScopeSameMission, q.MissionScope)
	}

	if q.MissionName != "recon-webapp" {
		t.Errorf("expected MissionName to be 'recon-webapp', got %s", q.MissionName)
	}

	if q.RunNumber == nil {
		t.Fatal("expected RunNumber to be set, got nil")
	}

	if *q.RunNumber != 3 {
		t.Errorf("expected RunNumber to be 3, got %d", *q.RunNumber)
	}

	if !q.IncludeRunMetadata {
		t.Error("expected IncludeRunMetadata to be true")
	}

	// Validate should pass
	if err := q.Validate(); err != nil {
		t.Errorf("expected query to be valid, got error: %v", err)
	}
}

// TestQuery_ScopeValidation verifies scope validation rules
func TestQuery_ScopeValidation(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Query)
		wantError bool
	}{
		{
			name: "ScopeSameMission without MissionName",
			setup: func(q *Query) {
				q.MissionScope = ScopeSameMission
				// MissionName is not set
			},
			wantError: true,
		},
		{
			name: "ScopeSameMission with MissionName",
			setup: func(q *Query) {
				q.MissionScope = ScopeSameMission
				q.MissionName = "test-mission"
			},
			wantError: false,
		},
		{
			name: "Invalid RunNumber (zero)",
			setup: func(q *Query) {
				zero := 0
				q.RunNumber = &zero
			},
			wantError: true,
		},
		{
			name: "Invalid RunNumber (negative)",
			setup: func(q *Query) {
				neg := -1
				q.RunNumber = &neg
			},
			wantError: true,
		},
		{
			name: "Valid RunNumber",
			setup: func(q *Query) {
				run := 5
				q.RunNumber = &run
			},
			wantError: false,
		},
		{
			name: "Invalid MissionScope",
			setup: func(q *Query) {
				q.MissionScope = MissionScope("invalid")
			},
			wantError: true,
		},
		{
			name: "Empty MissionScope (defaults to ScopeAll)",
			setup: func(q *Query) {
				q.MissionScope = ""
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewQuery("test query")
			tt.setup(q)

			err := q.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// TestQuery_BackwardsCompatibility ensures old queries still work
func TestQuery_BackwardsCompatibility(t *testing.T) {
	// Old-style query without scope fields
	q := NewQuery("test query")

	// Should have zero-value defaults
	if q.MissionScope != "" {
		t.Errorf("expected empty MissionScope, got %s", q.MissionScope)
	}

	if q.MissionName != "" {
		t.Errorf("expected empty MissionName, got %s", q.MissionName)
	}

	if q.RunNumber != nil {
		t.Errorf("expected nil RunNumber, got %v", q.RunNumber)
	}

	if q.IncludeRunMetadata {
		t.Error("expected IncludeRunMetadata to be false")
	}

	// Should still validate successfully
	if err := q.Validate(); err != nil {
		t.Errorf("backwards compatible query should validate, got error: %v", err)
	}
}

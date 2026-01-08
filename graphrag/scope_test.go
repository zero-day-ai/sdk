package graphrag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissionScope_String(t *testing.T) {
	tests := []struct {
		name     string
		scope    MissionScope
		expected string
	}{
		{
			name:     "current_run",
			scope:    ScopeCurrentRun,
			expected: "current_run",
		},
		{
			name:     "same_mission",
			scope:    ScopeSameMission,
			expected: "same_mission",
		},
		{
			name:     "all",
			scope:    ScopeAll,
			expected: "all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.scope.String())
		})
	}
}

func TestMissionScope_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		scope    MissionScope
		expected bool
	}{
		{
			name:     "valid: current_run",
			scope:    ScopeCurrentRun,
			expected: true,
		},
		{
			name:     "valid: same_mission",
			scope:    ScopeSameMission,
			expected: true,
		},
		{
			name:     "valid: all",
			scope:    ScopeAll,
			expected: true,
		},
		{
			name:     "invalid: empty string",
			scope:    MissionScope(""),
			expected: false,
		},
		{
			name:     "invalid: unknown value",
			scope:    MissionScope("invalid"),
			expected: false,
		},
		{
			name:     "invalid: wrong case",
			scope:    MissionScope("CURRENT_RUN"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.scope.IsValid())
		})
	}
}

func TestMissionScope_Validate(t *testing.T) {
	tests := []struct {
		name      string
		scope     MissionScope
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid: current_run",
			scope:     ScopeCurrentRun,
			wantError: false,
		},
		{
			name:      "valid: same_mission",
			scope:     ScopeSameMission,
			wantError: false,
		},
		{
			name:      "valid: all",
			scope:     ScopeAll,
			wantError: false,
		},
		{
			name:      "invalid: empty string",
			scope:     MissionScope(""),
			wantError: true,
			errorMsg:  "invalid mission scope",
		},
		{
			name:      "invalid: unknown value",
			scope:     MissionScope("invalid_scope"),
			wantError: true,
			errorMsg:  "invalid mission scope: invalid_scope",
		},
		{
			name:      "invalid: typo",
			scope:     MissionScope("current_runs"),
			wantError: true,
			errorMsg:  "invalid mission scope: current_runs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.scope.Validate()
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseMissionScope(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  MissionScope
		wantError bool
	}{
		{
			name:      "valid: current_run",
			input:     "current_run",
			expected:  ScopeCurrentRun,
			wantError: false,
		},
		{
			name:      "valid: same_mission",
			input:     "same_mission",
			expected:  ScopeSameMission,
			wantError: false,
		},
		{
			name:      "valid: all",
			input:     "all",
			expected:  ScopeAll,
			wantError: false,
		},
		{
			name:      "invalid: empty string",
			input:     "",
			wantError: true,
		},
		{
			name:      "invalid: unknown value",
			input:     "unknown",
			wantError: true,
		},
		{
			name:      "invalid: wrong case",
			input:     "CURRENT_RUN",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMissionScope(tt.input)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid mission scope")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestAllMissionScopes(t *testing.T) {
	scopes := AllMissionScopes()

	// Verify we have exactly 3 scopes
	assert.Len(t, scopes, 3)

	// Verify all scopes are present
	assert.Contains(t, scopes, ScopeCurrentRun)
	assert.Contains(t, scopes, ScopeSameMission)
	assert.Contains(t, scopes, ScopeAll)

	// Verify all returned scopes are valid
	for _, scope := range scopes {
		assert.True(t, scope.IsValid(), "scope %s should be valid", scope)
	}
}

func TestDefaultMissionScope(t *testing.T) {
	// Verify default is ScopeAll for backwards compatibility
	assert.Equal(t, ScopeAll, DefaultMissionScope)

	// Verify default is valid
	assert.True(t, DefaultMissionScope.IsValid())
	assert.NoError(t, DefaultMissionScope.Validate())
}

func TestMissionScope_BackwardsCompatibility(t *testing.T) {
	// Test that ScopeAll is "all" for backwards compatibility
	assert.Equal(t, "all", ScopeAll.String())

	// Test that default scope is "all"
	assert.Equal(t, "all", DefaultMissionScope.String())

	// Test that empty/unset scope can default to "all"
	var unsetScope MissionScope
	if unsetScope == "" {
		unsetScope = DefaultMissionScope
	}
	assert.Equal(t, ScopeAll, unsetScope)
}

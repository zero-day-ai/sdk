package graphrag

import "fmt"

// MissionScope defines the scope for filtering GraphRAG query results based on mission context.
// It controls whether results are limited to the current mission run, all runs of the same mission,
// or all missions in the system.
type MissionScope string

const (
	// ScopeCurrentRun limits results to only the current mission run.
	// This is the most restrictive scope, useful for analyzing results specific
	// to the active execution context.
	ScopeCurrentRun MissionScope = "current_run"

	// ScopeSameMission includes results from all runs with the same mission name.
	// This allows tracking trends and patterns across multiple executions of
	// the same mission while maintaining mission isolation.
	ScopeSameMission MissionScope = "same_mission"

	// ScopeAll includes results from all missions in the system.
	// This is the default scope and maintains backwards compatibility with
	// existing queries that do not specify a mission scope.
	ScopeAll MissionScope = "all"
)

// DefaultMissionScope is the default scope used when no scope is specified.
// It defaults to ScopeAll to maintain backwards compatibility with existing code.
const DefaultMissionScope = ScopeAll

// String returns the string representation of the mission scope.
func (s MissionScope) String() string {
	return string(s)
}

// IsValid returns true if the mission scope is a valid value.
func (s MissionScope) IsValid() bool {
	switch s {
	case ScopeCurrentRun, ScopeSameMission, ScopeAll:
		return true
	default:
		return false
	}
}

// Validate checks if the mission scope is valid.
// Returns an error if the scope is not one of the defined constants.
func (s MissionScope) Validate() error {
	if !s.IsValid() {
		return fmt.Errorf("invalid mission scope: %s (must be one of: current_run, same_mission, all)", s)
	}
	return nil
}

// ParseMissionScope parses a string into a MissionScope value.
// Returns an error if the string is not a valid mission scope.
func ParseMissionScope(s string) (MissionScope, error) {
	scope := MissionScope(s)
	if err := scope.Validate(); err != nil {
		return "", err
	}
	return scope, nil
}

// AllMissionScopes returns all valid mission scope values.
func AllMissionScopes() []MissionScope {
	return []MissionScope{
		ScopeCurrentRun,
		ScopeSameMission,
		ScopeAll,
	}
}

package types

import "time"

// MissionContext contains the operational context for a security testing mission.
// It provides agents with mission parameters, constraints, and tracking information.
type MissionContext struct {
	// ID is a unique identifier for the mission.
	ID string `json:"id"`

	// Name is a human-readable name for the mission.
	Name string `json:"name"`

	// CurrentAgent identifies the agent currently executing the mission.
	CurrentAgent string `json:"current_agent,omitempty"`

	// Phase indicates the current mission phase (e.g., "reconnaissance", "exploitation", "reporting").
	Phase string `json:"phase,omitempty"`

	// Constraints defines operational limits and requirements for the mission.
	Constraints MissionConstraints `json:"constraints"`

	// Metadata stores additional mission-specific information.
	// This can include start time, objectives, priorities, team assignments, etc.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// MissionConstraints defines operational limits for mission execution.
// These constraints ensure testing stays within acceptable boundaries.
type MissionConstraints struct {
	// MaxDuration is the maximum time allowed for mission execution.
	// Zero value means no time limit.
	MaxDuration time.Duration `json:"max_duration,omitempty"`

	// MaxFindings is the maximum number of findings to collect before stopping.
	// Zero value means no limit.
	MaxFindings int `json:"max_findings,omitempty"`

	// SeverityThreshold is the minimum severity level required to report findings.
	// Common values: "low", "medium", "high", "critical".
	SeverityThreshold string `json:"severity_threshold,omitempty"`

	// RequireEvidence indicates whether findings must include proof-of-concept evidence.
	RequireEvidence bool `json:"require_evidence"`
}

// Validate checks if the MissionContext has all required fields.
func (m *MissionContext) Validate() error {
	if m.ID == "" {
		return &ValidationError{Field: "ID", Message: "mission ID is required"}
	}

	if m.Name == "" {
		return &ValidationError{Field: "Name", Message: "mission name is required"}
	}

	return nil
}

// GetMetadata retrieves a metadata value by key.
func (m *MissionContext) GetMetadata(key string) (any, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	val, ok := m.Metadata[key]
	return val, ok
}

// SetMetadata sets a metadata value.
func (m *MissionContext) SetMetadata(key string, value any) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]any)
	}
	m.Metadata[key] = value
}

// IsExpired checks if the mission has exceeded its maximum duration.
// Returns false if no max duration is set or no start time is available.
func (m *MissionContext) IsExpired() bool {
	if m.Constraints.MaxDuration == 0 {
		return false
	}

	startTime, ok := m.GetMetadata("start_time")
	if !ok {
		return false
	}

	start, ok := startTime.(time.Time)
	if !ok {
		return false
	}

	return time.Since(start) > m.Constraints.MaxDuration
}

// ShouldStop checks if the mission should stop based on constraints.
// This checks both time limits and finding count limits.
func (m *MissionContext) ShouldStop(findingCount int) bool {
	// Check time limit
	if m.IsExpired() {
		return true
	}

	// Check finding count limit
	if m.Constraints.MaxFindings > 0 && findingCount >= m.Constraints.MaxFindings {
		return true
	}

	return false
}

// MeetsSeverityThreshold checks if a severity level meets the mission threshold.
func (m *MissionConstraints) MeetsSeverityThreshold(severity string) bool {
	if m.SeverityThreshold == "" {
		return true // No threshold set, accept all
	}

	severityLevels := map[string]int{
		"low":      1,
		"medium":   2,
		"high":     3,
		"critical": 4,
	}

	threshold, ok := severityLevels[m.SeverityThreshold]
	if !ok {
		return true // Unknown threshold, accept
	}

	level, ok := severityLevels[severity]
	if !ok {
		return true // Unknown severity, accept
	}

	return level >= threshold
}

// NewMissionContext creates a new mission context with default values.
func NewMissionContext(id, name string) *MissionContext {
	return &MissionContext{
		ID:       id,
		Name:     name,
		Metadata: make(map[string]any),
	}
}

// NewMissionConstraints creates mission constraints with default values.
func NewMissionConstraints() MissionConstraints {
	return MissionConstraints{
		RequireEvidence: true,
	}
}

// WithMaxDuration sets the maximum mission duration.
func (c MissionConstraints) WithMaxDuration(d time.Duration) MissionConstraints {
	c.MaxDuration = d
	return c
}

// WithMaxFindings sets the maximum number of findings.
func (c MissionConstraints) WithMaxFindings(count int) MissionConstraints {
	c.MaxFindings = count
	return c
}

// WithSeverityThreshold sets the minimum severity threshold.
func (c MissionConstraints) WithSeverityThreshold(threshold string) MissionConstraints {
	c.SeverityThreshold = threshold
	return c
}

// WithRequireEvidence sets whether evidence is required.
func (c MissionConstraints) WithRequireEvidence(require bool) MissionConstraints {
	c.RequireEvidence = require
	return c
}

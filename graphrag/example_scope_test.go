package graphrag_test

import (
	"fmt"

	"github.com/zero-day-ai/sdk/graphrag"
)

// ExampleQuery_WithMissionScope demonstrates using mission scope to filter query results
func ExampleQuery_WithMissionScope() {
	// Query only the current mission run
	currentRunQuery := graphrag.NewQuery("SQL injection vulnerabilities").
		WithMissionScope(graphrag.ScopeCurrentRun).
		WithIncludeRunMetadata(true)

	fmt.Printf("Current run query scope: %s\n", currentRunQuery.MissionScope)
	fmt.Printf("Include metadata: %t\n", currentRunQuery.IncludeRunMetadata)

	// Query all runs of the same mission
	sameMissionQuery := graphrag.NewQuery("XSS patterns").
		WithMissionScope(graphrag.ScopeSameMission).
		WithMissionName("web-security-scan").
		WithIncludeRunMetadata(true)

	fmt.Printf("Same mission query scope: %s\n", sameMissionQuery.MissionScope)
	fmt.Printf("Mission name: %s\n", sameMissionQuery.MissionName)

	// Query a specific run
	specificRunQuery := graphrag.NewQuery("authentication bypass").
		WithMissionScope(graphrag.ScopeSameMission).
		WithMissionName("web-security-scan").
		WithRunNumber(3)

	fmt.Printf("Specific run number: %d\n", *specificRunQuery.RunNumber)

	// Output:
	// Current run query scope: current_run
	// Include metadata: true
	// Same mission query scope: same_mission
	// Mission name: web-security-scan
	// Specific run number: 3
}

# Finding Package

Comprehensive security finding types for the Gibson AI/LLM security testing framework.

## Overview

The `finding` package provides strongly-typed structures for documenting security vulnerabilities discovered during AI/LLM testing missions. It includes support for:

- Security finding classification and severity scoring
- Evidence collection and reproduction steps
- MITRE ATT&CK and ATLAS framework mappings
- Multiple export formats (JSON, SARIF, CSV, HTML)
- Advanced filtering and querying capabilities

## Package Structure

```
finding/
├── doc.go              # Package documentation
├── finding.go          # Core Finding type and methods
├── severity.go         # Severity levels and scoring
├── category.go         # Security finding categories
├── evidence.go         # Evidence collection types
├── export.go           # Export formats and filtering
└── *_test.go          # Comprehensive test suite (96.5% coverage)
```

## Core Types

### Finding

The main type representing a discovered security vulnerability:

```go
type Finding struct {
    ID            string        // Unique identifier
    MissionID     string        // Parent mission ID
    AgentName     string        // Discovering agent
    Title         string        // Brief summary
    Description   string        // Detailed description
    Category      Category      // Security category
    Severity      Severity      // Severity level
    Confidence    float64       // Confidence score (0.0-1.0)
    MitreAttack   *MitreMapping // ATT&CK mapping
    MitreAtlas    *MitreMapping // ATLAS mapping
    Evidence      []Evidence    // Supporting evidence
    Reproduction  []ReproStep   // Reproduction steps
    CVSSScore     *float64      // CVSS score (0.0-10.0)
    RiskScore     float64       // Calculated risk score
    Remediation   string        // Fix guidance
    Status        Status        // Current status
    Tags          []string      // Custom tags
    // ... timestamps and metadata
}
```

### Severity Levels

```go
const (
    SeverityCritical Severity = "critical" // Weight: 10.0
    SeverityHigh     Severity = "high"     // Weight: 7.5
    SeverityMedium   Severity = "medium"   // Weight: 5.0
    SeverityLow      Severity = "low"      // Weight: 2.5
    SeverityInfo     Severity = "info"     // Weight: 1.0
)
```

### Categories

```go
const (
    CategoryJailbreak            // LLM safety bypass
    CategoryPromptInjection      // Malicious prompt injection
    CategoryDataExtraction       // Unauthorized data access
    CategoryPrivilegeEscalation  // Privilege elevation
    CategoryDOS                  // Denial of service
    CategoryModelManipulation    // Model behavior modification
    CategoryInformationDisclosure // Information leakage
)
```

### Evidence Types

```go
const (
    EvidenceHTTPRequest  // HTTP request capture
    EvidenceHTTPResponse // HTTP response capture
    EvidenceScreenshot   // Screenshot capture
    EvidenceLog          // Log output
    EvidencePayload      // Attack payload
    EvidenceConversation // Conversation transcript
)
```

## Usage Examples

### Creating a Finding

```go
// Create a new finding
finding := finding.NewFinding(
    "mission-123",
    "sql-injection-agent",
    "SQL Injection in Login Form",
    "Discovered SQL injection vulnerability allowing authentication bypass",
    finding.CategoryDataExtraction,
    finding.SeverityHigh,
)

// Add evidence
finding.AddEvidence(finding.Evidence{
    Type:      finding.EvidenceHTTPRequest,
    Title:     "Malicious Login Request",
    Content:   "POST /login\nusername=admin' OR '1'='1",
    Timestamp: time.Now(),
})

// Add reproduction steps
finding.AddReproductionStep(finding.NewReproStep(
    1,
    "Navigate to login page",
    "https://target.com/login",
    "Login form displayed",
))

finding.AddReproductionStep(finding.NewReproStep(
    2,
    "Submit malicious payload",
    "username=admin' OR '1'='1&password=anything",
    "Successfully logged in as admin",
))

// Set MITRE ATT&CK mapping
finding.SetMitreAttack(finding.NewMitreMapping(
    "enterprise",
    "TA0006",
    "Credential Access",
    "T1078",
    "Valid Accounts",
))

// Validate before use
if err := finding.Validate(); err != nil {
    log.Fatal(err)
}
```

### Filtering Findings

```go
// Create a filter
filter := &finding.Filter{
    MissionID:  "mission-123",
    Severities: []finding.Severity{
        finding.SeverityCritical,
        finding.SeverityHigh,
    },
    Status:    finding.StatusOpen,
    MinScore:  7.0,
    CreatedAfter: time.Now().Add(-24 * time.Hour),
    Limit:     50,
}

// Validate filter
if err := filter.Validate(); err != nil {
    log.Fatal(err)
}

// Check if a finding matches
if filter.Matches(*finding) {
    // Process matching finding
}
```

### Working with Severity

```go
// Check severity weight
weight := finding.SeverityHigh.Weight() // 7.5

// Compare severities
if finding.CompareSeverity(sev1, sev2) > 0 {
    // sev1 is more severe than sev2
}

// Parse from string
severity, err := finding.ParseSeverity("high")
if err != nil {
    log.Fatal(err)
}

// Get all severities
for _, sev := range finding.AllSeverities() {
    fmt.Printf("%s: %.1f\n", sev, sev.Weight())
}
```

### Setting Confidence and Risk Score

```go
// Set confidence (automatically recalculates risk score)
err := finding.SetConfidence(0.8)
if err != nil {
    log.Fatal(err)
}

// Risk score = severity weight * confidence
// For SeverityHigh (7.5) with confidence 0.8:
// RiskScore = 7.5 * 0.8 = 6.0
```

### Managing Finding Status

```go
// Update status
err := finding.SetStatus(finding.StatusConfirmed)
if err != nil {
    log.Fatal(err)
}

// Available statuses
const (
    StatusOpen          // Newly discovered
    StatusConfirmed     // Verified as valid
    StatusResolved      // Fixed/mitigated
    StatusFalsePositive // Determined invalid
)
```

## Export Formats

```go
const (
    FormatJSON  // JSON output
    FormatSARIF // SARIF (Static Analysis Results)
    FormatCSV   // Comma-separated values
    FormatHTML  // HTML report
)

// Get file extension and MIME type
format := finding.FormatJSON
ext := format.FileExtension()  // ".json"
mime := format.MimeType()       // "application/json"
```

## MITRE Framework Integration

### ATT&CK Mapping

```go
attackMapping := finding.NewMitreMapping(
    "enterprise",              // Matrix
    "TA0001",                 // Tactic ID
    "Initial Access",         // Tactic Name
    "T1059",                  // Technique ID
    "Command and Scripting Interpreter", // Technique Name
)
attackMapping.SubTechniques = []string{"T1059.001", "T1059.003"}
finding.SetMitreAttack(attackMapping)
```

### ATLAS Mapping

```go
atlasMapping := finding.NewMitreMapping(
    "atlas",                              // Matrix
    "AML.TA0000",                        // Tactic ID
    "ML Model Access",                   // Tactic Name
    "AML.T0000",                         // Technique ID
    "Infer Training Data Membership",    // Technique Name
)
finding.SetMitreAtlas(atlasMapping)
```

## Validation

All types include comprehensive validation:

```go
// Finding validation
if err := finding.Validate(); err != nil {
    // Checks: ID, MissionID, AgentName, Title, Description,
    // Category, Severity, Confidence, Status, CVSS score,
    // timestamps, evidence, reproduction steps, MITRE mappings
}

// Evidence validation
if err := evidence.Validate(); err != nil {
    // Checks: Type, Title, Content, Timestamp
}

// Filter validation
if err := filter.Validate(); err != nil {
    // Checks: Categories, Severities, Status, scores,
    // time ranges, pagination parameters
}

// MITRE mapping validation
if err := mapping.Validate(); err != nil {
    // Checks: Matrix, TacticID, TacticName,
    // TechniqueID, TechniqueName
}
```

## Best Practices

1. **Always validate** findings before storing or exporting
2. **Use constructors** (`NewFinding`, `NewEvidence`, etc.) to ensure proper initialization
3. **Set confidence levels** accurately to calculate meaningful risk scores
4. **Include evidence** to support findings and enable verification
5. **Add reproduction steps** to ensure findings are actionable
6. **Map to MITRE frameworks** for standardized classification
7. **Use appropriate severity** based on impact and exploitability
8. **Tag findings** for custom categorization and querying

## Testing

Comprehensive test coverage (96.5%):

```bash
go test ./finding/... -v -cover
```

## Performance

- Efficient filtering with early returns
- Minimal allocations for core types
- O(1) severity weight lookups
- Fast validation with cached maps

## Thread Safety

Individual finding operations are not thread-safe. Use synchronization when accessing findings from multiple goroutines:

```go
var mu sync.RWMutex

// Reading
mu.RLock()
matches := filter.Matches(*finding)
mu.RUnlock()

// Writing
mu.Lock()
finding.AddEvidence(evidence)
mu.Unlock()
```

## Dependencies

- `github.com/google/uuid` - UUID generation for finding IDs
- Go standard library only for all other functionality

## Integration with Gibson Framework

This package is designed to work seamlessly with:

- **gibson/types** - Mission and target types
- **gibson/schema** - JSON schema validation
- **gibson/llm** - LLM provider integration
- **gibson/agents** - Agent discovery and reporting

## Version

Part of Gibson SDK v1.0.0

## License

Proprietary - Zero Day AI

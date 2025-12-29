package finding

import (
	"testing"
	"time"
)

func TestNewFinding(t *testing.T) {
	missionID := "mission-123"
	agentName := "agent-sql"
	title := "SQL Injection Found"
	description := "Discovered SQL injection in login form"
	category := CategoryDataExtraction
	severity := SeverityHigh

	before := time.Now()
	finding := NewFinding(missionID, agentName, title, description, category, severity)
	after := time.Now()

	if finding.ID == "" {
		t.Error("NewFinding() ID is empty, want auto-generated UUID")
	}
	if finding.MissionID != missionID {
		t.Errorf("NewFinding() MissionID = %v, want %v", finding.MissionID, missionID)
	}
	if finding.AgentName != agentName {
		t.Errorf("NewFinding() AgentName = %v, want %v", finding.AgentName, agentName)
	}
	if finding.Title != title {
		t.Errorf("NewFinding() Title = %v, want %v", finding.Title, title)
	}
	if finding.Description != description {
		t.Errorf("NewFinding() Description = %v, want %v", finding.Description, description)
	}
	if finding.Category != category {
		t.Errorf("NewFinding() Category = %v, want %v", finding.Category, category)
	}
	if finding.Severity != severity {
		t.Errorf("NewFinding() Severity = %v, want %v", finding.Severity, severity)
	}
	if finding.Confidence != 1.0 {
		t.Errorf("NewFinding() Confidence = %v, want 1.0", finding.Confidence)
	}
	if finding.Status != StatusOpen {
		t.Errorf("NewFinding() Status = %v, want %v", finding.Status, StatusOpen)
	}
	if finding.CreatedAt.Before(before) || finding.CreatedAt.After(after) {
		t.Error("NewFinding() CreatedAt not in expected range")
	}
	if finding.UpdatedAt.Before(before) || finding.UpdatedAt.After(after) {
		t.Error("NewFinding() UpdatedAt not in expected range")
	}
	if finding.RiskScore != severity.Weight() {
		t.Errorf("NewFinding() RiskScore = %v, want %v", finding.RiskScore, severity.Weight())
	}
}

func TestNewFindingWithID(t *testing.T) {
	id := "custom-id-123"
	finding := NewFindingWithID(id, "mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityMedium)

	if finding.ID != id {
		t.Errorf("NewFindingWithID() ID = %v, want %v", finding.ID, id)
	}
}

func TestFinding_Validate(t *testing.T) {
	validFinding := NewFinding(
		"mission-1",
		"agent-1",
		"Test Finding",
		"Test Description",
		CategoryJailbreak,
		SeverityHigh,
	)

	tests := []struct {
		name     string
		finding  *Finding
		wantErr  bool
		errField string
	}{
		{
			name:    "valid finding",
			finding: validFinding,
			wantErr: false,
		},
		{
			name: "missing ID",
			finding: &Finding{
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "ID",
		},
		{
			name: "missing mission ID",
			finding: &Finding{
				ID:          "id-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "mission",
		},
		{
			name: "missing agent name",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "agent",
		},
		{
			name: "missing title",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "title",
		},
		{
			name: "missing description",
			finding: &Finding{
				ID:         "id-1",
				MissionID:  "mission-1",
				AgentName:  "agent-1",
				Title:      "Title",
				Category:   CategoryJailbreak,
				Severity:   SeverityHigh,
				Confidence: 1.0,
				Status:     StatusOpen,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			},
			wantErr:  true,
			errField: "description",
		},
		{
			name: "invalid category",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    Category("invalid"),
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "category",
		},
		{
			name: "invalid severity",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    Severity("invalid"),
				Confidence:  1.0,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "severity",
		},
		{
			name: "confidence too low",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  -0.1,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "confidence",
		},
		{
			name: "confidence too high",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.1,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "confidence",
		},
		{
			name: "invalid status",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      Status("invalid"),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "status",
		},
		{
			name: "CVSS score too low",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CVSSScore:   ptrFloat64(-0.1),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "CVSS",
		},
		{
			name: "CVSS score too high",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CVSSScore:   ptrFloat64(10.1),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "CVSS",
		},
		{
			name: "missing created_at",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				UpdatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "created_at",
		},
		{
			name: "missing updated_at",
			finding: &Finding{
				ID:          "id-1",
				MissionID:   "mission-1",
				AgentName:   "agent-1",
				Title:       "Title",
				Description: "Description",
				Category:    CategoryJailbreak,
				Severity:    SeverityHigh,
				Confidence:  1.0,
				Status:      StatusOpen,
				CreatedAt:   time.Now(),
			},
			wantErr:  true,
			errField: "updated_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.finding.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Finding.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFinding_AddEvidence(t *testing.T) {
	finding := NewFinding("mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityHigh)
	initialUpdateTime := finding.UpdatedAt

	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

	evidence := *NewEvidence(EvidenceHTTPRequest, "Test Request", "POST /api/test")
	finding.AddEvidence(evidence)

	if len(finding.Evidence) != 1 {
		t.Errorf("AddEvidence() resulted in %d evidence, want 1", len(finding.Evidence))
	}
	if finding.Evidence[0].Title != "Test Request" {
		t.Errorf("AddEvidence() evidence title = %v, want Test Request", finding.Evidence[0].Title)
	}
	if !finding.UpdatedAt.After(initialUpdateTime) {
		t.Error("AddEvidence() should update UpdatedAt timestamp")
	}
}

func TestFinding_AddReproductionStep(t *testing.T) {
	finding := NewFinding("mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityHigh)
	initialUpdateTime := finding.UpdatedAt

	time.Sleep(10 * time.Millisecond) // Ensure timestamp difference

	step := NewReproStep(1, "First step", "input data", "expected output")
	finding.AddReproductionStep(step)

	if len(finding.Reproduction) != 1 {
		t.Errorf("AddReproductionStep() resulted in %d steps, want 1", len(finding.Reproduction))
	}
	if finding.Reproduction[0].Description != "First step" {
		t.Errorf("AddReproductionStep() step description = %v, want First step", finding.Reproduction[0].Description)
	}
	if !finding.UpdatedAt.After(initialUpdateTime) {
		t.Error("AddReproductionStep() should update UpdatedAt timestamp")
	}
}

func TestFinding_AddTag(t *testing.T) {
	finding := NewFinding("mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityHigh)

	finding.AddTag("tag1")
	finding.AddTag("tag2")
	finding.AddTag("tag1") // Duplicate

	if len(finding.Tags) != 2 {
		t.Errorf("AddTag() resulted in %d tags, want 2", len(finding.Tags))
	}
	if finding.Tags[0] != "tag1" || finding.Tags[1] != "tag2" {
		t.Errorf("AddTag() tags = %v, want [tag1 tag2]", finding.Tags)
	}
}

func TestFinding_SetConfidence(t *testing.T) {
	finding := NewFinding("mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityHigh)
	initialRiskScore := finding.RiskScore

	err := finding.SetConfidence(0.8)
	if err != nil {
		t.Errorf("SetConfidence() error = %v, want nil", err)
	}
	if finding.Confidence != 0.8 {
		t.Errorf("SetConfidence() confidence = %v, want 0.8", finding.Confidence)
	}
	if finding.RiskScore == initialRiskScore {
		t.Error("SetConfidence() should recalculate RiskScore")
	}

	expectedRiskScore := SeverityHigh.Weight() * 0.8
	if finding.RiskScore != expectedRiskScore {
		t.Errorf("SetConfidence() RiskScore = %v, want %v", finding.RiskScore, expectedRiskScore)
	}

	// Test invalid confidence
	err = finding.SetConfidence(-0.1)
	if err == nil {
		t.Error("SetConfidence() with negative value should return error")
	}

	err = finding.SetConfidence(1.1)
	if err == nil {
		t.Error("SetConfidence() with value > 1.0 should return error")
	}
}

func TestFinding_SetStatus(t *testing.T) {
	finding := NewFinding("mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityHigh)

	err := finding.SetStatus(StatusConfirmed)
	if err != nil {
		t.Errorf("SetStatus() error = %v, want nil", err)
	}
	if finding.Status != StatusConfirmed {
		t.Errorf("SetStatus() status = %v, want %v", finding.Status, StatusConfirmed)
	}

	// Test invalid status
	err = finding.SetStatus(Status("invalid"))
	if err == nil {
		t.Error("SetStatus() with invalid status should return error")
	}
}

func TestFinding_SetMitreAttack(t *testing.T) {
	finding := NewFinding("mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityHigh)

	mapping := NewMitreMapping("enterprise", "TA0001", "Initial Access", "T1059", "Command and Scripting Interpreter")
	finding.SetMitreAttack(mapping)

	if finding.MitreAttack == nil {
		t.Fatal("SetMitreAttack() MitreAttack is nil")
	}
	if finding.MitreAttack.TechniqueID != "T1059" {
		t.Errorf("SetMitreAttack() TechniqueID = %v, want T1059", finding.MitreAttack.TechniqueID)
	}
}

func TestFinding_SetMitreAtlas(t *testing.T) {
	finding := NewFinding("mission-1", "agent-1", "Title", "Description", CategoryJailbreak, SeverityHigh)

	mapping := NewMitreMapping("atlas", "AML.TA0000", "ML Model Access", "AML.T0000", "Infer Training Data Membership")
	finding.SetMitreAtlas(mapping)

	if finding.MitreAtlas == nil {
		t.Fatal("SetMitreAtlas() MitreAtlas is nil")
	}
	if finding.MitreAtlas.TechniqueID != "AML.T0000" {
		t.Errorf("SetMitreAtlas() TechniqueID = %v, want AML.T0000", finding.MitreAtlas.TechniqueID)
	}
}

func TestMitreMapping_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mapping *MitreMapping
		wantErr bool
	}{
		{
			name: "valid mapping",
			mapping: &MitreMapping{
				Matrix:        "enterprise",
				TacticID:      "TA0001",
				TacticName:    "Initial Access",
				TechniqueID:   "T1059",
				TechniqueName: "Command and Scripting Interpreter",
			},
			wantErr: false,
		},
		{
			name: "missing matrix",
			mapping: &MitreMapping{
				TacticID:      "TA0001",
				TacticName:    "Initial Access",
				TechniqueID:   "T1059",
				TechniqueName: "Command and Scripting Interpreter",
			},
			wantErr: true,
		},
		{
			name: "missing tactic ID",
			mapping: &MitreMapping{
				Matrix:        "enterprise",
				TacticName:    "Initial Access",
				TechniqueID:   "T1059",
				TechniqueName: "Command and Scripting Interpreter",
			},
			wantErr: true,
		},
		{
			name: "missing tactic name",
			mapping: &MitreMapping{
				Matrix:        "enterprise",
				TacticID:      "TA0001",
				TechniqueID:   "T1059",
				TechniqueName: "Command and Scripting Interpreter",
			},
			wantErr: true,
		},
		{
			name: "missing technique ID",
			mapping: &MitreMapping{
				Matrix:        "enterprise",
				TacticID:      "TA0001",
				TacticName:    "Initial Access",
				TechniqueName: "Command and Scripting Interpreter",
			},
			wantErr: true,
		},
		{
			name: "missing technique name",
			mapping: &MitreMapping{
				Matrix:      "enterprise",
				TacticID:    "TA0001",
				TacticName:  "Initial Access",
				TechniqueID: "T1059",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mapping.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("MitreMapping.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReproStep_Validate(t *testing.T) {
	tests := []struct {
		name    string
		step    *ReproStep
		wantErr bool
	}{
		{
			name: "valid step",
			step: &ReproStep{
				Order:       1,
				Description: "First step",
				Input:       "input",
				Output:      "output",
			},
			wantErr: false,
		},
		{
			name: "order zero",
			step: &ReproStep{
				Order:       0,
				Description: "Step",
			},
			wantErr: true,
		},
		{
			name: "negative order",
			step: &ReproStep{
				Order:       -1,
				Description: "Step",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			step: &ReproStep{
				Order: 1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.step.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ReproStep.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMitreMapping(t *testing.T) {
	mapping := NewMitreMapping("enterprise", "TA0001", "Initial Access", "T1059", "Command and Scripting Interpreter")

	if mapping.Matrix != "enterprise" {
		t.Errorf("NewMitreMapping() Matrix = %v, want enterprise", mapping.Matrix)
	}
	if mapping.TacticID != "TA0001" {
		t.Errorf("NewMitreMapping() TacticID = %v, want TA0001", mapping.TacticID)
	}
	if mapping.TacticName != "Initial Access" {
		t.Errorf("NewMitreMapping() TacticName = %v, want Initial Access", mapping.TacticName)
	}
	if mapping.TechniqueID != "T1059" {
		t.Errorf("NewMitreMapping() TechniqueID = %v, want T1059", mapping.TechniqueID)
	}
	if mapping.TechniqueName != "Command and Scripting Interpreter" {
		t.Errorf("NewMitreMapping() TechniqueName = %v, want Command and Scripting Interpreter", mapping.TechniqueName)
	}
}

func TestNewReproStep(t *testing.T) {
	step := NewReproStep(1, "First step", "input data", "expected output")

	if step.Order != 1 {
		t.Errorf("NewReproStep() Order = %v, want 1", step.Order)
	}
	if step.Description != "First step" {
		t.Errorf("NewReproStep() Description = %v, want First step", step.Description)
	}
	if step.Input != "input data" {
		t.Errorf("NewReproStep() Input = %v, want input data", step.Input)
	}
	if step.Output != "expected output" {
		t.Errorf("NewReproStep() Output = %v, want expected output", step.Output)
	}
}

func TestCalculateRiskScore(t *testing.T) {
	tests := []struct {
		name       string
		severity   Severity
		confidence float64
		want       float64
	}{
		{"critical max confidence", SeverityCritical, 1.0, 10.0},
		{"high max confidence", SeverityHigh, 1.0, 7.5},
		{"medium max confidence", SeverityMedium, 1.0, 5.0},
		{"low max confidence", SeverityLow, 1.0, 2.5},
		{"info max confidence", SeverityInfo, 1.0, 1.0},
		{"critical half confidence", SeverityCritical, 0.5, 5.0},
		{"high zero confidence", SeverityHigh, 0.0, 0.0},
		{"medium 0.8 confidence", SeverityMedium, 0.8, 4.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateRiskScore(tt.severity, tt.confidence)
			if got != tt.want {
				t.Errorf("calculateRiskScore() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function for tests
func ptrFloat64(f float64) *float64 {
	return &f
}
